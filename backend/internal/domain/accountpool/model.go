package accountpool

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// JSONMap JSON 瀵硅薄绫诲瀷
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

// AccountPool 璐﹀彿姹犳ā鍨?
type AccountPool struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 鍩烘湰淇℃伅
	Name        string `gorm:"not null;size:255" json:"name"`
	Description string `gorm:"type:text" json:"description,omitempty"`

	// 鎻愪緵鍟嗙被鍨?
	Provider string `gorm:"column:provider_type;not null;size:50" json:"provider"`

	// 杞绛栫暐
	Strategy string `gorm:"not null;size:50;default:'round_robin'" json:"strategy"`

	// 鍋ュ悍妫€鏌ラ厤缃?
	HealthCheckInterval int `gorm:"not null;default:300" json:"health_check_interval"` // 绉?
	HealthCheckTimeout  int `gorm:"not null;default:10" json:"health_check_timeout"`   // 绉?
	MaxRetries          int `gorm:"not null;default:3" json:"max_retries"`

	// 鐘舵€?
	IsActive bool `gorm:"column:is_active;not null;default:true" json:"is_active"`

	// 缁熻
	TotalRequests int64 `gorm:"not null;default:0" json:"total_requests"`
	TotalErrors   int64 `gorm:"not null;default:0" json:"total_errors"`
}

// TableName 鎸囧畾琛ㄥ悕
func (AccountPool) TableName() string {
	return "account_pools"
}

// Activate 婵€娲昏处鍙锋睜
func (p *AccountPool) Activate() {
	p.IsActive = true
}

// Deactivate 鍋滅敤璐﹀彿姹?
func (p *AccountPool) Deactivate() {
	p.IsActive = false
}

// IncrementRequests 澧炲姞璇锋眰璁℃暟
func (p *AccountPool) IncrementRequests() {
	p.TotalRequests++
}

// IncrementErrors 澧炲姞閿欒璁℃暟
func (p *AccountPool) IncrementErrors() {
	p.TotalErrors++
}

// GetErrorRate 鑾峰彇閿欒鐜?
func (p *AccountPool) GetErrorRate() float64 {
	if p.TotalRequests == 0 {
		return 0
	}
	return float64(p.TotalErrors) / float64(p.TotalRequests)
}

// IsHealthy 妫€鏌ユ槸鍚﹀仴搴?
func (p *AccountPool) IsHealthy() bool {
	return p.IsActive && p.GetErrorRate() < 0.5
}

// AccountPoolRequestLog 璐﹀彿姹犺姹傛棩蹇楁ā鍨?
type AccountPoolRequestLog struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	CredentialID *uint  `gorm:"index" json:"credential_id,omitempty"`
	PoolID       *uint  `gorm:"index" json:"pool_id,omitempty"`
	Provider     string `gorm:"column:provider_type;not null;size:50" json:"provider"`

	// 璇锋眰淇℃伅
	Model      string `gorm:"not null;size:255" json:"model"`
	Method     string `gorm:"not null;size:10" json:"method"`
	StatusCode int    `json:"status_code,omitempty"`

	// 鎬ц兘
	ResponseTime int `json:"response_time,omitempty"` // 姣
	TokensUsed   int `json:"tokens_used,omitempty"`

	// 閿欒淇℃伅
	ErrorMessage string `gorm:"type:text" json:"error_message,omitempty"`

	// 鍏宠仈涓昏姹傛棩蹇?
	RequestLogID *uint `gorm:"index" json:"request_log_id,omitempty"`
}

// TableName 鎸囧畾琛ㄥ悕
func (AccountPoolRequestLog) TableName() string {
	return "account_pool_request_logs"
}

// IsSuccess 妫€鏌ヨ姹傛槸鍚︽垚鍔?
func (l *AccountPoolRequestLog) IsSuccess() bool {
	return l.StatusCode >= 200 && l.StatusCode < 300
}

// IsError 妫€鏌ヨ姹傛槸鍚﹀け璐?
func (l *AccountPoolRequestLog) IsError() bool {
	return l.StatusCode >= 400 || l.ErrorMessage != ""
}

// 璐﹀彿姹犵瓥鐣ュ父閲?
const (
	StrategyRoundRobin         = "round_robin"
	StrategyWeightedRoundRobin = "weighted_round_robin"
	StrategyLeastConnections   = "least_connections"
	StrategyRandom             = "random"
)

// ValidStrategies 鏈夋晥鐨勮处鍙锋睜绛栫暐鍒楄〃
var ValidStrategies = []string{
	StrategyRoundRobin,
	StrategyWeightedRoundRobin,
	StrategyLeastConnections,
	StrategyRandom,
}

// IsValidStrategy 妫€鏌ョ瓥鐣ユ槸鍚︽湁鏁?
func IsValidStrategy(strategy string) bool {
	for _, s := range ValidStrategies {
		if s == strategy {
			return true
		}
	}
	return false
}

