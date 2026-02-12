package accountpool

import (
	"api-aggregator/backend/internal/adapter"
	"api-aggregator/backend/internal/models"
	"context"
	"fmt"
	"net/http"
)

// KiroAdapterWrapper Kiro 适配器包装器
// 从凭据中提取必要信息并创建 Kiro 适配器
type KiroAdapterWrapper struct {
	credential *models.AccountCredential
	adapter    adapter.Adapter
}

// NewKiroAdapterWrapper 创建 Kiro 适配器包装器
func NewKiroAdapterWrapper(cred *models.AccountCredential, modelMapper adapter.KiroModelMapper) (*KiroAdapterWrapper, error) {
	// CredentialsData is already a map[string]interface{}
	credData := cred.CredentialsData

	// 提取必要字段
	accessToken, _ := credData["access_token"].(string)
	if accessToken == "" {
		return nil, fmt.Errorf("access_token not found in credentials")
	}

	profileArn, _ := credData["profile_arn"].(string)
	region, _ := credData["region"].(string)
	if region == "" {
		region = "us-east-1" // Default region
	}

	// 创建适配器配置
	config := &adapter.Config{
		Timeout: 120, // 2 minutes timeout for Kiro
	}

	// 创建 Kiro 适配器
	kiroAdapter := adapter.NewKiroAdapter(config, accessToken, profileArn, region, modelMapper)

	return &KiroAdapterWrapper{
		credential: cred,
		adapter:    kiroAdapter,
	}, nil
}

// Call 非流式请求
func (w *KiroAdapterWrapper) Call(ctx context.Context, req *adapter.ChatRequest) (*adapter.ChatResponse, error) {
	return w.adapter.Call(ctx, req)
}

// CallStream 流式请求
func (w *KiroAdapterWrapper) CallStream(ctx context.Context, req *adapter.ChatRequest) (*http.Response, error) {
	return w.adapter.CallStream(ctx, req)
}

// GetType 返回适配器类型
func (w *KiroAdapterWrapper) GetType() string {
	return "kiro"
}
