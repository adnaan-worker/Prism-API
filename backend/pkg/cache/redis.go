package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
)

// redisCache Redis缓存实现
type redisCache struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedis 创建Redis缓存实例
func NewRedis(config *RedisConfig) (Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         config.Addr,
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &redisCache{
		client: client,
		ctx:    ctx,
	}, nil
}

// Get 获取缓存
func (c *redisCache) Get(key string, value interface{}) error {
	data, err := c.client.Get(c.ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, value)
}

// Set 设置缓存
func (c *redisCache) Set(key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.Set(c.ctx, key, data, expiration).Err()
}

// Delete 删除缓存
func (c *redisCache) Delete(key string) error {
	return c.client.Del(c.ctx, key).Err()
}

// Exists 检查缓存是否存在
func (c *redisCache) Exists(key string) (bool, error) {
	n, err := c.client.Exists(c.ctx, key).Result()
	return n > 0, err
}

// Clear 清空所有缓存
func (c *redisCache) Clear() error {
	return c.client.FlushDB(c.ctx).Err()
}
