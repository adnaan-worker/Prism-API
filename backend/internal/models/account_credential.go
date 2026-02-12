package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// CredentialsData JSON 格式存储凭据数据
type CredentialsData map[string]interface{}

func (c CredentialsData) Value() (driver.Value, error) {
	if c == nil {
		return json.Marshal(map[string]interface{}{})
	}
	return json.Marshal(c)
}

func (c *CredentialsData) Scan(value interface{}) error {
	if value == nil {
		*c = map[string]interface{}{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}

// AccountCredential 账号凭据
type AccountCredential struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// 基本信息
	Name        string `gorm:"not null;size:255" json:"name"`
	Description string `gorm:"type:text" json:"description,omitempty"`

	// 提供商和认证类型
	Provider string `gorm:"column:provider_type;not null;size:50" json:"provider"`
	AuthType string `gorm:"column:auth_type;not null;size:50" json:"auth_type"`

	// 凭据数据 (JSON)
	CredentialsData CredentialsData `gorm:"type:jsonb;not null" json:"credentials_data"`

	// Token 过期时间
	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	// 状态
	IsActive     bool       `gorm:"column:is_active;not null;default:true" json:"is_active"`
	ErrorMessage string     `gorm:"column:health_check_error;type:text" json:"error_message,omitempty"`
	LastCheckAt  *time.Time `gorm:"column:last_health_check" json:"last_check_at,omitempty"`

	// 使用统计
	RequestCount int64      `gorm:"not null;default:0" json:"request_count"`
	SuccessCount int64      `gorm:"not null;default:0" json:"success_count"`
	ErrorCount   int64      `gorm:"not null;default:0" json:"error_count"`
	LastUsedAt   *time.Time `json:"last_used_at,omitempty"`

	// 配额
	DailyQuota   int64      `gorm:"not null;default:0" json:"daily_quota"`
	DailyUsed    int64      `gorm:"not null;default:0" json:"daily_used"`
	QuotaResetAt *time.Time `json:"quota_reset_at,omitempty"`
	
	// 关联池 ID (非数据库字段，用于创建时指定)
	PoolID uint `gorm:"-" json:"pool_id,omitempty"`
}

func (AccountCredential) TableName() string {
	return "account_credentials"
}

// IsExpired 检查是否过期
func (a *AccountCredential) IsExpired() bool {
	if a.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*a.ExpiresAt)
}

// IsExpiringSoon 检查是否即将过期
func (a *AccountCredential) IsExpiringSoon(duration time.Duration) bool {
	if a.ExpiresAt == nil {
		return false
	}
	return time.Now().Add(duration).After(*a.ExpiresAt)
}

// GetString 获取字符串值（同时支持驼峰和下划线命名）
func (a *AccountCredential) GetString(key string) string {
	// 先尝试原始 key
	if val, ok := a.CredentialsData[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	
	// 如果是下划线命名，尝试驼峰命名
	// refresh_token -> refreshToken
	// access_token -> accessToken
	// client_id -> clientId
	// client_secret -> clientSecret
	// profile_arn -> profileArn
	// auth_method -> authMethod
	camelKey := toCamelCase(key)
	if camelKey != key {
		if val, ok := a.CredentialsData[camelKey]; ok {
			if str, ok := val.(string); ok {
				return str
			}
		}
	}
	
	return ""
}

// toCamelCase 将下划线命名转换为驼峰命名
func toCamelCase(s string) string {
	if s == "" {
		return s
	}
	
	// 常见映射
	mapping := map[string]string{
		"refresh_token": "refreshToken",
		"access_token":  "accessToken",
		"client_id":     "clientId",
		"client_secret": "clientSecret",
		"profile_arn":   "profileArn",
		"auth_method":   "authMethod",
		"idc_region":    "idcRegion",
	}
	
	if camel, ok := mapping[s]; ok {
		return camel
	}
	
	return s
}

// SetString 设置字符串值（标准化为下划线命名）
func (a *AccountCredential) SetString(key, value string) {
	if a.CredentialsData == nil {
		a.CredentialsData = make(map[string]interface{})
	}
	
	// 标准化 key 为下划线命名
	standardKey := toSnakeCase(key)
	a.CredentialsData[standardKey] = value
}

// toSnakeCase 将驼峰命名转换为下划线命名
func toSnakeCase(s string) string {
	if s == "" {
		return s
	}
	
	// 常见映射
	mapping := map[string]string{
		"refreshToken": "refresh_token",
		"accessToken":  "access_token",
		"clientId":     "client_id",
		"clientSecret": "client_secret",
		"profileArn":   "profile_arn",
		"authMethod":   "auth_method",
		"idcRegion":    "idc_region",
	}
	
	if snake, ok := mapping[s]; ok {
		return snake
	}
	
	return s
}
