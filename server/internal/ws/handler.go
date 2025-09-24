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
	// Upgrade a WebSocket
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		// Permite WebSocket moderno; sin compresión para simplicidad
	})
	if err != nil {
		return
	}
	defer c.Close(websocket.StatusNormalClosure, "")

	// Estado de este peer
	var userID, deviceID string

	// Init mapa
	s.mu.Lock()
	if s.conns == nil {
		s.conns = make(map[string]map[string]*websocket.Conn)
	}
	s.mu.Unlock()

	// Loop de lectura
	for {
		var env types.Envelope
		if err := wsjson.Read(r.Context(), c, &env); err != nil {
			// cierre o error → cleanup
			if userID != "" && deviceID != "" {
				s.removeConn(userID, deviceID)
			}
			return
		}

		switch env.Type {
		case "hello":
			// Auth mínima
			if env.Hello == nil {
				continue
			}
			tok := env.Hello.Token
			uid := env.Hello.UserID
			dev := env.Hello.DeviceID
			if s.Auth != nil {
				if got, ok := s.Auth(tok); !ok || (uid != "" && got != uid) {
					// rechazar auth
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
				// drop silencioso
				continue
			}
			// broadcast a todos los devices del mismo usuario excepto el emisor
			s.broadcast(userID, deviceID, types.Envelope{
				Type: "clip",
				Clip: clip,
			})
		default:
			// desconocido → ignorar
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
	s.mu.RLock()
	targets := make([]*websocket.Conn, 0, 4)
	if peers := s.conns[userID]; peers != nil {
		for dev, c := range peers {
			if dev == fromDevice {
				continue
			}
			targets = append(targets, c)
		}
	}
	s.mu.RUnlock()

	// write con timeout corto
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
	// MIME por defecto
	if c.Mime == "" {
		c.Mime = "application/octet-stream"
	}
	// Si hay datos inline, deben cuadrar y respetar límite
	if len(c.Data) > 0 {
		if len(c.Data) != c.Size {
			return false
		}
		if s.MaxInlineBytes > 0 && c.Size > s.MaxInlineBytes {
			return false
		}
		return true
	}
	// Sin datos inline: se requiere URL y tamaño > 0
	if c.UploadURL == "" || c.Size <= 0 {
		return false
	}
	return true
}
