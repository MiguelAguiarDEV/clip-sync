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

func TestWSBroadcastSmallClip(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(app.NewMux())
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

	payload := []byte("hola")
	if err := wsjson.Write(ctx, cA, types.Envelope{
		Type: "clip",
		Clip: &types.Clip{MsgID: "m1", Mime: "text/plain", Size: len(payload), Data: payload},
	}); err != nil {
		t.Fatal(err)
	}

	var got types.Envelope
	if err := wsjson.Read(ctx, cB, &got); err != nil {
		t.Fatal(err)
	}
	if got.Type != "clip" || got.Clip == nil {
		t.Fatalf("esperaba clip, got=%+v", got)
	}
	if string(got.Clip.Data) != "hola" {
		t.Fatalf("data=%q", got.Clip.Data)
	}
	if got.From != "A" {
		t.Fatalf("from=%q", got.From)
	}

	shortCtx, cancel2 := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel2()
	var nothing types.Envelope
	if err := wsjson.Read(shortCtx, cA, &nothing); err == nil {
		t.Fatal("A no debería recibir su propio mensaje")
	}
}
