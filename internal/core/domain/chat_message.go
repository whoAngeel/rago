package domain

import "gorm.io/gorm"

type ChatMessage struct {
	gorm.Model
	SessionID int    `gorm:"index;not null" json:"session_id"`
	Role      string `gorm:"size:20;not null" json:"role"` // "user", "assistant"
	Content   string `gorm:"type:text;not null" json:"content"`
	Sources   string `gorm:"type:jsonb" json:"sources"`
}
