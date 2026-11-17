package eval

import (
	"context"
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

// Guards the real datasets against the real sample files: renaming a file or
// editing a question must break the build instead of the live eval. Optional
// directories (manuals, private) are validated only when they exist.
func TestRealDatasetMatchesTestdata(t *testing.T) {
	const root = "../../testdata"
	srcs, _, err := CollectSources(DefaultDataDirs(root))
	if err != nil {
		t.Fatalf("CollectSources: %v", err)
	}
	paths, err := DefaultQuestionPaths(root)
	if err != nil {
		t.Fatal(err)
	}
	qs, err := LoadQuestionFiles(paths)
	if err != nil {
		t.Fatalf("LoadQuestionFiles: %v", err)
	}
	if err := ValidateQuestions(qs, SourceNames(srcs)); err != nil {
		t.Fatalf("ValidateQuestions: %v", err)
	}

	base, err := LoadQuestions(root + "/eval/questions.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(base) < 24 || len(base) > 30 {
		t.Errorf("questions.json has %d questions, want 24..30", len(base))
	}
	unanswerable := 0
	for _, q := range base {
		if q.ExpectedSource == "" {
			unanswerable++
		}
	}
	if unanswerable == 0 {
		t.Error("questions.json has no unanswerable questions")
	}
}

// Every expected_contains of the committed question sets must exist in the
// chunks the real parsers produce. Private documents are excluded on purpose:
// a question there that fails because the parser lost text is the eval's job.
func TestExpectedContainsExistInChunks(t *testing.T) {
	const root = "../../testdata"
	srcs, _, err := CollectSources(DefaultDataDirs(root))
	if err != nil {
		t.Fatal(err)
	}
	paths, _ := DefaultQuestionPaths(root)
	qs, err := LoadQuestionFiles(paths)
	if err != nil {
		t.Fatal(err)
	}
	byName := map[string]Source{}
	for _, s := range srcs {
		byName[s.Name] = s
	}
	reg, chunker := NewParserRegistry(), NewChunker()
	cache := map[string][]string{}
	for _, q := range qs {
		src, ok := byName[q.ExpectedSource]
		if len(q.ExpectedContains) == 0 || !ok || src.Private() {
			continue
		}
		chunks, done := cache[src.Name]
		if !done {
			content, err := ReadSource(src)
			if err != nil {
				t.Fatal(err)
			}
			if chunks, err = BuildChunks(context.Background(), reg, chunker, src.Name, content); err != nil {
				t.Fatalf("BuildChunks(%s): %v", src.Name, err)
			}
			cache[src.Name] = chunks
		}
		found := false
		for _, c := range chunks {
			if ContainsAll(c, q.ExpectedContains) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("question %q (%s): no chunk of %s contains %q", q.ID, q.File, q.ExpectedSource, q.ExpectedContains)
		}
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

func TestContentTypeForRealFormats(t *testing.T) {
	tests := map[string]string{
		"m.pdf":  "application/pdf",
		"m.PDF":  "application/pdf",
		"w.docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"s.xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"i.png":  "image/png",
		"i.jpg":  "image/jpeg",
		"i.jpeg": "image/jpeg",
		"i.webp": "image/webp",
		"i.tif":  "image/tiff",
		"i.tiff": "image/tiff",
	}
	for name, want := range tests {
		if got, err := ContentTypeFor(name); err != nil || got != want {
			t.Errorf("ContentTypeFor(%q) = %q, %v; want %q", name, got, err, want)
		}
	}
}
