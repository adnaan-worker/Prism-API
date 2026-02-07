package adapter

import (
	"api-aggregator/backend/internal/models"
	"fmt"
)

// Factory creates adapters based on API configuration
type Factory struct{}

// NewFactory creates a new adapter factory
func NewFactory() *Factory {
	return &Factory{}
}

// CreateAdapter creates an adapter based on the API configuration
func (f *Factory) CreateAdapter(config *models.APIConfig) (Adapter, error) {
	adapterConfig := &Config{
		BaseURL: config.BaseURL,
		APIKey:  config.APIKey,
		Model:   "", // Model will be set per request
		Timeout: config.Timeout,
	}

	switch config.Type {
	case "openai":
		return NewOpenAIAdapter(adapterConfig), nil
	case "anthropic":
		return NewAnthropicAdapter(adapterConfig), nil
	case "gemini":
		return NewGeminiAdapter(adapterConfig), nil
	case "custom":
		// For custom type, default to OpenAI-compatible format
		return NewOpenAIAdapter(adapterConfig), nil
	default:
		return nil, fmt.Errorf("unsupported adapter type: %s", config.Type)
	}
}

// CreateAdapterByType creates an adapter by type string
func (f *Factory) CreateAdapterByType(adapterType, baseURL, apiKey string, timeout int) (Adapter, error) {
	config := &Config{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Timeout: timeout,
	}

	switch adapterType {
	case "openai":
		return NewOpenAIAdapter(config), nil
	case "anthropic":
		return NewAnthropicAdapter(config), nil
	case "gemini":
		return NewGeminiAdapter(config), nil
	case "custom":
		return NewOpenAIAdapter(config), nil
	default:
		return nil, fmt.Errorf("unsupported adapter type: %s", adapterType)
	}
}
