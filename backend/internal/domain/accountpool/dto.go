package accountpool

import "time"

// CreatePoolRequest 创建账号池请求
type CreatePoolRequest struct {
	Name                string `json:"name" binding:"required"`
	Description         string `json:"description"`
	Provider            string `json:"provider" binding:"required"`
	Strategy            string `json:"strategy"`
	HealthCheckInterval int    `json:"health_check_interval"`
	HealthCheckTimeout  int    `json:"health_check_timeout"`
	MaxRetries          int    `json:"max_retries"`
}

// UpdatePoolRequest 更新账号池请求
type UpdatePoolRequest struct {
	Name                *string `json:"name"`
	Description         *string `json:"description"`
	Strategy            *string `json:"strategy"`
	HealthCheckInterval *int    `json:"health_check_interval"`
	HealthCheckTimeout  *int    `json:"health_check_timeout"`
	MaxRetries          *int    `json:"max_retries"`
	IsActive            *bool   `json:"is_active"`
}

// UpdatePoolStatusRequest 更新账号池状态请求
type UpdatePoolStatusRequest struct {
	IsActive bool `json:"is_active" binding:"required"`
}

// PoolResponse 账号池响应
type PoolResponse struct {
	ID                  uint      `json:"id"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
	Name                string    `json:"name"`
	Description         string    `json:"description,omitempty"`
	Provider            string    `json:"provider"`
	Strategy            string    `json:"strategy"`
	HealthCheckInterval int       `json:"health_check_interval"`
	HealthCheckTimeout  int       `json:"health_check_timeout"`
	MaxRetries          int       `json:"max_retries"`
	IsActive            bool      `json:"is_active"`
	TotalRequests       int64     `json:"total_requests"`
	TotalErrors         int64     `json:"total_errors"`
	ErrorRate           float64   `json:"error_rate"`
}

// PoolListResponse 账号池列表响应
type PoolListResponse struct {
	Pools []*PoolResponse `json:"pools"`
	Total int64           `json:"total"`
}

// PoolStatsResponse 账号池统计响应
type PoolStatsResponse struct {
	PoolID        uint    `json:"pool_id"`
	PoolName      string  `json:"pool_name"`
	Provider      string  `json:"provider"`
	TotalCreds    int     `json:"total_creds"`
	ActiveCreds   int     `json:"active_creds"`
	TotalRequests int64   `json:"total_requests"`
	TotalErrors   int64   `json:"total_errors"`
	ErrorRate     float64 `json:"error_rate"`
	IsHealthy     bool    `json:"is_healthy"`
}

// RequestLogResponse 请求日志响应
type RequestLogResponse struct {
	ID           uint      `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	CredentialID *uint     `json:"credential_id,omitempty"`
	PoolID       *uint     `json:"pool_id,omitempty"`
	Provider     string    `json:"provider"`
	Model        string    `json:"model"`
	Method       string    `json:"method"`
	StatusCode   int       `json:"status_code,omitempty"`
	ResponseTime int       `json:"response_time,omitempty"`
	TokensUsed   int       `json:"tokens_used,omitempty"`
	ErrorMessage string    `json:"error_message,omitempty"`
	RequestLogID *uint     `json:"request_log_id,omitempty"`
}

// RequestLogListResponse 请求日志列表响应
type RequestLogListResponse struct {
	Logs  []*RequestLogResponse `json:"logs"`
	Total int64                 `json:"total"`
}

// PoolFilter 账号池过滤器
type PoolFilter struct {
	Provider *string
	Strategy *string
	IsActive *bool
}

// RequestLogFilter 请求日志过滤器
type RequestLogFilter struct {
	PoolID       *uint
	CredentialID *uint
	Provider     *string
	Model        *string
	StatusCode   *int
}

// ToPoolResponse 转换为账号池响应
func ToPoolResponse(pool *AccountPool) *PoolResponse {
	if pool == nil {
		return nil
	}
	return &PoolResponse{
		ID:                  pool.ID,
		CreatedAt:           pool.CreatedAt,
		UpdatedAt:           pool.UpdatedAt,
		Name:                pool.Name,
		Description:         pool.Description,
		Provider:            pool.Provider,
		Strategy:            pool.Strategy,
		HealthCheckInterval: pool.HealthCheckInterval,
		HealthCheckTimeout:  pool.HealthCheckTimeout,
		MaxRetries:          pool.MaxRetries,
		IsActive:            pool.IsActive,
		TotalRequests:       pool.TotalRequests,
		TotalErrors:         pool.TotalErrors,
		ErrorRate:           pool.GetErrorRate(),
	}
}

