# Hoja de ruta RAG (rago)

Guía para continuar la mejora de la calidad de respuestas del sistema RAG cuando no hay acceso al modelo principal, y reparto de trabajo entre modelos.

Estado de referencia: 2026-11-16. Rama `main`, cambios sin commitear.

## 0. Cómo usar este documento

- Leer primero la sección 2 (estado) y la sección 3 (pasos inmediatos).
- Marcar cada casilla solo cuando el resultado esté observado: un test, un comando o una comprobación manual.
- Si una tarea cambia el comportamiento de un comando, opción o mensaje, actualizar también la ayuda y la documentación.
- Un punto marcado **decisión de arquitectura** no se implementa sin el visto bueno del arquitecto.

### Pendiente de confirmar

Datos tomados de la sesión de auditoría y no revisados de nuevo al escribir este documento:

- [ ] Números de línea citados en las secciones 3 y 4 (los archivos existen, pero las líneas no se verificaron una a una).
- [ ] Valores reales de `.env` (ver paso 3b).

## 1. Roles por modelo

Regla: el modelo más capaz decide y revisa; los modelos más económicos escriben código.

| Modelo | ID | Rol |
|---|---|---|
| Fable 5.2 | `claude-fable-5-2` | Solo arquitecto y revisor. Decide qué construir, escribe el brief de cada tarea, revisa el diff resultante y toma la decisión final. No escribe código. |
| Sonnet 5.5 | `claude-sonnet-5-5` | Implementación de cada tarea (código, tests y documentación juntos) y auditorías de un flujo de código. |
| Haiku 5.5 | `claude-haiku-5-5` | Localizar elementos en el código, ediciones mecánicas pequeñas (uno o dos archivos) y documentos como este. |

### Cuando se agote el límite de Fable 5.2

- [ ] Cambiar el modelo de la sesión con `/model` a Opus 5.5 (`claude-opus-5-5`) o Sonnet 5.5, para el rol de arquitecto y revisor.
- [ ] Usar este documento como plan. Las tareas aquí descritas están especificadas lo bastante para ejecutarse sin Fable.
- [ ] Dejar para cuando vuelva Fable los puntos marcados **decisión de arquitectura**.

### Cómo pedir una tarea a un subagente (receta probada)

- [ ] Una tarea por subagente.
- [ ] Indicar el problema, los archivos exactos que puede editar y los tests que debe escribir antes del código.
- [ ] Indicar los comandos a ejecutar: `go build ./...`, `go vet ./...`, `go test ./...`.
- [ ] Pedir que informe qué dejó sin probar.
- [ ] La sesión principal vuelve a ejecutar los tests y lee el diff antes de aceptar.
- [ ] Dos tareas se ejecutan en paralelo solo si tocan archivos distintos.

## 2. Estado actual (primer lote, sin commitear)

Todo está en el árbol de trabajo de `main`, sin commit. `go build ./...`, `go vet ./...` y `go test ./...` pasan (63 tests).

- **Tamaño de recuperación.** El chat y `/ask` piden `RETRIEVAL_TOP_K` fragmentos (por defecto 8). Antes, el chat usaba `CONTEXT_WINDOW_LIMIT` (por defecto 8192) como límite de Qdrant, lo que enviaba prácticamente todo el corpus al LLM. `CONTEXT_WINDOW_LIMIT` se conserva solo por compatibilidad.
- **Umbral de relevancia.** Los resultados por debajo de `RETRIEVAL_SCORE_THRESHOLD` (por defecto 0.30; 0 lo desactiva) se descartan. Lógica en `internal/application/retrieval.go`.
- **Corrección de seguridad.** El chat de grupo público ya no busca sin filtro. Un grupo sin documentos responde "Este grupo no tiene documentos disponibles.". El adaptador de Qdrant rechaza una búsqueda sin filtro (`internal/infrastructure/qdrant/qdrant.go`). Antes, un grupo público sin documentos buscaba en los documentos de todos los usuarios.
- **Rol de sistema real.** El puerto del LLM recibe el texto de sistema y el de usuario por separado. El adaptador (`internal/infrastructure/openrouter/openrouter.go`) envía un mensaje `system` y un mensaje `user`. Nuevas variables: `LLM_TEMPERATURE` (por defecto 0.2) y `LLM_MAX_TOKENS` (por defecto 0, no se envía).
- **Prompt único.** Chat, `/ask` y chat público usan el mismo prompt en `internal/application/prompt.go`. Permite respuestas parciales y envuelve el contexto en `<contexto>` y `<documento fuente="...">`.
- **Prompt almacenado.** Al arrancar, la fila `system_prompt` de la base de datos se reemplaza solo si es idéntica al prompt por defecto anterior. Un prompt personalizado no se toca.

