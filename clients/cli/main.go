package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"clip-sync/server/pkg/types"
)

/* ---------- helpers ---------- */

const (
    exitOK     = 0
    exitUsage  = 2
    exitConn   = 10
    exitUpload = 11
    exitSend   = 12
)

func fatalf(code int, format string, args ...any) {
    fmt.Fprintf(os.Stderr, format+"\n", args...)
    os.Exit(code)
}

func sleepBackoff(attempt int) { time.Sleep(computeBackoff(attempt)) }

func computeBackoff(attempt int) time.Duration {
    d := 500 * time.Millisecond
    for i := 0; i < attempt && d < 5*time.Second; i++ {
        d *= 2
        if d > 5*time.Second {
            d = 5 * time.Second
        }
    }
    return d
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

func httpBaseFromWS(wsAddr string) string {
	if strings.HasPrefix(wsAddr, "wss://") {
		return "https://" + strings.TrimPrefix(wsAddr, "wss://")
	}
	if strings.HasPrefix(wsAddr, "ws://") {
		return "http://" + strings.TrimPrefix(wsAddr, "ws://")
	}
	return "http://" + wsAddr
}

func uploadFile(ctx context.Context, httpBase, path, contentType string) (uploadURL string, size int, err error) {
	f, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer f.Close()

	req, err := http.NewRequestWithContext(ctx, "POST", strings.TrimRight(httpBase, "/")+"/upload", f)
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Content-Type", contentType)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", 0, fmt.Errorf("upload failed: status=%d body=%s", resp.StatusCode, string(b))
	}

	var out struct {
		UploadURL string `json:"upload_url"`
		Size      int    `json:"size"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", 0, err
	}
	return out.UploadURL, out.Size, nil
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
				fmt.Printf("[from %s] %s\n", env.From, string(cl.Data))
			} else {
				fmt.Printf("[from %s] %s (%d bytes inline)\n", env.From, cl.Mime, len(cl.Data))
			}
		} else if cl.UploadURL != "" {
			fmt.Printf("[from %s] large clip: %s (%d bytes)\n", env.From, cl.UploadURL, cl.Size)
		}
	}
}

func runSendText(ctx context.Context, c *websocket.Conn, text string) error {
	data := []byte(text)
	if len(data) > types.MaxInlineBytes {
		return fmt.Errorf("text payload is %d bytes; exceeds MaxInlineBytes=%d — use --file",
			len(data), types.MaxInlineBytes)
	}
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

func runSendFile(ctx context.Context, c *websocket.Conn, wsAddr, path, mimeType string) error {
    base := httpBaseFromWS(wsAddr)
    upCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()

	if mimeType == "" {
		mimeType = detectMime(path, "application/octet-stream")
	}
	uploadURL, size, err := uploadFile(upCtx, base, path, mimeType)
	if err != nil {
		return err
	}

	env := types.Envelope{
		Type: "clip",
		Clip: &types.Clip{
			MsgID:     "m-" + time.Now().UTC().Format("20060102T150405.000Z0700"),
			Mime:      mimeType,
			Size:      size,
			UploadURL: uploadURL,
		},
	}
	if err := wsjson.Write(ctx, c, env); err != nil {
		return err
	}
	fmt.Printf("sent file: %s (%d bytes) url=%s\n", path, size, uploadURL)
	return nil
}

/* ---------- main ---------- */

func main() {
	addr := flag.String("addr", "ws://localhost:8080/ws", "WebSocket endpoint")
	token := flag.String("token", "u1", "user token (MVP: token == userID)")
	device := flag.String("device", "A", "device id (unique per device)")
	mode := flag.String("mode", "listen", "listen|send")
	text := flag.String("text", "", "text to send (send mode). If empty, read from stdin")
	file := flag.String("file", "", "path to file to send (uses HTTP /upload)")
    mime := flag.String("mime", "", "mime type for --file (auto-detect if empty)")
	flag.Parse()

	switch *mode {
	case "listen":
        for attempt := 0; ; attempt++ {
            ct, cancel := context.WithTimeout(context.Background(), 5*time.Second)
            c, err := dialAndHello(ct, *addr, *token, *device)
            cancel()
            if err != nil {
                fmt.Fprintln(os.Stderr, "connect failed:", err)
                sleepBackoff(attempt)
                continue
            }
            fmt.Printf("connected to %s as %s (listening)\n", *addr, *device)
            attempt = 0 // reset after a successful connect
            if err := runListen(context.Background(), c); err != nil {
                fmt.Fprintln(os.Stderr, "listen error:", err)
            }
            sleepBackoff(attempt)
        }

	case "send":
        var c *websocket.Conn
        var err error
        for attempt := 0; attempt < 5; attempt++ {
            ct, cancel := context.WithTimeout(context.Background(), 5*time.Second)
            c, err = dialAndHello(ct, *addr, *token, *device)
            cancel()
            if err == nil {
                break
            }
            if attempt == 4 {
                fatalf(exitConn, "connect failed: %v", err)
            }
            fmt.Fprintln(os.Stderr, "connect failed:", err)
            sleepBackoff(attempt)
        }
        defer c.Close(websocket.StatusNormalClosure, "")

		if *file != "" {
            if err := runSendFile(context.Background(), c, *addr, *file, *mime); err != nil {
                fatalf(exitUpload, "%v", err)
            }
            return
        }

		payload := strings.TrimSpace(*text)
		if payload == "" {
			b, _ := io.ReadAll(os.Stdin)
			payload = strings.TrimSpace(string(b))
		}
        if payload == "" {
            fatalf(exitUsage, "send mode: provide --text or --file (or pipe stdin)")
        }
        if err := runSendText(context.Background(), c, payload); err != nil {
            fatalf(exitSend, "%v", err)
        }
        fmt.Println("sent")

	default:
        fatalf(exitUsage, "unknown -mode=%q (use listen|send)", *mode)
    }
}

// detectMime returns mime type by file extension; fallback if unknown.
func detectMime(path, fallback string) string {
    ext := strings.ToLower(filepath.Ext(path))
    if v := extMime[ext]; v != "" {
        return v
    }
    if t := mime.TypeByExtension(ext); t != "" {
        if semi := strings.IndexByte(t, ';'); semi > 0 {
            t = strings.TrimSpace(t[:semi])
        }
        return t
    }
    if fallback == "" {
        return "application/octet-stream"
    }
    return fallback
}

var extMime = map[string]string{
    ".txt":  "text/plain",
    ".md":   "text/markdown",
    ".json": "application/json",
    ".pdf":  "application/pdf",
    ".png":  "image/png",
    ".jpg":  "image/jpeg",
    ".jpeg": "image/jpeg",
    ".gif":  "image/gif",
    ".html": "text/html",
}
