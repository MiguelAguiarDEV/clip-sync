package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCLI_Help(t *testing.T) {
	t.Parallel()

	// Build binary into temp dir
	tmp := t.TempDir()
	bin := filepath.Join(tmp, "cli")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}

	build := exec.Command("go", "build", "-o", bin, ".")
	build.Env = append(os.Environ(), "CGO_ENABLED=0")
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		t.Fatalf("build failed: %v", err)
	}

	// Run with --help
	cmd := exec.Command(bin, "--help")
	out, err := cmd.CombinedOutput()
	// Some CLIs exit non-zero on --help; allow it if output looks like help.
	text := strings.ToLower(string(out))
	if !strings.Contains(text, "usage") && !strings.Contains(text, "help") && !strings.Contains(text, "-h") {
		t.Fatalf("unexpected help output/exit. err=%v out=%q", err, string(out))
	}
}
