package log

import (
	"time"
)

// RequestLog 璇锋眰鏃ュ織妯″瀷
type RequestLog struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
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

// TableName 鎸囧畾琛ㄥ悕
func (RequestLog) TableName() string {
	return "request_logs"
}

// IsSuccess 妫€鏌ヨ姹傛槸鍚︽垚鍔?
func (l *RequestLog) IsSuccess() bool {
	return l.StatusCode >= 200 && l.StatusCode < 300
}

// IsError 妫€鏌ヨ姹傛槸鍚﹂敊璇?
func (l *RequestLog) IsError() bool {
	return l.StatusCode >= 400
}

// IsServerError 妫€鏌ユ槸鍚︽湇鍔″櫒閿欒
func (l *RequestLog) IsServerError() bool {
	return l.StatusCode >= 500
}

// IsClientError 妫€鏌ユ槸鍚﹀鎴风閿欒
func (l *RequestLog) IsClientError() bool {
	return l.StatusCode >= 400 && l.StatusCode < 500
}
