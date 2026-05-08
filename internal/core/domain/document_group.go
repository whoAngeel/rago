package domain

import (
	"crypto/rand"
	"time"
)

type DocumentGroup struct {
	ID             int       `gorm:"primaryKey" json:"id"`
	UserID         int       `gorm:"index;not null" json:"user_id"`
	Name           string    `gorm:"not null;size:255" json:"name"`
	Slug           string    `gorm:"uniqueIndex;not null;size:16" json:"slug"`
	IsActive       bool      `gorm:"default:true" json:"is_active"`
	AllowDownloads bool      `gorm:"default:true" json:"allow_downloads"`
	ChatQuota      int       `gorm:"default:100" json:"chat_quota"`
	ChatQuotaUsed  int       `gorm:"default:0" json:"chat_quota_used"`
	ChatAttempts   int       `gorm:"default:0" json:"chat_attempts"`
	DocumentIDs    []int     `gorm:"-" json:"document_ids"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type DocumentGroupItem struct {
	GroupID    int `gorm:"primaryKey" json:"group_id"`
	DocumentID int `gorm:"primaryKey" json:"document_id"`
}

const slugChars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func GenerateSlug() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i, v := range b {
		b[i] = slugChars[int(v)%62]
	}
	return string(b), nil
}
