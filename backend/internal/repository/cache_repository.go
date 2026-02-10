package repository

import (
	"context"
	"time"

	"api-aggregator/backend/internal/models"

	"gorm.io/gorm"
)

// CacheRepository 缓存数据访问层
type CacheRepository struct {
	db *gorm.DB
}

// NewCacheRepository 创建缓存仓库实例
func NewCacheRepository(db *gorm.DB) *CacheRepository {
	return &CacheRepository{db: db}
}

// FindByCacheKey 根据缓存键查找
func (r *CacheRepository) FindByCacheKey(ctx context.Context, cacheKey string) (*models.RequestCache, error) {
	var cache models.RequestCache
	err := r.db.WithContext(ctx).
		Where("cache_key = ? AND expires_at > ?", cacheKey, time.Now()).
		First(&cache).Error
	if err != nil {
		return nil, err
	}
	return &cache, nil
}

// FindByModel 根据模型查找有效缓存
func (r *CacheRepository) FindByModel(ctx context.Context, model string, limit int) ([]models.RequestCache, error) {
	var caches []models.RequestCache
	err := r.db.WithContext(ctx).
		Where("model = ? AND expires_at > ?", model, time.Now()).
		Order("created_at DESC").
		Limit(limit).
		Find(&caches).Error
	return caches, err
}

// Create 创建缓存记录
func (r *CacheRepository) Create(ctx context.Context, cache *models.RequestCache) error {
	return r.db.WithContext(ctx).Create(cache).Error
}

// IncrementHitCount 增加命中次数
func (r *CacheRepository) IncrementHitCount(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&models.RequestCache{}).
		Where("id = ?", id).
		UpdateColumn("hit_count", gorm.Expr("hit_count + ?", 1)).Error
}

// DeleteExpired 删除过期缓存
func (r *CacheRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&models.RequestCache{}).Error
}

// GetStatsByUser 获取用户的缓存统计
func (r *CacheRepository) GetStatsByUser(ctx context.Context, userID uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总命中次数
	var totalHits int64
	err := r.db.WithContext(ctx).
		Model(&models.RequestCache{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(hit_count), 0)").
		Scan(&totalHits).Error
	if err != nil {
		return nil, err
	}
	stats["total_hits"] = totalHits

	// 节省的 tokens
	var tokensSaved int64
	err = r.db.WithContext(ctx).
		Model(&models.RequestCache{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(tokens_saved * hit_count), 0)").
		Scan(&tokensSaved).Error
	if err != nil {
		return nil, err
	}
	stats["tokens_saved"] = tokensSaved

	// 缓存条目数
	var count int64
	err = r.db.WithContext(ctx).
		Model(&models.RequestCache{}).
		Where("user_id = ? AND expires_at > ?", userID, time.Now()).
		Count(&count).Error
	if err != nil {
		return nil, err
	}
	stats["cache_entries"] = count

	return stats, nil
}

// DeleteByUser 删除用户的所有缓存
func (r *CacheRepository) DeleteByUser(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&models.RequestCache{}).Error
}

// DeleteByCacheKey 删除指定缓存
func (r *CacheRepository) DeleteByCacheKey(ctx context.Context, cacheKey string) error {
	return r.db.WithContext(ctx).
		Where("cache_key = ?", cacheKey).
		Delete(&models.RequestCache{}).Error
}
