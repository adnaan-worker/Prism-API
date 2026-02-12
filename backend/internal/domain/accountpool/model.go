package accountpool

import (
	"time"

	"gorm.io/gorm"
)

// AccountPool 账号池模型
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

// TableName 指定表名
func (AccountPool) TableName() string {
	return "account_pools"
}

// Activate 激活账号池
func (p *AccountPool) Activate() {
	p.IsActive = true
}

// Deactivate 停用账号池
func (p *AccountPool) Deactivate() {
	p.IsActive = false
}

// IncrementRequests 增加请求计数
func (p *AccountPool) IncrementRequests() {
	p.TotalRequests++
}

// IncrementErrors 增加错误计数
func (p *AccountPool) IncrementErrors() {
	p.TotalErrors++
}

// GetErrorRate 获取错误率
func (p *AccountPool) GetErrorRate() float64 {
	if p.TotalRequests == 0 {
		return 0
	}
	return float64(p.TotalErrors) / float64(p.TotalRequests)
}

// IsHealthy 检查是否健康
func (p *AccountPool) IsHealthy() bool {
	return p.IsActive && p.GetErrorRate() < 0.5
}

// AccountPoolRequestLog 账号池请求日志模型
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

// TableName 指定表名
func (AccountPoolRequestLog) TableName() string {
	return "account_pool_request_logs"
}

// IsSuccess 检查请求是否成功
func (l *AccountPoolRequestLog) IsSuccess() bool {
	return l.StatusCode >= 200 && l.StatusCode < 300
}

// IsError 检查请求是否失败
func (l *AccountPoolRequestLog) IsError() bool {
	return l.StatusCode >= 400 || l.ErrorMessage != ""
}

// 账号池策略常量
const (
	StrategyRoundRobin         = "round_robin"
	StrategyWeightedRoundRobin = "weighted_round_robin"
	StrategyLeastConnections   = "least_connections"
	StrategyRandom             = "random"
)

// ValidStrategies 有效的账号池策略列表
var ValidStrategies = []string{
	StrategyRoundRobin,
	StrategyWeightedRoundRobin,
	StrategyLeastConnections,
	StrategyRandom,
}

// IsValidStrategy 检查策略是否有效
func IsValidStrategy(strategy string) bool {
	for _, s := range ValidStrategies {
		if s == strategy {
			return true
		}
	}
	return false
}
