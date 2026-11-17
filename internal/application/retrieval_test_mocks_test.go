package application_test

import (
	"context"

	"github.com/tmc/langchaingo/schema"
	"github.com/whoAngeel/rago/internal/core/domain"
	"github.com/whoAngeel/rago/internal/core/ports"
)

// Mocks local to the retrieval tests (prefixed rt to avoid clashing with other test files).

type rtVectorStore struct {
	results   []ports.SearchResult
	lastLimit int
}

func (m *rtVectorStore) Search(_ context.Context, _ string, _ []float32, _ int, _ []int, limit int) ([]ports.SearchResult, error) {
	m.lastLimit = limit
	return m.results, nil
}
func (m *rtVectorStore) CreateCollection(_ context.Context, _ string, _ int) error { return nil }
func (m *rtVectorStore) UpsertDocuments(_ context.Context, _ string, _ []schema.Document, _ [][]float32) error {
	return nil
}
func (m *rtVectorStore) GetPointsCount(_ context.Context, _ string) (uint64, error) { return 0, nil }
func (m *rtVectorStore) GetPointsCountByFilter(_ context.Context, _ string, _ int) (uint64, error) {
	return 0, nil
}
func (m *rtVectorStore) DeleteCollection(_ context.Context, _ string) error { return nil }
func (m *rtVectorStore) DeleteByDocumentID(_ context.Context, _ string, _ int) error {
	return nil
}

type rtEmbedder struct{}

func (rtEmbedder) EmbedText(_ context.Context, _ string) ([]float32, error) {
	return []float32{0.1}, nil
}
func (rtEmbedder) ComputeEmbeddings(_ context.Context, _ []string) ([][]float32, error) {
	return [][]float32{{0.1}}, nil
}

type rtLLM struct{ system, prompt string }

func (m *rtLLM) GenerateAnswer(_ context.Context, sys, p string) (ports.LLMResult, error) {
	m.system, m.prompt = sys, p
	return ports.LLMResult{Answer: "ok"}, nil
}
func (m *rtLLM) Stream(_ context.Context, sys, p string, _ func(string) error) (ports.LLMResult, error) {
	m.system, m.prompt = sys, p
	return ports.LLMResult{Answer: "ok"}, nil
}

// rtConfigRepo returns "SYS" as the stored system prompt, or err when set.
type rtConfigRepo struct{ err error }

func (r rtConfigRepo) Get(_ context.Context, _ string) (string, error) {
	if r.err != nil {
		return "", r.err
	}
	return "SYS", nil
}
func (rtConfigRepo) Set(_ context.Context, _, _ string) error { return nil }

type rtLogger struct{}

func (l rtLogger) Info(_ string, _ ...any)    {}
func (l rtLogger) Warn(_ string, _ ...any)    {}
func (l rtLogger) Debug(_ string, _ ...any)   {}
func (l rtLogger) Error(_ string, _ ...any)   {}
func (l rtLogger) Fatal(_ string, _ ...any)   {}
func (l rtLogger) With(_ ...any) ports.Logger { return l }

type rtChatRepo struct{}

func (rtChatRepo) CreateSession(_ context.Context, s *domain.ChatSession) error {
	s.ID = 1
	return nil
}
func (rtChatRepo) GetSession(_ context.Context, _, _ int) (*domain.ChatSession, error) {
	return &domain.ChatSession{}, nil
}
func (rtChatRepo) GetUserSessions(_ context.Context, _ int) ([]*domain.ChatSession, error) {
	return nil, nil
}
func (rtChatRepo) UpdateSessionTitle(_ context.Context, _, _ int, _ string) error { return nil }
func (rtChatRepo) DeleteSession(_ context.Context, _, _ int) error                { return nil }
func (rtChatRepo) CreateMessage(_ context.Context, _ *domain.ChatMessage) error   { return nil }
func (rtChatRepo) GetMessages(_ context.Context, _, _ int) ([]*domain.ChatMessage, error) {
	return nil, nil
}
func (rtChatRepo) GetAllMessages(_ context.Context, _ int) ([]*domain.ChatMessage, error) {
	return nil, nil
}

func rtResult(content, source string, score float32) ports.SearchResult {
	return ports.SearchResult{
		Document: schema.Document{PageContent: content, Metadata: map[string]any{"source": source}},
		Score:    score,
	}
}
