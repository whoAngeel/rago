# Guía de Implementación: Fase 1 (API REST y Arquitectura Hexagonal)

¡Bienvenido al inicio de la construcción de RAGO! Esta guía está diseñada para que tú mismo desarrolles el código.

## 🎯 Objetivo de la Fase
Establecer la estructura base del proyecto bajo los principios de la **Arquitectura Hexagonal**, implementar un servidor HTTP con **Gin**, y estructurar los puertos y casos de uso necesarios para responder preguntas.

---

## 🏗️ Requerimientos a Implementar

Ya tienes la infraestructura básica (Server, Logging, Healthcheck). Ahora debes agregar las siguientes piezas para completar la Fase 1:

### 1. Puertos del Sistema (`internal/core/ports`)
Debes definir las interfaces que tu aplicación necesita para funcionar, sin importar la tecnología subyacente.

#### VectorStore (`vector.store.go`)
Interfaz para interactuar con la base de datos vectorial (Qdrant).

| Método | Argumentos | Retorna | Descripción |
|--------|------------|--------|-------------|
| `CreateCollection` | `ctx context.Context`, `name string` (nombre de colección), `size int` (dimensión de vectores, ej: 1536) | `error` | Crea una nueva colección en Qdrant. El `name` debe ser único. El `size` debe coincidir con la dimensión del modelo de embeddings. |
| `UpsertDocuments` | `ctx context.Context`, `collection string` (nombre), `docs []schema.Document` (documentos con content/metadata), `vectors [][]float32` (embeddings alineados con docs) | `error` | Inserta o actualiza documentos con sus vectores asociados. Cada documento en `docs` debe tener un vector en `vectors` en el mismo índice. |
| `Search` | `ctx context.Context`, `collection string` (nombre), `queryVector []float32` (vector de búsqueda), `limit int` (número máximo de resultados) | `([]SearchResult, error)` | Busca los `limit` documentos más similares al `queryVector` usando cosine similarity. Retorna slice de SearchResult con Document y score. |
| `GetPointsCount` | `ctx context.Context`, `collection string` | `(int64, error)` | Retorna el número total de puntos/documentos en la colección. |
| `DeleteCollection` | `ctx context.Context`, `collection string` | `error` | Elimina una colección completa. Debe manejar el caso de colección no existente. |

**Tipo auxiliar a crear:**
```go
type SearchResult struct {
    Document schema.Document
    Score    float32
}
```

#### LLMProvider (`llm.provider.go`)
Interfaz para interactuar con modelos de lenguaje y embeddings.

| Método | Argumentos | Retorna | Descripción |
|--------|------------|--------|-------------|
| `GenerateAnswer` | `ctx context.Context`, `prompt string` (prompt completo con contexto + pregunta) | `(string, error)` | Genera una respuesta usando el LLM. El prompt ya debe contener el contexto recuperado + la pregunta del usuario. Retorna el texto de la respuesta generada. |
| `EmbedText` | `ctx context.Context`, `text string` | `([]float32, error)` | Convierte un texto en un vector de embedding. El vector retornado debe tener la dimensión configurada (ej: 1536 para text-embedding-3-small). |

**Notas:**
- Para `EmbedText`: el `[]float32` retornado debe tener exactamente la dimensión configurada en `EMBEDDING_DIMENSION`.
- Considera agregar un método adicional `GenerateAnswerWithOptions(ctx, prompt, options)` que acepte parámetros como temperature, max_tokens.

#### Logger (`logger.go`)
Interfaz para logging estructurado.

| Método | Argumentos | Retorna | Descripción |
|--------|------------|--------|-------------|
| `Debug` | `msg string`, `args ...any` | - | Loggear mensaje de debug. Args son key-value pairs para contexto adicional. |
| `Info` | `msg string`, `args ...any` | - | Loggear mensaje informativo. |
| `Warn` | `msg string`, `args ...any` | - | Loggear warning (no fatal pero requiere atención). |
| `Error` | `msg string`, `args ...any` | - | Loggear error (operación falló). |
| `Fatal` | `msg string`, `args ...any` | - | Loggear error fatal y terminar programa. |
| `With` | `args ...any` | `Logger` | Retorna una nueva instancia de Logger con el contexto agregado incluido en todos los logs subsecuentes. |

