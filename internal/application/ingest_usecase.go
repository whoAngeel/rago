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

// embedBatchSize is how many chunks go to the embedder per call.
const embedBatchSize = 32

// Execute embeds the chunks in batches and replaces the document's vectors.
// Any failed batch, or a batch returning the wrong number of vectors, fails
// the document: nothing is deleted or upserted, so the old vectors stay.
func (iu *IngestUsecase) Execute(ctx context.Context, doc *domain.Document, metadata map[string]any, chunks []string, recordStep StepFunc) error {
	allDocs := make([]schema.Document, 0, len(chunks))
	allVectors := make([][]float32, 0, len(chunks))

	doneEmbed := recordStep("embed")
	embedErr := func() error {
		for start := 0; start < len(chunks); start += embedBatchSize {
			end := min(start+embedBatchSize, len(chunks))
			batch := chunks[start:end]
			vectors, err := iu.Embedder.ComputeEmbeddings(ctx, batch)
			if err != nil {
				iu.Logger.Warn("Error embedding batch", "error", err, "first_chunk_idx", start, "batch_size", len(batch))
				return fmt.Errorf("embedding chunks %d-%d: %w", start, end-1, err)
			}
			if len(vectors) != len(batch) {
				return fmt.Errorf("embedding chunks %d-%d: got %d vectors for %d texts", start, end-1, len(vectors), len(batch))
			}
			for i, chunk := range batch {
				allVectors = append(allVectors, vectors[i])
				allDocs = append(allDocs, schema.Document{
					PageContent: chunk,
					Metadata:    buildMetadata(doc, metadata),
				})
			}
		}
		return nil
	}()
	doneEmbed(embedErr)
	if embedErr != nil {
		return embedErr
	}

	if len(allDocs) == 0 {
		return fmt.Errorf("no chunks embedded")
	}

	if err := iu.VectorStore.CreateCollection(ctx, iu.config.QdrantCollection, iu.config.EmbeddingDim); err != nil {
		iu.Logger.Warn("collection may already exist", "error", err)
	}

	if err := iu.VectorStore.DeleteByDocumentID(ctx, iu.config.QdrantCollection, doc.ID); err != nil {
		iu.Logger.Warn("could not delete old vectors before upsert", "doc_id", doc.ID, "error", err)
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

func (iu *IngestUsecase) DeleteDocumentVectors(ctx context.Context, documentID int) error {
	return iu.VectorStore.DeleteByDocumentID(ctx, iu.config.QdrantCollection, documentID)
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
