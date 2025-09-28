# Technical Notes: Linux and Windows

Audience: developers and recruiters. This folder is not for end‑user step‑by‑step guides.

## Table of Contents
- [Build and Cross‑Compile](#build-cross)
- [Run (Development)](#run-dev)
- [Relevant Flags/Env (server)](#flags-env)
- [Clipboard Integration](#clipboard)
- [Network and Firewall](#network)
- [Observability](#observability)
- [Security](#security)
- [CLI Modes](#cli-modes)
- [Limits and Dedupe](#limits-dedupe)
- [Tests and CI](#tests-ci)
- [Troubleshooting (technical)](#troubleshooting)
- [See Also](#see-also)

<a id="build-cross"></a>
## Build and Cross‑Compile
- CLI
  - Linux: `go -C clients/cli build -o ../../bin/cli .`
  - Windows (from Linux/macOS): `GOOS=windows GOARCH=amd64 go -C clients/cli build -o ../../bin/cli.exe .`
- Server
  - Linux: `go -C server build -o ../bin/server ./cmd/server`
  - Windows (cross): `GOOS=windows GOARCH=amd64 go -C server build -o ../bin/server.exe ./cmd/server`

Notes:
- The repo uses `go.work` for a multi‑module workspace (`server/`, `clients/cli/`).
- Prefer `go -C <module> ...` or `cd` into each module.

<a id="run-dev"></a>
## Run (Development)
- Server (bind all): `go -C server run ./cmd/server --addr :8080`
- CLI Windows: `.\bin\cli.exe --mode sync --addr ws://<IP>:8080/ws --token u1 --device W1`
- CLI Linux: `./bin/cli --mode sync --addr ws://<IP>:8080/ws --token u1 --device L1`

<a id="flags-env"></a>
## Relevant Flags/Env (server)
- `--addr` (`CLIPSYNC_ADDR`): default `:8080`.
- `--upload-dir` (`CLIPSYNC_UPLOAD_DIR`): default `./uploads`.
- `--upload-max-bytes` (`CLIPSYNC_UPLOAD_MAXBYTES`): default `50 MiB`.
- `--upload-allowed` (`CLIPSYNC_UPLOAD_ALLOWED`): MIME list (e.g., `text/plain,image/*`). Empty disables whitelist.
- `--inline-max-bytes` (`CLIPSYNC_INLINE_MAXBYTES`): default `64 KiB`.
- `--log-level` (`CLIPSYNC_LOG_LEVEL`): `debug|info|error|off`.
- `--pprof` (`CLIPSYNC_PPROF`) and `--expvar` (`CLIPSYNC_EXPVAR`): debug endpoints.

<a id="clipboard"></a>
## Clipboard Integration
- Implementation in `clients/cli/clipboard.go`:
  - Windows: uses `clip.exe` when available; otherwise PowerShell `Get-Clipboard`/`Set-Clipboard`.
  - Linux: prefers `wl-copy`/`wl-paste` (Wayland), then `xclip` or `xsel` (X11).
- Common error: if no backend found, CLI returns `no clipboard backend found`.

<a id="network"></a>
## Network and Firewall
- Service port: `8080/tcp`.
- Windows: `netsh advfirewall firewall add rule name="clip-sync" dir=in action=allow protocol=TCP localport=8080`.
- Linux (UFW): `sudo ufw allow 8080/tcp`.

<a id="observability"></a>
## Observability
- `/health`: simple `200 ok`.
- `/healthz`: JSON metrics (clips, drops, connections, per-device drops).
- Optional: `/debug/pprof/*` and `/debug/vars` (expvar) when enabled.

<a id="security"></a>
## Security
- MVP: if `CLIPSYNC_HMAC_SECRET` is unset, `token == user_id`.
- HMAC: token `user:exp_unix:hex(hmac_sha256(secret, user|exp))` (validated by server).
- TLS: use a reverse proxy (Caddy/Nginx) and connect via `wss://.../ws`.

<a id="cli-modes"></a>
## CLI Modes
- `listen`: print incoming clips (debug).
- `send`: send `--text`, `--file`, or stdin (stable pipe). MIME by extension if `--mime` is not provided.
- `recv`: apply incoming `text/*` to clipboard.
- `watch`: publish local clipboard changes (`--poll-ms`).
- `sync`: `watch` + `recv` in one process.

<a id="limits-dedupe"></a>
## Limits and Dedupe
- Inline: `CLIPSYNC_INLINE_MAXBYTES` (64 KiB default).
- Upload: `CLIPSYNC_UPLOAD_MAXBYTES` (50 MiB default).
- Dedup by `msg_id` on server and client (per-user LRU, `CLIPSYNC_DEDUPE`).

<a id="tests-ci"></a>
## Tests and CI
- Unit/integration tests under `server/tests` and `clients/cli`.
- GitHub Actions: vet, build, test, and CLI `--help` smoke.

<a id="troubleshooting"></a>
## Troubleshooting (technical)
- WS connection: verify IP/port and that the client uses `ws://<IP>:8080/ws`.
- 415 on `/upload`: MIME not allowed; adjust `--upload-allowed`.
- 413 on `/upload`: exceeds `--upload-max-bytes`.
- Missing clipboard backends on Linux: install `wl-clipboard` (Wayland) or `xclip`/`xsel` (X11).

<a id="see-also"></a>
## See Also
- Protocol v1: [protocol.md](protocol.md)
- README (overview and demo): [../README.md](../README.md)

