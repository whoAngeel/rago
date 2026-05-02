# RAGO - Architecture Decisions Log

Documento de referencia con todas las decisiones de diseño tomadas durante el desarrollo.
Úsalo como contexto para continuar el trabajo en otra máquina o sesión.

---

## 1. Database Schema Decisions

### 1.1 DocumentStatus como tipo custom
- `DocumentStatus` es un custom type con valores: `pending`, `processing`, `completed`, `failed`
- Implementa `database/sql.Scanner` y `driver.Valuer` para compatibilidad con GORM
- Validación a nivel de aplicación con método `Valid()`
- File: `internal/core/domain/document_status.go`

### 1.2 Hard delete en Documents
- No hay soft delete (`DeletedAt`) en la tabla `documents`
- Si un documento se borra, sus referencias en `chat_messages.sources` quedan huérfanas
- El frontend maneja fuentes huérfanas mostrando "Documento eliminado"

### 1.3 Role usa gorm.Model
- `Role` embebe `gorm.Model` para tener `CreatedAt`, `UpdatedAt`, `DeletedAt`
- No tiene campo `description` (roles son autoexplicativos: admin, editor, viewer)

### 1.4 Foreign Key Constraints
| Relación | On Delete | On Update |
|---|---|---|
| `users.role_id` → `roles.id` | RESTRICT | CASCADE |
| `sessions.user_id` → `users.id` | CASCADE | CASCADE |
| `documents.user_id` → `users.id` | CASCADE | CASCADE |

### 1.5 chat_sessions.title
- Auto-generado del primer mensaje del usuario
- Editable vía PATCH por el usuario

### 1.6 chat_messages.sources (JSONB)
Estructura de cada referencia:
```json
{
  "document_id": 5,
  "chunk_id": "qdrant_point_abc",
  "page": 3,
  "text_preview": "..."
}
```

### 1.6 ContentType como único selector de parser
- No se necesita campo `FileType` o `ParserType` separado
- Se deriva el parser del MIME type (`content_type`)

---

## 2. Authentication Decisions

### 2.1 Token Storage en Frontend
- **Refresh token**: HttpOnly Cookie (no accesible por JS)
- **Access token**: Devuelto en body de login, guardado en memoria del frontend (NO LocalStorage)
- Patrón híbrido: seguridad de refresh token + simplicidad de access token stateless

### 2.2 Password Hashing
- **bcrypt** con costo por defecto (10)
- Librería: `golang.org/x/crypto/bcrypt`

### 2.3 Token Durations
- **Access token**: 15 minutos
- **Refresh token**: 7 días (extensible con "Remember me" en futuro)

### 2.4 Refresh Token Rotation
- Cada refresh genera un nuevo refresh token
- El anterior se revoca automáticamente
- Implementación: actualizar `refresh_token` y `expires_at` en tabla `sessions`

### 2.5 Brute Force Protection
- **Rate limiting por IP**: 5 intentos de login por minuto
- No hay account lockout temporal (se puede agregar en futuro si es necesario)

### 2.6 User Registration
- **Registro abierto** con `POST /register`
- Rol por defecto: `viewer`
- Admin puede cambiar rol manualmente

### 2.7 Forgot Password
- **No implementado por ahora**
- Se puede resetear manualmente en BD o vía admin
- Diseñar interfaz extensible para futuro email-based reset

### 2.8 Logout
- Revoke refresh token: set `revoked_at` en tabla `sessions`
- Borrar HttpOnly Cookie del cliente
- Access token sigue válido hasta expiración (15 min) — comportamiento esperado con JWT stateless

---

## 3. Storage Decisions

### 3.1 Backend de Almacenamiento
- **MinIO** como único BlobStorage (ya montado en homelab)
- No hay LocalFS como fallback
- La interfaz `BlobStorage` se implementa como `MinioAdapter`
- MinIO es S3-compatible, futuro cambio a S3 requiere mínimo código

### 3.2 Bucket Strategy
- **Un solo bucket** para todos los usuarios (ej: `rago-documents`)
- Organización por prefijo: `{user_id}/`
- No se crean buckets por usuario

### 3.3 Object Key Pattern
- Formato: `{user_id}/{document_id}/{filename}`
- Ej: `42/123/reporte.pdf`
- `document_id` es el ID autoincremental de la BD

### 3.4 Upload Strategy
- Stream directo del request body a MinIO (sin buffer completo en memoria)
- Si falla, el registro en BD queda con estado `FAILED`
- Cleanup de objetos parciales: worker o rutina de limpieza

### 3.5 Deduplicación
- **Sin deduplicación** por ahora
- Cada upload genera un nuevo archivo en MinIO, incluso si el contenido es idéntico

