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

## Protocol

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

## CLI (MVP)

```
bin/cli --help
```

Planned:

* `send` mode: stdin or `--file` with MIME by extension.
* Auto reconnect with backoff.
* Clean logs and exit codes.

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
