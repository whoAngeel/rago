package application_test

import (
	"context"
	"testing"

	"github.com/whoAngeel/rago/internal/application"
)

type countingEmbedder struct{ calls int }

func (e *countingEmbedder) EmbedText(_ context.Context, _ string) ([]float32, error) {
	e.calls++
	return []float32{1, 2, 3}, nil
}
func (e *countingEmbedder) ComputeEmbeddings(_ context.Context, texts []string) ([][]float32, error) {
	e.calls++
	out := make([][]float32, len(texts))
	for i := range texts {
		out[i] = []float32{1, 2, 3}
	}
	return out, nil
}

type memEmbeddingCache struct{ m map[string][]float32 }

func (c *memEmbeddingCache) Get(_ context.Context, h, model string, dim int) ([]float32, bool, error) {
	v, ok := c.m[h+model]
	return v, ok, nil
}
func (c *memEmbeddingCache) Put(_ context.Context, h, model string, _ int, v []float32) error {
	c.m[h+model] = v
	return nil
}

func TestCachedEmbedder_SecondCallForSameTextSkipsInner(t *testing.T) {
	inner := &countingEmbedder{}
	cache := &memEmbeddingCache{m: map[string][]float32{}}
	e := application.NewCachedEmbedder(inner, cache, "m", 3, rtLogger{})

	first, err := e.EmbedText(context.Background(), "hello")
	if err != nil {
		t.Fatal(err)
	}
	if inner.calls != 1 || len(cache.m) != 1 {
		t.Fatalf("after miss: inner calls = %d, cached = %d; want 1, 1", inner.calls, len(cache.m))
	}
	second, err := e.EmbedText(context.Background(), "hello")
	if err != nil {
		t.Fatal(err)
	}
	if inner.calls != 1 {
		t.Errorf("inner called %d times, want 1", inner.calls)
	}
	if len(second) != 3 || second[0] != first[0] {
		t.Errorf("cached vector = %v, want %v", second, first)
	}
}