### 3.6 Cleanup de Archivos Huérfanos
- **Borrado sincrónico**: al hacer `DELETE /documents/:id`
- Primero se borra de MinIO, luego de la BD
- Si el borrado de MinIO falla, la transacción de BD no se hace

### 3.7 Tamaño Máximo de Archivo
- Configurable por variable de entorno: `MAX_FILE_SIZE`
- Default: 50MB (52428800 bytes)

### 3.8 Validación de Tipo de Archivo
- **Validar por extensión** en el endpoint antes de guardar en MinIO
- Extensiones soportadas: PDF, DOCX, XLSX, CSV, JSON, TXT (extensible)
- Validación de magic bytes no se necesita por ahora
- Si alguien sube un archivo con extensión válida pero contenido corrupto, el worker lo marca como `FAILED`

---

## 5. Worker & Ingestion Decisions (Phase 1.4)

### 5.1 Queue Mechanism
- **PostgreSQL-based queue** con `FOR UPDATE SKIP LOCKED`
- No se necesita Redis, RabbitMQ u otra infra nueva
- Persistente por naturaleza (si el server se reinicia, los docs siguen en BD)
- Soporta múltiples workers sin race conditions

### 5.2 Worker Pool Concurrency
- **Pool de goroutines** procesando docs en paralelo
- Default: **3 workers concurrentes**, configurable por `WORKER_CONCURRENCY`
- I/O bound (descarga MinIO + HTTP embeddings), así que concurrency va bien

### 5.3 Retry Strategy
- **Contador de reintentos** con `RetryCount` y `MaxRetries = 3`
- Campos agregados a `Document`: `retry_count int`, `error_message string`
- Si falla, incrementa `retry_count`. Si supera 3, se queda en `failed`
- No hay backoff exponencial en esta fase

### 5.4 Parser Location
- **En `infrastructure/parser/`**, no en `core/`
- `core` queda limpio (solo `domain/` y `ports/`)
- Todos los parsers (plaintext, PDF, DOCX, etc.) viven en `infrastructure/parser/`

### 5.5 Recovery de Stuck Documents
- Docs en `processing` por más de **5 minutos** se recuperan a `pending`
- Query: `WHERE status = 'pending' OR (status = 'processing' AND updated_at < NOW() - 5 min)`
- No requiere reset manual al arranque

### 5.6 Chunking Strategy
- **Chunking semántico**: split por párrafo/oración con fallback a token limit
- Párrafos completos se agrupan hasta llenar `CHUNK_SIZE`
- Párrafo más largo que `CHUNK_SIZE` → split por oración
- Fallback: split por caracteres
- Configurable: `CHUNK_SIZE` (default 512), `CHUNK_OVERLAP` (default 50)
- Sienta bases para chunking doc-aware en 1.5

### 5.7 Metadata Injection al Vector Store
- Se pasa `*domain.Document` completo al `IngestUsecase`
- El usecase extrae `user_id`, `filename`, `content_type` para metadata
- Así no hay que cambiar firmas cada vez que se necesita un nuevo campo
- Metadata en Qdrant: `{"user_id": N, "source": "filename", "content_type": "..."}`

### 5.8 Worker Logging
- **Structured JSON** con duración por documento
- Ej: `{"level":"info","doc_id":42,"status":"completed","duration_ms":3421,"user_id":5}`
- Counter de docs procesados en memoria (gratis y útil)

### 5.9 Document Progress Fields
Campos agregados a `Document` para frontend:
| Campo | Tipo | Uso |
|-------|------|-----|
| `ProcessingStartedAt` | `*time.Time` | Cuándo el worker empezó |
| `ErrorMessage` | `string` | Qué falló (último intento) |
| `RetryCount` | `int` | Cuántos reintentos |

### 5.10 Granularidad por Paso - Tabla `processing_steps`
Tabla separada para tracking detallado de cada etapa:

| Columna | Tipo | Descripción |
|---------|------|-------------|
| `id` | int (PK) | Auto-increment |
| `document_id` | int (FK) | → documents, CASCADE |
| `step_name` | string | "download", "parse", "chunk", "embed", "upsert" |
| `status` | string | "started", "completed", "failed" |
| `error_message` | string NULL | Error si falló |
| `duration_ms` | int NULL | Duración del paso |
| `created_at` | timestamp | Cuando se registró |

El frontend hace polling de steps para mostrar progress bar real.

---

## 6. Qdrant Vector Store Decisions

### 6.1 Payload Indexes
- Se crean **payload indexes keyword** para `user_id` y `document_id` al crear la colección
- Usar `CreateFieldIndex` con `FieldType_FieldTypeKeyword`
- Sin estos indexes, `Match_Keyword` en búsquedas funciona de forma impredecible

