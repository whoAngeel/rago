# Backend Implementation Reference

> Documento de referencia para el desarrollo frontend. Actualizar al completar cada iteración del backend.

---

## Iteración 1 — Checksum deduplication + límite de documentos

### Archivos modificados

| Archivo | Cambio |
|---|---|
| `internal/core/domain/document.go` | Campo `Checksum *string` (nullable, composite unique index con `user_id`) |
| `internal/core/domain/user.go` | Campo `MaxDocuments int` (default 10, `0` = ilimitado) |
| `internal/core/ports/database.go` | `FindByChecksum`, `CountDocumentsByUserID` en `DocumentRepository` |
| `internal/infrastructure/postgres/document_repo.go` | Implementación de los dos métodos anteriores |
| `internal/application/ingest_document_usecase.go` | Validaciones en `Upload()`: límite → checksum → upload |
| `internal/infrastructure/rest/handlers/document.go` | Mapeo de errores nuevos a 409/429 |
| `cmd/server/main.go` | `userRepo` pasado al constructor de `IngestDocumentUsecase` |

### Errores nuevos en upload

| Situación | HTTP | Body |
|---|---|---|
| Archivo ya subido por el usuario | `409 Conflict` | `{"error": "duplicate_document", "message": "Ya tenés este archivo subido"}` |
| Límite de documentos alcanzado | `429 Too Many Requests` | `{"error": "quota_exceeded", "message": "Alcanzaste el límite de documentos"}` |

### Lógica de negocio

- Orden de validaciones en upload: `max_documents` → `checksum` → upload a MinIO
- El checksum se calcula con SHA-256 vía `io.TeeReader` (una sola pasada del stream)
- `MaxDocuments = 0` en el usuario significa ilimitado (para admin)
- La columna `checksum` es nullable: docs existentes antes de esta iteración no generan conflictos en el índice

### Sentinels de error

```go
// internal/application/ingest_document_usecase.go
var (
    ErrNotFound             = errors.New("not found")
    ErrDuplicateDocument    = errors.New("duplicate document")
    ErrDocumentLimitReached = errors.New("document limit reached")
)
```

### DB (AutoMigrate automático al arrancar)

```sql
-- Columnas agregadas
ALTER TABLE documents ADD COLUMN checksum VARCHAR(64);
CREATE UNIQUE INDEX idx_documents_user_checksum ON documents(user_id, checksum);

ALTER TABLE users ADD COLUMN max_documents INT DEFAULT 10;

-- Post-deploy: hacer admin ilimitado
UPDATE users SET max_documents = 0 WHERE role_id = 1;
```

---

## Iteración 2 — CRUD de grupos de documentos (autenticado)

### Archivos nuevos

| Archivo | Contenido |
|---|---|
| `internal/core/domain/document_group.go` | Structs `DocumentGroup`, `DocumentGroupItem`, func `GenerateSlug()` |
| `internal/core/ports/database.go` | Interface `DocumentGroupRepository` (9 métodos) |
| `internal/infrastructure/postgres/document_group_repo.go` | Implementación completa |
| `internal/application/document_group_usecase.go` | CRUD + validaciones + auto-deactivate |
| `internal/infrastructure/rest/handlers/document_group_handler.go` | 7 endpoints |

### Archivos modificados

| Archivo | Cambio |
|---|---|
| `internal/infrastructure/rest/handlers/handlers.go` | Rutas `/api/v1/groups/` + refactor `setupRoutes` |
| `cmd/server/main.go` | AutoMigrate de 2 tablas nuevas + wiring `groupRepo` |

### Endpoints

Todos requieren `Authorization: Bearer <token>`.

| Método | Ruta | Handler | Body / Params |
|---|---|---|---|
| `POST` | `/api/v1/groups/` | `Create` | `{"name": "string", "document_ids": [1,2], "allow_downloads": true}` |
| `GET` | `/api/v1/groups/` | `List` | — |
| `GET` | `/api/v1/groups/:id` | `Get` | — |
| `PATCH` | `/api/v1/groups/:id` | `Update` | `{"name": "string", "is_active": bool, "allow_downloads": bool}` |
| `DELETE` | `/api/v1/groups/:id` | `Delete` | — |
| `POST` | `/api/v1/groups/:id/documents` | `AddDocuments` | `{"document_ids": [3,4]}` |
| `DELETE` | `/api/v1/groups/:id/documents/:doc_id` | `RemoveDocument` | — |

### Responses

