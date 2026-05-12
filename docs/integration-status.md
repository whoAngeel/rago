# Estado de integración Backend ↔ Frontend

## Backend — implementado

### Endpoints activos

| Método | Ruta | Handler | Estado |
|---|---|---|---|
| `GET` | `/api/v1/users/me` | `UserHandler.Me` | ✅ listo |
| `GET` | `/api/v1/documents/` | `DocumentHandler.List` | ✅ paginado |
| `POST` | `/api/v1/documents/` | `DocumentHandler.Upload` | ✅ con checksum + límite |
| `DELETE` | `/api/v1/documents/:id` | `DocumentHandler.Delete` | ✅ |
| `GET` | `/api/v1/documents/:id/steps` | `DocumentHandler.Steps` | ✅ |
| `GET` | `/api/v1/groups/` | `DocumentGroupHandler.List` | ✅ paginado |
| `POST` | `/api/v1/groups/` | `DocumentGroupHandler.Create` | ✅ |
| `GET` | `/api/v1/groups/:id` | `DocumentGroupHandler.Get` | ✅ |
| `PATCH` | `/api/v1/groups/:id` | `DocumentGroupHandler.Update` | ✅ |
| `DELETE` | `/api/v1/groups/:id` | `DocumentGroupHandler.Delete` | ✅ |
| `POST` | `/api/v1/groups/:id/documents` | `DocumentGroupHandler.AddDocuments` | ✅ |
| `DELETE` | `/api/v1/groups/:id/documents/:doc_id` | `DocumentGroupHandler.RemoveDocument` | ✅ |
| `GET` | `/api/v1/public/groups/:slug` | `PublicGroupHandler.GetGroupInfo` | ✅ |
| `POST` | `/api/v1/public/groups/:slug/chat` | `PublicGroupHandler.Chat` | ✅ con cuotas |
| `GET` | `/api/v1/public/groups/:slug/documents/:doc_id/download` | `PublicGroupHandler.DownloadDocument` | ✅ |
| `GET` | `/api/v1/stream` | `SSEHandler.Stream` | ✅ emite `document_status` + `document_step` |

### Features implementados

- **Checksum deduplication**: SHA-256 por usuario → 409 `duplicate_document`
- **Límite de documentos**: `User.MaxDocuments` (0 = ilimitado) → 429 `quota_exceeded`
- **Cuota de chat**: `User.ChatQuota / ChatQuotaUsed` (autoritative) + `Group.ChatQuota` (cap opcional) → 429 `quota_exceeded`
- **Token tracking**: tabla `llm_usage` con modelo + tokens (aprox) por cada llamada a LLM, vinculada a `user_id` y opcionalmente a `group_id`
- **SSE pasos**: el worker emite `document_step {doc_id, step_id, step_name, status, error}` en cada inicio y fin de step, no solo en cambios de estado del documento
- **RAG threshold**: score mínimo 0.40 en public chat, fallback top-2 si todo queda filtrado

### Pending post-deploy (SQL manual)

```sql
-- Admins sin límites
UPDATE users SET max_documents = 0, chat_quota = 0 WHERE role_id = 1;
```

---

## Frontend — implementado

- **Documentos**: tabla con SSE en tiempo real (dots de progreso por step), badge de retry con "Reintentando / Recuperado / Falló en N intentos", step activo visible durante processing (embed... con spinner), error truncado con title en failed
- **SSE hook**: maneja `document_status` (invalida lista) y `document_step` (escribe cache `["steps", doc.id]`). Limpia cache de steps en `document_status: processing` para evitar datos rancios de reintentos anteriores
- **Groups page**: UI de tarjetas con datos del grupo (documentos, mensajes, intentos), toggle activo/inactivo visual, botones de QR y Copy link

---

## Frontend — pendiente de integración

### 1. Errores de upload (documentos)

El backend retorna 409 y 429 con body `{error, message}`, pero `uploadMutation.onError` en `_authenticated.documents.tsx` solo muestra `err.message` (el mensaje de Axios, no el body del backend).

**Fix**: en `onError` leer `err.response?.data?.message` para mostrar "Ya tenés este archivo subido" / "Alcanzaste el límite de documentos".

```ts
onError: (err: any) => {
  const msg = err.response?.data?.message ?? err.message
  addToast("Error", msg, "error")
}
```

---

### 2. Endpoint `/me` — no consumido

