# RAGo — Retrieval Augmented Generation en Go

RAGO es una herramienta de inteligencia documental privada y autoalojada (self-hosted). Subes documentos (PDF, escaneos, imágenes), se procesan vía OCR + embeddings y se indexan en una base de datos vectorial. Los grupos de documentos exponen un chat público (sin cuenta) para consultarlos.

Backend en Go con arquitectura hexagonal (ports & adapters), frontend en React (SPA), Postgres para metadatos, Qdrant para vectores, MinIO para blobs.

## Tabla de Contenidos

1. [Cómo Funciona RAG](#cómo-funciona-rag)
2. [Arquitectura](#arquitectura)
3. [Estructura del Repositorio](#estructura-del-repositorio)
4. [Flujo de Ingesta](#flujo-de-ingesta)
5. [Flujo de Chat / Ask](#flujo-de-chat--ask)
6. [Autenticación](#autenticación)
7. [API REST](#api-rest)
8. [Worker (procesamiento en background)](#worker-procesamiento-en-background)
9. [Frontend](#frontend)
10. [Despliegue](#despliegue)
11. [Configuración](#configuración)
12. [Estado Actual](#estado-actual)
13. [Glosario](#glosario)
14. [Referencias](#referencias)

---

## Cómo Funciona RAG

RAG mejora las respuestas de un LLM dándole contexto recuperado desde una base de conocimiento externa, en vez de depender solo de lo que el modelo aprendió en entrenamiento.

```mermaid
sequenceDiagram
    actor U as Usuario
    participant E as Embedder
    participant Q as Qdrant
    participant L as LLM (OpenRouter)

    U->>E: "¿qué es Go?"
    E->>E: texto → vector [0.1, 0.8, -0.3...]
    E->>Q: Search(vector, limit)
    Q-->>E: chunks similares
    E->>L: prompt = contexto + pregunta
    L-->>U: respuesta generada
```

1. **Usuario pregunta** → texto plano.
2. **Embedder** → convierte la pregunta a vector.
3. **Qdrant Search** → busca chunks con vectores similares.
4. **Prompt + Contexto** → arma prompt con los chunks encontrados.
5. **LLM responde** → genera respuesta basada en ese contexto.

---

## Arquitectura

Hexagonal / ports & adapters: `internal/core` no importa nada de infraestructura, `internal/application` depende solo de `core/ports`, y `internal/infrastructure` implementa esos puertos. `cmd/server` es quien cablea todo.

```mermaid
graph TB
    subgraph FE["Frontend — React SPA (nginx :8080)"]
        UI[TanStack Router + Query]
    end

    subgraph API["cmd/server — Gin (:4000)"]
        MW[Middleware: CORS, RequestID, Auth JWT]
        H[Handlers REST]
    end

    subgraph APP["internal/application — casos de uso"]
        UC1[IngestDocumentUsecase]
        UC2[IngestUsecase]
        UC3[AskUsecase]
        UC4[ChatUsecase / PublicChatUsecase]
        UC5[AuthUseCase]
        UC6[DocumentGroupUsecase]
    end

    subgraph CORE["internal/core"]
        PORTS["ports/* — interfaces\nVectorStore · Embedder · LLMProvider\nParser · BlobStorage · Repos · SSE"]
        DOM["domain/* — entidades\nUser · Document · ChatSession\nDocumentGroup · LLMUsage"]
    end

    subgraph INFRA["internal/infrastructure — adapters"]
        PG[(Postgres / GORM)]
        QD[(Qdrant)]
        MI[(MinIO / S3)]
        OR[OpenRouter\nLLM + Embeddings]
        PA["Parsers\nPDF·DOCX·XLSX·CSV·JSON·Imagen+OCR"]
        SSE[SSE Manager]
        JWT[JWT Auth]
    end

    subgraph W["internal/worker"]
        IW["IngestWorker\npoll 10s · concurrency 3\nretries 3 · watchdog 15min"]
    end

    UI -->|JSON / SSE| H
    H --> MW --> APP
    APP --> PORTS
    IW --> PORTS
    PORTS -.implementado por.-> INFRA
    APP --> SSE
    IW --> SSE
    SSE -->|eventos live| UI
```

**Por qué hexagonal:** los casos de uso (`AskUsecase`, `ChatUsecase`, etc.) no saben si el vector store es Qdrant o el storage es MinIO — solo conocen la interfaz (`ports.VectorStore`, `ports.BlobStorage`). Cambiar de proveedor es escribir un adapter nuevo, no tocar la lógica de negocio.

---

## Estructura del Repositorio

```
rago/
├── cmd/
│   ├── server/           # entrypoint API — wiring completo (main.go)
│   ├── inspect_qdrant/   # utilidad debug: inspeccionar colecciones Qdrant
│   └── ocr-test/         # utilidad debug: probar OCR aislado
├── internal/
│   ├── core/
│   │   ├── domain/       # entidades: User, Document, ChatSession, DocumentGroup, LLMUsage...
│   │   └── ports/        # interfaces: VectorStore, Embedder, LLMProvider, Parser, BlobStorage, repos...
│   ├── application/      # casos de uso: Ingest, Ask, Chat, Auth, DocumentGroup
│   ├── infrastructure/
│   │   ├── auth/         # JWT (access + refresh tokens)
│   │   ├── chunker/      # FixedChunker (párrafo → oración → corte duro, 1000/200)
│   │   ├── config/       # carga de variables de entorno
│   │   ├── logger/
│   │   ├── openrouter/   # LLM + Embedder (langchaingo, API compatible OpenAI)
│   │   ├── parser/       # texto, csv, json, docx, xlsx, pdf, imagen+OCR (tesseract)
│   │   ├── postgres/     # repositorios GORM
│   │   ├── qdrant/       # adapter VectorStore
│   │   ├── rest/         # router, handlers, middleware (Gin)
│   │   ├── sse/          # pub/sub de eventos en vivo
│   │   └── storage/      # adapter MinIO
│   └── worker/           # IngestWorker: polling, concurrencia, reintentos, watchdog
├── frontend/              # SPA React 19 + TanStack Router/Query + Tailwind v4
├── docs/
├── docker-compose.yml     # frontend + app + postgres (Qdrant y MinIO corren aparte)
└── .env
```

---

## Flujo de Ingesta

Subida es **síncrona** (guarda archivo y responde rápido); el procesamiento pesado (parseo, OCR, chunking, embeddings, upsert) es **asíncrono**, lo hace `IngestWorker` por polling.

```mermaid
sequenceDiagram
    actor U as Usuario
    participant API as DocumentHandler
    participant UC as IngestDocumentUsecase
    participant PG as Postgres
    participant MI as MinIO
    participant W as IngestWorker
    participant P as Parser (+OCR/tesseract)
    participant C as Chunker
    participant E as Embedder
    participant Q as Qdrant

    U->>API: POST /documents (archivo)
    API->>UC: Upload(file)
    UC->>UC: SHA-256 + dedupe por (user_id, checksum)
    UC->>PG: crea Document (status=uploading)
    UC->>MI: sube bytes crudos
    UC->>PG: status=pending
    API-->>U: 201 {id, status: pending}

    rect rgb(235,235,235)
    note over W: cada 10s
    W->>PG: FindPendingDocuments(concurrency=3)
    W->>PG: status=processing
    W->>MI: descarga archivo
    W->>P: Parse, timeout 10min, OCR si es imagen o PDF escaneado
    P-->>W: texto extraído
    W->>C: Chunk(texto) — 1000 chars, overlap 200
    W->>E: EmbedText(chunk) por cada chunk
    W->>Q: DeleteByDocumentID → UpsertDocuments
    W->>PG: status=completed
    end
    W-->>U: SSE document_status / document_step
```

Puntos clave:

- **Deduplicación**: por hash SHA-256, rechaza re-subir el mismo archivo del mismo usuario.
- **OCR**: subprocess `tesseract` (idiomas `spa+eng` por defecto). PDFs primero intentan extracción de texto nativa; si una página no rinde texto (o se fuerza `force_ocr`), se rasteriza con `pdftoppm` y pasa por OCR.
- **Chunker**: `FixedChunker` — párrafo-consciente, 1000 caracteres objetivo, 200 de overlap, con fallback a corte por oración y corte duro para párrafos gigantes. (El paquete se llama `chunker/semantic.go` pero **no** es chunking semántico basado en embeddings.)
- **Delete antes de upsert**: cada ingesta borra los vectores previos del documento en Qdrant (`DeleteByDocumentID`) antes de insertar los nuevos — evita chunks duplicados/obsoletos en reprocesos.
- **Reintentos**: hasta 3; si el error persiste, `status=failed`.
- **Watchdog**: cada 2 minutos revisa documentos atascados en `processing` por más de 15 minutos y los regresa a `pending`.
- **Progreso en vivo**: cada paso (download, parse, chunk, embed, upsert) se registra como `ProcessingStep` y se emite por SSE (`GET /api/v1/stream`).

---

## Flujo de Chat / Ask

Tres variantes, misma idea base (embed pregunta → buscar en Qdrant → prompt con contexto → LLM):

- **`POST /ask`** — one-shot, sin historial, sin persistencia.
- **`POST /chats/send`** / **`POST /chats/send-stream`** — con sesión, historial, y streaming de tokens vía SSE.
- **`POST /public/groups/:slug/chat`** — sin autenticación, acotado a los documentos de un grupo público, con umbral de confianza (score ≥ 0.35).

```mermaid
sequenceDiagram
    actor U as Usuario
    participant API as ChatHandler
    participant UC as ChatUsecase
    participant Q as Qdrant
    participant LLM as OpenRouter LLM
    participant PG as Postgres

    U->>API: POST /chats/send-stream {question}
    API->>UC: SendStream(question)
    UC->>PG: carga historial (últimos N mensajes)
    UC->>Q: embed(question) + Search(user_id, limit)
    Q-->>UC: chunks relevantes
    UC->>UC: prompt = system_prompt + CONTEXTO + HISTORIAL + PREGUNTA
    UC->>LLM: Stream(prompt)
    loop por token
        LLM-->>API: chat_token (SSE)
    end
    UC->>PG: guarda ChatMessage (user + assistant, con sources)
    UC->>PG: registra LLMUsage (tokens in/out)
    API-->>U: chat_done (SSE)
```

---

## Autenticación

JWT stateless para access tokens (HS256), refresh tokens opacos persistidos en Postgres (`Session`) para poder revocarlos en logout.

```mermaid
sequenceDiagram
    actor U as Usuario
    participant API as AuthHandler
    participant UC as AuthUseCase
    participant PG as Postgres

    U->>API: POST /auth/login {email, password}
    API->>UC: Login()
    UC->>PG: valida usuario + password hash
    UC->>PG: guarda Session {refresh_token, expires_at}
    UC-->>U: {access_token (15m), refresh_token (30d)}

    U->>API: POST /auth/refresh {refresh_token}
    API->>UC: Refresh()
    UC->>PG: valida sesión no revocada/expirada
    UC-->>U: nuevo access_token
```

- `Authorization: Bearer <token>`, con fallback a `?token=` para clientes SSE (`EventSource` no permite headers custom).
- Rutas públicas (sin token): `/health`, `/auth/*`, `/public/*`.
- Existe middleware `RequireRole` para RBAC pero **no está aplicado a ninguna ruta todavía**.

---

## API REST

Router: `internal/infrastructure/rest/handlers`. Gin + `Recovery`, `RequestLogger`, CORS abierto, `X-Request-ID`.

### Públicas (sin auth)

| Método | Ruta | Descripción |
|--------|------|-------------|
| GET | `/health` | Health check profundo (Postgres + Qdrant + MinIO, timeout 3s) |
| POST | `/api/v1/auth/register` | Registro |
| POST | `/api/v1/auth/login` | Login |
| POST | `/api/v1/auth/refresh` | Refresh de access token |
| POST | `/api/v1/auth/logout` | Revoca sesión |
| GET | `/api/v1/public/groups/:slug` | Info + documentos de un grupo público |
| POST | `/api/v1/public/groups/:slug/chat` | Chat público contra los documentos del grupo |
| GET | `/api/v1/public/groups/:slug/documents/:doc_id/download` | Descargar documento |
| GET | `/api/v1/public/groups/:slug/documents/:doc_id/view` | Ver documento |

### Protegidas (Bearer JWT)

| Método | Ruta | Descripción |
|--------|------|-------------|
| GET | `/api/v1/users/me` | Perfil del usuario |
| GET | `/api/v1/users/me/stats` | Estadísticas de uso |
| POST | `/api/v1/ask` | Pregunta one-shot al RAG |
| GET | `/api/v1/stream` | Conexión SSE (eventos de documentos/chat en vivo) |
| GET | `/api/v1/config/system-prompt` | Ver system prompt |
| PUT | `/api/v1/config/system-prompt` | Editar system prompt |
| GET | `/api/v1/documents/select` | Listado corto (dropdown, excluye fallidos) |
| GET | `/api/v1/documents` | Listado paginado |
| POST | `/api/v1/documents` | Subir documento |
| DELETE | `/api/v1/documents/:id` | Eliminar documento |
| GET | `/api/v1/documents/:id/steps` | Historial de pasos de procesamiento |
| GET | `/api/v1/documents/:id/health` | Estado + steps + conteo de chunks en Qdrant |
| POST | `/api/v1/documents/:id/reprocess` | Reprocesar (opcional `force_ocr`) |
| POST | `/api/v1/groups` | Crear grupo |
| GET | `/api/v1/groups` | Listar grupos |
| GET | `/api/v1/groups/:id` | Detalle de grupo |
| GET | `/api/v1/groups/:id/usage` | Uso de LLM del grupo |
| PATCH | `/api/v1/groups/:id` | Editar grupo |
| DELETE | `/api/v1/groups/:id` | Eliminar grupo |
| POST | `/api/v1/groups/:id/documents` | Agregar documento al grupo |
| DELETE | `/api/v1/groups/:id/documents/:doc_id` | Quitar documento del grupo |
| POST | `/api/v1/chats/send` | Enviar mensaje (sin streaming) |
| POST | `/api/v1/chats/send-stream` | Enviar mensaje con streaming SSE |
| GET | `/api/v1/chats` | Listar sesiones de chat |
| GET | `/api/v1/chats/:id` | Detalle de sesión |
| PATCH | `/api/v1/chats/:id` | Renombrar sesión |
| DELETE | `/api/v1/chats/:id` | Eliminar sesión |

---

## Worker (procesamiento en background)

`internal/worker/ingest_worker.go` — un solo tipo de job: ingesta de documentos.

| Parámetro | Valor | Notas |
|---|---|---|
| Poll interval | 10s | busca documentos en `pending` |
| Concurrencia | 3 | goroutines simultáneas, gateadas por channel |
| Reintentos máx. | 3 | luego pasa a `failed` |
| Timeout de parseo/OCR | 10 min | por documento, cubre subprocess `tesseract`/`pdftoppm` |
| Watchdog — intervalo | 2 min | revisa documentos atascados |
| Watchdog — umbral | 15 min | doc en `processing` más de esto vuelve a `pending` |

---

## Frontend

React 19 SPA, servida por nginx (imagen propia, puerto `8080`), habla contra el backend en `:4000`.

Stack: TanStack Router (file-based) + TanStack Query, Zustand, Tailwind v4, `react-hook-form` + `zod`, `axios`, `react-markdown`, GSAP, `lucide-react`.

```
frontend/src/
├── routes/
│   ├── index.tsx                          # landing
│   ├── (auth)/login.tsx, register.tsx     # auth, sin login requerido
│   ├── _authenticated.tsx                 # guard/layout de rutas protegidas
│   ├── _authenticated.dashboard.tsx
│   ├── _authenticated.documents.tsx
│   ├── _authenticated.groups(.index/.{$id}).tsx
│   ├── _authenticated.chat.tsx
│   ├── _authenticated.settings.tsx
│   └── c.$slug.tsx                        # chat PÚBLICO por slug de grupo
├── components/{documents,groups,landing,layout,ui}/
├── hooks/ · lib/ · store/ · types/
```

Setup:

```bash
cd frontend
pnpm install
pnpm dev
```

---

## Despliegue

`docker-compose.yml` levanta 3 servicios; **Qdrant y MinIO corren aparte** (otro stack, NAS, bare-metal en el homelab) — no están en este compose.

```mermaid
graph LR
    subgraph "docker compose"
        FE["frontend\n(nginx :8080)"]
        APP["app\n(Go :4000)"]
        PG[("postgres\n:5433 → 5432")]
    end
    subgraph "Externo / Homelab"
        QD[("Qdrant\n:6334")]
        MI[("MinIO / S3")]
    end

    FE -->|API calls| APP
    APP -->|GORM| PG
    APP -->|gRPC| QD
    APP -->|S3 API| MI
```

`app` expone `healthcheck` sobre `GET /health` (el health check profundo descrito arriba).

---

## Configuración

Variables relevantes de `.env` (ver `internal/infrastructure/config/config.go`):

```bash
# Server
PORT=4000
HOST=0.0.0.0
ENV=production

# Postgres
DATABASE_URL="postgres://rago:rago@localhost:5433/rago"

# Qdrant (externo)
QDRANT_HOST="192.168.1.21"
QDRANT_PORT=6334
QDRANT_COLLECTION="default"

# MinIO (externo)
MINIO_ENDPOINT="192.168.1.21:9000"
MINIO_ROOT_USER="..."
MINIO_ROOT_PASS="..."
MINIO_BUCKET="rago"
MINIO_USE_SSL=false

# LLM / Embeddings (OpenRouter, API compatible OpenAI)
OPEN_ROUTER_API="sk-or-..."
OPEN_ROUTER_BASE_URL="https://openrouter.ai/api/v1"
LLM_MODEL="google/gemini-2.5-flash"
EMBEDDING_MODEL="text-embedding-3-small"

# Auth
JWT_SECRET="..."
ACCESS_TOKEN_EXPIRATION=15m
REFRESH_TOKEN_EXPIRATION=720h

# Chat
CHAT_HISTORY_LIMIT=10
CONTEXT_WINDOW_LIMIT=8192   # nota: es un límite de RESULTADOS de búsqueda, no de tokens
```

> `GOTENBERG_URL` puede aparecer en algún `.env` viejo — no se usa en el código actual, ignorar.

---

## Estado Actual

### ✅ Implementado

- Arquitectura hexagonal (core/ports/application/infrastructure)
- API REST completa (auth, documentos, grupos, chats, público)
- Ingesta asíncrona con worker, reintentos y watchdog
- OCR (tesseract) para imágenes y PDFs escaneados
- Parsers: TXT, MD, CSV, JSON, DOCX, XLSX, PDF, imágenes
- Vector store Qdrant + storage MinIO
- Chat con historial y streaming SSE
- Chat público por grupo (sin cuenta), con umbral de confianza
- JWT (access + refresh) con revocación vía sesiones en Postgres
- Health check profundo (Postgres + Qdrant + MinIO)
- Frontend React (dashboard, documentos, grupos, chat, settings)


---

## Glosario

| Término | Definición |
|---------|-------------|
| **Embedding** | Representación numérica de texto |
| **Chunk** | Fragmento de documento |
| **Vector Store** | Base de datos para vectores (Qdrant) |
| **Blob Storage** | Almacenamiento de archivos crudos (MinIO/S3) |
| **Port** | Interfaz en `core/ports` que define un contrato sin implementación |
| **Adapter** | Implementación concreta de un port en `infrastructure` |
| **SSE** | Server-Sent Events, canal unidireccional servidor→cliente para progreso en vivo |
| **Watchdog** | Proceso que detecta y recupera documentos atascados en `processing` |
| **RAG** | Retrieval Augmented Generation |

---

## Referencias

- [Qdrant Documentation](https://qdrant.tech/documentation/)
- [LangChain Go](https://tmc.github.io/langchaingo/)
- [OpenRouter Docs](https://openrouter.ai/docs)
- [MinIO Go SDK](https://pkg.go.dev/github.com/minio/minio-go/v7)
- [Gin Web Framework](https://gin-gonic.com/)
- [GORM](https://gorm.io/)
- [TanStack Router](https://tanstack.com/router)
