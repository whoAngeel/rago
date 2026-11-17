package chunker

import (
	"fmt"

	"github.com/tmc/langchaingo/schema"
	"github.com/whoAngeel/rago/internal/core/ports"
)

// ChunkPages chunks each parsed page on its own; structured documents bypass
// the chunker. With header set, every chunk is prefixed by a context header
// line carrying the document title and, for page-based documents, the chunk's
// own page number. The ingest worker and the eval share this path.
func ChunkPages(docs []schema.Document, fileName string, c ports.Chunker, header bool) ([]string, error) {
	title := ""
	if header && len(docs) > 0 {
		title = DocTitle(docs[0].PageContent, fileName)
	}
	var chunks []string
	for _, d := range docs {
		cs := []string{d.PageContent}
		if d.Metadata["chunk_type"] != "structured" {
			var err error
			if cs, err = c.Chunk(d.PageContent); err != nil {
				return nil, fmt.Errorf("chunking %s: %w", fileName, err)
			}
		}
		for _, chunk := range cs {
			if header {
				chunk = ContextHeader(title, pageNumber(d.Metadata)) + "\n" + chunk
			}
			chunks = append(chunks, chunk)
		}
	}
	return chunks, nil
}

// pageNumber reads the page_number metadata, which is an int when freshly
// parsed and a float64 after a JSON round trip; 0 when absent.
func pageNumber(md map[string]any) int {
	switch v := md["page_number"].(type) {
	case int:
		return v
	case float64:
		return int(v)
	}
	return 0
}
