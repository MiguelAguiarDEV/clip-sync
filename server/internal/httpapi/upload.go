package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

type UploadServer struct {
	Dir      string
	MaxBytes int64
}

type uploadResp struct {
	UploadURL string `json:"upload_url"`
	Size      int    `json:"size"`
}

var idRe = regexp.MustCompile(`^[a-f0-9]{32}$`)

func (s *UploadServer) Upload(w http.ResponseWriter, r *http.Request) {
	if s.MaxBytes > 0 {
		r.Body = http.MaxBytesReader(w, r.Body, s.MaxBytes)
	}
	if err := os.MkdirAll(s.Dir, 0o755); err != nil {
		http.Error(w, "storage error", http.StatusInternalServerError)
		return
	}

	id := randHex(16)
	final := filepath.Join(s.Dir, id)

	// escribir a tmp y luego rename
	tmp, err := os.CreateTemp(s.Dir, ".upload-*")
	if err != nil {
		http.Error(w, "storage error", http.StatusInternalServerError)
		return
	}
	tmpName := tmp.Name()
	defer func() {
		tmp.Close()
		_ = os.Remove(tmpName) // no pasa nada si ya se renombró
	}()

	n, err := io.Copy(tmp, r.Body)
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			http.Error(w, "payload too large", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "write error", http.StatusInternalServerError)
		return
	}

	// fsync para asegurar persistencia antes del rename (opcional)
	if err := tmp.Sync(); err != nil {
		http.Error(w, "fsync error", http.StatusInternalServerError)
		return
	}
	if err := tmp.Close(); err != nil {
		http.Error(w, "close error", http.StatusInternalServerError)
		return
	}

	if err := os.Rename(tmpName, final); err != nil {
		http.Error(w, "rename error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(uploadResp{
		UploadURL: "/d/" + id,
		Size:      int(n),
	})
}

func (s *UploadServer) Download(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !idRe.MatchString(id) {
		http.NotFound(w, r)
		return
	}
	fp := filepath.Join(s.Dir, id)
	f, err := os.Open(fp)
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "read error", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "inline; filename="+id)
	_, _ = io.Copy(w, f)
}

func randHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
