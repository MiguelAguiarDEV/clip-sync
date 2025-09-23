# Clip‑Sync

Sincroniza tu portapapeles y archivos ligeros entre dispositivos usando WebSockets y HTTP.

> Estado: MVP en desarrollo. `/upload` y tests unitarios listos. WS básico operativo. Ver `TODO.md`.

---

## Índice

* [Arquitectura](#arquitectura)
* [Protocolo](#protocolo)
* [Instalación](#instalación)
* [Configuración](#configuración)
* [Ejecución](#ejecución)
* [Endpoints HTTP](#endpoints-http)
* [WebSocket](#websocket)
* [Cliente CLI](#cliente-cli)
* [Seguridad](#seguridad)
* [Pruebas](#pruebas)
* [Roadmap](#roadmap)
* [Licencia](#licencia)

---

## Arquitectura

```
┌────────────┐      HTTP /upload      ┌───────────────┐
│  Cliente A │ ─────────────────────▶ │  Servidor     │
│  (CLI/App) │                        │  HTTP + WS    │
└────────────┘                        │  Hub (fanout) │
     ▲     │        WS /ws            └───────────────┘
     │     └───────────────────────────────────▶  ▲
     │                                           │
┌────────────┐                                   │
│  Cliente B │ ◀──────────────────────────────────┘
└────────────┘
```

* **Hub**: difunde cada *clip* a los dispositivos del mismo usuario.
* **/ws**: canal en tiempo real para clips en línea y señales de subida.
* **/upload**: subida de blobs cuando superan `MaxInlineBytes`.
* **Auth MVP**: token en query que hoy equivale a `userID`.

Estructura de carpetas relevante:

```
server/
  internal/
    app/        # mux, wiring
    hub/        # fan-out
    ws/         # servidor WS
    httpapi/    # handlers HTTP (/upload)
cli/
  ...           # cliente CLI (WIP)
```

## Protocolo

### Envelope WS (JSON)

```json
{
  "type": "clip",              // "hello" | "clip" | "ack"
  "user_id": "u123",           // derivado del token
  "device_id": "dev-a",
  "msg_id": "uuid-...",        // opcional; útil para dedup
  "mime": "text/plain",        // por defecto si falta
  "size": 5,                    // bytes del contenido
  "data": "aGVsbG8=",          // base64 si inline
  "upload_url": "http://..."   // si no va inline
}
```

Reglas previstas:

* `len(data)==size` si `data` existe.
* `size <= MaxInlineBytes` para enviar inline; si no, usar `upload_url`.

## Instalación

Requisitos: Go ≥ 1.22.

```bash
git clone <repo>
cd clip-sync
go mod download
```

## Configuración

Por flags o variables de entorno (WIP):

* `PORT` (HTTP)
* `WS_PORT` (opcional si se separa)
* `UPLOAD_DIR` (por defecto `./uploads`)
* `UPLOAD_MAX_BYTES` (por defecto `8MiB`)
* `LOG_LEVEL` (`debug|info|warn|error`)
* `MAX_INLINE_BYTES` (límite para `data` en WS)

## Ejecución

Servidor en local:

```bash
# Desde /server
go run ./cmd/server
```

Comprobación de salud:

```bash
curl -i http://localhost:8080/health
```

## Endpoints HTTP

### `POST /upload`

Sube un archivo multipart con el campo `file`.

```bash
curl -F "file=@./ejemplo.txt" http://localhost:8080/upload
```

Respuesta `201 Created`:

```json
{ "name": "<hash>.txt", "size": 100 }
```

Errores:

* `413 Payload Too Large` si excede `MaxBytes`.
* `400 Bad Request` si falta `file`.

### `GET /health`

Devuelve `ok`. En roadmap ampliar a métricas.

## WebSocket

URL:

```
ws://localhost:8080/ws?token=<TOKEN>&device_id=<ID>
```

Handshake mínimo y ejemplo con `websocat`:

```bash
websocat "ws://localhost:8080/ws?token=u123&device_id=dev-a"
```

Enviar un clip inline:

```json
{"type":"clip","mime":"text/plain","size":5,"data":"aGVsbG8="}
```

Recepción: todos los clientes del mismo usuario conectados al Hub.

## Cliente CLI

Objetivos v1:

* `clip-sync send --text "hola"`
* `clip-sync send --file ./nota.png`
* Modo pipe: `echo hola | clip-sync --mode send`
* Reconexión exponencial y logs limpios.

## Seguridad

* **Actual**: token sin firma, igual a `userID`.
* **Roadmap**: HMAC con caducidad, validación de `device_id`, lista de tipos permitidos en `/upload`.

## Pruebas

Unitarias:

```bash
go test ./server/...
```

* `httpapi/upload_test.go`: casos OK, tamaño excedido, método inválido.

E2E (WIP):

* WS básico + `upload_url` + señalización.
* Negativos: tamaño inconsistente, clip vacío.

## Roadmap

Ver [`TODO.md`](./TODO.md) para el checklist activo y criterios de aceptación.

## Licencia

WIP.
