package domain

import "time"

type Document struct {
	ID          string     `gorm:"primaryKey" json:"id"`
	UserID      string     `gorm:"not null" json:"user_id"`
	Filename    string     `gorm:"not null" json:"filename"`
	FilePath    string     `gorm:"" json:"file_path"`
	ContentType string     `json:"content_type"`
	Status      string     `json:"status" gorm:"default='pending'"`
	Size        int64      `gorm:"" json:"size"`
	CreatedAt   time.Time  `gorm:"not null" json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at"`
}
