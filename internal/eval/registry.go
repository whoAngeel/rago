package eval

import (
	"time"

	chunkerpkg "github.com/whoAngeel/rago/internal/infrastructure/chunker"
	parserpkg "github.com/whoAngeel/rago/internal/infrastructure/parser"
)

// Chunker settings and parse timeout mirror cmd/server/main.go and the ingest
// worker (internal/worker/ingest_worker.go).
const (
	ChunkSize    = 1000
	ChunkOverlap = 200
	ParseTimeout = 10 * time.Minute
)

// NewParserRegistry registers the same parsers as cmd/server/main.go.
func NewParserRegistry() *parserpkg.Registry {
	r := parserpkg.NewRegistry()
	r.Register("text/plain", parserpkg.NewPlainTextAdapter())
	r.Register("text/markdown", parserpkg.NewPlainTextAdapter())
	r.Register("text/csv", parserpkg.NewCSVParser())
	r.Register("application/json", parserpkg.NewJSONParser())
	r.Register("application/vnd.openxmlformats-officedocument.wordprocessingml.document", parserpkg.NewDOCXParser())
	r.Register("application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", parserpkg.NewXLSXParser())
	r.Register("application/pdf", parserpkg.NewPDFParser())
	img := parserpkg.NewImageParser()
	for _, ct := range []string{"image/png", "image/jpeg", "image/webp", "image/tiff"} {
		r.Register(ct, img)
	}
	return r
}

// NewChunker returns the chunker the server uses.
func NewChunker() *chunkerpkg.FixedChunker {
	return chunkerpkg.NewFixedChunker(ChunkSize, ChunkOverlap)
}
