package accountpool

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

// KiroRefreshService Kiro token 刷新服务
type KiroRefreshService struct {
	client *http.Client
}

type kiroRefreshResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
	TokenType    string `json:"tokenType"`
}

type kiroFreeTrialInfo struct {
	CurrentUsage              float64     `json:"currentUsage"`
	CurrentUsageWithPrecision float64     `json:"currentUsageWithPrecision"`
	UsageLimit                float64     `json:"usageLimit"`
	UsageLimitWithPrecision   float64     `json:"usageLimitWithPrecision"`
	FreeTrialStatus           string      `json:"freeTrialStatus"`
	FreeTrialExpiry           interface{} `json:"freeTrialExpiry"`
}

type kiroBonusInfo struct {
	BonusCode                 string      `json:"bonusCode"`
	DisplayName               string      `json:"displayName"`
	UsageLimit                float64     `json:"usageLimit"`
	UsageLimitWithPrecision   float64     `json:"usageLimitWithPrecision"`
	CurrentUsage              float64     `json:"currentUsage"`
	CurrentUsageWithPrecision float64     `json:"currentUsageWithPrecision"`
	ExpiresAt                 interface{} `json:"expiresAt"`
	Status                    string      `json:"status"`
}

type kiroUsageBreakdown struct {
	Type                      string             `json:"type"`
	ResourceType              string             `json:"resourceType"`
	DisplayName               string             `json:"displayName"`
	DisplayNamePlural         string             `json:"displayNamePlural"`
	CurrentUsage              float64            `json:"currentUsage"`
	CurrentUsageWithPrecision float64            `json:"currentUsageWithPrecision"`
	UsageLimit                float64            `json:"usageLimit"`
	UsageLimitWithPrecision   float64            `json:"usageLimitWithPrecision"`
	Currency                  string             `json:"currency"`
	Unit                      string             `json:"unit"`
	OverageRate               float64            `json:"overageRate"`
	OverageCap                float64            `json:"overageCap"`
	FreeTrialInfo             *kiroFreeTrialInfo `json:"freeTrialInfo"`
	Bonuses                   []kiroBonusInfo    `json:"bonuses"`
}

type kiroSubscriptionInfo struct {
	SubscriptionName             string `json:"subscriptionName"`
	SubscriptionTitle            string `json:"subscriptionTitle"`
	SubscriptionType             string `json:"subscriptionType"`
	Status                       string `json:"status"`
	Type                         string `json:"type"`
	SubscriptionManagementTarget string `json:"subscriptionManagementTarget"`
	UpgradeCapability            string `json:"upgradeCapability"`
	OverageCapability            string `json:"overageCapability"`
}

type kiroUsageLimitsResponse struct {
	UsageBreakdownList  []kiroUsageBreakdown `json:"usageBreakdownList"`
	NextDateReset       interface{}          `json:"nextDateReset"`
	SubscriptionInfo    *kiroSubscriptionInfo `json:"subscriptionInfo"`
	OverageConfiguration *struct {
		OverageEnabled bool `json:"overageEnabled"`
	} `json:"overageConfiguration"`
	UserInfo *struct {
		Email  string `json:"email"`
		UserID string `json:"userId"`
	} `json:"userInfo"`
}

