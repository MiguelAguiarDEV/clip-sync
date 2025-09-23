# Roadmap → v1

## 0) Verdes actuales (checkpoint)
- [ ] CLI: enviar texto y archivos grandes (probado manual).
- [ ] Tests E2E: WS básico + upload+señal (UploadAndSignal).
- [ ] Tests negativos WS: size inconsistente, clip vacío.

## 1) Endurecer servidor (core)
- [ ] Validar WS:
  - [ ] `len(Data)==Size`, `Size<=MaxInlineBytes`, `UploadURL` obligatorio si no hay `Data`.
  - [ ] `mime` por defecto cuando falte.
- [ ] Rate limit por dispositivo (p.ej. token bucket simple).
- [ ] Métricas básicas (contadores: clips, drops, conexiones).
- [ ] Logs estructurados (request id, device, user).
- [ ] `/healthz` mejorado (incluye métricas mínimas).

**Aceptación:** tests E2E existentes + nuevos tests: (a) drop por rate-limit, (b) métricas expuestas (unit).

## 2) Robustez WS
- [ ] Keep-alive (ya): ping/pong + timeout.
- [ ] Cierre amable: `http.Server` con `Shutdown`, drenar `Hub`.
- [ ] Deduplicación opcional por `msg_id` (cache LRU corta).
- [ ] Backpressure visible: contador de descartes por device.

**Aceptación:** test integración que mata y reinicia server; clientes CLI se reconectan y siguen recibiendo.

## 3) Seguridad mínima
- [ ] Auth mejorada: token firmado (HMAC) → `userID`, caducidad.
- [ ] Validar `device_id` (formato; opcional: registro previo).
- [ ] Limitar tamaño en `/upload` por config y rechazar tipos prohibidos.

**Aceptación:** tests unit para verificación HMAC + test E2E con token inválido (rechazo).

## 4) Configuración y DX
- [ ] Flags/env para: puertos, límites, dir uploads, log level.
- [ ] `/debug/pprof` y (opcional) `expvar`.
- [ ] Makefile/justfile para `build`, `test`, `run`.

**Aceptación:** arranque vía flags/env; documentación en README.

## 5) Cliente CLI “usable”
- [ ] Reconexión (ya): exponencial.
- [ ] Modo pipe estable: `echo hola | clip-sync --mode send`.
- [ ] Detección simple de mime para `--file` (por extensión).
- [ ] Salidas limpias (logs legibles), códigos de salida coherentes.

**Aceptación:** script E2E que orquesta 2 CLI + server (powershell/bash) y valida salida.

## 6) Observabilidad y pruebas
- [ ] CI (GitHub Actions): lint, `go test ./...`, build server+cli.
- [ ] Benchmarks ligeros del Hub (latencia fan-out, drops a presión).
- [ ] Fuzz JSON del envelope (corpus pequeño).

**Aceptación:** pipeline verde; benchmark no regresa errores y reporta números.

## 7) Pulido “v1”
- [ ] Documentar protocolo (hello, clip inline, clip con `upload_url`).
- [ ] README de arranque rápido (server + 2 CLI).
- [ ] Ejemplos Postman.
- [ ] Changelog v1.

**Listo v1:** servidor estable, CLI funcional (texto+archivo), tests verdes, docs claras.