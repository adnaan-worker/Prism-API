package service

import (
	"api-aggregator/backend/internal/models"
	"api-aggregator/backend/internal/repository"
	"context"
	"fmt"
)

type AccountPoolService struct {
	poolRepo *repository.AccountPoolRepository
	credRepo *repository.AccountCredentialRepository
}

func NewAccountPoolService(
	poolRepo *repository.AccountPoolRepository,
	credRepo *repository.AccountCredentialRepository,
) *AccountPoolService {
	return &AccountPoolService{
		poolRepo: poolRepo,
		credRepo: credRepo,
	}
}

// GetPools 获取所有账号池
func (s *AccountPoolService) GetPools(ctx context.Context, provider string) ([]*models.AccountPool, error) {
	if provider != "" {
		return s.poolRepo.FindByProvider(ctx, provider)
	}
	return s.poolRepo.FindAll(ctx)
}

// GetPool 获取指定账号池
func (s *AccountPoolService) GetPool(ctx context.Context, id uint) (*models.AccountPool, error) {
	return s.poolRepo.FindByID(ctx, id)
}

// CreatePool 创建账号池
func (s *AccountPoolService) CreatePool(ctx context.Context, pool *models.AccountPool) (*models.AccountPool, error) {
	if pool.Strategy == "" {
		pool.Strategy = "round_robin"
	}
	// IsActive 默认值由数据库处理，不需要在这里设置
	return s.poolRepo.Create(ctx, pool)
}

// UpdatePool 更新账号池
func (s *AccountPoolService) UpdatePool(ctx context.Context, pool *models.AccountPool) (*models.AccountPool, error) {
	existing, err := s.poolRepo.FindByID(ctx, pool.ID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("pool not found")
	}
	return s.poolRepo.Update(ctx, pool)
}

// DeletePool 删除账号池
func (s *AccountPoolService) DeletePool(ctx context.Context, id uint) error {
	return s.poolRepo.Delete(ctx, id)
}

// UpdatePoolStatus 更新账号池状态
func (s *AccountPoolService) UpdatePoolStatus(ctx context.Context, id uint, isActive bool) (*models.AccountPool, error) {
	pool, err := s.poolRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if pool == nil {
		return nil, fmt.Errorf("pool not found")
	}
	pool.IsActive = isActive
	return s.poolRepo.Update(ctx, pool)
}

// GetPoolStats 获取账号池统计信息
func (s *AccountPoolService) GetPoolStats(ctx context.Context, id uint) (map[string]interface{}, error) {
	pool, err := s.poolRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if pool == nil {
		return nil, fmt.Errorf("pool not found")
	}

	credentials, err := s.credRepo.FindByPoolID(ctx, id)
	if err != nil {
		return nil, err
	}

	activeCount := 0
	totalRequests := int64(0)
	for _, cred := range credentials {
		if cred.IsActive {
			activeCount++
		}
		totalRequests += cred.RequestCount
	}

	return map[string]interface{}{
		"pool_id":         pool.ID,
		"pool_name":       pool.Name,
		"provider":        pool.Provider,
		"total_creds":     len(credentials),
		"active_creds":    activeCount,
		"total_requests":  totalRequests,
	}, nil
}
