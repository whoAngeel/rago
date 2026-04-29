# Roadmap y Arquitectura de RAGO (Evolución a API REST)

Este documento forma parte de la documentación oficial del proyecto. Detalla la evolución planificada de RAGO, pasando de ser una herramienta CLI local a una plataforma SaaS completa con API REST, base de datos relacional y un Frontend interactivo.

## Arquitectura del Sistema

### Arquitectura Hexagonal (Independencia de Infraestructura)
Para asegurar que RAGO sea 100% agnóstico a las herramientas de terceros, el backend en Go adoptará la **Arquitectura Hexagonal (Puertos y Adaptadores)**.

- **Dominio (`/domain`):** Entidades de negocio como `Usuario`, `Documento` o `Sesión de Chat`.
- **Puertos (`/ports`):** Interfaces que definen contratos, por ejemplo: `type VectorStore interface { Search(...) }`.
- **Casos de Uso (`/application`):** Lógica que orquesta el flujo (ej. `PreguntarRAGUseCase`).
- **Adaptadores (`/infrastructure`):** Implementaciones tecnológicas concretas.
  - *¿Quieres cambiar Qdrant por Pinecone?* Simplemente creas una carpeta `/infrastructure/pinecone`, implementas la interfaz `VectorStore` y cambias la inyección de dependencias en el `main.go`. El núcleo de RAGO no sufrirá ningún cambio.

---

## FASE 1: BACKEND (Go + Gin + PostgreSQL)

Esta gran fase se enfoca en construir una API REST robusta, sin interfaz gráfica, que maneje toda la lógica pesada, bases de datos y procesamiento en segundo plano.

### Release 1.1: Refactorización a API REST Base y Logging
**Objetivo:** Abandonar la CLI, exponer RAGO a través de la web y establecer cimientos de observabilidad.
- Integración de **Gin** como framework web para enrutamiento HTTP de alto rendimiento.
- Refactorización del código actual (`cmd/rag/main.go`) hacia la estructura hexagonal `/internal`.
- **Servicio de Logging Estructurado:** Integrar un logger central (ej. `slog` estándar de Go 1.21+ o `Zap`) para registrar peticiones HTTP, errores y trazabilidad del sistema desde el día cero.
- Creación del endpoint `POST /api/v1/ask` que recibirá un JSON con la pregunta y devolverá la respuesta generada por el LLM.
- Creación de un endpoint básico de salud `GET /health`.

### Release 1.2: PostgreSQL, Identidad y Roles
**Objetivo:** Proveer una base de datos relacional y soporte multi-usuario.
- Integración de **PostgreSQL** para almacenar entidades relacionales (evitando guardar estados en memoria o en el Vector DB).
- Creación de tablas de `Usuarios` y `Roles`.
- Implementación de Autenticación basada en **JWT (JSON Web Tokens)**.
- Creación de middlewares en Gin para **Control de Acceso (RBAC)**: Validar que solo los administradores puedan subir archivos y que los visores solo puedan consultar.

### Release 1.3: Gestión de Documentos y Almacenamiento Blob
**Objetivo:** Permitir la subida de archivos físicos sin bloquear la aplicación por el procesamiento del LLM.
- Diseño de la tabla `Documentos` en PostgreSQL para llevar trazabilidad (nombre, fecha de subida, estado de procesamiento).
- Implementación de la interfaz `BlobStorage`.
- Creación del adaptador `LocalFS` para guardar temporal o permanentemente los archivos físicos subidos por los usuarios en el disco del servidor.
- Endpoint `POST /api/v1/documents` que reciba un `multipart/form-data`, guarde el archivo físico en el BlobStorage, e inserte un registro en PostgreSQL con el estado `PENDING`.

