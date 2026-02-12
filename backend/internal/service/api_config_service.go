package service

import (
	"api-aggregator/backend/internal/models"
	"api-aggregator/backend/internal/repository"
	"context"
	"errors"
	"fmt"
)

var (
	ErrConfigNotFound = errors.New("API configuration not found")
	ErrInvalidConfig  = errors.New("invalid API configuration")
)

type APIConfigService struct {
	configRepo *repository.APIConfigRepository
}

func NewAPIConfigService(configRepo *repository.APIConfigRepository) *APIConfigService {
	return &APIConfigService{
		configRepo: configRepo,
	}
}

// CreateConfigRequest represents a request to create an API configuration
type CreateConfigRequest struct {
	Name     string                 `json:"name" binding:"required"`
	Type     string                 `json:"type" binding:"required"`
	BaseURL  string                 `json:"base_url" binding:"required"`
	APIKey   string                 `json:"api_key"`
	Models   []string               `json:"models" binding:"required"`
	Headers  map[string]interface{} `json:"headers"`
	Priority int                    `json:"priority"`
	Weight   int                    `json:"weight"`
	MaxRPS   int                    `json:"max_rps"`
	Timeout  int                    `json:"timeout"`
}

// UpdateConfigRequest represents a request to update an API configuration
type UpdateConfigRequest struct {
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	BaseURL  string                 `json:"base_url"`
	APIKey   string                 `json:"api_key"`
	Models   []string               `json:"models"`
	Headers  map[string]interface{} `json:"headers"`
	Priority int                    `json:"priority"`
	Weight   int                    `json:"weight"`
	MaxRPS   int                    `json:"max_rps"`
	Timeout  int                    `json:"timeout"`
}

// CreateConfig creates a new API configuration
func (s *APIConfigService) CreateConfig(ctx context.Context, req *CreateConfigRequest) (*models.APIConfig, error) {
	// Validate type
	validTypes := map[string]bool{
		"openai":    true,
		"anthropic": true,
		"gemini":    true,
		"kiro":      true,
		"custom":    true,
	}
	if !validTypes[req.Type] {
		return nil, ErrInvalidConfig
	}

	// Validate models array
	if len(req.Models) == 0 {
		return nil, fmt.Errorf("models array cannot be empty")
	}

	// Set defaults
	if req.Priority == 0 {
		req.Priority = 100
	}
	if req.Weight == 0 {
		req.Weight = 1
	}
	if req.Timeout == 0 {
		req.Timeout = 30
	}

	config := &models.APIConfig{
		Name:     req.Name,
		Type:     req.Type,
		BaseURL:  req.BaseURL,
		APIKey:   req.APIKey,
		Models:   req.Models,
		Headers:  req.Headers,
		IsActive: true,
		Priority: req.Priority,
		Weight:   req.Weight,
		MaxRPS:   req.MaxRPS,
		Timeout:  req.Timeout,
	}

	if err := s.configRepo.Create(ctx, config); err != nil {
		return nil, fmt.Errorf("failed to create config: %w", err)
	}

	return config, nil
}

// GetConfig gets an API configuration by ID
func (s *APIConfigService) GetConfig(ctx context.Context, id uint) (*models.APIConfig, error) {
	config, err := s.configRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}
	if config == nil {
		return nil, ErrConfigNotFound
	}
	return config, nil
}

// GetAllConfigs gets all API configurations
func (s *APIConfigService) GetAllConfigs(ctx context.Context) ([]*models.APIConfig, error) {
	configs, err := s.configRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get configs: %w", err)
	}
	return configs, nil
}

// GetActiveConfigs gets all active API configurations
func (s *APIConfigService) GetActiveConfigs(ctx context.Context) ([]*models.APIConfig, error) {
	configs, err := s.configRepo.FindActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active configs: %w", err)
	}
	return configs, nil
}

// GetConfigsByModel gets all API configurations that support a specific model
func (s *APIConfigService) GetConfigsByModel(ctx context.Context, model string) ([]*models.APIConfig, error) {
	configs, err := s.configRepo.FindByModel(ctx, model)
	if err != nil {
		return nil, fmt.Errorf("failed to get configs by model: %w", err)
	}
	return configs, nil
}

