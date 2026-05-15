# Datos disponibles para el Dashboard

## APIs ya existentes

| Endpoint | Auth | Retorna |
|---|---|---|
| `GET /users/me` | JWT | `{id, name, email, role, max_documents, document_count, chat_quota, chat_quota_used}` |
| `GET /users/me/stats` | JWT | `{total_documents, by_status: {completed, failed, pending}, storage_used, total_groups, active_groups, total_attempts, unanswered_total}` |
| `GET /documents/?page=&limit=` | JWT | `{items: [Document], total, page, limit}` |
| `GET /groups/?page=&limit=` | JWT | `{items: [Group], total, page, limit}` |
| `GET /groups/:id` | JWT | `Group` con `document_ids` resueltos |
| `GET /groups/:id/usage` | JWT | `{total_calls, input_tokens, output_tokens}` |
| `GET /chats/` | JWT | `[ChatSession]` (sesiones de chat autenticado, no públicas) |

---

## Tablas y qué datos tienen

### `users`
| Campo | Descripción |
|---|---|
| `max_documents` | Límite de docs del usuario (0 = ilimitado) |
| `chat_quota` | Límite de chats global (0 = ilimitado) |
| `chat_quota_used` | Chats ya consumidos |
| `role_id` | 1=admin, 2=editor, 3=viewer |
| `created_at` | Fecha de registro |

### `documents`
| Campo | Descripción |
|---|---|
| `status` | `pending`, `processing`, `completed`, `failed` |
| `size` | Bytes del archivo |
| `content_type` | MIME (pdf, csv, json, xlsx, txt, docx) |
| `error_message` | Error si falló el procesamiento |
| `retry_count` | Reintentos realizados |
| `checksum` | SHA del archivo (para deduplicación) |
| `created_at` | Fecha de subida |

### `processing_steps`
| Campo | Descripción |
|---|---|
| `step_name` | `download`, `parse`, `chunk`, `embed`, `upsert` |
| `status` | `started`, `completed`, `failed` |
| `duration_ms` | Milisegundos que tomó el paso |
| `error_message` | Error si falló |

### `document_groups`
| Campo | Descripción |
|---|---|
| `is_active` | Grupo activo/inactivo |
| `allow_downloads` | Si permite descargar documentos vía link público |
| `chat_quota` | Límite de chats del grupo (0 = ilimitado) |
| `chat_quota_used` | Chats ya consumidos en ese grupo |
| `chat_attempts` | Intentos totales (incluye errores, quotas, etc.) |
| `unanswered_count` | Veces que el agente respondió "no tengo info suficiente" |
| `slug` | ID público para compartir (`/c/XXXXXX`) |
| `created_at` | Fecha de creación |

### `document_group_items`
| Campo | Descripción |
|---|---|
| `group_id` + `document_id` | Relación N:M entre grupos y documentos |

### `llm_usages`
| Campo | Descripción |
|---|---|
| `user_id` | Usuario dueño |
| `group_id` | Grupo desde el que se consultó (NULL si fue chat autenticado) |
| `model` | Modelo de LLM usado |
| `input_tokens` | Tokens de entrada |
| `output_tokens` | Tokens de salida |
| `created_at` | Fecha de la consulta |

### `chat_sessions` (chat autenticado, no público)
| Campo | Descripción |
|---|---|
| `user_id` | Usuario dueño |
| `title` | Título de la sesión |

### `chat_messages` (chat autenticado, no público)
| Campo | Descripción |
|---|---|
| `session_id` | FK a chat_sessions |
| `role` | `user` o `assistant` |
| `content` | Texto del mensaje |
| `sources` | JSONB con citas de documentos |

---

## Métricas que se pueden obtener

### A nivel usuario (scope: userID actual)

