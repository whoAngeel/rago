package eval

import (
	"time"

	"github.com/whoAngeel/rago/internal/core/ports"
	chunkerpkg "github.com/whoAngeel/rago/internal/infrastructure/chunker"
	parserpkg "github.com/whoAngeel/rago/internal/infrastructure/parser"
)

// ParseTimeout mirrors the ingest worker (internal/worker/ingest_worker.go).
const ParseTimeout = 10 * time.Minute

// NewParserRegistry returns the same parser registry the server uses.
func NewParserRegistry(ocr ports.OCR, pdfOpts ...parserpkg.PDFOption) *parserpkg.Registry {
	return parserpkg.NewDefaultRegistry(ocr, pdfOpts...)
}

// NewChunker returns the chunker the server uses by default.
func NewChunker() ports.Chunker {
	c, err := chunkerpkg.DefaultSettings().Chunker()
	if err != nil { // unreachable: the defaults are valid
		panic(err)
	}
	return c
}

// ChunkConfig selects the chunking strategy of an eval run.
type ChunkConfig = chunkerpkg.Settings
