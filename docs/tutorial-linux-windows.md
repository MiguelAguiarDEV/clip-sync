# Notas técnicas: Linux y Windows

Audiencia: desarrolladores y recruiters. En esta carpeta no hay guías paso a paso para usuarios finales.

## Índice
- [Compilación y cross‑compile](#compilacion-cross)
- [Ejecución (desarrollo)](#ejecucion-desarrollo)
- [Flags/env relevantes (server)](#flags-env-server)
- [Integración con portapapeles](#integracion-portapapeles)
- [Red y firewall](#red-firewall)
- [Observabilidad](#observabilidad)
- [Seguridad](#seguridad)
- [Modos del CLI](#modos-cli)
- [Límites y dedupe](#limites-dedupe)
- [Pruebas y CI](#pruebas-ci)
- [Troubleshooting (enfoque técnico)](#troubleshooting)
- [Ver también](#ver-tambien)

<a id="compilacion-cross"></a>
## Compilación y cross‑compile
- CLI
  - Linux: `go -C clients/cli build -o ../../bin/cli .`
  - Windows (desde Linux/macOS): `GOOS=windows GOARCH=amd64 go -C clients/cli build -o ../../bin/cli.exe .`
- Server
  - Linux: `go -C server build -o ../bin/server ./cmd/server`
  - Windows (cross): `GOOS=windows GOARCH=amd64 go -C server build -o ../bin/server.exe ./cmd/server`

Notas:
- El repo usa `go.work` para un workspace multi‑módulo (`server/`, `clients/cli/`).
- Preferir `go -C <mod> ...` o `cd` dentro de cada módulo.

<a id="ejecucion-desarrollo"></a>
## Ejecución (desarrollo)
- Server (bind all): `go -C server run ./cmd/server --addr :8080`
- CLI Windows: `.\bin\cli.exe --mode sync --addr ws://<IP>:8080/ws --token u1 --device W1`
- CLI Linux: `./bin/cli --mode sync --addr ws://<IP>:8080/ws --token u1 --device L1`

<a id="flags-env-server"></a>
## Flags/env relevantes (server)
- `--addr` (`CLIPSYNC_ADDR`): por defecto `:8080`.
- `--upload-dir` (`CLIPSYNC_UPLOAD_DIR`): por defecto `./uploads`.
- `--upload-max-bytes` (`CLIPSYNC_UPLOAD_MAXBYTES`): por defecto `50 MiB`.
- `--upload-allowed` (`CLIPSYNC_UPLOAD_ALLOWED`): lista de MIME (ej. `text/plain,image/*`). Vacío desactiva whitelist.
- `--inline-max-bytes` (`CLIPSYNC_INLINE_MAXBYTES`): por defecto `64 KiB`.
- `--log-level` (`CLIPSYNC_LOG_LEVEL`): `debug|info|error|off`.
- `--pprof` (`CLIPSYNC_PPROF`) y `--expvar` (`CLIPSYNC_EXPVAR`): endpoints de debug.

<a id="integracion-portapapeles"></a>
## Integración con portapapeles
- Implementación en `clients/cli/clipboard.go`:
  - Windows: usa `clip.exe` si está disponible; si no, PowerShell `Get-Clipboard`/`Set-Clipboard`.
  - Linux: busca `wl-copy`/`wl-paste` (Wayland), luego `xclip` o `xsel` (X11).
- Errores comunes: si no hay backend, el CLI devuelve `no clipboard backend found`.

<a id="red-firewall"></a>
## Red y firewall
- Puerto de servicio: `8080/tcp`.
- Windows: `netsh advfirewall firewall add rule name="clip-sync" dir=in action=allow protocol=TCP localport=8080`.
- Linux (UFW): `sudo ufw allow 8080/tcp`.

<a id="observabilidad"></a>
## Observabilidad
- `/health`: `200 ok` simple.
- `/healthz`: JSON con métricas (clips, drops, conexiones, drops por dispositivo).
- Opcional: `/debug/pprof/*` y `/debug/vars` (expvar) cuando se habilitan.

<a id="seguridad"></a>
## Seguridad
- MVP: si `CLIPSYNC_HMAC_SECRET` no está definido, `token == user_id`.
- HMAC: token `user:exp_unix:hex(hmac_sha256(secret, user|exp))` (validado en server).
- TLS: usar reverse proxy (Caddy/Nginx) y clientes por `wss://.../ws`.

<a id="modos-cli"></a>
## Modos del CLI
- `listen`: imprime clips entrantes (debug).
- `send`: envía `--text`, `--file` o `stdin` (pipe estable). MIME por extensión si no se pasa `--mime`.
- `recv`: aplica clips `text/*` al portapapeles.
- `watch`: publica cambios del portapapeles local (`--poll-ms`).
- `sync`: `watch` + `recv` en un solo proceso.

<a id="limites-dedupe"></a>
## Límites y dedupe
- Inline: `CLIPSYNC_INLINE_MAXBYTES` (64 KiB por defecto).
- Upload: `CLIPSYNC_UPLOAD_MAXBYTES` (50 MiB por defecto).
- Dedupe por `msg_id` en server y cliente (LRU configurable por usuario, `CLIPSYNC_DEDUPE`).

<a id="pruebas-ci"></a>
## Pruebas y CI
- Unit/integration tests en `server/tests` y `clients/cli`.
- GitHub Actions: vet, build, test y smoke del CLI `--help`.

<a id="troubleshooting"></a>
## Troubleshooting (enfoque técnico)
- Conexión WS: revisar IP/puerto, y que el cliente use `ws://<IP>:8080/ws`.
- 415 en `/upload`: MIME no permitido; ajustar `--upload-allowed`.
- 413 en `/upload`: excede `--upload-max-bytes`.
- Backends de clipboard ausentes en Linux: instalar `wl-clipboard` (Wayland) o `xclip`/`xsel` (X11).

<a id="ver-tambien"></a>
## Ver también
- Protocolo v1: [docs/protocol.md](protocol.md)
- README (visión general y demo): [../README.md](../README.md)


