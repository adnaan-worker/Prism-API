package apikey

import "time"

// CreateAPIKeyRequest 创建API密钥请求
type CreateAPIKeyRequest struct {
	Name      string `json:"name" binding:"required,min=1,max=100"`
	RateLimit int    `json:"rate_limit" binding:"omitempty,min=1,max=10000"`
}

// UpdateAPIKeyRequest 更新API密钥请求
type UpdateAPIKeyRequest struct {
	Name      string `json:"name" binding:"omitempty,min=1,max=100"`
	RateLimit int    `json:"rate_limit" binding:"omitempty,min=1,max=10000"`
	IsActive  *bool  `json:"is_active" binding:"omitempty"`
}

// GetAPIKeysRequest 获取API密钥列表请求
type GetAPIKeysRequest struct {
	Page     int  `form:"page" binding:"omitempty,min=1"`
	PageSize int  `form:"page_size" binding:"omitempty,min=1,max=100"`
	IsActive *bool `form:"is_active" binding:"omitempty"`
}

// APIKeyResponse API密钥响应
type APIKeyResponse struct {
	ID         uint       `json:"id"`
	Name       string     `json:"name"`
	Key        string     `json:"key"`
	IsActive   bool       `json:"is_active"`
	RateLimit  int        `json:"rate_limit"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// APIKeyListResponse API密钥列表响应
type APIKeyListResponse struct {
	Keys     []*APIKeyResponse `json:"keys"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}

// CreateAPIKeyResponse 创建API密钥响应
type CreateAPIKeyResponse struct {
	APIKey *APIKeyResponse `json:"api_key"`
}

// ToResponse 转换为响应对象
func (k *APIKey) ToResponse() *APIKeyResponse {
	return &APIKeyResponse{
		ID:         k.ID,
		Name:       k.Name,
		Key:        k.Key,
		IsActive:   k.IsActive,
		RateLimit:  k.RateLimit,
		LastUsedAt: k.LastUsedAt,
		CreatedAt:  k.CreatedAt,
		UpdatedAt:  k.UpdatedAt,
	}
}

// ToResponseList 批量转换为响应对象
func ToResponseList(keys []*APIKey) []*APIKeyResponse {
	responses := make([]*APIKeyResponse, len(keys))
	for i, key := range keys {
		responses[i] = key.ToResponse()
	}
	return responses
}
