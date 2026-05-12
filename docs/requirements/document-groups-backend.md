# Backend: Document Groups & Shared Chat
> Implementar por iteraciones en orden. Cada iteración es deployable y testeable de forma independiente.

---

## Iteración 1 — Fundamentos del modelo de documentos

Cambios al modelo y flujo de upload existente. Sin features nuevas visibles, solo guardianes.

### 1.1 Checksum / Deduplicación

**Objetivo:** evitar reprocesar el mismo archivo dos veces por usuario.

#### Migración DB
```sql
ALTER TABLE documents ADD COLUMN checksum VARCHAR(64) NOT NULL DEFAULT '';
CREATE UNIQUE INDEX idx_documents_user_checksum ON documents(user_id, checksum);
```

#### Cambios de código
| Archivo | Cambio |
|---|---|
| `internal/core/domain/document.go` | Agregar campo `Checksum string` con json/gorm tags |
| `internal/core/ports/database.go` | Agregar `FindByChecksum(ctx, userID int, checksum string) (*Document, error)` |
| `internal/infrastructure/postgres/document_repo.go` | Implementar `FindByChecksum` |
| `internal/application/ingest_document_usecase.go` | En `Upload()`: calcular SHA-256 del stream → llamar `FindByChecksum` → si existe, retornar error sentinel `ErrDuplicateDocument` |
| `internal/infrastructure/rest/handlers/document.go` | Mapear `ErrDuplicateDocument` a `409 Conflict` |

#### Detalle: cálculo del checksum
```go
// En Upload(), antes de BlobStorage.Upload()
// Leer el stream en un buffer, calcular hash, luego subir el buffer
h := sha256.New()
buf, err := io.ReadAll(io.TeeReader(fileReader, h))
checksum := hex.EncodeToString(h.Sum(nil))
```
Usar `io.TeeReader` para calcular el hash mientras se lee, sin leer el stream dos veces.

#### Contrato de error
```
POST /api/v1/documents/
→ 409 Conflict
{
  "error": "duplicate_document",
  "message": "Ya tenés este archivo subido"
}
```

---

### 1.2 Límite de archivos por usuario

**Objetivo:** preparar monetización futura. Por ahora, límite fijo de 10 documentos vivos por usuario.

#### Migración DB
```sql
ALTER TABLE users ADD COLUMN max_documents INT NOT NULL DEFAULT 10;
```

#### Cambios de código
| Archivo | Cambio |
|---|---|
| `internal/core/domain/user.go` | Agregar campo `MaxDocuments int` |
| `internal/core/ports/database.go` | Agregar `CountDocumentsByUserID(ctx, userID int) (int64, error)` |
| `internal/infrastructure/postgres/document_repo.go` | Implementar con `SELECT COUNT(*) WHERE user_id = ? AND deleted_at IS NULL` |
| `internal/application/ingest_document_usecase.go` | En `Upload()`: contar docs vivos → si `count >= user.MaxDocuments` → retornar `ErrDocumentLimitReached` |
| `internal/infrastructure/rest/handlers/document.go` | Mapear `ErrDocumentLimitReached` a `429 Too Many Requests` |

> **Qué cuenta como "doc vivo":** cualquier documento no eliminado, sin importar status (uploading, pending, processing, completed, failed).

#### Orden de validaciones en Upload()
1. Verificar `max_documents` → 429
2. Verificar checksum duplicado → 409
3. Proceder con upload

---

## Iteración 2 — CRUD de grupos (autenticado)

Endpoints para que el dueño cree y administre sus grupos. Sin chat público todavía.

### 2.1 Modelo de datos

