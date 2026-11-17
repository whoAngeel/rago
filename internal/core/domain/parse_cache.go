package domain

import "time"

// ParsedContent caches the text extracted from a file, keyed by the SHA-256 of
// its bytes and the signature of the parser configuration that produced it.
// It is shared across users: the key is the content, not the owner.
type ParsedContent struct {
	ID              int    `gorm:"primaryKey" json:"id"`
	ContentHash     string `gorm:"uniqueIndex:idx_parsed_content_hash_sig;size:64;not null" json:"content_hash"`
	ParserSignature string `gorm:"uniqueIndex:idx_parsed_content_hash_sig;not null" json:"parser_signature"`
	// Pages is a JSON array of {"text": string, "metadata": object}.
	Pages     string    `gorm:"type:text;not null" json:"-"`
	CreatedAt time.Time `json:"created_at"`
}
