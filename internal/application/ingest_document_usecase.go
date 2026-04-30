package application

import (
	"context"
	"fmt"
	"io"

	"github.com/whoAngeel/rago/internal/core/domain"
	"github.com/whoAngeel/rago/internal/core/ports"
)

type IngestDocumentUsecase struct {
	DocRepo     ports.DocumentRepository
	BlobStorage ports.BlobStorage
	IngestUC    *IngestUsecase
}

func NewIngestDocumentUsecase(
	docRepo ports.DocumentRepository,
	blobStorage ports.BlobStorage,
	ingestUC *IngestUsecase,
) *IngestDocumentUsecase {
	return &IngestDocumentUsecase{
		DocRepo:     docRepo,
		BlobStorage: blobStorage,
		IngestUC:    ingestUC,
	}
}

func (i *IngestDocumentUsecase) Upload(
	ctx context.Context,
	userId int,
	filename string,
	fileReader io.Reader,
	size int64,
	contentType string,
) (*domain.Document, error) {
	doc := &domain.Document{
		Filename:    filename,
		UserID:      userId,
		Status:      domain.StatusPending,
		ContentType: contentType,
		Size:        size,
	}
	document, err := i.DocRepo.CreateDocument(ctx, doc)
	if err != nil {
		return nil, err
	}

	objectKey := fmt.Sprintf("%d/%d/%s", userId, document.ID, filename)
	_, err = i.BlobStorage.Upload(ctx, objectKey, fileReader)
	if err != nil {
		i.DocRepo.UpdateDocumentStatus(ctx, document.ID, domain.StatusFailed)
		return nil, err
	}
	document.FilePath = objectKey
	_, err = i.DocRepo.UpdateDocument(ctx, document)
	return doc, nil

}

func (i *IngestDocumentUsecase) Process(ctx context.Context, docID int) error {
	doc, err := i.DocRepo.FindByID(ctx, docID)
	if err != nil {
		return err
	}

	if err := i.DocRepo.UpdateDocumentStatus(ctx, doc.ID, domain.StatusProcessing); err != nil {
		return err
	}

	file, err := i.BlobStorage.Download(ctx, doc.FilePath)
	if err != nil {
		return err
	}

	content, err := io.ReadAll(file)
	file.Close()
	if err != nil {
		i.DocRepo.UpdateDocumentStatus(ctx, doc.ID, domain.StatusFailed)
		return err
	}

	if err := i.IngestUC.Execute(ctx, doc.Filename, string(content)); err != nil {
		i.DocRepo.UpdateDocumentStatus(ctx, doc.ID, domain.StatusFailed)
		return err
	}

	return i.DocRepo.UpdateDocumentStatus(ctx, doc.ID, domain.StatusCompleted)

}

func (i *IngestDocumentUsecase) DeleteDocument(ctx context.Context, docID int) error {
	doc, err := i.DocRepo.FindByID(ctx, docID)
	if err != nil {
		return err
	}
	if err := i.BlobStorage.Delete(ctx, doc.FilePath); err != nil {
		return err
	}

	return i.DocRepo.DeleteDocument(ctx, doc.ID)
}

func (i *IngestDocumentUsecase) GetUsersDocuments(ctx context.Context, userID int) ([]*domain.Document, error) {
	return i.DocRepo.FindDocumentByUserID(ctx, userID)
}
