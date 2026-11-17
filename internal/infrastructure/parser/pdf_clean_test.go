package parser

import (
	"strings"
	"testing"
)

func TestNormalizeText(t *testing.T) {
	tests := []struct{ name, in, want string }{
		{"ligature fi", "conﬁguration", "configuration"},
		{"ligature ffl", "oﬄine", "offline"},
		{"fullwidth digits", "１２", "12"},
		{"plain unchanged", "hold for 3 seconds", "hold for 3 seconds"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeText(tt.in); got != tt.want {
				t.Errorf("normalizeText(%q) = %q; want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestDehyphenate(t *testing.T) {
	tests := []struct{ name, in, want string }{
		{"joins across newline", "config-\nuration", "configuration"},
		{"keeps compound on one line", "a well-known fact", "a well-known fact"},
		{"keeps hyphen before uppercase", "Wi-\nFi", "Wi-\nFi"},
		{"keeps hyphen before digit", "ref-\n12", "ref-\n12"},
		{"keeps dash after space", "a -\nb", "a -\nb"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dehyphenate(tt.in); got != tt.want {
				t.Errorf("dehyphenate(%q) = %q; want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestCollapseWhitespace(t *testing.T) {
	tests := []struct{ name, in, want string }{
		{"trailing spaces", "a  \nb\t\n", "a\nb"},
		{"three newlines", "a\n\n\nb", "a\n\nb"},
		{"many newlines", "a\n\n\n\n\nb", "a\n\nb"},
		{"two newlines kept", "a\n\nb", "a\n\nb"},
		{"spaces between blank lines", "a\n  \n \nb", "a\n\nb"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := collapseWhitespace(tt.in); got != tt.want {
				t.Errorf("collapseWhitespace(%q) = %q; want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestRemoveRepeatedLines(t *testing.T) {
	t.Run("removed on four pages with numbered footers", func(t *testing.T) {
		pages := []string{
			"ACME Manual\nbody one\nPage 1 of 4",
			"ACME Manual\nbody two\nPage 2 of 4",
			"ACME Manual\nbody three\nPage 3 of 4",
			"ACME Manual\nbody four\nPage 4 of 4",
		}
		got := removeRepeatedLines(pages)
		want := []string{"body one", "body two", "body three", "body four"}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("page %d = %q; want %q", i+1, got[i], want[i])
			}
		}
	})
	t.Run("untouched on two pages", func(t *testing.T) {
		pages := []string{"ACME\nbody one\nPage 1", "ACME\nbody two\nPage 2"}
		got := removeRepeatedLines(pages)
		for i := range pages {
			if got[i] != pages[i] {
				t.Errorf("page %d = %q; want unchanged %q", i+1, got[i], pages[i])
			}
		}
	})
	t.Run("line on one of four pages is kept", func(t *testing.T) {
		pages := []string{
			"Unique header\nbody one\nend one",
			"title two\nbody two\nend two",
			"title three\nbody three\nend three",
			"title four\nbody four\nend four",
		}
		got := removeRepeatedLines(pages)
		for i := range pages {
			if got[i] != pages[i] {
				t.Errorf("page %d = %q; want unchanged %q", i+1, got[i], pages[i])
			}
		}
	})
	t.Run("does not mutate input", func(t *testing.T) {
		pages := []string{"H\na\nF 1", "H\nb\nF 2", "H\nc\nF 3"}
		_ = removeRepeatedLines(pages)
		if pages[0] != "H\na\nF 1" {
			t.Errorf("input mutated: %q", pages[0])
		}
	})
}

func TestNeedsOCR(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"normal paragraph", "Press and hold the power button for three seconds to reset the device.", false},
		{"empty", "", true},
		{"whitespace only", " \n\t ", true},
		{"ten letters", "abcdefghij", true},
		{"replacement characters", strings.Repeat("�", 60) + " some real words here ok", true},
		{"private use characters", strings.Repeat("", 60) + " some real words here ok", true},
		{"few bad runes", "Press and hold the power button for three seconds � to reset.", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := needsOCR(tt.in); got != tt.want {
				t.Errorf("needsOCR(%q) = %v; want %v", tt.in, got, tt.want)
			}
		})
	}
}
