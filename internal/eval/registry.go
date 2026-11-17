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

// NewParserRegistry returns the same parser registry the server uses.
func NewParserRegistry() *parserpkg.Registry {
	return parserpkg.NewDefaultRegistry(parserpkg.NewTesseractOCR())
}

// NewChunker returns the chunker the server uses.
func NewChunker() *chunkerpkg.FixedChunker {
	return chunkerpkg.NewFixedChunker(ChunkSize, ChunkOverlap)
}
