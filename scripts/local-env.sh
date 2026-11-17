# Local development environment. Source this file; do not execute it.
#
#   set -a; . scripts/local-env.sh; set +a
#
# It points every data service at the containers from docker-compose.local.yml.
# The config package loads .env without overriding variables that are already
# set, so these values win and secrets (OPEN_ROUTER_API, JWTSECRET) still come
# from .env. Only local, non-secret development values belong here.

ENV=development
PORT=4004 # the Vite dev proxy in frontend/vite.config.ts targets this port

DATABASE_URL="postgres://rago:rago@localhost:5433/rago?sslmode=disable"

QDRANT_HOST=localhost
QDRANT_PORT=6334
QDRANT_COLLECTION=rago_local

MINIO_ENDPOINT=localhost:9000
MINIO_ROOT_USER=minioadmin
MINIO_ROOT_PASS=minioadmin
MINIO_BUCKET=rago-local
MINIO_USE_SSL=false
