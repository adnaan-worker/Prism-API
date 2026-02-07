package models

import (
	"time"

	"gorm.io/gorm"
)

type APIKey struct {
	ID         uint           `gorm:"primarykey" json:"id"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	UserID     uint           `gorm:"not null;index" json:"user_id"`
	Key        string         `gorm:"uniqueIndex;not null;size:255" json:"key"`
	Name       string         `gorm:"not null;size:255" json:"name"`
	IsActive   bool           `gorm:"not null;default:true" json:"is_active"`
	RateLimit  int            `gorm:"not null;default:60" json:"rate_limit"`
	LastUsedAt *time.Time     `json:"last_used_at,omitempty"`
	User       User           `gorm:"foreignKey:UserID" json:"-"`
}

func (APIKey) TableName() string {
	return "api_keys"
}
