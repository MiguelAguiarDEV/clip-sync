package ws

import (
	"context"
	"net/http"
	"sync"
	"time"

	"clip-sync/server/internal/hub"
	"clip-sync/server/pkg/types"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type Server struct {
	Hub            *hub.Hub
	Auth           func(token string) (string, bool)
	MaxInlineBytes int // límite para clips inline (bytes)

	mu    sync.RWMutex
	conns map[string]map[string]*websocket.Conn // userID -> deviceID -> conn
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
				continue
			}
			// incluir emisor
			out := types.Envelope{
				Type: "clip",
				From: deviceID,
				Clip: clip,
			}
			s.broadcast(userID, deviceID, out)

		default:
			// ignorar
		}
	}
}

func (s *Server) addConn(userID, deviceID string, c *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conns[userID] == nil {
		s.conns[userID] = make(map[string]*websocket.Conn)
	}
	s.conns[userID][deviceID] = c
}

func (s *Server) removeConn(userID, deviceID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if m := s.conns[userID]; m != nil {
		delete(m, deviceID)
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

	// Primer intento
	targets := buildTargets()
	// Si aún no hay receptores (p. ej. B todavía no procesó su hello),
	// espera una fracción de segundo y reintenta una vez.
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

// Validación de clips
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
