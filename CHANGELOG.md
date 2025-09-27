# Changelog

## v1.0.0 (2025-09-27)

- Server
  - WebSocket `/ws` with hello+clip envelopes
  - Fan-out hub per user, no echo to sender
  - Validations: `len(data)==size`, inline size limit, `upload_url` required for large clips
  - Rate limit per device (token bucket) with drop counters and metrics
  - Optional deduplication by `msg_id` (LRU per user)
  - HTTP `/upload` with max bytes and random IDs; `/d/{id}` download
  - MIME whitelist for `/upload` (exact and wildcard `type/*`)
  - `/healthz` metrics, optional `/debug/pprof` and `/debug/vars`
  - Structured logs and graceful shutdown
  - HMAC auth support (token `user:exp:mac`), fallback token==user
  - Flags/env for ports, limits, dirs, log level
- CLI
  - `listen` mode with exponential reconnection backoff
  - `send` mode: `--text` or `--file`; auto-detect MIME by extension
  - Stable pipe mode: stdin inline or auto-upload via temp file; UTF-8 → text/plain
  - Clean stderr output and explicit exit codes
- Tests & CI
  - Unit + integration tests; GitHub Actions vet/test/build
  - Hub fan-out benchmarks; Envelope JSON fuzz test
- Docs
  - Protocol spec, Quickstart, Postman collection

Thanks to all contributors.
