package middleware

import (
	"api-aggregator/backend/internal/models"
	"api-aggregator/backend/internal/repository"
	"api-aggregator/backend/internal/service"
	"api-aggregator/backend/pkg/redis"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var testDB *gorm.DB

func setupTestDB() error {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=api_aggregator port=5432 sslmode=disable"
	}

	var err error
	testDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	return nil
}

func cleanupTestDB() {
	if testDB != nil {
		testDB.Exec("TRUNCATE TABLE request_logs, sign_in_records, api_keys, api_configs, load_balancer_configs, users RESTART IDENTITY CASCADE")
	}
}

func setupTestRedis() error {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379"
	}

	return redis.InitRedis(redis.Config{
		URL:         redisURL,
		PoolSize:    10,
		MinIdleConn: 2,
	})
}

func cleanupTestRedis() {
	if redis.Client != nil {
		ctx := context.Background()
		redis.Client.FlushDB(ctx)
	}
}

func TestMain(m *testing.M) {
	if err := setupTestDB(); err != nil {
		fmt.Printf("Failed to setup test database: %v\n", err)
		os.Exit(1)
	}

	if err := setupTestRedis(); err != nil {
		fmt.Printf("Failed to setup test Redis: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()

	cleanupTestDB()
	cleanupTestRedis()

	os.Exit(code)
}

// Property 18: Rate Limiting Protection
// For any API key with a rate limit of N requests per minute,
// making more than N requests within one minute should result in
// 429 Too Many Requests errors for the excess requests.
func TestProperty_RateLimitProtection(t *testing.T) {
	cleanupTestDB()
	cleanupTestRedis()

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("Rate limit protection - excess requests return 429", prop.ForAll(
		func(rateLimit int) bool {
			cleanupTestRedis() // Clean Redis between tests

			// Create test user
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
			user := &models.User{
				Username:     fmt.Sprintf("ratelimit_user_%d_%d", rateLimit, time.Now().UnixNano()),
				Email:        fmt.Sprintf("ratelimit_%d_%d@test.com", rateLimit, time.Now().UnixNano()),
				PasswordHash: string(hashedPassword),
				Quota:        10000,
				UsedQuota:    0,
				IsAdmin:      false,
			}
			if err := testDB.Create(user).Error; err != nil {
				t.Logf("Failed to create user: %v", err)
				return false
			}

			// Create API key with specific rate limit
			apiKeyRepo := repository.NewAPIKeyRepository(testDB)
			apiKeyService := service.NewAPIKeyService(apiKeyRepo)

			apiKey, err := apiKeyService.CreateAPIKey(context.Background(), user.ID, &service.CreateAPIKeyRequest{
				Name:      "Test Key",
				RateLimit: rateLimit,
			})
			if err != nil {
				t.Logf("Failed to create API key: %v", err)
				return false
			}

			// Setup Gin router with rate limiter middleware
			gin.SetMode(gin.TestMode)
			router := gin.New()

			// Middleware that sets the API key in context
			router.Use(func(c *gin.Context) {
				c.Set("api_key", apiKey)
				c.Next()
			})

			// Apply rate limiter middleware
			router.Use(RateLimiterMiddleware())

			// Test endpoint
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			// Make rateLimit + 5 requests
			successCount := 0
			rateLimitedCount := 0
			totalRequests := rateLimit + 5

			for i := 0; i < totalRequests; i++ {
				req := httptest.NewRequest("GET", "/test", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				if w.Code == http.StatusOK {
					successCount++
				} else if w.Code == http.StatusTooManyRequests {
					rateLimitedCount++
				}
			}

			// Verify: first rateLimit requests should succeed
			if successCount != rateLimit {
				t.Logf("Expected %d successful requests, got %d", rateLimit, successCount)
				return false
			}

			// Verify: excess requests should be rate limited
			if rateLimitedCount != 5 {
				t.Logf("Expected 5 rate limited requests, got %d", rateLimitedCount)
				return false
			}

			return true
		},
		gen.IntRange(1, 20), // Test with rate limits from 1 to 20
	))

	properties.TestingRun(t)
}

// Unit test: Rate limiter middleware rejects requests without API key in context
func TestRateLimiterMiddleware_NoAPIKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RateLimiterMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

// Unit test: Rate limiter allows requests within limit
func TestRateLimiterMiddleware_WithinLimit(t *testing.T) {
	cleanupTestDB()
	cleanupTestRedis()

	// Create test user
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user := &models.User{
		Username:     fmt.Sprintf("test_user_%d", time.Now().UnixNano()),
		Email:        fmt.Sprintf("test_%d@test.com", time.Now().UnixNano()),
		PasswordHash: string(hashedPassword),
		Quota:        10000,
		UsedQuota:    0,
		IsAdmin:      false,
	}
	if err := testDB.Create(user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create API key
	apiKeyRepo := repository.NewAPIKeyRepository(testDB)
	apiKeyService := service.NewAPIKeyService(apiKeyRepo)

	apiKey, err := apiKeyService.CreateAPIKey(context.Background(), user.ID, &service.CreateAPIKeyRequest{
		Name:      "Test Key",
		RateLimit: 10,
	})
	if err != nil {
		t.Fatalf("Failed to create API key: %v", err)
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(func(c *gin.Context) {
		c.Set("api_key", apiKey)
		c.Next()
	})

	router.Use(RateLimiterMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Make 5 requests (within limit of 10)
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d: Expected status 200, got %d", i+1, w.Code)
		}
	}
}
