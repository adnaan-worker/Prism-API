package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Username     string         `gorm:"uniqueIndex;not null;size:255" json:"username"`
	Email        string         `gorm:"uniqueIndex;not null;size:255" json:"email"`
	PasswordHash string         `gorm:"not null;size:255" json:"-"`
	Quota        int64          `gorm:"not null;default:10000" json:"quota"`
	UsedQuota    int64          `gorm:"not null;default:0" json:"used_quota"`
	IsAdmin      bool           `gorm:"not null;default:false" json:"is_admin"`
	Status       string         `gorm:"not null;default:'active';size:50" json:"status"`
	LastSignIn   *time.Time     `json:"last_sign_in,omitempty"`
}

func (User) TableName() string {
	return "users"
}
