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
	Status      string `json:"status"`
	Description string `json:"description"`
	ConfigCount int    `json:"config_count"`
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
func (s *ModelService) GetAllModels(ctx context.Context) ([]*ModelInfo, error) {
	// Get all active configurations
	configs, err := s.configRepo.FindActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active configs: %w", err)
	}

	// Extract models and deduplicate
	modelMap := make(map[string]*ModelInfo)
	modelConfigCount := make(map[string]int)

	for _, config := range configs {
		for _, modelName := range config.Models {
			// Count configurations per model
			modelConfigCount[modelName]++

			// If model already exists, just update count
			if _, exists := modelMap[modelName]; exists {
				continue
			}

			// Infer provider from model name or config type
			provider := s.inferProvider(modelName, config.Type)

			// Create model info
			modelMap[modelName] = &ModelInfo{
				Name:        modelName,
				Provider:    provider,
				Type:        "chat",
				Status:      "active",
				Description: fmt.Sprintf("%s model", provider),
			}
		}
	}

	// Convert map to slice and add config counts
	models := make([]*ModelInfo, 0, len(modelMap))
	for modelName, modelInfo := range modelMap {
		modelInfo.ConfigCount = modelConfigCount[modelName]
		models = append(models, modelInfo)
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

// GetModelInfo returns information about a specific model
func (s *ModelService) GetModelInfo(ctx context.Context, modelName string) (*ModelInfo, error) {
	allModels, err := s.GetAllModels(ctx)
	if err != nil {
		return nil, err
	}

	for _, model := range allModels {
		if model.Name == modelName {
			return model, nil
		}
	}

	return nil, fmt.Errorf("model not found: %s", modelName)
}

// inferProvider infers the provider from model name or config type
func (s *ModelService) inferProvider(modelName, configType string) string {
	modelLower := strings.ToLower(modelName)

	// Check model name patterns
	if strings.HasPrefix(modelLower, "gpt-") || strings.Contains(modelLower, "davinci") ||
		strings.Contains(modelLower, "curie") || strings.Contains(modelLower, "babbage") {
		return "OpenAI"
	}
	if strings.HasPrefix(modelLower, "claude-") || strings.Contains(modelLower, "claude") {
		return "Anthropic"
	}
	if strings.HasPrefix(modelLower, "gemini-") || strings.Contains(modelLower, "gemini") ||
		strings.Contains(modelLower, "palm") {
		return "Google"
	}
	if strings.Contains(modelLower, "llama") {
		return "Meta"
	}
	if strings.Contains(modelLower, "mistral") {
		return "Mistral"
	}

	// Fall back to config type
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
