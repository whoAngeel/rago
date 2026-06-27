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
		return nil, ErrDuplicateDocument
	}

	// 4. Crear registro con checksum
	doc := &domain.Document{
		Filename:    filename,
		UserID:      userID,
		Status:      domain.StatusUploading,
		ContentType: contentType,
		Size:        size,
		Checksum:    &checksum,
	}
	document, err := i.DocRepo.CreateDocument(ctx, doc)
	if err != nil {
		return nil, err
	}

	// 5. Subir desde buffer (fileReader ya fue consumido por TeeReader)
	objectKey := fmt.Sprintf("%d/%d/%s", userID, document.ID, filename)
	if _, err = i.BlobStorage.Upload(ctx, objectKey, bytes.NewReader(buf)); err != nil {
		i.DocRepo.UpdateDocumentStatus(ctx, document.ID, domain.StatusFailed)
		return nil, err
	}
	document.FilePath = objectKey
	document.Status = domain.StatusPending
	_, err = i.DocRepo.UpdateDocument(ctx, document)
	return document, err
}

func (i *IngestDocumentUsecase) DeleteDocument(ctx context.Context, docID int) error {
	doc, err := i.DocRepo.FindByID(ctx, docID)
	if err != nil {
		return err
	}
	if doc.FilePath != "" {
		if err := i.BlobStorage.Delete(ctx, doc.FilePath); err != nil {
			return err
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
	Document      *domain.Document        `json:"document"`
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
