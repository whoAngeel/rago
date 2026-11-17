package eval

import (
	"bytes"
	"context"
	"fmt"

	"github.com/whoAngeel/rago/internal/core/ports"
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
			hits[i] = Hit{Source: src, Score: f.Score}
		}
		r := QuestionResult{ID: q.ID, ExpectedSource: q.ExpectedSource, Note: q.Note}
		if len(hits) > 0 {
			r.TopScore, r.TopSource = hits[0].Score, hits[0].Source
		}
		if q.Answerable() {
			r.Rank, r.Score = FirstRank(hits, q.ExpectedSource)
		}
		out = append(out, r)
	}
	return out, nil
}

// BuildChunks parses a document and chunks it the way the ingest worker does:
// parser chosen by content type, structured documents skip the chunker.
func BuildChunks(ctx context.Context, reg ports.ParserRegistry, chunker ports.Chunker, name string, content []byte) ([]string, error) {
	ct, err := ContentTypeFor(name)
	if err != nil {
		return nil, err
	}
	parser, err := reg.Resolve(ct)
	if err != nil {
		return nil, fmt.Errorf("resolving parser for %s: %w", name, err)
	}
	docs, err := parser.Parse(ctx, bytes.NewReader(content), ct)
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", name, err)
	}
	var chunks []string
	for _, d := range docs {
		if d.Metadata["chunk_type"] == "structured" {
			chunks = append(chunks, d.PageContent)
			continue
		}
		cs, err := chunker.Chunk(d.PageContent)
		if err != nil {
			return nil, fmt.Errorf("chunking %s: %w", name, err)
		}
		chunks = append(chunks, cs...)
	}
	return chunks, nil
}
