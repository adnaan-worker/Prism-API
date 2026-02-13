package settings

import "time"

// RuntimeConfigResponse 运行时配置响应
type RuntimeConfigResponse struct {
	CacheEnabled      bool `json:"cache_enabled"`
	CacheTTL          int  `json:"cache_ttl"`           // 秒
	MaxRetries        int  `json:"max_retries"`
	Timeout           int  `json:"timeout"`             // 秒
	EnableLoadBalance bool `json:"enable_load_balance"`
}

// UpdateRuntimeConfigRequest 更新运行时配置请求
type UpdateRuntimeConfigRequest struct {
	CacheEnabled      *bool `json:"cache_enabled"`
	CacheTTL          *int  `json:"cache_ttl"`
	MaxRetries        *int  `json:"max_retries"`
	Timeout           *int  `json:"timeout"`
	EnableLoadBalance *bool `json:"enable_load_balance"`
}

// SystemConfigResponse 系统配置响应
type SystemConfigResponse struct {
	SiteName        string `json:"site_name"`
	SiteDescription string `json:"site_description"`
	AdminEmail      string `json:"admin_email"`
	MaintenanceMode bool   `json:"maintenance_mode"`
}

// UpdatePasswordRequest 修改密码请求
type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// DefaultQuotaResponse 默认配额响应
type DefaultQuotaResponse struct {
	Daily   int64 `json:"daily"`
	Monthly int64 `json:"monthly"`
	Total   int64 `json:"total"`
}

// UpdateDefaultQuotaRequest 更新默认配额请求
type UpdateDefaultQuotaRequest struct {
	Daily   *int64 `json:"daily"`
	Monthly *int64 `json:"monthly"`
	Total   *int64 `json:"total"`
}

// DefaultRateLimitResponse 默认速率限制响应
type DefaultRateLimitResponse struct {
	PerMinute int `json:"per_minute"`
	PerHour   int `json:"per_hour"`
	PerDay    int `json:"per_day"`
}

// UpdateDefaultRateLimitRequest 更新默认速率限制请求
type UpdateDefaultRateLimitRequest struct {
	PerMinute *int `json:"per_minute"`
	PerHour   *int `json:"per_hour"`
	PerDay    *int `json:"per_day"`
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
