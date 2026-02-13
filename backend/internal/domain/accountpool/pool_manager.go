package accountpool

import (
	"api-aggregator/backend/internal/adapter"
	"api-aggregator/backend/pkg/errors"
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// PoolManager 账号池管理器
// 负责从账号池中选择凭据并创建适配器
type PoolManager struct {
	repo           Repository
	modelMapper    adapter.KiroModelMapper
	refreshService *KiroRefreshService
	mu             sync.RWMutex
	roundRobinIdx  map[uint]int // 轮询索引，key为poolID
}

// NewPoolManager 创建账号池管理器
func NewPoolManager(repo Repository, modelMapper adapter.KiroModelMapper) *PoolManager {
	return &PoolManager{
		repo:           repo,
		modelMapper:    modelMapper,
		refreshService: NewKiroRefreshService(),
		roundRobinIdx:  make(map[uint]int),
	}
}

// GetAdapter 从账号池获取适配器
// 返回：适配器实例、凭据ID、错误
func (pm *PoolManager) GetAdapter(ctx context.Context, poolID uint) (interface{}, uint, error) {
	// 获取账号池
	pool, err := pm.repo.FindByID(ctx, poolID)
	if err != nil {
		return nil, 0, errors.NewNotFoundError("account pool not found")
	}

	if !pool.IsActive {
		return nil, 0, errors.New(500001, "account pool is not active")
	}

	// 获取账号池的所有活跃凭据
	creds, err := pm.repo.FindActiveCredentialsByPoolID(ctx, poolID)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to get credentials")
	}

	if len(creds) == 0 {
		return nil, 0, errors.New(500001, "no active credentials in pool")
	}

	// 根据策略选择凭据
	cred, err := pm.selectCredential(pool, creds)
	if err != nil {
		return nil, 0, err
	}

	// 如果是 Kiro 凭据且已过期，尝试刷新
	if cred.Provider == "kiro" && cred.IsExpired() {
		if err := pm.refreshService.RefreshKiroToken(ctx, cred); err != nil {
			// 刷新失败，标记为不健康
			cred.UpdateHealthStatus(HealthStatusUnhealthy)
			cred.LastError = fmt.Sprintf("failed to refresh token: %v", err)
			pm.repo.UpdateCredential(ctx, cred)
			return nil, 0, errors.Wrap(err, 500001, "failed to refresh kiro token")
		}
		// 刷新成功，保存更新
		if err := pm.repo.UpdateCredential(ctx, cred); err != nil {
			return nil, 0, errors.Wrap(err, "failed to save refreshed credential")
		}
	}

	// 检查凭据是否健康
	if !cred.IsHealthy() {
		return nil, 0, errors.New(500001, "selected credential is not healthy")
	}

	// 检查速率限制
	if cred.IsRateLimited() {
		return nil, 0, errors.New(429002, "credential rate limit exceeded")
	}

	// 创建适配器
	adapterInstance, err := pm.createAdapterFromCredential(cred)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to create adapter from credential")
	}

	// 增加使用量
	cred.IncrementUsage()
	pm.repo.UpdateCredential(ctx, cred)

	return adapterInstance, cred.ID, nil
}

// selectCredential 根据策略选择凭据
func (pm *PoolManager) selectCredential(pool *AccountPool, creds []*AccountCredential) (*AccountCredential, error) {
	switch pool.Strategy {
	case StrategyRoundRobin:
		return pm.selectRoundRobin(pool.ID, creds), nil
	case StrategyWeightedRoundRobin:
		return pm.selectWeightedRoundRobin(creds), nil
	case StrategyLeastConnections:
		return pm.selectLeastConnections(creds), nil
	case StrategyRandom:
		return pm.selectRandom(creds), nil
	default:
		return pm.selectRoundRobin(pool.ID, creds), nil
	}
}

// selectRoundRobin 轮询选择
func (pm *PoolManager) selectRoundRobin(poolID uint, creds []*AccountCredential) *AccountCredential {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	idx := pm.roundRobinIdx[poolID]
	cred := creds[idx%len(creds)]
	pm.roundRobinIdx[poolID] = (idx + 1) % len(creds)

	return cred
}

// selectWeightedRoundRobin 加权轮询选择
func (pm *PoolManager) selectWeightedRoundRobin(creds []*AccountCredential) *AccountCredential {
	totalWeight := 0
	for _, cred := range creds {
		totalWeight += cred.Weight
	}

	if totalWeight == 0 {
		return creds[0]
	}

	random := rand.Intn(totalWeight)
	for _, cred := range creds {
		random -= cred.Weight
		if random < 0 {
			return cred
		}
	}

	return creds[0]
}

