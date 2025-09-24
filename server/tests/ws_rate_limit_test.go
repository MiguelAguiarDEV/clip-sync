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

func TestWSRateLimitPerDevice(t *testing.T) {
	t.Setenv("CLIPSYNC_RATE_LPS", "1") // 1 clip/seg por device para un test estable

	srv := httptest.NewServer(app.NewMux())
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cA, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer cA.Close(websocket.StatusNormalClosure, "")

	cB, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer cB.Close(websocket.StatusNormalClosure, "")

	// hello
	if err := wsjson.Write(ctx, cA, types.Envelope{
		Type:  "hello",
		Hello: &types.Hello{Token: "u1", UserID: "u1", DeviceID: "A"},
	}); err != nil {
		t.Fatal(err)
	}
	if err := wsjson.Write(ctx, cB, types.Envelope{
		Type:  "hello",
		Hello: &types.Hello{Token: "u1", UserID: "u1", DeviceID: "B"},
	}); err != nil {
		t.Fatal(err)
	}

	// enviar 10 clips rápido desde A
	payload := []byte("x")
	for i := 0; i < 10; i++ {
		_ = wsjson.Write(ctx, cA, types.Envelope{
			Type: "clip",
			Clip: &types.Clip{MsgID: "m", Mime: "text/plain", Size: len(payload), Data: payload},
		})
	}

	// leer durante ~500ms en B
	deadline := time.Now().Add(500 * time.Millisecond)
	got := 0
	for time.Now().Before(deadline) {
		rc, cancelR := context.WithTimeout(context.Background(), 50*time.Millisecond)
		var env types.Envelope
		if err := wsjson.Read(rc, cB, &env); err == nil && env.Type == "clip" && env.Clip != nil {
			got++
		}
		cancelR()
	}

	if got > 1 {
		t.Fatalf("rate-limit failed: received %d clips, want <= 1 in 500ms window", got)
	}
}
