# Estado del trabajo de mejora del RAG

Última actualización: 2026-11-17. **Este es el documento de entrada para cualquier agente o persona que retome el trabajo.** Manda sobre `docs/rag-roadmap.md` y sobre los demás documentos de `docs/` cuando difieran (muchos son de fases anteriores del proyecto).

Orden de lectura: secciones 1 a 3 para entender el contexto, sección 4 para saber qué hacer después, sección 8 para saber cómo trabajar.

## 1. Contexto en cinco minutos

**Qué es rago.** Un sistema RAG (preguntas sobre documentos propios). Backend en Go con arquitectura hexagonal (`internal/core/ports` = interfaces, `internal/core/domain` = modelos, `internal/application` = casos de uso, `internal/infrastructure` = adaptadores, `internal/worker` = procesamiento en segundo plano, `cmd/server/main.go` = cableado). Frontend en React (`frontend/`). Servicios: Postgres (GORM, migraciones con `AutoMigrate`), Qdrant (vectores), MinIO (archivos), OpenRouter (LLM de chat, OCR y embeddings).

**Uso previsto.** Soporte de productos: empresas suben manuales, garantías y fichas técnicas (sobre todo PDF) y sus clientes preguntan. Primer mercado: México. Por eso importa la precisión con números, unidades y modelos parecidos.

**Flujo de ingesta (estado actual).**
1. `POST /api/v1/documents/` (`internal/infrastructure/rest/handlers/document.go`): solo `.txt`, `.md`, `.pdf`; el tipo sale de la extensión; un `.pdf` debe empezar por `%PDF-`.
2. `IngestDocumentUsecase.Upload` (`internal/application/ingest_document_usecase.go`): SHA-256 del archivo; si el usuario ya lo tiene, devuelve el existente (`duplicate: true`); el archivo se guarda una vez en MinIO como `blobs/<sha256>`.
3. `IngestWorker` (`internal/worker/ingest_worker.go`): busca el texto en la caché de extracción; si no está, descarga y parsea.
4. Parser de PDF (`internal/infrastructure/parser/pdf.go`): `pdftotext` por página + limpieza (`pdf_clean.go`); por defecto **cada página se transcribe con un LLM multimodal** (OpenRouter) y la capa de texto queda como respaldo.
5. Chunking (`internal/infrastructure/chunker/pages.go`, `recursive.go`): por página, recursivo (párrafo, línea, oración), 1500 runas, solape 200, cada fragmento empieza con `[título — p. N]`.
6. Embeddings en lotes de 32 con caché por texto (`internal/application/ingest_usecase.go`, `cached_embedder.go`) y subida a Qdrant.

**Flujo de consulta.** `internal/application/chat_usecase.go` (chat con sesión), `ask_usecase.go`, `public_chat_usecase.go` (grupos públicos): se embebe la pregunta, se buscan los 8 fragmentos más parecidos del usuario en Qdrant, se filtran por puntaje (0.25) y se arma el prompt (`prompt.go`, contexto en `<contexto>`).

**Cómo se mide la calidad.** `go run ./cmd/eval` (ver `docs/eval.md`). Mide solo recuperación, no la redacción de la respuesta:
- **hit@1 / hit@3 / hit@8**: el fragmento correcto quedó primero / entre los 3 primeros / entre los 8 que recibe el LLM. Si falla hit@8, el LLM nunca ve el dato.
- **MRR**: promedio de 1/puesto (primero = 1, segundo = 0.5, ausente = 0). Resume qué tan arriba queda lo correcto.
- Nivel "documento" (acertó el archivo) y nivel "fragmento" (el fragmento contiene las cadenas esperadas, campo `expected_contains`).
- Corpus: `testdata/*.txt`, `testdata/manuals/` (12 manuales sintéticos casi idénticos), `testdata/private/` (6 PDF reales del propietario, **ignorados por git, nunca commitear**). Preguntas en `testdata/eval/*.json` y `testdata/private/questions.json`.

## 2. Dónde estamos

| Tanda | Qué | Estado |
|---|---|---|
| 1 | Top-k, umbral, fuga del chat público, rol de sistema, prompt único | En `main`, desplegada |
| 2 / 2b | Evaluación de recuperación y stack local | En `main`, desplegada |
| 3 | Formatos, extracción de PDF, OCR con LLM, CAS, chunking | Pasos 1-2 en `main` (merge `28ac76d`). Pasos 3-4 en la rama `feat/cas` |
| 4 | Recuperación: consulta con historial, búsqueda híbrida, reranker | **Pendiente, diseño propuesto en la sección 4** |

