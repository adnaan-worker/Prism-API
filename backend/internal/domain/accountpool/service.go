package accountpool

import (
	"api-aggregator/backend/pkg/errors"
	"api-aggregator/backend/pkg/query"
	"context"
)

// Service 账号池服务接口
type Service interface {
	CreatePool(ctx context.Context, req *CreatePoolRequest) (*PoolResponse, error)
	UpdatePool(ctx context.Context, id uint, req *UpdatePoolRequest) (*PoolResponse, error)
	DeletePool(ctx context.Context, id uint) error
	GetPool(ctx context.Context, id uint) (*PoolResponse, error)
	ListPools(ctx context.Context, filter *PoolFilter, opts *query.Options) (*PoolListResponse, error)
	UpdatePoolStatus(ctx context.Context, id uint, isActive bool) (*PoolResponse, error)
	GetPoolStats(ctx context.Context, id uint) (*PoolStatsResponse, error)
	
	// 凭据相关
	CreateCredential(ctx context.Context, req *CreateCredentialRequest) (*CredentialResponse, error)
	UpdateCredential(ctx context.Context, id uint, req *UpdateCredentialRequest) (*CredentialResponse, error)
	DeleteCredential(ctx context.Context, id uint) error
	GetCredential(ctx context.Context, id uint) (*CredentialResponse, error)
	ListCredentials(ctx context.Context, filter *CredentialFilter, opts *query.Options) (*CredentialListResponse, error)
	UpdateCredentialStatus(ctx context.Context, id uint, isActive bool) (*CredentialResponse, error)
	RefreshCredential(ctx context.Context, id uint) (*CredentialResponse, error)
	
	// 请求日志相关
	CreateRequestLog(ctx context.Context, log *AccountPoolRequestLog) error
	ListRequestLogs(ctx context.Context, filter *RequestLogFilter, opts *query.Options) (*RequestLogListResponse, error)
}

type service struct {
	repo Repository
}

// NewService 创建账号池服务实例
func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

// CreatePool 创建账号池
func (s *service) CreatePool(ctx context.Context, req *CreatePoolRequest) (*PoolResponse, error) {
	// 设置默认值
	if req.Strategy == "" {
		req.Strategy = StrategyRoundRobin
	}
	if req.HealthCheckInterval == 0 {
		req.HealthCheckInterval = 300
	}
	if req.HealthCheckTimeout == 0 {
		req.HealthCheckTimeout = 10
	}
	if req.MaxRetries == 0 {
		req.MaxRetries = 3
	}

	// 验证策略
	if !IsValidStrategy(req.Strategy) {
		return nil, errors.NewValidationError("invalid strategy", map[string]string{
			"strategy": "must be one of: round_robin, weighted_round_robin, least_connections, random",
		})
	}

	// 创建账号池
	pool := &AccountPool{
		Name:                req.Name,
		Description:         req.Description,
		Provider:            req.Provider,
		Strategy:            req.Strategy,
		HealthCheckInterval: req.HealthCheckInterval,
		HealthCheckTimeout:  req.HealthCheckTimeout,
		MaxRetries:          req.MaxRetries,
		IsActive:            true,
	}

	if err := s.repo.Create(ctx, pool); err != nil {
		return nil, errors.Wrap(err, "failed to create pool")
	}

	return ToPoolResponse(pool), nil
}

// UpdatePool 更新账号池
func (s *service) UpdatePool(ctx context.Context, id uint, req *UpdatePoolRequest) (*PoolResponse, error) {
	// 查找账号池
	pool, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.NewNotFoundError("account pool not found")
	}

	// 更新字段
	if req.Name != nil {
		pool.Name = *req.Name
	}
	if req.Description != nil {
		pool.Description = *req.Description
	}
	if req.Strategy != nil {
		if !IsValidStrategy(*req.Strategy) {
			return nil, errors.NewValidationError("invalid strategy", map[string]string{
				"strategy": "must be one of: round_robin, weighted_round_robin, least_connections, random",
			})
		}
		pool.Strategy = *req.Strategy
	}
	if req.HealthCheckInterval != nil {
		pool.HealthCheckInterval = *req.HealthCheckInterval
	}
	if req.HealthCheckTimeout != nil {
		pool.HealthCheckTimeout = *req.HealthCheckTimeout
	}
	if req.MaxRetries != nil {
		pool.MaxRetries = *req.MaxRetries
	}
	if req.IsActive != nil {
		pool.IsActive = *req.IsActive
	}

	// 保存更新
	if err := s.repo.Update(ctx, pool); err != nil {
		return nil, errors.Wrap(err, "failed to update pool")
	}

	return ToPoolResponse(pool), nil
}