// UpdateConfig updates an API configuration
func (s *APIConfigService) UpdateConfig(ctx context.Context, id uint, req *UpdateConfigRequest) (*models.APIConfig, error) {
	config, err := s.configRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}
	if config == nil {
		return nil, ErrConfigNotFound
	}

	// Update fields if provided
	if req.Name != "" {
		config.Name = req.Name
	}
	if req.Type != "" {
		validTypes := map[string]bool{
			"openai":    true,
			"anthropic": true,
			"gemini":    true,
			"kiro":      true,
			"custom":    true,
		}
		if !validTypes[req.Type] {
			return nil, ErrInvalidConfig
		}
		config.Type = req.Type
	}
	if req.BaseURL != "" {
		config.BaseURL = req.BaseURL
	}
	if req.APIKey != "" {
		config.APIKey = req.APIKey
	}
	if len(req.Models) > 0 {
		config.Models = req.Models
	}
	if req.Headers != nil {
		config.Headers = req.Headers
	}
	if req.Priority != 0 {
		config.Priority = req.Priority
	}
	if req.Weight != 0 {
		config.Weight = req.Weight
	}
	if req.MaxRPS != 0 {
		config.MaxRPS = req.MaxRPS
	}
	if req.Timeout != 0 {
		config.Timeout = req.Timeout
	}

	if err := s.configRepo.Update(ctx, config); err != nil {
		return nil, fmt.Errorf("failed to update config: %w", err)
	}

	return config, nil
}

// DeleteConfig deletes an API configuration
func (s *APIConfigService) DeleteConfig(ctx context.Context, id uint) error {
	config, err := s.configRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}
	if config == nil {
		return ErrConfigNotFound
	}

	if err := s.configRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete config: %w", err)
	}

	return nil
}

// ActivateConfig activates an API configuration
func (s *APIConfigService) ActivateConfig(ctx context.Context, id uint) error {
	config, err := s.configRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}
	if config == nil {
		return ErrConfigNotFound
	}

	if err := s.configRepo.UpdateStatus(ctx, id, true); err != nil {
		return fmt.Errorf("failed to activate config: %w", err)
	}

	return nil
}

// DeactivateConfig deactivates an API configuration
func (s *APIConfigService) DeactivateConfig(ctx context.Context, id uint) error {
	config, err := s.configRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}
	if config == nil {
		return ErrConfigNotFound
	}

	if err := s.configRepo.UpdateStatus(ctx, id, false); err != nil {
		return fmt.Errorf("failed to deactivate config: %w", err)
	}

	return nil
}

// BatchDeleteConfigs deletes multiple API configurations
func (s *APIConfigService) BatchDeleteConfigs(ctx context.Context, ids []uint) error {
	if len(ids) == 0 {
		return fmt.Errorf("no IDs provided")
	}

	if err := s.configRepo.BatchDelete(ctx, ids); err != nil {
		return fmt.Errorf("failed to batch delete configs: %w", err)
	}

	return nil
}

// BatchActivateConfigs activates multiple API configurations
func (s *APIConfigService) BatchActivateConfigs(ctx context.Context, ids []uint) error {
	if len(ids) == 0 {
		return fmt.Errorf("no IDs provided")
	}

	if err := s.configRepo.BatchUpdateStatus(ctx, ids, true); err != nil {
		return fmt.Errorf("failed to batch activate configs: %w", err)
	}

	return nil
}

// BatchDeactivateConfigs deactivates multiple API configurations
func (s *APIConfigService) BatchDeactivateConfigs(ctx context.Context, ids []uint) error {
	if len(ids) == 0 {
		return fmt.Errorf("no IDs provided")
	}

	if err := s.configRepo.BatchUpdateStatus(ctx, ids, false); err != nil {
		return fmt.Errorf("failed to batch deactivate configs: %w", err)
	}

	return nil
}
