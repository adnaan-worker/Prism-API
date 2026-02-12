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
	Status      string `json:"status"` // "active" or "inactive" for frontend
	ConfigCount int    `json:"config_count"` // Number of configs providing this model
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
// Models are deduplicated by name, with config_count showing how many configs provide each model
func (s *ModelService) GetAllModels(ctx context.Context) ([]*ModelInfo, error) {
	// Get all active configurations
	configs, err := s.configRepo.FindActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active configs: %w", err)
	}

	// Use map to deduplicate models and count configs
	modelMap := make(map[string]*ModelInfo)

	for _, config := range configs {
		provider := s.normalizeProvider(config.Type)
		
		for _, modelName := range config.Models {
			if existing, ok := modelMap[modelName]; ok {
				// Model already exists, increment config count
				existing.ConfigCount++
				// If any config is active, mark model as active
				if config.IsActive {
					existing.IsActive = true
					existing.Status = "active"
				}
			} else {
				// New model
				status := "inactive"
				if config.IsActive {
					status = "active"
				}
				
				modelMap[modelName] = &ModelInfo{
					Name:        modelName,
					Provider:    provider,
					Type:        "chat",
					ConfigID:    config.ID,
					ConfigName:  config.Name,
					BaseURL:     config.BaseURL,
					IsActive:    config.IsActive,
					Status:      status,
					ConfigCount: 1,
					Description: s.getModelDescription(modelName, provider),
				}
			}
		}
	}

	// Convert map to slice
	models := make([]*ModelInfo, 0, len(modelMap))
	for _, model := range modelMap {
		models = append(models, model)
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
// Returns deduplicated model info with config count
func (s *ModelService) GetModelInfo(ctx context.Context, modelName string) ([]*ModelInfo, error) {
	allModels, err := s.GetAllModels(ctx)
	if err != nil {
		return nil, err
	}

	for _, model := range allModels {
		if model.Name == modelName {
			return []*ModelInfo{model}, nil
		}
	}

	return nil, fmt.Errorf("model not found: %s", modelName)
}

// normalizeProvider normalizes provider names
func (s *ModelService) normalizeProvider(configType string) string {
	switch strings.ToLower(configType) {
	case "openai":
		return "openai"
	case "anthropic":
		return "anthropic"
	case "gemini":
		return "gemini"
	case "kiro":
		return "kiro"
	default:
		return "custom"
	}
}

// getModelDescription returns a description for a model
func (s *ModelService) getModelDescription(modelName, provider string) string {
	// Common model descriptions
	descriptions := map[string]string{
		"gpt-4":                    "OpenAI's most capable model, great for complex tasks",
		"gpt-4-turbo":              "Faster and more affordable GPT-4 variant",
		"gpt-3.5-turbo":            "Fast and efficient model for most tasks",
		"claude-3-opus":            "Anthropic's most powerful model for complex reasoning",
		"claude-3-sonnet":          "Balanced performance and speed",
		"claude-3-haiku":           "Fast and cost-effective Claude model",
		"claude-sonnet-4-5":        "Latest Claude Sonnet model with improved capabilities",
		"gemini-pro":               "Google's advanced AI model",
		"gemini-1.5-pro":           "Enhanced Gemini with larger context window",
	}
	
	if desc, ok := descriptions[modelName]; ok {
		return desc
	}
	
	// Default description
	return fmt.Sprintf("%s model provided by %s", modelName, provider)
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
