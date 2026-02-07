package loadbalancer

import (
	"api-aggregator/backend/internal/models"
	"context"
	"errors"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrNoAvailableConfig = errors.New("no available configuration")
)

// LoadBalancer is the interface for load balancing strategies
type LoadBalancer interface {
	// SelectConfig selects an API configuration based on the load balancing strategy
	SelectConfig(ctx context.Context, configs []*models.APIConfig) (*models.APIConfig, error)

	// GetStrategy returns the strategy name
	GetStrategy() string
}

// RoundRobinLB implements round-robin load balancing
type RoundRobinLB struct {
	counter *atomic.Uint64
}

// NewRoundRobinLB creates a new round-robin load balancer
func NewRoundRobinLB() *RoundRobinLB {
	return &RoundRobinLB{
		counter: &atomic.Uint64{},
	}
}

// GetStrategy returns the strategy name
func (lb *RoundRobinLB) GetStrategy() string {
	return "round_robin"
}

// SelectConfig selects a configuration using round-robin
func (lb *RoundRobinLB) SelectConfig(ctx context.Context, configs []*models.APIConfig) (*models.APIConfig, error) {
	if len(configs) == 0 {
		return nil, ErrNoAvailableConfig
	}

	// Get next index
	index := lb.counter.Add(1) - 1
	selectedIndex := int(index % uint64(len(configs)))

	return configs[selectedIndex], nil
}

// WeightedRoundRobinLB implements weighted round-robin load balancing
type WeightedRoundRobinLB struct {
	counter *atomic.Uint64
	mu      sync.RWMutex
}

// NewWeightedRoundRobinLB creates a new weighted round-robin load balancer
func NewWeightedRoundRobinLB() *WeightedRoundRobinLB {
	return &WeightedRoundRobinLB{
		counter: &atomic.Uint64{},
	}
}

// GetStrategy returns the strategy name
func (lb *WeightedRoundRobinLB) GetStrategy() string {
	return "weighted_round_robin"
}

// SelectConfig selects a configuration using weighted round-robin
func (lb *WeightedRoundRobinLB) SelectConfig(ctx context.Context, configs []*models.APIConfig) (*models.APIConfig, error) {
	if len(configs) == 0 {
		return nil, ErrNoAvailableConfig
	}

	// Calculate total weight
	totalWeight := 0
	for _, config := range configs {
		if config.Weight <= 0 {
			config.Weight = 1 // Default weight
		}
		totalWeight += config.Weight
	}

	// Get next position
	position := int(lb.counter.Add(1)-1) % totalWeight

	// Find the config based on weight
	currentWeight := 0
	for _, config := range configs {
		currentWeight += config.Weight
		if position < currentWeight {
			return config, nil
		}
	}

	// Fallback to first config
	return configs[0], nil
}

// LeastConnectionsLB implements least connections load balancing
type LeastConnectionsLB struct {
	connections map[uint]*atomic.Int32
	mu          sync.RWMutex
}

// NewLeastConnectionsLB creates a new least connections load balancer
func NewLeastConnectionsLB() *LeastConnectionsLB {
	return &LeastConnectionsLB{
		connections: make(map[uint]*atomic.Int32),
	}
}

// GetStrategy returns the strategy name
func (lb *LeastConnectionsLB) GetStrategy() string {
	return "least_connections"
}

// SelectConfig selects a configuration with the least connections
func (lb *LeastConnectionsLB) SelectConfig(ctx context.Context, configs []*models.APIConfig) (*models.APIConfig, error) {
	if len(configs) == 0 {
		return nil, ErrNoAvailableConfig
	}

	lb.mu.Lock()
	defer lb.mu.Unlock()

	// Initialize connection counters if needed
	for _, config := range configs {
		if _, exists := lb.connections[config.ID]; !exists {
			lb.connections[config.ID] = &atomic.Int32{}
		}
	}

	// Find config with least connections
	var selectedConfig *models.APIConfig
	minConnections := int32(1<<31 - 1) // Max int32

	for _, config := range configs {
		connections := lb.connections[config.ID].Load()
		if connections < minConnections {
			minConnections = connections
			selectedConfig = config
		}
	}

	if selectedConfig == nil {
		return configs[0], nil
	}

	// Increment connection count
	lb.connections[selectedConfig.ID].Add(1)

	return selectedConfig, nil
}

// ReleaseConnection decrements the connection count for a config
func (lb *LeastConnectionsLB) ReleaseConnection(configID uint) {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	if counter, exists := lb.connections[configID]; exists {
		counter.Add(-1)
	}
}

// RandomLB implements random load balancing
type RandomLB struct {
	rand *rand.Rand
	mu   sync.Mutex
}

// NewRandomLB creates a new random load balancer
func NewRandomLB() *RandomLB {
	return &RandomLB{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GetStrategy returns the strategy name
func (lb *RandomLB) GetStrategy() string {
	return "random"
}

// SelectConfig selects a random configuration
func (lb *RandomLB) SelectConfig(ctx context.Context, configs []*models.APIConfig) (*models.APIConfig, error) {
	if len(configs) == 0 {
		return nil, ErrNoAvailableConfig
	}

	lb.mu.Lock()
	defer lb.mu.Unlock()

	index := lb.rand.Intn(len(configs))
	return configs[index], nil
}

// Factory creates load balancers based on strategy
type Factory struct{}

// NewFactory creates a new load balancer factory
func NewFactory() *Factory {
	return &Factory{}
}

// CreateLoadBalancer creates a load balancer based on strategy
func (f *Factory) CreateLoadBalancer(strategy string) LoadBalancer {
	switch strategy {
	case "round_robin":
		return NewRoundRobinLB()
	case "weighted_round_robin":
		return NewWeightedRoundRobinLB()
	case "least_connections":
		return NewLeastConnectionsLB()
	case "random":
		return NewRandomLB()
	default:
		// Default to round robin
		return NewRoundRobinLB()
	}
}
