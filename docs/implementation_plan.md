# Plan de Implementación - RAGo

Este plan detalla los pasos técnicos necesarios para llevar el proyecto desde su estado actual hasta un MVP funcional.

## Fase 1: Capa de Almacenamiento (Store)
**Objetivo:** Implementar la integración completa con Qdrant.
- [ ] Definir interfaz `VectorStore` en `internal/store`.
- [ ] Implementar el cliente de Qdrant en `internal/store/qdrant.go`.
- [ ] Crear métodos para:
  - `CreateCollection(name string, vectorSize int)`
  - `UpsertDocuments(docs []schema.Document, vectors [][]float32)`
  - `Search(queryVector []float32, limit int) ([]schema.Document, error)`

## Fase 2: Orquestador (Engine)
**Objetivo:** Crear el motor que coordina el flujo de datos.
- [ ] Crear `internal/engine/rag_engine.go`.
- [ ] Implementar la lógica de "RAG":
  1. Recibir pregunta.
  2. Generar embedding de la pregunta.
  3. Buscar en Qdrant.
  4. Formatear el contexto recuperado en un Prompt.
  5. Llamar al LLM y retornar la respuesta.

## Fase 3: Interfaz de Usuario (CLI)
**Objetivo:** Exponer la funcionalidad al usuario final.
- [ ] Implementar `cmd/rag/main.go` con comandos básicos:
  - `ingest <path>`: Procesa y sube documentos a Qdrant.
  - `ask "<pregunta>"`: Ejecuta el flujo RAG completo y muestra la respuesta.

## Fase 4: Mejoras y Pulido
- [ ] Soporte para múltiples colecciones.
- [ ] Manejo de historial de conversación (Context Memory).
- [ ] Soporte para archivos PDF.
- [ ] Logs detallados y manejo de errores robusto.

---
**Estado Actual:**
- Ingest (Splitter): ✅
- Provider (Embeddings): ✅
- Store: 🛠️ (Conexión básica probada)
- Engine: ❌
- CLI: ❌
