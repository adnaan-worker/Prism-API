package apiconfig

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// StringArray 字符串数组类型（存储为 JSON）
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

// APIConfig API配置模型
type APIConfig struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Name      string         `gorm:"not null;size:255" json:"name"`
	Type      string         `gorm:"not null;size:50" json:"type"`
	BaseURL   string         `gorm:"not null;type:text" json:"base_url"`
	APIKey    string         `gorm:"type:text" json:"api_key,omitempty"`
	Models    StringArray    `gorm:"type:jsonb;not null;default:'[]'" json:"models"`
	Headers   JSONMap        `gorm:"type:jsonb" json:"headers,omitempty"`
	Metadata  JSONMap        `gorm:"type:jsonb" json:"metadata,omitempty"`
	IsActive  bool           `gorm:"not null;default:true" json:"is_active"`
	Priority  int            `gorm:"not null;default:100" json:"priority"`
	Weight    int            `gorm:"not null;default:1" json:"weight"`
	MaxRPS    int            `gorm:"not null;default:0" json:"max_rps"`
	Timeout   int            `gorm:"not null;default:30" json:"timeout"`
}

// TableName 指定表名
func (APIConfig) TableName() string {
	return "api_configs"
}

// IsValid 检查配置是否有效
func (c *APIConfig) IsValid() bool {
	return c.IsActive && c.DeletedAt.Time.IsZero()
}

// Activate 激活配置
func (c *APIConfig) Activate() {
	c.IsActive = true
}

// Deactivate 停用配置
func (c *APIConfig) Deactivate() {
	c.IsActive = false
}

// HasModel 检查是否支持指定模型
func (c *APIConfig) HasModel(model string) bool {
	for _, m := range c.Models {
		if m == model {
			return true
		}
	}
	return false
}

// AddModel 添加模型
func (c *APIConfig) AddModel(model string) {
	if !c.HasModel(model) {
		c.Models = append(c.Models, model)
	}
}

// RemoveModel 移除模型
func (c *APIConfig) RemoveModel(model string) {
	newModels := make([]string, 0, len(c.Models))
	for _, m := range c.Models {
		if m != model {
			newModels = append(newModels, m)
		}
	}
	c.Models = newModels
}

// SetHeader 设置请求头
func (c *APIConfig) SetHeader(key string, value interface{}) {
	if c.Headers == nil {
		c.Headers = make(JSONMap)
	}
	c.Headers[key] = value
}

// GetHeader 获取请求头
func (c *APIConfig) GetHeader(key string) (interface{}, bool) {
	if c.Headers == nil {
		return nil, false
	}
	val, ok := c.Headers[key]
	return val, ok
}

// SetMetadata 设置元数据
func (c *APIConfig) SetMetadata(key string, value interface{}) {
	if c.Metadata == nil {
		c.Metadata = make(JSONMap)
	}
	c.Metadata[key] = value
}

// GetMetadata 获取元数据
func (c *APIConfig) GetMetadata(key string) (interface{}, bool) {
	if c.Metadata == nil {
		return nil, false
	}
	val, ok := c.Metadata[key]
	return val, ok
}
