package models

import (
	"time"

	"gorm.io/gorm"
)

// AccountPool 账号池
type AccountPool struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// 基本信息
	Name        string `gorm:"not null;size:255" json:"name"`
	Description string `gorm:"type:text" json:"description,omitempty"`

	// 提供商类型
	Provider string `gorm:"column:provider_type;not null;size:50" json:"provider"`

	// 轮询策略
	Strategy string `gorm:"not null;size:50;default:'round_robin'" json:"strategy"`

	// 健康检查配置
	HealthCheckInterval int `gorm:"not null;default:300" json:"health_check_interval"` // 秒
	HealthCheckTimeout  int `gorm:"not null;default:10" json:"health_check_timeout"`   // 秒
	MaxRetries          int `gorm:"not null;default:3" json:"max_retries"`

	// 状态
	IsActive bool `gorm:"column:is_active;not null;default:true" json:"is_active"`

	// 统计
	TotalRequests int64 `gorm:"not null;default:0" json:"total_requests"`
	TotalErrors   int64 `gorm:"not null;default:0" json:"total_errors"`
}

func (AccountPool) TableName() string {
	return "account_pools"
}
