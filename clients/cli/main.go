package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"clip-sync/server/pkg/types"
)

/* ---------- helpers ---------- */

func sleepBackoff(attempt int) {
	d := 500 * time.Millisecond
	for i := 0; i < attempt && d < 5*time.Second; i++ {
		d *= 2
	}
	time.Sleep(d)
}

func dialAndHello(ctx context.Context, addr, token, device string) (*websocket.Conn, error) {
	c, _, err := websocket.Dial(ctx, addr, nil)
	if err != nil {
		return nil, err
	}
	hello := types.Envelope{
		Type:  "hello",
		Hello: &types.Hello{Token: token, UserID: token, DeviceID: device},
	}
	if err := wsjson.Write(ctx, c, hello); err != nil {
		c.Close(websocket.StatusNormalClosure, "")
		return nil, err
	}
	return c, nil
}

/* ---------- modes ---------- */

func runListen(ctx context.Context, c *websocket.Conn) error {
	for {
		var env types.Envelope
		if err := wsjson.Read(ctx, c, &env); err != nil {
			return err
		}
		if env.Type != "clip" || env.Clip == nil {
			continue
		}
		cl := env.Clip
		if len(cl.Data) > 0 {
			if strings.HasPrefix(cl.Mime, "text/") {
				fmt.Printf("[from %s] %s\n", cl.From, string(cl.Data))
			} else {
				fmt.Printf("[from %s] %s (%d bytes inline)\n", cl.From, cl.Mime, len(cl.Data))
			}
		} else if cl.UploadURL != "" {
			fmt.Printf("[from %s] large clip: %s (%d bytes)\n", cl.From, cl.UploadURL, cl.Size)
		}
	}
}

func runSend(ctx context.Context, c *websocket.Conn, text string) error {
	data := []byte(text)
	env := types.Envelope{
		Type: "clip",
		Clip: &types.Clip{
			MsgID: "m-" + time.Now().UTC().Format("20060102T150405.000Z0700"),
			Mime:  "text/plain",
			Size:  len(data),
			Data:  data,
		},
	}
	return wsjson.Write(ctx, c, env)
}

/* ---------- main ---------- */

func main() {
	addr := flag.String("addr", "ws://localhost:8080/ws", "WebSocket endpoint")
	token := flag.String("token", "u1", "user token (MVP: token == userID)")
	device := flag.String("device", "A", "device id (unique per device)")
	mode := flag.String("mode", "listen", "listen|send")
	text := flag.String("text", "", "text to send (send mode). If empty, read from stdin")
	flag.Parse()

	switch *mode {
	case "listen":
		for attempt := 0; ; attempt++ {
			ct, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			c, err := dialAndHello(ct, *addr, *token, *device)
			cancel()
			if err != nil {
				fmt.Println("connect failed:", err)
				sleepBackoff(attempt)
				continue
			}
			fmt.Printf("connected to %s as %s (listening)\n", *addr, *device)
			if err := runListen(context.Background(), c); err != nil {
				fmt.Println("listen error:", err)
			}
			sleepBackoff(attempt)
		}

	case "send":
		ct, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		c, err := dialAndHello(ct, *addr, *token, *device)
		cancel()
		if err != nil {
			log.Fatal(err)
		}
		defer c.Close(websocket.StatusNormalClosure, "")

		payload := strings.TrimSpace(*text)
		if payload == "" {
			b, _ := io.ReadAll(os.Stdin)
			payload = strings.TrimSpace(string(b))
		}
		if payload == "" {
			log.Fatal("send mode: provide --text or pipe stdin")
		}

		if err := runSend(context.Background(), c, payload); err != nil {
			log.Fatal(err)
		}
		fmt.Println("sent")

	default:
		log.Fatalf("unknown -mode=%q (use listen|send)", *mode)
	}
}