// AccountCredential 璐﹀彿鍑嵁妯″瀷
type AccountCredential struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 鍏宠仈璐﹀彿姹?
	PoolID uint `gorm:"not null;index" json:"pool_id"`

	// 鎻愪緵鍟嗙被鍨?
	Provider string `gorm:"column:provider_type;not null;size:50" json:"provider"`

	// 璁よ瘉绫诲瀷
	AuthType string `gorm:"not null;size:50;default:'api_key'" json:"auth_type"` // api_key, oauth

	// 鍑嵁淇℃伅锛堝姞瀵嗗瓨鍌級
	APIKey       string `gorm:"type:text" json:"api_key,omitempty"`
	AccessToken  string `gorm:"type:text" json:"access_token,omitempty"`
	RefreshToken string `gorm:"type:text" json:"refresh_token,omitempty"`

	// OAuth 鐩稿叧
	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	// 鎵╁睍淇℃伅锛圝SON 瀛樺偍锛屼笉鍚屾彁渚涘晢鍙互瀛樺偍涓嶅悓鐨勬暟鎹級
	Metadata JSONMap `gorm:"type:jsonb;default:'{}'" json:"metadata,omitempty"`

	// 鏉冮噸锛堢敤浜庡姞鏉冭疆璇級
	Weight int `gorm:"not null;default:1" json:"weight"`

	// 鐘舵€?
	IsActive bool `gorm:"column:is_active;not null;default:true" json:"is_active"`

	// 鍋ュ悍鐘舵€?
	HealthStatus string     `gorm:"size:50;default:'unknown'" json:"health_status"` // healthy, unhealthy, unknown
	LastError    string     `gorm:"type:text" json:"last_error,omitempty"`
	LastUsedAt   *time.Time `json:"last_used_at,omitempty"`

	// 缁熻
	TotalRequests int64 `gorm:"not null;default:0" json:"total_requests"`
	TotalErrors   int64 `gorm:"not null;default:0" json:"total_errors"`

	// 閫熺巼闄愬埗
	RateLimit        int        `gorm:"not null;default:0" json:"rate_limit"`         // 姣忓垎閽熻姹傛暟锛?琛ㄧず鏃犻檺鍒?
	CurrentUsage     int        `gorm:"not null;default:0" json:"current_usage"`      // 褰撳墠鍒嗛挓浣跨敤閲?
	RateLimitResetAt *time.Time `json:"rate_limit_reset_at,omitempty"`
}

// TableName 鎸囧畾琛ㄥ悕
func (AccountCredential) TableName() string {
	return "account_credentials"
}

// Activate 婵€娲诲嚟鎹?
func (c *AccountCredential) Activate() {
	c.IsActive = true
}

// Deactivate 鍋滅敤鍑嵁
func (c *AccountCredential) Deactivate() {
	c.IsActive = false
}

// IsExpired 妫€鏌ユ槸鍚﹁繃鏈?
func (c *AccountCredential) IsExpired() bool {
	if c.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*c.ExpiresAt)
}

// IsHealthy 妫€鏌ユ槸鍚﹀仴搴?
func (c *AccountCredential) IsHealthy() bool {
	// 鍏佽 unknown 鐘舵€佺殑鍑嵁锛堟柊瀵煎叆鐨勫嚟鎹級
	// 鍙湁鏄庣‘鏍囪涓?unhealthy 鐨勬墠鎷掔粷
	return c.IsActive && c.HealthStatus != "unhealthy" && !c.IsExpired()
}

// IncrementRequests 澧炲姞璇锋眰璁℃暟
func (c *AccountCredential) IncrementRequests() {
	c.TotalRequests++
	now := time.Now()
	c.LastUsedAt = &now
}

// IncrementErrors 澧炲姞閿欒璁℃暟
func (c *AccountCredential) IncrementErrors() {
	c.TotalErrors++
}

// GetErrorRate 鑾峰彇閿欒鐜?
func (c *AccountCredential) GetErrorRate() float64 {
	if c.TotalRequests == 0 {
		return 0
	}
	return float64(c.TotalErrors) / float64(c.TotalRequests)
}

// UpdateHealthStatus 鏇存柊鍋ュ悍鐘舵€?
func (c *AccountCredential) UpdateHealthStatus(status string) {
	c.HealthStatus = status
}

// IsRateLimited 妫€鏌ユ槸鍚﹁揪鍒伴€熺巼闄愬埗
func (c *AccountCredential) IsRateLimited() bool {
	if c.RateLimit == 0 {
		return false
	}
	
	// 妫€鏌ユ槸鍚﹂渶瑕侀噸缃?
	if c.RateLimitResetAt != nil && time.Now().After(*c.RateLimitResetAt) {
		return false
	}
	
	return c.CurrentUsage >= c.RateLimit
}

// IncrementUsage 澧炲姞浣跨敤閲?
func (c *AccountCredential) IncrementUsage() {
	// 濡傛灉闇€瑕侀噸缃?
	if c.RateLimitResetAt == nil || time.Now().After(*c.RateLimitResetAt) {
		c.CurrentUsage = 1
		resetAt := time.Now().Add(time.Minute)
		c.RateLimitResetAt = &resetAt
	} else {
		c.CurrentUsage++
	}
}

// 璁よ瘉绫诲瀷甯搁噺
const (
	AuthTypeAPIKey = "api_key"
	AuthTypeOAuth  = "oauth"
)

// ValidAuthTypes 鏈夋晥鐨勮璇佺被鍨嬪垪琛?
var ValidAuthTypes = []string{
	AuthTypeAPIKey,
	AuthTypeOAuth,
}

// IsValidAuthType 妫€鏌ヨ璇佺被鍨嬫槸鍚︽湁鏁?
func IsValidAuthType(authType string) bool {
	for _, t := range ValidAuthTypes {
		if t == authType {
			return true
		}
	}
	return false
}

// 鍋ュ悍鐘舵€佸父閲?
const (
	HealthStatusHealthy   = "healthy"
	HealthStatusUnhealthy = "unhealthy"
	HealthStatusUnknown   = "unknown"
)
