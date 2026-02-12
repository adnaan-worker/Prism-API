package loadbalancer

import (
	"time"

	"gorm.io/gorm"
)

// LoadBalancerConfig 负载均衡配置模型
type LoadBalancerConfig struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	ModelName string         `gorm:"not null;size:255;index" json:"model_name"`
	Strategy  string         `gorm:"not null;default:'round_robin';size:50" json:"strategy"`
	IsActive  bool           `gorm:"not null;default:true" json:"is_active"`
}

// TableName 指定表名
func (LoadBalancerConfig) TableName() string {
	return "load_balancer_configs"
}

// IsRoundRobin 检查是否为轮询策略
func (c *LoadBalancerConfig) IsRoundRobin() bool {
	return c.Strategy == StrategyRoundRobin
}

// IsWeightedRoundRobin 检查是否为加权轮询策略
func (c *LoadBalancerConfig) IsWeightedRoundRobin() bool {
	return c.Strategy == StrategyWeightedRoundRobin
}

// IsLeastConnections 检查是否为最少连接策略
func (c *LoadBalancerConfig) IsLeastConnections() bool {
	return c.Strategy == StrategyLeastConnections
}

// IsRandom 检查是否为随机策略
func (c *LoadBalancerConfig) IsRandom() bool {
	return c.Strategy == StrategyRandom
}

// Activate 激活配置
func (c *LoadBalancerConfig) Activate() {
	c.IsActive = true
}

// Deactivate 停用配置
func (c *LoadBalancerConfig) Deactivate() {
	c.IsActive = false
}

// 负载均衡策略常量
const (
	StrategyRoundRobin          = "round_robin"
	StrategyWeightedRoundRobin  = "weighted_round_robin"
	StrategyLeastConnections    = "least_connections"
	StrategyRandom              = "random"
)

// ValidStrategies 有效的负载均衡策略列表
var ValidStrategies = []string{
	StrategyRoundRobin,
	StrategyWeightedRoundRobin,
	StrategyLeastConnections,
	StrategyRandom,
}

// IsValidStrategy 检查策略是否有效
func IsValidStrategy(strategy string) bool {
	for _, s := range ValidStrategies {
		if s == strategy {
			return true
		}
	}
	return false
}
