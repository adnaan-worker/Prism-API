package repository

import (
	"api-aggregator/backend/internal/models"
	"context"

	"gorm.io/gorm"
)

type LoadBalancerRepository struct {
	db *gorm.DB
}

func NewLoadBalancerRepository(db *gorm.DB) *LoadBalancerRepository {
	return &LoadBalancerRepository{
		db: db,
	}
}

// FindAll retrieves all load balancer configurations
func (r *LoadBalancerRepository) FindAll(ctx context.Context) ([]*models.LoadBalancerConfig, error) {
	var configs []*models.LoadBalancerConfig
	err := r.db.WithContext(ctx).Find(&configs).Error
	return configs, err
}

// FindByID retrieves a load balancer configuration by ID
func (r *LoadBalancerRepository) FindByID(ctx context.Context, id uint) (*models.LoadBalancerConfig, error) {
	var config models.LoadBalancerConfig
	err := r.db.WithContext(ctx).First(&config, id).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// FindByModel retrieves a load balancer configuration by model name
func (r *LoadBalancerRepository) FindByModel(ctx context.Context, modelName string) (*models.LoadBalancerConfig, error) {
	var config models.LoadBalancerConfig
	err := r.db.WithContext(ctx).Where("model_name = ? AND is_active = ?", modelName, true).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// Create creates a new load balancer configuration
func (r *LoadBalancerRepository) Create(ctx context.Context, config *models.LoadBalancerConfig) error {
	return r.db.WithContext(ctx).Create(config).Error
}

// Update updates a load balancer configuration
func (r *LoadBalancerRepository) Update(ctx context.Context, id uint, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&models.LoadBalancerConfig{}).Where("id = ?", id).Updates(updates).Error
}

// Delete deletes a load balancer configuration
func (r *LoadBalancerRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.LoadBalancerConfig{}, id).Error
}
