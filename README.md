# clip-sync

Cross-device clipboard sync. Go server + CLI. WebSocket signaling + HTTP uploads.

## Features

* Sync text and files across your devices.
* Inline clips or upload-by-URL for large payloads.
* Fan-out hub per user. Sender is not echoed.
* Minimal auth hook (MVP).
* Unit + integration tests. GitHub Actions CI.

## Repository layout

```
.
├─ go.work                 # workspace for multi-module repo
├─ server/                 # Go module: server
│  ├─ cmd/server           # server entrypoint
│  ├─ internal/
│  │  ├─ app               # HTTP mux (routes: /health, /ws, /upload, /d/{id})
│  │  ├─ httpapi           # upload/download handlers
│  │  ├─ hub               # pub/sub hub (fan-out)
│  │  └─ ws                # WebSocket handler
│  ├─ pkg/types            # protocol envelopes
│  └─ tests                # integration/E2E tests
└─ clients/cli/            # Go module: CLI (MVP)
```

## Requirements

* Go 1.22+ (toolchain auto OK)
* Git

## Quick start

### Run the server

```bash
# from repo root
go -C server run ./cmd/server
# listens on :8080 (dev)
```

### Smoke test HTTP

```bash
# upload any file
curl -sS -X POST --data-binary @/bin/ls \
  -H "Content-Type: application/octet-stream" \
  http://localhost:8080/upload
# => {"upload_url":"/d/<id>","size":12345}

# download it
curl -sS -o /tmp/clip.bin "http://localhost:8080/d/<id>"
```

### Build binaries

```bash
# server
go -C server build -o ../bin/server ./cmd/server
# cli
go -C clients/cli build -o ../../bin/cli .
```

### Windows PowerShell smoke

```powershell
$bin = "$env:TEMP\clip-sync.bin"; [IO.File]::WriteAllBytes($bin,(New-Object byte[] (100000)))
$resp = Invoke-WebRequest "http://localhost:8080/upload" -Method POST -ContentType application/octet-stream -InFile $bin | ConvertFrom-Json
Invoke-WebRequest ("http://localhost:8080" + $resp.upload_url) -OutFile "$env:TEMP\clip-sync-dl.bin"
```

## User Guide

This section shows how to run the server, find its IP, open the firewall, and connect Linux and Windows clients.

### 1) Run the server

- Dev (from repo root):
  - `go -C server run ./cmd/server --addr 0.0.0.0:8080`
- Built binary:
  - `go -C server build -o ../bin/server ./cmd/server`
  - `./bin/server --addr 0.0.0.0:8080`

### 2) Find the server IP

- LAN IP candidates:
  - `hostname -I`  # Linux (one of these is your LAN IP)
  - `ip route get 1.1.1.1 | sed -n 's/.*src \([^ ]*\).*/\1/p'`  # primary IP
- Public IP (if directly exposed):
  - `curl -s ifconfig.me || curl -s https://api.ipify.org`

Use that IP in all client commands as `<SERVER_IP>`.

### 3) Open the firewall (Linux)

- UFW examples (adjust if needed):
  - `sudo ufw allow 8080/tcp`
  - `sudo ufw status`
- Firewalld examples:
  - `sudo firewall-cmd --add-port=8080/tcp --permanent && sudo firewall-cmd --reload`

If your server is behind a home router, forward external TCP port 8080 to your server’s LAN IP.

### 4) Run clients

Use the same `--token` for devices of the same user and a unique `--device` per device.

- Linux (Wayland/X11) — Sync clipboard both ways on this machine:
  - `go -C clients/cli run . --mode sync --addr ws://<SERVER_IP>:8080/ws --token u1 --device L1 --poll-ms 500`

- Linux — Only receive to clipboard:
  - `go -C clients/cli run . --mode recv --addr ws://<SERVER_IP>:8080/ws --token u1 --device L1`

- Linux — Only watch and broadcast clipboard changes:
  - `go -C clients/cli run . --mode watch --addr ws://<SERVER_IP>:8080/ws --token u1 --device L1 --poll-ms 500`

- Linux — Send a one-off clip:
  - `echo "hola" | go -C clients/cli run . --mode send --addr ws://<SERVER_IP>:8080/ws --token u1 --device L1`
  - `go -C clients/cli run . --mode send --file ./foto.png --mime image/png --addr ws://<SERVER_IP>:8080/ws --token u1 --device L1`

