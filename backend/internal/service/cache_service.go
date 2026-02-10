package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"api-aggregator/backend/internal/adapter"
	"api-aggregator/backend/internal/models"

	"gorm.io/gorm"
)

// CacheService 处理请求缓存逻辑
type CacheService struct {
	db               *gorm.DB
	embeddingService EmbeddingService
}

// NewCacheService 创建缓存服务实例
func NewCacheService(db *gorm.DB, embeddingService EmbeddingService) *CacheService {
	return &CacheService{
		db:               db,
		embeddingService: embeddingService,
	}
}

// EmbeddingService 向量嵌入服务接口
type EmbeddingService interface {
	GetEmbedding(ctx context.Context, text string) ([]float64, error)
	BatchGetEmbeddings(ctx context.Context, texts []string) ([][]float64, error)
	CosineSimilarity(vec1, vec2 []float64) float64
}

// CacheConfig 缓存配置
type CacheConfig struct {
	Enabled       bool          // 是否启用缓存
	TTL           time.Duration // 缓存过期时间
	SemanticMatch bool          // 是否启用语义匹配
	Threshold     float64       // 语义相似度阈值 (0-1)
}

// GetCachedResponse 获取缓存的响应
// 返回: response, isCached, error
func (s *CacheService) GetCachedResponse(
	ctx context.Context,
	req *adapter.ChatRequest,
	config *CacheConfig,
) (*adapter.ChatResponse, bool, error) {
	if !config.Enabled {
		return nil, false, nil
	}

	// 生成请求的哈希键
	cacheKey := s.generateCacheKey(req)

	// 1. 先尝试精确匹配
	cached, err := s.getExactMatch(ctx, cacheKey, config.TTL)
	if err == nil && cached != nil {
		return cached, true, nil
	}

	// 2. 如果启用语义匹配，尝试语义相似匹配
	if config.SemanticMatch {
		cached, err = s.getSemanticMatch(ctx, req, config.Threshold, config.TTL)
		if err == nil && cached != nil {
			return cached, true, nil
		}
	}

	return nil, false, nil
}

// SaveCachedResponse 保存响应到缓存
func (s *CacheService) SaveCachedResponse(
	ctx context.Context,
	req *adapter.ChatRequest,
	resp *adapter.ChatResponse,
	config *CacheConfig,
	userID uint,
) error {
	if !config.Enabled {
		return nil
	}

	cacheKey := s.generateCacheKey(req)
	
	// 提取最后一条用户消息作为查询文本（用于语义匹配）
	queryText := s.extractQueryText(req)

	// 序列化请求和响应
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	respJSON, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	// 生成查询文本的向量嵌入
	var embedding []float64
	if config.SemanticMatch && s.embeddingService != nil && queryText != "" {
		embedding, err = s.embeddingService.GetEmbedding(ctx, queryText)
		if err != nil {
			// 向量生成失败不影响缓存保存，只是不能进行语义匹配
			fmt.Printf("Warning: failed to generate embedding: %v\n", err)
		}
	}

	// 将向量序列化为 JSON
	var embeddingJSON string
	if len(embedding) > 0 {
		embBytes, _ := json.Marshal(embedding)
		embeddingJSON = string(embBytes)
	}

	cache := &models.RequestCache{
		UserID:      userID,
		CacheKey:    cacheKey,
		QueryText:   queryText,
		Embedding:   embeddingJSON,
		Model:       req.Model,
		Request:     string(reqJSON),
		Response:    string(respJSON),
		TokensSaved: resp.Usage.TotalTokens,
		HitCount:    0,
		ExpiresAt:   time.Now().Add(config.TTL),
	}

	return s.db.WithContext(ctx).Create(cache).Error
}

