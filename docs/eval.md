# Evaluación de recuperación

`go run ./cmd/eval` mide solo la **recuperación**: dada una pregunta, si vuelve el documento correcto (y el fragmento correcto), en qué posición y con qué puntaje. No llama al LLM de chat, así que no evalúa la redacción de la respuesta ni el prompt.

Solo necesita Qdrant y el proveedor de embeddings (no usa Postgres, MinIO ni el servidor). Genera peticiones de embeddings que se facturan a la API key configurada.

## Ejecución contra el stack local

```sh
set -a; . ./scripts/local-env.sh; set +a
go run ./cmd/eval --out /tmp/eval.json
```

En cada ejecución se vuelven a ingestar los documentos en la colección `rago_eval` con un usuario sintético, con los mismos parsers que el servidor (TXT, MD y PDF con respaldo de OCR), el mismo chunker que el servidor (por defecto `recursive` 1500/200 con cabecera de contexto) y `IngestUsecase`. Cada archivo se interpreta con un timeout de 10 minutos, como en el worker. Nunca se toca `QDRANT_COLLECTION`.

## Documentos y preguntas por defecto

- Documentos: los archivos directamente dentro de `testdata/` y, si existen, `testdata/manuals/` y `testdata/private/` (sin recursión; los `*.json` de `private/` se ignoran).
- Preguntas: todos los `*.json` de `testdata/eval/` y, si existe, `testdata/private/questions.json`.
- `testdata/private/` está en `.gitignore`: guarda documentos reales y no se versiona. **Al ejecutar la evaluación, esos documentos se envían al proveedor de embeddings.**
- Un documento se identifica por su nombre de archivo; el mismo nombre en dos directorios es un error. Si dos archivos tienen contenido idéntico, solo se ingesta uno y se imprime una advertencia con ambos nombres.

## Flags

- `--top-k N`: chunks por pregunta (por defecto `RETRIEVAL_TOP_K`).
- `--out ruta`: escribe además el reporte en JSON.
- `--skip-ingest`: reutiliza lo que ya hay en `rago_eval`.
- `--allow-remote`: permite un `QDRANT_HOST` que no sea local.
- `--data-dir ruta` (repetible): reemplaza la lista de directorios por defecto.
- `--questions ruta` (repetible): reemplaza la lista de archivos de preguntas por defecto. Los `id` deben ser únicos entre todos los archivos.
- `--dump-chunks`: parsea y trocea los documentos como lo haría la ingesta e imprime cada chunk (documento, índice y longitud en caracteres). No usa Qdrant, embeddings ni API key.
- `--force-ocr`: rasteriza y pasa por OCR todas las páginas de los PDF en lugar de usar su capa de texto; aplica a la ingesta y a `--dump-chunks`. Con `--dump-chunks` siempre usa Tesseract; en la ingesta usa el motor de `OCR_ENGINE`.
- La evaluación sigue `OCR_ENGINE` y `PDF_OCR_MODE`: con los valores por defecto envía las imágenes de las páginas de los PDF a OpenRouter.
- `--source nombre`: con `--dump-chunks`, limita la salida a un documento.
- `--chunker fixed|recursive`: estrategia de chunking (por defecto `CHUNKER` de la configuración). `recursive` corta por párrafo, línea, fin de frase y espacio, en ese orden. Aplica a la ingesta y a `--dump-chunks`.
- `--chunk-size N` y `--chunk-overlap N`: tamaño y solapamiento en caracteres (por defecto `CHUNK_SIZE` y `CHUNK_OVERLAP` de la configuración). Con los valores por defecto los chunks son idénticos a los del servidor.
- `--context-header`: (por defecto `CHUNK_CONTEXT_HEADER`; usa `--context-header=false` para quitarla) antepone a cada chunk una línea `[título — p. N]` (`[título]` si el documento no tiene páginas) antes de calcular el embedding. El título es la primera línea no vacía de la primera página, o el nombre del archivo si está vacía.
- `--parse-cache-dir ruta`: carpeta con las páginas ya parseadas (por defecto `<directorio de caché del usuario>/rago-eval/parse`; una cadena vacía desactiva la caché). La clave combina el contenido del archivo, la versión del parser, el motor y modelo de OCR, el modo de OCR de PDF y `--force-ocr`. Al terminar la ingesta se imprime cuántos archivos se parsearon y cuántos vinieron de la caché. Parsear PDF con un LLM tiene costo, por eso conviene no repetirlo al probar chunking.

