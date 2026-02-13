package runtime

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"
)

// Config 运行时配置
type Config struct {
	mu sync.RWMutex
	db *gorm.DB

	// 缓存配置
	CacheEnabled      bool
	CacheTTL          time.Duration
	SemanticEnabled   bool
	SemanticThreshold float64

	// Embedding 配置
	EmbeddingEnabled bool
	EmbeddingURL     string
	EmbeddingTimeout time.Duration

	// 运行时配置
	MaxRetries        int
	Timeout           time.Duration
	EnableLoadBalance bool

	// 默认配额
	DefaultQuotaDaily   int64
	DefaultQuotaMonthly int64
	DefaultQuotaTotal   int64

	// 默认速率限制
	DefaultRateLimitPerMinute int
	DefaultRateLimitPerHour   int
	DefaultRateLimitPerDay    int
}

// Manager 配置管理器
type Manager struct {
	config *Config
	db     *gorm.DB
	ctx    context.Context
	cancel context.CancelFunc
}

// NewManager 创建配置管理器
func NewManager(db *gorm.DB) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		config: &Config{db: db},
		db:     db,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Load 加载配置
func (m *Manager) Load() error {
	return m.loadFromDatabase()
}

// Reload 重新加载配置
func (m *Manager) Reload() error {
	return m.loadFromDatabase()
}

// Get 获取配置
func (m *Manager) Get() *Config {
	return m.config
}

// StartAutoReload 启动自动重载（每分钟检查一次）
func (m *Manager) StartAutoReload(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				m.Reload()
			case <-m.ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

// Stop 停止自动重载
func (m *Manager) Stop() {
	m.cancel()
}

// loadFromDatabase 从数据库加载配置
func (m *Manager) loadFromDatabase() error {
	m.config.mu.Lock()
	defer m.config.mu.Unlock()

	settings := make(map[string]string)
	
	var results []struct {
		Key   string
		Value string
	}
	
	err := m.db.Table("settings").
		Select("key, value").
		Scan(&results).Error
	
	if err != nil {
		return err
	}
	
	for _, r := range results {
		settings[r.Key] = r.Value
	}

	// 解析配置
	m.config.CacheEnabled = getBool(settings, "runtime.cache_enabled", true)
	m.config.CacheTTL = time.Duration(getDuration(settings, "runtime.cache_ttl", 3600)) * time.Second
	m.config.SemanticEnabled = getBool(settings, "runtime.semantic_cache_enabled", false)
	m.config.SemanticThreshold = getFloat(settings, "runtime.semantic_threshold", 0.85)
	
	m.config.EmbeddingEnabled = getBool(settings, "runtime.embedding_enabled", false)
	m.config.EmbeddingURL = getString(settings, "runtime.embedding_url", "http://localhost:8765")
	m.config.EmbeddingTimeout = time.Duration(getDuration(settings, "runtime.embedding_timeout", 30)) * time.Second
	
	m.config.MaxRetries = getInt(settings, "runtime.max_retries", 3)
	m.config.Timeout = time.Duration(getDuration(settings, "runtime.timeout", 30)) * time.Second
	m.config.EnableLoadBalance = getBool(settings, "runtime.enable_load_balance", true)
	
	m.config.DefaultQuotaDaily = getInt64(settings, "default_quota.daily", 1000)
	m.config.DefaultQuotaMonthly = getInt64(settings, "default_quota.monthly", 30000)
	m.config.DefaultQuotaTotal = getInt64(settings, "default_quota.total", 100000)
	
	m.config.DefaultRateLimitPerMinute = getInt(settings, "default_rate_limit.per_minute", 60)
	m.config.DefaultRateLimitPerHour = getInt(settings, "default_rate_limit.per_hour", 1000)
	m.config.DefaultRateLimitPerDay = getInt(settings, "default_rate_limit.per_day", 10000)

	return nil
}

// Config 方法

// IsCacheEnabled 缓存是否启用
func (c *Config) IsCacheEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.CacheEnabled
}

// GetCacheTTL 获取缓存 TTL
func (c *Config) GetCacheTTL() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.CacheTTL
}

// IsSemanticEnabled 语义缓存是否启用
func (c *Config) IsSemanticEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.SemanticEnabled && c.CacheEnabled
}

// GetSemanticThreshold 获取语义匹配阈值
func (c *Config) GetSemanticThreshold() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.SemanticThreshold
}

// IsEmbeddingEnabled Embedding 服务是否启用
func (c *Config) IsEmbeddingEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.EmbeddingEnabled
}

// GetEmbeddingURL 获取 Embedding 服务 URL
func (c *Config) GetEmbeddingURL() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.EmbeddingURL
}

// GetEmbeddingTimeout 获取 Embedding 超时时间
func (c *Config) GetEmbeddingTimeout() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.EmbeddingTimeout
}

// GetMaxRetries 获取最大重试次数
func (c *Config) GetMaxRetries() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.MaxRetries
}

// GetTimeout 获取超时时间
func (c *Config) GetTimeout() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Timeout
}

// IsLoadBalanceEnabled 负载均衡是否启用
func (c *Config) IsLoadBalanceEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.EnableLoadBalance
}

// GetDefaultQuota 获取默认配额
func (c *Config) GetDefaultQuota() (daily, monthly, total int64) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.DefaultQuotaDaily, c.DefaultQuotaMonthly, c.DefaultQuotaTotal
}

// GetDefaultRateLimit 获取默认速率限制
func (c *Config) GetDefaultRateLimit() (perMinute, perHour, perDay int) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.DefaultRateLimitPerMinute, c.DefaultRateLimitPerHour, c.DefaultRateLimitPerDay
}

// 辅助函数
func getString(settings map[string]string, key, defaultValue string) string {
	if val, ok := settings[key]; ok {
		return val
	}
	return defaultValue
}

func getInt(settings map[string]string, key string, defaultValue int) int {
	if val, ok := settings[key]; ok {
		var result int
		if _, err := fmt.Sscanf(val, "%d", &result); err == nil {
			return result
		}
	}
	return defaultValue
}

func getInt64(settings map[string]string, key string, defaultValue int64) int64 {
	if val, ok := settings[key]; ok {
		var result int64
		if _, err := fmt.Sscanf(val, "%d", &result); err == nil {
			return result
		}
	}
	return defaultValue
}

func getFloat(settings map[string]string, key string, defaultValue float64) float64 {
	if val, ok := settings[key]; ok {
		var result float64
		if _, err := fmt.Sscanf(val, "%f", &result); err == nil {
			return result
		}
	}
	return defaultValue
}

func getBool(settings map[string]string, key string, defaultValue bool) bool {
	if val, ok := settings[key]; ok {
		return val == "true" || val == "1" || val == "yes"
	}
	return defaultValue
}

func getDuration(settings map[string]string, key string, defaultValue int) int {
	if val, ok := settings[key]; ok {
		var result int
		if _, err := fmt.Sscanf(val, "%d", &result); err == nil {
			return result
		}
	}
	return defaultValue
}