```sql
CREATE TABLE document_groups (
    id              SERIAL PRIMARY KEY,
    user_id         INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name            VARCHAR(255) NOT NULL,
    slug            VARCHAR(16) NOT NULL UNIQUE,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    allow_downloads BOOLEAN NOT NULL DEFAULT TRUE,
    chat_quota      INT NOT NULL DEFAULT 100,
    chat_quota_used INT NOT NULL DEFAULT 0,
    chat_attempts   INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_document_groups_user ON document_groups(user_id);
CREATE UNIQUE INDEX idx_document_groups_slug ON document_groups(slug);

CREATE TABLE document_group_items (
    group_id    INT NOT NULL REFERENCES document_groups(id) ON DELETE CASCADE,
    document_id INT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    PRIMARY KEY (group_id, document_id)
);
```

### 2.2 Generación de slug

```go
// internal/core/domain/document_group.go
func GenerateSlug() (string, error) {
    b := make([]byte, 6)
    if _, err := rand.Read(b); err != nil {
        return "", err
    }
    return base62Encode(b)[:8], nil  // ej: "x7k2p9ab"
}
```
Reintentar si el slug ya existe en DB (colisión improbable pero posible).

### 2.3 Nuevas entidades de dominio

| Archivo | Contenido |
|---|---|
| `internal/core/domain/document_group.go` | Struct `DocumentGroup` + `GenerateSlug()` |

### 2.4 Ports

```go
// internal/core/ports/database.go
type DocumentGroupRepository interface {
    Create(ctx, group *domain.DocumentGroup) error
    FindByID(ctx, id, userID int) (*domain.DocumentGroup, error)
    FindByUserID(ctx, userID int) ([]*domain.DocumentGroup, error)
    Update(ctx, group *domain.DocumentGroup) error
    Delete(ctx, id, userID int) error
    AddDocument(ctx, groupID, documentID int) error
    RemoveDocument(ctx, groupID, documentID int) error
    FindDocumentIDs(ctx, groupID int) ([]int, error)
    SlugExists(ctx, slug string) (bool, error)
}
```

### 2.5 Usecase

`internal/application/document_group_usecase.go`

Métodos:
- `Create(ctx, userID int, name string, documentIDs []int) (*DocumentGroup, error)`
- `List(ctx, userID int) ([]*DocumentGroup, error)`
- `Get(ctx, id, userID int) (*DocumentGroup, error)`
- `Update(ctx, id, userID int, name string, isActive, allowDownloads bool) error`
- `Delete(ctx, id, userID int) error`
- `AddDocuments(ctx, groupID, userID int, documentIDs []int) error`
- `RemoveDocument(ctx, groupID, documentID, userID int) error`

**Regla de integridad:** después de `RemoveDocument`, si `len(FindDocumentIDs) == 0` → `Update(isActive: false)` automáticamente.

**Validación en Create:** solo aceptar `documentIDs` con status `completed` y que pertenezcan al `userID`.

### 2.6 Handler y rutas

```go
// internal/infrastructure/rest/handlers/document_group_handler.go
type DocumentGroupHandler struct { usecase *application.DocumentGroupUsecase }
```

```
POST   /api/v1/groups/                        → Create
GET    /api/v1/groups/                        → List
GET    /api/v1/groups/:id                     → Get
PATCH  /api/v1/groups/:id                     → Update (nombre, is_active, allow_downloads)
DELETE /api/v1/groups/:id                     → Delete
POST   /api/v1/groups/:id/documents           → AddDocuments  (body: { document_ids: [1,2,3] })
DELETE /api/v1/groups/:id/documents/:doc_id   → RemoveDocument
```

Todas requieren `AuthMiddleware()`.

---

## Iteración 3 — Chat público con cuota

Endpoints sin autenticación para el link compartido. Requiere actualizar el VectorStore interface.

### 3.1 Actualización de VectorStore interface

```go
// internal/core/ports/vector.store.go
Search(
    ctx         context.Context,
    collection  string,
    queryVector []float32,
    userID      int,
    documentIDs []int,   // nil = sin filtro (comportamiento actual)
    limit       int,
) ([]SearchResult, error)
```

