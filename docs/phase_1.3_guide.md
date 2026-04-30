# Guía: Release 1.3 - Gestión de Documentos y BlobStorage

## Contexto actual
- Go + Gin + Arquitectura Hexagonal
- PostgreSQL con GORM y AutoMigrate
- JWT auth + roles (admin, editor, viewer)
- Qdrant (vector store) + OpenRouter (LLM)
- Ingest actual: endpoint directo `POST /api/v1/ingest` (texto plano)

---

## Objetivo
Separar la subida de archivos de su procesamiento. El usuario sube un archivo → se guarda físicamente en MinIO → se registra en PostgreSQL → luego se procesa (extraer texto → chunk → embed → Qdrant).

---

## Decisiones tomadas (ver docs/architecture-decisions.md para detalles)

| Área | Decisión |
|------|----------|
| Backend | **MinIO** como único BlobStorage (S3-compatible) |
| Bucket | **Un solo bucket** para todos los usuarios |
| Object keys | `{user_id}/{document_id}/{filename}` |
| Upload | Stream directo del request body a MinIO |
| Deduplicación | **No** por ahora |
| Cleanup | Borrado sincrónico: MinIO primero, luego BD |
| Tamaño máximo | Env var `MAX_FILE_SIZE` (default: 50MB) |
| Validación | Por extensión en el endpoint |
| Status | Custom type `DocumentStatus` con Scan/Value |
| Delete | Hard delete, sin soft delete en documentos |

---

## Paso 1: Puerto BlobStorage

Crea `internal/core/ports/storage.go`:

Interfaz para almacenar archivos físicos, agnóstica al proveedor (MinIO, S3, etc):

| Método | Descripción |
|--------|-------------|
| `Upload(ctx, objectKey string, reader io.Reader, size int64) error` | Sube archivo a MinIO (stream) |
| `Download(ctx, objectKey string) (io.ReadCloser, error)` | Recupera archivo |
| `Delete(ctx, objectKey string) error` | Elimina archivo |
| `Exists(ctx, objectKey string) (bool, error)` | Verifica si existe |

**Nota:** La implementación será `MinioAdapter`. El objectKey sigue el patrón `{user_id}/{document_id}/{filename}`.

---

## Paso 2: Dominio - Entidad Document

El archivo ya existe en `internal/core/domain/document.go`. Actualizar si es necesario:

- Status es `DocumentStatus` (custom type), **NO** `string`
- **No** tiene `DeletedAt` (hard delete)
- Tiene constraint FK: `OnDelete:CASCADE` a `users`
- `CreatedAt`, `UpdatedAt` manuales (no gorm.Model)

El tipo `DocumentStatus` ya está en `internal/core/domain/document_status.go`:
- Valores: `pending`, `processing`, `completed`, `failed`
- Implementa `database/sql.Scanner` y `driver.Valuer`

---

## Paso 3: Puerto DocumentRepository

Ya existe en `internal/core/ports/database.go`. Verificar que `UpdateDocumentStatus` use el tipo correcto:

| Método | Descripción |
|--------|-------------|
| `CreateDocument(ctx, doc) (*Document, error)` | INSERT |
| `FindDocumentsByUserID(ctx, userID) ([]*Document, error)` | SELECT WHERE user_id |
| `UpdateDocumentStatus(ctx, id, status DocumentStatus) error` | UPDATE status (nota: tipo DocumentStatus, no string) |

---

## Paso 4: Adapter MinIO

Crea `internal/infrastructure/storage/minio.go`:

Implementa `BlobStorage` usando el SDK oficial de MinIO (`github.com/minio/minio-go/v7`):

- Config: `Endpoint`, `AccessKeyID`, `SecretAccessKey`, `BucketName`, `UseSSL`
- Conexión inicial en el constructor con `minio.New()`
- `Upload`: `client.PutObject(ctx, bucket, objectKey, reader, size, opts)`
- `Download`: `client.GetObject(ctx, bucket, objectKey, opts)`
- `Delete`: `client.RemoveObject(ctx, bucket, objectKey, opts)`
- `Exists`: `client.StatObject(ctx, bucket, objectKey, opts)`

**Env vars necesarias:**
| Variable | Default | Descripción |
|----------|---------|-------------|
| `MINIO_ENDPOINT` | - | URL del servidor MinIO (ej: `localhost:9000`) |
| `MINIO_ACCESS_KEY_ID` | - | Access key |
| `MINIO_SECRET_ACCESS_KEY` | - | Secret key |
| `MINIO_BUCKET_NAME` | `rago-documents` | Nombre del bucket |
| `MINIO_USE_SSL` | `false` | Usar HTTPS |

---

## Paso 5: Adapter PostgreSQL para Document

Ya existe en `internal/infrastructure/postgres/document_repo.go`. Verificar:

