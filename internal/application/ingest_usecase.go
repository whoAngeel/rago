package application

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/schema"
	"github.com/whoAngeel/rago/internal/core/domain"
	"github.com/whoAngeel/rago/internal/core/ports"
	"github.com/whoAngeel/rago/internal/infrastructure/config"
)

// StepFunc is called before a step starts; returns a done func to call when the step finishes.
type StepFunc func(stepName string) (done func(err error))

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

	doneEmbed := recordStep("embed")
	for i, chunk := range chunks {
		preview := chunk
		if len(preview) > 100 {
			preview = preview[:100] + "..."
		}
		vectors, err := iu.Embedder.EmbedText(ctx, chunk)
		if err != nil {
			iu.Logger.Warn("Error embedding text", "error", err, "chunk_idx", i, "chunk_len", len(chunk), "chunk_preview", preview)
			embedErr = err
			continue
		}
		allVectors = append(allVectors, vectors)
		allDocs = append(allDocs, schema.Document{
			PageContent: chunk,
			Metadata:    buildMetadata(doc, metadata),
		})
	}
	doneEmbed(embedErr)

	if len(allDocs) == 0 {
		return fmt.Errorf("no chunks embedded")
	}

	if err := iu.VectorStore.CreateCollection(ctx, iu.config.QdrantCollection, iu.config.EmbeddingDim); err != nil {
		iu.Logger.Warn("collection may already exist", "error", err)
	}

	doneUpsert := recordStep("upsert")
	err := iu.VectorStore.UpsertDocuments(ctx, iu.config.QdrantCollection, allDocs, allVectors)
	doneUpsert(err)
	if err != nil {
		return fmt.Errorf("error upserting: %w", err)
	}

	iu.Logger.Info("document ingested", "filename", doc.Filename)
	return nil
}

func (iu *IngestUsecase) GetPointsCountByDocumentID(ctx context.Context, documentID int) (uint64, error) {
	return iu.VectorStore.GetPointsCountByFilter(ctx, iu.config.QdrantCollection, documentID)
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
