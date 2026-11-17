package eval

import (
	"strings"
	"testing"
)

func TestDocTitleAndContextHeader(t *testing.T) {
	long := strings.Repeat("ñ", 100)
	tests := []struct {
		name, text, file string
		page             int
		want             string
	}{
		{"first non-empty line", "\n\nManual KC60\nsecond", "a.pdf", 3, "[Manual KC60 — p. 3]"},
		{"markdown hashes stripped", "## Guía  rápida\ntext", "guide.md", 0, "[Guía rápida]"},
		{"whitespace collapsed", "  A \t B   C ", "x.txt", 0, "[A B C]"},
		{"truncated to 80 runes", long, "x.txt", 1, "[" + strings.Repeat("ñ", 80) + " — p. 1]"},
		{"empty line falls back to file name", "###\nbody", "kb-faq.v2.txt", 0, "[kb-faq.v2]"},
		{"empty text falls back to file name", "", "dir/faq.txt", 2, "[faq — p. 2]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContextHeader(DocTitle(tt.text, tt.file), tt.page); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
