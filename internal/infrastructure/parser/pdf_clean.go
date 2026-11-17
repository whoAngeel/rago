package parser

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

const (
	// minTextRunes is the minimum number of letters/digits a page needs to be
	// considered as having a usable text layer.
	minTextRunes = 20
	// maxBadRuneRatio is the share of non-space runes that may be U+FFFD,
	// private-use or control characters before the text layer is garbage.
	maxBadRuneRatio = 0.10
	// minRepeatedLineTenths is the share of pages (in tenths) on which a first
	// or last line must repeat to be treated as a header or footer.
	minRepeatedLineTenths = 6
	// minPagesForHeaderRemoval is the smallest document that gets header and
	// footer removal.
	minPagesForHeaderRemoval = 3
)

var (
	hyphenBreak   = regexp.MustCompile(`(\p{L})-\n(\p{Ll})`)
	trailingSpace = regexp.MustCompile(`[ \t\r]+\n`)
	blankRuns     = regexp.MustCompile(`\n{3,}`)
	digitRun      = regexp.MustCompile(`\d+`)
)

// normalizeText applies Unicode NFKC, which also expands ligatures.
func normalizeText(s string) string {
	return norm.NFKC.String(s)
}

// dehyphenate joins words split across lines: letter, hyphen, newline,
// lowercase letter. Other hyphens are left alone.
func dehyphenate(s string) string {
	return hyphenBreak.ReplaceAllString(s, "$1$2")
}

// collapseWhitespace trims trailing blanks per line, collapses runs of three
// or more newlines into two and trims the ends.
func collapseWhitespace(s string) string {
	s = trailingSpace.ReplaceAllString(s, "\n")
	s = blankRuns.ReplaceAllString(s, "\n\n")
	return strings.TrimSpace(s)
}

// cleanPageText runs the per-page cleaning steps in order.
func cleanPageText(s string) string {
	return collapseWhitespace(dehyphenate(normalizeText(s)))
}

// lineKey normalises a line so that page numbers do not break matching.
func lineKey(line string) string {
	return digitRun.ReplaceAllString(strings.TrimSpace(line), "#")
}

// edgeLines returns the indexes of the first and last non-empty lines, or -1.
func edgeLines(lines []string) (first, last int) {
	first, last = -1, -1
	for i, l := range lines {
		if strings.TrimSpace(l) == "" {
			continue
		}
		if first == -1 {
			first = i
		}
		last = i
	}
	return first, last
}

// removeRepeatedLines drops running headers and footers: a first (or last)
// non-empty line that repeats, digits ignored, in the same position on at
// least 60% of the pages is removed from every page. Documents shorter than
// three pages are returned untouched. The input is not modified.
func removeRepeatedLines(pages []string) []string {
	out := make([]string, len(pages))
	copy(out, pages)
	if len(pages) < minPagesForHeaderRemoval {
		return out
	}

	split := make([][]string, len(pages))
	firstCount := map[string]int{}
	lastCount := map[string]int{}
	for i, p := range pages {
		split[i] = strings.Split(p, "\n")
		first, last := edgeLines(split[i])
		if first == -1 {
			continue
		}
		firstCount[lineKey(split[i][first])]++
		if last != first {
			lastCount[lineKey(split[i][last])]++
		}
	}

	repeated := func(counts map[string]int, key string) bool {
		return counts[key]*10 >= len(pages)*minRepeatedLineTenths
	}

	for i, lines := range split {
		first, last := edgeLines(lines)
		if first == -1 {
			continue
		}
		drop := map[int]bool{}
		if repeated(firstCount, lineKey(lines[first])) {
			drop[first] = true
		}
		if last != first && repeated(lastCount, lineKey(lines[last])) {
			drop[last] = true
		}
		if len(drop) == 0 {
			continue
		}
		kept := make([]string, 0, len(lines))
		for j, l := range lines {
			if !drop[j] {
				kept = append(kept, l)
			}
		}
		out[i] = strings.Join(kept, "\n")
	}
	return out
}

// needsOCR reports whether a cleaned page lacks a usable text layer: fewer
// than 20 letters/digits, or more than 10% of its non-space runes are U+FFFD,
// private-use or control characters.
func needsOCR(text string) bool {
	var useful, bad, nonSpace int
	for _, r := range text {
		if unicode.IsSpace(r) {
			continue
		}
		nonSpace++
		switch {
		case r == unicode.ReplacementChar || unicode.Is(unicode.Co, r) || unicode.IsControl(r):
			bad++
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			useful++
		}
	}
	if useful < minTextRunes {
		return true
	}
	return float64(bad) > maxBadRuneRatio*float64(nonSpace)
}
