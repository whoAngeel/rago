package application

import (
	"context"
	"fmt"

	"github.com/whoAngeel/rago/internal/core/ports"
	"github.com/whoAngeel/rago/internal/infrastructure/config"
)

const defaultLimit = 3

type AskUsecase struct {
	VectorStore ports.VectorStore
	LLMProvider ports.LLMProvider
	Embedder    ports.Embedder
	Logger      ports.Logger
	Config      *config.Config
}

func NewAskUsecase(
	vStore ports.VectorStore,
	llm ports.LLMProvider,
	log ports.Logger,
	embedder ports.Embedder,
	cfg *config.Config,
) *AskUsecase {
	return &AskUsecase{
		VectorStore: vStore,
		LLMProvider: llm,
		Embedder:    embedder,
		Logger:      log,
		Config:      cfg,
	}
}

func (au *AskUsecase) Execute(ctx context.Context, userID int, question string) (string, error) {
	// au.Logger.Info("Pregunta", "question", question)
	au.Logger.Info("Embedding question", "question", question)
	queryVector, err := au.Embedder.EmbedText(ctx, question)
	if err != nil {
		au.Logger.Error("error embedding question", "error", err)
		return "", fmt.Errorf("error embedding question: %w", err)
	}
	err = au.VectorStore.CreateCollection(ctx, au.Config.QdrantCollection, au.Config.EmbeddingDim)
	if err != nil {
		au.Logger.Warn("Collection may already exist", "error", err)
	}

	results, err := au.VectorStore.Search(ctx, au.Config.QdrantCollection, queryVector, userID, defaultLimit)
	if err != nil {
		return "", fmt.Errorf("error searching: %w", err)
	}

	if len(results) == 0 {
		return "[NOT FOUND RELEVANT INFO]", nil
	}

	var context string
	for _, r := range results {
		context += r.Document.PageContent + "\n"
	}
	prompt := fmt.Sprintf("Contexto: %s\n\nPregunta: %s\n\nResponde basándote solo en el contexto.", context, question)
	au.Logger.Info("Generating answer", "prompt", prompt)

	answer, err := au.LLMProvider.GenerateAnswer(ctx, prompt)
	if err != nil {
		au.Logger.Error("Error generating answer", "error", err)
		return "", fmt.Errorf("error generating answer: %w", err)
	}

	return answer, nil
}
