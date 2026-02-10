package service

import (
	"api-aggregator/backend/internal/adapter"
	"api-aggregator/backend/internal/loadbalancer"
	"api-aggregator/backend/internal/models"
	"api-aggregator/backend/internal/repository"
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

var (
	ErrInvalidAPIKey     = errors.New("invalid API key")
	ErrInsufficientQuota = errors.New("insufficient quota")
	ErrNoConfigAvailable = errors.New("no configuration available for model")
	ErrAPICallFailed     = errors.New("API call failed")
)

// ProxyService handles API proxy requests
type ProxyService struct {
	apiKeyRepo     *repository.APIKeyRepository
	configRepo     *repository.APIConfigRepository
	userRepo       *repository.UserRepository
	requestLogRepo *repository.RequestLogRepository
	lbRepo         *repository.LoadBalancerRepository
	billingTxRepo  *repository.BillingTransactionRepository
	quotaService   *QuotaService
	billingService *BillingService
	lbFactory      *loadbalancer.Factory
	adapterFactory *adapter.Factory
}

// NewProxyService creates a new proxy service
func NewProxyService(
	apiKeyRepo *repository.APIKeyRepository,
	configRepo *repository.APIConfigRepository,
	userRepo *repository.UserRepository,
	requestLogRepo *repository.RequestLogRepository,
	quotaService *QuotaService,
) *ProxyService {
	return &ProxyService{
		apiKeyRepo:     apiKeyRepo,
		configRepo:     configRepo,
		userRepo:       userRepo,
		requestLogRepo: requestLogRepo,
		lbRepo:         nil, // Will be set later if needed
		quotaService:   quotaService,
		lbFactory:      loadbalancer.NewFactory(),
		adapterFactory: adapter.NewFactory(),
	}
}

// NewProxyServiceWithLB creates a new proxy service with load balancer repository
func NewProxyServiceWithLB(
	apiKeyRepo *repository.APIKeyRepository,
	configRepo *repository.APIConfigRepository,
	userRepo *repository.UserRepository,
	requestLogRepo *repository.RequestLogRepository,
	lbRepo *repository.LoadBalancerRepository,
	quotaService *QuotaService,
	billingService *BillingService,
) *ProxyService {
	return &ProxyService{
		apiKeyRepo:     apiKeyRepo,
		configRepo:     configRepo,
		userRepo:       userRepo,
		requestLogRepo: requestLogRepo,
		lbRepo:         lbRepo,
		quotaService:   quotaService,
		billingService: billingService,
		lbFactory:      loadbalancer.NewFactory(),
		adapterFactory: adapter.NewFactory(),
	}
}

// ProxyRequest handles a proxy request
func (s *ProxyService) ProxyRequest(ctx context.Context, apiKey string, req *adapter.ChatRequest) (*adapter.ChatResponse, error) {
	startTime := time.Now()

	// 1. Validate API Key
	keyRecord, err := s.apiKeyRepo.FindByKey(ctx, apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to validate API key: %w", err)
	}
	if keyRecord == nil || !keyRecord.IsActive {
		return nil, ErrInvalidAPIKey
	}

	// Update last used time
	now := time.Now()
	keyRecord.LastUsedAt = &now
	s.apiKeyRepo.Update(ctx, keyRecord)

	// 2. Check user quota (pre-check with estimated tokens)
	user, err := s.userRepo.FindByID(ctx, keyRecord.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Estimate tokens for pre-check (more accurate estimation)
	estimatedTokens := s.estimateTokens(req)
	
	hasQuota, err := s.quotaService.CheckQuota(ctx, user.ID, estimatedTokens)
	if err != nil {
		return nil, fmt.Errorf("failed to check quota: %w", err)
	}
	if !hasQuota {
		s.logRequest(ctx, keyRecord.UserID, keyRecord.ID, 0, req.Model, "POST", "/v1/chat/completions", 429, 0, 0, "Insufficient quota")
		return nil, ErrInsufficientQuota
	}

	// 3. Find configurations for the model
	configs, err := s.configRepo.FindByModel(ctx, req.Model)
	if err != nil {
		return nil, fmt.Errorf("failed to find configs: %w", err)
	}
	if len(configs) == 0 {
		s.logRequest(ctx, keyRecord.UserID, keyRecord.ID, 0, req.Model, "POST", "/v1/chat/completions", 404, 0, 0, "No configuration available")
		return nil, ErrNoConfigAvailable
	}

	// 4. Select configuration using load balancer
	// Get load balancer strategy for this model
	strategy := "round_robin" // Default strategy
	if s.lbRepo != nil {
		lbConfig, err := s.lbRepo.FindByModel(ctx, req.Model)
		if err == nil && lbConfig != nil && lbConfig.IsActive {
			strategy = lbConfig.Strategy
		}
	}
	
	lb := s.lbFactory.CreateLoadBalancer(strategy)
	selectedConfig, err := lb.SelectConfig(ctx, configs)
	if err != nil {
		return nil, fmt.Errorf("failed to select config: %w", err)
	}

	// 5. Create adapter and call API
	adapterInstance, err := s.adapterFactory.CreateAdapter(selectedConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create adapter: %w", err)
	}

	// Make the API call
	resp, err := adapterInstance.Call(ctx, req)
	responseTime := int(time.Since(startTime).Milliseconds())

	if err != nil {
		// Log failed request
		s.logRequest(ctx, keyRecord.UserID, keyRecord.ID, selectedConfig.ID, req.Model, "POST", "/v1/chat/completions", 500, responseTime, 0, err.Error())
		return nil, fmt.Errorf("%w: %v", ErrAPICallFailed, err)
	}

	// 6. Calculate actual cost and deduct quota
	tokensUsed := int64(resp.Usage.TotalTokens)
	
	// Calculate cost using billing service
	var costToDeduct int64
	if s.billingService != nil {
		costCalc, err := s.billingService.CalculateCost(
			ctx,
			req.Model,
			selectedConfig.ID,
			resp.Usage.PromptTokens,
			resp.Usage.CompletionTokens,
			false, // Not an estimate
		)
		if err == nil {
			// Convert micro-credits to credits (round up)
			costToDeduct = (costCalc.MicroCredits + 999) / 1000
		}
	}
	
	// Fallback: use token count if billing service not available
	if costToDeduct == 0 {
		costToDeduct = tokensUsed
	}
	
	// Deduct quota
	if err := s.quotaService.DeductQuota(ctx, user.ID, costToDeduct); err != nil {
		fmt.Printf("Warning: failed to deduct quota: %v\n", err)
	}

	// 7. Log successful request
	s.logRequest(ctx, keyRecord.UserID, keyRecord.ID, selectedConfig.ID, req.Model, "POST", "/v1/chat/completions", 200, responseTime, resp.Usage.TotalTokens, "")

	return resp, nil
}

// logRequest logs a request asynchronously
func (s *ProxyService) logRequest(ctx context.Context, userID, apiKeyID, configID uint, model, method, path string, statusCode, responseTime, tokensUsed int, errorMsg string) {
	// Log asynchronously to avoid blocking
	go func() {
		log := &models.RequestLog{
			UserID:       userID,
			APIKeyID:     apiKeyID,
			APIConfigID:  configID,
			Model:        model,
			Method:       method,
			Path:         path,
			StatusCode:   statusCode,
			ResponseTime: responseTime,
			TokensUsed:   tokensUsed,
			ErrorMsg:     errorMsg,
		}
		s.requestLogRepo.Create(context.Background(), log)
	}()
}

// ValidateAPIKey validates an API key and returns the user ID
func (s *ProxyService) ValidateAPIKey(ctx context.Context, apiKey string) (uint, error) {
	keyRecord, err := s.apiKeyRepo.FindByKey(ctx, apiKey)
	if err != nil {
		return 0, fmt.Errorf("failed to validate API key: %w", err)
	}
	if keyRecord == nil || !keyRecord.IsActive {
		return 0, ErrInvalidAPIKey
	}
	return keyRecord.UserID, nil
}

// ProxyStreamRequest handles a streaming proxy request
// Returns the HTTP response for direct streaming to client
// Note: For streaming requests, we deduct estimated quota upfront since we cannot
// accurately count tokens in real-time streaming
func (s *ProxyService) ProxyStreamRequest(ctx context.Context, apiKey string, req *adapter.ChatRequest) (*models.APIConfig, *http.Response, error) {
	// 1. Validate API Key
	keyRecord, err := s.apiKeyRepo.FindByKey(ctx, apiKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to validate API key: %w", err)
	}
	if keyRecord == nil || !keyRecord.IsActive {
		return nil, nil, ErrInvalidAPIKey
	}

	// Update last used time
	now := time.Now()
	keyRecord.LastUsedAt = &now
	s.apiKeyRepo.Update(ctx, keyRecord)

	// 2. Check user quota
	user, err := s.userRepo.FindByID(ctx, keyRecord.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, nil, fmt.Errorf("user not found")
	}

	// 3. Find configurations for the model
	configs, err := s.configRepo.FindByModel(ctx, req.Model)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to find configs: %w", err)
	}
	if len(configs) == 0 {
		return nil, nil, ErrNoConfigAvailable
	}

	// 4. Select configuration using load balancer
	// Get load balancer strategy for this model
	strategy := "round_robin" // Default strategy
	if s.lbRepo != nil {
		lbConfig, err := s.lbRepo.FindByModel(ctx, req.Model)
		if err == nil && lbConfig != nil && lbConfig.IsActive {
			strategy = lbConfig.Strategy
		}
	}
	
	lb := s.lbFactory.CreateLoadBalancer(strategy)
	selectedConfig, err := lb.SelectConfig(ctx, configs)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to select config: %w", err)
	}

	// 5. Estimate tokens and calculate cost
	estimatedInputTokens := s.estimateInputTokens(req)
	estimatedOutputTokens := s.estimateOutputTokens(req)
	estimatedTotalTokens := estimatedInputTokens + estimatedOutputTokens
	
	// Calculate estimated cost using billing service
	var costToDeduct int64
	if s.billingService != nil {
		costCalc, err := s.billingService.CalculateCost(
			ctx,
			req.Model,
			selectedConfig.ID,
			int(estimatedInputTokens),
			int(estimatedOutputTokens),
			true, // Is an estimate
		)
		if err == nil {
			// Convert micro-credits to credits (round up)
			costToDeduct = (costCalc.MicroCredits + 999) / 1000
		}
	}
	
	// Fallback: use token count if billing service not available
	if costToDeduct == 0 {
		costToDeduct = estimatedTotalTokens
	}
	
	// Check if user has sufficient quota
	hasQuota, err := s.quotaService.CheckQuota(ctx, user.ID, costToDeduct)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to check quota: %w", err)
	}
	if !hasQuota {
		return nil, nil, ErrInsufficientQuota
	}

	// 6. Create adapter and call streaming API
	adapterInstance, err := s.adapterFactory.CreateAdapter(selectedConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create adapter: %w", err)
	}

	// Make the streaming API call
	resp, err := adapterInstance.CallStream(ctx, req)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %v", ErrAPICallFailed, err)
	}

	// 7. Deduct estimated quota upfront for streaming requests
	// Note: This is an estimation since we cannot accurately count tokens in streaming
	if err := s.quotaService.DeductQuota(ctx, user.ID, costToDeduct); err != nil {
		// Close the response if quota deduction fails
		resp.Body.Close()
		return nil, nil, fmt.Errorf("failed to deduct quota: %w", err)
	}
	
	// 8. Log the streaming request (with estimated tokens)
	go func() {
		log := &models.RequestLog{
			UserID:       keyRecord.UserID,
			APIKeyID:     keyRecord.ID,
			APIConfigID:  selectedConfig.ID,
			Model:        req.Model,
			Method:       "POST",
			Path:         "/v1/chat/completions",
			StatusCode:   200,
			ResponseTime: 0, // Unknown for streaming
			TokensUsed:   int(estimatedTotalTokens),
			ErrorMsg:     "",
		}
		s.requestLogRepo.Create(context.Background(), log)
	}()

	// Return config and response for caller to handle streaming
	return selectedConfig, resp, nil
}

