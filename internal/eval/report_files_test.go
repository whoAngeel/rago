package eval

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSummarizeByFileKeepsSameBaseNameApart(t *testing.T) {
	dirA, dirB := t.TempDir(), t.TempDir()
	pathA := filepath.Join(dirA, "questions.json")
	pathB := filepath.Join(dirB, "questions.json")
	if err := os.WriteFile(pathA, []byte(`[{"id":"a1","question":"A?"}]`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(pathB, []byte(`[{"id":"b1","question":"B?"}]`), 0o600); err != nil {
		t.Fatal(err)
	}

	qs, err := LoadQuestionFiles([]string{pathA, pathB})
	if err != nil {
		t.Fatal(err)
	}
	results := make([]QuestionResult, len(qs))
	for i, q := range qs {
		results[i] = QuestionResult{ID: q.ID, File: q.File}
	}

	files := summarizeByFile(results)
	if len(files) != 2 {
		t.Fatalf("got %d file totals, want 2: %+v", len(files), files)
	}
	if files[0].File != pathA || files[1].File != pathB {
		t.Errorf("labels = %q, %q; want %q, %q", files[0].File, files[1].File, pathA, pathB)
	}
}
