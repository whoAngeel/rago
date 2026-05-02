package domain

import "database/sql/driver"

type DocumentStatus string

const (
	StatusUploading  DocumentStatus = "uploading"
	StatusPending    DocumentStatus = "pending"
	StatusProcessing DocumentStatus = "processing"
	StatusCompleted  DocumentStatus = "completed"
	StatusFailed     DocumentStatus = "failed"
)

func (s DocumentStatus) String() string {
	return string(s)
}

func (s DocumentStatus) Valid() bool {
	switch s {
	case StatusUploading, StatusPending, StatusProcessing, StatusCompleted, StatusFailed:
		return true
	}
	return false
}

func (s *DocumentStatus) Scan(value interface{}) error {
	if value == nil {
		*s = StatusPending
		return nil
	}
	*s = DocumentStatus(value.(string))
	return nil
}

func (s DocumentStatus) Value() (driver.Value, error) {
	return string(s), nil
}
