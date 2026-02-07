package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"api-aggregator/backend/pkg/redis"
)

// CacheService handles caching operations
type CacheService struct {
	defaultTTL time.Duration
}

// NewCacheService creates a new cache service
func NewCacheService(defaultTTL time.Duration) *CacheService {
	return &CacheService{
		defaultTTL: defaultTTL,
	}
}

// Get retrieves a value from cache
func (s *CacheService) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := redis.Client.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(val), dest)
}

// Set stores a value in cache with default TTL
func (s *CacheService) Set(ctx context.Context, key string, value interface{}) error {
	return s.SetWithTTL(ctx, key, value, s.defaultTTL)
}

// SetWithTTL stores a value in cache with custom TTL
func (s *CacheService) SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return redis.Client.Set(ctx, key, data, ttl).Err()
}

// Delete removes a value from cache
func (s *CacheService) Delete(ctx context.Context, key string) error {
	return redis.Client.Del(ctx, key).Err()
}

// DeletePattern removes all keys matching a pattern
func (s *CacheService) DeletePattern(ctx context.Context, pattern string) error {
	iter := redis.Client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := redis.Client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}

// CacheKey generates a cache key with prefix
func CacheKey(prefix string, parts ...interface{}) string {
	key := prefix
	for _, part := range parts {
		key += fmt.Sprintf(":%v", part)
	}
	return key
}
