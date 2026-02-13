package cache

import (
	"api-aggregator/backend/pkg/errors"
	"api-aggregator/backend/pkg/logger"
	"api-aggregator/backend/pkg/query"
	"api-aggregator/backend/pkg/utils"
	"context"
)

// Service 缓存服务接口
type Service interface {
	GetCacheStats(ctx context.Context, userID *uint) (*CacheStatsResponse, error)
	GetCacheList(ctx context.Context, req *GetCacheListRequest) (*CacheListResponse, error)
	CleanExpiredCache(ctx context.Context) (*CleanExpiredCacheResponse, error)
	ClearUserCache(ctx context.Context, userID uint) (*ClearUserCacheResponse, error)
	DeleteCache(ctx context.Context, id uint) error
	
	// 缓存查询和存储
	FindByCacheKey(ctx context.Context, cacheKey string) (*RequestCache, error)
	FindByUserAndModel(ctx context.Context, userID uint, model string) ([]*RequestCache, error)
	IncrementHitCount(ctx context.Context, id uint) error
	CreateCacheWithEmbedding(ctx context.Context, cache *RequestCache, embedding []float64) error
}

// service 缓存服务实现
type service struct {
	repo   Repository
	logger logger.Logger
}

// NewService 创建缓存服务
func NewService(repo Repository, logger logger.Logger) Service {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

// GetCacheStats 获取缓存统计
func (s *service) GetCacheStats(ctx context.Context, userID *uint) (*CacheStatsResponse, error) {
	stats, err := s.repo.GetStats(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get cache stats", logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to get cache stats")
	}

	return stats, nil
}

