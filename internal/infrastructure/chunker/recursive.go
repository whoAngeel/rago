package chunker

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// separators from coarsest to finest. Each pattern matches the text that ends
// a piece; the match stays attached to the piece it ends, so concatenating
// the pieces reproduces the input.
var recursiveSeparators = []*regexp.Regexp{
	regexp.MustCompile(`\n[ \t]*\n\s*`), // blank line (paragraph)
	regexp.MustCompile(`\n`),            // single newline
	regexp.MustCompile(`[.?!]\s+`),      // sentence end, including newline forms
	regexp.MustCompile(`\s+`),           // space
}

// RecursiveChunker splits text on the coarsest separator that yields pieces
// within Size runes and merges adjacent small pieces up to Size. Sizes are in
// runes.
type RecursiveChunker struct {
	Size    int
	Overlap int
}

// NewRecursiveChunker returns a chunker with the given size and overlap in
// runes. Invalid values fall back to the same defaults as the fixed chunker.
func NewRecursiveChunker(size, overlap int) *RecursiveChunker {
	if size <= 0 {
		size = DefaultChunkSize
	}
	if overlap < 0 || overlap >= size {
		overlap = DefaultChunkOverlap
		if overlap >= size {
			overlap = size / 5
		}
	}
	return &RecursiveChunker{Size: size, Overlap: overlap}
}

type piece struct {
	text string
	n    int // rune count
}

// Chunk implements ports.Chunker.
func (c *RecursiveChunker) Chunk(text string) ([]string, error) {
	var pieces []piece
	for _, p := range c.split(text, 0) {
		pieces = append(pieces, piece{p, utf8.RuneCountInString(p)})
	}

	var chunks []string
	var cur []piece // current chunk: overlap pieces first, then new ones
	curLen, fresh := 0, 0
	emit := func() {
		var sb strings.Builder
		for _, p := range cur {
			sb.WriteString(p.text)
		}
		if s := strings.TrimSpace(sb.String()); s != "" {
			chunks = append(chunks, s)
		}
	}
	for _, p := range pieces {
		if curLen+p.n > c.Size && fresh > 0 {
			emit()
			cur, curLen = c.tail(cur, p.n), 0
			for _, q := range cur {
				curLen += q.n
			}
			fresh = 0
		}
		cur = append(cur, p)
		curLen += p.n
		fresh++
	}
	if fresh > 0 {
		emit()
	}
	return chunks, nil
}

// tail returns the trailing whole pieces of prev totalling at most Overlap
// runes, shrunk further so that next still fits in the chunk.
func (c *RecursiveChunker) tail(prev []piece, next int) []piece {
	limit := min(c.Overlap, c.Size-next)
	total, start := 0, len(prev)
	for start > 0 && total+prev[start-1].n <= limit {
		start--
		total += prev[start].n
	}
	return append([]piece(nil), prev[start:]...)
}

// split breaks text into pieces of at most Size runes, trying separators from
// level onwards and hard-cutting as a last resort.
func (c *RecursiveChunker) split(text string, level int) []string {
	if utf8.RuneCountInString(text) <= c.Size {
		return []string{text}
	}
	for ; level < len(recursiveSeparators); level++ {
		cuts := recursiveSeparators[level].FindAllStringIndex(text, -1)
		if len(cuts) == 0 {
			continue
		}
		var out []string
		prev := 0
		for _, m := range cuts {
			if m[1] > prev {
				out = append(out, c.split(text[prev:m[1]], level+1)...)
				prev = m[1]
			}
		}
		if prev < len(text) {
			out = append(out, c.split(text[prev:], level+1)...)
		}
		return out
	}
	runes := []rune(text)
	var out []string
	for i := 0; i < len(runes); i += c.Size {
		out = append(out, string(runes[i:min(i+c.Size, len(runes))]))
	}
	return out
}
