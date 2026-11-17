package application_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/tmc/langchaingo/schema"
	"github.com/whoAngeel/rago/internal/application"
	"github.com/whoAngeel/rago/internal/core/domain"
	"github.com/whoAngeel/rago/internal/infrastructure/config"
)

type batchEmbedder struct {
	rtEmbedder
	batches []int
	failAt  int // 1-based batch number that fails; 0 never
	short   bool
}

func (e *batchEmbedder) ComputeEmbeddings(_ context.Context, texts []string) ([][]float32, error) {
	e.batches = append(e.batches, len(texts))
	if e.failAt > 0 && len(e.batches) == e.failAt {
		return nil, errors.New("provider down")
	}
	n := len(texts)
	if e.short {
		n--
	}
	out := make([][]float32, n)
	for i := range out {
		out[i] = []float32{0.1}
	}
	return out, nil
}

type recordingStore struct {
	rtVectorStore
	upserted, deletes int
}

func (s *recordingStore) UpsertDocuments(_ context.Context, _ string, docs []schema.Document, _ [][]float32) error {
	s.upserted += len(docs)
	return nil
}
func (s *recordingStore) DeleteByDocumentID(context.Context, string, int) error {
	s.deletes++
	return nil
}

func makeChunks(n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = fmt.Sprintf("chunk %d", i)
	}
	return out
}

func noStep(string) func(error) { return func(error) {} }

func TestIngestExecute_BatchesEmbeddingsAndUpserts(t *testing.T) {
	emb, store := &batchEmbedder{}, &recordingStore{}
	uc := application.NewIngestUsecase(store, emb, rtLogger{}, config.Config{})
	if err := uc.Execute(context.Background(), &domain.Document{ID: 1}, nil, makeChunks(70), noStep); err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(emb.batches) != "[32 32 6]" {
		t.Errorf("batches = %v, want [32 32 6]", emb.batches)
	}
	if store.upserted != 70 {
		t.Errorf("upserted = %d, want 70", store.upserted)
	}
}

func TestIngestExecute_FailedBatchStoresNothing(t *testing.T) {
	for name, emb := range map[string]*batchEmbedder{
		"batch error":       {failAt: 2},
		"vector count diff": {short: true},
	} {
		t.Run(name, func(t *testing.T) {
			store := &recordingStore{}
			uc := application.NewIngestUsecase(store, emb, rtLogger{}, config.Config{})
			if err := uc.Execute(context.Background(), &domain.Document{ID: 1}, nil, makeChunks(70), noStep); err == nil {
				t.Fatal("expected an error")
			}
			if store.upserted != 0 || store.deletes != 0 {
				t.Errorf("upserted=%d deletes=%d, want 0/0", store.upserted, store.deletes)
			}
		})
	}
}
