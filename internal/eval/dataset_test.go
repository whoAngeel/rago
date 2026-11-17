package eval

import (
	"strings"
	"testing"
)

func TestValidateQuestions(t *testing.T) {
	files := []string{"a.txt", "b.csv"}
	ok := Question{ID: "q1", Question: "What is a?", ExpectedSource: "a.txt"}

	tests := []struct {
		name    string
		qs      []Question
		wantErr string
	}{
		{"valid answerable and unanswerable", []Question{ok, {ID: "q2", Question: "Unrelated?", ExpectedSource: ""}}, ""},
		{"duplicate id", []Question{ok, {ID: "q1", Question: "Other?", ExpectedSource: "b.csv"}}, "duplicate id"},
		{"unknown source file", []Question{{ID: "q1", Question: "x?", ExpectedSource: "gone.txt"}}, "gone.txt"},
		{"empty question", []Question{{ID: "q1", Question: "  ", ExpectedSource: "a.txt"}}, "empty question"},
		{"empty id", []Question{{ID: "", Question: "x?", ExpectedSource: "a.txt"}}, "empty id"},
		{"no questions", nil, "no questions"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateQuestions(tt.qs, files)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

// Guards the real dataset against the real sample files: renaming a file in
// testdata/ must break the build instead of the live eval.
func TestRealDatasetMatchesTestdata(t *testing.T) {
	qs, err := LoadQuestions("../../testdata/eval/questions.json")
	if err != nil {
		t.Fatalf("LoadQuestions: %v", err)
	}
	files, err := ListSourceFiles("../../testdata")
	if err != nil {
		t.Fatalf("ListSourceFiles: %v", err)
	}
	if err := ValidateQuestions(qs, files); err != nil {
		t.Fatalf("ValidateQuestions: %v", err)
	}
	if len(qs) < 24 || len(qs) > 30 {
		t.Errorf("dataset has %d questions, want 24..30", len(qs))
	}
	unanswerable := 0
	for _, q := range qs {
		if q.ExpectedSource == "" {
			unanswerable++
		}
	}
	if unanswerable == 0 {
		t.Error("dataset has no unanswerable questions")
	}
}

func TestContentTypeFor(t *testing.T) {
	tests := map[string]string{
		"a.txt":  "text/plain",
		"b.CSV":  "text/csv",
		"c.json": "application/json",
		"d.md":   "text/markdown",
	}
	for name, want := range tests {
		got, err := ContentTypeFor(name)
		if err != nil || got != want {
			t.Errorf("ContentTypeFor(%q) = %q, %v; want %q", name, got, err, want)
		}
	}
	if _, err := ContentTypeFor("x.bin"); err == nil {
		t.Error("expected error for unsupported extension")
	}
}
