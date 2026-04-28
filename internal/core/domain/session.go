package domain

import "time"

type Session struct {
	ID           int        `json:"id" gorm:"primaryKey"`
	UserID       int        `json:"user_id" gorm:"index;not null"`
	RefreshToken string     `json:"refresh_token" gorm:"uniqueIndex;not null"`
	AccessToken  string     `json:"access_token"`
	ExpiresAt    time.Time  `json:"expires_at" gorm:"not null"`
	RevokedAt    *time.Time `json:"revoked_at,omitempty"`
}