// NewKiroRefreshService 创建 Kiro 刷新服务
func NewKiroRefreshService() *KiroRefreshService {
	return &KiroRefreshService{
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

// RefreshKiroToken 刷新 Kiro token，并同步订阅/用量/封禁状态
func (s *KiroRefreshService) RefreshKiroToken(ctx context.Context, cred *AccountCredential) error {
	if cred.Provider != "kiro" {
		return fmt.Errorf("credential is not kiro type")
	}
	if cred.RefreshToken == "" {
		return fmt.Errorf("refresh token is empty")
	}
	if cred.Metadata == nil {
		cred.Metadata = make(JSONMap)
	}

	clientID, _ := cred.Metadata["client_id"].(string)
	clientSecret, _ := cred.Metadata["client_secret"].(string)
	if clientID == "" || clientSecret == "" {
		return fmt.Errorf("missing client_id or client_secret in metadata")
	}

	region := s.getRegion(cred)
	refreshResp, err := s.refreshAccessToken(ctx, cred.RefreshToken, clientID, clientSecret, region)
	if err != nil {
		return err
	}

	if refreshResp.AccessToken != "" {
		cred.AccessToken = refreshResp.AccessToken
	}
	if refreshResp.RefreshToken != "" {
		cred.RefreshToken = refreshResp.RefreshToken
	}
	if refreshResp.ExpiresIn > 0 {
		expiresAt := time.Now().Add(time.Duration(refreshResp.ExpiresIn) * time.Second)
		cred.ExpiresAt = &expiresAt
	}

	if err := s.syncAccountState(ctx, cred, region); err != nil {
		return err
	}

	status, _ := cred.Metadata["status"].(string)
	if status == "banned" {
		return nil
	}

	cred.HealthStatus = HealthStatusHealthy
	cred.LastError = ""
	cred.IsActive = true
	return nil
}

func (s *KiroRefreshService) refreshAccessToken(ctx context.Context, refreshToken, clientID, clientSecret, region string) (*kiroRefreshResponse, error) {
	endpoint := fmt.Sprintf("https://oidc.%s.amazonaws.com/token", region)
	reqBody := map[string]string{
		"clientId":     clientID,
		"clientSecret": clientSecret,
		"refreshToken": refreshToken,
		"grantType":    "refresh_token",
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("refresh failed with status %d: %s", resp.StatusCode, string(body))
	}

	var refreshResp kiroRefreshResponse
	if err := json.Unmarshal(body, &refreshResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &refreshResp, nil
}

func (s *KiroRefreshService) syncAccountState(ctx context.Context, cred *AccountCredential, region string) error {
	usageResp, err := s.fetchUsageLimits(ctx, cred, region)
	if err != nil {
		if isKiroBannedError(err) {
			cred.HealthStatus = HealthStatusUnhealthy
			cred.LastError = err.Error()
			cred.IsActive = false
			cred.Metadata["status"] = "banned"
			cred.Metadata["banned"] = true
			return nil
		}
		return err
	}

	now := time.Now()
	cred.Metadata["status"] = "active"
	cred.Metadata["banned"] = false
	cred.Metadata["last_synced_at"] = now.Format(time.RFC3339)

	if usageResp.UserInfo != nil {
		if usageResp.UserInfo.Email != "" {
			cred.Metadata["account_email"] = usageResp.UserInfo.Email
		}
		if usageResp.UserInfo.UserID != "" {
			cred.Metadata["user_id"] = usageResp.UserInfo.UserID
		}
	}

	cred.Metadata["subscription"] = buildSubscriptionMetadata(usageResp)
	cred.Metadata["usage"] = buildUsageMetadata(usageResp, now)
	return nil
}

func (s *KiroRefreshService) fetchUsageLimits(ctx context.Context, cred *AccountCredential, region string) (*kiroUsageLimitsResponse, error) {
	machineID, _ := cred.Metadata["machine_id"].(string)
	if machineID == "" {
		machineID = uuid.New().String()
		cred.Metadata["machine_id"] = machineID
	}

	baseURL := fmt.Sprintf("https://q.%s.amazonaws.com", region)
	url := fmt.Sprintf("%s/getUsageLimits?origin=AI_EDITOR&resourceType=AGENTIC_REQUEST&isEmailRequired=true", baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create usage request: %w", err)
	}
	setKiroRESTHeaders(req, cred.AccessToken, machineID)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send usage request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read usage response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("usage sync failed with status %d: %s", resp.StatusCode, string(body))
	}

	var usageResp kiroUsageLimitsResponse
	if err := json.Unmarshal(body, &usageResp); err != nil {
		return nil, fmt.Errorf("failed to parse usage response: %w", err)
	}
	return &usageResp, nil
}

func (s *KiroRefreshService) getRegion(cred *AccountCredential) string {
	if region, _ := cred.Metadata["region"].(string); region != "" {
		return region
	}
	return "us-east-1"
}

func setKiroRESTHeaders(req *http.Request, accessToken, machineID string) {
	kiroVersion := "0.6.18"
	userAgent := fmt.Sprintf("aws-sdk-js/1.0.18 ua/2.1 os/windows lang/js md/nodejs#20.16.0 api/codewhispererstreaming#1.0.18 m/E KiroIDE-%s-%s", kiroVersion, machineID)
	amzUserAgent := fmt.Sprintf("aws-sdk-js/1.0.18 KiroIDE-%s %s", kiroVersion, machineID)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("x-amz-user-agent", amzUserAgent)
	req.Header.Set("x-amzn-codewhisperer-optout-preference", "OPTIN")
}

func isKiroBannedError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "403") || strings.Contains(msg, "423") || strings.Contains(msg, "AccountSuspendedException")
}

