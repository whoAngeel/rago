# PRD - RAGo (RAG in Go)

## 1. Visión General
RAGo es un sistema de Generación Aumentada por Recuperación (RAG) ligero y eficiente desarrollado en Go. Su objetivo es permitir a los usuarios cargar documentos locales, procesarlos en fragmentos vectorizados y realizar consultas en lenguaje natural utilizando el contexto extraído de dichos documentos.

## 2. Objetivos
- Proporcionar una herramienta CLI para la ingesta y consulta de documentos.
- Utilizar una arquitectura modular que permita intercambiar proveedores de embeddings y LLMs.
- Garantizar una recuperación rápida de información utilizando Qdrant como motor vectorial.

## 3. Stack Tecnológico
- **Lenguaje:** Go 1.22+
- **Orquestación:** [langchaingo](https://github.com/tmc/langchaingo)
- **Base de Datos Vectorial:** Qdrant (vía gRPC/HTTP)
- **Embeddings:** OpenAI / OpenRouter
- **LLM:** OpenRouter (modelos configurables)
- **Configuración:** Variables de entorno (.env)

## 4. Funcionalidades Core

### 4.1. Ingesta de Documentos
- Lectura de archivos de texto (inicialmente .txt, .md).
- Segmentación inteligente (Splitting) para preservar el contexto semántico.
- Generación de embeddings para cada segmento.
- Almacenamiento en Qdrant con metadatos (fuente, posición).

### 4.2. Recuperación (Retrieval)
- Transformación de la consulta del usuario en un vector de búsqueda.
- Búsqueda de similitud de coseno en Qdrant para obtener los fragmentos más relevantes (Top-K).

### 4.3. Generación (Augmentation)
- Construcción de un prompt enriquecido que incluye:
  - Instrucciones del sistema.
  - Contexto recuperado.
  - Pregunta original del usuario.
- Inferencia mediante LLM para generar una respuesta coherente.

## 5. Requerimientos No Funcionales
- **Rendimiento:** Latencia de búsqueda < 200ms.
- **Mantenibilidad:** Código estructurado siguiendo los principios de Clean Architecture.
- **Extensibilidad:** Facilidad para añadir nuevos conectores (Pinecone, Weaviate, etc.).

## 6. User Stories
1. **Como usuario**, quiero poder ejecutar un comando para "aprender" de una carpeta de archivos.
2. **Como usuario**, quiero poder hacer preguntas sobre mis documentos y recibir respuestas precisas con fuentes citadas.
3. **Como desarrollador**, quiero poder cambiar el modelo de lenguaje simplemente editando mi archivo `.env`.
