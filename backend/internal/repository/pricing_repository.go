package repository

import (
	"api-aggregator/backend/internal/models"
	"context"

	"gorm.io/gorm"
)

type PricingRepository struct {
	db *gorm.DB
}

func NewPricingRepository(db *gorm.DB) *PricingRepository {
	return &PricingRepository{db: db}
}

// FindAll retrieves all pricing configurations with API config info
func (r *PricingRepository) FindAll(ctx context.Context) ([]*models.Pricing, error) {
	var pricings []*models.Pricing
	err := r.db.WithContext(ctx).
		Preload("APIConfig").
		Order("api_config_id ASC, model_name ASC").
		Find(&pricings).Error
	return pricings, err
}

// FindActive retrieves all active pricing configurations
func (r *PricingRepository) FindActive(ctx context.Context) ([]*models.Pricing, error) {
	var pricings []*models.Pricing
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Find(&pricings).Error
	return pricings, err
}

// FindByID retrieves a pricing configuration by ID with API config info
func (r *PricingRepository) FindByID(ctx context.Context, id uint) (*models.Pricing, error) {
	var pricing models.Pricing
	err := r.db.WithContext(ctx).Preload("APIConfig").First(&pricing, id).Error
	if err != nil {
		return nil, err
	}
	return &pricing, nil
}

// FindByModelAndAPIConfig retrieves pricing by model name and API config ID
func (r *PricingRepository) FindByModelAndAPIConfig(ctx context.Context, modelName string, apiConfigID uint) (*models.Pricing, error) {
	var pricing models.Pricing
	err := r.db.WithContext(ctx).
		Where("model_name = ? AND api_config_id = ? AND is_active = ?", modelName, apiConfigID, true).
		First(&pricing).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &pricing, nil
}

// FindByAPIConfig retrieves all pricings for an API config
func (r *PricingRepository) FindByAPIConfig(ctx context.Context, apiConfigID uint) ([]*models.Pricing, error) {
	var pricings []*models.Pricing
	err := r.db.WithContext(ctx).
		Preload("APIConfig").
		Where("api_config_id = ?", apiConfigID).
		Find(&pricings).Error
	return pricings, err
}

// Create creates a new pricing configuration
func (r *PricingRepository) Create(ctx context.Context, pricing *models.Pricing) error {
	return r.db.WithContext(ctx).Create(pricing).Error
}

// Update updates a pricing configuration
func (r *PricingRepository) Update(ctx context.Context, pricing *models.Pricing) error {
	return r.db.WithContext(ctx).Save(pricing).Error
}

// Delete deletes a pricing configuration
func (r *PricingRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Pricing{}, id).Error
}

// BatchCreate creates multiple pricing configurations
func (r *PricingRepository) BatchCreate(ctx context.Context, pricings []*models.Pricing) error {
	return r.db.WithContext(ctx).Create(&pricings).Error
}
