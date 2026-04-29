package domain

import (
	"time"

	"gorm.io/gorm"
)

type Role struct {
	gorm.Model
	Name string `gorm:"uniqueIndex;not null" json:"name"`
}

const (
	RoleAdmin  = "admin"
	RoleEditor = "editor"
	RoleViewer = "viewer"
)

type User struct {
	ID        int       `json:"id"  gorm:"primaryKey"`
	Email     string    `json:"email"  gorm:"uniqueIndex:idx_users_email;not null"`
	Password  string    `json:"-" gorm:"not null"`
	Name      string    `json:"name,omitempty" gorm:"default:''"`
	RoleID    int       `json:"role_id" gorm:"default:3;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT,ForeignKey:RoleIDReferences:ID"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
