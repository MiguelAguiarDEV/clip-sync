# Clip‑Sync — Checklist v1

*Actualizado: 2025-09-23*

\[X] Repo inicial y workspace Go configurado
\[X] HTTP mux básico con `/health`
\[X] Hub de difusión (fan‑out) creado
\[X] Servidor WS básico registrado en `/ws`
\[X] Auth MVP (token == userID)
\[X] Keep‑alive WS (ping/pong + timeout)
\[X] Endpoint `/upload` con límite `MaxBytes` y nombre aleatorio
\[X] Tests unitarios de `/upload` en verde
\[X] Ruta `/upload` registrada en el mux
\[ ] Tests E2E: WS básico + upload+señal (UploadAndSignal)
\[ ] Tests negativos WS: size inconsistente, clip vacío
\[ ] Validaciones WS: `len(Data)==Size`, `Size<=MaxInlineBytes`, `UploadURL` si no hay `Data`
\[ ] Asignar `mime` por defecto cuando falte
\[ ] Rate limit por dispositivo
\[ ] Métricas: contadores de clips, drops, conexiones
\[ ] Logs estructurados: request id, device, user
\[ ] `/healthz` con métricas mínimas
\[ ] Cierre amable: `http.Server` con `Shutdown` y drenaje de `Hub`
\[ ] Deduplicación opcional por `msg_id` (LRU corta)
\[ ] Backpressure visible: contador de descartes por device
\[ ] Auth HMAC firmado → `userID` + caducidad
\[ ] Validar `device_id` (formato; opcional registro)
\[ ] `/upload`: permitir solo tipos aceptados
\[ ] Flags/env: puertos, límites, dir uploads, log level
\[ ] `/debug/pprof` y opcional `expvar`
\[ ] Makefile/justfile para build/test/run
\[ ] CLI: reconexión exponencial
\[ ] CLI: modo pipe estable (`echo hola | clip-sync --mode send`)
\[ ] CLI: detección simple de mime por extensión
\[ ] CLI: salida limpia y códigos de salida coherentes
\[ ] CI (GitHub Actions): lint, `go test ./...`, build server+cli
\[ ] Benchmarks ligeros del Hub
\[ ] Fuzzing del envelope JSON (corpus pequeño)
\[ ] Documentar protocolo (hello, clip inline, `upload_url`)
\[ ] README de arranque rápido (server + 2 CLI)
\[ ] Colección Postman
\[ ] Changelog v1
\[ ] CLI: `watch` monitorea portapapeles y publica por WS
\[ ] CLI: `recv` aplica clips entrantes al portapapeles
\[ ] Clips grandes: subir a `/upload` en `watch` y descargar en `recv`
\[ ] Deduplicación por `msg_id` en server y cliente
\[ ] Binarios para Windows, macOS y Linux
