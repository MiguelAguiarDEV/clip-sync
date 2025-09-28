# clip-sync

Clip-Sync is a lightweight CLI to sync your clipboard across Windows and Linux, on the same LAN or over the Internet. Small payloads go inline via /ws; large payloads use /upload and are shared by an upload_url. Each user gets a hub with fan-out to devices and no echo to the sender. Built in Go with WebSocket signaling and HTTP for large transfers.

## Table of Contents

* [Introduction](#introduction)
* [Demo](#demo)
* [Quick Start](#quick-start)
  * [Server on Windows](#server-on-windows)
  * [Server on Linux](#server-on-linux)
* [Clients](#clients)
  * [Windows](#windows)
  * [Linux](#linux)
* [Releases](#releases)
* [Configuration](#configuration)
* [Technical Specs](#technical-specs)
* [Repository Layout](#repository-layout)
* [Technical Docs](#technical-docs)

<a id="introduction"></a>

## Introduction

Clip-Sync is a lightweight CLI to sync your clipboard across Windows and Linux, on the same LAN or over the Internet. Small payloads go inline via /ws; large payloads use /upload and are shared by an upload_url. Each user gets a hub with fan-out to devices and no echo to the sender. Built in Go with WebSocket signaling and HTTP for large transfers.

<a id="demo"></a>

## Demo

![Demo](demos/clip-sync-demo.gif)

<a id="quick-start"></a>

## Quick Start

Requirements:

* Go 1.22+ and Git, or use prebuilt binaries.
* TCP port `8080` reachable from clients.

<a id="server-on-windows"></a>

### Server on Windows

1) Run the server (PowerShell):

```powershell
go -C server run ./cmd/server --addr :8080
```

2) Open the firewall (admin):

```powershell
netsh advfirewall firewall add rule name="clip-sync" dir=in action=allow protocol=TCP localport=8080
```

<a id="server-on-linux"></a>

### Server on Linux

1) Run the server:

```bash
go -C server run ./cmd/server --addr 0.0.0.0:8080
```

2) Open the port (UFW or equivalent):

```bash
sudo ufw allow 8080/tcp
```

<a id="clients"></a>

## Clients

Use the same `--token` for devices of the same user and a unique `--device` per machine.

<a id="windows"></a>

### Windows

* Bidirectional sync with verbose logs:

```powershell
.\bin\cli.exe --mode sync --addr ws://<SERVER_IP>:8080/ws --token u1 --device W1 -v
```

* Receive only and apply to clipboard:

```powershell
.\bin\cli.exe --mode recv --addr ws://<SERVER_IP>:8080/ws --token u1 --device W1 -v
```

* One‑shot send (text or file):

```powershell
.\bin\cli.exe --mode send --text "hello" --addr ws://<SERVER_IP>:8080/ws --token u1 --device W1
.\bin\cli.exe --mode send --file .\photo.png --mime image/png --addr ws://<SERVER_IP>:8080/ws --token u1 --device W1
```

<a id="linux"></a>

### Linux

* Bidirectional sync with verbose logs:

```bash
./bin/cli --mode sync --addr ws://<SERVER_IP>:8080/ws --token u1 --device L1 -v
```

* Receive only and apply to clipboard:

```bash
./bin/cli --mode recv --addr ws://<SERVER_IP>:8080/ws --token u1 --device L1 -v
```

* One‑shot send (text or file):

```bash
echo "hello" | ./bin/cli --mode send --addr ws://<SERVER_IP>:8080/ws --token u1 --device L1
./bin/cli --mode send --file ./photo.png --mime image/png --addr ws://<SERVER_IP>:8080/ws --token u1 --device L1
```

<a id="releases"></a>

## Releases

Prebuilt artifacts are published under **Releases**. Verify checksums when provided. You can also build locally using `make dist` or the scripts under `scripts/`.

<a id="configuration"></a>

## Configuration

* `--addr` (`CLIPSYNC_ADDR`): listen address (default `:8080`).
* `--inline-max-bytes` (`CLIPSYNC_INLINE_MAXBYTES`): inline size limit (default 64 KiB).
* `--upload-dir`, `--upload-max-bytes`, `--upload-allowed`: upload directory, max size, allowed MIME whitelist (supports wildcards like `image/*`).
* `--log-level` (`CLIPSYNC_LOG_LEVEL`): `debug|info|error|off`.
* Optional security (HMAC): set `CLIPSYNC_HMAC_SECRET`. Token: `user:exp_unix:hex(hmac_sha256(secret, user|exp))`.
* TLS: use a reverse proxy (e.g., Caddy/Nginx) and connect via `wss://.../ws`.

<a id="technical-specs"></a>

## Technical Specs

* Stack: Go 1.22, WebSocket (`/ws`) + HTTP (`/upload`, `/d/{id}`, `/health`, `/healthz`).
* Architecture: per‑user hub with fan‑out to devices; no echo to sender.
* Scalability: client exponential backoff; dedup by `msg_id` on server and client.
* Clipboard backends:
  * Windows: `clip.exe` or PowerShell (`Get-Clipboard` / `Set-Clipboard`).
  * Linux: Wayland `wl-clipboard` or X11 `xclip` / `xsel`.
* Inline limit: 64 KiB. Large payloads via `/upload` (50 MiB default).
* Observability: `/health` liveness, `/healthz` JSON metrics, optional `/debug/pprof/*` and `/debug/vars`.
* Quality: unit + integration tests; GitHub Actions CI.

<a id="repository-layout"></a>

## Repository Layout

```
.
├─ go.work                 # multi-module workspace
├─ server/                 # Go module: server
│  ├─ cmd/server           # entrypoint
│  ├─ internal/
│  │  ├─ app               # HTTP mux (routes: /health, /healthz, /ws, /upload, /d/{id})
│  │  ├─ httpapi           # upload/download
│  │  ├─ hub               # pub/sub (fan-out)
│  │  └─ ws                # WebSocket handler
│  ├─ pkg/types            # envelopes
│  └─ tests                # integration/E2E
└─ clients/cli/            # Go module: CLI
```

<a id="technical-docs"></a>

## Technical Docs

* Protocol v1: `docs/protocol.md`
* Linux/Windows technical notes (dev/recruiters): `docs/tutorial-linux-windows.md`