Resultado de la verificación: SIN BLOQUEANTES (2026-11-16). Un verificador independiente revisó el diff completo: las seis reglas del lote se cumplen y `go build`, `go vet` y `go test ./... -count=1` pasan. Límite: todo se probó con mocks; la guardia de Qdrant y la consulta que actualiza el prompt en la base de datos se verificaron solo leyendo el código (ver paso 7).

## 3. Pasos inmediatos (checklist, en este orden)

- [x] **1. Verificación independiente.** Hecha, sin bloqueantes (ver sección 2).
- [ ] **2. Rama y commits.** Crear una rama de funcionalidad y commitear (nada está commiteado). Conventional Commits, sin líneas de atribución de IA. División sugerida:
  - retrieval: top-k y umbral;
  - corrección de la fuga del chat público;
  - rol de sistema y prompt unificado.
- [ ] **3. Variables de `.env`.** Añadir solo si se quieren valores distintos de los por defecto: `RETRIEVAL_TOP_K`, `RETRIEVAL_SCORE_THRESHOLD`, `LLM_TEMPERATURE`, `LLM_MAX_TOKENS`.
- [ ] **3b. Revisar `.env` existente.** Confirmar los valores reales de `EMBEDDING_MODEL` y `CONTEXT_WINDOW_LIMIT`. Durante la auditoría nadie leyó ese archivo; todo se juzgó con los valores por defecto del código (`text-embedding-3-small`, `google/gemini-2.5-flash`). Si el modelo de embeddings es de la familia e5, bge o nomic, faltan los prefijos de consulta y de pasaje: eso se convierte en una tarea nueva.
- [ ] **4. Reiniciar el servidor.** Revisar la línea de log que indica si el prompt de sistema fue actualizado o conservado. Si se conservó un prompt personalizado, actualizarlo a mano: la redacción antigua empuja al modelo a rechazar respuestas.
- [ ] **5. Calibrar el umbral 0.30.** Es un valor inicial no calibrado. Hacer algunas preguntas reales, revisar la línea de log `search completed` (resultados frente a `passed_threshold`) y bajar el umbral si se descartan fragmentos buenos.
- [ ] **6. Revisión receipt-driven (`gentle-ai review`).** Está activada globalmente y marca este cambio como revisión pendiente. No pudo evaluarse automáticamente por los archivos sin seguimiento. Decidir si se ejecuta o se omite.
- [ ] **7. Pruebas manuales.** Los tres puntos siguientes no están cubiertos por tests; comprobarlos a mano una vez:
  - lectura y actualización de la fila del prompt al arrancar;
  - streaming contra un proveedor real;
  - guardia del adaptador de Qdrant.

## 4. Trabajo pendiente (orden recomendado)

Cada lote es una lista de verificación. Para cada punto: qué y por qué, archivo de partida y modelo sugerido.

### Lote 2: conjunto de evaluación (hacer antes que nada)

Sin él no se pueden medir las mejoras.

- [ ] Crear 20-30 pares pregunta / respuesta esperada a partir de los documentos de ejemplo de `testdata/`.
- [ ] Añadir un script o un test en Go que ejecute los pares e informe cuántos recuperan el documento correcto.
- Archivo de partida: `testdata/`.
- Modelo: Sonnet para la implementación. Diseño de qué medir: arquitecto.

### Lote 3: ingesta (requiere reingestar todos los documentos; agrupar estos cambios)