// GetCacheList 获取缓存列表
func (s *service) GetCacheList(ctx context.Context, req *GetCacheListRequest) (*CacheListResponse, error) {
	// 设置默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	// 构建过滤条件
	var filters []query.Filter
	if req.UserID != nil {
		filters = append(filters, query.Filter{
			Field:    "user_id",
			Operator: "=",
			Value:    *req.UserID,
		})
	}
	if req.Model != "" {
		filters = append(filters, query.Filter{
			Field:    "model",
			Operator: "=",
			Value:    req.Model,
		})
	}

	// 构建排序
	sorts := []query.Sort{
		{Field: "created_at", Desc: true},
	}

	// 构建分页
	pagination := &query.Pagination{
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	// 查询缓存列表
	caches, total, err := s.repo.List(ctx, filters, sorts, pagination)
	if err != nil {
		s.logger.Error("Failed to get cache list", logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to get cache list")
	}

	return &CacheListResponse{
		Caches:   ToResponseList(caches),
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// CleanExpiredCache 清理过期缓存
func (s *service) CleanExpiredCache(ctx context.Context) (*CleanExpiredCacheResponse, error) {
	deleted, err := s.repo.DeleteExpired(ctx)
	if err != nil {
		s.logger.Error("Failed to clean expired cache", logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to clean expired cache")
	}

	s.logger.Info("Expired cache cleaned successfully", logger.Int64("deleted", deleted))

	return &CleanExpiredCacheResponse{
		Deleted: deleted,
		Message: "Expired cache cleaned successfully",
	}, nil
}

// ClearUserCache 清除用户缓存
func (s *service) ClearUserCache(ctx context.Context, userID uint) (*ClearUserCacheResponse, error) {
	deleted, err := s.repo.DeleteByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to clear user cache",
			logger.Uint("user_id", userID),
			logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to clear user cache")
	}

	s.logger.Info("User cache cleared successfully",
		logger.Uint("user_id", userID),
		logger.Int64("deleted", deleted))

	return &ClearUserCacheResponse{
		Deleted: deleted,
		Message: "User cache cleared successfully",
	}, nil
}

// DeleteCache 删除缓存
func (s *service) DeleteCache(ctx context.Context, id uint) error {
	// 检查缓存是否存在
	cache, err := s.repo.FindByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get cache", logger.Uint("cache_id", id), logger.Error(err))
		return errors.Wrap(err, 500002, "Failed to get cache")
	}
	if cache == nil {
		return errors.New(404001, "Cache not found")
	}

	// 删除缓存
	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to delete cache",
			logger.Uint("cache_id", id),
			logger.Error(err))
		return errors.Wrap(err, 500002, "Failed to delete cache")
	}

	s.logger.Info("Cache deleted successfully", logger.Uint("cache_id", id))

	return nil
}

// FindByCacheKey 根据缓存键查找缓存
func (s *service) FindByCacheKey(ctx context.Context, cacheKey string) (*RequestCache, error) {
	return s.repo.FindByCacheKey(ctx, cacheKey)
}

// FindByUserAndModel 根据用户和模型查找缓存
func (s *service) FindByUserAndModel(ctx context.Context, userID uint, model string) ([]*RequestCache, error) {
	return s.repo.FindByUserAndModel(ctx, userID, model)
}

// IncrementHitCount 增加缓存命中次数
func (s *service) IncrementHitCount(ctx context.Context, id uint) error {
	return s.repo.IncrementHitCount(ctx, id)
}

// FindSemanticMatch 查找语义匹配的缓存（已废弃，逻辑移到 proxy service）
func (s *service) FindSemanticMatch(ctx context.Context, userID uint, model string, queryText string, threshold float64) (*RequestCache, error) {
	// 获取用户该模型的所有有效缓存
	caches, err := s.repo.FindByUserAndModel(ctx, userID, model)
	if err != nil {
		s.logger.Error("Failed to find caches for semantic matching",
			logger.Uint("user_id", userID),
			logger.String("model", model),
			logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to find caches")
	}

	if len(caches) == 0 {
		return nil, nil
	}

	// 需要先获取查询文本的 embedding（由调用方提供）
	// 这里只是查找已有 embedding 的缓存并计算相似度
	var bestMatch *RequestCache
	var bestSimilarity float64

	for _, cache := range caches {
		if !cache.HasEmbedding() {
			continue
		}

		// 解析缓存的 embedding
		cacheEmbedding, err := utils.JSONToVector(cache.Embedding)
		if err != nil {
			s.logger.Warn("Failed to parse cache embedding",
				logger.Uint("cache_id", cache.ID),
				logger.Error(err))
			continue
		}

		// 注意：这里需要查询文本的 embedding，应该由调用方传入
		// 暂时返回 nil，实际使用时需要在上层集成 embedding 服务
		_ = cacheEmbedding
	}

	if bestMatch != nil && bestSimilarity >= threshold {
		// 增加命中次数
		if err := s.repo.IncrementHitCount(ctx, bestMatch.ID); err != nil {
			s.logger.Warn("Failed to increment hit count",
				logger.Uint("cache_id", bestMatch.ID),
				logger.Error(err))
		}
		return bestMatch, nil
	}

	return nil, nil
}

// CreateCacheWithEmbedding 创建带 embedding 的缓存
func (s *service) CreateCacheWithEmbedding(ctx context.Context, cache *RequestCache, embedding []float64) error {
	// 将 embedding 转换为 JSON 字符串
	if embedding != nil {
		embeddingJSON, err := utils.VectorToJSON(embedding)
		if err != nil {
			s.logger.Error("Failed to convert embedding to JSON",
				logger.Error(err))
			return errors.Wrap(err, 500002, "Failed to convert embedding")
		}
		cache.Embedding = embeddingJSON
	}

	// 创建缓存
	if err := s.repo.Create(ctx, cache); err != nil {
		s.logger.Error("Failed to create cache",
			logger.Uint("user_id", cache.UserID),
			logger.Error(err))
		return errors.Wrap(err, 500002, "Failed to create cache")
	}

	s.logger.Info("Cache created successfully",
		logger.Uint("cache_id", cache.ID),
		logger.Uint("user_id", cache.UserID),
		logger.Bool("has_embedding", embedding != nil))

	return nil
}
