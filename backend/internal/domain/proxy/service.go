package proxy

import (
	"api-aggregator/backend/internal/adapter"
	"api-aggregator/backend/internal/domain/apiconfig"
	"api-aggregator/backend/internal/domain/cache"
	"api-aggregator/backend/internal/domain/loadbalancer"
	"api-aggregator/backend/internal/domain/log"
	"api-aggregator/backend/internal/domain/pricing"
	"api-aggregator/backend/internal/domain/quota"
	"api-aggregator/backend/pkg/embedding"
	"api-aggregator/backend/pkg/errors"
	"api-aggregator/backend/pkg/logger"
	"api-aggregator/backend/pkg/runtime"
	"api-aggregator/backend/pkg/utils"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Service 代理服务接口
type Service interface {
	ChatCompletions(ctx context.Context, req *ProxyRequest) (*adapter.ChatResponse, error)
	ChatCompletionsStream(ctx context.Context, req *ProxyRequest) (*http.Response, error)
	SetEmbeddingClient(client *embedding.Client)
}

type service struct {
	adapterFactory  *adapter.Factory
	apiConfigRepo   apiconfig.Repository
	loadBalancerSvc loadbalancer.Service
	cacheService    cache.Service
	quotaService    quota.Service
	pricingService  pricing.Service
	logService      log.Service
	runtimeConfig   *runtime.Manager
	embeddingClient *embedding.Client
	logger          logger.Logger
}

// NewService 创建代理服务
func NewService(
	adapterFactory *adapter.Factory,
	apiConfigRepo apiconfig.Repository,
	loadBalancerSvc loadbalancer.Service,
	cacheService cache.Service,
	quotaService quota.Service,
	pricingService pricing.Service,
	logService log.Service,
	runtimeConfig *runtime.Manager,
	logger logger.Logger,
) Service {
	return &service{
		adapterFactory:  adapterFactory,
		apiConfigRepo:   apiConfigRepo,
		loadBalancerSvc: loadBalancerSvc,
		cacheService:    cacheService,
		quotaService:    quotaService,
		pricingService:  pricingService,
		logService:      logService,
		runtimeConfig:   runtimeConfig,
		logger:          logger,
	}
}

// SetEmbeddingClient 设置 embedding 客户端
func (s *service) SetEmbeddingClient(client *embedding.Client) {
	s.embeddingClient = client
}

// ChatCompletions 处理聊天补全请求
func (s *service) ChatCompletions(ctx context.Context, req *ProxyRequest) (*adapter.ChatResponse, error) {
	startTime := time.Now()
	
	// 1. 检查配额
	if err := s.checkQuota(ctx, req.UserID); err != nil {
		return nil, err
	}

	// 2. 生成缓存键
	cacheKey := s.generateCacheKey(req.ChatRequest)
	
	// 3. 查询缓存
	if s.runtimeConfig.Get().IsCacheEnabled() {
		cachedResp, err := s.checkCache(ctx, req.UserID, req.Model, cacheKey, req.ChatRequest)
		if err != nil {
			s.logger.Warn("Failed to check cache", logger.Error(err))
		} else if cachedResp != nil {
			// 缓存命中
			s.logger.Info("Cache hit",
				logger.Uint("user_id", req.UserID),
				logger.String("model", req.Model),
				logger.String("cache_key", cacheKey))
			
			cachedResp.Cached = true
			return cachedResp, nil
		}
	}

	// 4. 选择 API 配置（负载均衡）
	apiConfig, err := s.selectAPIConfig(ctx, req.Model)
	if err != nil {
		return nil, err
	}

	// 5. 创建适配器并调用
	adapterInstance, err := s.adapterFactory.CreateAdapter(apiConfig)
	if err != nil {
		return nil, errors.Wrap(err, 500003, "Failed to create adapter")
	}

	// 6. 调用上游 API
	resp, err := adapterInstance.Call(ctx, req.ChatRequest)
	if err != nil {
		// 记录失败日志
		s.logRequest(ctx, req, apiConfig.ID, 0, time.Since(startTime), err)
		return nil, errors.Wrap(err, 500004, "Failed to call upstream API")
	}

	// 7. 计算费用并扣除配额
	cost, err := s.calculateAndDeductCost(ctx, req.UserID, apiConfig.ID, req.Model, resp.Usage)
	if err != nil {
		s.logger.Warn("Failed to calculate cost", logger.Error(err))
	}

	// 8. 记录请求日志
	s.logRequest(ctx, req, apiConfig.ID, resp.Usage.TotalTokens, time.Since(startTime), nil)

	// 9. 存储到缓存
	if s.runtimeConfig.Get().IsCacheEnabled() && !req.Stream {
		go s.storeCache(context.Background(), req.UserID, req.Model, cacheKey, req.ChatRequest, resp, cost)
	}

	return resp, nil
}

// ChatCompletionsStream 处理流式聊天补全请求
func (s *service) ChatCompletionsStream(ctx context.Context, req *ProxyRequest) (*http.Response, error) {
	// 流式请求不使用缓存
	
	// 1. 检查配额
	if err := s.checkQuota(ctx, req.UserID); err != nil {
		return nil, err
	}

	// 2. 选择 API 配置
	apiConfig, err := s.selectAPIConfig(ctx, req.Model)
	if err != nil {
		return nil, err
	}

	// 3. 创建适配器并调用
	adapterInstance, err := s.adapterFactory.CreateAdapter(apiConfig)
	if err != nil {
		return nil, errors.Wrap(err, 500003, "Failed to create adapter")
	}

	// 4. 调用上游 API（流式）
	resp, err := adapterInstance.CallStream(ctx, req.ChatRequest)
	if err != nil {
		return nil, errors.Wrap(err, 500004, "Failed to call upstream API")
	}

	// 注意：流式请求的日志记录和配额扣除需要在流读取完成后处理
	// 这里只返回响应，由调用方处理

	return resp, nil
}

// checkQuota 检查用户配额
func (s *service) checkQuota(ctx context.Context, userID uint) error {
	quotaInfo, err := s.quotaService.GetQuotaInfo(ctx, userID)
	if err != nil {
		return errors.Wrap(err, 500005, "Failed to check quota")
	}

	if quotaInfo.UsedQuota >= quotaInfo.TotalQuota {
		return errors.New(403001, "Quota exceeded")
	}

	return nil
}

// generateCacheKey 生成缓存键
func (s *service) generateCacheKey(req *adapter.ChatRequest) string {
	// 将请求序列化为 JSON
	data, _ := json.Marshal(map[string]interface{}{
		"model":       req.Model,
		"messages":    req.Messages,
		"temperature": req.Temperature,
		"top_p":       req.TopP,
		"max_tokens":  req.MaxTokens,
	})
	
	// 计算 MD5 哈希
	hash := md5.Sum(data)
	return fmt.Sprintf("%x", hash)
}

// checkCache 检查缓存
func (s *service) checkCache(ctx context.Context, userID uint, model string, cacheKey string, req *adapter.ChatRequest) (*adapter.ChatResponse, error) {
	// 1. 精确匹配查询
	cachedItem, err := s.cacheService.FindByCacheKey(ctx, cacheKey)
	if err == nil && cachedItem != nil {
		// 解析响应
		var resp adapter.ChatResponse
		if err := json.Unmarshal([]byte(cachedItem.Response), &resp); err == nil {
			return &resp, nil
		}
	}

	// 2. 语义匹配查询（如果启用）
	if s.runtimeConfig.Get().IsSemanticEnabled() && s.embeddingClient != nil {
		return s.semanticCacheMatch(ctx, userID, model, req)
	}

	return nil, nil
}

// semanticCacheMatch 语义缓存匹配
func (s *service) semanticCacheMatch(ctx context.Context, userID uint, model string, req *adapter.ChatRequest) (*adapter.ChatResponse, error) {
	// 提取查询文本（最后一条用户消息）
	queryText := s.extractQueryText(req.Messages)
	if queryText == "" {
		return nil, nil
	}

	// 获取查询文本的 embedding
	queryEmbedding, err := s.embeddingClient.Embed(ctx, queryText)
	if err != nil {
		s.logger.Warn("Failed to get embedding", logger.Error(err))
		return nil, nil
	}

	// 查找语义匹配的缓存
	threshold := s.runtimeConfig.Get().GetSemanticThreshold()
	cachedItem, err := s.findSemanticMatch(ctx, userID, model, queryEmbedding, threshold)
	if err != nil || cachedItem == nil {
		return nil, nil
	}

	// 解析响应
	var resp adapter.ChatResponse
	if err := json.Unmarshal([]byte(cachedItem.Response), &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// findSemanticMatch 查找语义匹配的缓存
func (s *service) findSemanticMatch(ctx context.Context, userID uint, model string, queryEmbedding []float64, threshold float64) (*cache.RequestCache, error) {
	// 获取用户该模型的所有有效缓存
	caches, err := s.cacheService.FindByUserAndModel(ctx, userID, model)
	if err != nil || len(caches) == 0 {
		return nil, err
	}

	var bestMatch *cache.RequestCache
	var bestSimilarity float64

	for _, cachedItem := range caches {
		if !cachedItem.HasEmbedding() {
			continue
		}

		// 解析缓存的 embedding
		cacheEmbedding, err := utils.JSONToVector(cachedItem.Embedding)
		if err != nil {
			continue
		}

		// 计算余弦相似度
		similarity, err := utils.CosineSimilarity(queryEmbedding, cacheEmbedding)
		if err != nil {
			continue
		}

		if similarity > bestSimilarity {
			bestSimilarity = similarity
			bestMatch = cachedItem
		}
	}

	if bestMatch != nil && bestSimilarity >= threshold {
		// 增加命中次数
		s.cacheService.IncrementHitCount(ctx, bestMatch.ID)
		return bestMatch, nil
	}

	return nil, nil
}

// extractQueryText 提取查询文本
func (s *service) extractQueryText(messages []adapter.Message) string {
	// 提取最后一条用户消息
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			return messages[i].Content
		}
	}
	return ""
}

// storeCache 存储缓存
func (s *service) storeCache(ctx context.Context, userID uint, model string, cacheKey string, req *adapter.ChatRequest, resp *adapter.ChatResponse, tokensSaved int) {
	// 序列化请求和响应
	reqJSON, _ := json.Marshal(req)
	respJSON, _ := json.Marshal(resp)

	// 提取查询文本
	queryText := s.extractQueryText(req.Messages)

	// 创建缓存记录
	cacheItem := &cache.RequestCache{
		UserID:      userID,
		CacheKey:    cacheKey,
		QueryText:   queryText,
		Model:       model,
		Request:     string(reqJSON),
		Response:    string(respJSON),
		TokensSaved: tokensSaved,
		HitCount:    0,
		ExpiresAt:   time.Now().Add(s.runtimeConfig.Get().GetCacheTTL()),
	}

	// 如果启用 embedding，生成向量
	var embeddingVec []float64
	if s.runtimeConfig.Get().IsEmbeddingEnabled() && s.embeddingClient != nil && queryText != "" {
		vec, err := s.embeddingClient.Embed(ctx, queryText)
		if err != nil {
			s.logger.Warn("Failed to generate embedding", logger.Error(err))
		} else {
			embeddingVec = vec
		}
	}

	// 存储缓存
	if err := s.cacheService.CreateCacheWithEmbedding(ctx, cacheItem, embeddingVec); err != nil {
		s.logger.Warn("Failed to store cache", logger.Error(err))
	}
}

// selectAPIConfig 选择 API 配置（负载均衡）
func (s *service) selectAPIConfig(ctx context.Context, model string) (*apiconfig.APIConfig, error) {
	// 获取支持该模型的所有配置
	configs, err := s.apiConfigRepo.FindByModel(ctx, model)
	if err != nil {
		return nil, errors.Wrap(err, 500006, "Failed to find API configs")
	}

	if len(configs) == 0 {
		return nil, errors.New(404002, fmt.Sprintf("No API configuration found for model: %s", model))
	}

	// 如果只有一个配置，直接返回
	if len(configs) == 1 {
		return configs[0], nil
	}

	// 获取负载均衡配置
	lbConfig, err := s.loadBalancerSvc.GetConfigByModel(ctx, model)
	if err != nil || lbConfig == nil {
		// 没有负载均衡配置，使用第一个
		return configs[0], nil
	}

	// 根据策略选择配置
	return s.selectByStrategy(configs, lbConfig.Strategy)
}

// selectByStrategy 根据策略选择配置
func (s *service) selectByStrategy(configs []*apiconfig.APIConfig, strategy string) (*apiconfig.APIConfig, error) {
	switch strategy {
	case "round_robin":
		// 简单轮询：返回第一个（实际应该维护计数器）
		return configs[0], nil
	case "weighted_round_robin":
		// 加权轮询：根据权重选择
		return s.selectByWeight(configs), nil
	case "random":
		// 随机选择
		return configs[utils.Min(len(configs)-1, int(time.Now().UnixNano()%int64(len(configs))))], nil
	default:
		return configs[0], nil
	}
}

// selectByWeight 根据权重选择配置
func (s *service) selectByWeight(configs []*apiconfig.APIConfig) *apiconfig.APIConfig {
	totalWeight := 0
	for _, cfg := range configs {
		totalWeight += cfg.Weight
	}

	if totalWeight == 0 {
		return configs[0]
	}

	// 生成随机数
	random := int(time.Now().UnixNano() % int64(totalWeight))
	
	// 根据权重选择
	for _, cfg := range configs {
		random -= cfg.Weight
		if random < 0 {
			return cfg
		}
	}

	return configs[0]
}

// calculateAndDeductCost 计算费用并扣除配额
func (s *service) calculateAndDeductCost(ctx context.Context, userID uint, apiConfigID uint, model string, usage adapter.UsageInfo) (int, error) {
	// 计算费用
	costReq := &pricing.CalculateCostRequest{
		APIConfigID:  apiConfigID,
		ModelName:    model,
		InputTokens:  int64(usage.PromptTokens),
		OutputTokens: int64(usage.CompletionTokens),
	}

	costResp, err := s.pricingService.CalculateCost(ctx, costReq)
	if err != nil {
		return 0, err
	}

	// 扣除配额
	if err := s.quotaService.DeductQuota(ctx, userID, int64(costResp.TotalCost)); err != nil {
		return 0, err
	}

	return int(costResp.TotalCost), nil
}

// logRequest 记录请求日志
func (s *service) logRequest(ctx context.Context, req *ProxyRequest, apiConfigID uint, tokensUsed int, responseTime time.Duration, err error) {
	logReq := &log.CreateLogRequest{
		UserID:       req.UserID,
		APIKeyID:     req.APIKeyID,
		APIConfigID:  apiConfigID,
		Model:        req.Model,
		Method:       "POST",
		Path:         "/v1/chat/completions",
		StatusCode:   200,
		ResponseTime: int(responseTime.Milliseconds()),
		TokensUsed:   tokensUsed,
	}

	if err != nil {
		logReq.StatusCode = 500
		logReq.ErrorMsg = err.Error()
	}

	if err := s.logService.CreateLog(context.Background(), logReq); err != nil {
		s.logger.Warn("Failed to create log", logger.Error(err))
	}
}