func buildSubscriptionMetadata(resp *kiroUsageLimitsResponse) JSONMap {
	now := time.Now()
	result := JSONMap{}
	if resp == nil || resp.SubscriptionInfo == nil {
		result["type"] = "Free"
		result["title"] = "Free"
		return result
	}

	title := resp.SubscriptionInfo.SubscriptionTitle
	if title == "" {
		title = "Free"
	}
	typeName := normalizeSubscriptionType(title, resp.SubscriptionInfo.Type, resp.SubscriptionInfo.SubscriptionType)
	result["type"] = typeName
	result["title"] = title
	result["raw_type"] = firstNonEmpty(resp.SubscriptionInfo.Type, resp.SubscriptionInfo.SubscriptionType)
	result["status"] = resp.SubscriptionInfo.Status
	result["management_target"] = resp.SubscriptionInfo.SubscriptionManagementTarget
	result["upgrade_capability"] = resp.SubscriptionInfo.UpgradeCapability
	result["overage_capability"] = resp.SubscriptionInfo.OverageCapability

	if nextReset := parseFlexibleTime(resp.NextDateReset); nextReset != nil {
		result["expires_at"] = nextReset.Format(time.RFC3339)
		days := int(nextReset.Sub(now).Hours() / 24)
		if nextReset.After(now) && nextReset.Sub(now).Hours()/24 > float64(days) {
			days++
		}
		if days < 0 {
			days = 0
		}
		result["days_remaining"] = days
	}
	return result
}