### 6.2 Point ID Generation
- Point IDs se generan como: `document_id * 10000 + chunk_index`
- Ej: doc 42, chunk 0 → ID 420000; doc 42, chunk 1 → ID 420001
- Esto evita colisiones entre documentos (antes todos empezaban en 0)
- Máximo ~10,000 chunks por documento (suficiente para cualquier caso real)

### 6.3 Aislamiento Vectorial
- Búsquedas filtradas por `user_id` usando `Match_Keyword`
- Un usuario solo ve resultados de sus propios documentos
- El filtro requiere el payload index de 6.1 para funcionar correctamente

### 6.4 Metadata Payload
- `formatPayload` convierte todos los metadatos a strings via `fmt.Sprintf("%v", v)`
- Esto es suficiente para filtros keyword pero limita queries numéricos en el futuro
- Si se necesita filtering numérico (ej: `score > 0.8`), cambiar a `PayloadIndexParams_IntegerIndexParams`

---

## 7. Parsers & File Types Decisions (Phase 1.5)

### 7.1 Parser Architecture
- **Strategy Pattern** con `ParserRegistry`
- Registry mapea content-type → parser específico
- Registration en `main.go` o init function

```go
registry.Register("text/plain", parser.NewPlainTextParser())
registry.Register("application/pdf", parser.NewPDFParser())
registry.Register("application/vnd.openxmlformats-officedocument.wordprocessingml.document", parser.NewDOCXParser())
registry.Register("text/csv", parser.NewCSVParser())
registry.Register("application/json", parser.NewJSONParser())
registry.Register("application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", parser.NewXLSXParser())
```

### 7.2 Archivos Estructurados (CSV, JSON)
- **Parser devuelve `[]schema.Document` directamente**, saltándose el chunker genérico
- CSV: cada fila es un documento con headers como metadata
- JSON: cada objeto/array element es un documento
- Metadata incluye: `row_number`, `headers`, `source`

### 7.3 PDF — Texto Nativo + OCR Fallback
- Primero intentar extraer texto nativo (PDFs generados digitalmente)
- Si no se obtiene texto → llamar a **OCRmyPDF** (Docker service en homelab)
- OCRmyPDF detecta automáticamente si necesita OCR y aplica Tesseract
- Docker image: `jbarlow83/ocrmypdf`
- Se invoca vía `exec.Command` desde Go

### 7.4 XLSX — Fila por Fila
- Cada fila se convierte en documento con headers como metadata
- Similar al parser CSV
- Metadata: `sheet_name`, `row_number`, `headers`
- Para hojas con 10+ columnas, se puede agrupar por bloque (optimización futura)

### 7.5 DOCX — Extracción directa
- Librería Go: `github.com/unidoc/unioffice` o similar
- Extrae párrafos y tablas manteniendo estructura lógica
- No requiere dependencias externas en el servidor

### 7.6 PDFs — Extracción Página por Página
- Se extrae texto de cada página individualmente
- Cada página pasa por el chunker semántico
- Metadata incluye `page_number`
- Permite mejor tracking de progreso y recuperación de fallos parciales

### 7.7 Imágenes Embebidas
- **Opción 3: Placeholder** — Se registra `[IMAGEN: nombre_archivo.jpg]` en el texto
- No se extrae contenido de imágenes por ahora
- OCR de imágenes embebidas se deja para fase posterior

### 7.8 OCR Infrastructure
- Docker service: `jbarlow83/ocrmypdf`
- Se invoca vía CLI: `ocrmypdf --skip-text input.pdf output.pdf`
- `--skip-text` evita re-OCR de páginas que ya tienen texto
- Límite de recursos: 2GB RAM en docker-compose
- Flujo: detectar texto nativo → si no hay → OCR → extraer texto del output

---

## 8. Tables Status

### Existing (implemented)
- `roles` — Con gorm.Model
- `users` — Con FK a roles
- `sessions` — JWT tokens, FK a users (CASCADE)
- `documents` — Con DocumentStatus, FK a users (CASCADE)

### Planned (Roadmap 1.6)
- `chat_sessions` — User, title, timestamps
- `chat_messages` — Session, role, content, sources (JSONB)

---

## 9. Key Files

| File | Purpose |
|---|---|
| `internal/core/domain/user.go` | Role + User models |
| `internal/core/domain/session.go` | Session model (JWT tokens) |
| `internal/core/domain/document.go` | Document model |
| `internal/core/domain/document_status.go` | DocumentStatus custom type |
| `internal/core/ports/database.go` | Repository interfaces |
| `internal/infrastructure/postgres/*` | GORM implementations |
| `internal/infrastructure/qdrant/qdrant.go` | Qdrant vector store adapter |
| `cmd/server/main.go` | DB init, AutoMigrate, role seed |
| `docs/database-schema.md` | ER diagram + table details |
| `docs/architecture-decisions.md` | This file |
