# RAGO — Fases de Desarrollo Frontend

> **Contexto:** Este documento define las fases de construcción del frontend de RAGO, asumiendo que el backend en Go (Fase 1 del roadmap) ya está funcional y expone los endpoints documentados en `docs/roadmap.md` y `docs/architecture-decisions.md`.
>
> **Stack Tecnológico Confirmado (tras sesión de grill):**
> - **Runtime/Bundler:** Bun + Vite
> - **Framework:** React + TypeScript
> - **Styling:** Tailwind CSS (configuración personalizada con Design System Brainfish)
> - **Server State:** TanStack Query v5 (endpoints sin SSE)
> - **Client State:** Zustand (estado del chat)
> - **Forms:** React Hook Form + Zod (validación type-safe)
> - **Routing:** TanStack Router
> - **Components:** Custom components (botones, cards, inputs neo-brutalist)
> - **Icons:** Lucide React
>
> **Roles:**
> - **Usuario (@whoangel):** Setup inicial del proyecto, configuración de Bun/Vite/Tailwind, instalación de dependencias.
> - **Asistente (opencode/big-pickle):** Apoyo en levantamiento del frontend — creación de componentes UI, implementación de lógica de negocio, integración con backend Go, y seguimiento del Design System Brainfish.

---

## FASE 2: FRONTEND (App Web React)

### Release 2.0: Configuración Base y Design System
**Objetivo:** Preparar el entorno de desarrollo y definir los tokens visuales de Brainfish.

**Responsabilidades:**
- **[@whoangel]** Setup del proyecto:
  - Ejecutar `bun create vite` (seleccionar React + TypeScript)
  - Instalar dependencias: `bun add @tanstack/react-query @tanstack/router zustand react-hook-form zod lucide-react`
  - Instalar Tailwind CSS v4 y configurar `tailwind.config.ts`
- **[Asistente]** Configuración de Design System:
  - Crear `tailwind.config.ts` con paleta Brainfish (ver `docs/DESIGN-SYSTEM.md`):
    - Colores primarios: `#a3e635` (verde lima), `#84cc17`, `#4e7c10`
    - Neutros: `#171717` (dark-900), `#f5f5f5` (smokewhite)
    - Acentos: `#f5d1fe` (rosa), `#edeafe` (púrpura), `#fb923d` (naranja)
    - Sombras hard-offset: `box-shadow: 2px 2px 0 0 #0a0a0d`
    - Border-radius: `4px` (botones), `100px` (pills), `20px` (cards)
  - Configurar fuente **Satoshi** (vía Google Fonts o archivos locales)
  - Crear archivo de tipos TypeScript base (`types/`)

---

### Release 2.1: Autenticación y Diseño Base
**Objetivo:** Maquetación de UI premium y sistema de login seguro.

**Endpoints Go a consumir:**
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`

**Responsabilidades:**
- **[Asistente]** Lógica de autenticación:
  - Crear store de Zustand para manejo de JWT (`access_token` en memoria, `refresh_token` en HttpOnly cookie)
  - Integrar TanStack Query para mutaciones de login/register
  - Implementar hook `useAuth()` personalizado
- **[Asistente]** Componentes UI Brainfish:
  - `Button` (primario verde lima, secundario blanco, ambos con borde negro 1px y sombra 2px)
  - `Input` (borde `#737373` idle, `#171717` focus, radius 4px)
  - `Card` (bg blanco, borde 1px `#171717`, sombra hard-offset)
  - `Tag` / `Pill` (radius 100px, estilo navbar flotante)
- **[Asistente]** Páginas:
  - `Login.tsx` — Formulario con React Hook Form + Zod (validación email/password)
  - `Register.tsx` — Registro con roles (default: viewer)
  - `Layout.tsx` — Navbar flotante estilo Brainfish (logo, links, botones Sign in / Book Demo)

---

### Release 2.2: Dashboard Gestor de Documentos
**Objetivo:** Interfaz para gestión de documentos con estados visuales.

**Endpoints Go a consumir:**
- `POST /api/v1/documents` (multipart/form-data)
- `GET /api/v1/documents` (lista con estados)
- `DELETE /api/v1/documents/:id`
- `GET /api/v1/stream` (SSE para notificaciones de estado)

**Responsabilidades:**
- **[Asistente]** Integración TanStack Query:
  - `useDocuments()` — fetch lista de documentos
  - `useUploadDocument()` — mutación con progress tracking
  - `useDeleteDocument()` — mutación de borrado
- **[Asistente]** Componentes específicos:
  - `DocumentCard` — Muestra estado (PENDING, PROCESSING, COMPLETED, FAILED) con badges estilo Brainfish
  - `DropZone` — Área de drag & drop para subir archivos (PDF, DOCX, XLSX, CSV, JSON)
  - `ProgressBar` — Barra de progreso usando tabla `processing_steps` de Go
- **[Asistente]** Página `Dashboard.tsx`:
  - Grid de documentos del usuario
  - Indicadores visuales de espacio/cantidad
  - Manejo de estados huérfanos (documento eliminado)

---

