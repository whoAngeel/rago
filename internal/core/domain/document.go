package domain

import "time"

type Document struct {
	ID                  int            `gorm:"primaryKey" json:"id"`
	UserID              int            `gorm:"uniqueIndex:idx_documents_user_checksum;index;not null" json:"user_id"`
	Filename            string         `gorm:"not null" json:"filename"`
	FilePath            string         `json:"file_path"`
	ContentType         string         `json:"content_type"`
	Status              DocumentStatus `json:"status" gorm:"default:pending"`
	Size                int64          `json:"size"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	ProcessingStartedAt *time.Time     `json:"processing_started_at"`
	ErrorMessage        string         `json:"error_message" gorm:"type:text"`
	RetryCount          int            `json:"retry_count" gorm:"default:0"`
	ForceOCR            bool           `json:"force_ocr" gorm:"default:false"`
	Checksum            *string        `gorm:"uniqueIndex:idx_documents_user_checksum;size:64" json:"checksum,omitempty"`
}
