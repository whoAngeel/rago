package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/whoAngeel/rago/internal/core/ports"
	"github.com/whoAngeel/rago/internal/infrastructure/config"
)

const defaultLimit = 8

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

	var parts []string
	for i, r := range results {
		source := "desconocido"
		if s, ok := r.Document.Metadata["source"]; ok {
			source = fmt.Sprintf("%v", s)
		}
		parts = append(parts, fmt.Sprintf("[Fuente %d: %s]\n%s", i+1, source, r.Document.PageContent))
	}
	context := strings.Join(parts, "\n\n---\n\n")
	prompt := fmt.Sprintf("Usa SOLO el siguiente contexto para responder. Si no encuentras la respuesta, di 'No encontré información sobre eso en los documentos'.\n\nCONTEXTO:\n%s\n\nPREGUNTA: %s\n\nRESPUESTA:", context, question)
	au.Logger.Info("Generating answer", "sources_found", len(results), "prompt_len", len(prompt))

	answer, err := au.LLMProvider.GenerateAnswer(ctx, prompt)
	if err != nil {
		au.Logger.Error("Error generating answer", "error", err)
		return "", fmt.Errorf("error generating answer: %w", err)
	}

	return answer, nil
}
