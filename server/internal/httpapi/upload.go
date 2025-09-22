package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

type UploadServer struct {
	Dir      string // carpeta destino
	MaxBytes int64  // límite por subida (ej. 50 << 20)
}

func randID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

func (u *UploadServer) Upload(w http.ResponseWriter, r *http.Request) {
	// Limitar tamaño de lectura
	if u.MaxBytes > 0 {
		r.Body = http.MaxBytesReader(w, r.Body, u.MaxBytes)
	}

	// Asegurar directorio
	if err := os.MkdirAll(u.Dir, 0o755); err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	// Crear archivo con id aleatorio
	id := randID()
	fp := filepath.Join(u.Dir, id)
	f, err := os.Create(fp)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	// Copiar body a disco
	n, err := io.Copy(f, r.Body)
	if err != nil {
		http.Error(w, "bad upload", http.StatusBadRequest)
		return
	}

	// Responder JSON mínimo
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"upload_url":"/d/` + id + `","size":` + strconv.FormatInt(n, 10) + `}`))
}
