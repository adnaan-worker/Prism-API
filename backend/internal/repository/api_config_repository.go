package repository

import (
	"api-aggregator/backend/internal/models"
	"context"
	"errors"

	"gorm.io/gorm"
)

type APIConfigRepository struct {
	db *gorm.DB
}

func NewAPIConfigRepository(db *gorm.DB) *APIConfigRepository {
	return &APIConfigRepository{db: db}
}

// Create creates a new API configuration
func (r *APIConfigRepository) Create(ctx context.Context, config *models.APIConfig) error {
	return r.db.WithContext(ctx).Create(config).Error
}

// FindByID finds an API configuration by ID
func (r *APIConfigRepository) FindByID(ctx context.Context, id uint) (*models.APIConfig, error) {
	var config models.APIConfig
	err := r.db.WithContext(ctx).First(&config, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &config, nil
}

// FindAll finds all API configurations
func (r *APIConfigRepository) FindAll(ctx context.Context) ([]*models.APIConfig, error) {
	var configs []*models.APIConfig
	err := r.db.WithContext(ctx).Find(&configs).Error
	if err != nil {
		return nil, err
	}
	return configs, nil
}

// FindActive finds all active API configurations
func (r *APIConfigRepository) FindActive(ctx context.Context) ([]*models.APIConfig, error) {
	var configs []*models.APIConfig
	err := r.db.WithContext(ctx).Where("is_active = ? AND deleted_at IS NULL", true).Find(&configs).Error
	if err != nil {
		return nil, err
	}
	return configs, nil
}

// FindByModel finds all API configurations that support a specific model
func (r *APIConfigRepository) FindByModel(ctx context.Context, model string) ([]*models.APIConfig, error) {
	var configs []*models.APIConfig
	// Use JSONB contains operator to find configs with the model in their models array
	err := r.db.WithContext(ctx).
		Where("is_active = ? AND models @> ?", true, `["`+model+`"]`).
		Order("priority DESC, weight DESC").
		Find(&configs).Error
	if err != nil {
		return nil, err
	}
	return configs, nil
}

// Update updates an API configuration
func (r *APIConfigRepository) Update(ctx context.Context, config *models.APIConfig) error {
	return r.db.WithContext(ctx).Save(config).Error
}

// Delete soft deletes an API configuration
func (r *APIConfigRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.APIConfig{}, id).Error
}

// UpdateStatus updates the active status of an API configuration
func (r *APIConfigRepository) UpdateStatus(ctx context.Context, id uint, isActive bool) error {
	return r.db.WithContext(ctx).Model(&models.APIConfig{}).
		Where("id = ?", id).
		Update("is_active", isActive).Error
}

// BatchDelete soft deletes multiple API configurations
func (r *APIConfigRepository) BatchDelete(ctx context.Context, ids []uint) error {
	return r.db.WithContext(ctx).Delete(&models.APIConfig{}, ids).Error
}

// BatchUpdateStatus updates the active status of multiple API configurations
func (r *APIConfigRepository) BatchUpdateStatus(ctx context.Context, ids []uint, isActive bool) error {
	return r.db.WithContext(ctx).Model(&models.APIConfig{}).
		Where("id IN ?", ids).
		Update("is_active", isActive).Error
}
