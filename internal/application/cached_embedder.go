package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/whoAngeel/rago/internal/core/ports"
)

// CachedEmbedder wraps an Embedder with a content-addressed cache so the same
// text is never embedded (and paid for) twice for the same model and dimension.
type CachedEmbedder struct {
	inner ports.Embedder
	cache ports.EmbeddingCache
	model string
	dim   int
	log   ports.Logger
}

var _ ports.Embedder = (*CachedEmbedder)(nil)

func NewCachedEmbedder(inner ports.Embedder, cache ports.EmbeddingCache, model string, dim int, log ports.Logger) ports.Embedder {
	return &CachedEmbedder{inner: inner, cache: cache, model: model, dim: dim, log: log}
}

func textHash(text string) string {
	sum := sha256.Sum256([]byte(text))
	return hex.EncodeToString(sum[:])
}

// lookup returns the cached vector for text; cache errors and vectors of the
// wrong length count as a miss.
func (c *CachedEmbedder) lookup(ctx context.Context, hash string) ([]float32, bool) {
	vec, ok, err := c.cache.Get(ctx, hash, c.model, c.dim)
	if err != nil {
		c.log.Warn("embedding cache read failed", "error", err)
		return nil, false
	}
	if !ok || len(vec) != c.dim {
		return nil, false
	}
	return vec, true
}

func (c *CachedEmbedder) store(ctx context.Context, hash string, vec []float32) {
	if len(vec) != c.dim {
		return
	}
	if err := c.cache.Put(ctx, hash, c.model, c.dim, vec); err != nil {
		c.log.Warn("embedding cache write failed", "error", err)
	}
}

func (c *CachedEmbedder) EmbedText(ctx context.Context, text string) ([]float32, error) {
	hash := textHash(text)
	if vec, ok := c.lookup(ctx, hash); ok {
		return vec, nil
	}
	vec, err := c.inner.EmbedText(ctx, text)
	if err != nil {
		return nil, err
	}
	c.store(ctx, hash, vec)
	return vec, nil
}

func (c *CachedEmbedder) ComputeEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	out := make([][]float32, len(texts))
	hashes := make([]string, len(texts))
	var missIdx []int
	var missTexts []string
	for i, t := range texts {
		hashes[i] = textHash(t)
		if vec, ok := c.lookup(ctx, hashes[i]); ok {
			out[i] = vec
			continue
		}
		missIdx = append(missIdx, i)
		missTexts = append(missTexts, t)
	}
	if len(missTexts) == 0 {
		return out, nil
	}
	vecs, err := c.inner.ComputeEmbeddings(ctx, missTexts)
	if err != nil {
		return nil, err
	}
	if len(vecs) != len(missTexts) {
		return nil, fmt.Errorf("embedder returned %d vectors for %d texts", len(vecs), len(missTexts))
	}
	for j, i := range missIdx {
		out[i] = vecs[j]
		c.store(ctx, hashes[i], vecs[j])
	}
	return out, nil
}