// GetCacheStats 获取缓存统计信息
func (s *CacheService) GetCacheStats(ctx context.Context, userID uint) (*CacheStats, error) {
	var stats CacheStats

	// 总缓存命中次数
	err := s.db.WithContext(ctx).
		Model(&models.RequestCache{}).
		Select("COALESCE(SUM(hit_count), 0) as total_hits").
		Where("user_id = ?", userID).
		Scan(&stats.TotalHits).Error
	if err != nil {
		return nil, err
	}

	// 节省的总 tokens
	err = s.db.WithContext(ctx).
		Model(&models.RequestCache{}).
		Select("COALESCE(SUM(tokens_saved * hit_count), 0) as tokens_saved").
		Where("user_id = ?", userID).
		Scan(&stats.TokensSaved).Error
	if err != nil {
		return nil, err
	}

	// 缓存条目数
	err = s.db.WithContext(ctx).
		Model(&models.RequestCache{}).
		Where("user_id = ? AND expires_at > ?", userID, time.Now()).
		Count(&stats.CacheEntries).Error
	if err != nil {
		return nil, err
	}

	return &stats, nil
}

// CleanExpiredCache 清理过期的缓存
func (s *CacheService) CleanExpiredCache(ctx context.Context) error {
	return s.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&models.RequestCache{}).Error
}

// generateCacheKey 生成请求的缓存键
func (s *CacheService) generateCacheKey(req *adapter.ChatRequest) string {
	// 将请求的关键部分序列化
	data := struct {
		Model       string
		Messages    []adapter.Message
		Temperature float64
		MaxTokens   int
		TopP        float64
	}{
		Model:       req.Model,
		Messages:    req.Messages,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		TopP:        req.TopP,
	}

	jsonData, _ := json.Marshal(data)
	hash := md5.Sum(jsonData)
	return hex.EncodeToString(hash[:])
}

// extractQueryText 提取查询文本（最后一条用户消息）
func (s *CacheService) extractQueryText(req *adapter.ChatRequest) string {
	// 从后往前找第一条用户消息
	for i := len(req.Messages) - 1; i >= 0; i-- {
		if req.Messages[i].Role == "user" {
			return req.Messages[i].Content
		}
	}
	return ""
}

