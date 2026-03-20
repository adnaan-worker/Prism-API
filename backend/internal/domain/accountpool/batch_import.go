package accountpool

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// FlexibleImportTime 兼容 RFC3339 字符串、Unix 秒/毫秒时间戳的导入时间字段
type FlexibleImportTime struct {
	time.Time
	Valid bool
}

func (t *FlexibleImportTime) UnmarshalJSON(data []byte) error {
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" || trimmed == "null" {
		t.Time = time.Time{}
		t.Valid = false
		return nil
	}

	var raw interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	parsed, ok := parseFlexibleImportTime(raw)
	if !ok {
		return fmt.Errorf("invalid time value: %s", trimmed)
	}

	t.Time = parsed
	t.Valid = true
	return nil
}

func (t FlexibleImportTime) Ptr() *time.Time {
	if !t.Valid {
		return nil
	}
	tm := t.Time
	return &tm
}

func parseFlexibleImportTime(value interface{}) (time.Time, bool) {
	switch v := value.(type) {
	case string:
		v = strings.TrimSpace(v)
		if v == "" {
			return time.Time{}, false
		}
		if ts, err := time.Parse(time.RFC3339, v); err == nil {
			return ts, true
		}
		if num, err := strconv.ParseInt(v, 10, 64); err == nil {
			return unixAuto(num), true
		}
	case float64:
		if v <= 0 {
			return time.Time{}, false
		}
		return unixAuto(int64(v)), true
	case int64:
		if v <= 0 {
			return time.Time{}, false
		}
		return unixAuto(v), true
	case int:
		if v <= 0 {
			return time.Time{}, false
		}
		return unixAuto(int64(v)), true
	}
	return time.Time{}, false
}

func unixAuto(v int64) time.Time {
	if v > 1e12 {
		return time.UnixMilli(v)
	}
	return time.Unix(v, 0)
}

// KiroAccountImport Kiro 账号导入格式（从 Kiro Account Manager 导出）
type KiroAccountImport struct {
	Email    string `json:"email"`
	UserID   string `json:"userId"`
	Nickname string `json:"nickname"`
	IDP      string `json:"idp"`
	Credentials struct {
		AccessToken  string             `json:"accessToken"`
		CSRFToken    string             `json:"csrfToken"`
		RefreshToken string             `json:"refreshToken"`
		ClientID     string             `json:"clientId"`
		ClientSecret string             `json:"clientSecret"`
		Region       string             `json:"region"`
		ExpiresAt    FlexibleImportTime `json:"expiresAt"`
		AuthMethod   string             `json:"authMethod"`
		Provider     string             `json:"provider"`
	} `json:"credentials"`
	Subscription struct {
		Type      string             `json:"type"`
		Title     string             `json:"title"`
		ExpiresAt FlexibleImportTime `json:"expiresAt,omitempty"`
	} `json:"subscription"`
	Usage struct {
		Current     float64 `json:"current"`
		Limit       float64 `json:"limit"`
		PercentUsed float64 `json:"percentUsed"`
	} `json:"usage"`
	Status string `json:"status"`
	ID     string `json:"id"`
}

func normalizeUsagePercentValue(percent float64) float64 {
	if percent <= 0 {
		return 0
	}
	if percent <= 1 {
		return percent * 100
	}
	return percent
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
	now := time.Now()
	for i, acc := range accounts {
		// 计算到期天数
		var daysRemaining *int
		subscriptionExpiresAt := acc.Subscription.ExpiresAt.Ptr()
		if subscriptionExpiresAt != nil {
			days := int(subscriptionExpiresAt.Sub(now).Hours() / 24)
			if subscriptionExpiresAt.After(now) && subscriptionExpiresAt.Sub(now).Hours()/24 > float64(days) {
				days++
			}
			if days < 0 {
				days = 0
			}
			daysRemaining = &days
		}

		// 计算使用百分比
		usagePercent := normalizeUsagePercentValue(acc.Usage.PercentUsed)
		if usagePercent == 0 && acc.Usage.Limit > 0 {
			usagePercent = float64(acc.Usage.Current) / float64(acc.Usage.Limit) * 100
		}

		// 转换过期时间
		expiresAt := acc.Credentials.ExpiresAt.Ptr()

		// 构建 Metadata
		metadata := JSONMap{
			"client_id":     acc.Credentials.ClientID,
			"client_secret": acc.Credentials.ClientSecret,
			"account_name":  acc.Nickname,
			"account_email": acc.Email,
		}
		
		metadata["region"] = acc.Credentials.Region
		metadata["status"] = acc.Status
		metadata["banned"] = acc.Status == "banned"

		// 添加订阅信息
		if acc.Subscription.Type != "" {
			subscription := map[string]interface{}{
				"type":  acc.Subscription.Type,
				"title": acc.Subscription.Title,
			}
			if subscriptionExpiresAt != nil {
				subscription["expires_at"] = subscriptionExpiresAt.Format(time.RFC3339)
			}
			if daysRemaining != nil {
				subscription["days_remaining"] = *daysRemaining
			}
			metadata["subscription"] = subscription
		}
		
		// 添加使用量信息
		if acc.Usage.Limit > 0 {
			metadata["usage"] = map[string]interface{}{
				"current":      acc.Usage.Current,
				"limit":        acc.Usage.Limit,
				"percent":      usagePercent,
				"last_updated": time.Now().Format(time.RFC3339),
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
		
		if cred.Metadata == nil {
			cred.Metadata = make(JSONMap)
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
