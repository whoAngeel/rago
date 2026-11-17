package application_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/whoAngeel/rago/internal/application"
	"github.com/whoAngeel/rago/internal/core/ports"
)

func newRtChat(vs *rtVectorStore, llm *rtLLM, topK int, threshold float64) *application.ChatUsecase {
	return application.NewChatUsecase(
		rtChatRepo{}, rtConfigRepo{}, vs, rtEmbedder{}, llm, nil, rtLogger{},
		10, "col", topK, threshold, nil,
	)
}

func TestChat_SearchUsesTopK(t *testing.T) {
	vs := &rtVectorStore{}
	uc := newRtChat(vs, &rtLLM{}, 7, 0)
	if _, _, _, err := uc.SendMessage(context.Background(), 1, nil, "q"); err != nil {
		t.Fatal(err)
	}
	if vs.lastLimit != 7 {
		t.Errorf("SendMessage limit = %d; want 7", vs.lastLimit)
	}
	vs.lastLimit = 0
	if _, _, _, err := uc.SendStream(context.Background(), 1, nil, "q", func(string) error { return nil }); err != nil {
		t.Fatal(err)
	}
	if vs.lastLimit != 7 {
		t.Errorf("SendStream limit = %d; want 7", vs.lastLimit)
	}
}

func TestChat_DropsBelowThreshold(t *testing.T) {
	results := func() *rtVectorStore {
		return &rtVectorStore{results: []ports.SearchResult{
			rtResult("GOOD-CHUNK", "a.pdf", 0.8),
			rtResult("BAD-CHUNK", "b.pdf", 0.1),
		}}
	}
	t.Run("SendMessage", func(t *testing.T) {
		llm := &rtLLM{}
		_, sources, _, err := newRtChat(results(), llm, 8, 0.3).SendMessage(context.Background(), 1, nil, "q")
		if err != nil {
			t.Fatal(err)
		}
		if len(sources) != 1 || sources[0].Content != "GOOD-CHUNK" {
			t.Errorf("sources = %+v; want only GOOD-CHUNK", sources)
		}
		if !strings.Contains(llm.prompt, "GOOD-CHUNK") || strings.Contains(llm.prompt, "BAD-CHUNK") {
			t.Errorf("prompt wrong: %q", llm.prompt)
		}
	})
	t.Run("SendStream", func(t *testing.T) {
		llm := &rtLLM{}
		_, sources, _, err := newRtChat(results(), llm, 8, 0.3).SendStream(context.Background(), 1, nil, "q", func(string) error { return nil })
		if err != nil {
			t.Fatal(err)
		}
		if len(sources) != 1 || sources[0].Content != "GOOD-CHUNK" {
			t.Errorf("sources = %+v; want only GOOD-CHUNK", sources)
		}
		if !strings.Contains(llm.prompt, "GOOD-CHUNK") || strings.Contains(llm.prompt, "BAD-CHUNK") {
			t.Errorf("prompt wrong: %q", llm.prompt)
		}
	})
}

func TestChat_AllBelowThresholdTakesNoInfoBranch(t *testing.T) {
	vs := &rtVectorStore{results: []ports.SearchResult{rtResult("BAD-CHUNK", "b.pdf", 0.1)}}
	llm := &rtLLM{}
	_, sources, _, err := newRtChat(vs, llm, 8, 0.3).SendMessage(context.Background(), 1, nil, "q")
	if err != nil {
		t.Fatal(err)
	}
	if len(sources) != 0 {
		t.Errorf("sources = %+v; want none", sources)
	}
	if !strings.Contains(llm.prompt, "No se encontró información relevante en los documentos.") {
		t.Errorf("prompt missing no-info branch: %q", llm.prompt)
	}
	if strings.Contains(llm.prompt, "<documento") || strings.Contains(llm.prompt, "BAD-CHUNK") {
		t.Errorf("prompt should have no context: %q", llm.prompt)
	}
}

func TestChat_ZeroThresholdKeepsAll(t *testing.T) {
	vs := &rtVectorStore{results: []ports.SearchResult{
		rtResult("GOOD-CHUNK", "a.pdf", 0.8),
		rtResult("BAD-CHUNK", "b.pdf", 0.01),
	}}
	_, sources, _, err := newRtChat(vs, &rtLLM{}, 8, 0).SendMessage(context.Background(), 1, nil, "q")
	if err != nil {
		t.Fatal(err)
	}
	if len(sources) != 2 {
		t.Errorf("len(sources) = %d; want 2", len(sources))
	}
}

func TestChat_SystemPromptSeparateFromUserMessage(t *testing.T) {
	vs := &rtVectorStore{results: []ports.SearchResult{rtResult("CHUNK", "a.pdf", 0.8)}}
	t.Run("SendMessage", func(t *testing.T) {
		llm := &rtLLM{}
		if _, _, _, err := newRtChat(vs, llm, 8, 0).SendMessage(context.Background(), 1, nil, "mi pregunta"); err != nil {
			t.Fatal(err)
		}
		checkSplit(t, llm)
	})
	t.Run("SendStream", func(t *testing.T) {
		llm := &rtLLM{}
		if _, _, _, err := newRtChat(vs, llm, 8, 0).SendStream(context.Background(), 1, nil, "mi pregunta", func(string) error { return nil }); err != nil {
			t.Fatal(err)
		}
		checkSplit(t, llm)
	})
}

func checkSplit(t *testing.T, llm *rtLLM) {
	t.Helper()
	if llm.system != "SYS" {
		t.Errorf("system = %q; want the stored prompt SYS", llm.system)
	}
	if strings.Contains(llm.prompt, "SYS") {
		t.Errorf("user message repeats the system prompt: %q", llm.prompt)
	}
	if !strings.Contains(llm.prompt, `<documento fuente="a.pdf">`) || !strings.Contains(llm.prompt, "Pregunta: mi pregunta") {
		t.Errorf("user message wrong: %q", llm.prompt)
	}
}

func TestChat_FallsBackToDefaultPromptWhenConfigFails(t *testing.T) {
	newUC := func(llm *rtLLM) *application.ChatUsecase {
		return application.NewChatUsecase(
			rtChatRepo{}, rtConfigRepo{err: errors.New("db down")}, &rtVectorStore{}, rtEmbedder{}, llm, nil, rtLogger{},
			10, "col", 8, 0, nil,
		)
	}
	llm := &rtLLM{}
	if _, _, _, err := newUC(llm).SendMessage(context.Background(), 1, nil, "q"); err != nil {
		t.Fatal(err)
	}
	if llm.system != application.DefaultSystemPrompt {
		t.Errorf("SendMessage system = %q; want DefaultSystemPrompt", llm.system)
	}
	llm = &rtLLM{}
	if _, _, _, err := newUC(llm).SendStream(context.Background(), 1, nil, "q", func(string) error { return nil }); err != nil {
		t.Fatal(err)
	}
	if llm.system != application.DefaultSystemPrompt {
		t.Errorf("SendStream system = %q; want DefaultSystemPrompt", llm.system)
	}
}