func buildUsageMetadata(resp *kiroUsageLimitsResponse, now time.Time) JSONMap {
	result := JSONMap{
		"current":            0,
		"limit":              0,
		"percent":            0.0,
		"last_updated":       now.Format(time.RFC3339),
		"base_limit":         0,
		"base_current":       0,
		"free_trial_limit":   0,
		"free_trial_current": 0,
	}
	if resp == nil {
		return result
	}

	var credit *kiroUsageBreakdown
	for i := range resp.UsageBreakdownList {
		item := &resp.UsageBreakdownList[i]
		if item.ResourceType == "CREDIT" || strings.EqualFold(item.DisplayName, "Credits") {
			credit = item
			break
		}
	}
	if credit == nil {
		if nextReset := parseFlexibleTime(resp.NextDateReset); nextReset != nil {
			result["next_reset_date"] = nextReset.Format(time.RFC3339)
		}
		return result
	}

	baseCurrent := pickFloat(credit.CurrentUsageWithPrecision, credit.CurrentUsage)
	baseLimit := pickFloat(credit.UsageLimitWithPrecision, credit.UsageLimit)
	freeTrialCurrent := 0.0
	freeTrialLimit := 0.0
	if credit.FreeTrialInfo != nil && strings.EqualFold(credit.FreeTrialInfo.FreeTrialStatus, "ACTIVE") {
		freeTrialCurrent = pickFloat(credit.FreeTrialInfo.CurrentUsageWithPrecision, credit.FreeTrialInfo.CurrentUsage)
		freeTrialLimit = pickFloat(credit.FreeTrialInfo.UsageLimitWithPrecision, credit.FreeTrialInfo.UsageLimit)
		if expiry := parseFlexibleTime(credit.FreeTrialInfo.FreeTrialExpiry); expiry != nil {
			result["free_trial_expiry"] = expiry.Format(time.RFC3339)
		}
	}

	bonuses := make([]map[string]interface{}, 0)
	bonusCurrent := 0.0
	bonusLimit := 0.0
	for _, bonus := range credit.Bonuses {
		if !strings.EqualFold(bonus.Status, "ACTIVE") {
			continue
		}
		current := pickFloat(bonus.CurrentUsageWithPrecision, bonus.CurrentUsage)
		limit := pickFloat(bonus.UsageLimitWithPrecision, bonus.UsageLimit)
		item := map[string]interface{}{
			"code":    bonus.BonusCode,
			"name":    bonus.DisplayName,
			"current": int(current),
			"limit":   int(limit),
		}
		if expiresAt := parseFlexibleTime(bonus.ExpiresAt); expiresAt != nil {
			item["expires_at"] = expiresAt.Format(time.RFC3339)
		}
		bonuses = append(bonuses, item)
		bonusCurrent += current
		bonusLimit += limit
	}

	totalCurrent := baseCurrent + freeTrialCurrent + bonusCurrent
	totalLimit := baseLimit + freeTrialLimit + bonusLimit
	percent := 0.0
	if totalLimit > 0 {
		percent = totalCurrent / totalLimit * 100
	}

	result["current"] = int(totalCurrent)
	result["limit"] = int(totalLimit)
	result["percent"] = percent
	result["base_current"] = int(baseCurrent)
	result["base_limit"] = int(baseLimit)
	result["free_trial_current"] = int(freeTrialCurrent)
	result["free_trial_limit"] = int(freeTrialLimit)
	result["bonuses"] = bonuses
	result["resource_detail"] = JSONMap{
		"resource_type":       credit.ResourceType,
		"display_name":        credit.DisplayName,
		"display_name_plural": credit.DisplayNamePlural,
		"currency":            credit.Currency,
		"unit":                credit.Unit,
		"overage_rate":        credit.OverageRate,
		"overage_cap":         credit.OverageCap,
		"overage_enabled":     resp.OverageConfiguration != nil && resp.OverageConfiguration.OverageEnabled,
	}
	if nextReset := parseFlexibleTime(resp.NextDateReset); nextReset != nil {
		result["next_reset_date"] = nextReset.Format(time.RFC3339)
	}
	return result
}

func normalizeSubscriptionType(values ...string) string {
	joined := strings.ToUpper(strings.Join(values, " "))
	switch {
	case strings.Contains(joined, "PRO+") || strings.Contains(joined, "PRO_PLUS") || strings.Contains(joined, "PROPLUS"):
		return "Pro_Plus"
	case strings.Contains(joined, "POWER") || strings.Contains(joined, "ENTERPRISE"):
		return "Enterprise"
	case strings.Contains(joined, "TEAMS"):
		return "Teams"
	case strings.Contains(joined, "PRO"):
		return "Pro"
	default:
		return "Free"
	}
}

func parseFlexibleTime(value interface{}) *time.Time {
	switch v := value.(type) {
	case string:
		if v == "" {
			return nil
		}
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			return &t
		}
	case float64:
		if v <= 0 {
			return nil
		}
		t := time.Unix(int64(v), 0)
		return &t
	case int64:
		if v <= 0 {
			return nil
		}
		t := time.Unix(v, 0)
		return &t
	case int:
		if v <= 0 {
			return nil
		}
		t := time.Unix(int64(v), 0)
		return &t
	}
	return nil
}

func pickFloat(primary, fallback float64) float64 {
	if primary != 0 {
		return primary
	}
	return fallback
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
