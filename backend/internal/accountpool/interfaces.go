package accountpool

import (
	"api-aggregator/backend/internal/models"
	"context"
)

// Provider 定义账号池提供商的核心接口
// 所有提供商（Kiro、Gemini、Claude等）都必须实现此接口
type Provider interface {
	// Name 返回提供商名称（如 "kiro", "gemini"）
	Name() string

	// RefreshToken 刷新凭据的访问令牌
	RefreshToken(ctx context.Context, cred *models.AccountCredential) error

	// CheckHealth 检查凭据健康状态
	CheckHealth(ctx context.Context, cred *models.AccountCredential) error

	// CreateAdapter 为凭据创建 API 适配器实例
	// 返回 interface{} 避免循环依赖，实际返回 adapter.Adapter
	CreateAdapter(cred *models.AccountCredential) (interface{}, error)
}

// OAuthProvider 扩展 Provider 接口，支持 OAuth 认证
// 支持 OAuth 的提供商需要实现此接口
type OAuthProvider interface {
	Provider

	// GetAuthURL 获取 OAuth 授权 URL
	GetAuthURL(ctx context.Context, state string) (string, error)

	// ExchangeCode 用授权码交换访问令牌
	ExchangeCode(ctx context.Context, code string) (*models.AccountCredential, error)
}

// DeviceCodeProvider 扩展 Provider 接口，支持设备码认证
// 支持设备码流程的提供商需要实现此接口（如 AWS Builder ID）
type DeviceCodeProvider interface {
	Provider

	// InitiateDeviceCode 启动设备码流程
	InitiateDeviceCode(ctx context.Context) (map[string]interface{}, error)

	// PollDeviceCode 轮询设备码授权状态
	PollDeviceCode(ctx context.Context, deviceCode string) (*models.AccountCredential, error)
}
