# Tutorial: Clip‑Sync en Linux y Windows (misma red)

Este tutorial te guía para usar Clip‑Sync con el servidor hospedado en Linux o en Windows, y clientes en ambos sistemas. Está pensado para un entorno de red local doméstica/oficina.

Requisitos previos
- Red local: ambos equipos deben poder verse por IP (ej. 192.168.x.y).
- Puerto 8080 abierto en el host del servidor.
- Binarios o Go instalado:
  - Opción A (recomendado): descarga binarios desde la pestaña Releases (server y clip-sync para tu SO/arquitectura).
  - Opción B: compila desde el código con Go ≥ 1.22.
- Linux (cliente): necesitas un backend de portapapeles.
  - Wayland: wl-clipboard (`wl-copy`, `wl-paste`).
  - X11: xclip o xsel.
- Windows (cliente): PowerShell viene preinstalado y se usa para el portapapeles.

Variables y flags útiles
- `--addr` (server): dirección de escucha, por defecto `:8080`.
- `CLIPSYNC_UPLOAD_ALLOWED` o `--upload-allowed`: whitelist de MIME, por ej. `text/plain,image/*`.
- `CLIPSYNC_INLINE_MAXBYTES`: tamaño máximo de clip inline (por defecto 64 KiB).
- `CLIPSYNC_UPLOAD_MAXBYTES`: máximo de `/upload` (por defecto 50 MiB).
- `CLIPSYNC_LOG_LEVEL`: `debug|info|error|off`.
- Auth opcional HMAC: `CLIPSYNC_HMAC_SECRET` (ver más abajo).

## Escenario A — Servidor en Linux
1) Inicia el servidor
- Con binario:
  ```bash
  ./server-linux-amd64 -addr :8080 --pprof --expvar \
    --upload-allowed "text/plain,image/*" --inline-max-bytes 65536
  ```
- O con Go (desde repo):
  ```bash
  go -C server run ./cmd/server
  ```
- Abre el puerto en el firewall si usas UFW:
  ```bash
  sudo ufw allow 8080/tcp
  ```

2) Obtén la IP del servidor Linux
```bash
ip a | rg 'inet 192|inet 10'
# Ejemplo: 192.168.1.50
```

3) Cliente Windows (sin instalar nada extra)
- Descarga `clip-sync-windows-amd64.exe`.
- Ejecuta en PowerShell:
  ```powershell
  .\clip-sync-windows-amd64.exe -addr ws://192.168.1.50:8080/ws -token u1 -device WIN1 -mode sync
  ```
  - `-token` puede ser cualquier string en modo MVP. Con HMAC, ver “Auth HMAC”.
  - `-device` debe cumplir `^[A-Za-z0-9_-]{1,64}$` (ej. WIN1, LNX1).

4) Cliente Linux
- Instala backend del portapapeles (Wayland):
  ```bash
  sudo apt install wl-clipboard # o equivalente en tu distro
  ```
  (X11: `sudo apt install xclip`)
- Ejecuta:
  ```bash
  ./clip-sync-linux-amd64 -addr ws://192.168.1.50:8080/ws -token u1 -device LNX1 -mode sync
  ```

5) Prueba
- Copia texto en Windows → debe aparecer en Linux, y viceversa.
- Clips grandes de texto: el CLI hará upload automático al servidor.

## Escenario B — Servidor en Windows
1) Inicia el servidor
- Con binario:
  ```powershell
  .\server-windows-amd64.exe -addr :8080 --upload-allowed "text/plain,image/*"
  ```
- Abre el puerto en el firewall (PowerShell admin):
  ```powershell
  netsh advfirewall firewall add rule name="ClipSyncServer" dir=in action=allow protocol=TCP localport=8080
  ```

2) Obtén la IP de Windows
```powershell
ipconfig | findstr IPv4
# Ejemplo: 192.168.1.60
```

3) Cliente Linux
```bash
./clip-sync-linux-amd64 -addr ws://192.168.1.60:8080/ws -token u1 -device LNX1 -mode sync
```

4) Cliente Windows
```powershell
.\clip-sync-windows-amd64.exe -addr ws://192.168.1.60:8080/ws -token u1 -device WIN1 -mode sync
```

## Auth HMAC opcional (recomendado en redes compartidas)
1) Define un secreto en el servidor (Linux/Windows):
```bash
export CLIPSYNC_HMAC_SECRET=mi_super_secreto
# Windows (PowerShell): $env:CLIPSYNC_HMAC_SECRET = "mi_super_secreto"
```
2) Genera un token con expiración (ej. +1h)
- Linux/macOS (bash + openssl):
  ```bash
  uid="u1"; exp=$(date -d "+1 hour" +%s); secret="mi_super_secreto"
  mac=$(printf "%s|%s" "$uid" "$exp" | openssl dgst -sha256 -hmac "$secret" -binary | xxd -p -c 256)
  echo "$uid:$exp:$mac"
  ```
- Windows (PowerShell):
  ```powershell
  $uid = "u1"; $exp = [int][double]::Parse((Get-Date).ToUniversalTime().AddHours(1).Subtract([datetime]'1970-01-01').TotalSeconds)
  $secret = "mi_super_secreto"
  $payload = "$uid|$exp"
  $hmac = [System.BitConverter]::ToString((New-Object System.Security.Cryptography.HMACSHA256 ([Text.Encoding]::UTF8.GetBytes($secret))).ComputeHash([Text.Encoding]::UTF8.GetBytes($payload))).Replace("-", "").ToLower()
  "$uid:$exp:$hmac"
  ```
3) Usa ese token en los clientes (`-token`), y deja definido el `CLIPSYNC_HMAC_SECRET` en el entorno del servidor.

## Consejos de uso del CLI
- Modos:
  - `sync`: recomendado; combina recv + watch.
  - `recv`: solo aplica clips entrantes al portapapeles.
  - `watch`: solo publica cuando tu portapapeles cambia.
- Detección de MIME por extensión al enviar `--file`.
- Pipe estable: `echo hola | ./clip-sync -mode send`.
- Códigos de salida: uso=2, conexión=10, upload=11, envío=12.

## Solución de problemas
- “connect failed”: verifica la IP del servidor, el puerto 8080/tcp en el firewall y que usas `ws://IP:8080/ws` (no `localhost` desde otra máquina).
- “no clipboard backend found” (Linux): instala `wl-clipboard` (Wayland) o `xclip`/`xsel` (X11).
- “unsupported media type” en `/upload`: revisa `CLIPSYNC_UPLOAD_ALLOWED`. Para pruebas, `text/plain,image/*` suele ser suficiente.
- `device_id` inválido: usa alfanumérico, `_` o `-` (máx. 64 chars), por ejemplo `WIN1` o `LNX1`.
- “payload too large”: ajusta `--inline-max-bytes` para inline y `--upload-max-bytes` para `/upload`.

## (Avanzado) Ejecutar como servicio
- Linux systemd (servidor)
  ```ini
  [Unit]
  Description=Clip‑Sync Server
  After=network.target

  [Service]
  ExecStart=/usr/local/bin/server -addr :8080
  Environment=CLIPSYNC_UPLOAD_ALLOWED=text/plain,image/*
  Restart=on-failure

  [Install]
  WantedBy=multi-user.target
  ```
- Windows: crear una Tarea Programada que ejecute el binario al iniciar sesión, o usar NSSM para instalar como servicio.

¡Listo! Con esto puedes sincronizar texto entre tu Windows y tu Linux en la misma red.

