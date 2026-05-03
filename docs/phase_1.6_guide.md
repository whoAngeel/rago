# Guía: Release 1.6 - Chat Contextual RAG y Memoria

## Objetivo
Permitir conversaciones continuas (Hilos de chat) y no solo preguntas aisladas. El motor RAG concatenará el historial reciente de mensajes desde Postgres con el contexto extraído de Qdrant.

## Decisiones Aplicadas

| Fuente | Decisión |
|--------|----------|
| Architecture 9.1 | Mensajes con content + sources (JSONB) |
| Architecture 9.2 | Ventana fija: últimos N mensajes (default 10) |
| Architecture 9.3 | Modo Estricto: sin contexto = "No sé" |
| Architecture 9.4 | Títulos auto-generados del primer mensaje |
| Architecture 9.5 | Lazy Creation de sesión |
| Architecture 9.6 | Prompt con XML Tags |
| Architecture 9.7 | System Prompt en tabla `system_configs` |
| Architecture 9.8 | Historial completo en BD |
| Architecture 9.9 | Atomicidad: no crear sesión si falla |
| Architecture 9.10 | Estructura: chat_sessions + chat_messages |

---

## Estado Actual (Phase 1.5)

El sistema ya maneja documentos, parsers y búsqueda vectorial. El endpoint `ask` funciona para preguntas aisladas.

---

## Paso 1: Tablas y Modelos

Crea `internal/core/domain/chat_session.go` y `chat_message.go`.

### ChatSession
```go
type ChatSession struct {
    gorm.Model
    UserID int    `gorm:"index;not null"`
    Title  string `gorm:"size:255"`
}
```

### ChatMessage
```go
type ChatMessage struct {
    gorm.Model
    SessionID int    `gorm:"index;not null"`
    Role      string `gorm:"size:20;not null"` // "user", "assistant"
    Content   string `gorm:"type:text;not null"`
    Sources   string `gorm:"type:jsonb"` // JSON de fuentes citadas
}
```

### SystemConfig
```go
type SystemConfig struct {
    gorm.Model
    Key   string `gorm:"uniqueIndex;size:50;not null"`
    Value string `gorm:"type:text;not null"`
}
```

**AutoMigrate:** Agregar `&domain.ChatSession{}`, `&domain.ChatMessage{}`, `&domain.SystemConfig{}`.

---

## Paso 2: Puertos (Interfaces)

Crea en `internal/core/ports/database.go`:

```go
type ChatRepository interface {
    CreateSession(ctx context.Context, session *domain.ChatSession) error
    GetSession(ctx context.Context, id, userID int) (*domain.ChatSession, error)
    GetUserSessions(ctx context.Context, userID int) ([]*domain.ChatSession, error)
    UpdateSessionTitle(ctx context.Context, id int, title string) error
    DeleteSession(ctx context.Context, id, userID int) error

    CreateMessage(ctx context.Context, msg *domain.ChatMessage) error
    GetSessionMessages(ctx context.Context, sessionID, limit int) ([]*domain.ChatMessage, error)
}

type SystemConfigRepository interface {
    Get(ctx context.Context, key string) (string, error)
    Set(ctx context.Context, key, value string) error
}
```

---

## Paso 3: Adapter PostgreSQL

Implementa las interfaces en `internal/infrastructure/postgres/chat_repo.go` y `config_repo.go`.

- `GetSessionMessages` debe ordenar por `created_at DESC` y limitar (para la ventana de contexto).
- `CreateMessage` debe guardar el JSON de sources como string en la columna `jsonb`.

---

## Paso 4: Seed del System Prompt

En `main.go`, al iniciar:

```go
const DefaultSystemPrompt = `Eres un asistente experto que responde preguntas basándose ÚNICAMENTE en el contexto proporcionado.
Instrucciones:
1. Usa solo la información dentro de las etiquetas <context> para responder.
2. Si el contexto no tiene suficiente información, responde: "No tengo información suficiente en tus documentos para responder a esto."
3. No inventes ni uses conocimiento general.
4. Si mencionas datos, cita las fuentes proporcionadas en el contexto.`

seedSystemConfig(db, "system_prompt", DefaultSystemPrompt)
```

---

## Paso 5: ChatUsecase

Crea `internal/application/chat_usecase.go`.