- Windows — Build CLI (from Linux) and run on Windows:
  - Build: `GOOS=windows GOARCH=amd64 go -C clients/cli build -o ../../bin/cli.exe .`
  - On Windows PowerShell:
    - `.\\cli.exe --mode sync --addr ws://<SERVER_IP>:8080/ws --token u1 --device W1`
    - Or receive only: `.\\cli.exe --mode recv --addr ws://<SERVER_IP>:8080/ws --token u1 --device W1`

### 5) Optional security (HMAC tokens and TLS)

- HMAC tokens: set `CLIPSYNC_HMAC_SECRET` on the server. Token format: `user:exp_unix:hex(hmac_sha256(secret, user|exp))`.
  - Example generation (bash):
    - `uid=u1; exp=$(( $(date +%s) + 86400 )); secret=mys3cr3t; payload="$uid|$exp"; mac=$(printf "%s" "$payload" | openssl dgst -sha256 -hmac "$secret" -binary | xxd -p -c 256); echo "$uid:$exp:$mac"`

- TLS/wss via Caddy (example Caddyfile):
  - `:443 {`
  - `  reverse_proxy 127.0.0.1:8080`
  - `}`
  - Then use `wss://your.domain/ws` in clients.

## Protocol

See docs/protocol.md for the full v1 spec.

### WebSocket `/ws`

Envelope:

```json
{
  "type": "hello|clip",
  "hello": { "token": "string", "user_id": "string", "device_id": "string" },
  "clip": {
    "msg_id": "string",
    "mime": "text/plain|application/octet-stream|...",
    "size": 123,
    "data": "base64-optional",
    "upload_url": "/d/<id>-optional"
  }
}
```

Flow:

1. Client connects and sends `hello`.
2. Client sends `clip`:

   * Inline: `data` present and `size <= MaxInlineBytes`.
   * Large: omit `data`, set `upload_url` from HTTP `/upload`, and set `size`.
3. Server broadcasts to all devices of the same user except the sender.

### HTTP

* `GET /health` → `200 ok`
* `POST /upload`
  Body: `application/octet-stream`.
  Response:

  ```json
  { "upload_url": "/d/<id>", "size": 12345 }
  ```
* `GET /d/{id}` → streams the stored blob

## CLI

```
bin/cli --help
```

Modes:
- `listen`: prints incoming clips (debug).
- `send`: sends `--text`, `--file`, o stdin (pipe estable).
- `recv`: aplica clips entrantes al portapapeles (text/*).
- `watch`: monitorea portapapeles local y publica (text/*) con `--poll-ms`.
- `sync`: combina `watch` + `recv` en un solo proceso (recomendado).

Dependencias de portapapeles:
- Windows: PowerShell (Get-Clipboard / Set-Clipboard) — ya incluido.
- Linux: instala `wl-clipboard` (Wayland) o `xclip`/`xsel` (X11).

## Testing

### Server tests

```bash
go -C server test ./... -v
```

### Integration/E2E (in `server/tests`)

* `upload_and_signal_test.go`
* `integration_ws_test.go`
* `integration_ws_too_big_test.go`

### Optional CLI smoke test

If present:

```
clients/cli/main_test.go  # builds the CLI to a temp dir and checks --help output
```

## CI

GitHub Actions runs:

* `go -C server vet ./...`
* `go -C server test ./... -v`
* `go -C clients/cli vet ./...`
* Build server and CLI
* CLI `--help` smoke

## Configuration (dev defaults)

* Upload dir: `./uploads`
* Max upload size: `50 MiB`
* WebSocket auth: MVP stub (`token == userID`)
* Ports and limits via flags/env (WIP)

## Roadmap v1 (excerpt)

* WS validation: `len(data)==size`, `size<=MaxInlineBytes`, `upload_url` required if no `data`.
* Per-device rate limiting and drop counters.
* Keep-alive tuning, graceful shutdown, hub drain.
* HMAC tokens with expiry. Validate `device_id`.
* Flags/env for ports, limits, upload dir, log level.
* Metrics, `/healthz` with stats, `pprof`.
* CLI: pipe mode, reconnection, MIME by extension.
* Docs: protocol spec, Postman examples, changelog.

## Development notes

* No root `go.mod`. Workspace uses `go.work` with `server/` and `clients/cli/`.
* Run `go work sync` from repo root after dependency changes.
* Prefer `go -C <module> ...` or `cd` into the module.

## Make (optional)

If you use GNU Make:

```make
test:
	go test ./... -v
# Use -C or cd to run per module
```

## Contributing

* Open an issue with a minimal repro for bugs.
* Keep PRs small and covered by tests.

## License

MIT (to be added).
Tutoriales
- Español: docs/tutorial-linux-windows.md
