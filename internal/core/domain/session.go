package domain

import "time"

type Session struct {
	ID           int        `json:"id" gorm:"primaryKey"`
	UserID       int        `json:"user_id" gorm:"index:idx_sessions_user_id;not null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE,ForeignKey:UserIDReferences:ID"`
	RefreshToken string     `json:"refresh_token" gorm:"uniqueIndex:idx_sessions_refresh_token;not null"`
	AccessToken  string     `json:"access_token"`
	ExpiresAt    time.Time  `json:"expires_at" gorm:"not null"`
	RevokedAt    *time.Time `json:"revoked_at,omitempty"`
}
