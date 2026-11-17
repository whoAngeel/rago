package chunker

import (
	"strings"
	"testing"

	"github.com/tmc/langchaingo/schema"
)

type splitChunker struct{}

func (splitChunker) Chunk(text string) ([]string, error) { return strings.Split(text, " "), nil }

func TestChunkPages(t *testing.T) {
	docs := []schema.Document{
		{PageContent: "Title\nrow", Metadata: map[string]any{"chunk_type": "structured"}},
		{PageContent: "a b", Metadata: map[string]any{"page_number": 2}},
	}
	got, err := ChunkPages(docs, "x.txt", splitChunker{}, false)
	if err != nil {
		t.Fatal(err)
	}
	if want := "Title\nrow|a|b"; strings.Join(got, "|") != want {
		t.Errorf("without header: %q, want %q", strings.Join(got, "|"), want)
	}
	got, err = ChunkPages(docs, "x.txt", splitChunker{}, true)
	if err != nil {
		t.Fatal(err)
	}
	want := "[Title]\nTitle\nrow|[Title — p. 2]\na|[Title — p. 2]\nb"
	if strings.Join(got, "|") != want {
		t.Errorf("with header: %q, want %q", strings.Join(got, "|"), want)
	}
}
