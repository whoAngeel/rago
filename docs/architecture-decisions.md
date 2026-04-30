# RAGO - Architecture Decisions Log

Documento de referencia con todas las decisiones de diseño tomadas durante el desarrollo.
Úsalo como contexto para continuar el trabajo en otra máquina o sesión.

---

## 1. Database Schema Decisions

### 1.1 DocumentStatus como tipo custom
- `DocumentStatus` es un custom type con valores: `pending`, `processing`, `completed`, `failed`
- Implementa `database/sql.Scanner` y `driver.Valuer` para compatibilidad con GORM
- Validación a nivel de aplicación con método `Valid()`
- File: `internal/core/domain/document_status.go`

### 1.2 Hard delete en Documents
- No hay soft delete (`DeletedAt`) en la tabla `documents`
- Si un documento se borra, sus referencias en `chat_messages.sources` quedan huérfanas
- El frontend maneja fuentes huérfanas mostrando "Documento eliminado"

### 1.3 Role usa gorm.Model
- `Role` embebe `gorm.Model` para tener `CreatedAt`, `UpdatedAt`, `DeletedAt`
- No tiene campo `description` (roles son autoexplicativos: admin, editor, viewer)

### 1.4 Foreign Key Constraints
| Relación | On Delete | On Update |
|---|---|---|
| `users.role_id` → `roles.id` | RESTRICT | CASCADE |
| `sessions.user_id` → `users.id` | CASCADE | CASCADE |
| `documents.user_id` → `users.id` | CASCADE | CASCADE |

### 1.5 chat_sessions.title
- Auto-generado del primer mensaje del usuario
- Editable vía PATCH por el usuario

### 1.6 chat_messages.sources (JSONB)
Estructura de cada referencia:
```json
{
  "document_id": 5,
  "chunk_id": "qdrant_point_abc",
  "page": 3,
  "text_preview": "..."
}
```

### 1.6 ContentType como único selector de parser
- No se necesita campo `FileType` o `ParserType` separado
- Se deriva el parser del MIME type (`content_type`)

---

## 2. Authentication Decisions

### 2.1 Token Storage en Frontend
- **Refresh token**: HttpOnly Cookie (no accesible por JS)
- **Access token**: Devuelto en body de login, guardado en memoria del frontend (NO LocalStorage)
- Patrón híbrido: seguridad de refresh token + simplicidad de access token stateless

### 2.2 Password Hashing
- **bcrypt** con costo por defecto (10)
- Librería: `golang.org/x/crypto/bcrypt`

### 2.3 Token Durations
- **Access token**: 15 minutos
- **Refresh token**: 7 días (extensible con "Remember me" en futuro)

### 2.4 Refresh Token Rotation
- Cada refresh genera un nuevo refresh token
- El anterior se revoca automáticamente
- Implementación: actualizar `refresh_token` y `expires_at` en tabla `sessions`

### 2.5 Brute Force Protection
- **Rate limiting por IP**: 5 intentos de login por minuto
- No hay account lockout temporal (se puede agregar en futuro si es necesario)

### 2.6 User Registration
- **Registro abierto** con `POST /register`
- Rol por defecto: `viewer`
- Admin puede cambiar rol manualmente

### 2.7 Forgot Password
- **No implementado por ahora**
- Se puede resetear manualmente en BD o vía admin
- Diseñar interfaz extensible para futuro email-based reset

### 2.8 Logout
- Revoke refresh token: set `revoked_at` en tabla `sessions`
- Borrar HttpOnly Cookie del cliente
- Access token sigue válido hasta expiración (15 min) — comportamiento esperado con JWT stateless

---

## 3. Storage Decisions

### 3.1 Backend de Almacenamiento
- **MinIO** como único BlobStorage (ya montado en homelab)
- No hay LocalFS como fallback
- La interfaz `BlobStorage` se implementa como `MinioAdapter`
- MinIO es S3-compatible, futuro cambio a S3 requiere mínimo código

### 3.2 Bucket Strategy
- **Un solo bucket** para todos los usuarios (ej: `rago-documents`)
- Organización por prefijo: `{user_id}/`
- No se crean buckets por usuario

### 3.3 Object Key Pattern
- Formato: `{user_id}/{document_id}/{filename}`
- Ej: `42/123/reporte.pdf`
- `document_id` es el ID autoincremental de la BD

### 3.4 Upload Strategy
- Stream directo del request body a MinIO (sin buffer completo en memoria)
- Si falla, el registro en BD queda con estado `FAILED`
- Cleanup de objetos parciales: worker o rutina de limpieza

### 3.5 Deduplicación
- **Sin deduplicación** por ahora
- Cada upload genera un nuevo archivo en MinIO, incluso si el contenido es idéntico

### 3.6 Cleanup de Archivos Huérfanos
- **Borrado sincrónico**: al hacer `DELETE /documents/:id`
- Primero se borra de MinIO, luego de la BD
- Si el borrado de MinIO falla, la transacción de BD no se hace

### 3.7 Tamaño Máximo de Archivo
- Configurable por variable de entorno: `MAX_FILE_SIZE`
- Default: 50MB (52428800 bytes)

### 3.8 Validación de Tipo de Archivo
- **Validar por extensión** en el endpoint antes de guardar en MinIO
- Extensiones soportadas: PDF, DOCX, XLSX, CSV, JSON, TXT (extensible)
- Validación de magic bytes no se necesita por ahora
- Si alguien sube un archivo con extensión válida pero contenido corrupto, el worker lo marca como `FAILED`

---

## 4. Tables Status

### Existing (implemented)
- `roles` — Con gorm.Model
- `users` — Con FK a roles
- `sessions` — JWT tokens, FK a users (CASCADE)
- `documents` — Con DocumentStatus, FK a users (CASCADE)

### Planned (Roadmap 1.6)
- `chat_sessions` — User, title, timestamps
- `chat_messages` — Session, role, content, sources (JSONB)

---

## 5. Key Files

| File | Purpose |
|---|---|
| `internal/core/domain/user.go` | Role + User models |
| `internal/core/domain/session.go` | Session model (JWT tokens) |
| `internal/core/domain/document.go` | Document model |
| `internal/core/domain/document_status.go` | DocumentStatus custom type |
| `internal/core/ports/database.go` | Repository interfaces |
| `internal/infrastructure/postgres/*` | GORM implementations |
| `cmd/server/main.go` | DB init, AutoMigrate, role seed |
| `docs/database-schema.md` | ER diagram + table details |
| `docs/architecture-decisions.md` | This file |
