# Guía: Release 1.3 - Gestión de Documentos y BlobStorage

## ✅ Completado

| Componente | Archivo |
|-----------|---------|
| Puerto BlobStorage | `ports/storage.go` |
| Document entity + DocumentStatus | `domain/document.go`, `domain/document_status.go` |
| Puerto DocumentRepository | `ports/database.go` |
| Adapter MinIO | `storage/minio.go` |
| Adapter PostgreSQL Document | `postgres/document_repo.go` |
| IngestDocumentUsecase | `application/ingest_document_usecase.go` |
| DocumentHandler (List, Upload) | `rest/handlers/document.go` |
| Validaciones extensión + tamaño | handler Upload |
| Config MinIO + MAX_FILE_SIZE | `config/config.go` |
| AutoMigrate por modelo | `cmd/server/main.go` |

---

## ⬜ Release 1.3 - Faltante Final

### Delete Handler

Agrega en `rest/handlers/document.go`:

| Endpoint | Método | Auth | Response |
|----------|--------|------|----------|
| `/api/v1/documents/:id` | DELETE | Sí | 204 No Content |

```go
func (h *DocumentHandler) Delete(c *gin.Context) {
    ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
    defer cancel()

    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        rest.RespondError(c, 400, "Invalid document ID", err.Error())
        return
    }

    if err := h.usecase.DeleteDocument(ctx, id); err != nil {
        rest.RespondError(c, 500, "Delete failed", err.Error())
        return
    }

    c.JSON(http.StatusNoContent, nil)
}
```

**Ruta en handlers.go:**
```go
protected.DELETE("/documents/:id", docHandler.Delete)
```

---

## Decisiones aplicadas de architecture-decisions.md

| Decisión | Aplicada |
|----------|----------|
| 3.6 Cleanup sincrónico: MinIO first, luego BD | ✅ DeleteDocument use case |
| 3.7 MAX_FILE_SIZE | ✅ Validación en handler |
| 3.8 Validación por extensión | ✅ En Upload handler |
| 1.2 Hard delete en Documents | ✅ Sin DeletedAt |
| 1.4 FK CASCADE documents → users | ✅ En GORM AutoMigrate |

---

## Siguiente Fase: Release 1.4

Ver `docs/phase_1.4_guide.md` para el Worker de ingesta asíncrona.
