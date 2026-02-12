package apiconfig

import "time"

// CreateConfigRequest 创建配置请求
type CreateConfigRequest struct {
	Name     string                 `json:"name" binding:"required,min=1,max=255"`
	Type     string                 `json:"type" binding:"required,oneof=openai anthropic gemini kiro custom"`
	BaseURL  string                 `json:"base_url" binding:"required,url"`
	APIKey   string                 `json:"api_key" binding:"omitempty"`
	Models   []string               `json:"models" binding:"required,min=1"`
	Headers  map[string]interface{} `json:"headers" binding:"omitempty"`
	Metadata map[string]interface{} `json:"metadata" binding:"omitempty"`
	Priority int                    `json:"priority" binding:"omitempty,min=1,max=1000"`
	Weight   int                    `json:"weight" binding:"omitempty,min=1,max=100"`
	MaxRPS   int                    `json:"max_rps" binding:"omitempty,min=0"`
	Timeout  int                    `json:"timeout" binding:"omitempty,min=1,max=300"`
}

// UpdateConfigRequest 更新配置请求
type UpdateConfigRequest struct {
	Name     string                 `json:"name" binding:"omitempty,min=1,max=255"`
	Type     string                 `json:"type" binding:"omitempty,oneof=openai anthropic gemini kiro custom"`
	BaseURL  string                 `json:"base_url" binding:"omitempty,url"`
	APIKey   string                 `json:"api_key" binding:"omitempty"`
	Models   []string               `json:"models" binding:"omitempty,min=1"`
	Headers  map[string]interface{} `json:"headers" binding:"omitempty"`
	Metadata map[string]interface{} `json:"metadata" binding:"omitempty"`
	Priority *int                   `json:"priority" binding:"omitempty,min=1,max=1000"`
	Weight   *int                   `json:"weight" binding:"omitempty,min=1,max=100"`
	MaxRPS   *int                   `json:"max_rps" binding:"omitempty,min=0"`
	Timeout  *int                   `json:"timeout" binding:"omitempty,min=1,max=300"`
	IsActive *bool                  `json:"is_active" binding:"omitempty"`
}

// GetConfigsRequest 获取配置列表请求
type GetConfigsRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Type     string `form:"type" binding:"omitempty,oneof=openai anthropic gemini kiro custom"`
	IsActive *bool  `form:"is_active" binding:"omitempty"`
	Model    string `form:"model" binding:"omitempty"`
}

// BatchOperationRequest 批量操作请求
type BatchOperationRequest struct {
	IDs []uint `json:"ids" binding:"required,min=1"`
}

// ConfigResponse 配置响应
type ConfigResponse struct {
	ID        uint                   `json:"id"`
	Name      string                 `json:"name"`
	Type      string                 `json:"type"`
	BaseURL   string                 `json:"base_url"`
	APIKey    string                 `json:"api_key,omitempty"`
	Models    []string               `json:"models"`
	Headers   map[string]interface{} `json:"headers,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	IsActive  bool                   `json:"is_active"`
	Priority  int                    `json:"priority"`
	Weight    int                    `json:"weight"`
	MaxRPS    int                    `json:"max_rps"`
	Timeout   int                    `json:"timeout"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// ConfigListResponse 配置列表响应
type ConfigListResponse struct {
	Configs  []*ConfigResponse `json:"configs"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}

// BatchOperationResponse 批量操作响应
type BatchOperationResponse struct {
	Message string `json:"message"`
	Count   int    `json:"count"`
}

// ToResponse 转换为响应对象
func (c *APIConfig) ToResponse() *ConfigResponse {
	return &ConfigResponse{
		ID:        c.ID,
		Name:      c.Name,
		Type:      c.Type,
		BaseURL:   c.BaseURL,
		APIKey:    c.APIKey,
		Models:    c.Models,
		Headers:   c.Headers,
		Metadata:  c.Metadata,
		IsActive:  c.IsActive,
		Priority:  c.Priority,
		Weight:    c.Weight,
		MaxRPS:    c.MaxRPS,
		Timeout:   c.Timeout,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

// ToResponseList 批量转换为响应对象
func ToResponseList(configs []*APIConfig) []*ConfigResponse {
	responses := make([]*ConfigResponse, len(configs))
	for i, config := range configs {
		responses[i] = config.ToResponse()
	}
	return responses
}
