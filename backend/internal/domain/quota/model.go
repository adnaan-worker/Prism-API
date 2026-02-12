package quota

import (
	"time"

	"gorm.io/gorm"
)

// SignInRecord 签到记录模型
type SignInRecord struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	UserID       uint           `gorm:"not null;index" json:"user_id"`
	QuotaAwarded int            `gorm:"not null" json:"quota_awarded"`
}

// TableName 指定表名
func (SignInRecord) TableName() string {
	return "sign_in_records"
}

// IsToday 检查是否是今天的签到
func (r *SignInRecord) IsToday() bool {
	now := time.Now()
	return r.CreatedAt.Year() == now.Year() &&
		r.CreatedAt.Month() == now.Month() &&
		r.CreatedAt.Day() == now.Day()
}

// QuotaUsageRecord 配额使用记录（用于统计）
type QuotaUsageRecord struct {
	Date   string `json:"date"`
	Tokens int64  `json:"tokens"`
}

// Constants
const (
	DailySignInQuota = 1000 // 每日签到奖励配额
)
