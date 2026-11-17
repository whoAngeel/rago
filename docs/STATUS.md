# Estado del trabajo de mejora del RAG

Última actualización: 2026-11-16 (noche). Este documento manda sobre `docs/rag-roadmap.md` cuando difieran: el roadmap se escribió antes de las decisiones de OCR, CAS y formatos.

Para retomar: leer secciones 1 y 2, ejecutar el paso 0 de la sección 5 y seguir con la tanda 3.

## 1. Dónde estamos

| Tanda | Qué | Estado |
|---|---|---|
| 1 | Top-k, umbral, fuga del chat público, rol de sistema, prompt único | Hecha, commiteada en `main` |
| 2 | Evaluación de recuperación (`cmd/eval`) y stack local | Hecha, commiteada en `main` |
| 2b | Evaluación ampliada: PDF, acierto por fragmento, corpus difícil | Hecha, commiteada en `main` (`4136bb3`) |
| 3 | Ingesta: formatos, extracción PDF, OCR, CAS, contexto en chunks | Diseñada, **sin empezar** |
| 4 | Reescritura de consulta, búsqueda híbrida, reranker | Pendiente |

Tandas 1, 2 y 2b subidas a `origin/main` y desplegadas el 2026-11-17 (workflow «Deploy to Homelab» en verde para `49afc87` y `9cd9a8d`).

### Fuera de los commits (a propósito)

`AGENTS.md`, `CLAUDE.md`, `.atl/`, `.cursor/`, `.gemini/` y los cambios de diagramas Mermaid de `README.md` son ediciones del propietario, ajenas a este trabajo. `testdata/private/` está ignorada por git y contiene los PDF reales: no commitear nunca.

## 2. Cómo arrancar

```
./scripts/dev-local.sh          # contenedores locales + servidor en http://localhost:4004
set -a; . scripts/local-env.sh; set +a; go run ./cmd/eval --out /tmp/eval.json
go run ./cmd/eval --dump-chunks --data-dir testdata/private   # ver qué extrae el parser, sin red
```

- Contenedores propios (proyecto `rago-local`): postgres 5433, qdrant 6333/6334, minio 9000/9001. No tocar los contenedores `scory-production-*` ni los demás de la máquina.
- `scripts/local-env.sh` pisa las variables de servicios de datos; la clave de OpenRouter y el secreto JWT siguen viniendo de `.env`.
- La base local arranca vacía: crear usuario y volver a subir documentos.
- Guías: `docs/local-dev.md`, `docs/eval.md`.

## 3. Changelog

### Tanda 1 — commits `774569e`, `c19ab09`
- El chat pedía hasta 8192 fragmentos a Qdrant (`CONTEXT_WINDOW_LIMIT` usado como top-k). Ahora `RETRIEVAL_TOP_K` (8).
- Umbral `RETRIEVAL_SCORE_THRESHOLD` en chat y `/ask`.
- Seguridad: un grupo público sin documentos buscaba en los documentos de todos los usuarios. Corregido en el caso de uso y con una guardia en el adaptador de Qdrant.
- El LLM recibe un mensaje de sistema real; `LLM_TEMPERATURE` (0.2), `LLM_MAX_TOKENS` (0 = no se envía).
- Un único prompt y un único formato de contexto (`internal/application/prompt.go`).
- La fila `system_prompt` de la base de datos se actualiza al arrancar solo si es idéntica al prompt anterior.

### Tanda 2 — commits `d46dc99`, `537e267`, `00cfb67`, `2c2a322`
- `go run ./cmd/eval`: ingesta en la colección `rago_eval`, informa hit@k, MRR y barrido de umbrales. Se niega a ejecutarse si Qdrant no es local.
- Umbral por defecto bajado de 0.30 a 0.25 (0.30 perdía respuestas válidas).
- Stack local: `scripts/dev-local.sh`, `scripts/local-env.sh`, `docker-compose.local.yml` (imagen de MinIO cambiada a `cgr.dev/chainguard/minio:latest` porque `minio/minio` ya no se puede descargar).

### Tanda 2b — commits `4136bb3`, `b72972d`
- La evaluación ingesta PDF con los mismos parsers que el servidor, varias carpetas y varios archivos de preguntas.
- Campo `expected_contains`: acierto a nivel de fragmento, con normalización NFKC, minúsculas y espacios.
- `--dump-chunks`: muestra los fragmentos sin conectarse a nada.
- Corpus difícil: 12 manuales sintéticos casi idénticos entre sí (31 preguntas) y 24 preguntas sobre los 6 PDF reales de `testdata/private/`.

