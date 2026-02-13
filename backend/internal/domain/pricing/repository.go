package pricing

import (
	"api-aggregator/backend/pkg/query"
	"context"
	"errors"

	"gorm.io/gorm"
)

// Repository 定价仓储接口
type Repository interface {
	Create(ctx context.Context, pricing *Pricing) error
	Update(ctx context.Context, pricing *Pricing) error
	Delete(ctx context.Context, id uint) error
	FindByID(ctx context.Context, id uint) (*Pricing, error)
	FindByModelAndAPIConfig(ctx context.Context, modelName string, apiConfigID uint) (*Pricing, error)
	FindByAPIConfig(ctx context.Context, apiConfigID uint) ([]*Pricing, error)
	FindByModel(ctx context.Context, modelName string) ([]*Pricing, error)
	FindAll(ctx context.Context) ([]*Pricing, error)
	FindActive(ctx context.Context) ([]*Pricing, error)
	List(ctx context.Context, filters []query.Filter, sorts []query.Sort, pagination *query.Pagination) ([]*Pricing, int64, error)
	BatchCreate(ctx context.Context, pricings []*Pricing) error
	UpdateStatus(ctx context.Context, id uint, isActive bool) error
	CountAll(ctx context.Context) (int64, error)
	CountByAPIConfig(ctx context.Context, apiConfigID uint) (int64, error)
}

// repository 定价仓储实现
type repository struct {
	db *gorm.DB
}

// NewRepository 创建定价仓储
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// Create 创建定价
func (r *repository) Create(ctx context.Context, pricing *Pricing) error {
	return r.db.WithContext(ctx).Create(pricing).Error
}

// Update 更新定价
func (r *repository) Update(ctx context.Context, pricing *Pricing) error {
	return r.db.WithContext(ctx).Save(pricing).Error
}

// Delete 删除定价
func (r *repository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&Pricing{}, id).Error
}

// FindByID 根据ID查找定价
func (r *repository) FindByID(ctx context.Context, id uint) (*Pricing, error) {
	var pricing Pricing
	err := r.db.WithContext(ctx).First(&pricing, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &pricing, nil
}

// FindByModelAndAPIConfig 根据模型和API配置查找定价
func (r *repository) FindByModelAndAPIConfig(ctx context.Context, modelName string, apiConfigID uint) (*Pricing, error) {
	var pricing Pricing
	err := r.db.WithContext(ctx).
		Where("model_name = ? AND api_config_id = ?", modelName, apiConfigID).
		First(&pricing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &pricing, nil
}

// FindByAPIConfig 根据API配置查找所有定价
func (r *repository) FindByAPIConfig(ctx context.Context, apiConfigID uint) ([]*Pricing, error) {
	var pricings []*Pricing
	err := r.db.WithContext(ctx).
		Where("api_config_id = ?", apiConfigID).
		Order("model_name ASC").
		Find(&pricings).Error
	if err != nil {
		return nil, err
	}
	return pricings, nil
}

// FindByModel 根据模型查找所有定价
func (r *repository) FindByModel(ctx context.Context, modelName string) ([]*Pricing, error) {
	var pricings []*Pricing
	err := r.db.WithContext(ctx).
		Where("model_name = ?", modelName).
		Order("api_config_id ASC").
		Find(&pricings).Error
	if err != nil {
		return nil, err
	}
	return pricings, nil
}

// FindAll 查找所有定价
func (r *repository) FindAll(ctx context.Context) ([]*Pricing, error) {
	var pricings []*Pricing
	err := r.db.WithContext(ctx).
		Order("api_config_id ASC, model_name ASC").
		Find(&pricings).Error
	if err != nil {
		return nil, err
	}
	return pricings, nil
}

// FindActive 查找所有激活的定价
func (r *repository) FindActive(ctx context.Context) ([]*Pricing, error) {
	var pricings []*Pricing
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("api_config_id ASC, model_name ASC").
		Find(&pricings).Error
	if err != nil {
		return nil, err
	}
	return pricings, nil
}

// List 查询定价列表（支持过滤、排序、分页）
func (r *repository) List(ctx context.Context, filters []query.Filter, sorts []query.Sort, pagination *query.Pagination) ([]*Pricing, int64, error) {
	builder := query.NewBuilder(r.db.WithContext(ctx).Model(&Pricing{}))
	
	// 应用过滤
	builder.ApplyFilters(filters)
	
	// 统计总数
	var total int64
	builder.Count(&total)
	
	// 应用排序和分页
	if len(sorts) == 0 {
		sorts = []query.Sort{
			{Field: "api_config_id", Desc: false},
			{Field: "model_name", Desc: false},
		}
	}
	builder.ApplySort(sorts).ApplyPagination(pagination)
	
	// 查询结果
	var pricings []*Pricing
	err := builder.Find(&pricings)
	if err != nil {
		return nil, 0, err
	}
	
	return pricings, total, nil
}

// BatchCreate 批量创建定价
func (r *repository) BatchCreate(ctx context.Context, pricings []*Pricing) error {
	return r.db.WithContext(ctx).Create(&pricings).Error
}

// UpdateStatus 更新定价状态
func (r *repository) UpdateStatus(ctx context.Context, id uint, isActive bool) error {
	return r.db.WithContext(ctx).Model(&Pricing{}).
		Where("id = ?", id).
		Update("is_active", isActive).Error
}

// CountAll 统计所有定价数量
func (r *repository) CountAll(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Pricing{}).Count(&count).Error
	return count, err
}

// CountByAPIConfig 根据API配置统计定价数量
func (r *repository) CountByAPIConfig(ctx context.Context, apiConfigID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Pricing{}).
		Where("api_config_id = ?", apiConfigID).
		Count(&count).Error
	return count, err
}