// ToPoolListResponse 转换为账号池列表响应
func ToPoolListResponse(pools []*AccountPool, total int64) *PoolListResponse {
	responses := make([]*PoolResponse, len(pools))
	for i, pool := range pools {
		responses[i] = ToPoolResponse(pool)
	}
	return &PoolListResponse{
		Pools: responses,
		Total: total,
	}
}

// ToRequestLogResponse 转换为请求日志响应
func ToRequestLogResponse(log *AccountPoolRequestLog) *RequestLogResponse {
	if log == nil {
		return nil
	}
	return &RequestLogResponse{
		ID:           log.ID,
		CreatedAt:    log.CreatedAt,
		CredentialID: log.CredentialID,
		PoolID:       log.PoolID,
		Provider:     log.Provider,
		Model:        log.Model,
		Method:       log.Method,
		StatusCode:   log.StatusCode,
		ResponseTime: log.ResponseTime,
		TokensUsed:   log.TokensUsed,
		ErrorMessage: log.ErrorMessage,
		RequestLogID: log.RequestLogID,
	}
}

// ToRequestLogListResponse 转换为请求日志列表响应
func ToRequestLogListResponse(logs []*AccountPoolRequestLog, total int64) *RequestLogListResponse {
	responses := make([]*RequestLogResponse, len(logs))
	for i, log := range logs {
		responses[i] = ToRequestLogResponse(log)
	}
	return &RequestLogListResponse{
		Logs:  responses,
		Total: total,
	}
}

