package application

import (
	"context"
	"io"

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
) error {
	return nil
}

func (i *IngestDocumentUsecase) ProcessDocument(ctxt context.Context, docID int) error {
	return nil
}
