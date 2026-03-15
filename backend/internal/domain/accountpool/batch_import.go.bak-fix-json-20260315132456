package accountpool

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// KiroAccountImport Kiro 账号导入格式（从 Kiro Account Manager 导出）
type KiroAccountImport struct {
	Email    string `json:"email"`
	UserID   string `json:"userId"`
	Nickname string `json:"nickname"`
	IDP      string `json:"idp"`
	Credentials struct {
		AccessToken  string `json:"accessToken"`
		CSRFToken    string `json:"csrfToken"`
		RefreshToken string `json:"refreshToken"`
		ClientID     string `json:"clientId"`
		ClientSecret string `json:"clientSecret"`
		Region       string `json:"region"`
		ExpiresAt    int64  `json:"expiresAt"`
		AuthMethod   string `json:"authMethod"`
		Provider     string `json:"provider"`
	} `json:"credentials"`
	Subscription struct {
		Type       string `json:"type"`
		Title      string `json:"title"`
		ExpiresAt  int64  `json:"expiresAt,omitempty"`
	} `json:"subscription"`
	Usage struct {
		Current      int     `json:"current"`
		Limit        int     `json:"limit"`
		PercentUsed  float64 `json:"percentUsed"`
	} `json:"usage"`
	Status string `json:"status"`
	ID     string `json:"id"`
}

// BatchImportRequest 批量导入请求
type BatchImportRequest struct {
	PoolID   uint                 `json:"pool_id" binding:"required"`
	Accounts []KiroAccountImport  `json:"accounts" binding:"required,min=1"`
	Weight   int                  `json:"weight"`       // 默认权重
	RateLimit int                 `json:"rate_limit"`   // 默认速率限制
}

// BatchImportResponse 批量导入响应
type BatchImportResponse struct {
	Total     int      `json:"total"`
	Success   int      `json:"success"`
	Failed    int      `json:"failed"`
	Errors    []string `json:"errors,omitempty"`
	CreatedIDs []uint  `json:"created_ids,omitempty"`
}

// BatchImportFromJSON 从 JSON 字符串批量导入
func (s *service) BatchImportFromJSON(ctx context.Context, poolID uint, jsonData string, weight int, rateLimit int) (*BatchImportResponse, error) {
	// 解析 JSON
	var accounts []KiroAccountImport
	if err := json.Unmarshal([]byte(jsonData), &accounts); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return s.BatchImport(ctx, poolID, accounts, weight, rateLimit)
}

// BatchImport 批量导入账号
func (s *service) BatchImport(ctx context.Context, poolID uint, accounts []KiroAccountImport, defaultWeight int, defaultRateLimit int) (*BatchImportResponse, error) {
	// 验证账号池是否存在
	_, err := s.repo.FindByID(ctx, poolID)
	if err != nil {
		return nil, fmt.Errorf("account pool not found")
	}

	// 设置默认值
	if defaultWeight == 0 {
		defaultWeight = 1
	}

	response := &BatchImportResponse{
		Total:      len(accounts),
		Success:    0,
		Failed:     0,
		Errors:     []string{},
		CreatedIDs: []uint{},
	}

	// 逐个导入账号
	for i, acc := range accounts {
		// 计算到期天数
		var daysRemaining *int
		if acc.Subscription.ExpiresAt > 0 {
			days := int((acc.Subscription.ExpiresAt - time.Now().UnixMilli()) / (24 * 60 * 60 * 1000))
			daysRemaining = &days
		}

		// 计算使用百分比
		usagePercent := 0.0
		if acc.Usage.Limit > 0 {
			usagePercent = float64(acc.Usage.Current) / float64(acc.Usage.Limit) * 100
		}

		// 转换过期时间
		var subscriptionExpiresAt *time.Time
		if acc.Subscription.ExpiresAt > 0 {
			t := time.UnixMilli(acc.Subscription.ExpiresAt)
			subscriptionExpiresAt = &t
		}

		var expiresAt *time.Time
		if acc.Credentials.ExpiresAt > 0 {
			t := time.UnixMilli(acc.Credentials.ExpiresAt)
			expiresAt = &t
		}

		// 构建 Metadata
		metadata := JSONMap{
			"client_id":     acc.Credentials.ClientID,
			"client_secret": acc.Credentials.ClientSecret,
			"account_name":  acc.Nickname,
			"account_email": acc.Email,
		}
		
		// 添加订阅信息
		if acc.Subscription.Type != "" {
			metadata["subscription"] = map[string]interface{}{
				"type":            acc.Subscription.Type,
				"title":           acc.Subscription.Title,
				"expires_at":      subscriptionExpiresAt,
				"days_remaining":  daysRemaining,
			}
		}
		
		// 添加使用量信息
		if acc.Usage.Limit > 0 {
			metadata["usage"] = map[string]interface{}{
				"current": acc.Usage.Current,
				"limit":   acc.Usage.Limit,
				"percent": usagePercent,
			}
		}

		// 构建凭据
		cred := &AccountCredential{
			PoolID:       poolID,
			Provider:     "kiro",
			AuthType:     AuthTypeOAuth,
			AccessToken:  acc.Credentials.AccessToken,
			RefreshToken: acc.Credentials.RefreshToken,
			ExpiresAt:    expiresAt,
			Metadata:     metadata,
			Weight:       defaultWeight,
			IsActive:     acc.Status == "active",
			HealthStatus: HealthStatusUnknown,
			RateLimit:    defaultRateLimit,
		}
		
		// 设置使用量最后更新时间到 Metadata
		if cred.Metadata == nil {
			cred.Metadata = make(JSONMap)
		}
		if usage, ok := cred.Metadata["usage"].(map[string]interface{}); ok {
			usage["last_updated"] = time.Now()
		}

		// 尝试创建凭据
		if err := s.repo.CreateCredential(ctx, cred); err != nil {
			response.Failed++
			response.Errors = append(response.Errors, fmt.Sprintf("Account %d (%s): %v", i+1, acc.Email, err))
			continue
		}

		response.Success++
		response.CreatedIDs = append(response.CreatedIDs, cred.ID)
	}

	return response, nil
}
