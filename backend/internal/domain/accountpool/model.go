package accountpool

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// JSONMap JSON 对象类型
type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return json.Marshal(map[string]interface{}{})
	}
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = map[string]interface{}{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

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

// AccountCredential 账号凭据模型
type AccountCredential struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// 关联账号池
	PoolID uint `gorm:"not null;index" json:"pool_id"`

	// 提供商类型
	Provider string `gorm:"column:provider_type;not null;size:50" json:"provider"`

	// 认证类型
	AuthType string `gorm:"not null;size:50;default:'api_key'" json:"auth_type"` // api_key, oauth

	// 凭据信息（加密存储）
	APIKey       string `gorm:"type:text" json:"api_key,omitempty"`
	AccessToken  string `gorm:"type:text" json:"access_token,omitempty"`
	RefreshToken string `gorm:"type:text" json:"refresh_token,omitempty"`

	// OAuth 相关
	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	// 扩展信息（JSON 存储，不同提供商可以存储不同的数据）
	Metadata JSONMap `gorm:"type:jsonb;default:'{}'" json:"metadata,omitempty"`

	// 权重（用于加权轮询）
	Weight int `gorm:"not null;default:1" json:"weight"`

	// 状态
	IsActive bool `gorm:"column:is_active;not null;default:true" json:"is_active"`

	// 健康状态
	HealthStatus string     `gorm:"size:50;default:'unknown'" json:"health_status"` // healthy, unhealthy, unknown
	LastError    string     `gorm:"type:text" json:"last_error,omitempty"`
	LastUsedAt   *time.Time `json:"last_used_at,omitempty"`

	// 统计
	TotalRequests int64 `gorm:"not null;default:0" json:"total_requests"`
	TotalErrors   int64 `gorm:"not null;default:0" json:"total_errors"`

	// 速率限制
	RateLimit        int        `gorm:"not null;default:0" json:"rate_limit"`         // 每分钟请求数，0表示无限制
	CurrentUsage     int        `gorm:"not null;default:0" json:"current_usage"`      // 当前分钟使用量
	RateLimitResetAt *time.Time `json:"rate_limit_reset_at,omitempty"`
}

// TableName 指定表名
func (AccountCredential) TableName() string {
	return "account_credentials"
}

// Activate 激活凭据
func (c *AccountCredential) Activate() {
	c.IsActive = true
}

// Deactivate 停用凭据
func (c *AccountCredential) Deactivate() {
	c.IsActive = false
}

// IsExpired 检查是否过期
func (c *AccountCredential) IsExpired() bool {
	if c.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*c.ExpiresAt)
}

// IsHealthy 检查是否健康
func (c *AccountCredential) IsHealthy() bool {
	// 允许 unknown 状态的凭据（新导入的凭据）
	// 只有明确标记为 unhealthy 的才拒绝
	return c.IsActive && c.HealthStatus != "unhealthy" && !c.IsExpired()
}

// IncrementRequests 增加请求计数
func (c *AccountCredential) IncrementRequests() {
	c.TotalRequests++
	now := time.Now()
	c.LastUsedAt = &now
}

// IncrementErrors 增加错误计数
func (c *AccountCredential) IncrementErrors() {
	c.TotalErrors++
}

// GetErrorRate 获取错误率
func (c *AccountCredential) GetErrorRate() float64 {
	if c.TotalRequests == 0 {
		return 0
	}
	return float64(c.TotalErrors) / float64(c.TotalRequests)
}

// UpdateHealthStatus 更新健康状态
func (c *AccountCredential) UpdateHealthStatus(status string) {
	c.HealthStatus = status
}

// IsRateLimited 检查是否达到速率限制
func (c *AccountCredential) IsRateLimited() bool {
	if c.RateLimit == 0 {
		return false
	}
	
	// 检查是否需要重置
	if c.RateLimitResetAt != nil && time.Now().After(*c.RateLimitResetAt) {
		return false
	}
	
	return c.CurrentUsage >= c.RateLimit
}

// IncrementUsage 增加使用量
func (c *AccountCredential) IncrementUsage() {
	// 如果需要重置
	if c.RateLimitResetAt == nil || time.Now().After(*c.RateLimitResetAt) {
		c.CurrentUsage = 1
		resetAt := time.Now().Add(time.Minute)
		c.RateLimitResetAt = &resetAt
	} else {
		c.CurrentUsage++
	}
}

// 认证类型常量
const (
	AuthTypeAPIKey = "api_key"
	AuthTypeOAuth  = "oauth"
)

// ValidAuthTypes 有效的认证类型列表
var ValidAuthTypes = []string{
	AuthTypeAPIKey,
	AuthTypeOAuth,
}

// IsValidAuthType 检查认证类型是否有效
func IsValidAuthType(authType string) bool {
	for _, t := range ValidAuthTypes {
		if t == authType {
			return true
		}
	}
	return false
}

// 健康状态常量
const (
	HealthStatusHealthy   = "healthy"
	HealthStatusUnhealthy = "unhealthy"
	HealthStatusUnknown   = "unknown"
)
