# Clip‑Sync Protocol v1

This document describes the WebSocket envelope and HTTP side‑channel used by Clip‑Sync v1.

## Overview

- Session established over WebSocket at `/ws`.
- Client sends a `hello` envelope to authenticate and identify device.
- Clips are delivered as either inline payloads or via an HTTP upload URL.

## Envelopes

Top‑level shape:

```json
{
  "type": "hello|clip",
  "from": "<device_id>",
  "hello": { "token": "...", "user_id": "...", "device_id": "..." },
  "clip": { "msg_id": "...", "mime": "...", "size": 0, "data": "...", "upload_url": "..." }
}
```

### Hello

- `type`: `"hello"`
- `hello.token`: authentication token.
- `hello.user_id`: user identifier.
- `hello.device_id`: unique device id within the user namespace.

Validation:
- `device_id` must match `^[A-Za-z0-9_-]{1,64}$`.
- If HMAC auth is enabled (`CLIPSYNC_HMAC_SECRET`), token must be `userID:exp_unix:hex(hmac_sha256(secret, userID|exp_unix))` and `exp_unix` must be in the future.

### Clip

- `type`: `"clip"`
- `clip.msg_id` (optional): used for deduplication.
- `clip.mime`: MIME type. Defaults to `application/octet-stream` when empty.
- `clip.size`: total size in bytes.
- `clip.data` (optional): inline payload. When present:
  - `len(data) == size`
  - `size <= MaxInlineBytes` (64 KiB by default; see `CLIPSYNC_INLINE_MAXBYTES`).
- `clip.upload_url` (optional): HTTP path (e.g., `/d/<id>`) obtained from `/upload` when the clip is too large to send inline. When `data` is absent, `upload_url` must be present and `size > 0`.

Broadcast:
- The server fans out the clip to all other devices of the same user.
- The `from` field is set to the sender `device_id`.

Backpressure and rate limits:
- Per‑device token bucket controlled by `CLIPSYNC_RATE_LPS`.
- Drops are counted globally and per device.

Deduplication:
- Optional LRU per user controlled by `CLIPSYNC_DEDUPE` capacity (0 disables).
- When a duplicate `msg_id` is detected, the message is dropped.

Client dedupe (recommended):
- Receivers may drop repeated `msg_id` values locally to avoid reapplying the same clip.
- Senders may choose a stable `msg_id` for clipboard-driven events, e.g., `h-<sha|fnv>` of the text, so transient watchers do not flood duplicates.

## HTTP API

### POST /upload

Stores a blob and returns a download URL.

Request:
- Body: raw bytes.
- Header: `Content-Type` validated against whitelist when configured.

Env/flags:
- `CLIPSYNC_UPLOAD_DIR` or `--upload-dir` (default `./uploads`)
- `CLIPSYNC_UPLOAD_MAXBYTES` or `--upload-max-bytes` (default `50MiB`)
- `CLIPSYNC_UPLOAD_ALLOWED` or `--upload-allowed` (comma‑separated MIME list, supports wildcards like `image/*`). Empty disables whitelist.

Response:

```json
{ "upload_url": "/d/<id>", "size": 12345 }
```

Status codes:
- 200 OK: stored.
- 413 Payload Too Large: exceeds `MaxBytes`.
- 415 Unsupported Media Type: MIME not in whitelist.
- 5xx: storage or I/O errors.

### GET /d/{id}

Streams the stored blob with `Content-Type: application/octet-stream`.

### GET /healthz

Returns JSON with basic metrics: `clips_total`, `drops_total`, `conns_current`, and per‑device drops as `drops_device:<user|device>`.

## Server configuration

Flags (all have env equivalents):
- `--addr` (`CLIPSYNC_ADDR`): listen address, default `:8080`.
- `--upload-dir`, `--upload-max-bytes`, `--upload-allowed`.
- `--inline-max-bytes` (`CLIPSYNC_INLINE_MAXBYTES`).
- `--log-level` (`CLIPSYNC_LOG_LEVEL`): `debug|info|error|off`.
- `--pprof` (`CLIPSYNC_PPROF`) and `--expvar` (`CLIPSYNC_EXPVAR`).

Auth:
- MVP: `token == user_id` when `CLIPSYNC_HMAC_SECRET` is unset.
- HMAC mode: token format `userID:exp_unix:hex(hmac_sha256(secret, userID|exp))`.

## CLI behavior

- `listen` mode: reconnects with exponential backoff, resets after success.
- `send` mode:
  - `--text` inline if ≤ MaxInlineBytes.
  - `--file` uploads to `/upload` with MIME auto‑detected by extension when not provided.
  - Stable pipe: when input is piped to stdin, reads up to MaxInlineBytes inline; otherwise spills to a temp file and uploads; MIME heuristic: valid UTF‑8 → `text/plain`, else `application/octet-stream`.
- Exit codes: usage=2, connect=10, upload=11, send=12.

## Limits

- `MaxInlineBytes` default 64 KiB. Change via env/flag.
- Upload server `MaxBytes` default 50 MiB.

## Observability

- `/healthz` JSON metrics.
- Optional `/debug/pprof/*` and `/debug/vars` (expvar) when enabled.
