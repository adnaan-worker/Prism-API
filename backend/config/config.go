package config

import (
	"fmt"
	"log"
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
	Admin     AdminConfig
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

// AdminConfig holds initial admin user configuration
type AdminConfig struct {
	Username string
	Email    string
	Password string
}

// Load loads configuration from environment variables
// It automatically loads .env file if it exists
func Load() (*Config, error) {
	// Try to load .env file
	loadEnvFile()
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
		Admin: AdminConfig{
			Username: getEnv("ADMIN_USERNAME", ""),
			Email:    getEnv("ADMIN_EMAIL", ""),
			Password: getEnv("ADMIN_PASSWORD", ""),
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

// loadEnvFile loads environment variables from .env file
func loadEnvFile() {
	// Try multiple possible locations for .env file
	envPaths := []string{
		".env",
		"../.env",
		"../../.env",
	}

	for _, envPath := range envPaths {
		if err := loadEnvFromFile(envPath); err == nil {
			log.Printf("Loaded environment variables from %s", envPath)
			return
		}
	}
}

// loadEnvFromFile loads environment variables from a specific file
func loadEnvFromFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	// Simple .env file parser
	lines := string(data)
	currentLine := ""
	for _, char := range lines {
		if char == '\n' {
			processEnvLine(currentLine)
			currentLine = ""
		} else {
			currentLine += string(char)
		}
	}
	if currentLine != "" {
		processEnvLine(currentLine)
	}

	return nil
}

// processEnvLine processes a single line from .env file
func processEnvLine(line string) {
	// Remove leading/trailing whitespace
	line = trimSpace(line)
	
	// Skip empty lines and comments
	if line == "" || (len(line) > 0 && line[0] == '#') {
		return
	}

	// Find the = sign
	eqIndex := -1
	for i, char := range line {
		if char == '=' {
			eqIndex = i
			break
		}
	}

	if eqIndex == -1 {
		return // No = sign found
	}

	key := trimSpace(line[:eqIndex])
	value := trimSpace(line[eqIndex+1:])

	// Remove quotes if present
	if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') ||
		(value[0] == '\'' && value[len(value)-1] == '\'')) {
		value = value[1 : len(value)-1]
	}

	// Only set if not already set in environment
	if os.Getenv(key) == "" {
		os.Setenv(key, value)
	}
}

// trimSpace removes leading and trailing whitespace
func trimSpace(s string) string {
	start := 0
	end := len(s)
	
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\r') {
		start++
	}
	
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\r') {
		end--
	}
	
	return s[start:end]
}
