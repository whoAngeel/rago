# Guía: Release 1.7 - Websockets e Interactividad en Tiempo Real

## Objetivo
Implementar Server-Sent Events (SSE) para notificaciones en tiempo real y streaming de respuestas del LLM.

## Decisiones Aplicadas

| Fuente | Decisión |
|--------|----------|
| Architecture 11.1 | Protocolo SSE (no WebSockets) |
| Architecture 11.2 | Canal único `/api/v1/stream` para todos los eventos |
| Architecture 11.3 | Streaming token por token del LLM |
| Architecture 11.4 | Formato nativo SSE (`event:`, `data:`) |
| Architecture 11.5 | Heartbeat cada 15s |
| Architecture 11.6 | Mapa en memoria para conexiones |

---

## Estado Actual (Phase 1.6)

El sistema ya maneja sesiones de chat y historial. El endpoint `POST /chat/send` devuelve la respuesta completa una vez que el LLM termina.

---

## Paso 1: Puerto SSE Manager

Crea `internal/core/ports/sse.go`:

```go
type SSEClient struct {
    ID      string
    UserID  int
    Channel chan SSEEvent
}

type SSEEvent struct {
    Type string
    Data any
}

type SSEManager interface {
    AddClient(userID int, client *SSEClient)
    RemoveClient(userID int, clientID string)
    SendToUser(userID int, event SSEEvent)
    SendToAll(event SSEEvent)
}
```

---

## Paso 2: Implementación del Manager

Crea `internal/infrastructure/sse/manager.go`.

- Usa `sync.Map` o `sync.Mutex` para proteger el mapa de clientes (un usuario puede tener varias pestañas abiertas).
- `SendToUser` itera sobre los canales de ese usuario y envía el evento (con recover por si un canal está cerrado).

---

## Paso 3: Endpoint de Stream

Crea handler `Stream` en `internal/infrastructure/rest/handlers/sse_handler.go`.

Flujo del handler:
1.  Validar token JWT → obtener `userID`.
2.  Configurar headers SSE:
    ```go
    c.Header("Content-Type", "text/event-stream")
    c.Header("Cache-Control", "no-cache")
    c.Header("Connection", "keep-alive")
    c.Header("X-Accel-Buffering", "no") // Desactivar buffering de nginx
    ```
3.  Crear `SSEClient` y registrarlo en el Manager.
4.  Loop infinito (hasta que se cierre el contexto):
    - Enviar `ping` cada 15s.
    - Escuchar el canal del cliente para enviar eventos reales.
    - Si el cliente se desconecta, limpiar y salir.

Formato de escritura al `ResponseWriter`:
```text
event: {Type}
data: {JSON}
id: {Timestamp}

```

---

## Paso 4: Integrar Streaming en el Chat

Actualizar `ChatUsecase` para soportar streaming.

Opciones:
1.  **Nuevo método `SendStream`**: Que retorne un `chan string` para que el handler vaya escribiendo en el stream.
2.  **Callback**: Pasar una función `onToken(token string)` al método `SendMessage`.

**Recomendación:** Opción 2 (Callback).
Es más limpio para conectar con el `SSEManager`.

```go
func (uc *ChatUsecase) SendStream(ctx context.Context, userID int, sessionID *int, question string, onToken func(token string)) (fullAnswer string, sources []Source, err error) {
    // ... RAG setup ...
    // LLM call con stream:
    stream := uc.LLM.Stream(ctx, prompt)
    for token := range stream {
        onToken(token)
        fullAnswer += token
    }
    // Guardar mensajes...
}
```

---

## Paso 5: Integrar Eventos de Documentos

El Worker (Fase 1.4) cambia el estado de los documentos. Cuando lo haga, debe notificar al Manager.

Inyectar `SSEManager` en el `IngestWorker`.
En el loop de procesamiento:

```go
// Cuando cambia el status
uc.SSEManager.SendToUser(doc.UserID, SSEEvent{
    Type: "document_status",
    Data: map[string]any{
        "id":     doc.ID,
        "status": newStatus,
        "error":  doc.ErrorMessage,
    },
})
```

---

## Paso 6: Handler de Chat (Modo Stream)

Actualizar `chat_handler.go`.

Nuevo endpoint: `POST /chat/send-stream`.
Funciona igual que `/chat/send`, pero:
1.  Configura headers SSE.
2.  Llama a `ChatUsecase.SendStream`.
3.  En el callback `onToken`, envía el evento `chat_token` al cliente.
4.  Al finalizar, envía `chat_done` con las fuentes y cierra.

**Nota:** Si el cliente no soporta SSE o prefiere respuesta completa, mantener el endpoint clásico `/chat/send`.

---

## Estructura de archivos nuevos

```
internal/
├── core/
│   └── ports/
│       └── sse.go              ← nuevo
├── infrastructure/
│   ├── sse/
│   │   └── manager.go          ← nuevo
│   └── rest/
│       └── handlers/
│           ├── sse_handler.go  ← nuevo
└── worker/
    └── ingest_worker.go        ← inyectar SSEManager
```

---

## Orden sugerido

1.  Puertos y Manager de SSE.
2.  Endpoint de Stream (`/api/v1/stream`) + Heartbeat.
3.  Integrar Worker con SSE para notificaciones de documentos.
4.  Adaptar `ChatUsecase` para soportar streaming (callback).
5.  Endpoint `/chat/send-stream`.
6.  Pruebas:
    - Abrir dos pestañas con el stream.
    - Subir un archivo → recibir evento de status.
    - Enviar mensaje al chat → ver tokens llegando uno a uno.