Actualizar la implementación de Qdrant para agregar payload filter cuando `documentIDs != nil`:
```json
{ "must": [{ "key": "document_id", "match": { "any": [1, 3, 5] } }] }
```

Todos los callers existentes del `Search` pasan `nil` en `documentIDs` → sin breaking change en comportamiento.

### 3.2 Port público del grupo

```go
// internal/core/ports/database.go — en DocumentGroupRepository
FindBySlug(ctx context.Context, slug string) (*domain.DocumentGroup, error)
IncrementAttempts(ctx context.Context, groupID int) error
IncrementQuotaUsed(ctx context.Context, groupID int) error
```

### 3.3 Usecase público

`internal/application/public_chat_usecase.go`

```go
func (uc *PublicChatUsecase) GetGroupInfo(ctx, slug string) (*GroupPublicInfo, error)
func (uc *PublicChatUsecase) Chat(ctx, slug, message string) (string, error)
func (uc *PublicChatUsecase) DownloadDocument(ctx, slug string, docID int) (io.ReadCloser, string, error)
```

**Lógica de `Chat()`:**
1. `FindBySlug` → 404 si no existe
2. Si `!is_active` → 404
3. `IncrementAttempts` (siempre, incluyendo bloqueados)
4. Si `chat_quota_used >= chat_quota` → `ErrQuotaExceeded`
5. `FindDocumentIDs(groupID)` → obtener IDs del grupo
6. `VectorStore.Search(..., documentIDs, ...)` → resultados filtrados al grupo
7. LLM generate con contexto
8. `IncrementQuotaUsed`
9. Retornar respuesta

**Lógica de `DownloadDocument()`:**
1. `FindBySlug` → verificar `is_active` y `allow_downloads`
2. Verificar que `docID` pertenece al grupo
3. `BlobStorage.Download(doc.FilePath)`

### 3.4 Handler y rutas (sin auth)

```go
// internal/infrastructure/rest/handlers/public_group_handler.go
type PublicGroupHandler struct { usecase *application.PublicChatUsecase }
```

```
GET  /api/v1/public/groups/:slug                              → GetGroupInfo
POST /api/v1/public/groups/:slug/chat                         → Chat
GET  /api/v1/public/groups/:slug/documents/:doc_id/download   → DownloadDocument
```

Sin `AuthMiddleware()`. Rate limiting básico por IP recomendado (futuro).

#### Response de error por cuota
```json
HTTP 429
{
  "error": "quota_exceeded",
  "message": "Este chat ha alcanzado su límite de mensajes"
}
```

---

## Resumen de archivos nuevos/modificados por iteración

### Iteración 1
- `domain/document.go` — campo `Checksum`
- `domain/user.go` — campo `MaxDocuments`
- `ports/database.go` — `FindByChecksum`, `CountDocumentsByUserID`
- `postgres/document_repo.go` — implementaciones
- `application/ingest_document_usecase.go` — validaciones en `Upload()`
- `handlers/document.go` — mapeo de errores nuevos

### Iteración 2
- `domain/document_group.go` — nuevo
- `ports/database.go` — `DocumentGroupRepository` interface
- `postgres/document_group_repo.go` — nuevo
- `application/document_group_usecase.go` — nuevo
- `handlers/document_group_handler.go` — nuevo
- `handlers/handlers.go` — registrar rutas de grupos

### Iteración 3
- `ports/vector.store.go` — signature de `Search` con `documentIDs`
- `infrastructure/qdrant/` — implementación del filtro
- `application/public_chat_usecase.go` — nuevo
- `handlers/public_group_handler.go` — nuevo
- `handlers/handlers.go` — registrar rutas públicas

---

## Fuera de scope backend (para después)
- Migración Qdrant → pgvector / Pinecone
- Tabla de planes (por ahora solo `max_documents` en users)
- Rate limiting por IP en rutas públicas
- Expiración automática de links
