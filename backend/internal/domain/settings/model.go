package settings

import (
	"time"
)

// Setting 绯荤粺璁剧疆妯″瀷
type Setting struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 璁剧疆閿€?
	Key   string `gorm:"uniqueIndex;not null;size:255" json:"key"`
	Value string `gorm:"type:text" json:"value"`

	// 璁剧疆绫诲瀷
	Type string `gorm:"not null;size:50;default:'string'" json:"type"` // string, int, float, bool, json

	// 鎻忚堪
	Description string `gorm:"type:text" json:"description,omitempty"`

	// 鏄惁涓虹郴缁熻缃紙涓嶅彲鍒犻櫎锛?
	IsSystem bool `gorm:"not null;default:false" json:"is_system"`
}

// TableName 鎸囧畾琛ㄥ悕
func (Setting) TableName() string {
	return "settings"
}

// 璁剧疆閿父閲?
const (
	// 杩愯鏃堕厤缃?
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

	// 绯荤粺閰嶇疆
	KeySystemSiteName        = "system.site_name"
	KeySystemSiteDescription = "system.site_description"
	KeySystemAdminEmail      = "system.admin_email"
	KeySystemMaintenanceMode = "system.maintenance_mode"

	// 榛樿閰嶉
	KeyDefaultQuotaDaily   = "default_quota.daily"
	KeyDefaultQuotaMonthly = "default_quota.monthly"
	KeyDefaultQuotaTotal   = "default_quota.total"

	// 榛樿閫熺巼闄愬埗
	KeyDefaultRateLimitPerMinute = "default_rate_limit.per_minute"
	KeyDefaultRateLimitPerHour   = "default_rate_limit.per_hour"
	KeyDefaultRateLimitPerDay    = "default_rate_limit.per_day"
)

// 璁剧疆绫诲瀷甯搁噺
const (
	TypeString = "string"
	TypeInt    = "int"
	TypeFloat  = "float"
	TypeBool   = "bool"
	TypeJSON   = "json"
)
