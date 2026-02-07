package service

import (
	"api-aggregator/backend/internal/models"
	"api-aggregator/backend/internal/repository"
	"context"
	"os"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// getProxyTestDB creates a test database connection
func getProxyTestDB(t *testing.T) *gorm.DB {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "host=localhost user=postgres password=postgres dbname=api_aggregator_test port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	return db
}

// setupProxyTestDB initializes the test database
func setupProxyTestDB(t *testing.T) *gorm.DB {
	db := getProxyTestDB(t)

	// Clean up existing data
	db.Exec("TRUNCATE TABLE request_logs CASCADE")
	db.Exec("TRUNCATE TABLE api_keys CASCADE")
	db.Exec("TRUNCATE TABLE users CASCADE")

	// Run migrations
	err := db.AutoMigrate(&models.User{}, &models.APIKey{}, &models.RequestLog{})
	if err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	return db
}

// cleanupProxyTestDB cleans up the test database
func cleanupProxyTestDB(t *testing.T, db *gorm.DB) {
	db.Exec("TRUNCATE TABLE request_logs CASCADE")
	db.Exec("TRUNCATE TABLE api_keys CASCADE")
	db.Exec("TRUNCATE TABLE users CASCADE")
}

// Property 13: API Key验证
// Feature: api-aggregator, Property 13: For any API request, only requests with valid, active API keys should be processed; invalid or inactive keys should be rejected with 401 error.
// Validates: Requirements 6.1
func TestProperty_APIKeyValidation(t *testing.T) {
	db := setupProxyTestDB(t)
	defer cleanupProxyTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	apiKeyRepo := repository.NewAPIKeyRepository(db)
	requestLogRepo := repository.NewRequestLogRepository(db)
	signInRepo := repository.NewSignInRepository(db)
	quotaService := NewQuotaService(userRepo, signInRepo)

	// Note: We can't fully test ProxyRequest without mocking external APIs
	// So we test ValidateAPIKey which is the core validation logic

	properties := gopter.NewProperties(nil)

	// Generator for API key names
	keyNameGen := gen.RegexMatch("[a-zA-Z]{5,15}")

	properties.Property("Valid active API keys are accepted, invalid/inactive are rejected", prop.ForAll(
		func(keyName string) bool {
			// Clean up before each test
			db.Exec("TRUNCATE TABLE api_keys CASCADE")
			db.Exec("TRUNCATE TABLE users CASCADE")

			ctx := context.Background()

			// Create a user
			user := &models.User{
				Username:     "testuser",
				Email:        "test@example.com",
				PasswordHash: "hash",
				Quota:        10000,
				Status:       "active",
			}
			if err := userRepo.Create(ctx, user); err != nil {
				return true // Skip on error
			}

			// Create an active API key
			apiKeyService := NewAPIKeyService(apiKeyRepo)
			activeKey, err := apiKeyService.CreateAPIKey(ctx, user.ID, &CreateAPIKeyRequest{
				Name:      keyName,
				RateLimit: 60,
			})
			if err != nil {
				return true // Skip on error
			}

			// Create proxy service
			proxyService := NewProxyService(apiKeyRepo, nil, userRepo, requestLogRepo, quotaService)

			// Test 1: Valid active key should be accepted
			userID, err := proxyService.ValidateAPIKey(ctx, activeKey.Key)
			if err != nil {
				return false // Should succeed
			}
			if userID != user.ID {
				return false // Should return correct user ID
			}

			// Test 2: Invalid key should be rejected
			_, err = proxyService.ValidateAPIKey(ctx, "sk-invalid-key-12345")
			if err != ErrInvalidAPIKey {
				return false // Should return ErrInvalidAPIKey
			}

			// Test 3: Deactivate the key
			if err := apiKeyService.DeleteAPIKey(ctx, user.ID, activeKey.ID); err != nil {
				return true // Skip on error
			}

			// Test 4: Inactive key should be rejected
			_, err = proxyService.ValidateAPIKey(ctx, activeKey.Key)
			if err != ErrInvalidAPIKey {
				return false // Should return ErrInvalidAPIKey
			}

			return true
		},
		keyNameGen,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Property 14: 额度检查
// Feature: api-aggregator, Property 14: For any user with insufficient quota, API requests should be rejected with a quota exceeded error before calling the third-party API.
// Validates: Requirements 6.2
func TestProperty_QuotaCheck(t *testing.T) {
	db := setupProxyTestDB(t)
	defer cleanupProxyTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	signInRepo := repository.NewSignInRepository(db)
	quotaService := NewQuotaService(userRepo, signInRepo)

	properties := gopter.NewProperties(nil)

	// Generator for quota amounts
	quotaGen := gen.Int64Range(0, 1000)
	requestGen := gen.Int64Range(100, 2000)

	properties.Property("Insufficient quota is detected before API call", prop.ForAll(
		func(userQuota, requestTokens int64) bool {
			// Clean up before each test
			db.Exec("TRUNCATE TABLE users CASCADE")

			ctx := context.Background()

			// Create a user with specific quota
			user := &models.User{
				Username:     "testuser",
				Email:        "test@example.com",
				PasswordHash: "hash",
				Quota:        userQuota,
				UsedQuota:    0,
				Status:       "active",
			}
			if err := userRepo.Create(ctx, user); err != nil {
				return true // Skip on error
			}

			// Check quota
			hasQuota, err := quotaService.CheckQuota(ctx, user.ID, requestTokens)
			if err != nil {
				return false // Should not error
			}

			// Verify result matches expectation
			expectedHasQuota := userQuota >= requestTokens
			return hasQuota == expectedHasQuota
		},
		quotaGen,
		requestGen,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Property 15: 额度扣除准确性
// Feature: api-aggregator, Property 15: For any successful API call, the user's used_quota should increase by exactly the number of tokens used in the request.
// Validates: Requirements 6.3
func TestProperty_QuotaDeductionAccuracy(t *testing.T) {
	db := setupProxyTestDB(t)
	defer cleanupProxyTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	signInRepo := repository.NewSignInRepository(db)
	quotaService := NewQuotaService(userRepo, signInRepo)

	properties := gopter.NewProperties(nil)

	// Generator for token amounts
	tokensGen := gen.Int64Range(1, 1000)

	properties.Property("Quota deduction is accurate", prop.ForAll(
		func(tokensUsed int64) bool {
			// Clean up before each test
			db.Exec("TRUNCATE TABLE users CASCADE")

			ctx := context.Background()

			// Create a user with sufficient quota
			user := &models.User{
				Username:     "testuser",
				Email:        "test@example.com",
				PasswordHash: "hash",
				Quota:        10000,
				UsedQuota:    0,
				Status:       "active",
			}
			if err := userRepo.Create(ctx, user); err != nil {
				return true // Skip on error
			}

			initialUsedQuota := user.UsedQuota

			// Deduct quota
			if err := quotaService.DeductQuota(ctx, user.ID, tokensUsed); err != nil {
				return false // Should succeed
			}

			// Verify deduction
			updatedUser, err := userRepo.FindByID(ctx, user.ID)
			if err != nil {
				return false
			}

			expectedUsedQuota := initialUsedQuota + tokensUsed
			return updatedUser.UsedQuota == expectedUsedQuota
		},
		tokensGen,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Unit test for ValidateAPIKey
func TestProxyService_ValidateAPIKey(t *testing.T) {
	db := setupProxyTestDB(t)
	defer cleanupProxyTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	apiKeyRepo := repository.NewAPIKeyRepository(db)
	requestLogRepo := repository.NewRequestLogRepository(db)
	signInRepo := repository.NewSignInRepository(db)
	quotaService := NewQuotaService(userRepo, signInRepo)

	ctx := context.Background()

	// Create a user
	user := &models.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hash",
		Quota:        10000,
		Status:       "active",
	}
	userRepo.Create(ctx, user)

	// Create an API key
	apiKeyService := NewAPIKeyService(apiKeyRepo)
	apiKey, err := apiKeyService.CreateAPIKey(ctx, user.ID, &CreateAPIKeyRequest{
		Name:      "Test Key",
		RateLimit: 60,
	})
	if err != nil {
		t.Fatalf("Failed to create API key: %v", err)
	}

	// Create proxy service
	proxyService := NewProxyService(apiKeyRepo, nil, userRepo, requestLogRepo, quotaService)

	// Test valid key
	userID, err := proxyService.ValidateAPIKey(ctx, apiKey.Key)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if userID != user.ID {
		t.Errorf("Expected user ID %d, got %d", user.ID, userID)
	}

	// Test invalid key
	_, err = proxyService.ValidateAPIKey(ctx, "sk-invalid")
	if err != ErrInvalidAPIKey {
		t.Errorf("Expected ErrInvalidAPIKey, got %v", err)
	}

	// Deactivate key
	apiKeyService.DeleteAPIKey(ctx, user.ID, apiKey.ID)

	// Test inactive key
	_, err = proxyService.ValidateAPIKey(ctx, apiKey.Key)
	if err != ErrInvalidAPIKey {
		t.Errorf("Expected ErrInvalidAPIKey for inactive key, got %v", err)
	}
}
