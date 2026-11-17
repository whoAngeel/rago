package application

import "github.com/whoAngeel/rago/internal/core/ports"

// filterByScore drops search results scoring below threshold (higher score is
// better). A threshold <= 0 disables the cutoff.
func filterByScore(log ports.Logger, results []ports.SearchResult, threshold float64) []ports.SearchResult {
	if threshold <= 0 {
		log.Info("search completed", "results", len(results), "passed_threshold", len(results))
		return results
	}
	kept := make([]ports.SearchResult, 0, len(results))
	for _, r := range results {
		if float64(r.Score) >= threshold {
			kept = append(kept, r)
		}
	}
	log.Info("search completed", "results", len(results), "passed_threshold", len(kept), "threshold", threshold)
	return kept
}
