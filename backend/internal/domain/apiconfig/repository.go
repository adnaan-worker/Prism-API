package apiconfig

import (
	"api-aggregator/backend/pkg/query"
	"context"
	"errors"

	"gorm.io/gorm"
)

// Repository API配置仓储接口
type Repository interface {
	Create(ctx context.Context, config *APIConfig) error
	Update(ctx context.Context, config *APIConfig) error
	Delete(ctx context.Context, id uint) error
	FindByID(ctx context.Context, id uint) (*APIConfig, error)
	FindAll(ctx context.Context) ([]*APIConfig, error)
	FindActive(ctx context.Context) ([]*APIConfig, error)
	FindByType(ctx context.Context, configType string) ([]*APIConfig, error)
	FindByModel(ctx context.Context, model string) ([]*APIConfig, error)
	List(ctx context.Context, filters []query.Filter, sorts []query.Sort, pagination *query.Pagination) ([]*APIConfig, int64, error)
	UpdateStatus(ctx context.Context, id uint, isActive bool) error
	BatchDelete(ctx context.Context, ids []uint) error
	BatchUpdateStatus(ctx context.Context, ids []uint, isActive bool) error
	CountAll(ctx context.Context) (int64, error)
	CountByType(ctx context.Context, configType string) (int64, error)
	CountActive(ctx context.Context) (int64, error)
}

// repository API配置仓储实现
type repository struct {
	db *gorm.DB
}

// NewRepository 创建API配置仓储
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// Create 创建配置
func (r *repository) Create(ctx context.Context, config *APIConfig) error {
	return r.db.WithContext(ctx).Create(config).Error
}

// Update 更新配置
func (r *repository) Update(ctx context.Context, config *APIConfig) error {
	return r.db.WithContext(ctx).Save(config).Error
}

// Delete 删除配置
func (r *repository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&APIConfig{}, id).Error
}

// FindByID 根据ID查找配置
func (r *repository) FindByID(ctx context.Context, id uint) (*APIConfig, error) {
	var config APIConfig
	err := r.db.WithContext(ctx).First(&config, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &config, nil
}

// FindAll 查找所有配置
func (r *repository) FindAll(ctx context.Context) ([]*APIConfig, error) {
	var configs []*APIConfig
	err := r.db.WithContext(ctx).Order("priority ASC, created_at DESC").Find(&configs).Error
	if err != nil {
		return nil, err
	}
	return configs, nil
}

// FindActive 查找所有激活的配置
func (r *repository) FindActive(ctx context.Context) ([]*APIConfig, error) {
	var configs []*APIConfig
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("priority ASC, created_at DESC").
		Find(&configs).Error
	if err != nil {
		return nil, err
	}
	return configs, nil
}

// FindByType 根据类型查找配置
func (r *repository) FindByType(ctx context.Context, configType string) ([]*APIConfig, error) {
	var configs []*APIConfig
	err := r.db.WithContext(ctx).
		Where("type = ?", configType).
		Order("priority ASC, created_at DESC").
		Find(&configs).Error
	if err != nil {
		return nil, err
	}
	return configs, nil
}

// FindByModel 根据模型查找配置
func (r *repository) FindByModel(ctx context.Context, model string) ([]*APIConfig, error) {
	var configs []*APIConfig
	// 使用 PostgreSQL 的 JSONB 查询
	err := r.db.WithContext(ctx).
		Where("models @> ?", `["`+model+`"]`).
		Where("is_active = ?", true).
		Order("priority ASC, created_at DESC").
		Find(&configs).Error
	if err != nil {
		return nil, err
	}
	return configs, nil
}

// List 查询配置列表（支持过滤、排序、分页）
func (r *repository) List(ctx context.Context, filters []query.Filter, sorts []query.Sort, pagination *query.Pagination) ([]*APIConfig, int64, error) {
	builder := query.NewBuilder(r.db.WithContext(ctx).Model(&APIConfig{}))
	
	// 应用过滤
	builder.ApplyFilters(filters)
	
	// 统计总数
	var total int64
	builder.Count(&total)
	
	// 应用排序和分页
	if len(sorts) == 0 {
		sorts = []query.Sort{
			{Field: "priority", Desc: false},
			{Field: "created_at", Desc: true},
		}
	}
	builder.ApplySort(sorts).ApplyPagination(pagination)
	
	// 查询结果
	var configs []*APIConfig
	err := builder.Find(&configs)
	if err != nil {
		return nil, 0, err
	}
	
	return configs, total, nil
}

// UpdateStatus 更新配置状态
func (r *repository) UpdateStatus(ctx context.Context, id uint, isActive bool) error {
	return r.db.WithContext(ctx).Model(&APIConfig{}).
		Where("id = ?", id).
		Update("is_active", isActive).Error
}

// BatchDelete 批量删除配置
func (r *repository) BatchDelete(ctx context.Context, ids []uint) error {
	return r.db.WithContext(ctx).Delete(&APIConfig{}, ids).Error
}

// BatchUpdateStatus 批量更新配置状态
func (r *repository) BatchUpdateStatus(ctx context.Context, ids []uint, isActive bool) error {
	return r.db.WithContext(ctx).Model(&APIConfig{}).
		Where("id IN ?", ids).
		Update("is_active", isActive).Error
}

// CountAll 统计所有配置数量
func (r *repository) CountAll(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&APIConfig{}).Count(&count).Error
	return count, err
}

// CountByType 根据类型统计配置数量
func (r *repository) CountByType(ctx context.Context, configType string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&APIConfig{}).
		Where("type = ?", configType).
		Count(&count).Error
	return count, err
}

// CountActive 统计激活的配置数量
func (r *repository) CountActive(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&APIConfig{}).
		Where("is_active = ?", true).
		Count(&count).Error
	return count, err
}
