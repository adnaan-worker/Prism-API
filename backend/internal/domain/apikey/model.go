package apikey

import (
	"time"
)

// APIKey API瀵嗛挜妯″瀷
type APIKey struct {
	ID         uint           `gorm:"primarykey" json:"id"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	UserID     uint           `gorm:"not null;index" json:"user_id"`
	Key        string         `gorm:"uniqueIndex;not null;size:255" json:"key"`
	Name       string         `gorm:"not null;size:255" json:"name"`
	IsActive   bool           `gorm:"not null;default:true" json:"is_active"`
	RateLimit  int            `gorm:"not null;default:60" json:"rate_limit"`
	LastUsedAt *time.Time     `json:"last_used_at,omitempty"`
}

// TableName 鎸囧畾琛ㄥ悕
func (APIKey) TableName() string {
	return "api_keys"
}

// IsValid 妫€鏌ュ瘑閽ユ槸鍚︽湁鏁?
func (k *APIKey) IsValid() bool {
	return k.IsActive
}

// Activate 婵€娲诲瘑閽?
func (k *APIKey) Activate() {
	k.IsActive = true
}

// Deactivate 鍋滅敤瀵嗛挜
func (k *APIKey) Deactivate() {
	k.IsActive = false
}

// UpdateLastUsed 鏇存柊鏈€鍚庝娇鐢ㄦ椂闂?
func (k *APIKey) UpdateLastUsed() {
	now := time.Now()
	k.LastUsedAt = &now
}
