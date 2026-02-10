package models

import (
	"time"

	"gorm.io/gorm"
)

// Pricing represents the pricing configuration for a model from a specific API config
type Pricing struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	
	// API Config reference
	APIConfigID uint   `gorm:"not null;uniqueIndex:idx_config_model" json:"api_config_id"`
	APIConfig   *APIConfig `gorm:"foreignKey:APIConfigID" json:"api_config,omitempty"`
	
	// Model identification
	ModelName string `gorm:"not null;size:255;uniqueIndex:idx_config_model" json:"model_name"`
	
	// Pricing (per 1000 tokens)
	InputPrice  float64 `gorm:"not null;default:0" json:"input_price"`   // Price per 1000 input tokens
	OutputPrice float64 `gorm:"not null;default:0" json:"output_price"`  // Price per 1000 output tokens
	
	// Currency and unit
	Currency string `gorm:"not null;default:'credits';size:20" json:"currency"` // credits, usd, cny, etc.
	Unit     int    `gorm:"not null;default:1000" json:"unit"`                  // Pricing unit (usually 1000 tokens)
	
	// Status
	IsActive bool `gorm:"not null;default:true" json:"is_active"`
	
	// Description
	Description string `gorm:"size:500" json:"description"`
}

func (Pricing) TableName() string {
	return "pricings"
}
