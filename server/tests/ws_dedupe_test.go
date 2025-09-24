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

func TestWSDedupeByMsgID(t *testing.T) {
	t.Setenv("CLIPSYNC_DEDUPE", "64")

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

	// hello
	_ = wsjson.Write(ctx, cA, types.Envelope{Type: "hello", Hello: &types.Hello{Token: "u1", UserID: "u1", DeviceID: "A"}})
	_ = wsjson.Write(ctx, cB, types.Envelope{Type: "hello", Hello: &types.Hello{Token: "u1", UserID: "u1", DeviceID: "B"}})

	payload := []byte("dedupe")
	msgID := "m-dup-1"

	// primer envío
	_ = wsjson.Write(ctx, cA, types.Envelope{
		Type: "clip",
		Clip: &types.Clip{MsgID: msgID, Mime: "text/plain", Size: len(payload), Data: payload},
	})
	// duplicado
	_ = wsjson.Write(ctx, cA, types.Envelope{
		Type: "clip",
		Clip: &types.Clip{MsgID: msgID, Mime: "text/plain", Size: len(payload), Data: payload},
	})

	// B debería recibir solo uno
	read1Ctx, cancel1 := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel1()
	var got1 types.Envelope
	if err := wsjson.Read(read1Ctx, cB, &got1); err != nil {
		t.Fatalf("no recibio primero: %v", err)
	}
	read2Ctx, cancel2 := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel2()
	var got2 types.Envelope
	if err := wsjson.Read(read2Ctx, cB, &got2); err == nil {
		t.Fatalf("recibio duplicado: %+v", got2)
	}
}
