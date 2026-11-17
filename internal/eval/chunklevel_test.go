package eval

import (
	"context"
	"math"
	"testing"

	"github.com/tmc/langchaingo/schema"
	"github.com/whoAngeel/rago/internal/core/ports"
)

func TestFirstChunkRankNeedsSourceAndEveryString(t *testing.T) {
	h := []Hit{
		{Source: "a.txt", Text: "just x here", Score: 0.9},
		{Source: "b.txt", Text: "x and y", Score: 0.8},
		{Source: "a.txt", Text: "X  and\nY too", Score: 0.7},
	}
	rank, score := FirstChunkRank(h, "a.txt", []string{"x", "y"})
	if rank != 3 || score != 0.7 {
		t.Fatalf("chunk rank/score = %d/%v, want 3/0.7", rank, score)
	}
	// The file-level rank is untouched and better than the chunk-level one.
	if fr, _ := FirstRank(h, "a.txt"); fr != 1 {
		t.Errorf("file rank = %d, want 1", fr)
	}
	if rank, _ := FirstChunkRank(h, "b.txt", []string{"x", "zzz"}); rank != 0 {
		t.Errorf("missing string rank = %d, want 0", rank)
	}
	if rank, _ := FirstChunkRank(h, "c.txt", []string{"x"}); rank != 0 {
		t.Errorf("wrong source rank = %d, want 0", rank)
	}
}

func TestSummarizeBothLevels(t *testing.T) {
	results := []QuestionResult{
		{ID: "a", ExpectedSource: "a", Rank: 1, HasChunkCheck: true, ChunkRank: 3},
		{ID: "b", ExpectedSource: "b", Rank: 2},
		{ID: "c", ExpectedSource: "c", Rank: 0, HasChunkCheck: true},
		{ID: "d", ExpectedSource: "d", Rank: 1, HasChunkCheck: true, ChunkRank: 1},
		{ID: "u"},
	}
	s := Summarize(results)
	if s.Hit1 != 2 || s.Hit3 != 3 || s.HitK != 3 || math.Abs(s.MRR-0.625) > 1e-9 {
		t.Errorf("file level = %d/%d/%d mrr %v, want 2/3/3 0.625", s.Hit1, s.Hit3, s.HitK, s.MRR)
	}
	if s.ChunkCovered != 3 {
		t.Fatalf("chunk covered = %d, want 3", s.ChunkCovered)
	}
	if s.ChunkHit1 != 1 || s.ChunkHit3 != 2 || s.ChunkHitK != 2 {
		t.Errorf("chunk level = %d/%d/%d, want 1/2/2", s.ChunkHit1, s.ChunkHit3, s.ChunkHitK)
	}
	if want := (1.0/3 + 0 + 1) / 3; math.Abs(s.ChunkMRR-want) > 1e-9 {
		t.Errorf("chunk MRR = %v, want %v", s.ChunkMRR, want)
	}
}

func TestSweepUsesChunkLevelWhenChecked(t *testing.T) {
	results := []QuestionResult{
		{ID: "A", ExpectedSource: "a", Rank: 1, Score: 0.9, TopScore: 0.9, HasChunkCheck: true, ChunkRank: 2, ChunkScore: 0.30},
		{ID: "B", ExpectedSource: "b", Rank: 1, Score: 0.5, TopScore: 0.5},
		{ID: "C", ExpectedSource: "c", Rank: 1, Score: 0.8, TopScore: 0.8, HasChunkCheck: true},
	}
	rows := SweepThresholds(results, []float64{0, 0.35})
	if rows[0].AnswerableKept != 2 || rows[1].AnswerableKept != 1 {
		t.Errorf("kept = %d/%d, want 2/1", rows[0].AnswerableKept, rows[1].AnswerableKept)
	}
}

func TestPerFileTotals(t *testing.T) {
	results := []QuestionResult{
		{ID: "a", File: "x.json", ExpectedSource: "a", Rank: 1},
		{ID: "b", File: "y.json", ExpectedSource: "b", Rank: 0},
		{ID: "c", File: "x.json", ExpectedSource: "c", Rank: 2},
	}
	rep := BuildReport(results, 5, 0.3, nil)
	if len(rep.Files) != 2 || rep.Files[0].File != "x.json" || rep.Files[1].File != "y.json" {
		t.Fatalf("files = %+v", rep.Files)
	}
	if rep.Files[0].Summary.Answerable != 2 || rep.Files[0].Summary.HitK != 2 || rep.Files[1].Summary.HitK != 0 {
		t.Errorf("per-file summaries = %+v", rep.Files)
	}
	if rep.Summary.Answerable != 3 || rep.Summary.HitK != 2 {
		t.Errorf("overall = %+v", rep.Summary)
	}
}

type contentStore struct {
	ports.VectorStore
	res []ports.SearchResult
}

func (c contentStore) Search(context.Context, string, []float32, int, []int, int) ([]ports.SearchResult, error) {
	return c.res, nil
}

func TestRunQuestionsChunkLevel(t *testing.T) {
	mk := func(src, text string, score float32) ports.SearchResult {
		return ports.SearchResult{Document: schema.Document{PageContent: text, Metadata: map[string]any{"source": src}}, Score: score}
	}
	store := contentStore{res: []ports.SearchResult{
		mk("a.txt", "intro text", 0.6),
		mk("a.txt", "the ﬁnal  Code X-9", 0.5),
	}}
	qs := []Question{
		{ID: "q", File: "f.json", Question: "q", ExpectedSource: "a.txt", ExpectedContains: []string{"final code", "x-9"}},
		{ID: "p", File: "f.json", Question: "p", ExpectedSource: "a.txt"},
	}
	got, err := RunQuestions(context.Background(), fakeEmbedder{}, store, "c", 1, qs, 5)
	if err != nil {
		t.Fatal(err)
	}
	r := got[0]
	if r.Rank != 1 || !r.HasChunkCheck || r.ChunkRank != 2 || r.ChunkScore != 0.5 || r.File != "f.json" {
		t.Errorf("result = %+v", r)
	}
	if got[1].HasChunkCheck || got[1].ChunkRank != 0 {
		t.Errorf("question without expected_contains = %+v", got[1])
	}
}
