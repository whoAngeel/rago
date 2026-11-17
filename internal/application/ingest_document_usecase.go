package application

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	"github.com/whoAngeel/rago/internal/core/domain"
	"github.com/whoAngeel/rago/internal/core/ports"
)

var (
	ErrNotFound             = errors.New("not found")
	ErrDuplicateDocument    = errors.New("duplicate document")
	ErrDocumentLimitReached = errors.New("document limit reached")
)

type IngestDocumentUsecase struct {
	DocRepo     ports.DocumentRepository
	UserRepo    ports.UserRepository
	BlobStorage ports.BlobStorage
	IngestUC    *IngestUsecase
}

func NewIngestDocumentUsecase(
	docRepo ports.DocumentRepository,
	userRepo ports.UserRepository,
	blobStorage ports.BlobStorage,
	ingestUC *IngestUsecase,
) *IngestDocumentUsecase {
	return &IngestDocumentUsecase{
		DocRepo:     docRepo,
		UserRepo:    userRepo,
		BlobStorage: blobStorage,
		IngestUC:    ingestUC,
	}
}

func (i *IngestDocumentUsecase) Upload(
	ctx context.Context,
	userID int,
	filename string,
	fileReader io.Reader,
	size int64,
	contentType string,
) (*domain.Document, error) {
	// 1. Verificar límite de documentos (0 = ilimitado)
	user, err := i.UserRepo.FindById(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.MaxDocuments > 0 {
		count, err := i.DocRepo.CountDocumentsByUserID(ctx, userID)
		if err != nil {
			return nil, err
		}
		if int(count) >= user.MaxDocuments {
			return nil, ErrDocumentLimitReached
		}
	}

	// 2. Leer en buffer + calcular SHA-256 en una sola pasada
	h := sha256.New()
	buf, err := io.ReadAll(io.TeeReader(fileReader, h))
	if err != nil {
		return nil, err
	}
	checksum := hex.EncodeToString(h.Sum(nil))

	// 3. Verificar duplicado por usuario
	existing, err := i.DocRepo.FindByChecksum(ctx, userID, checksum)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return i.handleDuplicate(ctx, existing, buf)
	}

	// 4. Crear registro con checksum
	doc := &domain.Document{
		Filename:    filename,
		UserID:      userID,
		Status:      domain.StatusUploading,
		ContentType: contentType,
		Size:        size,
		Checksum:    &checksum,
		// Set before the blob is stored so a concurrent delete of another
		// document with the same content sees this row and keeps the blob.
		FilePath: blobKey(checksum),
	}
	document, err := i.DocRepo.CreateDocument(ctx, doc)
	if err != nil {
		return nil, err
	}

	// 5. Guardar el blob una sola vez por contenido
	objectKey := blobKey(checksum)
	if err := i.ensureBlob(ctx, objectKey, buf); err != nil {
		i.DocRepo.UpdateDocumentStatus(ctx, document.ID, domain.StatusFailed)
		return nil, err
	}
	document.FilePath = objectKey
	document.Status = domain.StatusPending
	_, err = i.DocRepo.UpdateDocument(ctx, document)
	return document, err
}

// blobKey is the content-addressed storage key for a file.
func blobKey(checksum string) string { return "blobs/" + checksum }

// ensureBlob uploads buf under key unless an object already exists there.
func (i *IngestDocumentUsecase) ensureBlob(ctx context.Context, key string, buf []byte) error {
	exists, err := i.BlobStorage.Exists(ctx, key)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	_, err = i.BlobStorage.Upload(ctx, key, bytes.NewReader(buf))
	return err
}

// handleDuplicate resolves an upload whose content the user already has. A
// failed document is requeued and the upload counts as successful; any other
// status returns the existing document together with ErrDuplicateDocument.
func (i *IngestDocumentUsecase) handleDuplicate(ctx context.Context, existing *domain.Document, buf []byte) (*domain.Document, error) {
	if existing.Status != domain.StatusFailed {
		return existing, ErrDuplicateDocument
	}
	// Make sure the blob is there: the earlier upload may have failed before
	// storing it, or a concurrent delete may have removed a shared blob.
	key := existing.FilePath
	if key == "" {
		key = blobKey(*existing.Checksum)
	}
	if err := i.ensureBlob(ctx, key, buf); err != nil {
		return nil, err
	}
	if existing.FilePath == "" {
		existing.FilePath = key
		if _, err := i.DocRepo.UpdateDocument(ctx, existing); err != nil {
			return nil, err
		}
	}
	if err := i.DocRepo.RequeueDocument(ctx, existing.ID); err != nil {
		return nil, err
	}
	existing.Status = domain.StatusPending
	existing.RetryCount = 0
	existing.ErrorMessage = ""
	return existing, nil
}