## 4. Línea base medida (2026-11-17, top-k 8, 81 preguntas)

Medición hecha tras restringir los formatos: 69 preguntas con respuesta y 12 sin respuesta.

| Nivel | hit@1 | hit@3 | hit@8 | MRR |
|---|---|---|---|---|
| Documento correcto | 57/69 (83 %) | 65/69 (94 %) | 69/69 (100 %) | 0.882 |
| Fragmento correcto | 38/69 (55 %) | 58/69 (84 %) | 65/69 (94 %) | 0.696 |
| Fragmento correcto, solo PDF reales privados | 7/20 | 13/20 | 16/20 | 0.509 |

Umbral 0.25: conserva 66/72 respuestas; filtra solo 1 de 12 preguntas sin respuesta.

Qué dicen los números:
1. **La extracción de PDF rompe el texto.** Los 4 fallos totales son de los PDF reales: `for3seconds` (palabras pegadas), `455. 1` (decimal partido), `}ERTY` por `AZERTY`, `3 440x1 440`. La librería actual es `ledongthuc/pdf`.
2. **Los documentos parecidos se confunden.** El documento correcto siempre está en el top 8, pero a menudo no primero: los fragmentos no dicen de qué modelo hablan. Las preguntas entre idiomas caen hasta el puesto 6.
3. **El umbral no detecta preguntas sin respuesta.** Las cercanas al tema puntúan hasta 0.646. Eso lo resuelven el prompt y un reranker, no el umbral.

Repetir la evaluación antes y después de cada cambio y comparar contra esta tabla.

## 5. Tanda 3: plan y decisiones

### Paso 0 (al retomar)
- [x] Corregir un defecto menor del informe: las preguntas privadas y las originales se llaman ambas `questions.json` y sus totales por archivo salen sumados. Etiquetar por ruta relativa (`internal/eval/report.go`, `dataset.go`). Modelo: Haiku.

### Paso 1 — seguridad y formatos (Sonnet)
- [x] **Borrado sin comprobar dueño.** Corregido en `49afc87`: `DeleteDocument` recibe el usuario y devuelve 404 si el documento no existe o es de otro. Desplegado el 2026-11-17.
- [x] **Solo `.txt`, `.md` y `.pdf`** (decisión del propietario). Tres listas que cambian juntas: mapa de extensiones en `handlers/document.go`, registro de parsers en `cmd/server/main.go`, y `internal/eval/registry.go`. Quitar csv, json, docx, xlsx e imágenes sueltas.
  - No borrar `internal/infrastructure/parser/image.go`: el OCR de PDF lo usa.
  - Fijar el tipo de contenido por extensión en el servidor (hoy se confía en la cabecera del navegador) y comprobar que un `.pdf` empieza por `%PDF-`.
  - Frontend: `frontend/src/components/documents/DropZone.tsx` (añadir `accept` y corregir la etiqueta de formatos).
  - Quitar de la evaluación `employees.csv`, `reports.json` y sus preguntas.
  - Los documentos de esos tipos ya ingestados no se borran.

### Paso 2 — extracción de PDF y OCR (Sonnet; diseño ya decidido)
Por página:
1. [x] Extraer texto con `pdftotext` (poppler, ya instalado) en lugar de `ledongthuc/pdf` (dependencia retirada).
2. [x] Limpiar: unir palabras cortadas con guion, quitar cabeceras y pies repetidos, normalizar NFKC.
3. [x] Si la página queda vacía o con texto basura, OCR de esa página a 300 DPI.

OCR detrás de un puerto (interfaz) con dos implementaciones, elegible por configuración:
- [x] Puerto `ports.OCR` hecho (`internal/core/ports/ocr.go`).
- [x] Tesseract local hecho (`parser.NewTesseractOCR`, a 300 DPI). Falta el modo de segmentación.
- [x] LLM multimodal por OpenRouter hecho (`openrouter.NewOCRAdapter`, imagen en base64 al endpoint de chat; con Tesseract de respaldo si falla).
- Con `OCR_ENGINE=llm`, las imágenes de las páginas que necesitan OCR se envían a OpenRouter.

