package models

import (
	"time"
)

// AccountPoolRequestLog 账号池请求日志
type AccountPoolRequestLog struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	CredentialID *uint  `gorm:"index" json:"credential_id,omitempty"`
	PoolID       *uint  `gorm:"index" json:"pool_id,omitempty"`
	Provider     string `gorm:"column:provider_type;not null;size:50" json:"provider"`

	// 请求信息
	Model      string `gorm:"not null;size:255" json:"model"`
	Method     string `gorm:"not null;size:10" json:"method"`
	StatusCode int    `json:"status_code,omitempty"`

	// 性能
	ResponseTime int `json:"response_time,omitempty"` // 毫秒
	TokensUsed   int `json:"tokens_used,omitempty"`

	// 错误信息
	ErrorMessage string `gorm:"type:text" json:"error_message,omitempty"`

	// 关联主请求日志
	RequestLogID *uint `gorm:"index" json:"request_log_id,omitempty"`
}

func (AccountPoolRequestLog) TableName() string {
	return "account_pool_request_logs"
}