// DeletePool 删除账号池
func (s *service) DeletePool(ctx context.Context, id uint) error {
	// 检查账号池是否存在
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return errors.NewNotFoundError("account pool not found")
	}

	// 删除账号池
	if err := s.repo.Delete(ctx, id); err != nil {
		return errors.Wrap(err, "failed to delete pool")
	}

	return nil
}

// GetPool 获取账号池
func (s *service) GetPool(ctx context.Context, id uint) (*PoolResponse, error) {
	pool, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.NewNotFoundError("account pool not found")
	}

	return ToPoolResponse(pool), nil
}

// ListPools 查询账号池列表
func (s *service) ListPools(ctx context.Context, filter *PoolFilter, opts *query.Options) (*PoolListResponse, error) {
	pools, total, err := s.repo.List(ctx, filter, opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list pools")
	}

	return ToPoolListResponse(pools, total), nil
}

// UpdatePoolStatus 更新账号池状态
func (s *service) UpdatePoolStatus(ctx context.Context, id uint, isActive bool) (*PoolResponse, error) {
	pool, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.NewNotFoundError("account pool not found")
	}

	pool.IsActive = isActive

	if err := s.repo.Update(ctx, pool); err != nil {
		return nil, errors.Wrap(err, "failed to update pool status")
	}

	return ToPoolResponse(pool), nil
}

// GetPoolStats 获取账号池统计
func (s *service) GetPoolStats(ctx context.Context, id uint) (*PoolStatsResponse, error) {
	pool, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.NewNotFoundError("account pool not found")
	}

	// 获取请求统计
	requestStats, err := s.repo.GetPoolRequestStats(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get pool request stats")
	}

	// 构建响应
	stats := &PoolStatsResponse{
		PoolID:        pool.ID,
		PoolName:      pool.Name,
		Provider:      pool.Provider,
		TotalRequests: pool.TotalRequests,
		TotalErrors:   pool.TotalErrors,
		ErrorRate:     pool.GetErrorRate(),
		IsHealthy:     pool.IsHealthy(),
	}

	// 添加请求统计信息
	if totalReqs, ok := requestStats["total_requests"].(int64); ok {
		stats.TotalRequests = totalReqs
	}

	return stats, nil
}

// CreateRequestLog 创建请求日志
func (s *service) CreateRequestLog(ctx context.Context, log *AccountPoolRequestLog) error {
	if err := s.repo.CreateRequestLog(ctx, log); err != nil {
		return errors.Wrap(err, "failed to create request log")
	}

	// 更新账号池统计
	if log.PoolID != nil {
		if err := s.repo.IncrementRequests(ctx, *log.PoolID); err != nil {
			// 记录错误但不影响主流程
			return errors.Wrap(err, "failed to increment pool requests")
		}

		if log.IsError() {
			if err := s.repo.IncrementErrors(ctx, *log.PoolID); err != nil {
				return errors.Wrap(err, "failed to increment pool errors")
			}
		}
	}

	return nil
}

// ListRequestLogs 查询请求日志列表
func (s *service) ListRequestLogs(ctx context.Context, filter *RequestLogFilter, opts *query.Options) (*RequestLogListResponse, error) {
	logs, total, err := s.repo.ListRequestLogs(ctx, filter, opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list request logs")
	}

	return ToRequestLogListResponse(logs, total), nil
}

// CreateCredential 创建凭据
func (s *service) CreateCredential(ctx context.Context, req *CreateCredentialRequest) (*CredentialResponse, error) {
	// 验证账号池是否存在
	_, err := s.repo.FindByID(ctx, req.PoolID)
	if err != nil {
		return nil, errors.NewNotFoundError("account pool not found")
	}

	// 验证认证类型
	if !IsValidAuthType(req.AuthType) {
		return nil, errors.NewValidationError("invalid auth type", map[string]string{
			"auth_type": "must be one of: api_key, oauth, session_token",
		})
	}

	// 设置默认值
	if req.Weight == 0 {
		req.Weight = 1
	}

	// 创建凭据
	cred := &AccountCredential{
		PoolID:       req.PoolID,
		Provider:     req.Provider,
		AuthType:     req.AuthType,
		APIKey:       req.APIKey,
		AccessToken:  req.AccessToken,
		RefreshToken: req.RefreshToken,
		SessionToken: req.SessionToken,
		AccountName:  req.AccountName,
		AccountEmail: req.AccountEmail,
		Weight:       req.Weight,
		IsActive:     true,
		HealthStatus: HealthStatusUnknown,
		RateLimit:    req.RateLimit,
	}

	if err := s.repo.CreateCredential(ctx, cred); err != nil {
		return nil, errors.Wrap(err, "failed to create credential")
	}

	return ToCredentialResponse(cred), nil
}

