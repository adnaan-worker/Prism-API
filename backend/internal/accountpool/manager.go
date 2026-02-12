package accountpool

import (
	"api-aggregator/backend/internal/adapter"
	"api-aggregator/backend/internal/models"
	"api-aggregator/backend/internal/repository"
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrNoHealthyCredential = errors.New("no healthy credential available")
	ErrProviderNotFound    = errors.New("provider not found")
)

// Manager 账号池管理器，负责凭据选择和负载均衡
type Manager struct {
	credRepo     *repository.AccountCredentialRepository
	poolCounters map[uint]*atomic.Uint64 // 轮询计数器
	modelMapper  adapter.KiroModelMapper  // 模型映射器
	mu           sync.RWMutex
}

// NewManager 创建账号池管理器
func NewManager(credRepo *repository.AccountCredentialRepository, modelMapper adapter.KiroModelMapper) *Manager {
	m := &Manager{
		credRepo:     credRepo,
		poolCounters: make(map[uint]*atomic.Uint64),
		modelMapper:  modelMapper,
	}
	
	// 注册 Kiro 提供商
	Register(NewKiroProvider(m))
	
	return m
}

// GetModelMapper 获取模型映射器
func (m *Manager) GetModelMapper() adapter.KiroModelMapper {
	return m.modelMapper
}

// GetAdapter 为指定池获取适配器实例
// 返回适配器实例和使用的凭据 ID
func (m *Manager) GetAdapter(ctx context.Context, poolID uint) (interface{}, uint, error) {
	// 1. 获取池的所有凭据
	credentials, err := m.credRepo.FindByPoolID(ctx, poolID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get credentials: %w", err)
	}

	// 2. 过滤健康凭据
	healthy := m.filterHealthy(credentials)
	if len(healthy) == 0 {
		return nil, 0, ErrNoHealthyCredential
	}

	// 3. 轮询选择凭据
	cred := m.selectRoundRobin(poolID, healthy)

	// 4. 获取提供商
	provider, err := Get(cred.Provider)
	if err != nil {
		return nil, 0, err
	}

	// 5. 检查是否需要刷新令牌
	if cred.IsExpiringSoon(5 * time.Minute) {
		go m.refreshAsync(provider, cred)
	}

	// 6. 创建适配器
	adapterInstance, err := provider.CreateAdapter(cred)
	if err != nil {
		m.markUnhealthy(ctx, cred, err)
		return nil, 0, fmt.Errorf("failed to create adapter: %w", err)
	}

	// 7. 更新使用统计
	go m.updateStats(ctx, cred)

	return adapterInstance, cred.ID, nil
}

// filterHealthy 过滤健康的凭据
func (m *Manager) filterHealthy(creds []*models.AccountCredential) []*models.AccountCredential {
	healthy := make([]*models.AccountCredential, 0, len(creds))
	for _, cred := range creds {
		if cred.IsActive && !cred.IsExpired() {
			healthy = append(healthy, cred)
		}
	}
	return healthy
}

// selectRoundRobin 轮询选择凭据
func (m *Manager) selectRoundRobin(poolID uint, creds []*models.AccountCredential) *models.AccountCredential {
	m.mu.Lock()
	if _, exists := m.poolCounters[poolID]; !exists {
		m.poolCounters[poolID] = &atomic.Uint64{}
	}
	counter := m.poolCounters[poolID]
	m.mu.Unlock()

	idx := counter.Add(1) - 1
	return creds[int(idx%uint64(len(creds)))]
}

// refreshAsync 异步刷新令牌
func (m *Manager) refreshAsync(provider Provider, cred *models.AccountCredential) {
	ctx := context.Background()
	if err := provider.RefreshToken(ctx, cred); err != nil {
		return
	}
	m.credRepo.Update(ctx, cred)
}

// markUnhealthy 标记凭据为不健康
func (m *Manager) markUnhealthy(ctx context.Context, cred *models.AccountCredential, err error) {
	cred.IsActive = false
	cred.ErrorMessage = err.Error()
	go m.credRepo.Update(ctx, cred)
}

// updateStats 更新使用统计（请求前调用）
func (m *Manager) updateStats(ctx context.Context, cred *models.AccountCredential) {
	cred.RequestCount++
	now := time.Now()
	cred.LastUsedAt = &now
	m.credRepo.Update(ctx, cred)
}

// RecordSuccess 记录请求成功（请求后调用）
func (m *Manager) RecordSuccess(ctx context.Context, credID uint) {
	cred, err := m.credRepo.FindByID(ctx, credID)
	if err != nil {
		return
	}
	cred.SuccessCount++
	m.credRepo.Update(ctx, cred)
}

// RecordError 记录请求失败（请求后调用）
func (m *Manager) RecordError(ctx context.Context, credID uint, errMsg string) {
	cred, err := m.credRepo.FindByID(ctx, credID)
	if err != nil {
		return
	}
	cred.ErrorCount++
	cred.ErrorMessage = errMsg
	m.credRepo.Update(ctx, cred)
}
