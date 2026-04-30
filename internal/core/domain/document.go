package domain

import "time"

type Document struct {
	ID          int            `gorm:"primaryKey" json:"id"`
	UserID      int            `gorm:"index;not null" json:"user_id"`
	Filename    string         `gorm:"not null" json:"filename"`
	FilePath    string         `json:"file_path"`
	ContentType string         `json:"content_type"`
	Status      DocumentStatus `json:"status" gorm:"default:pending"`
	Size        int64          `json:"size"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}
