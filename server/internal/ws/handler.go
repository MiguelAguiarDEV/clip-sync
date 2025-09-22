package ws

import (
	"context"
	"encoding/json"
	"net/http"

	"clip-sync/server/internal/hub"
	"clip-sync/server/pkg/types"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type Server struct {
	Hub  *hub.Hub
	Auth func(token string) (userID string, ok bool)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{})
	if err != nil {
		return
	}
	defer c.Close(websocket.StatusInternalError, "internal")

	ctx := r.Context()

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

	outCh, leave := s.Hub.Join(userID, deviceID)
	defer leave()

	go func(ctx context.Context) {
		for msg := range outCh {
			_ = c.Write(ctx, websocket.MessageText, msg)
		}
		c.Close(websocket.StatusNormalClosure, "")
	}(ctx)

	for {
		var in types.Envelope
		if err := wsjson.Read(ctx, c, &in); err != nil {
			break
		}
		if in.Type == "clip" && in.Clip != nil {
			in.Clip.From = deviceID
			b, _ := json.Marshal(in)
			s.Hub.Broadcast(userID, deviceID, b)
		}
	}

	c.Close(websocket.StatusNormalClosure, "")
}
