package application

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/schema"
	"github.com/whoAngeel/rago/internal/core/ports"
	"github.com/whoAngeel/rago/internal/infrastructure/config"
)

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

func (iu *IngestUsecase) Execute(ctx context.Context, filename, content string) error {
	docs :=
		[]schema.Document{{
			PageContent: content,
			Metadata:    map[string]any{"source": filename},
		}}
	// embed
	vectors, err := iu.Embedder.EmbedText(ctx, content)
	if err != nil {
		return fmt.Errorf("error embedding: %w", err)
	}

	err = iu.VectorStore.UpsertDocuments(ctx, iu.config.QdrantCollection, docs, [][]float32{vectors})
	if err != nil {
		return fmt.Errorf("error upserting: %w", err)
	}

	iu.Logger.Info("document ingested", "filename", filename)
	return nil
}
