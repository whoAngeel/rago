package eval

import (
	"fmt"
	"io"
	"unicode/utf8"
)

// DocChunks is the chunk list one document produces.
type DocChunks struct {
	Name   string
	Chunks []string
}

// WriteChunkDump prints every chunk with its document, index and character
// length. A non-empty source limits the output to that document.
func WriteChunkDump(w io.Writer, docs []DocChunks, source string) error {
	found := false
	for _, d := range docs {
		if source != "" && d.Name != source {
			continue
		}
		found = true
		fmt.Fprintf(w, "=== %s (%d chunks)\n", d.Name, len(d.Chunks))
		for i, c := range d.Chunks {
			fmt.Fprintf(w, "--- %s #%d (%d chars)\n%s\n", d.Name, i, utf8.RuneCountInString(c), c)
		}
	}
	if source != "" && !found {
		return fmt.Errorf("no document named %q among the selected data directories", source)
	}
	return nil
}