- `UpdateDocumentStatus` recibe `domain.DocumentStatus`, no `string`
- `CreateDocument` retorna `(*Document, error)` o actualiza el puntero pasado

**Verificar que `&domain.Document{}` esté en el AutoMigrate en main.go.**

---

## Paso 6: Nuevo IngestDocumentUseCase

Crea `internal/application/ingest_document_usecase.go`:

NO reemplazar el IngestUseCase actual. Este nuevo maneja el flujo con archivos:

```go
type IngestDocumentUsecase struct {
    DocRepo    ports.DocumentRepository
    BlobStore  ports.BlobStorage
    IngestUC   *IngestUsecase // reusa el existente para procesar texto
}
```

Método `Upload(ctx, userID int, filename string, reader io.Reader, size int64) (*Document, error)`:
1. Crea registro Document con status `pending` en PostgreSQL (sin FilePath aún)
2. Construye objectKey: `fmt.Sprintf("%d/%d/%s", userID, doc.ID, filename)`
3. Sube stream a MinIO con ese objectKey
4. Si el upload falla → marca doc como `failed` y retorna error
5. Si ok → actualiza `doc.FilePath` con el objectKey y retorna doc

Método `ProcessDocument(ctx, docID int) error`:
1. Obtiene documento de BD
2. Cambia status a `processing`
3. Lee archivo del BlobStorage con `Download(doc.FilePath)`
4. Extrae texto (por ahora asume texto plano)
5. Llama a `IngestUsecase.Execute` (el actual que ya tienes)
6. Si ok → status `completed`, si error → status `failed`

Método `DeleteDocument(ctx, docID int) error`:
1. Obtiene documento de BD
2. Borra de MinIO primero: `BlobStore.Delete(doc.FilePath)`
3. Si el borrado de MinIO falla → retorna error (NO borra de BD)
4. Si ok → borra de la BD

---

## Paso 7: Handlers HTTP

Crea endpoint en `rest/document_handler.go`:

| Endpoint | Método | Auth | Body | Response |
|----------|--------|------|------|----------|
| `/api/v1/documents` | POST | Sí | multipart/form-data (file) | `{id, filename, status}` |
| `/api/v1/documents` | GET | Sí | - | lista de documentos del usuario |
| `/api/v1/documents/:id` | DELETE | Sí (admin/editor) | - | 204 No Content |
| `/api/v1/documents/:id/process` | POST | Sí (admin) | - | trigger procesamiento |

**Tips para el handler:**
- `c.FormFile("file")` para obtener el archivo multipart
- Extraer `user_id` de `c.Get("user_id")` (del middleware JWT)
- Validar extensión del archivo antes de procesar (extensiones permitidas: `.pdf`, `.docx`, `.xlsx`, `.csv`, `.json`, `.txt`)
- Respetar `MAX_FILE_SIZE` (configurable, default 50MB)
- Llamar al use case Upload
- Responder con el documento creado

---

## Paso 8: Config

Agregar a `.env` y `config.go`:

| Variable | Default | Descripción |
|----------|---------|-------------|
| `MAX_FILE_SIZE` | `52428800` (50MB) | Tamaño máximo de upload en bytes |
| `MINIO_ENDPOINT` | - | URL del servidor MinIO |
| `MINIO_ACCESS_KEY_ID` | - | Access key de MinIO |
| `MINIO_SECRET_ACCESS_KEY` | - | Secret key de MinIO |
| `MINIO_BUCKET_NAME` | `rago-documents` | Nombre del bucket |
| `MINIO_USE_SSL` | `false` | Usar HTTPS |

**Eliminar `UPLOAD_DIR`** — ya no se usa con MinIO.

---

## Paso 9: Rutas protegidas

En `handlers.go`, dentro del grupo protegido:
```go
protected.POST("/documents", docHandler.Upload)
protected.GET("/documents", docHandler.List)
protected.DELETE("/documents/:id", docHandler.Delete)
protected.POST("/documents/:id/process", docHandler.Process)
```

**Nota:** El endpoint de process puede ser solo para admin. Usa middleware RBAC.

---

## Paso 10: Worker (opcional para Release 1.4)

Para no bloquear HTTP, el procesamiento real iría en un worker asíncrono.
Pero por ahora puedes llamar a `ProcessDocument` sincrónicamente en el handler Upload.

---

## Orden sugerido

```
1. Ports: BlobStorage + DocumentRepository (ya existente, verificar firma)
2. Domain: Document entity (ya existe) + DocumentStatus (ya existe)
3. Config: agregar vars de MinIO + MAX_FILE_SIZE
4. Adapter: MinioAdapter
5. Adapter: document_repo.go (verificar firma de UpdateDocumentStatus)
6. UseCase: IngestDocumentUsecase
7. Handler: document_handler.go
8. Rutas + main.go (wire de MinioAdapter)
```