### Estructura
```go
type ChatUsecase struct {
    ChatRepo     ports.ChatRepository
    ConfigRepo   ports.SystemConfigRepository
    VectorStore  ports.VectorStore
    Embedder     ports.Embedder
    LLM          ports.LLMProvider
    Logger       ports.Logger
    HistoryLimit int
}
```

### Método `SendMessage(ctx, userID int, sessionID *int, question string) (answer string, sources []Source, newSessionID int, err error)`

Flujo:
1.  **Obtener/Crear Sesión**:
    - Si `sessionID` es nil → Crear nueva sesión (en memoria).
    - Si `sessionID` existe → Validar que pertenezca a `userID`.
2.  **Obtener System Prompt** de `ConfigRepo`.
3.  **Obtener Historial Reciente** (últimos N mensajes) de `ChatRepo`.
4.  **Búsqueda Vectorial**:
    - Embed de `question`.
    - Search en Qdrant con filtro `user_id`.
5.  **Construir Prompt con XML**:
    ```xml
    <system>{SystemPrompt}</system>
    <context>
      {Chunks de Qdrant}
    </context>
    <history>
      {Historial formateado: User: ... Assistant: ...}
    </history>
    <question>{question}</question>
    ```
6.  **LLM Generation**.
7.  **Verificar Contexto**: Si el LLM responde "No tengo información..." o similar, marcar fuentes como vacías.
    *(O mejor: Si Qdrant no devuelve resultados, omitir el bloque `<context>` y forzar el prompt de "No sé").*
8.  **Guardar**:
    - Guardar `ChatMessage` (User).
    - Guardar `ChatMessage` (Assistant + Sources).
    - Si es sesión nueva, guardar `ChatSession` y generar título desde la pregunta.
    - Retornar `newSessionID` si aplica.

---

## Paso 6: Handlers HTTP

Crea `internal/infrastructure/rest/handlers/chat_handler.go`.

| Endpoint | Método | Auth | Body | Response |
|----------|--------|------|------|----------|
| `/api/v1/chat/send` | POST | Sí | `{"session_id": 123, "question": "..."}` | `{"answer": "...", "sources": [...], "session_id": 123}` |
| `/api/v1/chat/sessions` | GET | Sí | - | Lista de sesiones (id, title, updated_at) |
| `/api/v1/chat/sessions/:id` | PATCH | Sí | `{"title": "Nuevo título"}` | Sesión actualizada |
| `/api/v1/chat/sessions/:id` | DELETE | Sí | - | 204 No Content |

Nota: `session_id` es opcional en el body de `/send`. Si falta, se crea una nueva.

---

## Paso 7: Rutas

En el router:
```go
protected.POST("/chat/send", chatHandler.Send)
protected.GET("/chat/sessions", chatHandler.ListSessions)
protected.PATCH("/chat/sessions/:id", chatHandler.UpdateSession)
protected.DELETE("/chat/sessions/:id", chatHandler.DeleteSession)
```

---

## Paso 8: Config

Agregar a `.env` y `config.go`:
- `CHAT_HISTORY_LIMIT` (default: 10)

---

## Estructura de archivos nuevos

```
internal/
├── core/
│   ├── domain/
│   │   ├── chat_session.go     ← nuevo
│   │   └── chat_message.go     ← nuevo
│   └── ports/
│       └── database.go         ← agregar ChatRepository
├── infrastructure/
│   └── postgres/
│       ├── chat_repo.go        ← nuevo
│       └── config_repo.go      ← nuevo
├── application/
│   └── chat_usecase.go         ← nuevo
└── infrastructure/rest/
    └── handlers/
        └── chat_handler.go     ← nuevo
```

---

## Orden sugerido

1.  Dominio y Modelos (Sessions, Messages, Config)
2.  AutoMigrate y Seed de System Prompt
3.  Repositorios (Chat + Config)
4.  ChatUsecase (Core Logic)
5.  Handler HTTP + Rutas
6.  Pruebas:
    - Preguntar sin contexto → Refusal.
    - Preguntar con contexto → Respuesta con fuentes.
    - Segunda pregunta (con session_id) → Contexto de historial incluido.
    - Verificar que se guarda el historial completo pero el LLM solo ve los últimos N.
