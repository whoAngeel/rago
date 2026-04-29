# Plan de Implementación - RAGo

## Fase 1: Capa de Almacenamiento (Store) ✅
**Objetivo:** Implementar la integración completa con Qdrant.
- [x] Definir interfaz `VectorStore` en `internal/store/store.go`.
- [x] Implementar el cliente de Qdrant en `internal/store/qdrant.go`.
- [x] Crear métodos:
  - `CreateCollection(ctx, name, size)` → Crea colección con dimensión
  - `UpsertDocuments(ctx, collection, docs, vectors)` → Inserta puntos con vectores
  - `Search(ctx, collection, queryVector, limit)` → Busca documentos similares
  - `GetPointsCount(ctx, collection)` → Cuenta puntos en colección

## Fase 2: Orquestador (Engine) ✅
**Objetivo:** Crear el motor que coordina el flujo de datos.
- [x] Crear `internal/engine/rag_engine.go`.
- [x] Implementar la lógica de "RAG":
  1. Recibir pregunta.
  2. Generar embedding de la pregunta.
  3. Buscar en Qdrant.
  4. Formatear el contexto recuperado en un Prompt.
  5. Llamar al LLM y retornar la respuesta.

## Fase 3: Interfaz de Usuario (CLI) ✅
**Objetivo:** Exponer la funcionalidad al usuario final.
- [x] Implementar `cmd/rag/main.go` con comandos:
  - `ingest <archivo>` → Procesa y sube documento a Qdrant
  - `ask "<pregunta>"` → Ejecuta flujo RAG
  - `debug` → Muestra estado de colecciones
  - `delete <coleccion>` → Elimina colección
  - `reset` → Elimina todas las colecciones

## Fase 4: Soporte de Tipos de Archivo ✅
**Objetivo:** Procesar múltiples formatos.
- [x] `.txt` → Texto plano
- [x] `.md` → Markdown
- [ ] `.pdf` → PDF (pendiente)

## Fase 5: API REST (Pendiente)
- [ ] cmd/api/main.go → Entry point con Gin
- [ ] POST /ingest → Ingestar archivo
- [ ] POST /ask → Preguntar al RAG
- [ ] GET /health → Health check
- [ ] Autenticación (JWT/API key)

## Fase 6: Ingesta Automática (Pendiente)
- [ ] Watcher para MinIO/S3 → Polling de archivos nuevos
- [ ] Deduplicación → Hash de contenido
- [ ] Incremental updates → Solo documentos nuevos
- [ ] Soporte DOCX, CSV

## Fase 7: Mejoras RAG (Pendiente)
- [ ] Historial de conversación (Context Memory)
- [ ] Filtros por metadata (source, fecha)
- [ ] Reranking de resultados
- [ ] Cache de embeddings

## Fase 8: Producción (Pendiente)
- [ ] Logs estructurados (JSON)
- [ ] Rate limiting
- [ ] Múltiples colecciones (por tema/proyecto)
- [ ] Metrics (Prometheus)
- [ ] Docker/Deploy

---

---

## Puertos del Sistema (Arquitectura Hexagonal)

Los puertos son las interfaces que definen cómo la aplicación se comunica con el mundo exterior, sin depender de implementaciones concretas.

### VectorStore (`internal/core/ports/vector.store.go`)

Interfaz para operations con la base de datos vectorial (Qdrant).

| Método | Arguments | Returns | Descripción |
|--------|-----------|---------|-------------|
| `CreateCollection` | `ctx context.Context`, `name string` (nombre de colección), `size int` (dimensión de vectores, ej: 1536) | `error` | Crea una nuova colección en Qdrant. El `name` debe ser único. El `size` debe coincidir con la dimensión del modelo de embeddings. |
| `UpsertDocuments` | `ctx context.Context`, `collection string` (nombre), `docs []schema.Document` (documentos con content/metadata), `vectors [][]float32` (embeddings alineados con docs) | `error` | Inserta o actualiza documentos con sus vectores asociados. Cada documento en `docs` debe tener对应的vector en `vectors` en el mismo índice. |
| `Search` | `ctx context.Context`, `collection string` (nombre), `queryVector []float32` (vector de búsqueda), `limit int` (número máximo de resultados) | `([]SearchResult, error)` | Busca los `limit` documentos más similares al `queryVector` usando cosine similarity. Retorna slice de SearchResult con Document y score. |
| `GetPointsCount` | `ctx context.Context`, `collection string` | `(int64, error)` | Retorna el número total de puntos/documentos en la colección. |
| `DeleteCollection` | `ctx context.Context`, `collection string` | `error` | Elimina una colección completa. Debe manejar el caso de colección no existente. |

**Tipos auxiliares a crear:**
```go
type SearchResult struct {
    Document schema.Document
    Score    float32
}
```

---

