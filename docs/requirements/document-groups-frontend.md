# Frontend: Document Groups & Shared Chat
> Implementar después de que el backend de cada iteración esté deployado y testeado.
> Depende de: `document-groups-backend.md`

---

## Iteración 1-F — Feedback de límites en documentos

Corresponde al backend Iteración 1 (checksum + max_documents).

### Cambios en upload

- Si API devuelve `409 Conflict` con `error: "duplicate_document"` → toast de error: *"Ya tenés este archivo subido"*
- Si API devuelve `429 Too Many Requests` con `error: "quota_exceeded"` → toast de error: *"Alcanzaste el límite de X documentos"*

**Archivos a modificar:**
- `routes/_authenticated.documents.tsx` — manejo de errores en `uploadMutation.onError`

---

## Iteración 2-F — Creación y gestión de grupos

Corresponde al backend Iteración 2.

### 2-F.1 Selección de documentos para crear grupo

En la tabla de documentos (`DocumentTable`):
- Agregar columna de checkbox (solo habilitada para docs con `status === 'completed'`)
- Toolbar flotante que aparece cuando hay al menos 1 seleccionado: *"Crear grupo con X documentos"*
- Click → abre modal `CreateGroupModal`

**Archivos a crear/modificar:**
- `components/documents/DocumentTable.tsx` — checkboxes + toolbar de selección
- `components/groups/CreateGroupModal.tsx` — nuevo

### 2-F.2 Modal de creación de grupo

Campos:
- Nombre del grupo (input texto, obligatorio)
- Toggle "Permitir descargar archivos" (default: ON)
- Botón "Crear grupo"

On submit → `POST /api/v1/groups/` → redirigir a `/groups` (lista de grupos).

### 2-F.3 Sección `/groups`

Nueva ruta en el router: `_authenticated.groups.tsx`

**Vista lista de grupos:**
- Tabla o cards con: nombre, cantidad de docs, estado (activo/inactivo), `chat_attempts` intentos, `chat_quota_used / chat_quota` mensajes usados
- Botones: copiar link, ver QR, editar, activar/desactivar, eliminar

**Vista detalle `/groups/:id`:**
- Nombre del grupo (editable inline o con botón editar)
- Lista de documentos incluidos con botón para remover cada uno
- Botón "Agregar documentos" → abre selector de docs completados
- Stats: intentos de chat, cuota usada
- Link compartible con botón de copy
- QR generado en frontend (librería: `qrcode.react`)
- Toggle allow_downloads
- Botón desactivar/reactivar link

**Archivos a crear:**
- `routes/_authenticated.groups.tsx`
- `routes/_authenticated.groups.$id.tsx`
- `components/groups/GroupCard.tsx`
- `components/groups/GroupQR.tsx`
- `hooks/useGroups.ts` — queries y mutations de grupos

### 2-F.4 Tipos nuevos

```typescript
// types/index.ts
interface DocumentGroup {
    id: number
    user_id: number
    name: string
    slug: string
    is_active: boolean
    allow_downloads: boolean
    chat_quota: number
    chat_quota_used: number
    chat_attempts: number
    document_ids: number[]
    created_at: string
    updated_at: string
}
```

### 2-F.5 Navegación

Agregar "Grupos" al sidebar de la app autenticada.

---

## Iteración 3-F — Página pública de chat

Corresponde al backend Iteración 3. Esta ruta es pública, sin layout autenticado.

### 3-F.1 Ruta pública `/c/:slug`

**No usa el layout autenticado.** Ruta separada, sin sidebar, sin auth guard.

**Layout de la página:**
```
┌─────────────────────────────────────────────┐
│  [Nombre del grupo]               logo app  │
├──────────────┬──────────────────────────────┤
│              │                              │
│  Documentos  │       Chat                   │
│  (panel)     │                              │
│              │  [mensaje del visitante]     │
│  📄 doc1.pdf │  [respuesta del LLM]         │
│  📄 doc2.pdf │                              │
│  ↓ descargar │  [input + botón enviar]      │
│  (si enabled)│                              │
└──────────────┴──────────────────────────────┘
```

**Comportamiento:**
- Al cargar: `GET /api/v1/public/groups/:slug` → obtener nombre, lista de docs, `allow_downloads`
- Si grupo no existe o inactivo (404) → página de error: *"Este chat no está disponible"*
- Chat: stateless, sin historial al recargar
- On send: `POST /api/v1/public/groups/:slug/chat` con `{ message: "..." }`
- Si `429 quota_exceeded`: deshabilitar input, mostrar banner *"Este chat ha alcanzado su límite de mensajes"*
- Descargas: botón por doc → `GET /api/v1/public/groups/:slug/documents/:id/download` (solo si `allow_downloads`)

**Archivos a crear:**
- `routes/c.$slug.tsx` — ruta pública
- `components/public/PublicChat.tsx`
- `components/public/DocumentPanel.tsx`

### 3-F.2 QR en la gestión (generación en frontend)

```tsx
// components/groups/GroupQR.tsx
import QRCode from 'qrcode.react'

const shareUrl = `${window.location.origin}/c/${group.slug}`
<QRCode value={shareUrl} size={200} />
```

Instalar: `npm install qrcode.react`

---

## Resumen de archivos por iteración frontend

### Iteración 1-F
- `routes/_authenticated.documents.tsx` — manejo 409/429 en upload

### Iteración 2-F
- `types/index.ts` — `DocumentGroup`
- `components/documents/DocumentTable.tsx` — checkboxes + toolbar
- `components/groups/CreateGroupModal.tsx` — nuevo
- `components/groups/GroupCard.tsx` — nuevo
- `components/groups/GroupQR.tsx` — nuevo
- `hooks/useGroups.ts` — nuevo
- `routes/_authenticated.groups.tsx` — nuevo
- `routes/_authenticated.groups.$id.tsx` — nuevo
- Sidebar — agregar "Grupos"

### Iteración 3-F
- `routes/c.$slug.tsx` — nuevo (ruta pública)
- `components/public/PublicChat.tsx` — nuevo
- `components/public/DocumentPanel.tsx` — nuevo
- `package.json` — agregar `qrcode.react`

---

## Fuera de scope frontend (para después)
- CTA de registro cuando se agota la cuota
- Analytics del link (quién abrió, cuándo)
- Chat con historial por sesión
- Dark mode en página pública
