package repository

import (
	"api-aggregator/backend/internal/models"
	"context"

	"gorm.io/gorm"
)

type AccountCredentialRepository struct {
	db *gorm.DB
}

func NewAccountCredentialRepository(db *gorm.DB) *AccountCredentialRepository {
	return &AccountCredentialRepository{db: db}
}

func (r *AccountCredentialRepository) Create(ctx context.Context, cred *models.AccountCredential) (*models.AccountCredential, error) {
	if err := r.db.WithContext(ctx).Create(cred).Error; err != nil {
		return nil, err
	}
	return cred, nil
}

func (r *AccountCredentialRepository) FindByID(ctx context.Context, id uint) (*models.AccountCredential, error) {
	var cred models.AccountCredential
	if err := r.db.WithContext(ctx).First(&cred, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &cred, nil
}

func (r *AccountCredentialRepository) FindByPoolID(ctx context.Context, poolID uint) ([]*models.AccountCredential, error) {
	var creds []*models.AccountCredential
	err := r.db.WithContext(ctx).
		Joins("JOIN account_pool_credentials ON account_pool_credentials.credential_id = account_credentials.id").
		Where("account_pool_credentials.pool_id = ?", poolID).
		Find(&creds).Error
	return creds, err
}

func (r *AccountCredentialRepository) FindByProvider(ctx context.Context, provider string) ([]*models.AccountCredential, error) {
	var creds []*models.AccountCredential
	err := r.db.WithContext(ctx).Where("provider_type = ?", provider).Find(&creds).Error
	return creds, err
}

func (r *AccountCredentialRepository) FindByStatus(ctx context.Context, status string) ([]*models.AccountCredential, error) {
	var creds []*models.AccountCredential
	// status 映射到 is_active 字段
	isActive := status == "active"
	err := r.db.WithContext(ctx).Where("is_active = ?", isActive).Find(&creds).Error
	return creds, err
}

func (r *AccountCredentialRepository) FindAll(ctx context.Context) ([]*models.AccountCredential, error) {
	var creds []*models.AccountCredential
	err := r.db.WithContext(ctx).Find(&creds).Error
	return creds, err
}

func (r *AccountCredentialRepository) Update(ctx context.Context, cred *models.AccountCredential) (*models.AccountCredential, error) {
	if err := r.db.WithContext(ctx).Save(cred).Error; err != nil {
		return nil, err
	}
	return cred, nil
}

func (r *AccountCredentialRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.AccountCredential{}, id).Error
}
