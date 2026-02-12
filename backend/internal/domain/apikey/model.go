package apikey

import (
	"time"

	"gorm.io/gorm"
)

// APIKey API密钥模型
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
}

// TableName 指定表名
func (APIKey) TableName() string {
	return "api_keys"
}

// IsValid 检查密钥是否有效
func (k *APIKey) IsValid() bool {
	return k.IsActive && k.DeletedAt.Time.IsZero()
}

// Activate 激活密钥
func (k *APIKey) Activate() {
	k.IsActive = true
}

// Deactivate 停用密钥
func (k *APIKey) Deactivate() {
	k.IsActive = false
}

// UpdateLastUsed 更新最后使用时间
func (k *APIKey) UpdateLastUsed() {
	now := time.Now()
	k.LastUsedAt = &now
}
