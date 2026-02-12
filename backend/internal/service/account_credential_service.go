package service

import (
	"api-aggregator/backend/internal/accountpool"
	"api-aggregator/backend/internal/models"
	"api-aggregator/backend/internal/repository"
	"context"
	"fmt"
)

type AccountCredentialService struct {
	credRepo *repository.AccountCredentialRepository
	poolRepo *repository.AccountPoolRepository
}

func NewAccountCredentialService(
	credRepo *repository.AccountCredentialRepository,
	poolRepo *repository.AccountPoolRepository,
) *AccountCredentialService {
	return &AccountCredentialService{
		credRepo: credRepo,
		poolRepo: poolRepo,
	}
}

// GetCredentials 获取凭据列表
func (s *AccountCredentialService) GetCredentials(ctx context.Context, poolID *uint, provider, status string) ([]*models.AccountCredential, error) {
	if poolID != nil {
		return s.credRepo.FindByPoolID(ctx, *poolID)
	}
	if provider != "" {
		return s.credRepo.FindByProvider(ctx, provider)
	}
	if status != "" {
		return s.credRepo.FindByStatus(ctx, status)
	}
	return s.credRepo.FindAll(ctx)
}

// GetCredential 获取指定凭据
func (s *AccountCredentialService) GetCredential(ctx context.Context, id uint) (*models.AccountCredential, error) {
	return s.credRepo.FindByID(ctx, id)
}

// CreateCredential 创建凭据
func (s *AccountCredentialService) CreateCredential(ctx context.Context, cred *models.AccountCredential) (*models.AccountCredential, error) {
	// 验证池存在
	pool, err := s.poolRepo.FindByID(ctx, cred.PoolID)
	if err != nil {
		return nil, err
	}
	if pool == nil {
		return nil, fmt.Errorf("pool not found")
	}

	// 验证提供商匹配
	if cred.Provider != pool.Provider {
		return nil, fmt.Errorf("provider mismatch")
	}

	// 如果是 refresh_token 类型，立即刷新以验证有效性
	if cred.AuthType == "refresh_token" {
		provider, err := accountpool.Get(cred.Provider)
		if err != nil {
			return nil, fmt.Errorf("provider not found: %w", err)
		}
		
		// 尝试刷新令牌
		if err := provider.RefreshToken(ctx, cred); err != nil {
			return nil, fmt.Errorf("invalid refresh_token: %w", err)
		}
	}

	// 创建凭据
	createdCred, err := s.credRepo.Create(ctx, cred)
	if err != nil {
		return nil, err
	}

	// 添加到账号池关联表
	if err := s.poolRepo.AddCredentialToPool(ctx, cred.PoolID, createdCred.ID); err != nil {
		// 如果关联失败，删除已创建的凭据
		s.credRepo.Delete(ctx, createdCred.ID)
		return nil, fmt.Errorf("failed to add credential to pool: %w", err)
	}

	return createdCred, nil
}

// UpdateCredential 更新凭据
func (s *AccountCredentialService) UpdateCredential(ctx context.Context, cred *models.AccountCredential) (*models.AccountCredential, error) {
	existing, err := s.credRepo.FindByID(ctx, cred.ID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("credential not found")
	}
	return s.credRepo.Update(ctx, cred)
}

// DeleteCredential 删除凭据
func (s *AccountCredentialService) DeleteCredential(ctx context.Context, id uint) error {
	return s.credRepo.Delete(ctx, id)
}

// RefreshCredential 刷新凭据令牌
func (s *AccountCredentialService) RefreshCredential(ctx context.Context, id uint) (*models.AccountCredential, error) {
	cred, err := s.credRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if cred == nil {
		return nil, fmt.Errorf("credential not found")
	}

	// 获取提供商
	provider, err := accountpool.Get(cred.Provider)
	if err != nil {
		return nil, err
	}

	// 刷新令牌
	if err := provider.RefreshToken(ctx, cred); err != nil {
		cred.IsActive = false
		cred.ErrorMessage = err.Error()
		s.credRepo.Update(ctx, cred)
		return nil, err
	}

	return s.credRepo.Update(ctx, cred)
}

// UpdateCredentialStatus 更新凭据状态
func (s *AccountCredentialService) UpdateCredentialStatus(ctx context.Context, id uint, isActive bool) (*models.AccountCredential, error) {
	cred, err := s.credRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if cred == nil {
		return nil, fmt.Errorf("credential not found")
	}
	cred.IsActive = isActive
	return s.credRepo.Update(ctx, cred)
}
