package pricing

import (
	"time"
)

// Pricing 瀹氫环妯″瀷
type Pricing struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	APIConfigID uint           `gorm:"not null;uniqueIndex:idx_config_model" json:"api_config_id"`
	ModelName   string         `gorm:"not null;size:255;uniqueIndex:idx_config_model" json:"model_name"`
	InputPrice  float64        `gorm:"not null;default:0" json:"input_price"`
	OutputPrice float64        `gorm:"not null;default:0" json:"output_price"`
	Currency    string         `gorm:"not null;default:'credits';size:20" json:"currency"`
	Unit        int            `gorm:"not null;default:1000" json:"unit"`
	IsActive    bool           `gorm:"not null;default:true" json:"is_active"`
	Description string         `gorm:"size:500" json:"description"`
}

// TableName 鎸囧畾琛ㄥ悕
func (Pricing) TableName() string {
	return "pricings"
}

// IsValid 妫€鏌ュ畾浠锋槸鍚︽湁鏁?
func (p *Pricing) IsValid() bool {
	return p.IsActive
}

// Activate 婵€娲诲畾浠?
func (p *Pricing) Activate() {
	p.IsActive = true
}

// Deactivate 鍋滅敤瀹氫环
func (p *Pricing) Deactivate() {
	p.IsActive = false
}

// CalculateInputCost 璁＄畻杈撳叆鎴愭湰
func (p *Pricing) CalculateInputCost(tokens int64) float64 {
	if p.Unit == 0 {
		return 0
	}
	return float64(tokens) / float64(p.Unit) * p.InputPrice
}

// CalculateOutputCost 璁＄畻杈撳嚭鎴愭湰
func (p *Pricing) CalculateOutputCost(tokens int64) float64 {
	if p.Unit == 0 {
		return 0
	}
	return float64(tokens) / float64(p.Unit) * p.OutputPrice
}

// CalculateTotalCost 璁＄畻鎬绘垚鏈?
func (p *Pricing) CalculateTotalCost(inputTokens, outputTokens int64) float64 {
	return p.CalculateInputCost(inputTokens) + p.CalculateOutputCost(outputTokens)
}
