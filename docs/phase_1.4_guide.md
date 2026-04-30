# Guía: Release 1.4 - Ingesta Automática y Asíncrona (Workers)

## Objetivo
Aislar el procesamiento pesado (descargar archivo, extraer texto, hacer chunks, embeddings, guardar en Qdrant) del ciclo de vida HTTP. El usuario sube un archivo → el handler responde inmediatamente → un worker en background lo procesa.

## Decisiones Aplicadas

| Fuente | Decisión |
|--------|----------|
| Roadmap 1.4 | Worker Pool con goroutines + channels |
| Roadmap 1.4 | Polling de documentos `PENDING` |
| ADR 3.3 | Object key: `{user_id}/{document_id}/{filename}` |
| ADR 3.4 | Stream directo a MinIO (ya implementado en Upload handler) |
| ADR 1.6 | ContentType como selector de parser (futuro 1.5) |
| ADR 1.4 | FK documents.user_id → CASCADE (ya aplicado) |

---

## Estado Actual (Release 1.3)

El flujo Upload ya funciona:
```
POST /api/v1/documents (multipart)
  → Handler: validar extensión + tamaño
  → UseCase.Upload: crear registro BD (pending) + subir a MinIO
  → Response: {id, filename, status: "pending"}
```

Falta el procesamiento posterior.

---

## Paso 1: Puerto - Parser

Crea `internal/core/ports/parser.go`:

| Método | Descripción |
|--------|-------------|
| `Parse(ctx, reader io.Reader, contentType string) (string, error)` | Extrae texto del archivo |

En Release 1.4 solo implementas **texto plano** (`.txt`, contenido sin chunk). En 1.5 agregas PDF, DOCX, etc.

---

## Paso 2: Adapter - PlainText Parser

Crea `internal/core/parser/plaintext.go` (o en infrastructure):

Implementa `ports.Parser`:
- Lee todo el contenido con `io.ReadAll`
- Convierte a string
- Retorna

Solo maneja `text/plain`. Para otros content-types, retorna error "unsupported content type".

---

## Paso 3: Ampliar DocumentRepository

Agrega a `ports/database.go` y `postgres/document_repo.go`:

| Método | Descripción |
|--------|-------------|
| `FindDocumentsByStatus(ctx, status, limit) ([]*Document, error)` | Documentos con status específico, con límite |

```go
// ports
FindDocumentsByStatus(ctx context.Context, status domain.DocumentStatus, limit int) ([]*domain.Document, error)
```

En GORM:
```go
r.db.WithContext(ctx).Where("status = ?", status).Limit(limit).Find(&docs)
```

---

## Paso 4: Puerto - Worker

Crea `internal/core/ports/worker.go`:

```go
type Worker interface {
    Start(ctx context.Context)
    Stop()
}
```

Interfaz simple: arrancar y detener.

---

## Paso 5: IngestWorker

Crea `internal/worker/ingest_worker.go`:

```go
type IngestWorker struct {
    DocRepo      ports.DocumentRepository
    BlobStorage  ports.BlobStorage
    Parser       ports.Parser
    Embedder     ports.Embedder
    IngestUC     *application.IngestUsecase
    Logger       ports.Logger
    PollInterval time.Duration // ej. 10 segundos
    BatchSize    int           // cuántos documentos procesar por ciclo
    stopCh       chan struct{}
}
```

**Flujo del worker (método `Start`):**

1. Lanza goroutine con `time.NewTicker(PollInterval)`
2. Cada tick, llama método interno `processPendingDocuments`
3. `processPendingDocuments`:
   - Busca documentos con status `pending` (limit BatchSize)
   - Por cada documento:
     - Cambia status a `processing`
     - Descarga archivo de MinIO (`BlobStorage.Download(doc.FilePath)`)
     - Parsea texto (`Parser.Parse(reader, doc.ContentType)`)
     - Ingresa a Qdrant (`IngestUC.Execute(filename, content)`)
     - Si ok → status `completed`
     - Si error → status `failed` (con log del error)
   - El ciclo sigue aunque un documento falle

**Stop:**
```go
func (w *IngestWorker) Stop() {
    close(w.stopCh)
}
```

**Select en el ticker** para manejar graceful shutdown:
```go
select {
case <-ticker.C:
    w.processPendingDocuments()
case <-w.stopCh:
    return
}
```

---

## Paso 6: Composition Root (main.go)

Agregar al arranque:

```go
// Inicializar parser (por ahora plain text)
parser := parser.NewPlainTextParser()

// Inicializar worker
ingestWorker := worker.NewIngestWorker(
    docRepo,
    minio,         // BlobStorage
    parser,        // Parser
    embedder,      // ya existe
    ingestUC,      // ya existe
    log,
    10*time.Second, // PollInterval
    5,              // BatchSize
)
ingestWorker.Start(ctx)
```

**Graceful shutdown** (ya tienes el bloque, agrega):
```go
ingestWorker.Stop()
```

---

## Paso 7: Aislamiento Vectorial (Roadmap 1.4)

Cuando el worker inserta en Qdrant vía `IngestUC.Execute`, debe incluir `user_id` en los metadatos del documento. El `IngestUsecase` actual ya usa `domain.Document.Metadata` (vía `schema.Document`). 

Si no está implementado, agrega en `IngestUsecase.Execute`:
- Agregar `user_id` al metadata del chunk antes de hacer `UpsertDocuments`

Ejemplo del documento schema:
```go
doc := schema.Document{
    PageContent: chunk,
    Metadata: map[string]any{
        "source":    filename,
        "user_id":   userID,
    },
}
```

---

## Estructura de archivos nuevos

```
internal/
├── core/
│   ├── ports/
│   │   ├── parser.go     ← nuevo
│   │   └── worker.go     ← nuevo
│   └── parser/          ← nuevo directorio
│       └── plaintext.go
├── worker/              ← nuevo directorio
│   └── ingest_worker.go
```

---

## Orden sugerido

```
1. Ports/parser.go + ports/worker.go
2. parser/plaintext.go
3. FindDocumentsByStatus en port + adapter
4. worker/ingest_worker.go
5. main.go (wire worker + graceful shutdown)
6. Probar: subir archivo → status "pending" → worker lo procesa → "completed"
```