// estimateTokens estimates the number of tokens for a request
// This is a rough estimation for pre-check, actual usage will be deducted after API call
func (s *ProxyService) estimateTokens(req *adapter.ChatRequest) int64 {
	var totalChars int
	
	// Count all message content
	for _, msg := range req.Messages {
		totalChars += len(msg.Content)
		// Add overhead for role and structure
		totalChars += 10
	}
	
	// Add overhead for tools if present
	if len(req.Tools) > 0 {
		for _, tool := range req.Tools {
			totalChars += len(tool.Function.Name)
			totalChars += len(tool.Function.Description)
			// Rough estimate for parameters JSON
			totalChars += 100
		}
	}
	
	// Estimate tokens (1 token ≈ 4 characters for English, 2 for Chinese)
	// Use conservative estimate: 1 token per 3 characters
	estimatedInputTokens := int64(totalChars / 3)
	
	// Estimate output tokens based on max_tokens or default
	estimatedOutputTokens := int64(req.MaxTokens)
	if estimatedOutputTokens == 0 {
		// Default estimate: assume output is 50% of input, min 100, max 1000
		estimatedOutputTokens = estimatedInputTokens / 2
		if estimatedOutputTokens < 100 {
			estimatedOutputTokens = 100
		}
		if estimatedOutputTokens > 1000 {
			estimatedOutputTokens = 1000
		}
	}
	
	// Total estimated tokens
	totalEstimated := estimatedInputTokens + estimatedOutputTokens
	
	// Minimum estimate
	if totalEstimated < 50 {
		totalEstimated = 50
	}
	
	return totalEstimated
}

// estimateInputTokens estimates input tokens from request
func (s *ProxyService) estimateInputTokens(req *adapter.ChatRequest) int64 {
	var totalChars int
	
	for _, msg := range req.Messages {
		totalChars += len(msg.Content)
		totalChars += 10
	}
	
	if len(req.Tools) > 0 {
		totalChars += len(req.Tools) * 150
	}
	
	estimated := int64(totalChars / 3)
	if estimated < 10 {
		estimated = 10
	}
	
	return estimated
}

// estimateOutputTokens estimates output tokens from request
func (s *ProxyService) estimateOutputTokens(req *adapter.ChatRequest) int64 {
	if req.MaxTokens > 0 {
		return int64(req.MaxTokens)
	}
	
	inputTokens := s.estimateInputTokens(req)
	estimated := inputTokens / 2
	if estimated < 100 {
		estimated = 100
	}
	if estimated > 2000 {
		estimated = 2000
	}
	
	return estimated
}

// normalizeProvider normalizes provider name from config type
func (s *ProxyService) normalizeProvider(configType string) string {
	switch configType {
	case "openai":
		return "openai"
	case "anthropic":
		return "anthropic"
	case "gemini":
		return "gemini"
	default:
		return "custom"
	}
}
