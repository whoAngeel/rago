#!/usr/bin/env bash
# Levanta el stack local (postgres, qdrant, minio) y arranca el servidor Go
# usando las variables de scripts/local-env.sh, que sobrescriben a .env.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

COMPOSE=(docker compose -p rago-local -f docker-compose.local.yml)
ENV_FILE="scripts/local-env.sh"

if [[ ! -f "$ENV_FILE" ]]; then
  echo "Falta $ENV_FILE en $REPO_ROOT" >&2
  exit 1
fi

"${COMPOSE[@]}" up -d

echo "Esperando a que postgres esté healthy..."
pg_container="$("${COMPOSE[@]}" ps -q postgres)"
if [[ -z "$pg_container" ]]; then
  echo "No se encontró el contenedor de postgres" >&2
  exit 1
fi
for _ in $(seq 1 60); do
  status="$(docker inspect --format '{{.State.Health.Status}}' "$pg_container" 2>/dev/null || echo unknown)"
  if [[ "$status" == "healthy" ]]; then
    break
  fi
  sleep 2
done
if [[ "$status" != "healthy" ]]; then
  echo "postgres no llegó a healthy (estado: $status)" >&2
  exit 1
fi

set -a
# shellcheck disable=SC1090
. "./$ENV_FILE"
set +a

# Muestra solo hosts y puertos; nunca credenciales.
printf 'postgres: %s\n' "$(printf '%s' "$DATABASE_URL" | sed -E 's#^[^@]*@##; s#/.*$##; s#\?.*$##')"
printf 'qdrant:   %s:%s (collection %s)\n' "$QDRANT_HOST" "$QDRANT_PORT" "$QDRANT_COLLECTION"
printf 'minio:    %s (bucket %s)\n' "$MINIO_ENDPOINT" "$MINIO_BUCKET"

exec go run ./cmd/server
