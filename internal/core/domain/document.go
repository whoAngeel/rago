package domain

import "time"

type Document struct {
	ID          int        `gorm:"primaryKey" json:"id"`
	UserID      int        `gorm:"index:idx_documents_user_id;not null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE,ForeignKey:UserIDReferences:ID" json:"user_id"`
	Filename    string     `gorm:"not null" json:"filename"`
	FilePath    string     `json:"file_path"`
	ContentType string     `json:"content_type"`
	Status      string     `json:"status" gorm:"default:'pending'"`
	Size        int64      `json:"size"`
	CreatedAt   time.Time  `gorm:"not null" json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at" gorm:"index"`
}