### Release 1.4: Ingesta Automática y Asíncrona (Goroutines)
**Objetivo:** Aislar el procesamiento pesado (Chunking y Embeddings) del ciclo de vida HTTP.
- Desarrollo de un **Worker Pool** utilizando Goroutines nativas y Canales de Go.
- El worker observará continuamente la base de datos en busca de documentos `PENDING`.
- Al encontrar uno, lo descargará del BlobStorage, extraerá el texto, lo dividirá en Chunks, calculará los embeddings y lo insertará en Qdrant.
- Tras el éxito, actualizará el estado en PostgreSQL a `COMPLETED`. En caso de error, a `FAILED`.
- **Aislamiento Vectorial:** Se inyectará el `user_id` en los metadatos de los puntos de Qdrant para garantizar que las búsquedas futuras estén estrictamente acotadas a los documentos que le pertenecen al usuario que consulta.

### Release 1.5: Ecosistema de Parsers y Tipos de Archivos
**Objetivo:** Permitir que el sistema ingiera el 90% de los formatos ofimáticos estándar.
- Refactorizar la lógica de lectura actual para soportar una arquitectura de *Extractores Modulares*.
- Soporte para formatos nativos estructurados: `JSON`, `CSV`.
- Integración de librerías para ofimática y maquetación: `PDF`, `DOCX`, `XLSX`.
- Optimización de estrategias de *Chunking* según el tipo de archivo (ej. no cortar un CSV a mitad de una fila).

### Release 1.6: Chat Contextual RAG y Memoria
**Objetivo:** Permitir conversaciones continuas (Hilos de chat) y no solo preguntas aisladas de un solo turno (Zero-Shot).
- Creación de tablas `Chat_Sessions` y `Chat_Messages` en PostgreSQL.
- Endpoints CRUD para manejar sesiones de chat.
- Modificación del motor RAG: Antes de enviar el Prompt al LLM, la aplicación concatenará el historial reciente de mensajes desde Postgres con el contexto extraído de Qdrant, logrando que el modelo recuerde lo dicho en turnos anteriores.

### Release 1.7: Websockets e Interactividad en Tiempo Real
**Objetivo:** Preparar la API para un consumo reactivo por parte de un Frontend moderno.
- Integración de un servidor de **WebSockets** dentro de Gin.
- Canal de Eventos: Emitir un evento cuando un documento cambie de `PROCESSING` a `COMPLETED` para que el frontend actualice su interfaz sin recargar.
- (Opcional) Canal de Streaming: Enviar los tokens de respuesta del LLM a medida que se generan (Server-Sent Events o WebSockets) para reducir la percepción de latencia en la UI.

---

## FASE 2: FRONTEND (Web App)

Con la API REST madura, segura y asíncrona, se construirá una aplicación de cliente enriquecida (Single Page Application o Server-Side Rendered App).

### Estructura Tecnológica Recomendada
- **Framework:** React / Next.js (o Vue / Nuxt).
- **Estilos:** TailwindCSS o una librería de componentes basada en Vanilla CSS si se requiere máximo control custom.
- **Gestión de Estado:** Zustand (React) o Pinia (Vue) para estados globales, además de integración con Websockets.

### Release 2.1: Autenticación y Diseño Base
- Maquetación de la UI con un diseño premium y responsive.
- Pantallas de Login y Registro.
- Manejo seguro del JWT almacenándolo de manera adecuada (preferiblemente HttpOnly Cookies).

### Release 2.2: Dashboard Gestor de Documentos
- Interfaz para ver todos los documentos subidos con su estado (Pendiente, Procesando, Listo, Error).
- Zona de *Drag & Drop* para subir archivos que impactará directamente al endpoint `/documents` de la API.
- Indicadores visuales de espacio o cantidad de documentos por usuario.

### Release 2.3: Interfaz Conversacional RAG (El Chat)
- Desarrollo de la vista principal del usuario: Un chat estilo ChatGPT.
- Barra lateral con el historial de "Sesiones de Chat".
- Área principal con burbujas de conversación diferenciadas por `Rol` (Usuario vs Asistente).
- Integración de WebSockets/SSE para imprimir el texto del LLM de manera fluida (efecto de "escribiendo...").
- *Feature:* Mostrar debajo de la respuesta del LLM un acordeón o lista pequeña de las "Citas" o "Fuentes" de donde se extrajo la información (mapeando a los Chunks del documento original).