// UpdateCredential 更新凭据
func (s *service) UpdateCredential(ctx context.Context, id uint, req *UpdateCredentialRequest) (*CredentialResponse, error) {
	// 查找凭据
	cred, err := s.repo.FindCredentialByID(ctx, id)
	if err != nil {
		return nil, errors.NewNotFoundError("credential not found")
	}

	// 更新字段
	if req.APIKey != nil {
		cred.APIKey = *req.APIKey
	}
	if req.AccessToken != nil {
		cred.AccessToken = *req.AccessToken
	}
	if req.RefreshToken != nil {
		cred.RefreshToken = *req.RefreshToken
	}
	if req.SessionToken != nil {
		cred.SessionToken = *req.SessionToken
	}
	if req.AccountName != nil {
		cred.AccountName = *req.AccountName
	}
	if req.AccountEmail != nil {
		cred.AccountEmail = *req.AccountEmail
	}
	if req.Weight != nil {
		cred.Weight = *req.Weight
	}
	if req.IsActive != nil {
		cred.IsActive = *req.IsActive
	}
	if req.RateLimit != nil {
		cred.RateLimit = *req.RateLimit
	}

	// 保存更新
	if err := s.repo.UpdateCredential(ctx, cred); err != nil {
		return nil, errors.Wrap(err, "failed to update credential")
	}

	return ToCredentialResponse(cred), nil
}

// DeleteCredential 删除凭据
func (s *service) DeleteCredential(ctx context.Context, id uint) error {
	// 检查凭据是否存在
	_, err := s.repo.FindCredentialByID(ctx, id)
	if err != nil {
		return errors.NewNotFoundError("credential not found")
	}

	// 删除凭据
	if err := s.repo.DeleteCredential(ctx, id); err != nil {
		return errors.Wrap(err, "failed to delete credential")
	}

	return nil
}

// GetCredential 获取凭据
func (s *service) GetCredential(ctx context.Context, id uint) (*CredentialResponse, error) {
	cred, err := s.repo.FindCredentialByID(ctx, id)
	if err != nil {
		return nil, errors.NewNotFoundError("credential not found")
	}

	return ToCredentialResponse(cred), nil
}

// ListCredentials 查询凭据列表
func (s *service) ListCredentials(ctx context.Context, filter *CredentialFilter, opts *query.Options) (*CredentialListResponse, error) {
	creds, total, err := s.repo.ListCredentials(ctx, filter, opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list credentials")
	}

	return ToCredentialListResponse(creds, total), nil
}

// UpdateCredentialStatus 更新凭据状态
func (s *service) UpdateCredentialStatus(ctx context.Context, id uint, isActive bool) (*CredentialResponse, error) {
	cred, err := s.repo.FindCredentialByID(ctx, id)
	if err != nil {
		return nil, errors.NewNotFoundError("credential not found")
	}

	if err := s.repo.UpdateCredentialStatus(ctx, id, isActive); err != nil {
		return nil, errors.Wrap(err, "failed to update credential status")
	}

	cred.IsActive = isActive

	return ToCredentialResponse(cred), nil
}

// RefreshCredential 刷新凭据
func (s *service) RefreshCredential(ctx context.Context, id uint) (*CredentialResponse, error) {
	cred, err := s.repo.FindCredentialByID(ctx, id)
	if err != nil {
		return nil, errors.NewNotFoundError("credential not found")
	}

	// TODO: 实现实际的刷新逻辑，根据不同的认证类型调用相应的刷新接口
	// 这里只是更新健康状态和检查时间
	cred.UpdateHealthStatus(HealthStatusHealthy)

	if err := s.repo.UpdateCredential(ctx, cred); err != nil {
		return nil, errors.Wrap(err, "failed to refresh credential")
	}

	return ToCredentialResponse(cred), nil
}
