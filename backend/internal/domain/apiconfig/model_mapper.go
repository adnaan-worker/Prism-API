package apiconfig

import (
	"context"
	"sync"
	"time"
)

// ModelMapper 模型映射器
// 用于 Kiro 等需要特殊模型 ID 的场景
type ModelMapper struct {
	repo  Repository
	cache map[string]string
	mu    sync.RWMutex
	ttl   time.Duration
}

// NewModelMapper 创建模型映射器
func NewModelMapper(repo Repository) *ModelMapper {
	return &ModelMapper{
		repo:  repo,
		cache: make(map[string]string),
		ttl:   5 * time.Minute,
	}
}

// GetModelMapping 获取模型映射
// 从 APIConfig 的 Metadata 中读取 model_mappings
// 格式: {"model_mappings": {"claude-3-5-sonnet": "anthropic.claude-3-5-sonnet-20241022-v2:0"}}
func (m *ModelMapper) GetModelMapping(ctx context.Context, modelName string) (string, error) {
	// 先查缓存
	m.mu.RLock()
	if mapped, ok := m.cache[modelName]; ok {
		m.mu.RUnlock()
		return mapped, nil
	}
	m.mu.RUnlock()

	// 查询所有 Kiro 类型的配置
	configs, err := m.repo.FindByType(ctx, "kiro")
	if err != nil || len(configs) == 0 {
		// 如果没有配置或出错，返回原模型名
		return modelName, nil
	}

	// 遍历配置，查找映射
	for _, config := range configs {
		if config.Metadata != nil {
			if mappings, ok := config.Metadata["model_mappings"].(map[string]interface{}); ok {
				if mapped, ok := mappings[modelName].(string); ok {
					// 找到映射，缓存并返回
					m.mu.Lock()
					m.cache[modelName] = mapped
					m.mu.Unlock()
					return mapped, nil
				}
			}
		}
	}

	// 没有找到映射，返回原模型名
	return modelName, nil
}

// ClearCache 清除缓存
func (m *ModelMapper) ClearCache() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cache = make(map[string]string)
}
