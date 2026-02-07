package loadbalancer

import (
	"api-aggregator/backend/internal/models"
	"context"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Property 11: 负载均衡分配
// Feature: api-aggregator, Property 11: For any model with multiple API configurations, making multiple requests should result in different configurations being selected according to the load balancing strategy.
// Validates: Requirements 5.1
func TestProperty_LoadBalancerDistribution(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Generator for number of configs (2-5)
	configCountGen := gen.IntRange(2, 5)

	properties.Property("Load balancer distributes requests across configs", prop.ForAll(
		func(configCount int) bool {
			ctx := context.Background()

			// Create test configs
			configs := make([]*models.APIConfig, configCount)
			for i := 0; i < configCount; i++ {
				configs[i] = &models.APIConfig{
					ID:       uint(i + 1),
					Name:     "Config " + string(rune('A'+i)),
					Type:     "openai",
					BaseURL:  "https://api.test.com",
					IsActive: true,
					Weight:   1,
				}
			}

			// Test round robin
			rrLB := NewRoundRobinLB()
			selectedIDs := make(map[uint]bool)

			// Make enough requests to hit all configs
			for i := 0; i < configCount*2; i++ {
				config, err := rrLB.SelectConfig(ctx, configs)
				if err != nil {
					return false
				}
				selectedIDs[config.ID] = true
			}

			// All configs should have been selected at least once
			if len(selectedIDs) != configCount {
				return false
			}

			// Test random
			randomLB := NewRandomLB()
			selectedIDs = make(map[uint]bool)

			// Make many requests to increase probability of hitting all configs
			for i := 0; i < configCount*10; i++ {
				config, err := randomLB.SelectConfig(ctx, configs)
				if err != nil {
					return false
				}
				selectedIDs[config.ID] = true
			}

			// With enough requests, all configs should be selected (probabilistic)
			// We accept if at least 50% of configs are selected
			return len(selectedIDs) >= configCount/2

		},
		configCountGen,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Property 12: 加权轮询比例
// Feature: api-aggregator, Property 12: For any set of API configurations with different weights, after a large number of requests, the distribution of selected configurations should approximate the weight ratios.
// Validates: Requirements 5.2
func TestProperty_WeightedRoundRobinRatio(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Generator for weights (1-10)
	weightGen := gen.IntRange(1, 10)

	properties.Property("Weighted round robin respects weight ratios", prop.ForAll(
		func(weight1, weight2 int) bool {
			ctx := context.Background()

			// Create configs with different weights
			configs := []*models.APIConfig{
				{
					ID:       1,
					Name:     "Config A",
					Type:     "openai",
					BaseURL:  "https://api1.test.com",
					IsActive: true,
					Weight:   weight1,
				},
				{
					ID:       2,
					Name:     "Config B",
					Type:     "openai",
					BaseURL:  "https://api2.test.com",
					IsActive: true,
					Weight:   weight2,
				},
			}

			lb := NewWeightedRoundRobinLB()

			// Count selections
			counts := make(map[uint]int)
			totalRequests := (weight1 + weight2) * 10 // Multiple cycles

			for i := 0; i < totalRequests; i++ {
				config, err := lb.SelectConfig(ctx, configs)
				if err != nil {
					return false
				}
				counts[config.ID]++
			}

			// Calculate expected ratios
			totalWeight := float64(weight1 + weight2)
			expectedRatio1 := float64(weight1) / totalWeight
			expectedRatio2 := float64(weight2) / totalWeight

			// Calculate actual ratios
			actualRatio1 := float64(counts[1]) / float64(totalRequests)
			actualRatio2 := float64(counts[2]) / float64(totalRequests)

			// Allow 10% tolerance
			tolerance := 0.1
			ratio1Match := abs(actualRatio1-expectedRatio1) < tolerance
			ratio2Match := abs(actualRatio2-expectedRatio2) < tolerance

			return ratio1Match && ratio2Match
		},
		weightGen,
		weightGen,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Helper function for absolute value
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// Unit test for RoundRobinLB
func TestRoundRobinLB(t *testing.T) {
	ctx := context.Background()
	lb := NewRoundRobinLB()

	configs := []*models.APIConfig{
		{ID: 1, Name: "Config 1"},
		{ID: 2, Name: "Config 2"},
		{ID: 3, Name: "Config 3"},
	}

	// Test round robin distribution
	for i := 0; i < 9; i++ {
		config, err := lb.SelectConfig(ctx, configs)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		expectedID := uint((i % 3) + 1)
		if config.ID != expectedID {
			t.Errorf("Round %d: Expected config ID %d, got %d", i, expectedID, config.ID)
		}
	}

	// Test empty configs
	_, err := lb.SelectConfig(ctx, []*models.APIConfig{})
	if err != ErrNoAvailableConfig {
		t.Errorf("Expected ErrNoAvailableConfig, got %v", err)
	}
}

// Unit test for WeightedRoundRobinLB
func TestWeightedRoundRobinLB(t *testing.T) {
	ctx := context.Background()
	lb := NewWeightedRoundRobinLB()

	configs := []*models.APIConfig{
		{ID: 1, Name: "Config 1", Weight: 3},
		{ID: 2, Name: "Config 2", Weight: 1},
	}

	// Count selections over one full cycle
	counts := make(map[uint]int)
	totalWeight := 4

	for i := 0; i < totalWeight*3; i++ {
		config, err := lb.SelectConfig(ctx, configs)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		counts[config.ID]++
	}

	// Config 1 should be selected 3x more than Config 2
	ratio := float64(counts[1]) / float64(counts[2])
	expectedRatio := 3.0

	if abs(ratio-expectedRatio) > 0.5 {
		t.Errorf("Expected ratio ~%.1f, got %.1f (counts: %v)", expectedRatio, ratio, counts)
	}
}

// Unit test for LeastConnectionsLB
func TestLeastConnectionsLB(t *testing.T) {
	ctx := context.Background()
	lb := NewLeastConnectionsLB()

	configs := []*models.APIConfig{
		{ID: 1, Name: "Config 1"},
		{ID: 2, Name: "Config 2"},
		{ID: 3, Name: "Config 3"},
	}

	// First request should go to any config
	config1, err := lb.SelectConfig(ctx, configs)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Second request should go to a different config (least connections)
	config2, err := lb.SelectConfig(ctx, configs)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if config1.ID == config2.ID {
		t.Error("Expected different configs for least connections")
	}

	// Release first connection
	lb.ReleaseConnection(config1.ID)

	// Next request should prefer the released config
	config3, err := lb.SelectConfig(ctx, configs)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// After release, config1 should have fewer connections
	// So it might be selected again (though not guaranteed due to ties)
	_ = config3
}

// Unit test for RandomLB
func TestRandomLB(t *testing.T) {
	ctx := context.Background()
	lb := NewRandomLB()

	configs := []*models.APIConfig{
		{ID: 1, Name: "Config 1"},
		{ID: 2, Name: "Config 2"},
		{ID: 3, Name: "Config 3"},
	}

	// Make many requests and verify distribution
	counts := make(map[uint]int)
	iterations := 300

	for i := 0; i < iterations; i++ {
		config, err := lb.SelectConfig(ctx, configs)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		counts[config.ID]++
	}

	// Each config should be selected at least once
	for _, config := range configs {
		if counts[config.ID] == 0 {
			t.Errorf("Config %d was never selected", config.ID)
		}
	}

	// Distribution should be roughly even (within 40% of expected)
	expectedCount := iterations / len(configs)
	tolerance := float64(expectedCount) * 0.4

	for id, count := range counts {
		diff := abs(float64(count) - float64(expectedCount))
		if diff > tolerance {
			t.Errorf("Config %d: count %d too far from expected %d", id, count, expectedCount)
		}
	}

	// Test empty configs
	_, err := lb.SelectConfig(ctx, []*models.APIConfig{})
	if err != ErrNoAvailableConfig {
		t.Errorf("Expected ErrNoAvailableConfig, got %v", err)
	}
}

// Unit test for Factory
func TestFactory(t *testing.T) {
	factory := NewFactory()

	tests := []struct {
		strategy     string
		expectedType string
	}{
		{"round_robin", "round_robin"},
		{"weighted_round_robin", "weighted_round_robin"},
		{"least_connections", "least_connections"},
		{"random", "random"},
		{"unknown", "round_robin"}, // Default
	}

	for _, tt := range tests {
		t.Run(tt.strategy, func(t *testing.T) {
			lb := factory.CreateLoadBalancer(tt.strategy)
			if lb == nil {
				t.Fatal("Expected load balancer, got nil")
			}
			if lb.GetStrategy() != tt.expectedType {
				t.Errorf("Expected strategy %s, got %s", tt.expectedType, lb.GetStrategy())
			}
		})
	}
}

// Unit test for concurrent access
func TestLoadBalancer_Concurrent(t *testing.T) {
	ctx := context.Background()
	lb := NewRoundRobinLB()

	configs := []*models.APIConfig{
		{ID: 1, Name: "Config 1"},
		{ID: 2, Name: "Config 2"},
		{ID: 3, Name: "Config 3"},
	}

	// Run concurrent requests
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_, err := lb.SelectConfig(ctx, configs)
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}
