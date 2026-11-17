// Package eval holds the pure logic of the retrieval evaluation harness:
// dataset handling, metrics, threshold sweep, reporting and the safety guard.
// Nothing here talks to the network; external systems come in through ports.
package eval

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Question is one evaluation item. An empty ExpectedSource marks a question
// whose answer is in no document.
type Question struct {
	ID             string `json:"id"`
	Question       string `json:"question"`
	ExpectedSource string `json:"expected_source"`
	Note           string `json:"note"`
}

// Answerable reports whether some document is expected to answer the question.
func (q Question) Answerable() bool { return q.ExpectedSource != "" }

// LoadQuestions reads a JSON dataset from path.
func LoadQuestions(path string) ([]Question, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading questions: %w", err)
	}
	var qs []Question
	if err := json.Unmarshal(data, &qs); err != nil {
		return nil, fmt.Errorf("decoding questions %s: %w", path, err)
	}
	return qs, nil
}

// ListSourceFiles returns the sorted names of regular files directly in dir.
func ListSourceFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("listing %s: %w", dir, err)
	}
	var names []string
	for _, e := range entries {
		if e.Type().IsRegular() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

// ValidateQuestions checks ids, texts and that every expected source exists
// among files.
func ValidateQuestions(qs []Question, files []string) error {
	if len(qs) == 0 {
		return fmt.Errorf("no questions in dataset")
	}
	known := make(map[string]bool, len(files))
	for _, f := range files {
		known[f] = true
	}
	seen := make(map[string]bool, len(qs))
	for i, q := range qs {
		if strings.TrimSpace(q.ID) == "" {
			return fmt.Errorf("question #%d: empty id", i+1)
		}
		if seen[q.ID] {
			return fmt.Errorf("duplicate id %q", q.ID)
		}
		seen[q.ID] = true
		if strings.TrimSpace(q.Question) == "" {
			return fmt.Errorf("question %q: empty question", q.ID)
		}
		if q.ExpectedSource != "" && !known[q.ExpectedSource] {
			return fmt.Errorf("question %q: expected_source %q not found in data directory", q.ID, q.ExpectedSource)
		}
	}
	return nil
}

// ContentTypeFor maps a file name to the content type the parser registry
// uses, the same way an upload would be tagged.
func ContentTypeFor(name string) (string, error) {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".txt":
		return "text/plain", nil
	case ".md":
		return "text/markdown", nil
	case ".csv":
		return "text/csv", nil
	case ".json":
		return "application/json", nil
	}
	return "", fmt.Errorf("unsupported file extension for %q", name)
}
