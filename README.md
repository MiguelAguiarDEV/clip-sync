# clip-sync

Sincroniza tu portapapeles entre dispositivos (Windows/Linux). Server en Go + CLI. WebSocket para señalización y HTTP para cargas grandes.

## Índice
- [Introducción](#introduccion)
- [Demo](#demo)
- [Inicio Rápido](#inicio-rapido)
  - [Servidor en Windows](#servidor-en-windows)
  - [Servidor en Linux](#servidor-en-linux)
- [Clientes](#clientes)
  - [Windows](#clientes-windows)
  - [Linux](#clientes-linux)
- [Releases](#releases)
- [Configuración](#configuracion)
- [Especificaciones Técnicas](#especificaciones-tecnicas)
- [Estructura del Repositorio](#estructura-del-repositorio)
- [Documentación Técnica](#documentacion-tecnica)

<a id="introduccion"></a>
## Introducción
Clip‑Sync permite compartir texto entre dispositivos en la misma red o a través de Internet. Para cargas pequeñas se envía inline, y para cargas grandes se usa `/upload` y se comparte un `upload_url`.

<a id="demo"></a>
## Demo
![Demo](demos/clip-sync-demo.gif)

<a id="inicio-rapido"></a>
## Inicio Rápido

Requisitos:
- Go 1.22+ y Git, o usar los binarios en `bin/`.
- Puerto `8080/tcp` accesible desde los clientes.

<a id="servidor-en-windows"></a>
### Servidor en Windows
1) Ejecutar el servidor (PowerShell):
```powershell
go -C server run ./cmd/server --addr :8080
```
2) Abrir el puerto en el firewall (admin):
```powershell
netsh advfirewall firewall add rule name="clip-sync" dir=in action=allow protocol=TCP localport=8080
```

<a id="servidor-en-linux"></a>
### Servidor en Linux
1) Ejecutar el servidor:
```bash
go -C server run ./cmd/server --addr 0.0.0.0:8080
```
2) Abrir el puerto en el firewall (UFW o equivalente):
```bash
sudo ufw allow 8080/tcp
```

<a id="clientes"></a>
## Clientes
Usa el mismo `--token` para todos los dispositivos del mismo usuario y un `--device` único por equipo.

<a id="clientes-windows"></a>
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

<a id="clientes-linux"></a>
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

<a id="configuracion"></a>
## Configuración
- `--addr` (`CLIPSYNC_ADDR`): dirección de escucha del server (por defecto `:8080`).
- `--inline-max-bytes` (`CLIPSYNC_INLINE_MAXBYTES`): tamaño inline máx. (default 64 KiB).
- `--upload-dir`, `--upload-max-bytes`, `--upload-allowed` (whitelist de MIME, admite comodines `image/*`).
- `--log-level` (`CLIPSYNC_LOG_LEVEL`): `debug|info|error|off`.
- Seguridad opcional (HMAC): define `CLIPSYNC_HMAC_SECRET` en el servidor. Token: `user:exp_unix:hex(hmac_sha256(secret, user|exp))`.
- TLS: usa un reverse proxy (ej. Caddy/Nginx) y conecta por `wss://.../ws`.

<a id="especificaciones-tecnicas"></a>
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

<a id="estructura-del-repositorio"></a>
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

<a id="documentacion-tecnica"></a>
## Documentación Técnica
- Protocolo v1: `docs/protocol.md`
- Notas técnicas Linux/Windows (dev/recruiters): `docs/tutorial-linux-windows.md`
