package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

// Config holds Redis configuration
type Config struct {
	URL         string
	PoolSize    int
	MinIdleConn int
}

// InitRedis initializes the Redis client with connection pooling
func InitRedis(cfg Config) error {
	opt, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return err
	}

	// Configure connection pool
	opt.PoolSize = cfg.PoolSize
	opt.MinIdleConns = cfg.MinIdleConn
	opt.PoolTimeout = 4 * time.Second
	opt.ConnMaxIdleTime = 5 * time.Minute

	Client = redis.NewClient(opt)

	// Test connection
	ctx := context.Background()
	if err := Client.Ping(ctx).Err(); err != nil {
		return err
	}

	return nil
}

// CloseRedis closes the Redis connection
func CloseRedis() error {
	if Client != nil {
		return Client.Close()
	}
	return nil
}
