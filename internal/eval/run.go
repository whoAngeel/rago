package eval

import (
	"bytes"
	"context"
	"fmt"

	"github.com/tmc/langchaingo/schema"
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

// ChunkDocs chunks each parsed page on its own, as the ingest worker does;
// structured documents skip the chunker. With header set, every chunk is
// prefixed by a context header line carrying the document title and, for
// page-based documents, the chunk's own page number.
func ChunkDocs(docs []schema.Document, chunker ports.Chunker, name string, header bool) ([]string, error) {
	title := ""
	if header && len(docs) > 0 {
		title = DocTitle(docs[0].PageContent, name)
	}
	var chunks []string
	for _, d := range docs {
		cs := []string{d.PageContent}
		if d.Metadata["chunk_type"] != "structured" {
			var err error
			if cs, err = chunker.Chunk(d.PageContent); err != nil {
				return nil, fmt.Errorf("chunking %s: %w", name, err)
			}
		}
		for _, c := range cs {
			if header {
				c = ContextHeader(title, pageNumber(d.Metadata)) + "\n" + c
			}
			chunks = append(chunks, c)
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

// BuildChunks parses a document and chunks it the way the ingest worker does:
// parser chosen by content type, structured documents skip the chunker.
func BuildChunks(ctx context.Context, reg ports.ParserRegistry, chunker ports.Chunker, name string, content []byte) ([]string, error) {
	docs, err := ParseDocs(ctx, reg, name, content)
	if err != nil {
		return nil, err
	}
	return ChunkDocs(docs, chunker, name, false)
}
