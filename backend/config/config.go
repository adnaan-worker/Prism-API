package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	Database  DatabaseConfig
	Redis     RedisConfig
	Server    ServerConfig
	JWT       JWTConfig
	Embedding EmbeddingConfig
	Cache     CacheConfig
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	URL         string
	PoolSize    int
	MinIdleConn int
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port           string
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	RequestTimeout time.Duration
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret string
}

// EmbeddingConfig holds embedding service configuration
type EmbeddingConfig struct {
	URL     string
	Timeout time.Duration
	Enabled bool
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	Enabled       bool
	TTL           time.Duration
	SemanticMatch bool
	Threshold     float64
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Database: DatabaseConfig{
			URL:             getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/api_aggregator?sslmode=disable"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		Redis: RedisConfig{
			URL:         getEnv("REDIS_URL", "redis://localhost:6379"),
			PoolSize:    getEnvAsInt("REDIS_POOL_SIZE", 10),
			MinIdleConn: getEnvAsInt("REDIS_MIN_IDLE_CONN", 2),
		},
		Server: ServerConfig{
			Port:           getEnv("PORT", "8080"),
			ReadTimeout:    getEnvAsDuration("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout:   getEnvAsDuration("SERVER_WRITE_TIMEOUT", 10*time.Second),
			RequestTimeout: getEnvAsDuration("REQUEST_TIMEOUT", 30*time.Second),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		},
		Embedding: EmbeddingConfig{
			URL:     getEnv("EMBEDDING_URL", "http://localhost:8765"),
			Timeout: getEnvAsDuration("EMBEDDING_TIMEOUT", 30*time.Second),
			Enabled: getEnvAsBool("EMBEDDING_ENABLED", true),
		},
		Cache: CacheConfig{
			Enabled:       getEnvAsBool("CACHE_ENABLED", true),
			TTL:           getEnvAsDuration("CACHE_TTL", 24*time.Hour),
			SemanticMatch: getEnvAsBool("CACHE_SEMANTIC_MATCH", true),
			Threshold:     getEnvAsFloat("CACHE_THRESHOLD", 0.85),
		},
	}

	// Validate required fields
	if cfg.JWT.Secret == "your-secret-key-change-in-production" {
		fmt.Println("WARNING: Using default JWT secret. Please set JWT_SECRET environment variable in production.")
	}

	return cfg, nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as int or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsDuration gets an environment variable as duration or returns a default value
func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// getEnvAsBool gets an environment variable as bool or returns a default value
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// getEnvAsFloat gets an environment variable as float64 or returns a default value
func getEnvAsFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}