### Release 2.3: Interfaz Conversacional RAG (El Chat)
**Objetivo:** Chat estilo ChatGPT con historial de sesiones y streaming de respuesta.

**Endpoints Go a consumir:**
- `POST /api/v1/chat/sessions` (crear sesión implícita)
- `GET /api/v1/chat/sessions` (listar sesiones)
- `POST /api/v1/chat/send-stream` (streaming de tokens)
- `GET /api/v1/chat/sessions/:id/messages` (historial)

**Responsabilidades:**
- **[Asistente]** Estado del Chat (Zustand):
  - Store: `currentSession`, `messages[]`, `isStreaming`, `sources[]`
  - Manejo de flujo de streaming (acumulación de tokens en memoria)
- **[Asistente]** Componentes de Chat:
  - `ChatWindow` — Área principal de mensajes
  - `MessageBubble` — Burbujas diferenciadas por Rol (Usuario vs Asistente) con estilo neo-brutalist
  - `ChatSidebar` — Historial de "Sesiones de Chat" con scroll
  - `SourcesAccordion` — Acordeón que muestra "Citas" o "Fuentes" (mapeo a chunks del documento original)
  - `TypingIndicator` — Efecto "escribiendo..." durante streaming
- **[Asistente]** Integración SSE:
  - Consumo de `POST /chat/send-stream` para imprimir texto de LLM de manera fluida
  - Manejo de eventos: `chat_token`, `chat_done`, `error`
- **[Asistente]** Página `Chat.tsx`:
  - Layout 2 columnas (sidebar + chat)
  - Auto-generación de título de sesión basado en primer mensaje

---

### Release 2.4: Refinamiento Neo-Brutalist y UX
**Objetivo:** Alinear completamente con el Design System Brainfish y pulir interactividad.

**Responsabilidades:**
- **[Asistente]** Patrones interactivos Brainfish:
  - Hover en botones/cards: `transform: translate(2px, 2px)` + sombra `0 0 0 0` (efecto "hundirse")
  - Focus states con outline negro accesible
  - Disabled states con colores `--dark-300` / `--dark-400`
- **[Asistente]** Elementos de marca:
  - Peces verde lima dispersos como decoración flotante
  - Burbujas (círculos negros) alrededor de secciones
  - Estrellas de 4 puntas antes de secciones
  - Ondas SVG como separadores entre secciones
- **[Asistente]** Responsive y accesibilidad:
  - Adaptación a viewports de 1440px (diseño original)
  - Navegación por teclado en chat
  - Anuncios de estado para lectores de pantalla

---

## Estructura de Carpetas (Frontend)

```
frontend/
├── public/
│   ├── fonts/                # Satoshi (woff2)
│   └── images/               # Logos, peces, decoraciones
├── src/
│   ├── components/
│   │   ├── ui/              # Button, Card, Input, Badge, Tag, Pill
│   │   ├── chat/            # ChatWindow, MessageBubble, SourcesAccordion
│   │   ├── documents/       # DocumentCard, DropZone, ProgressBar
│   │   ├── layout/          # Navbar, Sidebar, Footer
│   │   └── icons/           # Star, Fish, Bubble (Brainfish brand)
│   ├── routes/              # TanStack Router pages (Login, Dashboard, Chat)
│   ├── hooks/               # useAuth, useChat, useDocuments
│   ├── store/               # Zustand stores (auth, chat)
│   ├── lib/                 # Tailwind config, API client, constants
│   ├── types/               # TypeScript interfaces (Document, Message, Session)
│   ├── utils/               # Formatters, helpers
│   └── main.tsx             # Entry point
├── tailwind.config.ts        # Brainfish Design System tokens
├── vite.config.ts
└── package.json
```

---

## Notas de Implementación

1. **SSE vs Polling:** Por ahora usamos `POST /chat/send-stream` para streaming de chat. El documento `docs/architecture-decisions.md` menciona que SSE es el protocolo principal (11.1). No usar WebSockets para chat (overkill según 11.1).

2. **Manejo de Fuentes Huérfanas:** Si un documento se borra, el frontend debe mostrar "Documento eliminado" en las fuentes de mensajes (ver 1.2 en architecture-decisions).

3. **Tokens JWT:** Access token en memoria (no LocalStorage por seguridad), refresh token en HttpOnly cookie (ver 2.1 en architecture-decisions).

4. **Debilidades Conocidas:** El backend tiene problemas de streaming (tokens con `\n` rompen SSE, ver 10.8). El frontend debe manejar estos casos gracefully.

5. **Simultaneidad:** Mientras el asistente trabaja en componentes y lógica, el usuario puede estar configurando nuevas herramientas o ajustando la configuración de Tailwind/Vite.

---

## Próximos Pasos

1. **[@whoangel]** Ejecutar: `bun create vite frontend --template react-ts`
2. **[@whoangel]** Instalar dependencias y configurar Tailwind
3. **[Asistente]** Crear archivos base: `tailwind.config.ts`, `types/`, `lib/api.ts`
4. **[Asistente]** Implementar Release 2.1 (Autenticación + Componentes UI base)

¿Procedo a apoyarte con el paso 3 (archivos base) una vez que tengas el setup listo?
