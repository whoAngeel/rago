package chunker

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestRecursiveChunker(t *testing.T) {
	short := "Un texto corto."
	paras := "Primera parrafada completa.\n\nSegunda parrafada completa.\n\nTercera parrafada completa."
	sentences := "Alpha beta gamma. Delta epsilon zeta. Eta theta iota. Kappa lambda mu."
	accents := strings.Repeat("Añoranza canción pingüino. ", 12)

	tests := []struct {
		name    string
		text    string
		size    int
		overlap int
		check   func(t *testing.T, chunks []string)
	}{
		{"under size gives one chunk", short, 100, 10, func(t *testing.T, c []string) {
			if len(c) != 1 || c[0] != short {
				t.Fatalf("got %q", c)
			}
		}},
		{"paragraphs kept whole", paras, 60, 0, func(t *testing.T, c []string) {
			if len(c) != 2 || c[0] != "Primera parrafada completa.\n\nSegunda parrafada completa." || c[1] != "Tercera parrafada completa." {
				t.Fatalf("got %q", c)
			}
		}},
		{"long paragraph splits at sentence ends", sentences, 40, 0, func(t *testing.T, c []string) {
			if len(c) < 2 {
				t.Fatalf("expected a split, got %q", c)
			}
			for _, ch := range c {
				if !strings.HasSuffix(ch, ".") {
					t.Errorf("chunk %q does not end at a sentence end", ch)
				}
			}
		}},
		{"multi-byte text is never cut mid-rune", accents, 25, 0, func(t *testing.T, c []string) {
			for _, ch := range c {
				if !utf8.ValidString(ch) {
					t.Errorf("invalid UTF-8 in %q", ch)
				}
			}
			hard := NewRecursiveChunker(7, 0)
			hc, _ := hard.Chunk(strings.Repeat("ñandú", 10))
			for _, ch := range hc {
				if !utf8.ValidString(ch) || utf8.RuneCountInString(ch) > 7 {
					t.Errorf("bad hard cut %q", ch)
				}
			}
		}},
		{"overlap repeats the tail of the previous chunk", sentences, 40, 20, func(t *testing.T, c []string) {
			if len(c) < 2 {
				t.Fatalf("got %q", c)
			}
			for i := 1; i < len(c); i++ {
				lastSentence := c[i-1][strings.LastIndex(c[i-1][:len(c[i-1])-1], ". ")+2:]
				if !strings.HasPrefix(c[i], lastSentence) {
					t.Errorf("chunk %d %q does not start with tail %q", i, c[i], lastSentence)
				}
			}
		}},
		{"blank input gives no chunks", " \n\n  ", 50, 5, func(t *testing.T, c []string) {
			if len(c) != 0 {
				t.Fatalf("got %q", c)
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunks, err := NewRecursiveChunker(tt.size, tt.overlap).Chunk(tt.text)
			if err != nil {
				t.Fatal(err)
			}
			for _, ch := range chunks {
				if n := utf8.RuneCountInString(ch); n > tt.size {
					t.Errorf("chunk of %d runes exceeds %d: %q", n, tt.size, ch)
				}
				if strings.TrimSpace(ch) != ch || ch == "" {
					t.Errorf("chunk not trimmed or empty: %q", ch)
				}
			}
			tt.check(t, chunks)
		})
	}
}
