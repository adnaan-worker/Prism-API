package accountpool

import (
	"fmt"
	"sync"
)

// Registry 提供商注册表，管理所有已注册的提供商
type Registry struct {
	providers map[string]Provider
	mu        sync.RWMutex
}

// NewRegistry 创建新的提供商注册表
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
	}
}

// Register 注册提供商
func (r *Registry) Register(provider Provider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[provider.Name()] = provider
}

// Get 获取提供商
func (r *Registry) Get(name string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	provider, ok := r.providers[name]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", name)
	}
	return provider, nil
}

// List 列出所有已注册的提供商名称
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}

// 全局注册表实例
var globalRegistry = NewRegistry()

// Register 向全局注册表注册提供商
func Register(provider Provider) {
	globalRegistry.Register(provider)
}

// Get 从全局注册表获取提供商
func Get(name string) (Provider, error) {
	return globalRegistry.Get(name)
}

// List 列出全局注册表中的所有提供商
func List() []string {
	return globalRegistry.List()
}
