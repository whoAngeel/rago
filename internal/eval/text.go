package eval

import (
	"strings"

	"golang.org/x/text/unicode/norm"
)

// NormalizeText prepares text for chunk-level matching: Unicode NFKC (folds
// PDF ligatures such as "ﬁ" into "fi"), lower-case, and every whitespace run
// collapsed to one space. Accents are kept, so "é" and "e" stay distinct.
func NormalizeText(s string) string {
	s = strings.ToLower(norm.NFKC.String(s))
	return strings.Join(strings.Fields(s), " ")
}

// ContainsAll reports whether the normalised text contains every normalised
// string in want. An empty list never matches: it means "no chunk-level check".
func ContainsAll(text string, want []string) bool {
	if len(want) == 0 {
		return false
	}
	t := NormalizeText(text)
	for _, w := range want {
		if !strings.Contains(t, NormalizeText(w)) {
			return false
		}
	}
	return true
}