Candidatos para el OCR con LLM (precios del catálogo público de OpenRouter, consultados el 2026-11-17 UTC; coste por 1000 páginas estimado):

| Modelo | Entrada / salida ($/M tokens) | ~$/1000 pág | Nota |
|---|---|---|---|
| `qwen/qwen3.7-flash` | 0.03 / 0.13 | 0.2–0.4 | El más barato; los tokens de razonamiento pueden inflar el coste |
| `google/gemini-2.5-flash-lite` | 0.10 / 0.40 | ~0.4 | Modelo antiguo |
| `openai/gpt-6-luna` | 0.10 / 0.50 | ~0.5 | Regla de tokens de imagen sin verificar |
| `anthropic/claude-haiku-5.5` | 0.10 / 0.50 | 0.6–0.85 | Coste calculado con regla documentada |
| Motor PDF `mistral-ocr` de OpenRouter | — | 2.00 | Único OCR especializado; recibe el PDF directo |

**No hay benchmark público que compare la calidad de OCR de estos modelos.** Elegir con una prueba local. Los modelos gratuitos no son aptos para manuales privados mientras no se confirme su política de datos.

Prueba local antes de elegir:
- [ ] Los 6 PDF reales tienen capa de texto, así que nunca pasan por OCR. Rasterizarlos a imagen (`pdftoppm`) para crear versiones escaneadas con texto correcto conocido.
- [ ] Comparar: capa de texto, Tesseract 300 DPI y cada LLM. Medir aciertos exactos de números y unidades, filas de tabla, texto inventado en zonas en blanco, coste real por página y la evaluación de recuperación.
- [ ] Caso real a vigilar: la ficha del monitor tiene la fuente rota en el propio PDF (`energéca` sin «ti»); ahí el OCR puede dar mejor resultado que la capa de texto.

Medición en los 6 PDF privados (nivel chunk, 20 preguntas con respuesta, top-k 8, 2026-11-17):

| Fuente del texto | hit@1 | hit@3 | hit@8 | MRR |
|---|---|---|---|---|
| Capa de texto | 8 | 12 | 16 | 0.518 |
| Tesseract forzado | 7 | 11 | 14 | 0.458 |
| LLM forzado (`google/gemini-2.5-flash-lite`) | 9 | 16 | 20 | 0.637 |

Decisión: por defecto `OCR_ENGINE=llm` con `PDF_OCR_MODE=always`. Cada imagen de página de un PDF se envía a OpenRouter; la capa de texto es el respaldo si el LLM falla o devuelve vacío (Tesseract solo cubre las páginas sin capa de texto). Para mantener las páginas en local: `PDF_OCR_MODE=auto` u `OCR_ENGINE=tesseract`. Límites de la medición: 6 PDF y un solo modelo probado.

Tests pendientes:
- Ruta `ForceOCR`, OCR nulo, error de OCR, entrada que no es PDF y fallos de `pdftotext`/`pdftoppm`.
- Tesseract real, `stripCodeFence` y el error "no choices".
- Selección de `NewEngine`/`PDFOptions` y el flag `--force-ocr`.
- Caché (paso 3): acierto/fallo de la caché de parseo en el worker, `ForceOCR` sin escritura, repositorios GORM de ambas cachés, `ComputeEmbeddings` con aciertos parciales, vector de longitud distinta como fallo, respuesta 200 `duplicate` del handler y `RequeueDocument`.

Riesgos del OCR con LLM: inventa texto en zonas borrosas, omite filas de tablas, reordena columnas. Pedir transcripción literal, tablas en Markdown y `[ilegible]` cuando no se lea.

### Paso 3 — almacenamiento por contenido, CAS (Sonnet; **decisión de arquitectura** ya tomada)
Hoy solo existe un rechazo de duplicados: índice único `(user_id, checksum)`, respuesta 409, clave de MinIO `<usuario>/<doc>/<nombre>`, sin caché; reprocesar repite descarga, parseo, OCR y embeddings.

