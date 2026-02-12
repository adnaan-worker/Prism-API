package cache

import (
	"time"
)

// Cache 缓存接口
type Cache interface {
	Get(key string, value interface{}) error
	Set(key string, value interface{}, expiration time.Duration) error
	Delete(key string) error
	Exists(key string) (bool, error)
	Clear() error
}

// RedisConfig Redis配置
type RedisConfig struct {
	Addr         string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
}
