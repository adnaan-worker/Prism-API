package accountpool

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// KiroRefreshService Kiro token 刷新服务
type KiroRefreshService struct {
	client *http.Client
}

// NewKiroRefreshService 创建 Kiro 刷新服务
func NewKiroRefreshService() *KiroRefreshService {
	return &KiroRefreshService{
		client: &http.Client{
			Timeout: 60 * time.Second, // 增加超时时间
		},
	}
}

// RefreshKiroToken 刷新 Kiro token
// 参考 Kiro Account Manager 的实现
func (s *KiroRefreshService) RefreshKiroToken(ctx context.Context, cred *AccountCredential) error {
	if cred.Provider != "kiro" {
		return fmt.Errorf("credential is not kiro type")
	}

	if cred.RefreshToken == "" {
		return fmt.Errorf("refresh token is empty")
	}

	// 从 SessionToken 中提取 clientId 和 clientSecret
	// SessionToken 存储格式: {"clientId":"xxx","clientSecret":"yyy"}
	var tokenData struct {
		ClientID     string `json:"clientId"`
		ClientSecret string `json:"clientSecret"`
	}

	if cred.SessionToken != "" {
		if err := json.Unmarshal([]byte(cred.SessionToken), &tokenData); err != nil {
			return fmt.Errorf("failed to parse session token: %w", err)
		}
	}

	// 如果没有 clientId/clientSecret，返回错误
	if tokenData.ClientID == "" || tokenData.ClientSecret == "" {
		return fmt.Errorf("missing clientId or clientSecret in session token")
	}

	// AWS OIDC 刷新端点
	region := "us-east-1" // Kiro 默认使用 us-east-1
	endpoint := fmt.Sprintf("https://oidc.%s.amazonaws.com/token", region)

	// 构建刷新请求（完全按照 Kiro Account Manager 的格式）
	reqBody := map[string]string{
		"clientId":     tokenData.ClientID,
		"clientSecret": tokenData.ClientSecret,
		"refreshToken": cred.RefreshToken,
		"grantType":    "refresh_token",
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("refresh failed with status %d: %s", resp.StatusCode, string(body))
	}

	// 解析响应（按照 AWS OIDC 的响应格式）
	var refreshResp struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
		ExpiresIn    int64  `json:"expiresIn"`
		TokenType    string `json:"tokenType"`
	}

	if err := json.Unmarshal(body, &refreshResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// 更新凭据
	if refreshResp.AccessToken != "" {
		cred.AccessToken = refreshResp.AccessToken
	}
	
	// 如果返回了新的 refreshToken，更新它
	if refreshResp.RefreshToken != "" {
		cred.RefreshToken = refreshResp.RefreshToken
	}

	// 更新过期时间
	if refreshResp.ExpiresIn > 0 {
		expiresAt := time.Now().Add(time.Duration(refreshResp.ExpiresIn) * time.Second)
		cred.ExpiresAt = &expiresAt
	}

	// 更新状态
	cred.Status = "active"
	cred.HealthStatus = HealthStatusHealthy
	cred.LastError = ""

	return nil
}