Rama `feat/cas` (sin push, sin merge):
- `6278b2b` cachés por contenido (paso 3).
- `31f8162` herramientas de chunking en la evaluación.
- Sin commitear al cerrar esta sesión, si `git status` lo muestra: chunker del servidor configurable, embeddings por lotes, ruta de chunking compartida (paso 4). Build, vet y 188 tests en verde el 2026-11-17.

**Antes de desplegar `feat/cas`** (cada push a `main` despliega producción por `.github/workflows/deploy.yml`):
1. Se crean dos tablas nuevas (`parsed_contents`, `embedding_cache_entries`) con `AutoMigrate`. No se modifica ninguna tabla existente.
2. Los documentos ya ingestados conservan sus fragmentos viejos (fijo 1000/200, sin cabecera, extracción vieja). Para que mejoren hay que reprocesarlos.
3. Con los valores por defecto, **cada página de cada PDF se envía como imagen a OpenRouter**. Para mantener los archivos en local: `PDF_OCR_MODE=auto` u `OCR_ENGINE=tesseract`.
4. Riesgo de tiempo: el worker corta el parseo a los 10 minutos. Con 4 páginas en paralelo, un PDF de 100 páginas entra solo si cada llamada al LLM tarda menos de ~24 s. No se probó con un PDF grande.
5. El frontend todavía muestra "subido correctamente" para un duplicado (ver sección 5).

### Fuera de los commits (a propósito)

`AGENTS.md`, `CLAUDE.md`, `.atl/`, `.cursor/`, `.gemini/` y los diagramas Mermaid de `README.md` son ediciones del propietario. `testdata/private/` contiene los PDF reales.

## 3. Resultados medidos

### Línea base vigente (default del servidor)

Medida el 2026-11-17 con `go run ./cmd/eval` y la configuración por defecto (OCR `llm` `google/gemini-2.5-flash-lite` modo `always`; chunking `recursive` 1500/200 con cabecera). Top-k 8, 81 preguntas: 69 con respuesta, 12 sin respuesta.

| Conjunto | Nivel | hit@1 | hit@3 | hit@8 | MRR |
|---|---|---|---|---|---|
| Todo | Documento | 59/69 (86 %) | 64/69 (93 %) | 69/69 (100 %) | 0.902 |
| Todo | Fragmento | 46/69 (67 %) | 61/69 (88 %) | 68/69 (99 %) | 0.782 |
| Manuales sintéticos parecidos (27) | Fragmento | 17 | 23 | 26 | 0.752 |
| Textos generales (22) | Fragmento | 17 | 20 | 22 | 0.854 |
| PDF reales privados (20) | Fragmento | 12 | 18 | 20 | 0.744 |

Umbral 0.25: conserva 66/69 respuestas y deja vacías solo 1/12 preguntas sin respuesta.

Comparar cada cambio contra esta tabla.

### Evolución en esta tanda (nivel fragmento, top-k 8, 69 preguntas con respuesta)

| Momento | hit@1 | hit@3 | hit@8 | MRR | PDF reales (20) hit@1 / hit@8 / MRR |
|---|---|---|---|---|---|
| Inicio de la tanda 3 | 38 | 58 | 65 | 0.696 | 7 / 16 / 0.509 |
| `pdftotext` + limpieza | 39 | 57 | 65 | 0.700 | 8 / 16 / 0.518 |
| OCR con LLM en todas las páginas | 40 | 60 | 69 | 0.733 | 9 / 20 / 0.633 |
| + chunking recursivo 1500/200 con cabecera | 46 | 61 | 68 | 0.782 | 12 / 20 / 0.744 |

Lectura: el fragmento correcto ya llega casi siempre al LLM (hit@8). El problema que queda es el **orden**: solo 2 de cada 3 veces queda primero. Eso es trabajo de la tanda 4.

### OCR: fuente del texto en los 6 PDF reales (nivel fragmento, 20 preguntas)

| Fuente | hit@1 | hit@3 | hit@8 | MRR |
|---|---|---|---|---|
| Capa de texto (`pdftotext`) | 8 | 12 | 16 | 0.518 |
| Tesseract forzado | 7 | 11 | 14 | 0.458 |
| LLM forzado (`google/gemini-2.5-flash-lite`) | 9 | 16 | 20 | 0.637 |

