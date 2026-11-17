package eval

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func write(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestCollectSourcesDuplicateBaseNameFails(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "one"), "guide.txt", "first")
	write(t, filepath.Join(root, "two"), "guide.txt", "second")
	_, _, err := CollectSources([]string{filepath.Join(root, "one"), filepath.Join(root, "two")})
	if err == nil || !strings.Contains(err.Error(), "guide.txt") {
		t.Fatalf("error = %v, want one naming guide.txt", err)
	}
}

func TestCollectSourcesSkipsIdenticalContent(t *testing.T) {
	root := t.TempDir()
	write(t, root, "manual (1).txt", "same bytes")
	write(t, root, "manual.txt", "same bytes")
	write(t, root, "other.txt", "different")
	srcs, warns, err := CollectSources([]string{root})
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, s := range srcs {
		names = append(names, s.Name)
	}
	if strings.Join(names, ",") != "manual.txt,other.txt" {
		t.Errorf("names = %v, want manual.txt,other.txt", names)
	}
	if len(warns) != 1 || !strings.Contains(warns[0], "manual (1).txt") || !strings.Contains(warns[0], "manual.txt") {
		t.Errorf("warnings = %q, want one naming both files", warns)
	}
}

func TestCollectSourcesDeterministicAndPrivateJSONIgnored(t *testing.T) {
	root := t.TempDir()
	write(t, root, "b.txt", "b")
	write(t, root, "a.txt", "a")
	write(t, filepath.Join(root, "private"), "doc.txt", "d")
	write(t, filepath.Join(root, "private"), "questions.json", "[]")
	write(t, filepath.Join(root, "eval"), "q.json", "[]")
	dirs := []string{root, filepath.Join(root, "private")}
	first, _, err := CollectSources(dirs)
	if err != nil {
		t.Fatal(err)
	}
	second, _, _ := CollectSources(dirs)
	var got []string
	for _, s := range first {
		got = append(got, s.Name)
	}
	if strings.Join(got, ",") != "a.txt,b.txt,doc.txt" {
		t.Errorf("names = %v", got)
	}
	for i := range first {
		if first[i].Name != second[i].Name {
			t.Error("order is not deterministic")
		}
	}
}

func TestCollectSourcesMissingExplicitDirFails(t *testing.T) {
	if _, _, err := CollectSources([]string{filepath.Join(t.TempDir(), "nope")}); err == nil {
		t.Error("expected error for a missing explicit directory")
	}
}

func TestDefaultDataDirsSkipsMissingOptionalOnes(t *testing.T) {
	root := t.TempDir()
	got := DefaultDataDirs(root)
	if len(got) != 1 || got[0] != root {
		t.Fatalf("dirs = %v, want only the root", got)
	}
	if err := os.MkdirAll(filepath.Join(root, "private"), 0o755); err != nil {
		t.Fatal(err)
	}
	got = DefaultDataDirs(root)
	if len(got) != 2 || got[1] != filepath.Join(root, "private") {
		t.Errorf("dirs = %v, want root and private", got)
	}
}

func TestLoadQuestionFilesDuplicateIDAcrossFiles(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "a.json", `[{"id":"q1","question":"x?","expected_source":"a.txt"}]`)
	write(t, dir, "b.json", `[{"id":"q1","question":"y?","expected_source":"a.txt"}]`)
	_, err := LoadQuestionFiles([]string{filepath.Join(dir, "a.json"), filepath.Join(dir, "b.json")})
	if err == nil || !strings.Contains(err.Error(), "q1") {
		t.Fatalf("error = %v, want duplicate id q1", err)
	}
}

func TestLoadQuestionFilesRecordsOrigin(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "a.json", `[{"id":"q1","question":"x?","expected_source":"a.txt","expected_contains":["k"]}]`)
	qs, err := LoadQuestionFiles([]string{filepath.Join(dir, "a.json")})
	if err != nil {
		t.Fatal(err)
	}
	if qs[0].File != "a.json" || len(qs[0].ExpectedContains) != 1 {
		t.Errorf("question = %+v", qs[0])
	}
}

func TestDefaultQuestionPaths(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "eval"), "b.json", "[]")
	write(t, filepath.Join(root, "eval"), "a.json", "[]")
	write(t, filepath.Join(root, "eval"), "notes.txt", "")
	got, err := DefaultQuestionPaths(root)
	if err != nil || len(got) != 2 || filepath.Base(got[0]) != "a.json" {
		t.Fatalf("paths = %v, %v", got, err)
	}
	write(t, filepath.Join(root, "private"), "questions.json", "[]")
	got, _ = DefaultQuestionPaths(root)
	if len(got) != 3 || filepath.Base(got[2]) != "questions.json" {
		t.Errorf("paths with private = %v", got)
	}
}
