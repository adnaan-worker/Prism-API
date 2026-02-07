package models

import (
	"time"

	"gorm.io/gorm"
)

type LoadBalancerConfig struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	ModelName string         `gorm:"not null;size:255;index" json:"model_name"`
	Strategy  string         `gorm:"not null;default:'round_robin';size:50" json:"strategy"`
	IsActive  bool           `gorm:"not null;default:true" json:"is_active"`
}

func (LoadBalancerConfig) TableName() string {
	return "load_balancer_configs"
}