El LLM gana porque transcribe las tablas fila por fila (etiqueta y valor juntos) y lee texto que solo existe en imágenes. Límites: 6 PDF y un solo modelo probado.

### Chunking (69 preguntas con respuesta)

| Configuración | hit@1 | hit@3 | hit@8 | MRR | PDF reales hit@1 / MRR |
|---|---|---|---|---|---|
| fixed 1000/200 (anterior) | 40 | 61 | 69 | 0.734 | 9 / 0.637 |
| fixed 1000/200 + cabecera | 40 | 60 | 68 | 0.732 | 10 / 0.668 |
| recursive 1000/200 | 37 | 54 | 65 | 0.683 | 8 / 0.569 |
| recursive 1000/200 + cabecera | 46 | 63 | 68 | 0.795 | 9 / 0.627 |
| recursive 600/100 + cabecera | 36 | 59 | 68 | 0.698 | 7 / 0.527 |
| recursive 400/80 + cabecera | 37 | 60 | 66 | 0.705 | 8 / 0.571 |
| **recursive 1500/200 + cabecera (elegido)** | 46 | 61 | 68 | 0.782 | 12 / 0.744 |
| recursive 2000/200 + cabecera | 51 | 66 | 68 | 0.841 | 10 / 0.717 |

Se eligió 1500: 2000 puntúa más en total, pero la métrica de fragmento favorece mecánicamente a los fragmentos grandes (contienen más texto esperado) y 1500 es mejor en los PDF reales. Con 20 preguntas la diferencia entre ambos está dentro del ruido; es configurable (`CHUNK_SIZE`).

### Umbral de puntaje

0.25 conserva 66-67 de 69 respuestas pero solo deja vacías 1 de 12 preguntas sin respuesta (las preguntas cercanas al tema puntúan hasta 0.65). **El umbral no sirve para detectar "no está en tus documentos".** Eso depende del prompt (ya pide responder "No encontré información sobre eso en tus documentos.") y de un reranker (tanda 4).

## 4. Plan: lo que sigue (tanda 4)

Estas son **propuestas de diseño**, no decisiones cerradas. Regla para todas: medir con `go run ./cmd/eval` antes y después, y quedarse con el cambio solo si mejora hit@1 o MRR sin bajar hit@8. Implementar de a una, en este orden.

### 4.1 Reescribir la pregunta con el historial (prioridad alta, barato)

- **Problema.** En el chat, la búsqueda usa solo la última pregunta. "¿Y cuánto pesa?" no encuentra nada porque no nombra el producto. El historial se pasa al LLM (`chat_usecase.go`, `historyStr`), pero no a la búsqueda.
- **Propuesta.** Si la sesión tiene historial, antes de embeber pedir a un modelo barato que reescriba la pregunta como autónoma ("¿Cuánto pesa el teclado MX Keys Mini?"). Usar la reescrita solo para buscar; al LLM de respuesta le sigue llegando la original. Configurable (`QUERY_REWRITE`, por defecto activo). Si la reescritura falla, usar la original.
- **Archivos.** `internal/application/chat_usecase.go` (`SendMessage` y `SendStream` duplican ~80 líneas: unificarlas primero), un puerto nuevo o un método en `ports.LLMProvider`.
- **Cómo medir.** La evaluación no tiene preguntas con historial. Añadir un campo opcional `history` a las preguntas y 10-15 preguntas de seguimiento.
- **Costo.** Una llamada extra al LLM por mensaje con historial.

### 4.2 Búsqueda híbrida: vectores + palabras exactas (prioridad alta)

- **Problema.** La búsqueda por vectores falla con tokens exactos: `10 m`, `144 Hz`, `AZERTY`, códigos de modelo (`Z207`, `KC-60`). Son justamente las preguntas de soporte.
- **Propuesta.** Colección nueva en Qdrant con dos vectores por punto: el denso actual y uno disperso tipo BM25. Consulta con la Query API de Qdrant y fusión RRF. Cliente actual: `github.com/qdrant/go-client v1.17.1`.
- **A verificar antes de implementar** (con Context7 o la documentación de Qdrant): si el servidor Qdrant desplegado calcula BM25 por su cuenta o si hay que generar el vector disperso en Go (tokenizar + hash) con `modifier: idf` en la colección.
- **Migración.** No se puede añadir un vector disperso a una colección existente: crear `<colección>_v2`, reingestar todo y cambiar `QDRANT_COLLECTION`. Gracias a las cachés del paso 3, reingestar no repite OCR ni embeddings.
- **Archivos.** `internal/infrastructure/qdrant/qdrant.go`, `ports.VectorStore`, `ingest_usecase.go`. Mantener el filtro obligatorio por usuario/documentos (hoy `Search` rechaza búsquedas sin filtro: no romper esa guardia, evita fugas entre usuarios).

