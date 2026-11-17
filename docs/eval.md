# Evaluación de recuperación

`go run ./cmd/eval` mide solo la **recuperación**: dada una pregunta, si vuelve el documento correcto, en qué posición y con qué puntaje. No llama al LLM de chat, así que no evalúa la calidad de la respuesta redactada ni el prompt.

Solo necesita Qdrant y el proveedor de embeddings (no usa Postgres, MinIO ni el servidor). Genera peticiones de embeddings que se facturan a la API key configurada.

## Ejecución contra el stack local

```sh
set -a; . ./scripts/local-env.sh; set +a
go run ./cmd/eval --out /tmp/eval.json
```

En cada ejecución se vuelven a ingestar los documentos de `testdata/` en la colección `rago_eval` con un usuario sintético, usando los mismos parsers, el mismo chunker (1000/200) y `IngestUsecase` que el worker. Nunca se toca `QDRANT_COLLECTION`.

## Flags

- `--top-k N`: chunks por pregunta (por defecto `RETRIEVAL_TOP_K`).
- `--out ruta`: escribe además el reporte en JSON.
- `--skip-ingest`: reutiliza lo que ya hay en `rago_eval`.
- `--allow-remote`: permite un `QDRANT_HOST` que no sea local.
- `--questions ruta` y `--data-dir ruta`: dataset y documentos (por defecto `testdata/eval/questions.json` y `testdata`).

## Cómo leer el reporte

1. **Por pregunta**: posición del primer chunk del documento esperado (`miss` si no apareció), su puntaje y el puntaje más alto devuelto. Las preguntas sin respuesta muestran `(none)`.
2. **Totales** sobre las preguntas con respuesta: hit@1, hit@3, hit@k y MRR (media del recíproco de la posición; una falla cuenta 0).
3. **Umbral**: para el valor configurado en `RETRIEVAL_SCORE_THRESHOLD` y para 0.20 a 0.45, cuántas preguntas con respuesta conservan su chunk esperado (puntaje >= corte) y cuántas preguntas sin respuesta terminan correctamente sin resultados. Se busca un corte que mantenga alto el primer valor sin bajar el segundo.

Las búsquedas no aplican umbral, de modo que todos los puntajes son visibles.

## Agregar una pregunta

Añade un objeto a `testdata/eval/questions.json` con `id` único, `question`, `expected_source` y `note`:

- `expected_source` es el nombre exacto de un archivo de `testdata/`; déjalo vacío (`""`) si ningún documento responde la pregunta.
- Redacta la pregunta en el idioma del documento y sin copiar una frase literal.
- Marca en `note` las preguntas vagas o de seguimiento.

`go test ./internal/eval` valida el dataset contra `testdata/`: ids duplicados, preguntas vacías o archivos inexistentes rompen el build.

## Protección contra producción

El comando se niega a ejecutarse (código de salida distinto de cero) si `QDRANT_HOST` no es `localhost`, `127.0.0.1` o `::1`, porque crea y reescribe una colección. Si tu `.env` apunta a producción, carga `scripts/local-env.sh` como se indica arriba. Usa `--allow-remote` solo si estás seguro del destino.
