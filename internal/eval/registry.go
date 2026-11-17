package eval

import (
	"fmt"
	"time"

	"github.com/whoAngeel/rago/internal/core/ports"
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
func NewParserRegistry(ocr ports.OCR, pdfOpts ...parserpkg.PDFOption) *parserpkg.Registry {
	return parserpkg.NewDefaultRegistry(ocr, pdfOpts...)
}

// NewChunker returns the chunker the server uses.
func NewChunker() *chunkerpkg.FixedChunker {
	return chunkerpkg.NewFixedChunker(ChunkSize, ChunkOverlap)
}

// ChunkConfig selects the chunking strategy of an eval run.
type ChunkConfig struct {
	Kind    string // "fixed" or "recursive"
	Size    int
	Overlap int
	Header  bool
}

// DefaultChunkConfig is the server's configuration: the baseline.
func DefaultChunkConfig() ChunkConfig {
	return ChunkConfig{Kind: "fixed", Size: ChunkSize, Overlap: ChunkOverlap}
}

// Chunker builds the configured chunker.
func (c ChunkConfig) Chunker() (ports.Chunker, error) {
	if c.Size <= 0 || c.Overlap < 0 || c.Overlap >= c.Size {
		return nil, fmt.Errorf("invalid chunking: size %d, overlap %d (need size > 0 and 0 <= overlap < size)", c.Size, c.Overlap)
	}
	switch c.Kind {
	case "fixed":
		return chunkerpkg.NewFixedChunker(c.Size, c.Overlap), nil
	case "recursive":
		return chunkerpkg.NewRecursiveChunker(c.Size, c.Overlap), nil
	}
	return nil, fmt.Errorf("unknown chunker %q (want fixed or recursive)", c.Kind)
}

// String is the configuration line printed at the start of a run.
func (c ChunkConfig) String() string {
	return fmt.Sprintf("chunking: %s, size %d, overlap %d, context header %t", c.Kind, c.Size, c.Overlap, c.Header)
}
