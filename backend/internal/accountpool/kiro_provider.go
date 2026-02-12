package accountpool

import (
	"api-aggregator/backend/internal/models"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// AdapterConfig 适配器配置（避免导入 adapter 包）
type AdapterConfig struct {
	BaseURL string
	APIKey  string
	Timeout int
}

// KiroProvider Kiro 提供商实现
type KiroProvider struct {
	client  *http.Client
	manager *Manager // 引用 Manager 以获取 modelMapper
}

// NewKiroProvider 创建 Kiro 提供商
func NewKiroProvider(manager *Manager) *KiroProvider {
	return &KiroProvider{
		client:  &http.Client{Timeout: 30 * time.Second},
		manager: manager,
	}
}

// Name 返回提供商名称
func (p *KiroProvider) Name() string {
	return "kiro"
}

// RefreshToken 刷新访问令牌
func (p *KiroProvider) RefreshToken(ctx context.Context, cred *models.AccountCredential) error {
	refreshToken := cred.GetString("refresh_token")
	if refreshToken == "" {
		return errors.New("refresh_token not found")
	}

	authMethod := cred.GetString("auth_method")
	region := cred.GetString("region")
	if region == "" {
		region = "us-east-1"
	}

	// 自动检测认证方式
	// 如果有 client_id 和 client_secret，说明是 Builder ID 方式
	clientID := cred.GetString("client_id")
	clientSecret := cred.GetString("client_secret")
	
	if clientID != "" && clientSecret != "" {
		// 强制使用 Builder ID 方式
		authMethod = "builder-id"
	}

	// 如果没有指定 auth_method，默认使用 social（最常见的方式）
	if authMethod == "" {
		authMethod = "social"
	}

	switch authMethod {
	case "social", "refresh_token": // refresh_token 类型也使用 social 刷新接口
		return p.refreshSocial(ctx, cred, refreshToken, region)
	case "builder-id", "builder_id":
		return p.refreshBuilderID(ctx, cred, refreshToken, region)
	default:
		// 未知类型，尝试 social 方式
		return p.refreshSocial(ctx, cred, refreshToken, region)
	}
}

// refreshSocial 刷新社交登录令牌
func (p *KiroProvider) refreshSocial(ctx context.Context, cred *models.AccountCredential, refreshToken, region string) error {
	url := fmt.Sprintf("https://prod.%s.auth.desktop.kiro.dev/refreshToken", region)
	
	reqBody, _ := json.Marshal(map[string]string{"refreshToken": refreshToken})
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("refresh failed (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
		ProfileArn   string `json:"profileArn"`
		ExpiresIn    int    `json:"expiresIn"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// 更新凭据
	cred.SetString("access_token", result.AccessToken)
	cred.SetString("refresh_token", result.RefreshToken)
	if result.ProfileArn != "" {
		cred.SetString("profile_arn", result.ProfileArn)
	}
	cred.SetString("auth_method", "social")
	expiresAt := time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)
	cred.ExpiresAt = &expiresAt
	
	// 重置每日配额（如果已过重置时间）
	p.resetDailyQuotaIfNeeded(cred)

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// refreshBuilderID 刷新 Builder ID 令牌
func (p *KiroProvider) refreshBuilderID(ctx context.Context, cred *models.AccountCredential, refreshToken, region string) error {
	idcRegion := cred.GetString("idc_region")
	if idcRegion == "" {
		idcRegion = region
	}

	url := fmt.Sprintf("https://oidc.%s.amazonaws.com/token", idcRegion)
	
	clientID := cred.GetString("client_id")
	clientSecret := cred.GetString("client_secret")
	
	if clientID == "" || clientSecret == "" {
		return errors.New("client_id or client_secret not found for Builder ID auth")
	}
	
	reqBody, _ := json.Marshal(map[string]string{
		"clientId":     clientID,
		"clientSecret": clientSecret,
		"refreshToken": refreshToken,
		"grantType":    "refresh_token",
	})

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("refresh failed (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
		ExpiresIn    int    `json:"expiresIn"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// 更新凭据
	cred.SetString("access_token", result.AccessToken)
	cred.SetString("refresh_token", result.RefreshToken)
	cred.SetString("auth_method", "builder-id")
	expiresAt := time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)
	cred.ExpiresAt = &expiresAt
	
	// 重置每日配额（如果已过重置时间）
	p.resetDailyQuotaIfNeeded(cred)

	return nil
}

// CheckHealth 检查凭据健康状态
func (p *KiroProvider) CheckHealth(ctx context.Context, cred *models.AccountCredential) error {
	if cred.IsExpired() {
		return errors.New("token expired")
	}
	if cred.IsExpiringSoon(5 * time.Minute) {
		return errors.New("token expiring soon")
	}
	return nil
}

// CreateAdapter 创建 Kiro 适配器
func (p *KiroProvider) CreateAdapter(cred *models.AccountCredential) (interface{}, error) {
	modelMapper := p.manager.GetModelMapper()
	if modelMapper == nil {
		return nil, fmt.Errorf("model mapper not initialized")
	}
	return NewKiroAdapterWrapper(cred, modelMapper)
}

// GetAuthURL 获取 OAuth 授权 URL
func (p *KiroProvider) GetAuthURL(ctx context.Context, state string) (string, error) {
	region := "us-east-1"
	return fmt.Sprintf("https://prod.%s.auth.desktop.kiro.dev/authorize?state=%s", region, state), nil
}

// ExchangeCode 交换授权码
func (p *KiroProvider) ExchangeCode(ctx context.Context, code string) (*models.AccountCredential, error) {
	return nil, errors.New("not implemented")
}

// InitiateDeviceCode 启动设备码流程
func (p *KiroProvider) InitiateDeviceCode(ctx context.Context) (map[string]interface{}, error) {
	region := "us-east-1" // 默认区域
	ssoOIDCEndpoint := fmt.Sprintf("https://oidc.%s.amazonaws.com", region)
	
	// 1. 注册 OIDC 客户端
	regBody, _ := json.Marshal(map[string]interface{}{
		"clientName": "Kiro IDE",
		"clientType": "public",
		"scopes": []string{
			"codewhisperer:completions",
			"codewhisperer:analysis",
			"codewhisperer:conversations",
		},
	})
	
	regReq, err := http.NewRequestWithContext(ctx, "POST", ssoOIDCEndpoint+"/client/register", bytes.NewBuffer(regBody))
	if err != nil {
		return nil, err
	}
	regReq.Header.Set("Content-Type", "application/json")
	regReq.Header.Set("User-Agent", "KiroIDE")
	
	regResp, err := p.client.Do(regReq)
	if err != nil {
		return nil, err
	}
	defer regResp.Body.Close()
	
	if regResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(regResp.Body)
		return nil, fmt.Errorf("client registration failed: %s", string(body))
	}
	
	var regData struct {
		ClientID     string `json:"clientId"`
		ClientSecret string `json:"clientSecret"`
	}
	if err := json.NewDecoder(regResp.Body).Decode(&regData); err != nil {
		return nil, err
	}
	
	// 2. 启动设备授权
	authBody, _ := json.Marshal(map[string]string{
		"clientId":     regData.ClientID,
		"clientSecret": regData.ClientSecret,
		"startUrl":     "https://view.awsapps.com/start",
	})
	
	authReq, err := http.NewRequestWithContext(ctx, "POST", ssoOIDCEndpoint+"/device_authorization", bytes.NewBuffer(authBody))
	if err != nil {
		return nil, err
	}
	authReq.Header.Set("Content-Type", "application/json")
	
	authResp, err := p.client.Do(authReq)
	if err != nil {
		return nil, err
	}
	defer authResp.Body.Close()
	
	if authResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(authResp.Body)
		return nil, fmt.Errorf("device authorization failed: %s", string(body))
	}
	
	var deviceAuth struct {
		DeviceCode              string `json:"deviceCode"`
		UserCode                string `json:"userCode"`
		VerificationURI         string `json:"verificationUri"`
		VerificationURIComplete string `json:"verificationUriComplete"`
		ExpiresIn               int    `json:"expiresIn"`
		Interval                int    `json:"interval"`
	}
	if err := json.NewDecoder(authResp.Body).Decode(&deviceAuth); err != nil {
		return nil, err
	}
	
	return map[string]interface{}{
		"device_code":                deviceAuth.DeviceCode,
		"user_code":                  deviceAuth.UserCode,
		"verification_uri":           deviceAuth.VerificationURI,
		"verification_uri_complete":  deviceAuth.VerificationURIComplete,
		"expires_in":                 deviceAuth.ExpiresIn,
		"interval":                   deviceAuth.Interval,
		"client_id":                  regData.ClientID,
		"client_secret":              regData.ClientSecret,
		"region":                     region,
	}, nil
}

// PollDeviceCode 轮询设备码状态
func (p *KiroProvider) PollDeviceCode(ctx context.Context, deviceCode string) (*models.AccountCredential, error) {
	// 这个方法需要从 OAuth handler 传入完整的参数
	// 实际轮询逻辑在 OAuth handler 中实现
	return nil, errors.New("use InitiateOAuth endpoint for device code flow")
}

// resetDailyQuotaIfNeeded 检查并重置每日配额
func (p *KiroProvider) resetDailyQuotaIfNeeded(cred *models.AccountCredential) {
	now := time.Now()
	
	// 如果没有设置重置时间，或者已经过了重置时间，则重置配额
	if cred.QuotaResetAt == nil || now.After(*cred.QuotaResetAt) {
		// 重置每日使用量
		cred.DailyUsed = 0
		
		// 设置下一个重置时间（明天的 00:00:00 UTC）
		tomorrow := now.AddDate(0, 0, 1)
		nextReset := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, time.UTC)
		cred.QuotaResetAt = &nextReset
	}
}

// 注册 Kiro 提供商
// 注意：不再使用 init() 自动注册，而是在 Manager 中手动注册
// func init() {
// 	Register(NewKiroProvider(nil))
// }
