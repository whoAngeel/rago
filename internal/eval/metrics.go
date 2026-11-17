package eval

// Hit is one retrieved chunk reduced to what the evaluation needs.
type Hit struct {
	Source string
	Score  float32
}

// QuestionResult is the outcome of retrieving for one question. Rank is the
// 1-based position of the first chunk from the expected source, 0 for a miss
// (or for unanswerable questions).
type QuestionResult struct {
	ID             string  `json:"id"`
	ExpectedSource string  `json:"expected_source"`
	Rank           int     `json:"rank"`
	Score          float32 `json:"score"`
	TopScore       float32 `json:"top_score"`
	TopSource      string  `json:"top_source"`
	Note           string  `json:"note,omitempty"`
}

// FirstRank returns the rank and score of the first hit from expected, or
// (0, 0) when none matches.
func FirstRank(hits []Hit, expected string) (int, float32) {
	for i, h := range hits {
		if h.Source == expected {
			return i + 1, h.Score
		}
	}
	return 0, 0
}

// Summary aggregates metrics over the answerable questions.
type Summary struct {
	Answerable   int     `json:"answerable"`
	Unanswerable int     `json:"unanswerable"`
	Hit1         int     `json:"hit_at_1"`
	Hit3         int     `json:"hit_at_3"`
	HitK         int     `json:"hit_at_k"`
	MRR          float64 `json:"mrr"`
}

// Summarize computes hit@1, hit@3, hit@k and MRR. Every result's rank was
// obtained with the same top-k, so hit@k is any non-miss.
func Summarize(results []QuestionResult) Summary {
	var s Summary
	var rr float64
	for _, r := range results {
		if r.ExpectedSource == "" {
			s.Unanswerable++
			continue
		}
		s.Answerable++
		if r.Rank == 0 {
			continue
		}
		s.HitK++
		if r.Rank <= 1 {
			s.Hit1++
		}
		if r.Rank <= 3 {
			s.Hit3++
		}
		rr += 1 / float64(r.Rank)
	}
	if s.Answerable > 0 {
		s.MRR = rr / float64(s.Answerable)
	}
	return s
}

// ThresholdRow is the effect of one score cutoff. A chunk is kept when its
// score is >= the threshold, matching the application's filter; a threshold
// <= 0 keeps everything.
type ThresholdRow struct {
	Threshold         float64 `json:"threshold"`
	Configured        bool    `json:"configured,omitempty"`
	AnswerableKept    int     `json:"answerable_kept"`
	AnswerableTotal   int     `json:"answerable_total"`
	UnanswerableClean int     `json:"unanswerable_clean"`
	UnanswerableTotal int     `json:"unanswerable_total"`
}

// SweepThresholds counts, per cutoff, answerable questions whose expected
// chunk survives and unanswerable questions that end with zero results.
func SweepThresholds(results []QuestionResult, thresholds []float64) []ThresholdRow {
	rows := make([]ThresholdRow, 0, len(thresholds))
	for _, th := range thresholds {
		row := ThresholdRow{Threshold: th}
		for _, r := range results {
			if r.ExpectedSource == "" {
				row.UnanswerableTotal++
				if noResultsAbove(r, th) {
					row.UnanswerableClean++
				}
				continue
			}
			row.AnswerableTotal++
			if r.Rank > 0 && (th <= 0 || float64(r.Score) >= th) {
				row.AnswerableKept++
			}
		}
		rows = append(rows, row)
	}
	return rows
}

// noResultsAbove reports whether nothing retrieved for r survives the cutoff.
// Hits are score-ordered, so the top score decides; with no cutoff only a
// search that returned nothing at all is clean.
func noResultsAbove(r QuestionResult, th float64) bool {
	if th <= 0 {
		return r.TopSource == "" && r.TopScore == 0
	}
	return float64(r.TopScore) < th
}
