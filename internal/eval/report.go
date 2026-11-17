package eval

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"text/tabwriter"
)

// DefaultSweep is the threshold sweep shown next to the configured value.
var DefaultSweep = []float64{0.20, 0.25, 0.30, 0.35, 0.40, 0.45}

// Report is the full evaluation output, printed as a table and as JSON.
type Report struct {
	TopK                int              `json:"top_k"`
	ConfiguredThreshold float64          `json:"configured_threshold"`
	Questions           []QuestionResult `json:"questions"`
	Summary             Summary          `json:"summary"`
	Files               []FileSummary    `json:"files,omitempty"`
	Thresholds          []ThresholdRow   `json:"thresholds"`
}

// FileSummary holds the totals of the questions loaded from one question file.
type FileSummary struct {
	File    string  `json:"file"`
	Summary Summary `json:"summary"`
}

func summarizeByFile(results []QuestionResult) []FileSummary {
	var order []string
	groups := map[string][]QuestionResult{}
	for _, r := range results {
		if r.File == "" {
			continue
		}
		if _, ok := groups[r.File]; !ok {
			order = append(order, r.File)
		}
		groups[r.File] = append(groups[r.File], r)
	}
	out := make([]FileSummary, 0, len(order))
	for _, f := range order {
		out = append(out, FileSummary{File: f, Summary: Summarize(groups[f])})
	}
	return out
}

// BuildReport computes the summary and the threshold view. The configured
// threshold is listed first and flagged; a sweep value equal to it is not
// repeated.
func BuildReport(results []QuestionResult, topK int, configured float64, sweep []float64) Report {
	ths := []float64{configured}
	for _, t := range sweep {
		if math.Abs(t-configured) > 1e-9 {
			ths = append(ths, t)
		}
	}
	rows := SweepThresholds(results, ths)
	rows[0].Configured = true
	return Report{
		TopK:                topK,
		ConfiguredThreshold: configured,
		Questions:           results,
		Summary:             Summarize(results),
		Files:               summarizeByFile(results),
		Thresholds:          rows,
	}
}

// WriteJSON writes the report as indented JSON.
func (r Report) WriteJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(r)
}

// WriteTable prints the report as plain tables.
func (r Report) WriteTable(w io.Writer) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tFILE\tEXPECTED SOURCE\tRANK\tSCORE\tCHUNK RANK\tCHUNK SCORE\tTOP SCORE\tTOP SOURCE")
	for _, q := range r.Questions {
		expected, rank, score := q.ExpectedSource, "miss", fmt.Sprintf("%.3f", q.Score)
		chunkRank, chunkScore := "-", "-"
		switch {
		case expected == "":
			expected, rank, score = "(none)", "-", "-"
		case q.Rank > 0:
			rank = fmt.Sprint(q.Rank)
		default:
			score = "-"
		}
		if q.HasChunkCheck {
			chunkRank = "miss"
			if q.ChunkRank > 0 {
				chunkRank, chunkScore = fmt.Sprint(q.ChunkRank), fmt.Sprintf("%.3f", q.ChunkScore)
			}
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%.3f\t%s\n", q.ID, q.File, expected, rank, score, chunkRank, chunkScore, q.TopScore, q.TopSource)
	}
	if err := tw.Flush(); err != nil {
		return err
	}

	fmt.Fprintln(w)
	r.writeTotals(w, "Overall", r.Summary)
	for _, f := range r.Files {
		r.writeTotals(w, f.File, f.Summary)
	}
	fmt.Fprintln(w)

	tw = tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "THRESHOLD\tANSWERABLE KEPT\tUNANSWERABLE EMPTY")
	for _, t := range r.Thresholds {
		label := fmt.Sprintf("%.2f", t.Threshold)
		if t.Configured {
			label += " (configured)"
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\n", label, ratio(t.AnswerableKept, t.AnswerableTotal), ratio(t.UnanswerableClean, t.UnanswerableTotal))
	}
	return tw.Flush()
}

func (r Report) writeTotals(w io.Writer, label string, s Summary) {
	fmt.Fprintf(w, "[%s] answerable %d (unanswerable %d), top-k = %d\n", label, s.Answerable, s.Unanswerable, r.TopK)
	fmt.Fprintf(w, "  file  : hit@1 %s  hit@3 %s  hit@%d %s  MRR %.3f\n", ratio(s.Hit1, s.Answerable), ratio(s.Hit3, s.Answerable), r.TopK, ratio(s.HitK, s.Answerable), s.MRR)
	fmt.Fprintf(w, "  chunk : hit@1 %s  hit@3 %s  hit@%d %s  MRR %.3f  (covers %d of %d answerable)\n", ratio(s.ChunkHit1, s.ChunkCovered), ratio(s.ChunkHit3, s.ChunkCovered), r.TopK, ratio(s.ChunkHitK, s.ChunkCovered), s.ChunkMRR, s.ChunkCovered, s.Answerable)
}

func ratio(n, d int) string {
	if d == 0 {
		return fmt.Sprintf("%d/0", n)
	}
	return fmt.Sprintf("%d/%d (%.0f%%)", n, d, 100*float64(n)/float64(d))
}
