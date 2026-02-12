package adapter

import (
	"api-aggregator/backend/internal/models"
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// PoolManagerInterface 定义池管理器接口（避免循环依赖）
type PoolManagerInterface interface {
	GetAdapter(ctx context.Context, poolID uint) (interface{}, uint, error)
	RecordSuccess(ctx context.Context, credID uint)
	RecordError(ctx context.Context, credID uint, errMsg string)
}

// Factory creates adapters based on API configuration
type Factory struct {
	poolManager PoolManagerInterface
}

// NewFactory creates a new adapter factory
func NewFactory() *Factory {
	return &Factory{}
}

// SetPoolManager sets the pool manager for account pool support
func (f *Factory) SetPoolManager(poolManager PoolManagerInterface) {
	f.poolManager = poolManager
}

// CreateAdapter creates an adapter based on the API configuration
func (f *Factory) CreateAdapter(config *models.APIConfig) (Adapter, error) {
	// Check if this is an account pool configuration
	// Format: "account_pool:kiro:123" where 123 is the pool ID
	if strings.HasPrefix(config.Type, "account_pool:") {
		return f.createAccountPoolAdapter(config)
	}

	// Regular adapter creation
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
	case "kiro":
		// Kiro adapter requires account pool
		// If Type is just "kiro", it should use account pool
		// This is a fallback - normally should use "account_pool:kiro:X" format
		return nil, fmt.Errorf("kiro adapter requires account pool configuration. Use type 'account_pool:kiro:pool_id'")
	case "custom":
		// For custom type, default to OpenAI-compatible format
		return NewOpenAIAdapter(adapterConfig), nil
	default:
		return nil, fmt.Errorf("unsupported adapter type: %s", config.Type)
	}
}

// createAccountPoolAdapter creates an account pool adapter
// Expected format: "account_pool:kiro:123" or "account_pool:gemini:456"
func (f *Factory) createAccountPoolAdapter(config *models.APIConfig) (Adapter, error) {
	if f.poolManager == nil {
		return nil, fmt.Errorf("account pool feature is not enabled")
	}

	// Parse the type string
	parts := strings.Split(config.Type, ":")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid account pool type format: %s (expected: account_pool:provider:pool_id)", config.Type)
	}

	poolIDStr := parts[2]

	// Parse pool ID
	poolID, err := strconv.ParseUint(poolIDStr, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid pool ID: %s", poolIDStr)
	}

	// Create account pool adapter wrapper
	return &accountPoolAdapterWrapper{
		poolManager: f.poolManager,
		poolID:      uint(poolID),
	}, nil
}

// accountPoolAdapterWrapper 包装池管理器以匹配 Adapter 接口
type accountPoolAdapterWrapper struct {
	poolManager PoolManagerInterface
	poolID      uint
}

func (w *accountPoolAdapterWrapper) Call(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	adapterInterface, credID, err := w.poolManager.GetAdapter(ctx, w.poolID)
	if err != nil {
		return nil, err
	}
	
	// 类型断言
	adapter, ok := adapterInterface.(Adapter)
	if !ok {
		return nil, fmt.Errorf("invalid adapter type")
	}
	
	// 调用适配器
	resp, err := adapter.Call(ctx, req)
	
	// 记录结果
	if err != nil {
		w.poolManager.RecordError(ctx, credID, err.Error())
	} else {
		w.poolManager.RecordSuccess(ctx, credID)
	}
	
	return resp, err
}

func (w *accountPoolAdapterWrapper) CallStream(ctx context.Context, req *ChatRequest) (*http.Response, error) {
	adapterInterface, credID, err := w.poolManager.GetAdapter(ctx, w.poolID)
	if err != nil {
		return nil, err
	}
	
	// 类型断言
	adapter, ok := adapterInterface.(Adapter)
	if !ok {
		return nil, fmt.Errorf("invalid adapter type")
	}
	
	// 调用适配器
	resp, err := adapter.CallStream(ctx, req)
	
	// 记录结果（流式请求只在获取响应时记录，实际成功与否在流读取时确定）
	if err != nil {
		w.poolManager.RecordError(ctx, credID, err.Error())
	} else {
		// 流式请求开始成功，记录为成功
		// 注意：流式传输中的错误不会在这里捕获
		w.poolManager.RecordSuccess(ctx, credID)
	}
	
	return resp, err
}

func (w *accountPoolAdapterWrapper) GetType() string {
	return "account_pool"
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