// getExactMatch 精确匹配缓存
func (s *CacheService) getExactMatch(
	ctx context.Context,
	cacheKey string,
	ttl time.Duration,
) (*adapter.ChatResponse, error) {
	var cache models.RequestCache
	
	err := s.db.WithContext(ctx).
		Where("cache_key = ? AND expires_at > ?", cacheKey, time.Now()).
		First(&cache).Error
	
	if err != nil {
		return nil, err
	}

	// 增加命中次数
	s.db.WithContext(ctx).
		Model(&cache).
		UpdateColumn("hit_count", gorm.Expr("hit_count + ?", 1))

	// 反序列化响应
	var resp adapter.ChatResponse
	if err := json.Unmarshal([]byte(cache.Response), &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// getSemanticMatch 语义匹配缓存（使用向量相似度）
func (s *CacheService) getSemanticMatch(
	ctx context.Context,
	req *adapter.ChatRequest,
	threshold float64,
	ttl time.Duration,
) (*adapter.ChatResponse, error) {
	queryText := s.extractQueryText(req)
	if queryText == "" {
		return nil, fmt.Errorf("no query text found")
	}

	// 如果没有嵌入服务，回退到简单文本匹配
	if s.embeddingService == nil {
		return s.getSemanticMatchSimple(ctx, req, threshold, ttl)
	}

	// 生成查询文本的向量
	queryEmbedding, err := s.embeddingService.GetEmbedding(ctx, queryText)
	if err != nil {
		// 向量生成失败，回退到简单匹配
		fmt.Printf("Warning: failed to generate query embedding, falling back to simple match: %v\n", err)
		return s.getSemanticMatchSimple(ctx, req, threshold, ttl)
	}

	// 获取同模型的所有有效缓存（包含向量）
	var caches []models.RequestCache
	err = s.db.WithContext(ctx).
		Where("model = ? AND expires_at > ? AND embedding IS NOT NULL AND embedding != ''", req.Model, time.Now()).
		Order("created_at DESC").
		Limit(100). // 限制查询数量
		Find(&caches).Error
	
	if err != nil {
		return nil, err
	}

	if len(caches) == 0 {
		return nil, fmt.Errorf("no cached embeddings found")
	}

	// 计算相似度，找到最相似的缓存
	var bestMatch *models.RequestCache
	var bestScore float64

	for i := range caches {
		// 反序列化缓存的向量
		var cachedEmbedding []float64
		if err := json.Unmarshal([]byte(caches[i].Embedding), &cachedEmbedding); err != nil {
			continue
		}

		// 计算余弦相似度
		score := s.embeddingService.CosineSimilarity(queryEmbedding, cachedEmbedding)
		
		if score > bestScore && score >= threshold {
			bestScore = score
			bestMatch = &caches[i]
		}
	}

	if bestMatch == nil {
		return nil, fmt.Errorf("no semantic match found (best score: %.3f, threshold: %.3f)", bestScore, threshold)
	}

	// 增加命中次数
	s.db.WithContext(ctx).
		Model(bestMatch).
		UpdateColumn("hit_count", gorm.Expr("hit_count + ?", 1))

	// 反序列化响应
	var resp adapter.ChatResponse
	if err := json.Unmarshal([]byte(bestMatch.Response), &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// getSemanticMatchSimple 简单的语义匹配（回退方案）
func (s *CacheService) getSemanticMatchSimple(
	ctx context.Context,
	req *adapter.ChatRequest,
	threshold float64,
	ttl time.Duration,
) (*adapter.ChatResponse, error) {
	queryText := s.extractQueryText(req)
	if queryText == "" {
		return nil, fmt.Errorf("no query text found")
	}

	// 获取同模型的所有有效缓存
	var caches []models.RequestCache
	err := s.db.WithContext(ctx).
		Where("model = ? AND expires_at > ?", req.Model, time.Now()).
		Order("created_at DESC").
		Limit(100).
		Find(&caches).Error
	
	if err != nil {
		return nil, err
	}

	// 使用简单的 Jaccard 相似度
	var bestMatch *models.RequestCache
	var bestScore float64

	for i := range caches {
		score := s.calculateSimilarity(queryText, caches[i].QueryText)
		if score > bestScore && score >= threshold {
			bestScore = score
			bestMatch = &caches[i]
		}
	}

	if bestMatch == nil {
		return nil, fmt.Errorf("no semantic match found")
	}

	// 增加命中次数
	s.db.WithContext(ctx).
		Model(bestMatch).
		UpdateColumn("hit_count", gorm.Expr("hit_count + ?", 1))

	// 反序列化响应
	var resp adapter.ChatResponse
	if err := json.Unmarshal([]byte(bestMatch.Response), &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// calculateSimilarity 计算两个文本的相似度（简单实现，作为回退方案）
// 使用 Jaccard 相似度
func (s *CacheService) calculateSimilarity(text1, text2 string) float64 {
	words1 := s.tokenize(text1)
	words2 := s.tokenize(text2)

	if len(words1) == 0 || len(words2) == 0 {
		return 0
	}

	// 计算交集和并集
	intersection := 0
	union := make(map[string]bool)

	for word := range words1 {
		union[word] = true
	}
	for word := range words2 {
		union[word] = true
		if words1[word] {
			intersection++
		}
	}

	return float64(intersection) / float64(len(union))
}

// tokenize 简单的分词
func (s *CacheService) tokenize(text string) map[string]bool {
	words := make(map[string]bool)
	
	// 简单按空格分词（生产环境应使用更好的分词器）
	word := ""
	for _, char := range text {
		if char == ' ' || char == '\n' || char == '\t' {
			if word != "" {
				words[word] = true
				word = ""
			}
		} else {
			word += string(char)
		}
	}
	if word != "" {
		words[word] = true
	}

	return words
}

// ClearUserCache 清除用户的所有缓存
func (s *CacheService) ClearUserCache(ctx context.Context, userID uint) error {
	return s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&models.RequestCache{}).Error
}

// CacheStats 缓存统计
type CacheStats struct {
	TotalHits    int64 `json:"total_hits"`
	TokensSaved  int64 `json:"tokens_saved"`
	CacheEntries int64 `json:"cache_entries"`
}