Diseño elegido: caché por contenido, con filas de documento por usuario.
- [x] Blob una sola vez en `blobs/<sha256>`; se omite la subida si ya existe. Los objetos antiguos no se mueven. Al borrar, el blob se elimina solo si ningún otro documento referencia el mismo `file_path` (`CountByFilePath`).
- [x] Caché de texto extraído (`ParsedContent`), clave `(content_hash, parser_signature)`; el worker la consulta antes de descargar y parsear. No se guarda el resultado de un parseo con `ForceOCR`.
- [x] Caché de embeddings (`EmbeddingCacheEntry`), clave `(text_hash, model, dimension)`, solo en la ingesta (`CachedEmbedder`); las consultas de chat no pasan por ella. Se omitió la "versión de chunker": la clave es el texto exacto del fragmento.
- [x] Duplicado: `Upload` devuelve el documento existente con `ErrDuplicateDocument` y el handler responde 200 con `"duplicate": true`; si estaba fallido, se reencola (pendiente, reintentos 0) y responde 201.
- [ ] Dos subidas idénticas simultáneas chocan con el índice único y devuelven 500: mapear ese error.
- [ ] Frontend: mostrar el aviso de duplicado (ahora llega 200 con `duplicate: true`; hoy muestra "subido correctamente") (`frontend/src/routes/_authenticated.documents.tsx`).

Cachés compartidas entre usuarios por hash de contenido (no contienen datos de usuario, solo texto y vectores derivados del archivo). Firma del parser: `parser.Version|motor|modelo|modo`. Para invalidar: subir `parser.Version` (cuando cambie la extracción o limpieza), cambiar la configuración de OCR (motor, modelo o modo) o cambiar el modelo/dimensión de embeddings; las filas antiguas quedan sin uso.

Descartado: guardar los vectores una sola vez con una lista de usuarios autorizados. Ahorra espacio, pero toca los filtros de búsqueda y el borrado; un error ahí es una fuga entre usuarios.

Hacer el paso 3 antes de cambiar el chunker: con el texto en caché, reingestar todo no repite el OCR.

### Paso 4 — calidad de los fragmentos (Sonnet)
- [ ] Anteponer título o producto a cada fragmento antes de embeber, y guardar página e índice en el payload (`internal/worker/ingest_worker.go`, `internal/application/ingest_usecase.go`). Ataca la confusión entre modelos parecidos.
- [ ] Fallos de embedding: hoy se omiten fragmentos y el documento queda completado. Agrupar llamadas, reintentar y fallar el documento si falta alguno.
- [ ] Chunker (`internal/infrastructure/chunker/semantic.go`): tamaño en bytes, corta en cada punto. Nuevo tamaño y estrategia: **decisión de arquitectura**, tomarla con la evaluación.

## 6. Pendientes sueltos

- [ ] Confirmar `EMBEDDING_MODEL` y `CONTEXT_WINDOW_LIMIT` en `.env`: nadie pudo leer ese archivo y no se verificó si apunta a producción.
- [ ] Sin test automático: actualización de la fila del prompt al arrancar, streaming contra el proveedor real, guardia de Qdrant, y OCR dentro de la evaluación.
- [ ] `BuildUserMessage` solo escapa `"`; un fragmento con `</contexto>` puede falsificar etiquetas.
- [ ] `.gitignore` tiene el patrón `server`, que ignora archivos nuevos en `cmd/server/`. Cambiarlo a `/server`.
- [ ] `gentle-ai review` marca revisión pendiente; no se ejecutó.
- [ ] Limpiezas menores: ver `docs/rag-roadmap.md`, sección «Limpiezas menores».

## 7. Cómo trabajar (reparto por modelo)

- Fable 5.2: solo arquitecto y revisor. Sonnet 5.5: implementa y audita. Haiku 5.5: busca, ediciones pequeñas y documentos.
- Sin Fable: cambiar con `/model` a Opus 5.5 o Sonnet 5.5 y seguir este documento. Lo marcado como decisión de arquitectura ya tomada se puede implementar; el tamaño del chunker y las decisiones de la tanda 4 esperan al arquitecto.
- Receta por tarea: un subagente por tarea, con el problema, los archivos que puede editar, los tests que debe escribir primero y los comandos (`go build ./...`, `go vet ./...`, `go test ./...`). Después, la sesión principal vuelve a ejecutar los tests y lee el diff. Dos tareas en paralelo solo si tocan archivos distintos.
- Limitaciones del entorno vistas en esta sesión: la lectura y escritura de `.env*` está denegada, y guardar en Engram falla en modo automático porque un hook modifica la llamada.
