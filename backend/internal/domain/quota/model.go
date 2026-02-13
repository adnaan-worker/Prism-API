package quota

import (
	"time"
)

// SignInRecord 绛惧埌璁板綍妯″瀷
type SignInRecord struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	UserID       uint           `gorm:"not null;index" json:"user_id"`
	QuotaAwarded int            `gorm:"not null" json:"quota_awarded"`
}

// TableName 鎸囧畾琛ㄥ悕
func (SignInRecord) TableName() string {
	return "sign_in_records"
}

// IsToday 妫€鏌ユ槸鍚︽槸浠婂ぉ鐨勭鍒?
func (r *SignInRecord) IsToday() bool {
	now := time.Now()
	return r.CreatedAt.Year() == now.Year() &&
		r.CreatedAt.Month() == now.Month() &&
		r.CreatedAt.Day() == now.Day()
}

// QuotaUsageRecord 閰嶉浣跨敤璁板綍锛堢敤浜庣粺璁★級
type QuotaUsageRecord struct {
	Date   string `json:"date"`
	Tokens int64  `json:"tokens"`
}

// Constants
const (
	DailySignInQuota = 1000 // 姣忔棩绛惧埌濂栧姳閰嶉
)
