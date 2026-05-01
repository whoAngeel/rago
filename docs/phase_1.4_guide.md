# Guía: Release 1.4 - Ingesta Automática y Asíncrona (Workers)

## Objetivo
Aislar el procesamiento pesado (descargar archivo, extraer texto, hacer chunks, embeddings, guardar en Qdrant) del ciclo de vida HTTP. El usuario sube un archivo → el handler responde inmediatamente → un worker en background lo procesa.

## Decisiones Aplicadas

| Fuente | Decisión |
|--------|----------|
| Roadmap 1.4 | Worker Pool con goroutines + channels |
| Architecture 5.1 | PG Queue con `FOR UPDATE SKIP LOCKED` |
| Architecture 5.2 | Pool de 3 goroutines (`WORKER_CONCURRENCY`) |
| Architecture 5.3 | Retry con `max_retries=3`, `retry_count`, `error_message` |
| Architecture 5.4 | Parser en `infrastructure/parser/` |
| Architecture 5.5 | Recovery de docs stuck en `processing` > 5 min |
| Architecture 5.6 | Chunking semántico (párrafo/oración con fallback) |
| Architecture 5.7 | Pasar `*domain.Document` al usecase para metadata |
| Architecture 5.9 | Campos de progreso en `Document` |
| Architecture 5.10 | Tabla `processing_steps` para granularidad |

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

```go
type Parser interface {
    Parse(ctx context.Context, reader io.Reader, contentType string) (string, error)
}
```

En Release 1.4 solo implementas **texto plano** (`.txt`). En 1.5 agregas PDF, DOCX, etc.

---

## Paso 2: Adapter - PlainText Parser

Crea `internal/infrastructure/parser/plaintext.go`:

Implementa `ports.Parser`:
- Lee todo el contenido con `io.ReadAll`
- Convierte a string
- Retorna
- Solo maneja `text/plain`. Para otros content-types, retorna error "unsupported content type".

---

## Paso 3: Ampliar Document Model

Agregar a `internal/core/domain/document.go`:

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `ProcessingStartedAt` | `*time.Time` | Cuándo el worker empezó a procesar |
| `ErrorMessage` | `string` | Mensaje del último error |
| `RetryCount` | `int` | Contador de reintentos |

GORM tags apropiados para cada uno.

---

## Paso 4: Tabla Processing Steps

Crear modelo y migrar. Estructura:

```go
type ProcessingStep struct {
    ID           int       `gorm:"primaryKey"`
    DocumentID   int       `gorm:"index;not null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
    StepName     string    `gorm:"size:50;not null"` // download, parse, chunk, embed, upsert
    Status       string    `gorm:"size:20;not null"`  // started, completed, failed
    ErrorMessage *string   `gorm:"type:text"`
    DurationMS   *int      `json:"duration_ms"`
    CreatedAt    time.Time `gorm:"not null"`
}
```

Steps del pipeline:
1. **download** — Descargar de MinIO
2. **parse** — Extraer texto
3. **chunk** — Dividir en chunks
4. **embed** — Calcular embeddings
5. **upsert** — Insertar en Qdrant

Flujo para cada paso:
1. Insertar fila con `status=started`
2. Ejecutar operación
3. Actualizar fila con `status=completed/failed`, `duration_ms`, `error_message`

**No olvides agregar `&domain.ProcessingStep{}` al AutoMigrate.**

---

## Paso 5: Ampliar DocumentRepository

Agregar a `ports/database.go` y `postgres/document_repo.go`:

| Método | Descripción |
|--------|-------------|
| `FindPendingDocuments(ctx, limit) ([]*Document, error)` | Docs `pending` o `processing` viejo (recuperación) |
| `CreateProcessingStep(ctx, step *ProcessingStep) error` | INSERT paso |
| `UpdateProcessingStep(ctx, id, status, error, duration) error` | UPDATE paso |

Query para `FindPendingDocuments`:
```go
r.db.WithContext(ctx).
    Where("status = ? OR (status = ? AND updated_at < ?)",
        domain.StatusPending,
        domain.StatusProcessing,
        time.Now().Add(-5*time.Minute),
    ).
    Order("created_at ASC").
    Limit(limit).
    Find(&docs)
