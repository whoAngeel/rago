package domain

import "gorm.io/gorm"

type SystemConfig struct {
	gorm.Model
	Key   string `gorm:"uniqueIndex;size:50;not null" json:"key"`
	Value string `gorm:"type:text;not null" json:"value"`
}
