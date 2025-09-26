package app

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "net/http"
    "os"
    "strconv"
    "strings"
    "time"

    "clip-sync/server/internal/httpapi"
    "clip-sync/server/internal/hub"
    "clip-sync/server/internal/logx"
    "clip-sync/server/internal/ws"
)

type App struct {
	Mux *http.ServeMux
	WSS *ws.Server
}

func NewApp() *App {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	h := hub.New(32)
    wss := &ws.Server{
        Hub: h,
        Auth: func(token string) (string, bool) {
            secret := os.Getenv("CLIPSYNC_HMAC_SECRET")
            if secret == "" {
                if token == "" {
                    return "", false
                }
                // modo MVP: token == userID
                return token, true
            }
            return verifyHMACToken(token, secret)
        },
        MaxInlineBytes:     64 << 10,
        RateLimitPerSecond: envInt("CLIPSYNC_RATE_LPS", 0),
        Log: func(event string, fields map[string]any) {
            logx.Info(event, fields)
        },
	}
	// dedupe: capacidad LRU por usuario desde env (0 = off)
	wss.SetDedupeCapacity(envInt("CLIPSYNC_DEDUPE", 128))

	mux.Handle("/ws", wss)

	up := &httpapi.UploadServer{
		Dir:      "./uploads",
		MaxBytes: 50 << 20,
	}
	mux.HandleFunc("POST /upload", up.Upload)
	mux.HandleFunc("GET /d/{id}", up.Download)

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(wss.MetricsSnapshot())
	})

	return &App{Mux: mux, WSS: wss}
}

// Back-compat
func NewMux() http.Handler { return NewApp().Mux }

func envInt(name string, def int) int {
    v := os.Getenv(name)
    if v == "" {
        return def
    }
    n, err := strconv.Atoi(v)
    if err != nil {
        return def
    }
    return n
}

// verifyHMACToken valida tokens con formato: userID:exp_unix:hex(hmac_sha256(secret, userID|exp_unix))
func verifyHMACToken(token, secret string) (string, bool) {
    parts := strings.Split(token, ":")
    if len(parts) != 3 {
        return "", false
    }
    uid, expStr, macHex := parts[0], parts[1], parts[2]
    if uid == "" || expStr == "" || macHex == "" {
        return "", false
    }
    exp, err := strconv.ParseInt(expStr, 10, 64)
    if err != nil {
        return "", false
    }
    if time.Now().Unix() > exp {
        return "", false
    }
    payload := uid + "|" + expStr
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write([]byte(payload))
    sum := mac.Sum(nil)
    want := hex.EncodeToString(sum)
    if !hmac.Equal([]byte(macHex), []byte(want)) {
        return "", false
    }
    return uid, true
}