| Métrica | Cómo obtenerla |
|---|---|
| **Total documentos** | `documents` count WHERE user_id |
| **Docs por estado** | `documents` count WHERE user_id GROUP BY status |
| **Docs completados/fallidos/pendientes** | COUNT con `status = 'completed'/'failed'/'pending'` |
| **Docs procesando** | `total - completados - fallidos - pendientes` |
| **Almacenamiento usado** | SUM(`size`) WHERE user_id |
| **% de cuota de docs** | `document_count / max_documents * 100` |
| **% de cuota de chats** | `chat_quota_used / chat_quota * 100` |
| **Total grupos** | `document_groups` count WHERE user_id |
| **Grupos activos** | Count WHERE user_id AND `is_active = true` |
| **Grupos con downloads permitidos** | Count WHERE `allow_downloads = true` |
| **Intentos de chat totales** | SUM(`chat_attempts`) across user groups |
| **Consultas respondidas** | `total_attempts - unanswered_total` |
| **% sin respuesta** | `unanswered_total / total_attempts * 100` |
| **Preguntas sin contexto** | SUM(`unanswered_count`) across user groups |
| **Cuota por grupo** | `chat_quota_used / chat_quota` por cada grupo |
| **Grupos al límite** | Count WHERE `chat_quota > 0 AND chat_quota_used >= chat_quota` |
| **Grupos sin docs** | Count WHERE `document_ids` array empty |
| **Tokens consumidos** | SUM(`input_tokens + output_tokens`) FROM `llm_usages` WHERE user_id |
| **Tokens por día/semana/mes** | Tokens agrupados por `created_at` con trunc de fecha |
| **Modelos más usados** | Count GROUP BY `model` en llm_usages |
| **Costo estimado** | Tokens × precio por modelo (requiere lookup table externa) |
| **Consultas hoy** | Count llm_usages WHERE `created_at >= today` |
| **Última consulta** | MAX(`created_at`) FROM llm_usages WHERE user_id |
| **Último doc subido** | MAX(`created_at`) FROM documents WHERE user_id |
| **Sesiones de chat autenticado** | Count FROM `chat_sessions` WHERE user_id |
| **Mensajes intercambiados** | Count FROM `chat_messages` JOIN chat_sessions |
| **Tasa de error en procesamiento** | `failed / total_documents * 100` |
| **Tamaño promedio de documento** | AVG(`size`) WHERE user_id |
| **Docs duplicados (mismo checksum)** | Documents with same checksum per user |
| **Tiempo promedio de procesamiento** | AVG(`duration_ms`) FROM `processing_steps` WHERE step_name='upsert' AND status='completed' (el último paso) |
| **Pasos que más fallan** | Count FROM processing_steps WHERE status='failed' GROUP BY step_name |

### A nivel grupo individual (scope: groupID)

| Métrica | Dónde está |
|---|---|
| **Docs en el grupo** | `group.document_ids.length` |
| **Chats consumidos del cupo** | `group.chat_quota_used / group.chat_quota` |
| **Intentos totales** | `group.chat_attempts` |
| **Sin respuesta** | `group.unanswered_count` |
| **Actividad reciente** | `group.chat_attempts > 0` |
| **Grupo al límite** | `chat_quota > 0 && chat_quota_used >= chat_quota` |

### A nivel admin (scope: todos los usuarios)

Todo lo anterior, más:

| Métrica | Cómo |
|---|---|
| **Total usuarios** | Count FROM `users` |
| **Usuarios por rol** | Count GROUP BY `role_id` |
| **Nuevos usuarios (hoy/semana/mes)** | Count WHERE `created_at >= today/this week` |
| **Usuarios activos** | Users con `chat_quota_used > 0` o `documents > 0` |
| **Docs totales del sistema** | Count FROM `documents` |
| **Grupos totales/públicos** | Count FROM `document_groups` |
| **Tokens totales del sistema** | SUM FROM `llm_usages` |
| **Top grupos por uso** | ORDER BY `chat_attempts` DESC LIMIT N |

---

## Por tipo de documento (PDF, CSV, JSON, etc.)

| Métrica | Cómo |
|---|---|
| **Cantidad por tipo** | `documents` GROUP BY content_type |
| **Tamaño total por tipo** | SUM(`size`) GROUP BY content_type |
| **Tasa de éxito por tipo** | completed/total per content_type |
| **Tipo más fallido** | content_type con mayor `failed/total` ratio |

---

## Lo que NO está implementado aún

| Métrica | Qué falta |
|---|---|
| **Actividad por hora/día/semana** | Agrupar por `created_at` truncado en queries |
| **Tendencia (últimos 7/30 días)** | Varios COUNT con `WHERE created_at >=` |
| **Comparativa mes actual vs anterior** | Dos queries de rango de fechas |
| **Retención de usuarios** | Comparar docs/chats de 30d vs 60d |
| **RAG quality score** | No se trackea (necesitaría feedback del usuario) |
| **Fuentes más citadas por el LLM** | `chat_messages.sources` JSONB (solo chat autenticado) |
| **Latencia promedio del LLM** | No se trackea la duración de `GenerateAnswer` |

---

## Endpoint sugerido para extender

Actualmente `/users/me/stats` ya retorna:
```
total_documents, by_status {completed, failed, pending},
storage_used, total_groups, active_groups,
total_attempts, unanswered_total
```

Se puede extender para agregar (sin romper compatibilidad):
```
chat_quota_used, chat_quota, document_count, max_documents,
tokens_consumed, tokens_last_30d, last_activity_at,
groups_at_limit, docs_by_type: {pdf, csv, json, ...},
processing_error_rate, avg_processing_time_ms
```
