package service

import (
	"api-aggregator/backend/internal/models"
	"api-aggregator/backend/internal/repository"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
)

var (
	ErrAPIKeyNotFound     = errors.New("api key not found")
	ErrAPIKeyInactive     = errors.New("api key is inactive")
	ErrUnauthorizedAccess = errors.New("unauthorized access to api key")
)

type APIKeyService struct {
	apiKeyRepo *repository.APIKeyRepository
}

func NewAPIKeyService(apiKeyRepo *repository.APIKeyRepository) *APIKeyService {
	return &APIKeyService{
		apiKeyRepo: apiKeyRepo,
	}
}

// CreateAPIKeyRequest represents a request to create an API key
type CreateAPIKeyRequest struct {
	Name      string `json:"name" binding:"required"`
	RateLimit int    `json:"rate_limit"`
}

// CreateAPIKey creates a new API key for a user
func (s *APIKeyService) CreateAPIKey(ctx context.Context, userID uint, req *CreateAPIKeyRequest) (*models.APIKey, error) {
	// Generate unique API key
	key, err := s.generateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate api key: %w", err)
	}

	// Set default rate limit if not provided
	rateLimit := req.RateLimit
	if rateLimit == 0 {
		rateLimit = 60
	}

	apiKey := &models.APIKey{
		UserID:    userID,
		Key:       key,
		Name:      req.Name,
		IsActive:  true,
		RateLimit: rateLimit,
	}

	if err := s.apiKeyRepo.Create(ctx, apiKey); err != nil {
		return nil, fmt.Errorf("failed to create api key: %w", err)
	}

	return apiKey, nil
}

// GetAPIKeysByUserID retrieves all API keys for a user
func (s *APIKeyService) GetAPIKeysByUserID(ctx context.Context, userID uint) ([]*models.APIKey, error) {
	apiKeys, err := s.apiKeyRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get api keys: %w", err)
	}
	return apiKeys, nil
}

// DeleteAPIKey deletes an API key (soft delete)
func (s *APIKeyService) DeleteAPIKey(ctx context.Context, userID uint, keyID uint) error {
	// First, verify the key belongs to the user
	apiKey, err := s.apiKeyRepo.FindByID(ctx, keyID)
	if err != nil {
		return fmt.Errorf("failed to find api key: %w", err)
	}
	if apiKey == nil {
		return ErrAPIKeyNotFound
	}
	if apiKey.UserID != userID {
		return ErrUnauthorizedAccess
	}

	if err := s.apiKeyRepo.Delete(ctx, keyID); err != nil {
		return fmt.Errorf("failed to delete api key: %w", err)
	}

	return nil
}

// ValidateAPIKey validates an API key and returns the associated user ID
func (s *APIKeyService) ValidateAPIKey(ctx context.Context, key string) (uint, error) {
	apiKey, err := s.apiKeyRepo.FindByKey(ctx, key)
	if err != nil {
		return 0, fmt.Errorf("failed to find api key: %w", err)
	}
	if apiKey == nil {
		return 0, ErrAPIKeyNotFound
	}
	if !apiKey.IsActive {
		return 0, ErrAPIKeyInactive
	}

	// Update last used timestamp
	_ = s.apiKeyRepo.UpdateLastUsedAt(ctx, apiKey.ID)

	return apiKey.UserID, nil
}

// GetAPIKeyByKey retrieves an API key by its key string
func (s *APIKeyService) GetAPIKeyByKey(ctx context.Context, key string) (*models.APIKey, error) {
	apiKey, err := s.apiKeyRepo.FindByKey(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to find api key: %w", err)
	}
	if apiKey == nil {
		return nil, ErrAPIKeyNotFound
	}

	// Update last used timestamp
	_ = s.apiKeyRepo.UpdateLastUsedAt(ctx, apiKey.ID)

	return apiKey, nil
}

// generateAPIKey generates a unique API key with "sk-" prefix
func (s *APIKeyService) generateAPIKey() (string, error) {
	// Generate 32 random bytes
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Convert to hex string and add "sk-" prefix
	return "sk-" + hex.EncodeToString(bytes), nil
}
