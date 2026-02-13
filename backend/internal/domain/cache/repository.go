package cache

import (
	"api-aggregator/backend/pkg/query"
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

// Repository 缓存仓储接口
type Repository interface {
	Create(ctx context.Context, cache *RequestCache) error
	Update(ctx context.Context, cache *RequestCache) error
	Delete(ctx context.Context, id uint) error
	FindByID(ctx context.Context, id uint) (*RequestCache, error)
	FindByCacheKey(ctx context.Context, cacheKey string) (*RequestCache, error)
	FindByUserID(ctx context.Context, userID uint) ([]*RequestCache, error)
	FindByModel(ctx context.Context, model string, limit int) ([]*RequestCache, error)
	FindByUserAndModel(ctx context.Context, userID uint, model string) ([]*RequestCache, error)
	List(ctx context.Context, filters []query.Filter, sorts []query.Sort, pagination *query.Pagination) ([]*RequestCache, int64, error)
	IncrementHitCount(ctx context.Context, id uint) error
	GetStats(ctx context.Context, userID *uint) (*CacheStatsResponse, error)
	DeleteExpired(ctx context.Context) (int64, error)
	DeleteByUserID(ctx context.Context, userID uint) (int64, error)
	CountAll(ctx context.Context) (int64, error)
	CountByUserID(ctx context.Context, userID uint) (int64, error)
}

// repository 缓存仓储实现
type repository struct {
	db *gorm.DB
}

// NewRepository 创建缓存仓储
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// Create 创建缓存
func (r *repository) Create(ctx context.Context, cache *RequestCache) error {
	return r.db.WithContext(ctx).Create(cache).Error
}

// Update 更新缓存
func (r *repository) Update(ctx context.Context, cache *RequestCache) error {
	return r.db.WithContext(ctx).Save(cache).Error
}

// Delete 删除缓存
func (r *repository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&RequestCache{}, id).Error
}

// FindByID 根据ID查找缓存
func (r *repository) FindByID(ctx context.Context, id uint) (*RequestCache, error) {
	var cache RequestCache
	err := r.db.WithContext(ctx).First(&cache, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &cache, nil
}

// FindByCacheKey 根据缓存键查找缓存
func (r *repository) FindByCacheKey(ctx context.Context, cacheKey string) (*RequestCache, error) {
	var cache RequestCache
	err := r.db.WithContext(ctx).
		Where("cache_key = ? AND expires_at > ?", cacheKey, time.Now()).
		First(&cache).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &cache, nil
}

// FindByUserID 根据用户ID查找所有缓存
func (r *repository) FindByUserID(ctx context.Context, userID uint) ([]*RequestCache, error) {
	var caches []*RequestCache
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND expires_at > ?", userID, time.Now()).
		Order("created_at DESC").
		Find(&caches).Error
	if err != nil {
		return nil, err
	}
	return caches, nil
}

// FindByModel 根据模型查找缓存
func (r *repository) FindByModel(ctx context.Context, model string, limit int) ([]*RequestCache, error) {
	var caches []*RequestCache
	err := r.db.WithContext(ctx).
		Where("model = ? AND expires_at > ?", model, time.Now()).
		Order("created_at DESC").
		Limit(limit).
		Find(&caches).Error
	if err != nil {
		return nil, err
	}
	return caches, nil
}

// List 查询缓存列表（支持过滤、排序、分页）
func (r *repository) List(ctx context.Context, filters []query.Filter, sorts []query.Sort, pagination *query.Pagination) ([]*RequestCache, int64, error) {
	builder := query.NewBuilder(r.db.WithContext(ctx).Model(&RequestCache{}))
	
	// 只查询未过期的缓存
	builder.Where("expires_at > ?", time.Now())
	
	// 应用过滤
	builder.ApplyFilters(filters)
	
	// 统计总数
	var total int64
	builder.Count(&total)
	
	// 应用排序和分页
	if len(sorts) == 0 {
		sorts = []query.Sort{
			{Field: "created_at", Desc: true},
		}
	}
	builder.ApplySort(sorts).ApplyPagination(pagination)
	
	// 查询结果
	var caches []*RequestCache
	err := builder.Find(&caches)
	if err != nil {
		return nil, 0, err
	}
	
	return caches, total, nil
}

// IncrementHitCount 增加命中次数
func (r *repository) IncrementHitCount(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&RequestCache{}).
		Where("id = ?", id).
		UpdateColumn("hit_count", gorm.Expr("hit_count + ?", 1)).Error
}

// GetStats 获取缓存统计
func (r *repository) GetStats(ctx context.Context, userID *uint) (*CacheStatsResponse, error) {
	query := r.db.WithContext(ctx).Model(&RequestCache{}).
		Where("expires_at > ?", time.Now())
	
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	
	var stats struct {
		TotalHits    int64
		TokensSaved  int64
		CacheEntries int64
	}
	
	// 总命中次数
	query.Select("COALESCE(SUM(hit_count), 0)").Scan(&stats.TotalHits)
	
	// 节省的tokens
	query.Select("COALESCE(SUM(tokens_saved * hit_count), 0)").Scan(&stats.TokensSaved)
	
	// 缓存条目数
	query.Count(&stats.CacheEntries)
	
	return &CacheStatsResponse{
		TotalHits:    stats.TotalHits,
		TokensSaved:  stats.TokensSaved,
		CacheEntries: stats.CacheEntries,
	}, nil
}

// DeleteExpired 删除过期缓存
func (r *repository) DeleteExpired(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&RequestCache{})
	return result.RowsAffected, result.Error
}

// DeleteByUserID 删除用户的所有缓存
func (r *repository) DeleteByUserID(ctx context.Context, userID uint) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&RequestCache{})
	return result.RowsAffected, result.Error
}

// CountAll 统计所有缓存数量
func (r *repository) CountAll(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&RequestCache{}).
		Where("expires_at > ?", time.Now()).
		Count(&count).Error
	return count, err
}

// CountByUserID 根据用户ID统计缓存数量
func (r *repository) CountByUserID(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&RequestCache{}).
		Where("user_id = ? AND expires_at > ?", userID, time.Now()).
		Count(&count).Error
	return count, err
}


// FindByUserAndModel 根据用户ID和模型查找所有有效缓存（用于语义匹配）
func (r *repository) FindByUserAndModel(ctx context.Context, userID uint, model string) ([]*RequestCache, error) {
	var caches []*RequestCache
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND model = ? AND expires_at > ? AND embedding IS NOT NULL AND embedding != ''", 
			userID, model, time.Now()).
		Order("created_at DESC").
		Limit(100). // 限制最多检查100条，避免性能问题
		Find(&caches).Error
	if err != nil {
		return nil, err
	}
	return caches, nil
}