// DeleteDocument removes a document owned by userID. A document that does not
// exist or belongs to another user yields ErrNotFound, so callers cannot tell
// the two cases apart.
func (i *IngestDocumentUsecase) DeleteDocument(ctx context.Context, docID, userID int) error {
	doc, err := i.DocRepo.FindByID(ctx, docID)
	if err != nil || doc == nil || doc.UserID != userID {
		return ErrNotFound
	}
	if doc.FilePath != "" {
		// Blobs are shared by every document with the same content: keep the
		// object while another row still references it.
		others, err := i.DocRepo.CountByFilePath(ctx, doc.FilePath, doc.ID)
		if err != nil {
			return err
		}
		if others == 0 {
			if err := i.BlobStorage.Delete(ctx, doc.FilePath); err != nil {
				return err
			}
		}
	}
	if err := i.IngestUC.DeleteDocumentVectors(ctx, docID); err != nil {
		i.IngestUC.Logger.Warn("could not delete vectors on document deletion", "doc_id", docID, "error", err)
	}
	return i.DocRepo.DeleteDocument(ctx, doc.ID)
}

func (i *IngestDocumentUsecase) GetUsersDocuments(ctx context.Context, userID int, page, limit int) ([]*domain.Document, int64, error) {
	return i.DocRepo.FindDocumentByUserID(ctx, userID, page, limit)
}

func (i *IngestDocumentUsecase) GetDocumentSteps(ctx context.Context, docID, userID int) ([]*domain.ProcessingStep, error) {
	doc, err := i.DocRepo.FindByID(ctx, docID)
	if err != nil || doc.UserID != userID {
		return nil, ErrNotFound
	}
	steps, err := i.DocRepo.FindStepsByDocumentID(ctx, docID)
	if err != nil {
		return nil, err
	}
	if steps == nil {
		return []*domain.ProcessingStep{}, nil
	}
	return steps, nil
}

type DocumentHealth struct {
	Document      *domain.Document         `json:"document"`
	Steps         []*domain.ProcessingStep `json:"steps"`
	ChunksIndexed uint64                   `json:"chunks_indexed"`
}

func (i *IngestDocumentUsecase) GetDocumentHealth(ctx context.Context, docID, userID int) (*DocumentHealth, error) {
	doc, err := i.DocRepo.FindByID(ctx, docID)
	if err != nil || doc.UserID != userID {
		return nil, ErrNotFound
	}

	steps, err := i.DocRepo.FindStepsByDocumentID(ctx, docID)
	if err != nil {
		return nil, err
	}
	if steps == nil {
		steps = []*domain.ProcessingStep{}
	}

	chunksIndexed, err := i.IngestUC.GetPointsCountByDocumentID(ctx, docID)
	if err != nil {
		return nil, err
	}

	return &DocumentHealth{
		Document:      doc,
		Steps:         steps,
		ChunksIndexed: chunksIndexed,
	}, nil
}

func (uc *IngestDocumentUsecase) GetDocumentsForSelect(ctx context.Context, userID int) ([]*domain.Document, error) {
	return uc.DocRepo.FindDocumentsForSelect(ctx, userID)
}

func (uc *IngestDocumentUsecase) ReprocessDocument(ctx context.Context, docID, userID int, forceOCR bool) error {
	doc, err := uc.DocRepo.FindByID(ctx, docID)
	if err != nil || doc.UserID != userID {
		return ErrNotFound
	}
	if doc.Status != domain.StatusFailed && doc.Status != domain.StatusCompleted {
		return fmt.Errorf("solo se pueden reprocesar documentos fallidos o completados")
	}
	if err := uc.DocRepo.DeleteProcessingStepsByDocumentID(ctx, docID); err != nil {
		return err
	}
	doc.Status = domain.StatusPending
	doc.ErrorMessage = ""
	doc.ProcessingStartedAt = nil
	doc.ForceOCR = forceOCR
	if _, err := uc.DocRepo.UpdateDocument(ctx, doc); err != nil {
		return err
	}
	return nil
}
