package tests

import (
	"context"
	"strings"
	"testing"
	"time"

	"clip-sync/server/internal/app"
	"clip-sync/server/pkg/types"
	"net/http/httptest"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

func TestWSDropsTooBigInlineClip(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(app.NewMux())
	defer srv.Close()
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Conectar A y B + hello
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

	// A envía inline > MaxInlineBytes → debe descartarse
	n := types.MaxInlineBytes + 1
	big := make([]byte, n)
	for i := range big {
		big[i] = 'X'
	}

	if err := wsjson.Write(ctx, cA, types.Envelope{
		Type: "clip",
		Clip: &types.Clip{
			MsgID: "too-big",
			Mime:  "application/octet-stream",
			Size:  int(n),
			Data:  big,
		},
	}); err != nil {
		t.Fatal(err)
	}

	// B NO debe recibir nada
	shortCtx, cancel2 := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel2()
	var got types.Envelope
	if err := wsjson.Read(shortCtx, cB, &got); err == nil {
		t.Fatalf("no debería recibir; got=%+v", got)
	}
}
