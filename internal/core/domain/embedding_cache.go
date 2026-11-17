package domain

import "time"

// EmbeddingCacheEntry caches the embedding of one exact chunk text for a given
// model and dimension. Vector holds little-endian float32 values.
type EmbeddingCacheEntry struct {
	ID        int       `gorm:"primaryKey" json:"id"`
	TextHash  string    `gorm:"uniqueIndex:idx_embedding_cache_key;size:64;not null" json:"text_hash"`
	Model     string    `gorm:"uniqueIndex:idx_embedding_cache_key;not null" json:"model"`
	Dimension int       `gorm:"uniqueIndex:idx_embedding_cache_key;not null" json:"dimension"`
	Vector    []byte    `gorm:"type:bytea;not null" json:"-"`
	CreatedAt time.Time `json:"created_at"`
}
