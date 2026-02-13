package pricing

import "time"

// CreatePricingRequest 创建定价请求
type CreatePricingRequest struct {
	APIConfigID uint    `json:"api_config_id" binding:"required"`
	ModelName   string  `json:"model_name" binding:"required,min=1,max=255"`
	InputPrice  float64 `json:"input_price" binding:"required,min=0"`
	OutputPrice float64 `json:"output_price" binding:"required,min=0"`
	Currency    string  `json:"currency" binding:"omitempty,oneof=credits usd cny eur"`
	Unit        int     `json:"unit" binding:"omitempty,min=1"`
	Description string  `json:"description" binding:"omitempty,max=500"`
}

// UpdatePricingRequest 更新定价请求
type UpdatePricingRequest struct {
	InputPrice  *float64 `json:"input_price" binding:"omitempty,min=0"`
	OutputPrice *float64 `json:"output_price" binding:"omitempty,min=0"`
	Currency    string   `json:"currency" binding:"omitempty,oneof=credits usd cny eur"`
	Unit        *int     `json:"unit" binding:"omitempty,min=1"`
	IsActive    *bool    `json:"is_active" binding:"omitempty"`
	Description string   `json:"description" binding:"omitempty,max=500"`
}

// GetPricingsRequest 获取定价列表请求
type GetPricingsRequest struct {
	Page        int    `form:"page" binding:"omitempty,min=1"`
	PageSize    int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	APIConfigID uint   `form:"api_config_id" binding:"omitempty"`
	ModelName   string `form:"model_name" binding:"omitempty"`
	IsActive    *bool  `form:"is_active" binding:"omitempty"`
}

// CalculateCostRequest 计算成本请求
type CalculateCostRequest struct {
	ModelName    string `json:"model_name" binding:"required"`
	APIConfigID  uint   `json:"api_config_id" binding:"required"`
	InputTokens  int64  `json:"input_tokens" binding:"required,min=0"`
	OutputTokens int64  `json:"output_tokens" binding:"required,min=0"`
}

// BatchCreatePricingRequest 批量创建定价请求
type BatchCreatePricingRequest struct {
	Pricings []CreatePricingRequest `json:"pricings" binding:"required,min=1"`
}

// APIConfigInfo API配置基本信息
type APIConfigInfo struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// PricingResponse 定价响应
type PricingResponse struct {
	ID          uint           `json:"id"`
	APIConfigID uint           `json:"api_config_id"`
	APIConfig   *APIConfigInfo `json:"api_config,omitempty"`
	ModelName   string         `json:"model_name"`
	InputPrice  float64        `json:"input_price"`
	OutputPrice float64        `json:"output_price"`
	Currency    string         `json:"currency"`
	Unit        int            `json:"unit"`
	IsActive    bool           `json:"is_active"`
	Description string         `json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// PricingListResponse 定价列表响应
type PricingListResponse struct {
	Pricings []*PricingResponse `json:"pricings"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
}

// CostCalculationResponse 成本计算响应
type CostCalculationResponse struct {
	ModelName    string  `json:"model_name"`
	InputTokens  int64   `json:"input_tokens"`
	OutputTokens int64   `json:"output_tokens"`
	InputCost    float64 `json:"input_cost"`
	OutputCost   float64 `json:"output_cost"`
	TotalCost    float64 `json:"total_cost"`
	Currency     string  `json:"currency"`
	Unit         int     `json:"unit"`
}

// BatchCreatePricingResponse 批量创建定价响应
type BatchCreatePricingResponse struct {
	Created int      `json:"created"`
	Failed  int      `json:"failed"`
	Errors  []string `json:"errors,omitempty"`
}

// ToResponse 转换为响应对象
func (p *Pricing) ToResponse() *PricingResponse {
	return &PricingResponse{
		ID:          p.ID,
		APIConfigID: p.APIConfigID,
		ModelName:   p.ModelName,
		InputPrice:  p.InputPrice,
		OutputPrice: p.OutputPrice,
		Currency:    p.Currency,
		Unit:        p.Unit,
		IsActive:    p.IsActive,
		Description: p.Description,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

// ToResponseList 批量转换为响应对象
func ToResponseList(pricings []*Pricing) []*PricingResponse {
	responses := make([]*PricingResponse, len(pricings))
	for i, pricing := range pricings {
		responses[i] = pricing.ToResponse()
	}
	return responses
}
