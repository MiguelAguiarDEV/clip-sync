package app

import (
	"net/http"

	"clip-sync/server/internal/httpapi"
	"clip-sync/server/internal/hub"
	"clip-sync/server/internal/ws"
)

func NewMux() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	h := hub.New(32)
	wss := &ws.Server{
		Hub: h,
		Auth: func(token string) (string, bool) {
			if token == "" {
				return "", false
			}
			// MVP: token == userID
			return token, true
		},
	}
	mux.Handle("/ws", wss)

	up := &httpapi.UploadServer{
		Dir:      "./uploads",
		MaxBytes: 50 << 20, // 50 MiB
	}
	mux.HandleFunc("POST /upload", up.Upload)
	mux.HandleFunc("GET /d/{id}", up.Download)

	return mux
}
