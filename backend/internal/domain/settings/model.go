package settings

import (
	"time"

	"gorm.io/gorm"
)

// Setting 系统设置模型
type Setting struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// 设置键值
	Key   string `gorm:"uniqueIndex;not null;size:255" json:"key"`
	Value string `gorm:"type:text" json:"value"`

	// 设置类型
	Type string `gorm:"not null;size:50;default:'string'" json:"type"` // string, int, float, bool, json

	// 描述
	Description string `gorm:"type:text" json:"description,omitempty"`

	// 是否为系统设置（不可删除）
	IsSystem bool `gorm:"not null;default:false" json:"is_system"`
}

// TableName 指定表名
func (Setting) TableName() string {
	return "settings"
}

// 设置键常量
const (
	// 运行时配置
	KeyRuntimeCacheEnabled         = "runtime.cache_enabled"
	KeyRuntimeCacheTTL             = "runtime.cache_ttl"
	KeyRuntimeSemanticCacheEnabled = "runtime.semantic_cache_enabled"
	KeyRuntimeSemanticThreshold    = "runtime.semantic_threshold"
	KeyRuntimeEmbeddingEnabled     = "runtime.embedding_enabled"
	KeyRuntimeEmbeddingURL         = "runtime.embedding_url"
	KeyRuntimeEmbeddingTimeout     = "runtime.embedding_timeout"
	KeyRuntimeMaxRetries           = "runtime.max_retries"
	KeyRuntimeTimeout              = "runtime.timeout"
	KeyRuntimeEnableLoadBalance    = "runtime.enable_load_balance"

	// 系统配置
	KeySystemSiteName        = "system.site_name"
	KeySystemSiteDescription = "system.site_description"
	KeySystemAdminEmail      = "system.admin_email"
	KeySystemMaintenanceMode = "system.maintenance_mode"

	// 默认配额
	KeyDefaultQuotaDaily   = "default_quota.daily"
	KeyDefaultQuotaMonthly = "default_quota.monthly"
	KeyDefaultQuotaTotal   = "default_quota.total"

	// 默认速率限制
	KeyDefaultRateLimitPerMinute = "default_rate_limit.per_minute"
	KeyDefaultRateLimitPerHour   = "default_rate_limit.per_hour"
	KeyDefaultRateLimitPerDay    = "default_rate_limit.per_day"
)

// 设置类型常量
const (
	TypeString = "string"
	TypeInt    = "int"
	TypeFloat  = "float"
	TypeBool   = "bool"
	TypeJSON   = "json"
)
