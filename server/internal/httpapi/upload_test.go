package httpapi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestUpload_RawOK(t *testing.T) {
	dir := t.TempDir()
	s := &UploadServer{Dir: dir, MaxBytes: 1 << 20}

	body := bytes.Repeat([]byte("X"), 100_000)
	req := httptest.NewRequest(http.MethodPost, "/upload", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/octet-stream")
	rr := httptest.NewRecorder()

	http.HandlerFunc(s.Upload).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	ents, _ := os.ReadDir(dir)
	if len(ents) != 1 || ents[0].IsDir() {
		t.Fatalf("esperaba 1 archivo")
	}
}

func TestUpload_TooLarge(t *testing.T) {
	dir := t.TempDir()
	s := &UploadServer{Dir: dir, MaxBytes: 10}

	body := bytes.Repeat([]byte("X"), 100_000)
	req := httptest.NewRequest(http.MethodPost, "/upload", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	http.HandlerFunc(s.Upload).ServeHTTP(rr, req)

	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status=%d, want 413", rr.Code)
	}
	ents, _ := os.ReadDir(dir)
	if len(ents) != 0 {
		t.Fatalf("no debe guardar archivos cuando es demasiado grande")
	}
}