**Notas:**
- El método `With` debe retornar una nueva implementación (no mutar la actual) para evitar side effects.
- Uso típico: `logger.With("request_id", id).Info("request started")`.

---

### 2. Capa de Aplicación / Casos de Uso (`internal/application`)
Esta es la capa que contiene tu lógica de negocio principal.
*   Crea un archivo `ask_usecase.go`.
*   Define un struct `AskUseCase` que reciba por inyección de dependencias tus puertos (`VectorStore`, `LLMProvider`, `Logger`).
*   Implementa un método `Execute(ctx context.Context, question string) (string, error)` que orqueste el flujo.

#### Flujo Detallado de AskUsecase.Execute

| Paso | Acción | Descripción |
|------|--------|-------------|
| 1 | Recibe `question` | La pregunta del usuario |
| 2 | `EmbedText(question)` | Convierte la pregunta en vector usando LLMProvider |
| 3 | `VectorStore.Search(collection, queryVector, limit)` | Busca documentos similares (limit default: 3-5) |
| 4 | Formatea contexto | Concatenar todos los `SearchResult[i].Document.PageContent` |
| 5 | Arma prompt | `Contexto: {contexto}\n\nPregunta: {question}\n\nResponde basándote solo en el contexto.` |
| 6 | `GenerateAnswer(prompt)` | Envía el prompt armado al LLM |
| 7 | Retorna respuesta | Devuelve el texto generado al handler |

**Ejemplo de prompt armado:**
```
Contexto: "El sol es una estrella. La luna refleja su luz. Los planetas orbitan alrededor."

Pregunta: "¿Qué ilumina la luna?"

Responde basándote solo en el contexto.
```

**Notas:**
- La colección puede venir de configuración o del request
- Si `Search` retorna vacío, responder "No encontré información relevante"
- El `limit` default puede ser 3-5 documentos
- Manejar errores en cada paso con logs apropiados

### 3. Capa de Infraestructura REST (`internal/infrastructure/rest`)
Aquí expondremos el caso de uso hacia el mundo exterior a través de HTTP.
*   En tu router de Gin, agrega un nuevo endpoint `POST /api/v1/ask`.
*   El handler debe recibir un body JSON como: `{"question": "¿De qué trata este documento?"}`.
*   Debe llamar a `AskUseCase.Execute()`.
*   Debe responder con un JSON: `{"answer": "..."}`.

### 4. Composition Root (`cmd/server/main.go`)
Aquí debes unir todas las piezas.
*   Instancia la configuración.
*   Instancia el Logger.
*   Instancia las implementaciones reales de tus puertos (tu cliente real de Qdrant, tu cliente real de LLM).
*   Inyecta esas implementaciones en una nueva instancia de `AskUseCase`.
*   Pásale el `AskUseCase` a tus Handlers/Router.
*   Arranca el servidor (esto ya lo tienes).

---

## 🧪 Verificación
Una vez implementado, arranca tu servidor y prueba con:

```bash
curl -X POST http://localhost:4004/api/v1/ask \
-H "Content-Type: application/json" \
-d '{"question": "test"}'
```

Deberías recibir la respuesta armada por tu caso de uso.

---

## 🏁 Siguiente Paso
**Regla Estricta:** Una vez que hayas programado, testeado y estés satisfecho con esta fase, debes decirme exactamente la siguiente frase en el chat:

> **"He completado la Fase 1, por favor genera la documentación para la Fase 2"**


Al decirme eso, evaluaré si hay algo que mejorar, y te prepararé la Guía de Implementación para integrar PostgreSQL y JWT. ¡Éxito con el código!