// selectLeastConnections 最少连接选择
func (pm *PoolManager) selectLeastConnections(creds []*AccountCredential) *AccountCredential {
	var selected *AccountCredential
	minRequests := int64(-1)

	for _, cred := range creds {
		if minRequests == -1 || cred.TotalRequests < minRequests {
			minRequests = cred.TotalRequests
			selected = cred
		}
	}

	return selected
}

// selectRandom 随机选择
func (pm *PoolManager) selectRandom(creds []*AccountCredential) *AccountCredential {
	return creds[rand.Intn(len(creds))]
}

// createAdapterFromCredential 从凭据创建适配器
func (pm *PoolManager) createAdapterFromCredential(cred *AccountCredential) (adapter.Adapter, error) {
	// 根据提供商类型创建适配器
	switch cred.Provider {
	case "kiro":
		return pm.createKiroAdapter(cred)
	case "openai":
		return pm.createOpenAIAdapter(cred)
	case "anthropic":
		return pm.createAnthropicAdapter(cred)
	case "gemini":
		return pm.createGeminiAdapter(cred)
	default:
		return nil, errors.New(500001, fmt.Sprintf("unsupported provider: %s", cred.Provider))
	}
}

// createKiroAdapter 创建 Kiro 适配器
func (pm *PoolManager) createKiroAdapter(cred *AccountCredential) (adapter.Adapter, error) {
	// Kiro 需要特殊处理
	if cred.AuthType != AuthTypeOAuth {
		return nil, errors.New(500001, "kiro requires oauth authentication")
	}

	config := &adapter.Config{
		BaseURL: "https://q.us-east-1.amazonaws.com",
		Timeout: 120,
	}

	// 从 Metadata 中获取 machine_id
	machineID, _ := cred.Metadata["machine_id"].(string)
	if machineID == "" {
		// 如果没有，生成一个新的
		machineID = generateMachineID()
		if cred.Metadata == nil {
			cred.Metadata = make(JSONMap)
		}
		cred.Metadata["machine_id"] = machineID
	}

	// 创建 Kiro 适配器
	// profileArn 对于 Social Auth (OAuth) 不需要，传空字符串
	kiroAdapter := adapter.NewKiroAdapter(
		config,
		cred.AccessToken,
		"",            // profileArn - not needed for Social Auth
		"us-east-1",   // region
		pm.modelMapper, // model mapper
	)

	return kiroAdapter, nil
}

// generateMachineID 生成机器 ID
func generateMachineID() string {
	// 简单的随机 ID 生成
	return fmt.Sprintf("machine-%d", time.Now().UnixNano())
}

// createOpenAIAdapter 创建 OpenAI 适配器
func (pm *PoolManager) createOpenAIAdapter(cred *AccountCredential) (adapter.Adapter, error) {
	config := &adapter.Config{
		BaseURL: "https://api.openai.com",
		APIKey:  cred.APIKey,
		Timeout: 60,
	}

	return adapter.NewOpenAIAdapter(config), nil
}

// createAnthropicAdapter 创建 Anthropic 适配器
func (pm *PoolManager) createAnthropicAdapter(cred *AccountCredential) (adapter.Adapter, error) {
	config := &adapter.Config{
		BaseURL: "https://api.anthropic.com",
		APIKey:  cred.APIKey,
		Timeout: 60,
	}

	return adapter.NewAnthropicAdapter(config), nil
}

// createGeminiAdapter 创建 Gemini 适配器
func (pm *PoolManager) createGeminiAdapter(cred *AccountCredential) (adapter.Adapter, error) {
	config := &adapter.Config{
		BaseURL: "https://generativelanguage.googleapis.com",
		APIKey:  cred.APIKey,
		Timeout: 60,
	}

	return adapter.NewGeminiAdapter(config), nil
}

// RecordSuccess 记录成功请求
func (pm *PoolManager) RecordSuccess(ctx context.Context, credID uint) {
	cred, err := pm.repo.FindCredentialByID(ctx, credID)
	if err != nil {
		return
	}

	cred.IncrementRequests()
	cred.UpdateHealthStatus(HealthStatusHealthy)
	pm.repo.UpdateCredential(ctx, cred)
}

// RecordError 记录失败请求
func (pm *PoolManager) RecordError(ctx context.Context, credID uint, errMsg string) {
	cred, err := pm.repo.FindCredentialByID(ctx, credID)
	if err != nil {
		return
	}

	cred.IncrementRequests()
	cred.IncrementErrors()

	// 如果错误率过高，标记为不健康
	if cred.GetErrorRate() > 0.5 {
		cred.UpdateHealthStatus(HealthStatusUnhealthy)
	}

	pm.repo.UpdateCredential(ctx, cred)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
