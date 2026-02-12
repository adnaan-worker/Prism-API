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
