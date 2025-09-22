# TODO — Clip Sync

## ✅ Hecho
- [x] Núcleo **Hub** (`internal/hub`): `New`, `Join`, `Broadcast` (colas por dispositivo, no bloqueante).
- [x] Tests unitarios Hub:
  - [x] Salta al emisor y entrega a otros.
  - [x] No bloquea con cola llena (descarta).
  - [x] `leave()` elimina y cierra canal.
  - [x] Rooms aislados por `userID`.
- [x] Wiring servidor (`internal/app`):
  - [x] `NewMux()` con `/health` (200 "ok").
  - [x] Ruta `/ws` registrada.
- [x] Handler WebSocket (`internal/ws/handler.go`):
  - [x] `websocket.Accept`, handshake **hello** (token, user, device).
  - [x] `Auth` MVP (token == userID).
  - [x] `Join` al Hub, writer goroutine, reader loop → `Broadcast`.
- [x] **Tipos compartidos** movidos a `pkg/types` (Hello, Clip, Envelope, `MaxInlineBytes`).
- [x] Test de **integración WS** (dos clientes; A envía, B recibe; A no se reenvía).
- [x] Prueba manual con **Postman** de `/ws` y `/health`.

## 🚧 Siguiente(s) inmediato(s)
- [ ] HTTP **/upload** (POST) y **/d/{id}** (GET) en `internal/httpapi`.
- [ ] Test E2E: subir 100KB → señal WS con `upload_url` → otro cliente descarga y verifica tamaño.

## 🛠️ Robustez servidor
- [ ] Validaciones: `size`, `mime`, respetar `MaxInlineBytes` (en WS).
- [ ] **Ping/Pong** WS + timeouts (desconectar zombies).
- [ ] Cierre amable (`http.Server.Shutdown`, drenaje).
- [ ] Métricas de backpressure (drops por dispositivo) + contadores de clips.
- [ ] **Dedup** por `msg_id`/hash (evitar reenvíos).
- [ ] Rate limiting por dispositivo/usuario.
- [ ] Auth mejor (token firmado/HMAC; `device_id` registrado).
- [ ] TLS (reverse proxy o `ListenAndServeTLS`).
- [ ] Logs estructurados + niveles.
- [ ] `/debug/pprof`, `/healthz` listo para prod.
- [ ] Config por flags/env (puertos, límites, dirs).

## 🧪 Calidad
- [ ] Más unit tests (carreras `leave` vs `broadcast`).
- [ ] Fuzz de JSON (envelopes malformados).
- [ ] Benchmarks (latencia de broadcast, presión de colas).
- [ ] CI (GitHub Actions: lint, test, build).

## 👩‍💻 Cliente CLI (validar protocolo)
- [ ] Workspace `go.work` (server + client).
- [ ] Módulo `clients/cli`: comandos `listen` (recibe) y `send --text`.
- [ ] Flags: `--addr`, `--token`, `--device`, `--mode`.
- [ ] Reconexión exponencial.
- [ ] Modo pipe: `echo hola | clip-sync --mode send`.

## 🖥️ GUI Desktop (Fyne)
- [ ] Ventana simple + historial.
- [ ] Watcher de portapapeles (Win/macOS/Linux).
- [ ] Tray icon + autostart.
- [ ] Preferencias (servidor, token, límites).
- [ ] Copiar al clipboard al recibir.
- [ ] Empaquetado: .exe, .app, AppImage.

## 📱 Móvil (fase posterior)
- [ ] UI básica (historial + copiar).
- [ ] Foreground/limitaciones SO (Android/iOS).
- [ ] Empaquetado: APK / Xcode (TestFlight).

## 🚀 Entrega
- [ ] GitHub Releases (binarios Win/macOS/Linux) + checksums.
- [ ] GoReleaser (automatizar).
- [ ] Opcional: Homebrew/Scoop/Chocolatey.

## 📚 Docs
- [ ] README (arranque rápido).
- [ ] Especificación de protocolo (hello, clip inline, clip con `upload_url`).
- [ ] Guía cliente CLI y Postman.
- [ ] Roadmap y contribución.
