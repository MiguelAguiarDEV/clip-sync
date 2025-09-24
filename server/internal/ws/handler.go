package ws

import (
	"context"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"clip-sync/server/internal/hub"
	"clip-sync/server/pkg/types"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type limiter struct {
	rate     float64
	capacity float64
	tokens   float64
	last     time.Time
}

func newLimiter(rps int) *limiter {
	r := float64(rps)
	if r <= 0 {
		r = 1e9 // effectively unlimited
	}
	now := time.Now()
	return &limiter{rate: r, capacity: r, tokens: r, last: now}
}

func (l *limiter) allow() bool {
	now := time.Now()
	el := now.Sub(l.last).Seconds()
	l.last = now
	l.tokens += el * l.rate
	if l.tokens > l.capacity {
		l.tokens = l.capacity
	}
	if l.tokens >= 1 {
		l.tokens -= 1
		return true
	}
	return false
}

type Server struct {
	Hub                *hub.Hub
	Auth               func(token string) (string, bool)
	MaxInlineBytes     int
	RateLimitPerSecond int

	mu    sync.RWMutex
	conns map[string]map[string]*websocket.Conn // userID -> deviceID -> conn

	rlmu sync.Mutex
	rl   map[string]*limiter // key: userID|deviceID

	metrics struct {
		clips int64 // accepted clips
		drops int64 // dropped clips
		conns int64 // current connections
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{})
	if err != nil {
		return
	}
	defer c.Close(websocket.StatusNormalClosure, "")

	var userID, deviceID string

	s.mu.Lock()
	if s.conns == nil {
		s.conns = make(map[string]map[string]*websocket.Conn)
	}
	s.mu.Unlock()

	for {
		var env types.Envelope
		if err := wsjson.Read(r.Context(), c, &env); err != nil {
			if userID != "" && deviceID != "" {
				s.removeConn(userID, deviceID)
			}
			return
		}

		switch env.Type {
		case "hello":
			if env.Hello == nil {
				continue
			}
			tok := env.Hello.Token
			uid := env.Hello.UserID
			dev := env.Hello.DeviceID
			if s.Auth != nil {
				if got, ok := s.Auth(tok); !ok || (uid != "" && got != uid) {
					_ = c.Close(websocket.StatusPolicyViolation, "unauthorized")
					return
				}
			}
			userID, deviceID = uid, dev
			s.addConn(uid, dev, c)

		case "clip":
			if env.Clip == nil {
				continue
			}
			clip := env.Clip
			if !s.validateClip(clip) {
				atomic.AddInt64(&s.metrics.drops, 1)
				continue
			}
			if !s.allow(userID, deviceID) {
				atomic.AddInt64(&s.metrics.drops, 1)
				continue
			}
			atomic.AddInt64(&s.metrics.clips, 1)

			out := types.Envelope{
				Type: "clip",
				From: deviceID,
				Clip: clip,
			}
			s.broadcast(userID, deviceID, out)

		default:
			// ignore
		}
	}
}

func (s *Server) allow(userID, deviceID string) bool {
	if s.RateLimitPerSecond <= 0 {
		return true
	}
	key := userID + "|" + deviceID
	s.rlmu.Lock()
	if s.rl == nil {
		s.rl = make(map[string]*limiter)
	}
	lim := s.rl[key]
	if lim == nil {
		lim = newLimiter(s.RateLimitPerSecond)
		s.rl[key] = lim
	}
	s.rlmu.Unlock()
	return lim.allow()
}

func (s *Server) addConn(userID, deviceID string, c *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conns[userID] == nil {
		s.conns[userID] = make(map[string]*websocket.Conn)
	}
	s.conns[userID][deviceID] = c
	atomic.AddInt64(&s.metrics.conns, 1)
}

func (s *Server) removeConn(userID, deviceID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if m := s.conns[userID]; m != nil {
		if _, ok := m[deviceID]; ok {
			delete(m, deviceID)
			atomic.AddInt64(&s.metrics.conns, -1)
		}
		if len(m) == 0 {
			delete(s.conns, userID)
		}
	}
}

func (s *Server) broadcast(userID, fromDevice string, env types.Envelope) {
	buildTargets := func() []*websocket.Conn {
		s.mu.RLock()
		defer s.mu.RUnlock()
		list := make([]*websocket.Conn, 0, 4)
		if peers := s.conns[userID]; peers != nil {
			for dev, c := range peers {
				if dev == fromDevice {
					continue
				}
				list = append(list, c)
			}
		}
		return list
	}

	targets := buildTargets()
	if len(targets) == 0 {
		time.Sleep(50 * time.Millisecond)
		targets = buildTargets()
	}

	for _, c := range targets {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		_ = wsjson.Write(ctx, c, env)
		cancel()
	}
}

func (s *Server) validateClip(c *types.Clip) bool {
	if c == nil {
		return false
	}
	if c.Mime == "" {
		c.Mime = "application/octet-stream"
	}
	if len(c.Data) > 0 {
		if len(c.Data) != c.Size {
			return false
		}
		if s.MaxInlineBytes > 0 && c.Size > s.MaxInlineBytes {
			return false
		}
		return true
	}
	if c.UploadURL == "" || c.Size <= 0 {
		return false
	}
	return true
}

func (s *Server) MetricsSnapshot() map[string]int64 {
	return map[string]int64{
		"clips_total":   atomic.LoadInt64(&s.metrics.clips),
		"drops_total":   atomic.LoadInt64(&s.metrics.drops),
		"conns_current": atomic.LoadInt64(&s.metrics.conns),
	}
}
