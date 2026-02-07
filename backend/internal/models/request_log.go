package models

import (
	"time"

	"gorm.io/gorm"
)

type RequestLog struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	UserID       uint           `gorm:"not null;index" json:"user_id"`
	APIKeyID     uint           `gorm:"not null;index" json:"api_key_id"`
	APIConfigID  uint           `gorm:"not null;index" json:"api_config_id"`
	Model        string         `gorm:"not null;size:255;index" json:"model"`
	Method       string         `gorm:"not null;size:10" json:"method"`
	Path         string         `gorm:"not null;type:text" json:"path"`
	StatusCode   int            `gorm:"not null;index" json:"status_code"`
	ResponseTime int            `gorm:"not null" json:"response_time"`
	TokensUsed   int            `gorm:"not null;default:0" json:"tokens_used"`
	ErrorMsg     string         `gorm:"type:text" json:"error_msg,omitempty"`
	User         User           `gorm:"foreignKey:UserID" json:"-"`
	APIKey       APIKey         `gorm:"foreignKey:APIKeyID" json:"-"`
}

func (RequestLog) TableName() string {
	return "request_logs"
}
