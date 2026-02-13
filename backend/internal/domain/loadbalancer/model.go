package loadbalancer

import (
	"time"
)

// LoadBalancerConfig 璐熻浇鍧囪　閰嶇疆妯″瀷
type LoadBalancerConfig struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	ModelName string         `gorm:"not null;size:255;index" json:"model_name"`
	Strategy  string         `gorm:"not null;default:'round_robin';size:50" json:"strategy"`
	IsActive  bool           `gorm:"not null;default:true" json:"is_active"`
}

// TableName 鎸囧畾琛ㄥ悕
func (LoadBalancerConfig) TableName() string {
	return "load_balancer_configs"
}

// IsRoundRobin 妫€鏌ユ槸鍚︿负杞绛栫暐
func (c *LoadBalancerConfig) IsRoundRobin() bool {
	return c.Strategy == StrategyRoundRobin
}

// IsWeightedRoundRobin 妫€鏌ユ槸鍚︿负鍔犳潈杞绛栫暐
func (c *LoadBalancerConfig) IsWeightedRoundRobin() bool {
	return c.Strategy == StrategyWeightedRoundRobin
}

// IsLeastConnections 妫€鏌ユ槸鍚︿负鏈€灏戣繛鎺ョ瓥鐣?
func (c *LoadBalancerConfig) IsLeastConnections() bool {
	return c.Strategy == StrategyLeastConnections
}

// IsRandom 妫€鏌ユ槸鍚︿负闅忔満绛栫暐
func (c *LoadBalancerConfig) IsRandom() bool {
	return c.Strategy == StrategyRandom
}

// Activate 婵€娲婚厤缃?
func (c *LoadBalancerConfig) Activate() {
	c.IsActive = true
}

// Deactivate 鍋滅敤閰嶇疆
func (c *LoadBalancerConfig) Deactivate() {
	c.IsActive = false
}

// 璐熻浇鍧囪　绛栫暐甯搁噺
const (
	StrategyRoundRobin          = "round_robin"
	StrategyWeightedRoundRobin  = "weighted_round_robin"
	StrategyLeastConnections    = "least_connections"
	StrategyRandom              = "random"
)

// ValidStrategies 鏈夋晥鐨勮礋杞藉潎琛＄瓥鐣ュ垪琛?
var ValidStrategies = []string{
	StrategyRoundRobin,
	StrategyWeightedRoundRobin,
	StrategyLeastConnections,
	StrategyRandom,
}

// IsValidStrategy 妫€鏌ョ瓥鐣ユ槸鍚︽湁鏁?
func IsValidStrategy(strategy string) bool {
	for _, s := range ValidStrategies {
		if s == strategy {
			return true
		}
	}
	return false
}
