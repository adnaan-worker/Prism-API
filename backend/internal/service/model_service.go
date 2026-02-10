package service

import (
	"api-aggregator/backend/internal/repository"
	"context"
	"fmt"
	"strings"
)

// ModelInfo represents information about a model
type ModelInfo struct {
	Name        string `json:"name"`
	Provider    string `json:"provider"`
	Type        string `json:"type"`
	ConfigID    uint   `json:"config_id"`
	ConfigName  string `json:"config_name"`
	BaseURL     string `json:"base_url"`
	IsActive    bool   `json:"is_active"`
	Description string `json:"description"`
}

// ModelGroup represents models grouped by provider
type ModelGroup struct {
	Provider string       `json:"provider"`
	Models   []*ModelInfo `json:"models"`
	Count    int          `json:"count"`
}

// ModelService handles model-related operations
type ModelService struct {
	configRepo *repository.APIConfigRepository
}

// NewModelService creates a new model service
func NewModelService(configRepo *repository.APIConfigRepository) *ModelService {
	return &ModelService{
		configRepo: configRepo,
	}
}

// GetAllModels returns all available models from active API configurations
// Each configuration's models are listed separately (no deduplication)
func (s *ModelService) GetAllModels(ctx context.Context) ([]*ModelInfo, error) {
	// Get all active configurations
	configs, err := s.configRepo.FindActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active configs: %w", err)
	}

	models := make([]*ModelInfo, 0)

	for _, config := range configs {
		provider := s.normalizeProvider(config.Type)
		
		for _, modelName := range config.Models {
			models = append(models, &ModelInfo{
				Name:        modelName,
				Provider:    provider,
				Type:        "chat",
				ConfigID:    config.ID,
				ConfigName:  config.Name,
				BaseURL:     config.BaseURL,
				IsActive:    config.IsActive,
				Description: fmt.Sprintf("%s - %s", provider, config.Name),
			})
		}
	}

	return models, nil
}

// GetModelsByProvider returns models filtered by provider
func (s *ModelService) GetModelsByProvider(ctx context.Context, provider string) ([]*ModelInfo, error) {
	allModels, err := s.GetAllModels(ctx)
	if err != nil {
		return nil, err
	}

	filtered := make([]*ModelInfo, 0)
	for _, model := range allModels {
		if strings.EqualFold(model.Provider, provider) {
			filtered = append(filtered, model)
		}
	}

	return filtered, nil
}

// GetModelsGroupedByProvider returns models grouped by provider
func (s *ModelService) GetModelsGroupedByProvider(ctx context.Context) ([]*ModelGroup, error) {
	allModels, err := s.GetAllModels(ctx)
	if err != nil {
		return nil, err
	}

	// Group by provider
	groupMap := make(map[string][]*ModelInfo)
	for _, model := range allModels {
		groupMap[model.Provider] = append(groupMap[model.Provider], model)
	}

	// Convert to slice
	groups := make([]*ModelGroup, 0, len(groupMap))
	for provider, models := range groupMap {
		groups = append(groups, &ModelGroup{
			Provider: provider,
			Models:   models,
			Count:    len(models),
		})
	}

	return groups, nil
}

// GetModelInfo returns information about a specific model
// If multiple configs provide the same model, returns all of them
func (s *ModelService) GetModelInfo(ctx context.Context, modelName string) ([]*ModelInfo, error) {
	allModels, err := s.GetAllModels(ctx)
	if err != nil {
		return nil, err
	}

	matches := make([]*ModelInfo, 0)
	for _, model := range allModels {
		if model.Name == modelName {
			matches = append(matches, model)
		}
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("model not found: %s", modelName)
	}

	return matches, nil
}

// normalizeProvider normalizes provider names
func (s *ModelService) normalizeProvider(configType string) string {
	switch strings.ToLower(configType) {
	case "openai":
		return "OpenAI"
	case "anthropic":
		return "Anthropic"
	case "gemini":
		return "Google"
	default:
		return "Custom"
	}
}

// GetUniqueModels returns deduplicated list of model names
func (s *ModelService) GetUniqueModels(ctx context.Context) ([]string, error) {
	configs, err := s.configRepo.FindActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active configs: %w", err)
	}

	// Use map for deduplication
	modelSet := make(map[string]bool)
	for _, config := range configs {
		for _, modelName := range config.Models {
			modelSet[modelName] = true
		}
	}

	// Convert to slice
	models := make([]string, 0, len(modelSet))
	for modelName := range modelSet {
		models = append(models, modelName)
	}

	return models, nil
}
