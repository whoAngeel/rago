# Guía: Release 1.3 - Gestión de Documentos y BlobStorage

## Contexto actual
- Go + Gin + Arquitectura Hexagonal
- PostgreSQL con GORM y AutoMigrate
- JWT auth + roles (admin, editor, viewer)
- Qdrant (vector store) + OpenRouter (LLM)
- Ingest actual: endpoint directo `POST /api/v1/ingest` (texto plano)

---

## Objetivo
Separar la subida de archivos de su procesamiento. El usuario sube un archivo → se guarda físicamente → se registra en PostgreSQL → luego se procesa (extraer texto → chunk → embed → Qdrant).

---

## Paso 1: Puerto BlobStorage

Crea `internal/core/ports/storage.go`:

Interfaz para almacenar archivos físicos, agnóstica al proveedor (LocalFS, S3, etc):

| Método | Descripción |
|--------|-------------|
| `Save(filename string, reader io.Reader) (path string, err error)` | Guarda archivo, retorna ruta |
| `Get(path string) (io.ReadCloser, error)` | Recupera archivo por ruta |
| `Delete(path string) error` | Elimina archivo |

---

## Paso 2: Dominio - Entidad Document

Crea `internal/core/domain/document.go`:

Struct Document con:
- ID, UserID (quién subió), Filename, FilePath (blob), Size, ContentType
- Status: string (pending, processing, completed, failed)
- CreatedAt, UpdatedAt

Tags GORM: `primaryKey`, `index` en UserID, `default:pending` en Status

---

## Paso 3: Puerto DocumentRepository

Crea en `internal/core/ports/database.go` (agrega a lo existente):

| Método | Descripción |
|--------|-------------|
| `CreateDocument(ctx, doc) error` | INSERT |
| `FindDocumentsByUserID(ctx, userID) ([]Document, error)` | SELECT WHERE user_id |
| `UpdateDocumentStatus(ctx, id, status) error` | UPDATE status |

---

## Paso 4: Adapter LocalFS

Crea `internal/infrastructure/storage/localfs.go`:

Implementa `BlobStorage` guardando en disco:
- Config: `UploadDir string` (ej: "./uploads")
- `Save`: crea archivo con nombre único (timestamp + UUID), retorna path relativo
- `Get`: abre archivo desde UploadDir + path
- `Delete`: os.Remove

---

## Paso 5: Adapter PostgreSQL para Document

Crea `internal/infrastructure/postgres/document_repo.go`:

Implementa `DocumentRepository` con GORM:
- CreateDocument: `db.Create`
- FindDocumentsByUserID: `db.Where("user_id = ?", userID).Find`
- UpdateDocumentStatus: `db.Model(&doc).Update("status", status)`

**No olvides agregar `&domain.Document{}` al AutoMigrate en main.go.**

---

## Paso 6: Nuevo IngestUseCase (asíncrono)

Crea `internal/application/ingest_document_usecase.go`:

NO reemplazar el IngestUseCase actual. Este nuevo maneja el flujo con archivos:

```go
type IngestDocumentUsecase struct {
    DocRepo    ports.DocumentRepository
    BlobStore  ports.BlobStorage
    IngestUC   *IngestUsecase // reuse el existente para procesar texto
}
```

Método `Upload(ctx, userID, fileName, fileReader) error`:
1. Guarda en BlobStorage → obtiene path
2. Crea registro Document con status "pending" en PostgreSQL
3. Retorna (el procesamiento real se hará después en un worker)

Método `ProcessDocument(ctx, docID) error`:
1. Obtiene documento de BD
2. Cambia status a "processing"
3. Lee archivo del BlobStorage
4. Extrae texto (por ahora asume texto plano)
5. Llama a `IngestUsecase.Execute` (el actual que ya tienes)
6. Si ok → status "completed", si error → status "failed"

---

## Paso 7: Handlers HTTP

Crea endpoint en `rest/document_handler.go`:

| Endpoint | Método | Auth | Body | Response |
|----------|--------|------|------|----------|
| `/api/v1/documents` | POST | Sí | multipart/form-data (file) | `{id, filename, status}` |
| `/api/v1/documents` | GET | Sí | - | lista de documentos del usuario |
| `/api/v1/documents/:id/process` | POST | Sí (admin) | - | trigger procesamiento |

**Tips para el handler:**
- `c.FormFile("file")` para obtener el archivo multipart
- Necesitas extraer `user_id` de `c.Get("user_id")` (del middleware JWT)
- Llamar al use case Upload
- Responder con el documento creado

---

## Paso 8: Config

Agregar a `.env` y `config.go`:
- `UPLOAD_DIR` (default: "./uploads")
- `MAX_UPLOAD_SIZE` (ya existe)

---

## Paso 9: Rutas protegidas

En `handlers.go`, dentro del grupo protegido:
```go
protected.POST("/documents", docHandler.Upload)
protected.GET("/documents", docHandler.List)
protected.POST("/documents/:id/process", docHandler.Process)
```

---

## Paso 10: Worker (opcional para Release 1.4)

Para no bloquear HTTP, el procesamiento real iría en un worker asíncrono.
Pero por ahora puedes llamar a `ProcessDocument` sincrónicamente en el handler Upload.

---

## Orden sugerido

```
1. Ports: BlobStorage + DocumentRepository
2. Domain: Document entity + tags GORM
3. Adapter: LocalFS
4. Adapter: document_repo.go
5. UseCase: IngestDocumentUsecase
6. Handler: document_handler.go
7. Rutas + main.go
```
