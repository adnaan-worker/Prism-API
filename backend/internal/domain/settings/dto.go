package settings

import "time"

// RuntimeConfigResponse 运行时配置响应
type RuntimeConfigResponse struct {
	CacheEnabled         bool    `json:"cache_enabled"`
	CacheTTL             string  `json:"cache_ttl"`              // 格式: "24h", "1h30m"
	SemanticCacheEnabled bool    `json:"semantic_cache_enabled"`
	SemanticThreshold    float64 `json:"semantic_threshold"`     // 0.0 ~ 1.0
	EmbeddingEnabled     bool    `json:"embedding_enabled"`
}

// UpdateRuntimeConfigRequest 更新运行时配置请求
type UpdateRuntimeConfigRequest struct {
	CacheEnabled         *bool    `json:"cache_enabled"`
	CacheTTL             *string  `json:"cache_ttl"`
	SemanticCacheEnabled *bool    `json:"semantic_cache_enabled"`
	SemanticThreshold    *float64 `json:"semantic_threshold"`
	EmbeddingEnabled     *bool    `json:"embedding_enabled"`
}

// SystemConfigResponse 系统运行信息响应
type SystemConfigResponse struct {
	// 缓存配置
	CacheEnabled         bool    `json:"cache_enabled"`
	CacheTTL             string  `json:"cache_ttl"`
	SemanticCacheEnabled bool    `json:"semantic_cache_enabled"`
	SemanticThreshold    float64 `json:"semantic_threshold"`
	EmbeddingEnabled     bool    `json:"embedding_enabled"`
	// 速率限制
	RateLimitEnabled  bool   `json:"rate_limit_enabled"`
	RateLimitRequests int    `json:"rate_limit_requests"`
	RateLimitWindow   string `json:"rate_limit_window"`
	// 服务信息
	Version   string `json:"version"`
	Uptime    string `json:"uptime"`
	GoVersion string `json:"go_version"`
}

// UpdatePasswordRequest 修改密码请求
type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// DefaultQuotaResponse 默认配额响应
type DefaultQuotaResponse struct {
	DefaultQuota int64 `json:"default_quota"`
}

// UpdateDefaultQuotaRequest 更新默认配额请求
type UpdateDefaultQuotaRequest struct {
	DefaultQuota *int64 `json:"default_quota"`
}

// DefaultRateLimitResponse 默认速率限制响应
type DefaultRateLimitResponse struct {
	RequestsPerMinute int `json:"requests_per_minute"`
	RequestsPerDay    int `json:"requests_per_day"`
}

// UpdateDefaultRateLimitRequest 更新默认速率限制请求
type UpdateDefaultRateLimitRequest struct {
	RequestsPerMinute *int `json:"requests_per_minute"`
	RequestsPerDay    *int `json:"requests_per_day"`
}

// SettingResponse 设置响应
type SettingResponse struct {
	ID          uint      `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Key         string    `json:"key"`
	Value       string    `json:"value"`
	Type        string    `json:"type"`
	Description string    `json:"description,omitempty"`
	IsSystem    bool      `json:"is_system"`
}

// ToSettingResponse 转换为设置响应
func ToSettingResponse(setting *Setting) *SettingResponse {
	if setting == nil {
		return nil
	}
	return &SettingResponse{
		ID:          setting.ID,
		CreatedAt:   setting.CreatedAt,
		UpdatedAt:   setting.UpdatedAt,
		Key:         setting.Key,
		Value:       setting.Value,
		Type:        setting.Type,
		Description: setting.Description,
		IsSystem:    setting.IsSystem,
	}
}
