package tests

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"clip-sync/server/internal/app"
	"clip-sync/server/pkg/types"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

func dialWS(t *testing.T, url string) (*websocket.Conn, context.Context, func()) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	c, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		cancel()
		t.Fatalf("ws dial: %v", err)
	}
	return c, ctx, func() {
		c.Close(websocket.StatusNormalClosure, "")
		cancel()
	}
}

func hello(t *testing.T, ctx context.Context, c *websocket.Conn, dev string) {
	t.Helper()
	if err := wsjson.Write(ctx, c, types.Envelope{
		Type:  "hello",
		Hello: &types.Hello{Token: "u1", UserID: "u1", DeviceID: dev},
	}); err != nil {
		t.Fatalf("hello: %v", err)
	}
}

// Drop si len(Data) != Size
func TestWSDropsMismatchedSize(t *testing.T) {
	srv := httptest.NewServer(app.NewMux())
	defer srv.Close()
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"

	cA, ctxA, doneA := dialWS(t, wsURL)
	defer doneA()
	cB, _, doneB := dialWS(t, wsURL)
	defer doneB()

	hello(t, ctxA, cA, "A")
	hello(t, ctxA, cB, "B")

	// Data = 2 bytes, Size = 3
	if err := wsjson.Write(ctxA, cA, types.Envelope{
		Type: "clip",
		Clip: &types.Clip{MsgID: "m1", Mime: "text/plain", Size: 3, Data: []byte("hi")},
	}); err != nil {
		t.Fatalf("write: %v", err)
	}

	// B no debe recibir nada
	readCtx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	var got types.Envelope
	if err := wsjson.Read(readCtx, cB, &got); err == nil {
		t.Fatalf("esperaba drop; B recibio: %+v", got)
	}
}

// Drop si falta Data y UploadURL
func TestWSDropsEmptyClip(t *testing.T) {
	srv := httptest.NewServer(app.NewMux())
	defer srv.Close()
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"

	cA, ctxA, doneA := dialWS(t, wsURL)
	defer doneA()
	cB, _, doneB := dialWS(t, wsURL)
	defer doneB()

	hello(t, ctxA, cA, "A")
	hello(t, ctxA, cB, "B")

	if err := wsjson.Write(ctxA, cA, types.Envelope{
		Type: "clip",
		Clip: &types.Clip{MsgID: "m2", Mime: "application/octet-stream", Size: 0},
	}); err != nil {
		t.Fatalf("write: %v", err)
	}

	readCtx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	var got types.Envelope
	if err := wsjson.Read(readCtx, cB, &got); err == nil {
		t.Fatalf("esperaba drop; B recibio: %+v", got)
	}
}
