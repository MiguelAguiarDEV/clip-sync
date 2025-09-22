package ws

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"clip-sync/server/internal/hub"
	"clip-sync/server/pkg/types"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

// Keep-alive / timeouts
const (
	pongWait    = 20 * time.Second // sin actividad de lectura en este tiempo → cerramos
	pingEvery   = pongWait / 2     // frecuencia de pings de servidor a cliente
	readTimeout = pongWait         // timeout por cada lectura wsjson.Read
)

// Server conecta WebSocket con el Hub y la autenticación.
type Server struct {
	Hub  *hub.Hub
	Auth func(token string) (userID string, ok bool)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 1) Aceptar WS
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{})
	if err != nil {
		return
	}
	defer c.Close(websocket.StatusInternalError, "internal")

	ctx := r.Context()

	// 1.1) Ping goroutine (keep-alive)
	pingCtx, stopPing := context.WithCancel(ctx)
	defer stopPing()
	go func() {
		t := time.NewTicker(pingEvery)
		defer t.Stop()
		for {
			select {
			case <-pingCtx.Done():
				return
			case <-t.C:
				if err := c.Ping(pingCtx); err != nil {
					c.Close(websocket.StatusPolicyViolation, "ping failed")
					return
				}
			}
		}
	}()

	// 2) Handshake: esperar "hello"
	var env types.Envelope
	if err := wsjson.Read(ctx, c, &env); err != nil || env.Type != "hello" || env.Hello == nil {
		c.Close(websocket.StatusProtocolError, "need hello")
		return
	}
	userID, ok := s.Auth(env.Hello.Token)
	if !ok {
		c.Close(websocket.StatusPolicyViolation, "auth")
		return
	}
	deviceID := env.Hello.DeviceID
	if deviceID == "" {
		c.Close(websocket.StatusProtocolError, "need device_id")
		return
	}

	// 3) Unirse al Hub
	outCh, leave := s.Hub.Join(userID, deviceID)
	defer leave()

	// 4) Writer: del Hub → cliente
	go func(ctx context.Context) {
		for msg := range outCh {
			_ = c.Write(ctx, websocket.MessageText, msg)
		}
		c.Close(websocket.StatusNormalClosure, "")
	}(ctx)

	// 5) Reader: del cliente → Hub (con validaciones y timeout por lectura)
	for {
		rc, cancel := context.WithTimeout(ctx, readTimeout)
		var in types.Envelope
		err := wsjson.Read(rc, c, &in)
		cancel()
		if err != nil {
			break // cierre del cliente o timeout
		}

		if in.Type == "clip" && in.Clip != nil {
			cl := in.Clip

			// Validaciones:
			if len(cl.Data) > 0 {
				// Inline: tamaño consistente y límite
				if len(cl.Data) != cl.Size {
					continue
				}
				if cl.Size > types.MaxInlineBytes {
					continue // demasiado grande para WS → usar /upload
				}
				if cl.Mime == "" {
					cl.Mime = "application/octet-stream"
				}
			} else {
				// Grande: debe traer URL
				if cl.UploadURL == "" {
					continue
				}
			}

			in.Clip.From = deviceID
			b, _ := json.Marshal(in)
			s.Hub.Broadcast(userID, deviceID, b)
		}
	}

	// 6) Cierre limpio
	c.Close(websocket.StatusNormalClosure, "")
}
