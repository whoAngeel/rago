package domain

import "time"

type LLMUsage struct {
	ID           int       `gorm:"primaryKey" json:"id"`
	UserID       int       `gorm:"not null;index" json:"user_id"`
	GroupID      *int      `gorm:"index" json:"group_id,omitempty"`
	Model        string    `gorm:"not null;size:100" json:"model"`
	InputTokens  int       `gorm:"not null;default:0" json:"input_tokens"`
	OutputTokens int       `gorm:"not null;default:0" json:"output_tokens"`
	CreatedAt    time.Time `json:"created_at"`
}
