# clip-sync

Sincroniza tu portapapeles entre dispositivos (Windows/Linux). Server en Go + CLI. WebSocket para señalización y HTTP para cargas grandes.

![Demo](demos/clip-sync-demo.gif)

## Índice
- Introducción
- Demo
- Inicio Rápido
  - Servidor en Windows
  - Servidor en Linux
- Clientes
  - Windows
  - Linux
- Releases
- Configuración
- Especificaciones Técnicas
- Estructura del Repositorio
- Documentación Técnica

## Introducción
Clip‑Sync permite compartir texto entre dispositivos en la misma red o a través de Internet. Para cargas pequeñas se envía inline, y para cargas grandes se usa `/upload` y se comparte un `upload_url`.

## Demo
Arriba puedes ver un ejemplo de sincronización en tiempo real entre Windows y Linux.

## Inicio Rápido

Requisitos:
- Go 1.22+ y Git, o usar los binarios en `bin/`.
- Puerto `8080/tcp` accesible desde los clientes.

### Servidor en Windows
1) Ejecutar el servidor (PowerShell):
```powershell
go -C server run ./cmd/server --addr :8080
```
2) Abrir el puerto en el firewall (admin):
```powershell
netsh advfirewall firewall add rule name="clip-sync" dir=in action=allow protocol=TCP localport=8080
```

### Servidor en Linux
1) Ejecutar el servidor:
```bash
go -C server run ./cmd/server --addr 0.0.0.0:8080
```
2) Abrir el puerto en el firewall (UFW o equivalente):
```bash
sudo ufw allow 8080/tcp
```

## Clientes
Usa el mismo `--token` para todos los dispositivos del mismo usuario y un `--device` único por equipo.

### Windows
- Sincronización bidireccional y logs verbosos:
```powershell
.\bin\cli.exe --mode sync --addr ws://<SERVER_IP>:8080/ws --token u1 --device W1 -v
```
- Solo recibir y aplicar al portapapeles:
```powershell
.\bin\cli.exe --mode recv --addr ws://<SERVER_IP>:8080/ws --token u1 --device W1 -v
```
- Enviar una vez (texto o archivo):
```powershell
.\bin\cli.exe --mode send --text "hola" --addr ws://<SERVER_IP>:8080/ws --token u1 --device W1
.\bin\cli.exe --mode send --file .\foto.png --mime image/png --addr ws://<SERVER_IP>:8080/ws --token u1 --device W1
```

### Linux
- Sincronización bidireccional y logs verbosos:
```bash
./bin/cli --mode sync --addr ws://<SERVER_IP>:8080/ws --token u1 --device L1 -v
```
- Solo recibir y aplicar al portapapeles:
```bash
./bin/cli --mode recv --addr ws://<SERVER_IP>:8080/ws --token u1 --device L1 -v
```
- Enviar una vez (texto o archivo):
```bash
echo "hola" | ./bin/cli --mode send --addr ws://<SERVER_IP>:8080/ws --token u1 --device L1
./bin/cli --mode send --file ./foto.png --mime image/png --addr ws://<SERVER_IP>:8080/ws --token u1 --device L1
```

## Releases
- Tags `v*` (ej. `v0.1.0`) disparan el workflow de publicación y adjuntan binarios para Linux/Windows/macOS (amd64/arm64):
  - `server_<GOOS>_<GOARCH>[.exe]`
  - `cli_<GOOS>_<GOARCH>[.exe]`
- Cómo crear un release:
  - `git tag v0.1.0 && git push origin v0.1.0`
  - Revisa la pestaña “Releases” para descargar activos.
- Instalación rápida (ejemplos):
  - Linux: `chmod +x cli_linux_amd64 && sudo mv cli_linux_amd64 /usr/local/bin/clip-sync`
  - Windows: renombra `cli_windows_amd64.exe` a `cli.exe` y colócalo en una carpeta del `PATH`.

## Configuración
- `--addr` (`CLIPSYNC_ADDR`): dirección de escucha del server (por defecto `:8080`).
- `--inline-max-bytes` (`CLIPSYNC_INLINE_MAXBYTES`): tamaño inline máx. (default 64 KiB).
- `--upload-dir`, `--upload-max-bytes`, `--upload-allowed` (whitelist de MIME, admite comodines `image/*`).
- `--log-level` (`CLIPSYNC_LOG_LEVEL`): `debug|info|error|off`.
- Seguridad opcional (HMAC): define `CLIPSYNC_HMAC_SECRET` en el servidor. Token: `user:exp_unix:hex(hmac_sha256(secret, user|exp))`.
- TLS: usa un reverse proxy (ej. Caddy/Nginx) y conecta por `wss://.../ws`.

## Especificaciones Técnicas
- Stack: Go 1.22, WebSocket (`/ws`) + HTTP (`/upload`, `/d/{id}`, `/health`, `/healthz`).
- Arquitectura: hub por usuario, fan‑out a dispositivos, no se hace echo al emisor.
- Escalabilidad: backoff exponencial en clientes, dedupe por `msg_id` en server y cliente.
- Portapapeles:
  - Windows: `clip.exe` o PowerShell (`Get-Clipboard`/`Set-Clipboard`).
  - Linux: Wayland `wl-clipboard` o X11 `xclip`/`xsel`.
- Límite inline: 64 KiB; cargas grandes via `/upload` (50 MiB por defecto).
- Observabilidad: `/health` simple, `/healthz` JSON de métricas, `/debug/pprof/*` y `/debug/vars` opcionales.
- Calidad: tests unitarios e integración; CI en GitHub Actions.

## Estructura del Repositorio
```
.
├─ go.work                 # workspace multi-módulo
├─ server/                 # Módulo Go: server
│  ├─ cmd/server           # entrypoint
│  ├─ internal/
│  │  ├─ app               # HTTP mux (routes: /health, /healthz, /ws, /upload, /d/{id})
│  │  ├─ httpapi           # upload/download
│  │  ├─ hub               # pub/sub (fan-out)
│  │  └─ ws                # WebSocket handler
│  ├─ pkg/types            # envelopes
│  └─ tests                # integración/E2E
└─ clients/cli/            # Módulo Go: CLI
```

## Documentación Técnica
- Protocolo v1: `docs/protocol.md`
- Notas técnicas Linux/Windows (dev/recruiters): `docs/tutorial-linux-windows.md`
