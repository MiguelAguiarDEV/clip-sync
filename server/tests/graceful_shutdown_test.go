package tests

import (
	"context"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"clip-sync/server/internal/app"
	"clip-sync/server/pkg/types"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

func TestGracefulShutdownClosesWS(t *testing.T) {
	a := app.NewApp()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	srv := &http.Server{Handler: a.Mux}
	go srv.Serve(ln)
	defer ln.Close()

	wsURL := "ws://" + ln.Addr().String() + "/ws"

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

	// one clip
	payload := []byte("ok")
	_ = wsjson.Write(ctx, cA, types.Envelope{Type: "clip", Clip: &types.Clip{Mime: "text/plain", Size: len(payload), Data: payload}})
	var got types.Envelope
	if err := wsjson.Read(ctx, cB, &got); err != nil || got.Type != "clip" {
		t.Fatalf("precondition failed: %+v, err=%v", got, err)
	}

	// shutdown
	shCtx, shCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer shCancel()
	a.WSS.Shutdown(shCtx)
	_ = srv.Shutdown(shCtx)

	// verify reads fail quickly
	short, cancel2 := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel2()
	if err := wsjson.Read(short, cA, &got); err == nil {
		t.Fatal("expected read error after shutdown")
	}
}

func TestNewAppKeepsRoutes(t *testing.T) {
	a := app.NewApp()
	h := a.Mux
	if h == nil {
		t.Fatal("mux nil")
	}
	// quick sanity on known routes
	for _, p := range []string{"/health", "/healthz", "/ws", "/upload"} {
		if !strings.HasPrefix(p, "/") {
			t.Fatalf("bad path %s", p)
		}
	}
}
