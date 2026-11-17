package eval

import (
	"bytes"
	"context"
	"fmt"

	"github.com/tmc/langchaingo/schema"
	"github.com/whoAngeel/rago/internal/core/ports"
	chunkerpkg "github.com/whoAngeel/rago/internal/infrastructure/chunker"
)

// RunQuestions embeds each question and searches the collection with no
// score threshold, so every score is visible.
func RunQuestions(ctx context.Context, emb ports.Embedder, store ports.VectorStore, collection string, userID int, qs []Question, topK int) ([]QuestionResult, error) {
	out := make([]QuestionResult, 0, len(qs))
	for _, q := range qs {
		vec, err := emb.EmbedText(ctx, q.Question)
		if err != nil {
			return nil, fmt.Errorf("embedding %q: %w", q.ID, err)
		}
		found, err := store.Search(ctx, collection, vec, userID, nil, topK)
		if err != nil {
			return nil, fmt.Errorf("searching %q: %w", q.ID, err)
		}
		hits := make([]Hit, len(found))
		for i, f := range found {
			src, _ := f.Document.Metadata["source"].(string)
			hits[i] = Hit{Source: src, Text: f.Document.PageContent, Score: f.Score}
		}
		r := QuestionResult{ID: q.ID, ExpectedSource: q.ExpectedSource, Note: q.Note, File: q.File}
		if len(hits) > 0 {
			r.TopScore, r.TopSource = hits[0].Score, hits[0].Source
		}
		if q.Answerable() {
			r.Rank, r.Score = FirstRank(hits, q.ExpectedSource)
			if len(q.ExpectedContains) > 0 {
				r.HasChunkCheck = true
				r.ChunkRank, r.ChunkScore = FirstChunkRank(hits, q.ExpectedSource, q.ExpectedContains)
			}
		}
		out = append(out, r)
	}
	return out, nil
}

// ParseDocs parses a document with the parser chosen by its content type.
func ParseDocs(ctx context.Context, reg ports.ParserRegistry, name string, content []byte) ([]schema.Document, error) {
	ct, err := ContentTypeFor(name)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, ParseTimeout)
	defer cancel()
	parser, err := reg.Resolve(ct)
	if err != nil {
		return nil, fmt.Errorf("resolving parser for %s: %w", name, err)
	}
	docs, err := parser.Parse(ctx, bytes.NewReader(content), ct)
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", name, err)
	}
	return docs, nil
}

// BuildChunks parses a document and chunks it the way the ingest worker does:
// parser chosen by content type, structured documents skip the chunker.
func BuildChunks(ctx context.Context, reg ports.ParserRegistry, chunker ports.Chunker, name string, content []byte) ([]string, error) {
	docs, err := ParseDocs(ctx, reg, name, content)
	if err != nil {
		return nil, err
	}
	return chunkerpkg.ChunkPages(docs, name, chunker, false)
}
