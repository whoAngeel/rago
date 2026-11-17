package application_test

import (
	"strings"
	"testing"

	"github.com/whoAngeel/rago/internal/application"
	"github.com/whoAngeel/rago/internal/core/ports"
)

func TestBuildUserMessage_RendersDocumentBlocks(t *testing.T) {
	got := application.BuildUserMessage(
		[]ports.SearchResult{rtResult("UNO", "a.pdf", 0.9), rtResult("DOS", "b.pdf", 0.8)},
		"", "¿qué hay?",
	)
	want := "<contexto>\n" +
		"<documento fuente=\"a.pdf\">\nUNO\n</documento>\n" +
		"<documento fuente=\"b.pdf\">\nDOS\n</documento>\n" +
		"</contexto>\n\nPregunta: ¿qué hay?"
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestBuildUserMessage_HistoryOnlyWhenPresent(t *testing.T) {
	res := []ports.SearchResult{rtResult("UNO", "a.pdf", 0.9)}
	if got := application.BuildUserMessage(res, "", "q"); strings.Contains(got, "<historial>") {
		t.Errorf("unexpected history block: %q", got)
	}
	got := application.BuildUserMessage(res, "user: hola", "q")
	want := "</contexto>\n\n<historial>\nuser: hola\n</historial>\n\nPregunta: q"
	if !strings.Contains(got, want) {
		t.Errorf("history block wrong: %q", got)
	}
}

func TestBuildUserMessage_NoResults(t *testing.T) {
	got := application.BuildUserMessage(nil, "", "q")
	want := "<contexto>\nNo se encontró información relevante en los documentos.\n</contexto>\n\nPregunta: q"
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestBuildUserMessage_EscapesQuotesInSource(t *testing.T) {
	got := application.BuildUserMessage([]ports.SearchResult{rtResult("X", `a"b.pdf`, 0.9)}, "", "q")
	if !strings.Contains(got, `<documento fuente="a&quot;b.pdf">`) {
		t.Errorf("source not escaped: %q", got)
	}
}

func TestBuildUserMessage_UnknownSource(t *testing.T) {
	r := rtResult("X", "", 0.9)
	r.Document.Metadata = nil
	if got := application.BuildUserMessage([]ports.SearchResult{r}, "", "q"); !strings.Contains(got, `fuente="desconocido"`) {
		t.Errorf("missing default source: %q", got)
	}
}
