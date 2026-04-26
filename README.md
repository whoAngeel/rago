# RAGo - Retrieval Augmented Generation en Go

Sistema RAG (Retrieval Augmented Generation) implementado en Go usando Qdrant como base de vectores y LangChain para embeddings y LLM.

## Tabla de Contenidos

1. [Cómo Funciona RAG](#cómo-funciona-rag)
2. [Arquitectura del Sistema](#arquitectura-del-sistema)
3. [Componentes](#componentes)
4. [Configuración](#configuración)
5. [Uso](#uso)
6. [Conversión a API REST](#conversión-a-api-rest)
7. [Ingesta Automática desde MinIO/S3](#ingesta-automática-desde-minios3)
8. [Estado Actual](#estado-actual)
9. [Próximos Pasos](#próximos-pasos)

---

## Cómo Funciona RAG

RAG es una técnica que mejora las respuestas de un LLM proporcionando contexto relevante recuperado desde una base de conocimiento externa.

### Flujo Completo

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         FLUJO RAG                                   │
└─────────────────────────────────────────────────────────────────────┘

  USUARIO                 SISTEMA                    QDRANT
    │                        │                          │
    │  Pregunta:             │                          │
    │ "¿qué es Go?"         │                          │
    │───────────────────────>                          │
    │                        │                         │
    │                 ┌────▼────┐                   │
    │                 │ EMBEDDER│ (convierte texto     │
    │                 │        │  a vector)           │
    │                 └────┬────┘                   │
    │                      │ [0.1, 0.8, -0.3...]   │
    │                      │                        │
    │                      ▼                        │
    │            ┌─────────────────┐                │
    │            │ SEARCH (busca   │                │
    │            │ similares)     │─────────────────>│
    │            └────────┬────────┘                │
    │                     │   documentos similares │
    │                     │<───────────────────────┤
    │                     │                        │
    │               ┌─────▼──────┐                 │
    │               │ PROMPT      │ (arma prompt     │
    │               │ +contexto  │  con contexto)   │
    │               └─────┬──────┘                 │
    │                     │                         │
    │                     ▼                        │
    │            ┌──────────────┐                 │
    │            │ LLM (OpenAI) │                  │
    │            │ respuesta    │                  │
    │            └──────┬───────┘                 │
    │                   │                        │
    │  Respuesta:      │                        │
    │ "Go es un        │<─────────────────────── │
    │  lenguaje..."    │                         │
    │<────────────────│                         │
```

### Paso a Paso

1. **Usuario pregunta** → "¿qué es Go?"
2. **Embedder** → Convierte pregunta a vector numérico
3. **Qdrant Search** → Busca documentos con vectores similares
4. **Formatear Contexto** → Junta los chunks encontrados
5. **Prompt + Contexto** → Envia pregunta + contexto al LLM
6. **LLM responde** → Genera respuesta basada en el contexto

---

## Arquitectura del Sistema

```
rago/
├── cmd/rag/main.go           # CLI - Punto de entrada
├── internal/
│   ├── engine/
│   │   ├── config.go        # Carga configuración
│   │   └── rag_engine.go  # Orquestador principal
│   ├── ingest/
│   │   ├── splitter.go    # Fragmenta texto en chunks
│   │   └── process.go    # Prepara documentos
│   ├── provider/
│   │   └── embeddings.go # Genera embeddings
│   └── store/
│       ├── store.go       # Interfaz VectorStore
│       └── qdrant.go    # Implementación Qdrant
├── docs/
│   └── implementation_plan.md
├── .env                   # Variables de entorno
└── test.txt              # Archivo de prueba
```

---

## Componentes

### 1. Store (Qdrant) - `internal/store/`

**Propósito:** Almacenar y buscar vectores (embeddings)

```go
type VectorStore interface {
    CreateCollection(ctx context.Context, name string, size int) error
    UpsertDocuments(ctx context.Context, collection string, docs []schema.Document, vectors [][]float32) error
    Search(ctx context.Context, collection string, queryVector []float32, limit int) ([]schema.Document, error)
}
```

** Métodos:**
- `CreateCollection` → Crea colección en Qdrant con dimensión específica
- `UpsertDocuments` → Inserta documentos con sus vectores
- `Search` → Busca documentos similares al vector de consulta

**Por qué Qdrant:**
- Base de datos de vectores optimizada para búsqueda semántica
- API gRPC rápida
- Maneja millions de vectores eficientemente
- Fuzzy search, filtros, aggregations

### 2. Provider (Embeddings) - `internal/provider/`

**Propósito:** Convertir texto a vectores numéricos

```go
type Embedder struct {
    client *embeddings.EmbedderImpl
}

func (e *Embedder) ComputeEmbeddings(ctx context.Context, texts []string) ([][]float32, error)
```

**Por qué embeddings:**
- Texto → números (ej: "Go" → [0.1, 0.8, -0.3, ...])
- Palabras similares = vectores similares
- Permiten búsqueda semántica

**Modelos disponibles:**
- `text-embedding-3-small` (OpenAI) - 1536 dimensiones
- `text-embedding-3-large` (OpenAI) - 3072 dimensiones
- `gemini-embedding-001` (Google)

### 3. Ingest (Splitter) - `internal/ingest/`

**Propósito:** Fragmentar documentos grandes en chunks manejables

```go
type Splitter struct {
    splitter textsplitter.RecursiveCharacter
}

func (s *Splitter) SplitDocuments(docs []schema.Document) ([]schema.Document, error)
```

**Parámetros:**
- `chunkSize` = 500 caracteres (tamaño de cada fragmento)
- `chunkOverlap` = 100 caracteres (superposición entre chunks)

**Por qué fragmentar:**
- LLM tiene límite de contexto (ej: 8K, 32K, 128K tokens)
- Documentos grandes no caben en un solo prompt
-Chunks pequeños = búsqueda más precisa

### 4. Engine (RAG Engine) - `internal/engine/`

**Propósito:** Orquestar todo el flujo RAG

```go
type RAGEngine struct {
    store   store.VectorStore   // Qdrant
    embedder *provider.Embedder //Embeddings
    llm     llms.Model         // LLM (OpenAI)
}

func (e *RAGEngine) Ask(ctx context.Context, collection string, question string) (string, error)
```

**Flujo interno:**
1. `ComputeEmbeddings([question])` → vector
2. `store.Search()` → documentos similares
3. `formatContext()` �� formatea docs encontrados
4. `buildPrompt()` → crea prompt con contexto
5. `llm.Call()` → obtiene respuesta del LLM

### 5. CLI - `cmd/rag/main.go`

**Comandos:**

```bash
# Ingest: procesar documento y subir a Qdrant
./rago ingest archivo.txt

# Ask: preguntar al RAG
./rago ask "¿qué es Go?"
```

---

## Configuración

Variables en `.env`:

```bash
# Qdrant (vector database)
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

## Uso

### Ingestar Documento

```bash
./rago ingest test.txt
# Output: Documento procesado: 1 chunks
```

**Qué happening:**
1. Lee `test.txt`
2. Lo fragmenta en chunks (splitter)
3. Genera embeddings de cada chunk
4. Crea colección "docs" en Qdrant (si no existe)
5. Sube documentos + vectores a Qdrant

### Preguntar

```bash
./rago ask "¿qué es Go?"
```

**Qué happening:**
1. Convierte pregunta a embedding
2. Busca chunks similares en Qdrant
3. arma prompt con chunks encontrados
4. Envia a LLM
5. Retorna respuesta

---

## Conversión a API REST

Para exponer RAG como API REST:

### Estructura sugerida

```go
// cmd/api/main.go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/whoAngeel/rago/internal/engine"
    "github.com/whoAngeel/rago/internal/provider"
    "github.com/whoAngeel/rago/internal/store"
)

func main() {
    r := gin.Default()
    
    // Inicializar componentes
    cfg, _ := engine.LoadConfig()
    qdrantStore, _ := store.NewQdrantStore(cfg.QdrantHost, cfg.QdrantPort)
    embedder, _ := provider.NewEmbedder(cfg.OpenRouterKey, cfg.BaseUrl, cfg.EmbeddingModel)
    ragEngine, _ := engine.NewRAGEngine(qdrantStore, embedder, cfg)
    
    // Rutas
    r.POST("/ask", func(c *gin.Context) {
        var req struct {
            Question string `json:"question"`
        }
        c.ShouldBindJSON(&req)
        
        answer, err := ragEngine.Ask(c.Request.Context(), "docs", req.Question)
        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }
        c.JSON(200, gin.H{"answer": answer})
    })
    
    r.POST("/ingest", func(c *gin.Context) {
        // Endpoint para ingest desde URL o base64
    })
    
    r.Run(":8080")
}
```

### Endpoints sugeridos

| Método | Ruta | Descripción |
|--------|------|-------------|
| POST | `/ask` | Preguntar al RAG |
| POST | `/ingest` | Ingestar documento |
| GET | `/health` | Health check |
| GET | `/collections` | Listar colecciones |

---

## Ingesta Automática desde MinIO/S3

Para mantener RAG actualizado con archivos desde MinIO:

### Idea General

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   MinIO/S3   │────>│  Watcher    │────>│   Ingest   │
│  (archivos)  │     │  (intervalo)│     │  (procesa) │
└──────────────┘     └──────────────┘     └──────────────┘
                                                │
                                                ▼
                                        ┌──────────────┐
                                        │   Qdrant     │
                                        │ (vector DB)  │
                                        └──────────────┘
```

### Implementación sugerida

```go
// internal/sync/watcher.go
package sync

import (
    "context"
    "fmt"
    "time"
    
    "github.com/minio/minio-go/v7"
    "github.com/minio/minio-go/v7/pkg/notification"
)

type Watcher struct {
    client      *minio.Client
    qdrantStore store.VectorStore
    embedder    *provider.Embedder
    bucket      string
    interval    time.Duration
}

func NewWatcher(endpoint, accessKey, secretKey, bucket string) (*Watcher, error) {
    client, err := minio.New(endpoint, &minio.Options{
        Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
        Secure: false,
    })
    if err != nil {
        return nil, err
    }
    
    return &Watcher{
        client:   client,
        bucket:   bucket,
        interval: 5 * time.Minute, // verificar cada 5 minutos
    }, nil
}

func (w *Watcher) Start(ctx context.Context) error {
    ticker := time.NewTicker(w.interval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            if err := w.checkNewFiles(ctx); err != nil {
                fmt.Printf("Error: %v\n", err)
            }
        case <-ctx.Done():
            return ctx.Err()
        }
    }
}

func (w *Watcher) checkNewFiles(ctx context.Context) error {
    // 1. Listar objetos en bucket
    objects := w.client.ListObjects(w.bucket, minio.ListObjectsOptions{})
    
    for obj := range objects {
        // 2. Descargar archivo
        data, err := w.client.GetObject(ctx, w.bucket, obj.Key, minio.GetObjectOptions{})
        if err != nil {
            continue
        }
        
        // 3. Procesar e ingestar
        // ... (mismo proceso que CLI ingest)
    }
    
    return nil
}
```

### Opciones de Sincronización

1. **Polling (intervalo)**
   - Verificar cada N minutos
   - Simple, confiable
   - Ej: cada 5 minutos

2. **Event-driven (AWS SNS/SQS o MinIO Kafka)**
   - Notificaciones en tiempo real
   - Más complejo
   - Recomendado para producción

### Procesamiento por Tipo de Archivo

```go
func processFile(filename string, data []byte) error {
    switch ext := strings.ToLower(filepath.Ext(filename)); ext {
    case ".txt", ".md":
        return processText(data)
    case ".pdf":
        return processPDF(data)
    case ".docx":
        return processDOCX(data)
    case ".csv":
        return processCSV(data)
    default:
        return fmt.Errorf("tipo no soportado: %s", ext)
    }
}
```

---

## Estado Actual

### ✅ Completado

- [x] Capa de almacenamiento (Qdrant)
- [x] Generador de embeddings
- [x] Fragmentador de texto
- [x] Motor RAG (orquestador)
- [x] CLI básico (ingest + ask + debug + delete + reset)
- [x] Soporte TXT y MD
- [x] IDs únicos para documentos

### ❌ Pendiente

- [ ] API REST
- [ ] Ingesta automática desde MinIO/S3
- [ ] Soporte PDF
- [ ] Historial de conversación
- [ ] Manejo de errores robusto
- [ ] Logs estructurados

---

## Bugs Conocidos y Soluciones

### 1. ID duplicado (sobrescribía documentos)
**Problema:** Cada archivo usaba IDs desde 0, sobrescribiendo anteriores.
**Solución:** Obtener `GetPointsCount()` y asignar IDs secuenciales únicos.

### 2. Dimensión de vectores incorrecta
**Problema:** `expected dim: 1024, got 1536` - colección con dimensión diferente.
**Solución:** Usar `EMBEDDING_DIMENSION=1536` (para text-embedding-3-small) y hacer `reset` para recrear.

### 3. Vectores no se guardaban (Vectors: 0)
**Problema:** Estructura incorrecta para asignar vectores en `UpsertDocuments`.
**Solución:** Usar helper `qdrant.NewVectorsDense(vectors[i])`.

---

## Comandos CLI Disponibles

```bash
./rago ingest <archivo>    # Ingestar documento (txt, md)
./rago ask "<pregunta>"     # Preguntar al RAG
./rago debug               # Ver estado de BD (colecciones, puntos, docs)
./rago delete <coleccion>   # Eliminar colección específica
./rago reset              # Eliminar TODAS las colecciones
```

---

## Ejemplo de Uso

```bash
# Resetear BD
./rago reset

# Ingestar documentos
./rago ingest documentos/go.txt
./rago ingest documentos/python.txt
./rago ingest documentos/rag.txt

# Ver estado
./rago debug
# Output:
# Colecciones:
#   - docs
#     Points: 57
#     Documentos:
#       [1] Graduate Job Classification Case Study...

# Preguntar
./rago ask "¿qué es RAG?"
```

---

## Próximos Pasos

### 1. API REST (HIGH PRIORITY)

```bash
# Instalación
go install github.com/gin-gonic/gin@latest

# Estructura sugerida
cmd/api/main.go       # Entry point
internal/handlers/   # HTTP handlers
internal/middleware/ # Auth, CORS, etc.
```

### 2. Ingesta Automática (HIGH PRIORITY)

- Crear watcher para MinIO
- Procesamiento por tipo de archivo (PDF, DOCX, CSV)
- Deduplicación por hash
- Incremental updates

### 3. Mejoras (MEDIUM PRIORITY)

- Soporte PDF (librería pdf)
- Historial de conversación
- Cache de embeddings
- Rate limiting
- Autenticación

---

## Glosario

| Término | Definición |
|---------|-------------|
| **Embedding** | Representación numérica de texto |
| **Chunk** | Fragmento de documento |
| **Vector Store** | Base de datos para vectores |
| **Similarity Search** | Búsqueda por cercanía vectorial |
| **Prompt** | Texto enviado al LLM |
| **RAG** | Retrieval Augmented Generation |
| **MinIO** | Storage compatible con S3 |
| **Qdrant** | Vector database |

---

## Referencias

- [Qdrant Documentation](https://qdrant.tech/documentation/)
- [LangChain Go](https://tmc.github.io/langchaingo/)
- [OpenAI Embeddings](https://platform.openai.com/docs/guides/embeddings)
- [MinIO SDK](https://pkg.go.dev/github.com/minio/minio-go/v7)