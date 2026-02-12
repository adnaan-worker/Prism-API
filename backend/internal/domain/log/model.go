package log

import (
	"time"

	"gorm.io/gorm"
)

// RequestLog 请求日志模型
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
}

// TableName 指定表名
func (RequestLog) TableName() string {
	return "request_logs"
}

// IsSuccess 检查请求是否成功
func (l *RequestLog) IsSuccess() bool {
	return l.StatusCode >= 200 && l.StatusCode < 300
}

// IsError 检查请求是否错误
func (l *RequestLog) IsError() bool {
	return l.StatusCode >= 400
}

// IsServerError 检查是否服务器错误
func (l *RequestLog) IsServerError() bool {
	return l.StatusCode >= 500
}

// IsClientError 检查是否客户端错误
func (l *RequestLog) IsClientError() bool {
	return l.StatusCode >= 400 && l.StatusCode < 500
}
