package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"clip-sync/server/pkg/types"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

func TestHealthzMetrics(t *testing.T) {
	t.Setenv("CLIPSYNC_RATE_LPS", "1")

	srv := httptest.NewServer(NewMux())
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
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
	_ = wsjson.Write(ctx, cA, types.Envelope{Type: "hello", Hello: &types.Hello{Token: "u1", UserID: "u1", DeviceID: "A"}})
	_ = wsjson.Write(ctx, cB, types.Envelope{Type: "hello", Hello: &types.Hello{Token: "u1", UserID: "u1", DeviceID: "B"}})

	// clip inválido (drop)
	_ = wsjson.Write(ctx, cA, types.Envelope{Type: "clip", Clip: &types.Clip{Mime: "text/plain", Size: 0}})
	// clip válido
	payload := []byte("ok")
	_ = wsjson.Write(ctx, cA, types.Envelope{Type: "clip", Clip: &types.Clip{Mime: "text/plain", Size: len(payload), Data: payload}})

	// leer uno
	rctx, cancelRead := context.WithTimeout(context.Background(), 200*time.Millisecond)
	var dummy types.Envelope
	_ = wsjson.Read(rctx, cB, &dummy)
	cancelRead()

	// /healthz
	resp, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("healthz status=%d", resp.StatusCode)
	}
	var m map[string]int64
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		t.Fatal(err)
	}
	if m["conns_current"] < 2 {
		t.Fatalf("conns_current=%d", m["conns_current"])
	}
	if m["clips_total"] < 1 {
		t.Fatalf("clips_total=%d", m["clips_total"])
	}
	if m["drops_total"] < 1 {
		t.Fatalf("drops_total=%d", m["drops_total"])
	}
}
