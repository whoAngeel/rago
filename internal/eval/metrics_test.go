package eval

import (
	"math"
	"testing"
)

func hits(sources ...string) []Hit {
	out := make([]Hit, len(sources))
	for i, s := range sources {
		out[i] = Hit{Source: s, Score: float32(0.9) - float32(i)*0.1}
	}
	return out
}

func TestFirstRank(t *testing.T) {
	h := hits("x.txt", "a.txt", "a.txt")
	rank, score := FirstRank(h, "a.txt")
	if rank != 2 {
		t.Fatalf("rank = %d, want 2", rank)
	}
	if score != h[1].Score {
		t.Errorf("score = %v, want %v", score, h[1].Score)
	}
	if rank, _ := FirstRank(h, "missing.txt"); rank != 0 {
		t.Errorf("miss rank = %d, want 0", rank)
	}
	if rank, _ := FirstRank(nil, "a.txt"); rank != 0 {
		t.Errorf("empty hits rank = %d, want 0", rank)
	}
}

func TestSummarize(t *testing.T) {
	results := []QuestionResult{
		{ID: "a", ExpectedSource: "a.txt", Rank: 1},
		{ID: "b", ExpectedSource: "b.txt", Rank: 3},
		{ID: "c", ExpectedSource: "c.txt", Rank: 0},
		{ID: "d", ExpectedSource: "d.txt", Rank: 2},
		{ID: "u", ExpectedSource: "", Rank: 0},
	}
	s := Summarize(results)
	if s.Answerable != 4 || s.Unanswerable != 1 {
		t.Fatalf("counts = %d/%d, want 4/1", s.Answerable, s.Unanswerable)
	}
	if s.Hit1 != 1 || s.Hit3 != 3 || s.HitK != 3 {
		t.Errorf("hits = %d/%d/%d, want 1/3/3", s.Hit1, s.Hit3, s.HitK)
	}
	wantMRR := (1.0 + 1.0/3 + 0 + 0.5) / 4
	if math.Abs(s.MRR-wantMRR) > 1e-9 {
		t.Errorf("MRR = %v, want %v", s.MRR, wantMRR)
	}
}

func TestSummarizeNoAnswerable(t *testing.T) {
	s := Summarize([]QuestionResult{{ID: "u"}})
	if s.MRR != 0 || s.Answerable != 0 {
		t.Errorf("summary = %+v, want zeroes", s)
	}
}

func TestSweepThresholds(t *testing.T) {
	results := []QuestionResult{
		{ID: "A", ExpectedSource: "a", Rank: 1, Score: 0.42, TopScore: 0.42},
		{ID: "B", ExpectedSource: "b", Rank: 2, Score: 0.28, TopScore: 0.50},
		{ID: "C", ExpectedSource: "c", Rank: 0, TopScore: 0.45},
		{ID: "U1", TopScore: 0.22},
		{ID: "U2", TopScore: 0.35},
		{ID: "U3"}, // nothing returned at all
	}
	rows := SweepThresholds(results, []float64{0, 0.20, 0.30, 0.40})
	want := []ThresholdRow{
		{Threshold: 0, AnswerableKept: 2, AnswerableTotal: 3, UnanswerableClean: 1, UnanswerableTotal: 3},
		{Threshold: 0.20, AnswerableKept: 2, AnswerableTotal: 3, UnanswerableClean: 1, UnanswerableTotal: 3},
		{Threshold: 0.30, AnswerableKept: 1, AnswerableTotal: 3, UnanswerableClean: 2, UnanswerableTotal: 3},
		{Threshold: 0.40, AnswerableKept: 1, AnswerableTotal: 3, UnanswerableClean: 3, UnanswerableTotal: 3},
	}
	for i, w := range want {
		if rows[i] != w {
			t.Errorf("row %d = %+v, want %+v", i, rows[i], w)
		}
	}
}