El backend expone `GET /api/v1/users/me` con:
```json
{
  "id": 1, "name": "...", "email": "...", "role": "editor",
  "max_documents": 10, "document_count": 3,
  "chat_quota": 500, "chat_quota_used": 42
}
```
El tipo `User` en `types/index.ts` ya tiene los campos. Nadie los muestra en la UI.

**Pendiente**: mostrar en navbar o settings la cuota de chat (`chat_quota_used / chat_quota`) y el uso de documentos (`document_count / max_documents`).

---

### 3. Página pública de grupos — no existe

El backend tiene `/api/v1/public/groups/:slug` (sin auth). No hay ninguna ruta en el frontend para esto.

**Pendiente**: crear `/public/groups/$slug.tsx` (ruta pública, fuera de `_authenticated`) con:
- Fetch `GET /api/v1/public/groups/:slug` → nombre, documentos, allow_downloads
- Input de chat + llamada a `POST /api/v1/public/groups/:slug/chat`
- Lista de documentos con botón de descarga si `allow_downloads: true`
- Manejo de 404 (inactivo) y 429 (cuota excedida)

---

### 4. Groups page — stub incompleto

| Feature | Estado |
|---|---|
| Listar grupos | ✅ funciona (API paginada, frontend ignora total/page/limit) |
| Toggle activo/inactivo | ❌ `console.log` — falta `PATCH /groups/:id` |
| Eliminar grupo | ❌ `console.log` — falta `DELETE /groups/:id` |
| Crear grupo | ❌ botón "Nuevo grupo" no hace nada |
| Editar nombre / downloads | ❌ no implementado |
| Agregar/quitar documentos | ❌ no implementado |
| Copiar enlace (`/public/groups/:slug`) | ❌ botón no hace nada |
| QR del slug | ❌ botón no hace nada |
| Paginación | ❌ `GroupsResponse` no tiene `total`/`page`/`limit`, UI no pagina |

---

### 5. Download de documentos — stub

En `_authenticated.documents.tsx`:
```ts
onDownload={(id) => console.log(`download ${id}`)}
```
Falta llamar a `GET /api/v1/documents/:id` (o implementar presigned URL / proxy de descarga).

---

### Resumen de prioridades sugeridas

1. **Groups page** — es la feature central del producto, está scaffolded pero sin funcionalidad real
2. **Página pública** — el endpoint está, solo falta el frontend
3. **Errores de upload** — fix de 3 líneas, mejora inmediata
4. **`/me` + cuotas** — mostrar al usuario sus límites actuales
5. **Download de documentos** — flujo básico de uso

---

## Pendientes de diseño / decisiones técnicas

### A. Notificación al dueño cuando el chat está bloqueado por cuota agotada

Hoy el dueño no sabe que su link está siendo usado pero rechazado. Los datos existen pero no hay notificación activa.

#### Comportamiento actual de los contadores

El contador `chat_attempts` se incrementa en **toda** petición al chat del grupo, incluyendo las rechazadas. `chat_quota_used` solo se incrementa en chats exitosos. Por lo tanto:

- Usuario externo intenta chatear → grupo agotó su cuota (`group.chat_quota_used >= group.chat_quota`) → `attempts +1`, `quota_used +0`
- Usuario externo intenta chatear → el dueño del grupo agotó su cuota de usuario (`user.chat_quota_used >= user.chat_quota`) → `attempts +1`, `quota_used +0`

En ambos casos el resultado en DB es idéntico — **no se puede distinguir cuál fue el motivo del bloqueo** con los contadores actuales.

#### Opciones de implementación

**Opción 1 — Pasiva (UI badge, costo cero):** si `chat_quota_used >= chat_quota` o `user.chat_quota_used >= user.chat_quota`, mostrar banner en el detalle del grupo: *"Tu grupo está bloqueado — cuota agotada"*. Datos ya disponibles con `/me` + el grupo.

**Opción 2 — Activa (SSE):** emitir evento `quota_exceeded` desde `public_chat_usecase.go` vía `SSEManager` cuando se rechaza por cuota. Requiere que el dueño esté conectado al SSE stream.

**Pendiente también:** separar los contadores para distinguir causa de rechazo. Alternativas:
- Agregar columnas `quota_user_blocked int` y `quota_group_blocked int` en `document_groups`
- O tabla de eventos de rechazo con timestamp y motivo

**Decisión pendiente:** implementar al menos Opción 1 junto con la pantalla de detalle de grupo. Opción 2 y separación de contadores quedan para después.
