package apikey

import (
	"api-aggregator/backend/pkg/query"
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

// Repository API密钥仓储接口
type Repository interface {
	Create(ctx context.Context, apiKey *APIKey) error
	Update(ctx context.Context, apiKey *APIKey) error
	Delete(ctx context.Context, id uint) error
	FindByID(ctx context.Context, id uint) (*APIKey, error)
	FindByKey(ctx context.Context, key string) (*APIKey, error)
	FindByUserID(ctx context.Context, userID uint) ([]*APIKey, error)
	List(ctx context.Context, userID uint, filters []query.Filter, sorts []query.Sort, pagination *query.Pagination) ([]*APIKey, int64, error)
	UpdateLastUsedAt(ctx context.Context, id uint) error
	UpdateStatus(ctx context.Context, id uint, isActive bool) error
	CountByUserID(ctx context.Context, userID uint) (int64, error)
}

// repository API密钥仓储实现
type repository struct {
	db *gorm.DB
}

// NewRepository 创建API密钥仓储
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// Create 创建API密钥
func (r *repository) Create(ctx context.Context, apiKey *APIKey) error {
	return r.db.WithContext(ctx).Create(apiKey).Error
}

// Update 更新API密钥
func (r *repository) Update(ctx context.Context, apiKey *APIKey) error {
	return r.db.WithContext(ctx).Save(apiKey).Error
}

// Delete 删除API密钥
func (r *repository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&APIKey{}, id).Error
}

// FindByID 根据ID查找API密钥
func (r *repository) FindByID(ctx context.Context, id uint) (*APIKey, error) {
	var apiKey APIKey
	err := r.db.WithContext(ctx).First(&apiKey, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &apiKey, nil
}

// FindByKey 根据密钥字符串查找API密钥
func (r *repository) FindByKey(ctx context.Context, key string) (*APIKey, error) {
	var apiKey APIKey
	err := r.db.WithContext(ctx).Where("key = ?", key).First(&apiKey).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &apiKey, nil
}

// FindByUserID 根据用户ID查找所有API密钥
func (r *repository) FindByUserID(ctx context.Context, userID uint) ([]*APIKey, error) {
	var apiKeys []*APIKey
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&apiKeys).Error
	if err != nil {
		return nil, err
	}
	return apiKeys, nil
}

// List 查询API密钥列表（支持过滤、排序、分页）
func (r *repository) List(ctx context.Context, userID uint, filters []query.Filter, sorts []query.Sort, pagination *query.Pagination) ([]*APIKey, int64, error) {
	builder := query.NewBuilder(r.db.WithContext(ctx).Model(&APIKey{}))
	
	// 添加用户ID过滤
	builder.Where("user_id = ?", userID)
	
	// 应用过滤
	builder.ApplyFilters(filters)
	
	// 统计总数
	var total int64
	builder.Count(&total)
	
	// 应用排序和分页
	builder.ApplySort(sorts).ApplyPagination(pagination)
	
	// 查询结果
	var apiKeys []*APIKey
	err := builder.Find(&apiKeys)
	if err != nil {
		return nil, 0, err
	}
	
	return apiKeys, total, nil
}

// UpdateLastUsedAt 更新最后使用时间
func (r *repository) UpdateLastUsedAt(ctx context.Context, id uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&APIKey{}).
		Where("id = ?", id).
		Update("last_used_at", now).Error
}

// UpdateStatus 更新API密钥状态
func (r *repository) UpdateStatus(ctx context.Context, id uint, isActive bool) error {
	return r.db.WithContext(ctx).Model(&APIKey{}).
		Where("id = ?", id).
		Update("is_active", isActive).Error
}

// CountByUserID 统计用户的API密钥数量
func (r *repository) CountByUserID(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&APIKey{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}