### LLMProvider (`internal/core/ports/llm.provider.go`)

Interfaz para interacting con modelos de lenguaje y embeddings.

| Método | Arguments | Returns | Descripción |
|--------|-----------|---------|-------------|
| `GenerateAnswer` | `ctx context.Context`, `prompt string` (prompt completo con contexto + pregunta) | `(string, error)` | Genera una respuesta usando el LLM. El prompt ya debe contener el contexto recuperado + la pregunta del usuario. Retorna el texto de la respuesta generada. |
| `EmbedText` | `ctx context.Context`, `text string` | `([]float32, error)` | Convierte un texto en un vector de embedding. El vector retornado debe tener la dimensión configurada (ej: 1536 para text-embedding-3-small). |

**Notas de implementación:**
- Para `EmbedText`: el `[]float32` retornado debe tener exactamente la dimensión configurada en `EMBEDDING_DIMENSION`.
- Para `GenerateAnswer`: considera agregar un método `GenerateAnswerWithOptions(ctx, prompt, options)` que acepte parámetros como temperature, max_tokens.

---

### Logger (`internal/core/ports/logger.go`)

Interfaz para logging estructurado.

| Método | Arguments | Returns | Descripción |
|--------|-----------|---------|-------------|
| `Debug` | `msg string`, `args ...any` | - | Loggear mensaje de debug. Args son key-value pairs para contexto adicional. |
| `Info` | `msg string`, `args ...any` | - | Loggear mensaje informativo. |
| `Warn` | `msg string`, `args ...any` | - | Loggear warning (no fatal pero requiere atención). |
| `Error` | `msg string`, `args ...any` | - | Loggear error (operación falló). |
| `Fatal` | `msg string`, `args ...any` | - | Loggear error fatal y terminar programa. Equivalente a Error + os.Exit(1). |
| `With` | `args ...any` | `Logger` | Retorna una nueva instancia de Logger con el contexto agregado included en todos los logs subsecuentes. |

**Notas de implementación:**
- Implementar como wrapper alrededor de charmbracelet/log ozer.
- El método `With` debe retornar una nueva implementación (no mutar la actual) para evitar side effects.
- Formato recomendado: `logger.With("request_id", id).Info("request started")`.

---

## Estado Actual:

| Componente | Estado |
|------------|--------|
| Store (Qdrant) | ✅ Completado |
| Provider (Embeddings) | ✅ Completado |
| Ingest (Splitter) | ✅ Completado |
| Engine (RAG) | ✅ Completado |
| CLI | ✅ Completado |
| API REST | ❌ Pendiente |
| Ingesta Automática | ❌ Pendiente |
| Soporte PDF | ❌ Pendiente |
| Context Memory | ❌ Pendiente |

---

## Comandos CLI Disponibles:

```bash
./rago ingest <archivo>    # Ingestar documento
./rago ask "<pregunta>"     # Preguntar al RAG
./rago debug               # Ver estado de BD
./rago delete <coleccion>  # Eliminar colección
./rago reset              # Eliminar todas las colecciones
```

---

## Estructura del Proyecto:

```
rago/
├── cmd/
│   └── rag/main.go           # CLI
├── internal/
│   ├── engine/
│   │   ├── config.go        # Configuración
│   │   └── rag_engine.go    # Motor RAG
│   ├── ingest/
│   │   ├── process.go      # Procesador de documentos
│   │   └── splitter.go     # Fragmentador de texto
│   ├── provider/
│   │   └── embeddings.go   # Generador de embeddings
│   └── store/
│       ├── store.go       # Interfaz VectorStore
│       └── qdrant.go      # Implementación Qdrant
├── docs/
│   └── implementation_plan.md
├── .env                    # Variables de entorno
├── README.md              # Documentación completa
└── go.mod                 # Dependencias
```

---

## Configuración (.env):

```bash
# Qdrant
QDRANT_HOST="192.168.1.21"
QDRANT_PORT=6334

# Embeddings
EMBEDDING_MODEL="text-embedding-3-small"
EMBEDDING_DIMENSION=1536

# LLM (OpenRouter)
OPEN_ROUTER_API="sk-or-..."
OPEN_ROUTER_BASE_URL="https://openrouter.ai/api/v1"
LLM_MODEL="google/gemini-2.5-flash"
```

---

## Bugs Conocidos y Soluciones:

1. **ID duplicado** → Cada archivo nuevo obtiene IDs únicos basados en conteo de puntos
2. **Dimensión de vectores** → Crear colección con dimensión correcta (1536 para text-embedding-3-small)
3. **Vectores vacíos** → Usar `qdrant.NewVectorsDense()` helper

---

## Próxima Sesión:

1. **API REST** → Exponer funcionalidad como HTTP endpoints
2. **Soporte PDF** → Parser para documentos PDF
3. **Watcher MinIO** → Sincronización automática desde S3