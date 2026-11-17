package application_test

import (
	"context"
	"strings"
	"testing"

	"github.com/whoAngeel/rago/internal/application"
	"github.com/whoAngeel/rago/internal/core/ports"
	"github.com/whoAngeel/rago/internal/infrastructure/config"
)

func newRtAsk(vs *rtVectorStore, llm *rtLLM, topK int, threshold float64) *application.AskUsecase {
	cfg := &config.Config{QdrantCollection: "col", EmbeddingDim: 1, RetrievalTopK: topK, RetrievalScoreThreshold: threshold}
	return application.NewAskUsecase(vs, llm, rtLogger{}, rtEmbedder{}, cfg, nil)
}

func TestAsk_UsesTopKAndDropsBelowThreshold(t *testing.T) {
	vs := &rtVectorStore{results: []ports.SearchResult{
		rtResult("GOOD-CHUNK", "a.pdf", 0.8),
		rtResult("BAD-CHUNK", "b.pdf", 0.1),
	}}
	llm := &rtLLM{}
	if _, err := newRtAsk(vs, llm, 4, 0.3).Execute(context.Background(), 1, "q"); err != nil {
		t.Fatal(err)
	}
	if vs.lastLimit != 4 {
		t.Errorf("limit = %d; want 4", vs.lastLimit)
	}
	if !strings.Contains(llm.prompt, "GOOD-CHUNK") || strings.Contains(llm.prompt, "BAD-CHUNK") {
		t.Errorf("prompt wrong: %q", llm.prompt)
	}
}

func TestAsk_AllBelowThresholdReturnsNotFound(t *testing.T) {
	vs := &rtVectorStore{results: []ports.SearchResult{rtResult("BAD-CHUNK", "b.pdf", 0.1)}}
	llm := &rtLLM{}
	got, err := newRtAsk(vs, llm, 4, 0.3).Execute(context.Background(), 1, "q")
	if err != nil {
		t.Fatal(err)
	}
	if got != "[NOT FOUND RELEVANT INFO]" {
		t.Errorf("got %q; want not-found marker", got)
	}
	if llm.prompt != "" {
		t.Errorf("LLM should not be called, prompt=%q", llm.prompt)
	}
}

func TestAsk_SystemPromptSeparateFromUserMessage(t *testing.T) {
	vs := &rtVectorStore{results: []ports.SearchResult{rtResult("CHUNK", "a.pdf", 0.8)}}
	llm := &rtLLM{}
	if _, err := newRtAsk(vs, llm, 4, 0).Execute(context.Background(), 1, "mi pregunta"); err != nil {
		t.Fatal(err)
	}
	if llm.system != application.DefaultSystemPrompt {
		t.Errorf("system = %q; want DefaultSystemPrompt", llm.system)
	}
	if strings.Contains(llm.prompt, "Reglas:") {
		t.Errorf("user message repeats the system prompt: %q", llm.prompt)
	}
	if !strings.Contains(llm.prompt, `<documento fuente="a.pdf">`) || !strings.Contains(llm.prompt, "Pregunta: mi pregunta") {
		t.Errorf("user message wrong: %q", llm.prompt)
	}
}
