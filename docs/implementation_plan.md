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