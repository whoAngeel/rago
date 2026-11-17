package eval

import "testing"

func TestNormalizeText(t *testing.T) {
	tests := []struct{ name, in, want string }{
		{"ligature folded", "ﬁnal ﬂow", "final flow"},
		{"case folded", "Model HT-X200", "model ht-x200"},
		{"whitespace collapsed", "a \n\t b   c\u00a0d", "a b c d"},
		{"trimmed", "  x  ", "x"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeText(tt.in); got != tt.want {
				t.Errorf("NormalizeText(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
	if NormalizeText("café") == NormalizeText("cafe") {
		t.Error("accented letter must stay distinct from the plain one")
	}
}

func TestContainsAll(t *testing.T) {
	text := "The  ﬁrst Button is\nLabelled PAIR."
	if !ContainsAll(text, []string{"first button", "pair"}) {
		t.Error("expected all strings found after normalisation")
	}
	if ContainsAll(text, []string{"pair", "missing"}) {
		t.Error("one missing string must fail the match")
	}
	if ContainsAll(text, nil) {
		t.Error("an empty list is not a chunk-level check")
	}
}
