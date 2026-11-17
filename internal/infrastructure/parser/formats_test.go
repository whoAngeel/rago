package parser

import "testing"

func TestContentTypeForExtension(t *testing.T) {
	tests := []struct {
		ext    string
		want   string
		wantOK bool
	}{
		{".txt", "text/plain", true},
		{".md", "text/markdown", true},
		{".pdf", "application/pdf", true},
		{".PDF", "application/pdf", true},
		{".csv", "", false},
		{".json", "", false},
		{".docx", "", false},
		{".xlsx", "", false},
		{".png", "", false},
		{"", "", false},
	}
	for _, tt := range tests {
		got, ok := ContentTypeForExtension(tt.ext)
		if got != tt.want || ok != tt.wantOK {
			t.Errorf("ContentTypeForExtension(%q) = %q, %v; want %q, %v", tt.ext, got, ok, tt.want, tt.wantOK)
		}
	}
}

func TestNewDefaultRegistry(t *testing.T) {
	r := NewDefaultRegistry(nil)
	for _, ct := range []string{"text/plain", "text/markdown", "application/pdf"} {
		if _, err := r.Get(ct); err != nil {
			t.Errorf("Get(%q) failed: %v", ct, err)
		}
	}
	for _, ct := range []string{"text/csv", "application/json", "image/png"} {
		if _, err := r.Get(ct); err == nil {
			t.Errorf("Get(%q) resolved; want unsupported", ct)
		}
	}
}
