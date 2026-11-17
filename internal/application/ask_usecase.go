package application

import (
	"context"
	"fmt"

	"github.com/whoAngeel/rago/internal/core/domain"
	"github.com/whoAngeel/rago/internal/core/ports"
	"github.com/whoAngeel/rago/internal/infrastructure/config"
)

type AskUsecase struct {
	VectorStore ports.VectorStore
	LLMProvider ports.LLMProvider
	Embedder    ports.Embedder
	Logger      ports.Logger
	UsageRepo   ports.LLMUsageRepository
	Config      *config.Config
}

func NewAskUsecase(
	vStore ports.VectorStore,
	llm ports.LLMProvider,
	log ports.Logger,
	embedder ports.Embedder,
	cfg *config.Config,
	usageRepo ports.LLMUsageRepository,
) *AskUsecase {
	return &AskUsecase{
		VectorStore: vStore,
		LLMProvider: llm,
		Embedder:    embedder,
		Logger:      log,
		UsageRepo:   usageRepo,
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

	results, err := au.VectorStore.Search(ctx, au.Config.QdrantCollection, queryVector, userID, nil, au.Config.RetrievalTopK)
	if err != nil {
		return "", fmt.Errorf("error searching: %w", err)
	}
	results = filterByScore(au.Logger, results, au.Config.RetrievalScoreThreshold)

	if len(results) == 0 {
		return "[NOT FOUND RELEVANT INFO]", nil
	}

	prompt := BuildUserMessage(results, "", question)
	au.Logger.Info("Generating answer", "sources_found", len(results), "prompt_len", len(prompt))

	llmResult, err := au.LLMProvider.GenerateAnswer(ctx, DefaultSystemPrompt, prompt)
	if err != nil {
		au.Logger.Error("Error generating answer", "error", err)
		return "", fmt.Errorf("error generating answer: %w", err)
	}
	if au.UsageRepo != nil {
		_ = au.UsageRepo.Create(ctx, &domain.LLMUsage{
			UserID:       userID,
			Model:        llmResult.Model,
			InputTokens:  llmResult.InputTokens,
			OutputTokens: llmResult.OutputTokens,
		})
	}
	return llmResult.Answer, nil
}
