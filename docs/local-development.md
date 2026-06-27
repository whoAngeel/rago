# Local Development — rago

## Prerrequisitos

| Herramienta | Versión |
|---|---|
| [Go](https://go.dev/dl/) | 1.26+ |
| [pnpm](https://pnpm.io/installation) | Última |
| [Docker](https://docs.docker.com/engine/install/) + [Compose](https://docs.docker.com/compose/install/) | Última |
| [Air](https://github.com/air-verse/air) (opcional) | `go install github.com/air-verse/air@latest` |

## 1. Clonar e instalar dependencias

```bash
git clone <repo> rago
cd rago

# Backend
go mod download

# Frontend
cd frontend && pnpm install && cd ..
```

## 2. Configurar variables de entorno

El proyecto lee un archivo `.env` en la raíz. **No hay `.env.example`**; crea uno propio con las siguientes variables mínimas:

```env
# === Requeridas (la app falla si faltan) ===
DATABASE_URL=postgres://rago:rago@localhost:5433/rago
OPEN_ROUTER_API=sk-or-v1-...
QDRANT_HOST=localhost
QDRANT_PORT=6334
MINIO_ENDPOINT=localhost:9000

# === Opcionales con defaults ===
HOST=0.0.0.0
PORT=4004
ENV=development
JWTSECRET=my-secret-key
LLM_MODEL=google/gemini-2.5-flash
EMBEDDING_MODEL=text-embedding-3-small
EMBEDDING_DIMENSION=1536
QDRANT_COLLECTION=default
MINIO_ROOT_USER=minioadmin
MINIO_ROOT_PASS=minioadmin
MINIO_BUCKET=rago
MINIO_USE_SSL=false
MAX_UPLOAD_SIZE=52428800
CHAT_HISTORY_LIMIT=10
CONTEXT_WINDOW_LIMIT=8192
JWT_ACCESS_EXPIRATION=15m
JWT_REFRESH_TOKEN_EXPIRATION=720h
```

## 3. Levantar servicios externos

### PostgreSQL (vía Docker)

```bash
docker compose up postgres -d
```

Puerto host: `5433`, user/pass/db: `rago`/`rago`/`rago`.

### Qdrant (vector DB)

```bash
docker run -d --name qdrant -p 6333:6333 -p 6334:6334 qdrant/qdrant
```

### MinIO (S3-compatible storage)

```bash
docker run -d --name minio -p 9000:9000 -p 9001:9001 \
  -e MINIO_ROOT_USER=minioadmin -e MINIO_ROOT_PASS=minioadmin \
  minio/minio server /data --console-address ":9001"
```

Luego crea un bucket llamado `rago` (o el que configures en `MINIO_BUCKET`).

### OpenRouter (LLM)

Regístrate en [openrouter.ai](https://openrouter.ai) y genera una API key. Colócala en `OPEN_ROUTER_API`.

## 4. Ejecutar en desarrollo

Dos terminales:

```bash
# Terminal 1 — Backend (hot-reload con Air)
air

# Terminal 2 — Frontend (Vite dev server)
cd frontend && pnpm dev
```

O sin Air:

```bash
go run ./cmd/server
```

## 5. Acceder

| Servicio | URL |
|---|---|
| Frontend | http://localhost:5173 |
| API Backend | http://localhost:4004 |
| Health check | `GET http://localhost:4004/health` |
| MinIO Console | http://localhost:9001 |
| Qdrant Dashboard | http://localhost:6333/dashboard |

El frontend en Vite proxydea `/api` al backend (`http://localhost:4004`).

## 6. Tests

```bash
go test ./...
```

## 7. Docker full-stack

```bash
docker compose up --build
```

Esto levanta: `postgres`, `app` (Go server en `:4000`), `frontend` (Nginx en `:8080`). Requiere tener las imágenes construidas previamente o un registro donde se resuelvan.
