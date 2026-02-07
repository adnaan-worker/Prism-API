package repository

import (
	"api-aggregator/backend/internal/models"
	"context"
	"errors"

	"gorm.io/gorm"
)

type APIKeyRepository struct {
	db *gorm.DB
}

func NewAPIKeyRepository(db *gorm.DB) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

// Create creates a new API key
func (r *APIKeyRepository) Create(ctx context.Context, apiKey *models.APIKey) error {
	return r.db.WithContext(ctx).Create(apiKey).Error
}

// FindByID finds an API key by ID
func (r *APIKeyRepository) FindByID(ctx context.Context, id uint) (*models.APIKey, error) {
	var apiKey models.APIKey
	err := r.db.WithContext(ctx).First(&apiKey, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &apiKey, nil
}

// FindByKey finds an API key by key string
func (r *APIKeyRepository) FindByKey(ctx context.Context, key string) (*models.APIKey, error) {
	var apiKey models.APIKey
	err := r.db.WithContext(ctx).Where("key = ?", key).First(&apiKey).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &apiKey, nil
}

// FindByUserID finds all API keys for a user
func (r *APIKeyRepository) FindByUserID(ctx context.Context, userID uint) ([]*models.APIKey, error) {
	var apiKeys []*models.APIKey
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&apiKeys).Error
	if err != nil {
		return nil, err
	}
	return apiKeys, nil
}

// Update updates an API key
func (r *APIKeyRepository) Update(ctx context.Context, apiKey *models.APIKey) error {
	return r.db.WithContext(ctx).Save(apiKey).Error
}

// Delete soft deletes an API key
func (r *APIKeyRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.APIKey{}, id).Error
}

// UpdateLastUsedAt updates the last used timestamp
func (r *APIKeyRepository) UpdateLastUsedAt(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&models.APIKey{}).Where("id = ?", id).Update("last_used_at", gorm.Expr("NOW()")).Error
}