### 4.3 Reranker (prioridad media)

- **Problema.** El fragmento correcto está en el top 8 pero no primero (hit@1 46/69).
- **Propuesta.** Traer 20-30 candidatos y reordenarlos con un reranker antes de pasar 8 al LLM. Opciones: API de rerank dedicada (Cohere, Jina, Voyage: requiere otra clave y otro proveedor que vea los fragmentos) o un LLM barato de OpenRouter que puntúe. **Decisión pendiente del propietario**: proveedor y privacidad.
- **Beneficio extra.** El puntaje del reranker sí separa "no hay respuesta" mejor que el coseno.

### 4.4 Fuentes con página en la respuesta (prioridad media, barato)

- Guardar `page_number` y el índice del fragmento en el payload de Qdrant (hoy `IngestUsecase.Execute` recibe metadatos `nil`). Mostrar "Manual X, p. 4" en el chat y en el chat público.
- Ojo: tras pasar por la caché de extracción, los números de página vuelven como `float64` (JSON). Convertir antes de guardar.

### 4.5 Medir respuestas, no solo recuperación (prioridad media)

La evaluación actual no llama al LLM. Añadir un modo que genere la respuesta y la compare con `expected_contains` o con un juez LLM, y que compruebe que las 12 preguntas sin respuesta reciben la frase de rechazo.

### 4.6 Probar otros modelos de OCR (prioridad baja)

Solo se midió `google/gemini-2.5-flash-lite`. Candidato más barato: `qwen/qwen3.7-flash`. Medir con `OCR_MODEL=<modelo> go run ./cmd/eval --force-ocr --parse-cache-dir ""` (sin caché para que realmente transcriba).

## 5. Pendientes chicos (cualquier modelo puede hacerlos)

- [ ] **Frontend, duplicado.** El servidor responde 200 con `duplicate: true`; el frontend muestra "Documento subido correctamente". Mostrar "Ya tenías este archivo" (`frontend/src/routes/_authenticated.documents.tsx`).
- [ ] **Dos subidas idénticas simultáneas** del mismo usuario chocan con el índice único y devuelven 500. Mapear el error a la respuesta de duplicado.
- [ ] **Reprocesar no limpia el error anterior.** `DocumentRepository.UpdateDocument` usa `Updates(struct)` de GORM, que ignora valores cero: `error_message` y `processing_started_at` no se borran en `ReprocessDocument`. Usar actualización con mapa (como `RequeueDocument`).
- [ ] **PDF grandes.** Medir un PDF de 100+ páginas con OCR por LLM contra el límite de 10 minutos del worker; si no entra, subir el límite o la concurrencia.
- [ ] **Seguridad del prompt.** `BuildUserMessage` solo escapa `"`; un fragmento con `</contexto>` puede falsificar etiquetas.
- [ ] **`.gitignore`** tiene el patrón `server`, que ignora archivos nuevos en `cmd/server/`. Cambiarlo a `/server`.
- [ ] **Confirmar `EMBEDDING_MODEL`** en el `.env` de producción. Nadie pudo leer ese archivo.
- [ ] **Tests pendientes** (se pidió tests mínimos en esta tanda; un pase posterior debe completarlos):
  - Parser de PDF: ruta `ForceOCR`, OCR nulo, error de OCR, entrada que no es PDF, fallos de `pdftotext`/`pdftoppm`, Tesseract real.
  - OCR con LLM: `stripCodeFence`, error "no choices", selección en `ocr.NewEngine`/`ocr.PDFOptions`, flag `--force-ocr`.
  - Cachés: acierto/fallo de la caché de parseo en el worker, `ForceOCR` sin escritura, repositorios GORM de ambas cachés, `ComputeEmbeddings` con aciertos parciales, respuesta 200 `duplicate` del handler, `RequeueDocument`.
  - Sin test desde antes: actualización de la fila del prompt al arrancar, streaming contra el proveedor real, guardia de Qdrant.

## 6. Configuración nueva de esta tanda

