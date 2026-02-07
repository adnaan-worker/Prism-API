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
	quotaService   *QuotaService
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
		quotaService:   quotaService,
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

	// 2. Check user quota
	user, err := s.userRepo.FindByID(ctx, keyRecord.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Estimate tokens (rough estimate: 1 token per 4 characters)
	estimatedTokens := int64(len(req.Messages[0].Content) / 4)
	if estimatedTokens < 10 {
		estimatedTokens = 10
	}

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
	lb := s.lbFactory.CreateLoadBalancer("round_robin") // Default strategy
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

	// 6. Deduct quota
	tokensUsed := int64(resp.Usage.TotalTokens)
	if err := s.quotaService.DeductQuota(ctx, user.ID, tokensUsed); err != nil {
		// Log the error but don't fail the request since API call succeeded
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

	// Estimate tokens (rough estimate: 1 token per 4 characters)
	estimatedTokens := int64(len(req.Messages[0].Content) / 4)
	if estimatedTokens < 10 {
		estimatedTokens = 10
	}

	hasQuota, err := s.quotaService.CheckQuota(ctx, user.ID, estimatedTokens)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to check quota: %w", err)
	}
	if !hasQuota {
		return nil, nil, ErrInsufficientQuota
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
	lb := s.lbFactory.CreateLoadBalancer("round_robin")
	selectedConfig, err := lb.SelectConfig(ctx, configs)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to select config: %w", err)
	}

	// 5. Create adapter and call streaming API
	adapterInstance, err := s.adapterFactory.CreateAdapter(selectedConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create adapter: %w", err)
	}

	// Make the streaming API call
	resp, err := adapterInstance.CallStream(ctx, req)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %v", ErrAPICallFailed, err)
	}

	// Return config and response for caller to handle streaming
	return selectedConfig, resp, nil
}