- [ ] **Metadatos de fragmentos.** El worker descarta los metadatos del parser: los fragmentos no guardan contexto ni posición. Guardar página e índice de fragmento en el payload y anteponer el título del documento al texto que se embebe. Archivo: `internal/worker/ingest_worker.go` (alrededor de las líneas 180-201).
- [ ] **Fallos de embeddings silenciados.** `internal/application/ingest_usecase.go` (alrededor de las líneas 44-60) omite los fragmentos fallidos y el documento termina como completado, dejando huecos invisibles. Agrupar las llamadas, reintentar con backoff y marcar el documento como fallido si falla cualquier fragmento.
- [ ] **PDF con páginas escaneadas.** El OCR solo se ejecuta si ninguna página tiene texto; un PDF mixto pierde sus páginas escaneadas. Aplicar OCR página a página. Archivo: `internal/infrastructure/parser/pdf.go` (alrededor de las líneas 62-80). Además, el OCR renderiza a 150 DPI sin modo de segmentación de página; usar 300 DPI.
- [ ] **Limpieza del texto de PDF.** No se reúnen las palabras con guion de final de línea ni se eliminan cabeceras y pies repetidos. Su impacto depende de los PDF del propietario y no fue medido.
- [ ] **Chunker.** Archivo: `internal/infrastructure/chunker/semantic.go`. Tamaño 1000 medido en bytes (unos 250 tokens, pequeño); solapamiento en runas; corta en cada punto ("Art. 5", "Dr."). Elegir nuevo tamaño y estrategia: **decisión de arquitectura**.
- [ ] **Tests.** No hay tests para chunker, parsers ni ingesta. Cada tarea anterior debe incluir los suyos.

### Lote 4: calidad de recuperación (misma arquitectura)

- [ ] **Preguntas de seguimiento.** Hoy se embebe la pregunta cruda y el historial no se usa para recuperar; "¿y el segundo?" recupera ruido. Reescribir la pregunta como autónoma con una llamada barata al LLM antes de embeber. Archivo de partida: `internal/application/chat_usecase.go` (llamadas a `EmbedText` en `SendMessage` y `SendStream`). Modelo: Sonnet.
- [ ] **Búsqueda híbrida** (densa + palabras clave) para términos exactos como nombres, códigos y números; Qdrant soporta vectores dispersos. **Decisión de arquitectura.**
- [ ] **Reranker** sobre los 20-30 primeros resultados, reducido a 5-8. **Decisión de arquitectura.**
- [ ] **Chat público.** No tiene historial de conversación y no devuelve fuentes.
- [ ] **Registro.** En el chat autenticado, registrar fragmentos recuperados y puntuaciones (hoy solo se registra el número).

### Limpiezas menores (Haiku o Sonnet, cuando convenga)

- [ ] Escape incompleto al armar el contexto: `BuildUserMessage` en `internal/application/prompt.go` solo escapa `"` en el nombre del archivo. Un fragmento que contenga `</documento>` o `</contexto>`, o un nombre con `<`, `>` o `&`, puede falsificar etiquetas. Escapar esos caracteres en nombre y texto. Lo señaló el verificador como endurecimiento, no como regresión.
- [ ] `SendMessage` y `SendStream` en `internal/application/chat_usecase.go` duplican unas 80 líneas.
- [ ] El uso de tokens se estima como caracteres / 4, no es el uso real.
- [ ] El cliente HTTP del LLM no tiene timeout.
- [ ] No hay comprobación de que una colección de Qdrant existente coincida con `EMBEDDING_DIMENSION`.
- [ ] El README lista `JWT_SECRET`, `ACCESS_TOKEN_EXPIRATION` y `REFRESH_TOKEN_EXPIRATION`, pero el código lee `JWTSECRET`, `JWT_ACCESS_EXPIRATION` y `JWT_REFRESH_TOKEN_EXPIRATION`.

## 5. Decisión de arquitectura ya tomada

No cambiar la arquitectura RAG ahora. Los problemas encontrados son errores y configuración, independientes de la arquitectura. El diseño hexagonal, con puertos para embedder, almacén vectorial y LLM, soporta todos los lotes anteriores.

Reconsiderar solo después del lote 4 y solo si el conjunto de evaluación muestra un techo:

- **RAG agéntico:** para preguntas de varios pasos o entre documentos.
- **GraphRAG:** para preguntas globales sobre relaciones en el corpus. Es caro en la ingesta y rara vez compensa en una instalación autoalojada.

Esa decisión corresponde al arquitecto, con números.