| Variable | Defecto | Qué hace |
|---|---|---|
| `OCR_ENGINE` | `llm` | `llm` (OpenRouter, con Tesseract de respaldo) o `tesseract`. Sin clave de OpenRouter se usa Tesseract. |
| `OCR_MODEL` | `google/gemini-2.5-flash-lite` | Modelo multimodal para el OCR. |
| `PDF_OCR_MODE` | `always` | `always`: toda página pasa por el LLM; `auto`: solo las páginas sin capa de texto. |
| `CHUNKER` | `recursive` | `recursive` o `fixed`. |
| `CHUNK_SIZE` / `CHUNK_OVERLAP` | 1500 / 200 | En runas. |
| `CHUNK_CONTEXT_HEADER` | `true` | Antepone `[título — p. N]` a cada fragmento. |

Cachés (paso 3), compartidas entre usuarios por hash de contenido:
- Texto extraído: tabla `parsed_contents`, clave `(hash del archivo, firma del parser)`. Firma = `parser.Version|motor|modelo|modo`. **Subir `parser.Version` (`internal/infrastructure/parser/formats.go`) cada vez que cambie la extracción o la limpieza**, si no se seguirá usando el texto viejo.
- Embeddings: tabla `embedding_cache_entries`, clave `(hash del texto del fragmento, modelo, dimensión)`. Solo en la ingesta; las preguntas del chat no pasan por ella.
- Archivos: `blobs/<sha256>` en MinIO; se borran solo cuando ningún documento los referencia.
- Descartado: guardar los vectores una sola vez con lista de usuarios. Ahorra espacio pero un error en los filtros sería una fuga entre usuarios.

## 7. Cómo arrancar en local

```
./scripts/dev-local.sh                     # contenedores rago-local + servidor en http://localhost:4004
set -a; . scripts/local-env.sh; set +a
go run ./cmd/eval                          # mide con la configuración del servidor (~2-3 min, usa red)
go run ./cmd/eval --dump-chunks --source <archivo>   # ver los fragmentos, sin red ni API
```

- Contenedores propios (proyecto `rago-local`): postgres 5433, qdrant 6333/6334, minio 9000/9001. **No tocar** los demás contenedores de la máquina (`scory-production-*`, `nginx-reverse-proxy`, `n8n-app`, etc.). Siempre `-p rago-local -f docker-compose.local.yml`; nunca `docker system prune`, `docker volume rm` ni `compose down -v`.
- `scripts/local-env.sh` pisa las variables de datos; la clave de OpenRouter y el secreto JWT vienen de `.env`. **No leer ni copiar `.env`.**
- La evaluación guarda el texto extraído en `~/.cache/rago-eval/parse`: repetirla no vuelve a pagar OCR. Para forzar OCR de nuevo: `--parse-cache-dir ""`.
- Usuarios de prueba locales: `smoke@local.test`, `cas-a@local.test`, `cas-b@local.test` (contraseña de los dos últimos: `local-smoke-pass-1`).

## 8. Cómo trabajar

- **Reparto por modelo** (pedido del propietario): el modelo más caro (Fable, o Opus si no hay Fable) solo diseña y revisa; Sonnet implementa; Haiku busca y hace ediciones pequeñas. Las decisiones marcadas como "decisión pendiente del propietario" se preguntan, no se asumen.
- **Receta por tarea.** Un subagente por tarea con: el problema, los archivos que puede editar (lista exacta), los tests mínimos y los comandos de verificación (`go build ./...`, `go vet ./...`, `go test ./internal/... ./cmd/...`). Después, la sesión principal vuelve a correr build y tests y lee el diff. Tareas en paralelo solo si tocan archivos distintos.
- **Tests.** El propietario pidió tests mínimos durante la implementación; la cobertura completa se hace en un pase aparte (lista en la sección 5).
- **Medir.** Toda mejora de recuperación se valida con `go run ./cmd/eval` contra la sección 3.
- **Commits.** Conventional commits, sin atribución a IA. Solo commitear cuando el propietario lo pide. Nunca push a `main` sin pedirlo: despliega producción.
- **No usar `gentle-ai review`** en este repo (decisión del propietario, 2026-11-17; desactivado con `gentle-ai review mode disable --scope clone`). Memoria persistente: Engram.
- **Gasto de contexto.** Lo que más consume: correr la evaluación muchas veces seguidas (matrices de configuraciones), leer diffs grandes y tener muchos subagentes. Preferir una medición antes y otra después, y leer solo los archivos que se van a tocar.
- Limitación del entorno: la lectura y escritura de `.env*` está denegada.