**GET /groups/** y **GET /groups/:id** devuelven el tipo `DocumentGroup`:

```typescript
interface DocumentGroup {
    id: number
    user_id: number
    name: string
    slug: string          // 8 chars alfanumérico, usado para la URL pública /c/:slug
    is_active: boolean
    allow_downloads: boolean
    chat_quota: number    // default 100
    chat_quota_used: number
    chat_attempts: number
    document_ids: number[]
    created_at: string
    updated_at: string
}
```

### Lógica de negocio

- Solo se pueden agregar al grupo documentos con `status === "completed"` que pertenezcan al usuario autenticado
- Si se remueve el último documento de un grupo → `is_active` se pone `false` automáticamente
- El `slug` se genera con 8 caracteres alfanuméricos aleatorios (`[0-9a-zA-Z]`), único en toda la tabla
- El slug se usa como identificador para la URL pública: `/c/:slug`

### Sentinels de error

```go
// internal/application/document_group_usecase.go
var (
    ErrGroupNotFound   = errors.New("group not found")
    ErrInvalidDocument = errors.New("invalid document: must be completed and owned by user")
    ErrSlugCollision   = errors.New("slug collision after max retries")
)
```

### Errores HTTP

| Situación | HTTP | Dónde |
|---|---|---|
| Grupo no existe o no pertenece al usuario | `404 Not Found` | Get, Update, Delete, AddDocuments, RemoveDocument |
| Documento no completado o no pertenece al usuario | `400 Bad Request` | Create, AddDocuments |

### DB (AutoMigrate automático)

```sql
CREATE TABLE document_groups (
    id              SERIAL PRIMARY KEY,
    user_id         INT NOT NULL REFERENCES users(id),
    name            VARCHAR(255) NOT NULL,
    slug            VARCHAR(16) NOT NULL UNIQUE,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    allow_downloads BOOLEAN NOT NULL DEFAULT TRUE,
    chat_quota      INT NOT NULL DEFAULT 100,
    chat_quota_used INT NOT NULL DEFAULT 0,
    chat_attempts   INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMP,
    updated_at      TIMESTAMP
);

CREATE TABLE document_group_items (
    group_id    INT NOT NULL REFERENCES document_groups(id) ON DELETE CASCADE,
    document_id INT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    PRIMARY KEY (group_id, document_id)
);
```

---

## Iteración 3 — Chat público con cuota

### Archivos nuevos

| Archivo | Contenido |
|---|---|
| `internal/application/public_chat_usecase.go` | `GetGroupInfo`, `Chat`, `DownloadDocument` |
| `internal/infrastructure/rest/handlers/public_group_handler.go` | 3 endpoints sin auth |

### Archivos modificados

| Archivo | Cambio |
|---|---|
| `internal/core/ports/vector.store.go` | `Search` ahora acepta `documentIDs []int` (nil = sin filtro) |
| `internal/infrastructure/qdrant/qdrant.go` | Filtro dinámico: `userID > 0` → filtra por user_id; `documentIDs != nil` → filtra por document_id IN [...] |
| `internal/application/chat_usecase.go` | Callers de Search actualizados a `nil` (comportamiento idéntico) |
| `internal/application/ask_usecase.go` | Idem |
| `internal/core/ports/database.go` | `DocumentGroupRepository` + `FindBySlug`, `IncrementAttempts`, `IncrementQuotaUsed` |
| `internal/infrastructure/postgres/document_group_repo.go` | Implementación de los 3 métodos anteriores |
| `internal/infrastructure/rest/handlers/handlers.go` | Rutas públicas bajo `/api/v1/public/` (fuera del auth middleware) |
| `cmd/server/main.go` | Wiring de `PublicChatUsecase` y `PublicGroupHandler` |

### Endpoints públicos (sin `Authorization`)

| Método | Ruta | Descripción |
|---|---|---|
| `GET` | `/api/v1/public/groups/:slug` | Info del grupo + lista de documentos |
| `POST` | `/api/v1/public/groups/:slug/chat` | Chat con el grupo |
| `GET` | `/api/v1/public/groups/:slug/documents/:doc_id/download` | Descargar documento (si `allow_downloads`) |

### Response types

**GET /public/groups/:slug**
```typescript
interface GroupPublicInfo {
    name: string
    slug: string
    allow_downloads: boolean
    documents: Array<{ id: number; filename: string }>
}
```

**POST /public/groups/:slug/chat**
- Body: `{ "message": "string" }`
- Response 200: `{ "answer": "string" }`
- Response 404: grupo inactivo o inexistente → `{ "error": "...", "details": "Este chat no está disponible" }`
- Response 429: cuota agotada → `{ "error": "quota_exceeded", "message": "Este chat ha alcanzado su límite de mensajes" }`

### Lógica de negocio

- `IncrementAttempts` se llama **siempre**, incluso cuando la cuota está agotada
- Si `chat_quota_used >= chat_quota` → 429 (cuota agotada)
- `IncrementQuotaUsed` solo se llama después de una respuesta exitosa
- Búsqueda vectorial filtrada por `document_ids` del grupo (`userID = 0` → sin filtro de user)
- Descarga: verifica `is_active`, `allow_downloads`, y que el doc pertenezca al grupo

### Sentinels de error

```go
// internal/application/public_chat_usecase.go
var ErrQuotaExceeded = errors.New("quota exceeded")
var ErrGroupInactive = errors.New("group not found or inactive")
```

### Cambio de firma VectorStore.Search

```go
// ANTES
Search(ctx, collection, queryVector, userID, limit) ([]SearchResult, error)

// AHORA
Search(ctx, collection, queryVector, userID, documentIDs []int, limit) ([]SearchResult, error)
// userID=0 + documentIDs=[1,2,3] → filtra solo por document_id (caso público)
// userID=X + documentIDs=nil   → filtra solo por user_id (caso autenticado, sin cambio)
```

---

## Roles

Seeded en DB en `cmd/server/main.go`:

| ID | Nombre | Descripción |
|---|---|---|
| 1 | `admin` | Sin límite de documentos (`max_documents = 0`). Ver todo. |
| 2 | `viewer` | Reservado. Enforcing pendiente Iteración 2-F. |
| 3 | `editor` | Usuario regular. Default para nuevos registros. |

El rol se extrae del JWT en `internal/infrastructure/rest/middleware/jwt.go` y se almacena en el contexto Gin como `"role"` (string) y `"user_id"` (int).

Enforcing por rol en rutas: **pendiente, se implementa en frontend Iteración 2-F**.
