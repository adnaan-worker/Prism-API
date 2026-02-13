package apiconfig

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// StringArray 瀛楃涓叉暟缁勭被鍨嬶紙瀛樺偍涓?JSON锛?
type StringArray []string

func (s StringArray) Value() (driver.Value, error) {
	if s == nil {
		return json.Marshal([]string{})
	}
	return json.Marshal(s)
}

func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		*s = []string{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, s)
}

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

// 閰嶇疆绫诲瀷甯搁噺
const (
	ConfigTypeDirect      = "direct"       // 鐩存帴璋冪敤绗笁鏂?API
	ConfigTypeAccountPool = "account_pool" // 浣跨敤璐﹀彿姹?
)

// APIConfig API閰嶇疆妯″瀷
type APIConfig struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	Name      string         `gorm:"not null;size:255" json:"name"`
	Type      string         `gorm:"not null;size:50" json:"type"` // openai, anthropic, gemini, kiro
	
	// 閰嶇疆绫诲瀷
	ConfigType string `gorm:"not null;size:50;default:'direct'" json:"config_type"` // direct, account_pool
	
	// 鐩存帴璋冪敤閰嶇疆
	BaseURL string `gorm:"type:text" json:"base_url,omitempty"`
	APIKey  string `gorm:"type:text" json:"api_key,omitempty"`
	
	// 璐﹀彿姹犻厤缃?
	AccountPoolID *uint `gorm:"index" json:"account_pool_id,omitempty"`
	
	Models   StringArray `gorm:"type:jsonb;not null;default:'[]'" json:"models"`
	Headers  JSONMap     `gorm:"type:jsonb" json:"headers,omitempty"`
	Metadata JSONMap     `gorm:"type:jsonb" json:"metadata,omitempty"`
	IsActive bool        `gorm:"not null;default:true" json:"is_active"`
	Priority int         `gorm:"not null;default:100" json:"priority"`
	Weight   int         `gorm:"not null;default:1" json:"weight"`
	MaxRPS   int         `gorm:"not null;default:0" json:"max_rps"`
	Timeout  int         `gorm:"not null;default:30" json:"timeout"`
}

// TableName 鎸囧畾琛ㄥ悕
func (APIConfig) TableName() string {
	return "api_configs"
}

// IsValid 妫€鏌ラ厤缃槸鍚︽湁鏁?
func (c *APIConfig) IsValid() bool {
	return c.IsActive
}

// Activate 婵€娲婚厤缃?
func (c *APIConfig) Activate() {
	c.IsActive = true
}

// Deactivate 鍋滅敤閰嶇疆
func (c *APIConfig) Deactivate() {
	c.IsActive = false
}

// HasModel 妫€鏌ユ槸鍚︽敮鎸佹寚瀹氭ā鍨?
func (c *APIConfig) HasModel(model string) bool {
	for _, m := range c.Models {
		if m == model {
			return true
		}
	}
	return false
}

// AddModel 娣诲姞妯″瀷
func (c *APIConfig) AddModel(model string) {
	if !c.HasModel(model) {
		c.Models = append(c.Models, model)
	}
}

// RemoveModel 绉婚櫎妯″瀷
func (c *APIConfig) RemoveModel(model string) {
	newModels := make([]string, 0, len(c.Models))
	for _, m := range c.Models {
		if m != model {
			newModels = append(newModels, m)
		}
	}
	c.Models = newModels
}

// SetHeader 璁剧疆璇锋眰澶?
func (c *APIConfig) SetHeader(key string, value interface{}) {
	if c.Headers == nil {
		c.Headers = make(JSONMap)
	}
	c.Headers[key] = value
}

// GetHeader 鑾峰彇璇锋眰澶?
func (c *APIConfig) GetHeader(key string) (interface{}, bool) {
	if c.Headers == nil {
		return nil, false
	}
	val, ok := c.Headers[key]
	return val, ok
}

// SetMetadata 璁剧疆鍏冩暟鎹?
func (c *APIConfig) SetMetadata(key string, value interface{}) {
	if c.Metadata == nil {
		c.Metadata = make(JSONMap)
	}
	c.Metadata[key] = value
}

// GetMetadata 鑾峰彇鍏冩暟鎹?
func (c *APIConfig) GetMetadata(key string) (interface{}, bool) {
	if c.Metadata == nil {
		return nil, false
	}
	val, ok := c.Metadata[key]
	return val, ok
}

// GetType 鑾峰彇绫诲瀷锛堝疄鐜?adapter.APIConfigInterface锛?
func (c *APIConfig) GetType() string {
	return c.Type
}

// GetBaseURL 鑾峰彇 BaseURL锛堝疄鐜?adapter.APIConfigInterface锛?
func (c *APIConfig) GetBaseURL() string {
	return c.BaseURL
}

// GetAPIKey 鑾峰彇 APIKey锛堝疄鐜?adapter.APIConfigInterface锛?
func (c *APIConfig) GetAPIKey() string {
	return c.APIKey
}

// GetTimeout 鑾峰彇瓒呮椂鏃堕棿锛堝疄鐜?adapter.APIConfigInterface锛?
func (c *APIConfig) GetTimeout() int {
	return c.Timeout
}

// IsDirect 鏄惁鏄洿鎺ヨ皟鐢?
func (c *APIConfig) IsDirect() bool {
	return c.ConfigType == ConfigTypeDirect
}

// IsAccountPool 鏄惁浣跨敤璐﹀彿姹?
func (c *APIConfig) IsAccountPool() bool {
	return c.ConfigType == ConfigTypeAccountPool
}
