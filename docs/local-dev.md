# Entorno local de desarrollo

## Qué es

Un stack local con sus propios contenedores: PostgreSQL, Qdrant y MinIO, definido en `docker-compose.local.yml` con el proyecto `rago-local`. Permite desarrollar y experimentar sin tocar los servicios de producción.

Las variables de infraestructura están en `scripts/local-env.sh`, que sobrescribe a `.env` en ejecuciones locales. Las claves externas (`OPEN_ROUTER_API`, `JWTSECRET`) siguen tomándose de `.env`.

## Puesta en marcha

```bash
./scripts/dev-local.sh
```

El script:

1. Levanta los contenedores con `docker compose -p rago-local -f docker-compose.local.yml up -d`.
2. Espera a que PostgreSQL esté `healthy`.
3. Carga `scripts/local-env.sh` y muestra una línea por servicio con el host que usará.
4. Ejecuta `go run ./cmd/server` (puerto 4000 por defecto).

## Detener sin perder datos

```bash
docker compose -p rago-local -f docker-compose.local.yml stop
```

No uses `down -v`: elimina los volúmenes.

## Dónde están los datos

Volúmenes con nombre de Docker:

- `rago-local_pg_data_local`: base de datos PostgreSQL.
- `rago-local_qdrant_data_local`: colecciones de vectores.
- `rago-local_minio_data_local`: archivos subidos.

## Estado inicial

La base de datos local empieza vacía. Hay que crear un usuario y volver a subir los documentos. Los datos de producción no se copian.

## Verificar que no se usa producción

- Al arrancar, el script imprime los hosts: `postgres: localhost:5433`, `qdrant: localhost:6334`, `minio: localhost:9000`.
- En los logs del servidor, el arranque muestra `mode=development`.
- Los hosts de `scripts/local-env.sh` deben ser `localhost`. Si una variable de datos no aparece en `scripts/local-env.sh`, se toma de `.env`.

## Frontend

El frontend (`frontend/`) usa una URL base vacía (`frontend/src/lib/constants.ts`) y Vite redirige `/api` mediante proxy. El destino del proxy está fijo en `frontend/vite.config.ts` como `http://localhost:4004`, no como 4000. No hay variable de entorno de Vite para cambiarlo.

Para usar el servidor en `http://localhost:4000`, hay dos opciones:

- Arrancar el servidor con `PORT=4004` (`PORT=4004 ./scripts/dev-local.sh`). Esto no requiere cambios de código.
- Cambiar el `target` del proxy en `frontend/vite.config.ts`.

Después, desde `frontend/`:

```bash
npm run dev
```
