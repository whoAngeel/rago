package domain

import "time"

type ProcessingStep struct {
	ID           int       `gorm:"primaryKey" json:"id"`
	DocumentID   int       `gorm:"index;not null;constraint:OnUpdate:CASCADE,onDelete:CASCADE" json:"document_id"`
	StepName     string    `gorm:"size:50;not null" json:"step_name"`
	Status       string    `gorm:"size:20;not null" json:"status"`
	ErrorMessage *string   `gorm:"type:text" json:"error_message"`
	DurationMS   *int      `json:"duration_ms"`
	CreatedAt    time.Time `gorm:"not null" json:"created_at"`
}