## Experimentos de chunking

Los valores por defecto de `--chunker`, `--chunk-size`, `--chunk-overlap` y `--context-header` salen de la configuración del servidor (`CHUNKER`, `CHUNK_SIZE`, `CHUNK_OVERLAP`, `CHUNK_CONTEXT_HEADER`), así que `go run ./cmd/eval` sin flags mide lo que hace el servidor; los flags explícitos los reemplazan. `--dump-chunks` no lee la configuración y usa los mismos valores integrados (recursive, 1500, 200, cabecera activa). Cada ejecución imprime en stderr la configuración de chunking que usa. Compara variantes contra el valor base (el default del servidor; el anterior era `fixed` 1000/200, que se reproduce con `--chunker fixed --chunk-size 1000 --context-header=false`) guardando un reporte por variante; la caché de parseo evita volver a llamar al OCR:

```sh
go run ./cmd/eval --chunker recursive --chunk-size 800 --chunk-overlap 120 --out /tmp/recursive.json
go run ./cmd/eval --chunker recursive --chunk-size 800 --chunk-overlap 120 --context-header --out /tmp/recursive-header.json
```

Para revisar los chunks antes de ingestar, añade `--dump-chunks --source nombre` con las mismas opciones.

## Dos niveles de acierto

- **Archivo**: hay un chunk de `expected_source` entre los resultados. Es lo que mide `rank`.
- **Chunk**: el campo opcional `expected_contains` (lista de cadenas cortas). Un chunk acierta si viene de `expected_source` y contiene todas las cadenas, comparando tras normalizar ambos lados: Unicode NFKC (las ligaduras de PDF como `ﬁ` equivalen a `fi`), minúsculas y espacios colapsados. Las tildes se conservan.

Las preguntas sin `expected_contains` solo cuentan a nivel de archivo; los denominadores a nivel de chunk indican cuántas preguntas cubren.

Para escribir `expected_contains`, ejecuta `--dump-chunks` y elige cadenas distintivas (un código de modelo, un número con su unidad, el nombre de un botón) que existan en un solo chunk. Evita frases largas: pueden quedar partidas entre dos chunks. Si el parser pierde o deforma el texto, escribe la cadena tal como aparece en el PDF: la pregunta fallará y eso es lo que la evaluación debe mostrar.

## Cómo leer el reporte

1. **Por pregunta**: archivo de origen, posición y puntaje del primer chunk del documento esperado, posición y puntaje del primer chunk que cumple `expected_contains` (`miss` si no apareció, `-` si no aplica) y el puntaje más alto devuelto. Las preguntas sin respuesta muestran `(none)`.
2. **Totales** global y por archivo de preguntas, a ambos niveles: hit@1, hit@3, hit@k y MRR (una falla cuenta 0). Así se leen por separado los manuales privados y los conjuntos sintéticos.
3. **Umbral**: para `RETRIEVAL_SCORE_THRESHOLD` y de 0.20 a 0.45, cuántas preguntas con respuesta conservan su acierto (a nivel de chunk si la pregunta tiene `expected_contains`, si no a nivel de archivo) y cuántas sin respuesta terminan sin resultados.

Las búsquedas no aplican umbral, de modo que todos los puntajes son visibles.

## Agregar una pregunta

Añade un objeto a un archivo de `testdata/eval/` (o a `testdata/private/questions.json`) con `id` único, `question`, `expected_source`, `note` y, si procede, `expected_contains`:

- `expected_source` es el nombre exacto de un documento; déjalo vacío (`""`) si ningún documento responde.
- Redacta la pregunta como la haría un usuario, en el idioma del documento, sin copiar una frase literal.
- Marca en `note` las preguntas vagas, de seguimiento o entre idiomas.

`go test ./internal/eval` valida los archivos que existan: ids duplicados, preguntas vacías, documentos inexistentes y, fuera de `private/`, cadenas de `expected_contains` que no aparecen en ningún chunk.

## Protección contra producción

El comando se niega a ejecutarse (código de salida distinto de cero) si `QDRANT_HOST` no es `localhost`, `127.0.0.1` o `::1`, porque crea y reescribe una colección. Si tu `.env` apunta a producción, carga `scripts/local-env.sh` como se indica arriba. Usa `--allow-remote` solo si estás seguro del destino.
