package application

import (
	"context"
	"fmt"
	"time"

	"github.com/tmc/langchaingo/schema"
	"github.com/whoAngeel/rago/internal/core/domain"
	"github.com/whoAngeel/rago/internal/core/ports"
	"github.com/whoAngeel/rago/internal/infrastructure/config"
)

type StepFunc func(stepName string, start time.Time, err error)

type IngestUsecase struct {
	VectorStore ports.VectorStore
	Embedder    ports.Embedder
	Logger      ports.Logger
	config      config.Config
}

func NewIngestUsecase(
	vStore ports.VectorStore,
	embedder ports.Embedder,
	logger ports.Logger,
	config config.Config,
) *IngestUsecase {
	return &IngestUsecase{
		VectorStore: vStore,
		Embedder:    embedder,
		Logger:      logger,
		config:      config,
	}
}

func (iu *IngestUsecase) Execute(ctx context.Context, doc *domain.Document, metadata map[string]any, chunks []string, recordStep StepFunc) error {

	var allDocs []schema.Document
	var allVectors [][]float32
	var embedErr error

	embedStart := time.Now()
	for _, chunk := range chunks {
		vectors, err := iu.Embedder.EmbedText(ctx, chunk)
		if err != nil {
			iu.Logger.Warn("Error embedding text", "error", err)
			embedErr = err
			continue
		}
		allVectors = append(allVectors, vectors)
		allDocs = append(allDocs, schema.Document{
			PageContent: chunk,
			Metadata:    buildMetadata(doc, metadata),
		})
	}
	recordStep("embed", embedStart, embedErr)

	if len(allDocs) == 0 {
		return fmt.Errorf("no chunks embedded")
	}

	upsertStart := time.Now()
	err := iu.VectorStore.UpsertDocuments(ctx, iu.config.QdrantCollection, allDocs, allVectors)
	recordStep("upsert", upsertStart, err)
	if err != nil {
		return fmt.Errorf("error upserting: %w", err)
	}

	iu.Logger.Info("document ingested", "filename", doc.Filename)
	return nil
}

func buildMetadata(doc *domain.Document, metadata map[string]any) map[string]any {
	merged := map[string]any{
		"source":       doc.Filename,
		"user_id":      doc.UserID,
		"content_type": doc.ContentType,
		"document_id":  doc.ID,
	}

	// Merge: sobrescribe si hay claves repetidas
	for k, v := range metadata {
		merged[k] = v
	}

	return merged
}
