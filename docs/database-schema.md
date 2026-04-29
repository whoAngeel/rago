# RAGO - Database Schema Design

## Relational Diagram

```mermaid
erDiagram
    roles ||--o{ users : "has"
    users ||--o{ sessions : "owns"
    users ||--o{ documents : "owns"
    users ||--o{ chat_sessions : "creates"
    chat_sessions ||--o{ chat_messages : "contains"

    roles {
        int id PK
        string name UK "admin, editor, viewer"
        text description
        timestamp created_at
    }

    users {
        int id PK
        string email UK
        string password
        string name
        int role_id FK "RESTRICT on delete"
        timestamp created_at
        timestamp updated_at
    }

    sessions {
        int id PK
        int user_id FK "CASCADE on delete"
        string refresh_token UK
        string access_token
        timestamp expires_at
        timestamp revoked_at
    }

    documents {
        int id PK
        int user_id FK "CASCADE on delete"
        string filename
        string file_path
        string content_type
        string status "pending, processing, completed, failed"
        bigint size
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at "soft delete"
    }

    chat_sessions {
        int id PK
        int user_id FK "CASCADE on delete"
        string title
        timestamp created_at
        timestamp updated_at
    }

    chat_messages {
        int id PK
        int session_id FK "CASCADE on delete"
        string role "user | assistant"
        text content
        json sources "document citations"
        timestamp created_at
    }
```

## Tables Detail

### `roles`

| Column      | Type          | Constraints          | Description               |
| ----------- | ------------- | -------------------- | ------------------------- |
| `id`        | int (SERIAL)  | PK                   | Auto-increment            |
| `name`      | string (50)   | UNIQUE, NOT NULL     | `admin`, `editor`, `viewer` |
| `description` | text        |                      | Role description          |
| `created_at`| timestamp     | NOT NULL, DEFAULT NOW|                           |

Seed: `admin` (id=1), `editor` (id=2), `viewer` (id=3)

### `users`

| Column       | Type           | Constraints                       | Description          |
| ------------ | -------------- | --------------------------------- | -------------------- |
| `id`         | int (SERIAL)   | PK                                |                      |
| `email`      | string (255)   | UNIQUE, NOT NULL                  |                      |
| `password`   | string (255)   | NOT NULL                          | Bcrypt hash          |
| `name`       | string (255)   | DEFAULT ''                        |                      |
| `role_id`    | int            | FK → roles(id), RESTRICT on delete| Default: 3 (viewer)  |
| `created_at` | timestamp      | NOT NULL, DEFAULT NOW             |                      |
| `updated_at` | timestamp      | DEFAULT NOW                       |                      |

### `sessions`

| Column          | Type           | Constraints                          | Description               |
| --------------- | -------------- | ------------------------------------ | ------------------------- |
| `id`            | int (SERIAL)   | PK                                   |                           |
| `user_id`       | int            | FK → users(id), **CASCADE on delete**|                           |
| `refresh_token` | string (255)   | UNIQUE, NOT NULL                     | Session token             |
| `access_token`  | string (255)   |                                      | JWT access                |
| `expires_at`    | timestamp      | NOT NULL                             | Token expiry              |
| `revoked_at`    | timestamp NULL |                                      | Soft revoke               |

### `documents`

| Column          | Type           | Constraints                          | Description                         |
| --------------- | -------------- | ------------------------------------ | ----------------------------------- |
| `id`            | int (SERIAL)   | PK                                   |                                     |
| `user_id`       | int            | FK → users(id), **CASCADE on delete**|                                     |
| `filename`      | text           | NOT NULL                             | Original filename                   |
| `file_path`     | text           |                                      | Physical path in BlobStorage        |
| `content_type`  | text           |                                      | MIME type                           |
| `status`        | text           | DEFAULT `'pending'`                  | `pending`, `processing`, `completed`, `failed` |
| `size`          | bigint         |                                      | File size in bytes                  |
| `created_at`    | timestamp      | NOT NULL, DEFAULT NOW                |                                     |
| `updated_at`    | timestamp NULL |                                      |                                     |
| `deleted_at`    | timestamp NULL | INDEX                                | Soft delete (GORM)                  |

### `chat_sessions` *(Roadmap 1.6)*

| Column       | Type           | Constraints                          | Description          |
| ------------ | -------------- | ------------------------------------ | -------------------- |
| `id`         | int (SERIAL)   | PK                                   |                      |
| `user_id`    | int            | FK → users(id), **CASCADE on delete**|                      |
| `title`      | string (255)   |                                      | Session title        |
| `created_at` | timestamp      | NOT NULL, DEFAULT NOW                |                      |
| `updated_at` | timestamp      | DEFAULT NOW                          |                      |

### `chat_messages` *(Roadmap 1.6)*

| Column        | Type           | Constraints                               | Description              |
| ------------- | -------------- | ----------------------------------------- | ------------------------ |
| `id`          | int (SERIAL)   | PK                                        |                          |
| `session_id`  | int            | FK → chat_sessions(id), **CASCADE on delete** |                          |
| `role`        | string (20)    | NOT NULL                                  | `user` or `assistant`    |
| `content`     | text           | NOT NULL                                  | Message body             |
| `sources`     | jsonb          |                                           | Citations / document refs|
| `created_at`  | timestamp      | NOT NULL, DEFAULT NOW                     |                          |

## Constraints Summary

| Relationship              | On Delete | On Update |
| ------------------------- | --------- | --------- |
| `users.role_id` → `roles.id`       | RESTRICT  | CASCADE   |
| `sessions.user_id` → `users.id`    | CASCADE   | CASCADE   |
| `documents.user_id` → `users.id`   | CASCADE   | CASCADE   |
| `chat_sessions.user_id` → `users.id` | CASCADE | CASCADE   |
| `chat_messages.session_id` → `chat_sessions.id` | CASCADE | CASCADE |

## Roadmap Mapping

| Release   | Tables Affected                     |
| --------- | ----------------------------------- |
| **1.2**   | `roles`, `users`, `sessions`        |
| **1.3**   | `documents`                         |
| **1.4**   | `documents` (status flow)           |
| **1.6**   | `chat_sessions`, `chat_messages`    |