```

---

## Paso 6: Puerto - Worker

Crea `internal/core/ports/worker.go`:

```go
type Worker interface {
    Start(ctx context.Context)
    Stop()
}
```

Interfaz simple: arrancar y detener.

---

## Paso 7: Chunker (Semantic)

Crea `internal/core/ports/chunker.go`:

```go
type Chunker interface {
    Chunk(text string) ([]string, error)
}
```

Crea `internal/infrastructure/chunker/semantic.go`:

Implementación:
1. Split por párrafos (`\n\n`)
2. Agrupar párrafos hasta alcanzar `CHUNK_SIZE` (configurable, default 512)
3. Si un párrafo individual > `CHUNK_SIZE` → split por oraciones (`. `)
4. Si una oración > `CHUNK_SIZE` → fallback split por caracteres
5. Overlap: última oración del chunk se repite al inicio del siguiente (configurable `CHUNK_OVERLAP`, default 50)

Config por env vars:
- `CHUNK_SIZE` (default: 512)
- `CHUNK_OVERLAP` (default: 50)

---

## Paso 8: IngestWorker

Crea `internal/worker/ingest_worker.go`:

```go
type IngestWorker struct {
    DocRepo      ports.DocumentRepository
    BlobStorage  ports.BlobStorage
    Parser       ports.Parser
    Chunker      ports.Chunker
    Embedder     ports.Embedder
    IngestUC     *application.IngestUsecase
    Logger       ports.Logger
    PollInterval time.Duration
    Concurrency  int
    MaxRetries   int
    stopCh       chan struct{}
    processed    int64 // counter atómico
}
```

**Método `Start`:**
1. Lanza N goroutines (donde N = `Concurrency`)
2. Cada goroutine corre un loop:
   - `SELECT ... FOR UPDATE SKIP LOCKED` (limit 1) para obtener un doc
   - Si no hay docs → espera `PollInterval` y reintenta
   - Si hay doc → llama `processDocument(doc)`
3. Graceful shutdown con `select` en `stopCh`

**Método `processDocument(doc)`:**
1. Cambia status a `processing`, set `ProcessingStartedAt`
2. Para cada paso del pipeline, ejecutar y registrar en `processing_steps`:

```
download → parse → chunk → embed → upsert
```

3. Si todos exitosos → status `completed`
4. Si algún paso falla:
   - Incrementar `RetryCount`
   - Si `RetryCount >= MaxRetries` → status `failed`, guardar `ErrorMessage`
   - Si no → volver a `pending` para próximo ciclo
5. El ciclo sigue aunque un documento falle

**Método `Stop`:**
```go
func (w *IngestWorker) Stop() {
    close(w.stopCh)
}
```

---

## Paso 9: Adaptar IngestUsecase

Cambiar `IngestUsecase.Execute` para recibir el documento completo:

```go
type IngestUsecase struct {
    // ...existing fields
}

func (uc *IngestUsecase) Execute(ctx context.Context, doc *domain.Document, content string) error {
    // ...existing logic
    // Metadata injection:
    doc := schema.Document{
        PageContent: chunk,
        Metadata: map[string]any{
            "source":       doc.Filename,
            "user_id":      doc.UserID,
            "content_type": doc.ContentType,
            "document_id":  doc.ID,
        },
    }
}
```

---

## Paso 10: Composition Root (main.go)

Agregar al arranque:

```go
// Inicializar parser
parser := parser.NewPlainTextParser()

// Inicializar chunker
chunker := chunker.NewSemanticChunker(
    config.GetInt("CHUNK_SIZE", 512),
    config.GetInt("CHUNK_OVERLAP", 50),
)

// Inicializar worker
ingestWorker := worker.NewIngestWorker(
    docRepo,
    minio,         // BlobStorage
    parser,        // Parser
    chunker,       // Chunker
    embedder,      // ya existe
    ingestUC,      // ya existe
    log,
    10*time.Second, // PollInterval
    config.GetInt("WORKER_CONCURRENCY", 3),
    3,              // MaxRetries
)

// Iniciar worker con contexto
go ingestWorker.Start(ctx)

// Graceful shutdown (ya tienes el bloque, agrega):
ingestWorker.Stop()
```

---

## Paso 11: Aislamiento Vectorial (Roadmap 1.4)

Ya cubierto en Paso 9 — el `IngestUsecase` inyecta `user_id` en los metadatos de Qdrant. Asegúrate de que el `VectorStore.Search` filtre por `user_id` en los metadatos.

---

## Estructura de archivos nuevos

```
internal/
├── core/
│   ├── domain/
│   │   └── processing_step.go     ← nuevo
│   ├── ports/
│   │   ├── parser.go              ← nuevo
│   │   ├── chunker.go             ← nuevo
│   │   └── worker.go              ← nuevo
├── infrastructure/
│   ├── parser/
│   │   └── plaintext.go           ← nuevo
│   └── chunker/
│       └── semantic.go            ← nuevo
└── worker/
    └── ingest_worker.go           ← nuevo
```

---

## Orden sugerido

```
1. Ports: parser.go, chunker.go, worker.go
2. Domain: ProcessingStep model
3. Infrastructure: parser/plaintext.go, chunker/semantic.go
4. DB: FindPendingDocuments, CreateProcessingStep, UpdateProcessingStep
5. Adaptar IngestUsecase para recibir *domain.Document
6. Worker: ingest_worker.go
7. Config: WORKER_CONCURRENCY, CHUNK_SIZE, CHUNK_OVERLAP
8. main.go: wire worker + graceful shutdown
9. Probar: subir archivo → pending → worker procesa → steps registrados → completed
```
