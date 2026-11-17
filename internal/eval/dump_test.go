package eval

import (
	"bytes"
	"strings"
	"testing"
)

func TestWriteChunkDump(t *testing.T) {
	docs := []DocChunks{
		{Name: "a.pdf", Chunks: []string{"héllo", "second chunk"}},
		{Name: "b.txt", Chunks: []string{"only"}},
	}
	var buf bytes.Buffer
	if err := WriteChunkDump(&buf, docs, ""); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{"=== a.pdf (2 chunks)", "--- a.pdf #0 (5 chars)", "--- a.pdf #1 (12 chars)", "second chunk", "=== b.txt (1 chunks)"} {
		if !strings.Contains(out, want) {
			t.Errorf("dump missing %q:\n%s", want, out)
		}
	}

	buf.Reset()
	if err := WriteChunkDump(&buf, docs, "b.txt"); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(buf.String(), "a.pdf") || !strings.Contains(buf.String(), "only") {
		t.Errorf("filtered dump = %s", buf.String())
	}
	if err := WriteChunkDump(&buf, docs, "zzz"); err == nil {
		t.Error("expected error for an unknown source")
	}
}
