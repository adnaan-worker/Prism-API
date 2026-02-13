package adapter

import (
	"fmt"
)

// APIConfigInterface 定义 API 配置接口（避免循环依赖）
type APIConfigInterface interface {
	GetType() string
	GetBaseURL() string
	GetAPIKey() string
	GetTimeout() int
}

// Factory creates adapters based on API configuration
type Factory struct{}

// NewFactory creates a new adapter factory
func NewFactory() *Factory {
	return &Factory{}
}

// CreateAdapter creates an adapter based on the API configuration
func (f *Factory) CreateAdapter(config APIConfigInterface) (Adapter, error) {
	adapterConfig := &Config{
		BaseURL: config.GetBaseURL(),
		APIKey:  config.GetAPIKey(),
		Model:   "", // Model will be set per request
		Timeout: config.GetTimeout(),
	}

	configType := config.GetType()
	switch configType {
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
		return nil, fmt.Errorf("unsupported adapter type: %s", configType)
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