// CreateCredentialRequest 创建凭据请求
type CreateCredentialRequest struct {
	PoolID       uint   `json:"pool_id" binding:"required"`
	Provider     string `json:"provider" binding:"required"`
	AuthType     string `json:"auth_type" binding:"required"`
	APIKey       string `json:"api_key"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	SessionToken string `json:"session_token"`
	AccountName  string `json:"account_name"`
	AccountEmail string `json:"account_email"`
	Weight       int    `json:"weight"`
	RateLimit    int    `json:"rate_limit"`
}

// UpdateCredentialRequest 更新凭据请求
type UpdateCredentialRequest struct {
	APIKey       *string `json:"api_key"`
	AccessToken  *string `json:"access_token"`
	RefreshToken *string `json:"refresh_token"`
	SessionToken *string `json:"session_token"`
	AccountName  *string `json:"account_name"`
	AccountEmail *string `json:"account_email"`
	Weight       *int    `json:"weight"`
	IsActive     *bool   `json:"is_active"`
	RateLimit    *int    `json:"rate_limit"`
}

// UpdateCredentialStatusRequest 更新凭据状态请求
type UpdateCredentialStatusRequest struct {
	IsActive bool `json:"is_active" binding:"required"`
}

// RefreshCredentialRequest 刷新凭据请求
type RefreshCredentialRequest struct {
	// 可以添加刷新相关的参数
}

// CredentialResponse 凭据响应
type CredentialResponse struct {
	ID            uint       `json:"id"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	PoolID        uint       `json:"pool_id"`
	Provider      string     `json:"provider"`
	AuthType      string     `json:"auth_type"`
	AccountName   string     `json:"account_name,omitempty"`
	AccountEmail  string     `json:"account_email,omitempty"`
	Weight        int        `json:"weight"`
	IsActive      bool       `json:"is_active"`
	Status        string     `json:"status"`
	LastError     string     `json:"last_error,omitempty"`
	HealthStatus  string     `json:"health_status"`
	LastCheckedAt *time.Time `json:"last_checked_at,omitempty"`
	LastUsedAt    *time.Time `json:"last_used_at,omitempty"`
	TotalRequests int64      `json:"total_requests"`
	TotalErrors   int64      `json:"total_errors"`
	ErrorRate     float64    `json:"error_rate"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	IsExpired     bool       `json:"is_expired"`
	RateLimit     int        `json:"rate_limit"`
	CurrentUsage  int        `json:"current_usage"`
	
	// 订阅信息
	SubscriptionType          string     `json:"subscription_type,omitempty"`
	SubscriptionTitle         string     `json:"subscription_title,omitempty"`
	SubscriptionExpiresAt     *time.Time `json:"subscription_expires_at,omitempty"`
	SubscriptionDaysRemaining *int       `json:"subscription_days_remaining,omitempty"`
	
	// 使用量详情
	UsageCurrent        int        `json:"usage_current"`
	UsageLimit          int        `json:"usage_limit"`
	UsagePercent        float64    `json:"usage_percent"`
	UsageLastUpdated    *time.Time `json:"usage_last_updated,omitempty"`
	BaseLimit           int        `json:"base_limit"`
	BaseCurrent         int        `json:"base_current"`
	FreeTrialLimit      int        `json:"free_trial_limit"`
	FreeTrialCurrent    int        `json:"free_trial_current"`
	FreeTrialExpiry     *time.Time `json:"free_trial_expiry,omitempty"`
	NextResetDate       *time.Time `json:"next_reset_date,omitempty"`
	
	// 机器码
	MachineID string `json:"machine_id,omitempty"`
	// 敏感信息不返回
}

// CredentialListResponse 凭据列表响应
type CredentialListResponse struct {
	Credentials []*CredentialResponse `json:"credentials"`
	Total       int64                 `json:"total"`
}

// CredentialFilter 凭据过滤器
type CredentialFilter struct {
	PoolID       *uint
	Provider     *string
	AuthType     *string
	IsActive     *bool
	HealthStatus *string
}

// ToCredentialResponse 转换为凭据响应
func ToCredentialResponse(cred *AccountCredential) *CredentialResponse {
	if cred == nil {
		return nil
	}
	return &CredentialResponse{
		ID:            cred.ID,
		CreatedAt:     cred.CreatedAt,
		UpdatedAt:     cred.UpdatedAt,
		PoolID:        cred.PoolID,
		Provider:      cred.Provider,
		AuthType:      cred.AuthType,
		AccountName:   cred.AccountName,
		AccountEmail:  cred.AccountEmail,
		Weight:        cred.Weight,
		IsActive:      cred.IsActive,
		Status:        cred.Status,
		LastError:     cred.LastError,
		HealthStatus:  cred.HealthStatus,
		LastCheckedAt: cred.LastCheckedAt,
		LastUsedAt:    cred.LastUsedAt,
		TotalRequests: cred.TotalRequests,
		TotalErrors:   cred.TotalErrors,
		ErrorRate:     cred.GetErrorRate(),
		ExpiresAt:     cred.ExpiresAt,
		IsExpired:     cred.IsExpired(),
		RateLimit:     cred.RateLimit,
		CurrentUsage:  cred.CurrentUsage,
		
		// 订阅信息
		SubscriptionType:          cred.SubscriptionType,
		SubscriptionTitle:         cred.SubscriptionTitle,
		SubscriptionExpiresAt:     cred.SubscriptionExpiresAt,
		SubscriptionDaysRemaining: cred.SubscriptionDaysRemaining,
		
		// 使用量详情
		UsageCurrent:        cred.UsageCurrent,
		UsageLimit:          cred.UsageLimit,
		UsagePercent:        cred.UsagePercent,
		UsageLastUpdated:    cred.UsageLastUpdated,
		BaseLimit:           cred.BaseLimit,
		BaseCurrent:         cred.BaseCurrent,
		FreeTrialLimit:      cred.FreeTrialLimit,
		FreeTrialCurrent:    cred.FreeTrialCurrent,
		FreeTrialExpiry:     cred.FreeTrialExpiry,
		NextResetDate:       cred.NextResetDate,
		
		// 机器码
		MachineID: cred.MachineID,
	}
}

// ToCredentialListResponse 转换为凭据列表响应
func ToCredentialListResponse(creds []*AccountCredential, total int64) *CredentialListResponse {
	responses := make([]*CredentialResponse, len(creds))
	for i, cred := range creds {
		responses[i] = ToCredentialResponse(cred)
	}
	return &CredentialListResponse{
		Credentials: responses,
		Total:       total,
	}
}
