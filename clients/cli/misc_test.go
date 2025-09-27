package main

import (
    "os"
    "os/exec"
    "path/filepath"
    "runtime"
    "testing"
    "time"
)

func TestComputeBackoff(t *testing.T) {
    cases := []struct{ in int; wantMin, wantMax time.Duration }{
        {0, 500 * time.Millisecond, 500 * time.Millisecond},
        {1, 1 * time.Second, 1 * time.Second},
        {2, 2 * time.Second, 2 * time.Second},
        {3, 4 * time.Second, 4 * time.Second},
        {4, 5 * time.Second, 5 * time.Second},
        {10, 5 * time.Second, 5 * time.Second},
    }
    for _, c := range cases {
        got := computeBackoff(c.in)
        if got < c.wantMin || got > c.wantMax {
            t.Fatalf("attempt %d: got %v", c.in, got)
        }
    }
}

func TestDetectMime(t *testing.T) {
    if got := detectMime("/tmp/a.txt", ""); got != "text/plain" { t.Fatalf("txt: %s", got) }
    if got := detectMime("a/b/c.png", ""); got != "image/png" { t.Fatalf("png: %s", got) }
    if got := detectMime("noext", "application/octet-stream"); got != "application/octet-stream" { t.Fatalf("fallback: %s", got) }
}

func TestCLI_UnknownMode_ExitCode(t *testing.T) {
    t.Parallel()
    // build binary
    tmp := t.TempDir()
    bin := filepath.Join(tmp, "cli")
    if runtime.GOOS == "windows" { bin += ".exe" }
    build := exec.Command("go", "build", "-o", bin, ".")
    build.Env = append(os.Environ(), "CGO_ENABLED=0")
    if out, err := build.CombinedOutput(); err != nil {
        t.Fatalf("build: %v out=%s", err, string(out))
    }
    // run with bad mode
    cmd := exec.Command(bin, "--mode", "nope")
    err := cmd.Run()
    if err == nil { t.Fatal("expected non-zero exit") }
}

