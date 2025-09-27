package main

import (
    "bytes"
    "errors"
    "fmt"
    "os/exec"
    "runtime"
)

func getClipboardText() (string, error) {
    if runtime.GOOS == "windows" {
        // Prefer Windows PowerShell; -Raw keeps text as-is.
        cmd := exec.Command("powershell", "-NoProfile", "-Command", "Get-Clipboard -Raw")
        out, err := cmd.CombinedOutput()
        if err != nil {
            return "", fmt.Errorf("Get-Clipboard: %v", err)
        }
        return string(out), nil
    }
    // Wayland wl-paste
    if _, err := exec.LookPath("wl-paste"); err == nil {
        cmd := exec.Command("wl-paste", "-n")
        out, err := cmd.CombinedOutput()
        if err == nil { return string(out), nil }
    }
    // X11 xclip
    if _, err := exec.LookPath("xclip"); err == nil {
        cmd := exec.Command("xclip", "-selection", "clipboard", "-o")
        out, err := cmd.CombinedOutput()
        if err == nil { return string(out), nil }
    }
    // X11 xsel
    if _, err := exec.LookPath("xsel"); err == nil {
        cmd := exec.Command("xsel", "--clipboard", "--output")
        out, err := cmd.CombinedOutput()
        if err == nil { return string(out), nil }
    }
    return "", errors.New("no clipboard backend found (install wl-clipboard or xclip)")
}

func setClipboardText(s string) error {
    if runtime.GOOS == "windows" {
        cmd := exec.Command("powershell", "-NoProfile", "-Command", "Set-Clipboard")
        cmd.Stdin = bytes.NewBufferString(s)
        if out, err := cmd.CombinedOutput(); err != nil {
            return fmt.Errorf("Set-Clipboard: %v (%s)", err, string(out))
        }
        return nil
    }
    if _, err := exec.LookPath("wl-copy"); err == nil {
        cmd := exec.Command("wl-copy")
        cmd.Stdin = bytes.NewBufferString(s)
        if out, err := cmd.CombinedOutput(); err != nil {
            return fmt.Errorf("wl-copy: %v (%s)", err, string(out))
        }
        return nil
    }
    if _, err := exec.LookPath("xclip"); err == nil {
        cmd := exec.Command("xclip", "-selection", "clipboard")
        cmd.Stdin = bytes.NewBufferString(s)
        if out, err := cmd.CombinedOutput(); err != nil {
            return fmt.Errorf("xclip: %v (%s)", err, string(out))
        }
        return nil
    }
    if _, err := exec.LookPath("xsel"); err == nil {
        cmd := exec.Command("xsel", "--clipboard", "--input")
        cmd.Stdin = bytes.NewBufferString(s)
        if out, err := cmd.CombinedOutput(); err != nil {
            return fmt.Errorf("xsel: %v (%s)", err, string(out))
        }
        return nil
    }
    return errors.New("no clipboard backend found (install wl-clipboard or xclip)")
}

