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

func main() {
	addr := flag.String("addr", "ws://localhost:8080/ws", "WebSocket endpoint")
	token := flag.String("token", "u1", "user token (MVP: token == userID)")
	device := flag.String("device", "A", "device id (unique per device)")
	mode := flag.String("mode", "listen", "listen|send")
	text := flag.String("text", "", "text to send (if empty, read from stdin)")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c, _, err := websocket.Dial(ctx, *addr, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close(websocket.StatusNormalClosure, "")

	hello := types.Envelope{
		Type: "hello",
		Hello: &types.Hello{
			Token:    *token,
			UserID:   *token, // MVP: token == userID
			DeviceID: *device,
		},
	}
	if err := wsjson.Write(ctx, c, hello); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("connected to %s as device %s (mode=%s)\n", *addr, *device, *mode)

	switch *mode {
	case "listen":
		if err := runListen(context.Background(), c); err != nil {
			log.Fatal(err)
		}
	case "send":
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
	default:
		log.Fatalf("unknown mode: %s (use listen|send)", *mode)
	}
}

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
		// Mostrar información útil
		if len(cl.Data) > 0 {
			// Asumimos texto si mime empieza con "text/"
			if strings.HasPrefix(cl.Mime, "text/") {
				fmt.Printf("[from %s] %s\n", cl.From, string(cl.Data))
			} else {
				fmt.Printf("[from %s] %s (%d bytes inline)\n", cl.From, cl.Mime, len(cl.Data))
			}
		} else if cl.UploadURL != "" {
			fmt.Printf("[from %s] large clip: %s (%d bytes) — download via HTTP\n", cl.From, cl.UploadURL, cl.Size)
		} else {
			fmt.Printf("[from %s] empty clip (size=%d mime=%s)\n", cl.From, cl.Size, cl.Mime)
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
			Data:  data, // inline (pequeño)
			From:  "",   // lo rellena el servidor
		},
	}
	return wsjson.Write(ctx, c, env)
}
