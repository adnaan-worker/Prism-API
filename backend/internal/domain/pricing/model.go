package pricing

import (
	"time"

	"gorm.io/gorm"
)

// Pricing 定价模型
type Pricing struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	APIConfigID uint           `gorm:"not null;uniqueIndex:idx_config_model" json:"api_config_id"`
	ModelName   string         `gorm:"not null;size:255;uniqueIndex:idx_config_model" json:"model_name"`
	InputPrice  float64        `gorm:"not null;default:0" json:"input_price"`
	OutputPrice float64        `gorm:"not null;default:0" json:"output_price"`
	Currency    string         `gorm:"not null;default:'credits';size:20" json:"currency"`
	Unit        int            `gorm:"not null;default:1000" json:"unit"`
	IsActive    bool           `gorm:"not null;default:true" json:"is_active"`
	Description string         `gorm:"size:500" json:"description"`
}

// TableName 指定表名
func (Pricing) TableName() string {
	return "pricings"
}

// IsValid 检查定价是否有效
func (p *Pricing) IsValid() bool {
	return p.IsActive && p.DeletedAt.Time.IsZero()
}

// Activate 激活定价
func (p *Pricing) Activate() {
	p.IsActive = true
}

// Deactivate 停用定价
func (p *Pricing) Deactivate() {
	p.IsActive = false
}

// CalculateInputCost 计算输入成本
func (p *Pricing) CalculateInputCost(tokens int64) float64 {
	if p.Unit == 0 {
		return 0
	}
	return float64(tokens) / float64(p.Unit) * p.InputPrice
}

// CalculateOutputCost 计算输出成本
func (p *Pricing) CalculateOutputCost(tokens int64) float64 {
	if p.Unit == 0 {
		return 0
	}
	return float64(tokens) / float64(p.Unit) * p.OutputPrice
}

// CalculateTotalCost 计算总成本
func (p *Pricing) CalculateTotalCost(inputTokens, outputTokens int64) float64 {
	return p.CalculateInputCost(inputTokens) + p.CalculateOutputCost(outputTokens)
}
