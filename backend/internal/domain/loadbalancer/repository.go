package loadbalancer

import (
	"api-aggregator/backend/pkg/query"
	"context"

	"gorm.io/gorm"
)

// Repository 负载均衡配置仓储接口
type Repository interface {
	Create(ctx context.Context, config *LoadBalancerConfig) error
	Update(ctx context.Context, config *LoadBalancerConfig) error
	Delete(ctx context.Context, id uint) error
	FindByID(ctx context.Context, id uint) (*LoadBalancerConfig, error)
	FindByModel(ctx context.Context, modelName string) (*LoadBalancerConfig, error)
	List(ctx context.Context, filter *ConfigFilter, opts *query.Options) ([]*LoadBalancerConfig, int64, error)
	FindAll(ctx context.Context) ([]*LoadBalancerConfig, error)
	ExistsByModel(ctx context.Context, modelName string) (bool, error)
}

type repository struct {
	db *gorm.DB
}

// NewRepository 创建负载均衡配置仓储实例
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// Create 创建负载均衡配置
func (r *repository) Create(ctx context.Context, config *LoadBalancerConfig) error {
	return r.db.WithContext(ctx).Create(config).Error
}

// Update 更新负载均衡配置
func (r *repository) Update(ctx context.Context, config *LoadBalancerConfig) error {
	return r.db.WithContext(ctx).Save(config).Error
}

// Delete 删除负载均衡配置
func (r *repository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&LoadBalancerConfig{}, id).Error
}

// FindByID 根据ID查找负载均衡配置
func (r *repository) FindByID(ctx context.Context, id uint) (*LoadBalancerConfig, error) {
	var config LoadBalancerConfig
	err := r.db.WithContext(ctx).First(&config, id).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// FindByModel 根据模型名称查找负载均衡配置
func (r *repository) FindByModel(ctx context.Context, modelName string) (*LoadBalancerConfig, error) {
	var config LoadBalancerConfig
	err := r.db.WithContext(ctx).
		Where("model_name = ? AND is_active = ?", modelName, true).
		First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// List 查询负载均衡配置列表
func (r *repository) List(ctx context.Context, filter *ConfigFilter, opts *query.Options) ([]*LoadBalancerConfig, int64, error) {
	var configs []*LoadBalancerConfig
	var total int64

	db := r.db.WithContext(ctx).Model(&LoadBalancerConfig{})

	// 应用过滤器
	if filter != nil {
		if filter.ModelName != nil {
			db = db.Where("model_name = ?", *filter.ModelName)
		}
		if filter.Strategy != nil {
			db = db.Where("strategy = ?", *filter.Strategy)
		}
		if filter.IsActive != nil {
			db = db.Where("is_active = ?", *filter.IsActive)
		}
	}

	// 计算总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用查询选项
	if opts != nil {
		db = query.ApplyOptions(db, opts)
	}

	// 查询数据
	if err := db.Find(&configs).Error; err != nil {
		return nil, 0, err
	}

	return configs, total, nil
}

// FindAll 查询所有负载均衡配置
func (r *repository) FindAll(ctx context.Context) ([]*LoadBalancerConfig, error) {
	var configs []*LoadBalancerConfig
	err := r.db.WithContext(ctx).Find(&configs).Error
	return configs, err
}

// ExistsByModel 检查模型是否已存在配置
func (r *repository) ExistsByModel(ctx context.Context, modelName string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&LoadBalancerConfig{}).
		Where("model_name = ?", modelName).
		Count(&count).Error
	return count > 0, err
}
