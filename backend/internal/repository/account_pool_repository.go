package repository

import (
	"api-aggregator/backend/internal/models"
	"context"

	"gorm.io/gorm"
)

type AccountPoolRepository struct {
	db *gorm.DB
}

func NewAccountPoolRepository(db *gorm.DB) *AccountPoolRepository {
	return &AccountPoolRepository{db: db}
}

func (r *AccountPoolRepository) Create(ctx context.Context, pool *models.AccountPool) (*models.AccountPool, error) {
	if err := r.db.WithContext(ctx).Create(pool).Error; err != nil {
		return nil, err
	}
	return pool, nil
}

func (r *AccountPoolRepository) FindByID(ctx context.Context, id uint) (*models.AccountPool, error) {
	var pool models.AccountPool
	if err := r.db.WithContext(ctx).First(&pool, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &pool, nil
}

func (r *AccountPoolRepository) FindByProvider(ctx context.Context, provider string) ([]*models.AccountPool, error) {
	var pools []*models.AccountPool
	err := r.db.WithContext(ctx).Where("provider_type = ?", provider).Find(&pools).Error
	return pools, err
}

func (r *AccountPoolRepository) FindAll(ctx context.Context) ([]*models.AccountPool, error) {
	var pools []*models.AccountPool
	err := r.db.WithContext(ctx).Find(&pools).Error
	return pools, err
}

func (r *AccountPoolRepository) Update(ctx context.Context, pool *models.AccountPool) (*models.AccountPool, error) {
	if err := r.db.WithContext(ctx).Save(pool).Error; err != nil {
		return nil, err
	}
	return pool, nil
}

func (r *AccountPoolRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.AccountPool{}, id).Error
}

func (r *AccountPoolRepository) GetPoolCredentials(ctx context.Context, poolID uint) ([]*models.AccountCredential, error) {
	var creds []*models.AccountCredential
	err := r.db.WithContext(ctx).
		Joins("JOIN account_pool_credentials ON account_pool_credentials.credential_id = account_credentials.id").
		Where("account_pool_credentials.pool_id = ?", poolID).
		Find(&creds).Error
	return creds, err
}

// AddCredentialToPool 将凭据添加到池
func (r *AccountPoolRepository) AddCredentialToPool(ctx context.Context, poolID, credentialID uint) error {
	sql := `
		INSERT INTO account_pool_credentials (pool_id, credential_id, created_at)
		VALUES (?, ?, NOW())
		ON CONFLICT (pool_id, credential_id) DO NOTHING
	`
	return r.db.WithContext(ctx).Exec(sql, poolID, credentialID).Error
}

// RemoveCredentialFromPool 从池中移除凭据
func (r *AccountPoolRepository) RemoveCredentialFromPool(ctx context.Context, poolID, credentialID uint) error {
	sql := `DELETE FROM account_pool_credentials WHERE pool_id = ? AND credential_id = ?`
	return r.db.WithContext(ctx).Exec(sql, poolID, credentialID).Error
}
