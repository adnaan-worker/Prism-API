package models

import (
	"time"

	"gorm.io/gorm"
)

type SignInRecord struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	UserID       uint           `gorm:"not null;index" json:"user_id"`
	QuotaAwarded int            `gorm:"not null;default:0" json:"quota_awarded"`
	User         User           `gorm:"foreignKey:UserID" json:"-"`
}

func (SignInRecord) TableName() string {
	return "sign_in_records"
}
