package eval

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/tmc/langchaingo/schema"
	"github.com/whoAngeel/rago/internal/core/ports"
)

type fakeEmbedder struct{ order map[string]float32 }

func (f fakeEmbedder) ComputeEmbeddings(context.Context, []string) ([][]float32, error) {
	return nil, nil
}
func (f fakeEmbedder) EmbedText(_ context.Context, text string) ([]float32, error) {
	return []float32{f.order[text]}, nil
}

type fakeStore struct {
	ports.VectorStore
	byKey map[float32][]ports.SearchResult
	limit int
	user  int
}

func (f *fakeStore) Search(_ context.Context, _ string, v []float32, userID int, _ []int, limit int) ([]ports.SearchResult, error) {
	f.limit, f.user = limit, userID
	return f.byKey[v[0]], nil
}

func res(source string, score float32) ports.SearchResult {
	return ports.SearchResult{Document: schema.Document{Metadata: map[string]any{"source": source}}, Score: score}
}

func TestRunQuestions(t *testing.T) {
	emb := fakeEmbedder{order: map[string]float32{"q-a": 1, "q-u": 2}}
	store := &fakeStore{byKey: map[float32][]ports.SearchResult{
		1: {res("x.txt", 0.6), res("a.txt", 0.5), res("a.txt", 0.4)},
		2: {res("x.txt", 0.2)},
	}}
	qs := []Question{
		{ID: "a", Question: "q-a", ExpectedSource: "a.txt"},
		{ID: "u", Question: "q-u"},
	}
	got, err := RunQuestions(context.Background(), emb, store, "rago_eval", 42, qs, 5)
	if err != nil {
		t.Fatal(err)
	}
	if store.limit != 5 || store.user != 42 {
		t.Errorf("search called with limit=%d user=%d, want 5/42", store.limit, store.user)
	}
	if got[0].Rank != 2 || got[0].Score != 0.5 || got[0].TopScore != 0.6 || got[0].TopSource != "x.txt" {
		t.Errorf("answerable result = %+v", got[0])
	}
	if got[1].Rank != 0 || got[1].TopScore != 0.2 {
		t.Errorf("unanswerable result = %+v", got[1])
	}
}

type fakeParser struct{}

func (fakeParser) Parse(context.Context, io.Reader, string) ([]schema.Document, error) {
	return []schema.Document{
		{PageContent: "row", Metadata: map[string]any{"chunk_type": "structured"}},
		{PageContent: "a b"},
	}, nil
}

type fakeRegistry struct{}

func (fakeRegistry) Resolve(string) (ports.Parser, error) { return fakeParser{}, nil }

type fakeChunker struct{}

func (fakeChunker) Chunk(text string) ([]string, error) { return strings.Split(text, " "), nil }

func TestBuildChunksSkipsChunkerForStructured(t *testing.T) {
	chunks, err := BuildChunks(context.Background(), fakeRegistry{}, fakeChunker{}, "x.txt", []byte("{}"))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"row", "a", "b"}
	if strings.Join(chunks, "|") != strings.Join(want, "|") {
		t.Errorf("chunks = %v, want %v", chunks, want)
	}
}

func TestReportOutput(t *testing.T) {
	results := []QuestionResult{
		{ID: "a", ExpectedSource: "a.txt", Rank: 1, Score: 0.5, TopScore: 0.5, TopSource: "a.txt"},
		{ID: "m", ExpectedSource: "b.txt", TopScore: 0.3, TopSource: "x.txt"},
		{ID: "u", TopScore: 0.2, TopSource: "x.txt"},
	}
	rep := BuildReport(results, 8, 0.30, []float64{0.2, 0.3})
	var buf bytes.Buffer
	if err := rep.WriteTable(&buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{"hit@1", "MRR", "0.30", "(none)", "miss"} {
		if !strings.Contains(out, want) {
			t.Errorf("table missing %q:\n%s", want, out)
		}
	}
	buf.Reset()
	if err := rep.WriteJSON(&buf); err != nil || !strings.Contains(buf.String(), `"mrr"`) {
		t.Errorf("json output = %q, err = %v", buf.String(), err)
	}
	// configured 0.30 duplicates a sweep value, so it appears once and is flagged.
	if len(rep.Thresholds) != 2 || !rep.Thresholds[0].Configured && !rep.Thresholds[1].Configured {
		t.Errorf("thresholds = %+v, want 2 rows with one configured", rep.Thresholds)
	}
}
