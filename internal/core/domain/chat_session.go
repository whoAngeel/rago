package domain

import (
	"gorm.io/gorm"
)

type ChatSession struct {
	gorm.Model
	UserID int    `gorm:"index;not null" json:"user_id"`
	Title  string `gorm:"size:255" json:"title"`
}
