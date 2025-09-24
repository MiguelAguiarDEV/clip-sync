package app

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	"clip-sync/server/internal/httpapi"
	"clip-sync/server/internal/hub"
	"clip-sync/server/internal/ws"
)

func NewMux() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	h := hub.New(32)
	wss := &ws.Server{
		Hub: h,
		Auth: func(token string) (string, bool) {
			if token == "" {
				return "", false
			}
			return token, true
		},
		MaxInlineBytes:     64 << 10,                       // 64 KiB
		RateLimitPerSecond: envInt("CLIPSYNC_RATE_LPS", 0), // 0 = ilimitado por defecto
	}
	mux.Handle("/ws", wss)

	up := &httpapi.UploadServer{
		Dir:      "./uploads",
		MaxBytes: 50 << 20, // 50 MB
	}
	mux.HandleFunc("POST /upload", up.Upload)
	mux.HandleFunc("GET /d/{id}", up.Download)

	// /healthz con métricas
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(wss.MetricsSnapshot())
	})

	return mux
}

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
