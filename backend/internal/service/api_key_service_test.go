package service

import (
	"api-aggregator/backend/internal/models"
	"api-aggregator/backend/internal/repository"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// getTestDBForAPIKey creates a test database connection
func getTestDBForAPIKey(t *testing.T) *gorm.DB {
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

// setupTestDBForAPIKey initializes the test database
func setupTestDBForAPIKey(t *testing.T) *gorm.DB {
	db := getTestDBForAPIKey(t)

	// Clean up existing data
	db.Exec("TRUNCATE TABLE api_keys CASCADE")
	db.Exec("TRUNCATE TABLE users CASCADE")

	// Run migrations
	err := db.AutoMigrate(&models.User{}, &models.APIKey{})
	if err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	return db
}

// cleanupTestDBForAPIKey cleans up the test database
func cleanupTestDBForAPIKey(t *testing.T, db *gorm.DB) {
	db.Exec("TRUNCATE TABLE api_keys CASCADE")
	db.Exec("TRUNCATE TABLE users CASCADE")
}

// Property 5: API密钥唯一性和格式
// Feature: api-aggregator, Property 5: For any user creating an API key, the generated key should be unique, start with "sk-", and be associated with the correct user_id.
// Validates: Requirements 2.1
func TestProperty_APIKeyUniquenessAndFormat(t *testing.T) {
	db := setupTestDBForAPIKey(t)
	defer cleanupTestDBForAPIKey(t, db)

	userRepo := repository.NewUserRepository(db)
	apiKeyRepo := repository.NewAPIKeyRepository(db)
	apiKeyService := NewAPIKeyService(apiKeyRepo)

	properties := gopter.NewProperties(nil)

	// Custom generators
	nameGen := gen.RegexMatch("[a-zA-Z][a-zA-Z0-9_-]{2,49}")
	rateLimitGen := gen.IntRange(1, 1000)

	properties.Property("API key uniqueness and format", prop.ForAll(
		func(name string, rateLimit int) bool {
			// Clean up before each test
			db.Exec("TRUNCATE TABLE api_keys CASCADE")
			db.Exec("TRUNCATE TABLE users CASCADE")

			ctx := context.Background()

			// Create a test user
			user := &models.User{
				Username:     "testuser",
				Email:        "test@example.com",
				PasswordHash: "hash",
				Quota:        10000,
			}
			if err := userRepo.Create(ctx, user); err != nil {
				return true // Skip if user creation fails
			}

			// Create first API key
			req1 := &CreateAPIKeyRequest{
				Name:      strings.TrimSpace(name),
				RateLimit: rateLimit,
			}
			apiKey1, err := apiKeyService.CreateAPIKey(ctx, user.ID, req1)
			if err != nil {
				return false // Should succeed
			}

			// Check format: should start with "sk-"
			if !strings.HasPrefix(apiKey1.Key, "sk-") {
				return false
			}

			// Check user association
			if apiKey1.UserID != user.ID {
				return false
			}

			// Create second API key
			req2 := &CreateAPIKeyRequest{
				Name:      strings.TrimSpace(name) + "_2",
				RateLimit: rateLimit,
			}
			apiKey2, err := apiKeyService.CreateAPIKey(ctx, user.ID, req2)
			if err != nil {
				return false // Should succeed
			}

			// Check uniqueness: keys should be different
			if apiKey1.Key == apiKey2.Key {
				return false
			}

			// Check format for second key
			if !strings.HasPrefix(apiKey2.Key, "sk-") {
				return false
			}

			return true
		},
		nameGen,
		rateLimitGen,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Property 6: API密钥权限隔离
// Feature: api-aggregator, Property 6: For any two different users, each user should only be able to view and manage their own API keys, not the other user's keys.
// Validates: Requirements 2.2
func TestProperty_APIKeyPermissionIsolation(t *testing.T) {
	db := setupTestDBForAPIKey(t)
	defer cleanupTestDBForAPIKey(t, db)

	userRepo := repository.NewUserRepository(db)
	apiKeyRepo := repository.NewAPIKeyRepository(db)
	apiKeyService := NewAPIKeyService(apiKeyRepo)

	properties := gopter.NewProperties(nil)

	// Custom generators
	nameGen := gen.RegexMatch("[a-zA-Z][a-zA-Z0-9_-]{2,49}")

	properties.Property("API key permission isolation", prop.ForAll(
		func(name1, name2 string) bool {
			// Clean up before each test
			db.Exec("TRUNCATE TABLE api_keys CASCADE")
			db.Exec("TRUNCATE TABLE users CASCADE")

			ctx := context.Background()

			// Create two test users
			user1 := &models.User{
				Username:     "user1",
				Email:        "user1@example.com",
				PasswordHash: "hash",
				Quota:        10000,
			}
			if err := userRepo.Create(ctx, user1); err != nil {
				return true // Skip if user creation fails
			}

			user2 := &models.User{
				Username:     "user2",
				Email:        "user2@example.com",
				PasswordHash: "hash",
				Quota:        10000,
			}
			if err := userRepo.Create(ctx, user2); err != nil {
				return true // Skip if user creation fails
			}

			// Create API key for user1
			req1 := &CreateAPIKeyRequest{
				Name:      strings.TrimSpace(name1),
				RateLimit: 60,
			}
			apiKey1, err := apiKeyService.CreateAPIKey(ctx, user1.ID, req1)
			if err != nil {
				return false // Should succeed
			}

			// Create API key for user2
			req2 := &CreateAPIKeyRequest{
				Name:      strings.TrimSpace(name2),
				RateLimit: 60,
			}
			apiKey2, err := apiKeyService.CreateAPIKey(ctx, user2.ID, req2)
			if err != nil {
				return false // Should succeed
			}

			// User1 should only see their own keys
			user1Keys, err := apiKeyService.GetAPIKeysByUserID(ctx, user1.ID)
			if err != nil {
				return false
			}
			if len(user1Keys) != 1 {
				return false
			}
			if user1Keys[0].ID != apiKey1.ID {
				return false
			}

			// User2 should only see their own keys
			user2Keys, err := apiKeyService.GetAPIKeysByUserID(ctx, user2.ID)
			if err != nil {
				return false
			}
			if len(user2Keys) != 1 {
				return false
			}
			if user2Keys[0].ID != apiKey2.ID {
				return false
			}

			// User2 should not be able to delete user1's key
			err = apiKeyService.DeleteAPIKey(ctx, user2.ID, apiKey1.ID)
			if err != ErrUnauthorizedAccess {
				return false // Should return unauthorized error
			}

			// User1 should be able to delete their own key
			err = apiKeyService.DeleteAPIKey(ctx, user1.ID, apiKey1.ID)
			if err != nil {
				return false // Should succeed
			}

			// Verify user1's key is deleted
			user1Keys, err = apiKeyService.GetAPIKeysByUserID(ctx, user1.ID)
			if err != nil {
				return false
			}
			if len(user1Keys) != 0 {
				return false
			}

			// Verify user2's key is still there
			user2Keys, err = apiKeyService.GetAPIKeysByUserID(ctx, user2.ID)
			if err != nil {
				return false
			}
			if len(user2Keys) != 1 {
				return false
			}

			return true
		},
		nameGen,
		nameGen,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Unit test for API key creation
func TestAPIKeyService_CreateAPIKey(t *testing.T) {
	db := setupTestDBForAPIKey(t)
	defer cleanupTestDBForAPIKey(t, db)

	userRepo := repository.NewUserRepository(db)
	apiKeyRepo := repository.NewAPIKeyRepository(db)
	apiKeyService := NewAPIKeyService(apiKeyRepo)

	ctx := context.Background()

	// Create a test user
	user := &models.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hash",
		Quota:        10000,
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Test successful API key creation
	req := &CreateAPIKeyRequest{
		Name:      "Test Key",
		RateLimit: 100,
	}

	apiKey, err := apiKeyService.CreateAPIKey(ctx, user.ID, req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if apiKey.Name != req.Name {
		t.Errorf("Expected name %s, got %s", req.Name, apiKey.Name)
	}
	if apiKey.RateLimit != req.RateLimit {
		t.Errorf("Expected rate limit %d, got %d", req.RateLimit, apiKey.RateLimit)
	}
	if !strings.HasPrefix(apiKey.Key, "sk-") {
		t.Errorf("Expected key to start with 'sk-', got %s", apiKey.Key)
	}
	if apiKey.UserID != user.ID {
		t.Errorf("Expected user ID %d, got %d", user.ID, apiKey.UserID)
	}
	if !apiKey.IsActive {
		t.Error("Expected is_active true, got false")
	}

	// Test default rate limit
	req2 := &CreateAPIKeyRequest{
		Name: "Test Key 2",
	}
	apiKey2, err := apiKeyService.CreateAPIKey(ctx, user.ID, req2)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if apiKey2.RateLimit != 60 {
		t.Errorf("Expected default rate limit 60, got %d", apiKey2.RateLimit)
	}
}

// Unit test for getting API keys by user ID
func TestAPIKeyService_GetAPIKeysByUserID(t *testing.T) {
	db := setupTestDBForAPIKey(t)
	defer cleanupTestDBForAPIKey(t, db)

	userRepo := repository.NewUserRepository(db)
	apiKeyRepo := repository.NewAPIKeyRepository(db)
	apiKeyService := NewAPIKeyService(apiKeyRepo)

	ctx := context.Background()

	// Create a test user
	user := &models.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hash",
		Quota:        10000,
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create multiple API keys
	for i := 0; i < 3; i++ {
		req := &CreateAPIKeyRequest{
			Name:      "Test Key",
			RateLimit: 60,
		}
		_, err := apiKeyService.CreateAPIKey(ctx, user.ID, req)
		if err != nil {
			t.Fatalf("Failed to create API key: %v", err)
		}
	}

	// Get all keys for user
	apiKeys, err := apiKeyService.GetAPIKeysByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(apiKeys) != 3 {
		t.Errorf("Expected 3 API keys, got %d", len(apiKeys))
	}
}

// Unit test for deleting API key
func TestAPIKeyService_DeleteAPIKey(t *testing.T) {
	db := setupTestDBForAPIKey(t)
	defer cleanupTestDBForAPIKey(t, db)

	userRepo := repository.NewUserRepository(db)
	apiKeyRepo := repository.NewAPIKeyRepository(db)
	apiKeyService := NewAPIKeyService(apiKeyRepo)

	ctx := context.Background()

	// Create two test users
	user1 := &models.User{
		Username:     "user1",
		Email:        "user1@example.com",
		PasswordHash: "hash",
		Quota:        10000,
	}
	if err := userRepo.Create(ctx, user1); err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}

	user2 := &models.User{
		Username:     "user2",
		Email:        "user2@example.com",
		PasswordHash: "hash",
		Quota:        10000,
	}
	if err := userRepo.Create(ctx, user2); err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}

	// Create API key for user1
	req := &CreateAPIKeyRequest{
		Name:      "Test Key",
		RateLimit: 60,
	}
	apiKey, err := apiKeyService.CreateAPIKey(ctx, user1.ID, req)
	if err != nil {
		t.Fatalf("Failed to create API key: %v", err)
	}

	// Test unauthorized deletion (user2 trying to delete user1's key)
	err = apiKeyService.DeleteAPIKey(ctx, user2.ID, apiKey.ID)
	if err != ErrUnauthorizedAccess {
		t.Errorf("Expected ErrUnauthorizedAccess, got %v", err)
	}

	// Test successful deletion (user1 deleting their own key)
	err = apiKeyService.DeleteAPIKey(ctx, user1.ID, apiKey.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify key is deleted
	apiKeys, err := apiKeyService.GetAPIKeysByUserID(ctx, user1.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(apiKeys) != 0 {
		t.Errorf("Expected 0 API keys after deletion, got %d", len(apiKeys))
	}

	// Test deleting non-existent key
	err = apiKeyService.DeleteAPIKey(ctx, user1.ID, 99999)
	if err != ErrAPIKeyNotFound {
		t.Errorf("Expected ErrAPIKeyNotFound, got %v", err)
	}
}

// Unit test for validating API key
func TestAPIKeyService_ValidateAPIKey(t *testing.T) {
	db := setupTestDBForAPIKey(t)
	defer cleanupTestDBForAPIKey(t, db)

	userRepo := repository.NewUserRepository(db)
	apiKeyRepo := repository.NewAPIKeyRepository(db)
	apiKeyService := NewAPIKeyService(apiKeyRepo)

	ctx := context.Background()

	// Create a test user
	user := &models.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hash",
		Quota:        10000,
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create API key
	req := &CreateAPIKeyRequest{
		Name:      "Test Key",
		RateLimit: 60,
	}
	apiKey, err := apiKeyService.CreateAPIKey(ctx, user.ID, req)
	if err != nil {
		t.Fatalf("Failed to create API key: %v", err)
	}

	// Test valid API key
	userID, err := apiKeyService.ValidateAPIKey(ctx, apiKey.Key)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if userID != user.ID {
		t.Errorf("Expected user ID %d, got %d", user.ID, userID)
	}

	// Test invalid API key
	_, err = apiKeyService.ValidateAPIKey(ctx, "sk-invalid")
	if err != ErrAPIKeyNotFound {
		t.Errorf("Expected ErrAPIKeyNotFound, got %v", err)
	}

	// Test inactive API key
	apiKey.IsActive = false
	if err := apiKeyRepo.Update(ctx, apiKey); err != nil {
		t.Fatalf("Failed to update API key: %v", err)
	}
	_, err = apiKeyService.ValidateAPIKey(ctx, apiKey.Key)
	if err != ErrAPIKeyInactive {
		t.Errorf("Expected ErrAPIKeyInactive, got %v", err)
	}
}
