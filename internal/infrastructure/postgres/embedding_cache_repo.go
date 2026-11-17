package postgres

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math"

	"github.com/whoAngeel/rago/internal/core/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type EmbeddingCacheRepository struct {
	db *gorm.DB
}

func NewEmbeddingCacheRepository(db *gorm.DB) *EmbeddingCacheRepository {
	return &EmbeddingCacheRepository{db: db}
}

func (r *EmbeddingCacheRepository) Get(ctx context.Context, textHash, model string, dim int) ([]float32, bool, error) {
	var row domain.EmbeddingCacheEntry
	err := r.db.WithContext(ctx).
		Where("text_hash = ? AND model = ? AND dimension = ?", textHash, model, dim).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	vec, err := decodeVector(row.Vector)
	if err != nil {
		return nil, false, err
	}
	return vec, true, nil
}

func (r *EmbeddingCacheRepository) Put(ctx context.Context, textHash, model string, dim int, vector []float32) error {
	row := domain.EmbeddingCacheEntry{TextHash: textHash, Model: model, Dimension: dim, Vector: encodeVector(vector)}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&row).Error
}

// encodeVector serialises v as little-endian float32 values.
func encodeVector(v []float32) []byte {
	out := make([]byte, 4*len(v))
	for i, f := range v {
		binary.LittleEndian.PutUint32(out[4*i:], math.Float32bits(f))
	}
	return out
}

func decodeVector(b []byte) ([]float32, error) {
	if len(b)%4 != 0 {
		return nil, fmt.Errorf("invalid vector blob length %d", len(b))
	}
	out := make([]float32, len(b)/4)
	for i := range out {
		out[i] = math.Float32frombits(binary.LittleEndian.Uint32(b[4*i:]))
	}
	return out, nil
}
