package models

import (
	"time"

	"gorm.io/gorm"
)

// BillingTransaction records all billing operations for audit trail
type BillingTransaction struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	
	// User and request info
	UserID       uint   `gorm:"not null;index" json:"user_id"`
	APIKeyID     uint   `gorm:"not null" json:"api_key_id"`
	APIConfigID  uint   `gorm:"not null" json:"api_config_id"`
	RequestLogID uint   `gorm:"index" json:"request_log_id"` // Link to request log
	
	// Transaction details
	Type        string `gorm:"not null;size:20" json:"type"` // charge, refund, adjustment, sign_in
	Status      string `gorm:"not null;size:20;default:'completed'" json:"status"` // pending, completed, failed, reversed
	
	// Amount details (in micro-credits, 1 credit = 1000 micro-credits)
	MicroCredits int64 `gorm:"not null" json:"micro_credits"` // Amount in micro-credits
	
	// Token usage
	InputTokens  int `gorm:"not null;default:0" json:"input_tokens"`
	OutputTokens int `gorm:"not null;default:0" json:"output_tokens"`
	TotalTokens  int `gorm:"not null;default:0" json:"total_tokens"`
	
	// Pricing details
	PricingID   uint    `gorm:"index" json:"pricing_id"`
	InputPrice  float64 `json:"input_price"`  // Price per unit at time of transaction
	OutputPrice float64 `json:"output_price"` // Price per unit at time of transaction
	
	// Model and metadata
	Model       string `gorm:"size:255" json:"model"`
	IsEstimate  bool   `gorm:"not null;default:false" json:"is_estimate"` // True for streaming requests
	Description string `gorm:"size:500" json:"description"`
	
	// Balance snapshot (for audit)
	BalanceBefore int64 `json:"balance_before"` // Available balance before transaction
	BalanceAfter  int64 `json:"balance_after"`  // Available balance after transaction
}

func (BillingTransaction) TableName() string {
	return "billing_transactions"
}
