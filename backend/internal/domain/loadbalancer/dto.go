package loadbalancer

import "time"

// CreateConfigRequest 创建负载均衡配置请求
type CreateConfigRequest struct {
	ModelName string `json:"model_name" binding:"required"`
	Strategy  string `json:"strategy" binding:"required"`
}

// UpdateConfigRequest 更新负载均衡配置请求
type UpdateConfigRequest struct {
	Strategy string `json:"strategy"`
	IsActive *bool  `json:"is_active"`
}

// ConfigResponse 负载均衡配置响应
type ConfigResponse struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	ModelName string    `json:"model_name"`
	Strategy  string    `json:"strategy"`
	IsActive  bool      `json:"is_active"`
}

// ConfigListResponse 配置列表响应
type ConfigListResponse struct {
	Configs []*ConfigResponse `json:"configs"`
	Total   int64             `json:"total"`
}

// EndpointInfo 端点信息
type EndpointInfo struct {
	ConfigID     uint    `json:"config_id"`
	ConfigName   string  `json:"config_name"`
	Type         string  `json:"type"`
	BaseURL      string  `json:"base_url"`
	Priority     int     `json:"priority"`
	Weight       int     `json:"weight"`
	IsActive     bool    `json:"is_active"`
	HealthStatus string  `json:"health_status"`
	ResponseTime *int64  `json:"response_time,omitempty"`
	SuccessRate  *float64 `json:"success_rate,omitempty"`
}

// ModelEndpointsResponse 模型端点响应
type ModelEndpointsResponse struct {
	ModelName string          `json:"model_name"`
	Endpoints []*EndpointInfo `json:"endpoints"`
	Total     int             `json:"total"`
}

// AvailableModelsResponse 可用模型响应
type AvailableModelsResponse struct {
	Models []string `json:"models"`
	Total  int      `json:"total"`
}

// ConfigFilter 配置过滤器
type ConfigFilter struct {
	ModelName *string
	Strategy  *string
	IsActive  *bool
}

// ToConfigResponse 转换为配置响应
func ToConfigResponse(config *LoadBalancerConfig) *ConfigResponse {
	if config == nil {
		return nil
	}
	return &ConfigResponse{
		ID:        config.ID,
		CreatedAt: config.CreatedAt,
		UpdatedAt: config.UpdatedAt,
		ModelName: config.ModelName,
		Strategy:  config.Strategy,
		IsActive:  config.IsActive,
	}
}

// ToConfigListResponse 转换为配置列表响应
func ToConfigListResponse(configs []*LoadBalancerConfig, total int64) *ConfigListResponse {
	responses := make([]*ConfigResponse, len(configs))
	for i, config := range configs {
		responses[i] = ToConfigResponse(config)
	}
	return &ConfigListResponse{
		Configs: responses,
		Total:   total,
	}
}
