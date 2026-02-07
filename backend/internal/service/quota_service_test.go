package service

import (
	"api-aggregator/backend/internal/models"
	"api-aggregator/backend/internal/repository"
	"context"
	"os"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// getQuotaTestDB creates a test database connection
func getQuotaTestDB(t *testing.T) *gorm.DB {
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

// setupQuotaTestDB initializes the test database
func setupQuotaTestDB(t *testing.T) *gorm.DB {
	db := getQuotaTestDB(t)

	// Clean up existing data
	db.Exec("TRUNCATE TABLE sign_in_records CASCADE")
	db.Exec("TRUNCATE TABLE users CASCADE")

	// Run migrations
	err := db.AutoMigrate(&models.User{}, &models.SignInRecord{})
	if err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	return db
}

// cleanupQuotaTestDB cleans up the test database
func cleanupQuotaTestDB(t *testing.T, db *gorm.DB) {
	db.Exec("TRUNCATE TABLE sign_in_records CASCADE")
	db.Exec("TRUNCATE TABLE users CASCADE")
}

// Property 16: 签到额度增加
// Feature: api-aggregator, Property 16: For any user performing daily sign-in, the user's quota should increase by exactly 1000 tokens.
// Validates: Requirements 8.1
func TestProperty_SignInQuotaIncrease(t *testing.T) {
	db := setupQuotaTestDB(t)
	defer cleanupQuotaTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	signInRepo := repository.NewSignInRepository(db)
	quotaService := NewQuotaService(userRepo, signInRepo)

	properties := gopter.NewProperties(nil)

	// Generator for initial quota (0 to 100000)
	initialQuotaGen := gen.Int64Range(0, 100000)

	properties.Property("Sign-in increases quota by exactly 1000", prop.ForAll(
		func(initialQuota int64) bool {
			// Clean up before each test
			db.Exec("TRUNCATE TABLE sign_in_records CASCADE")
			db.Exec("TRUNCATE TABLE users CASCADE")

			ctx := context.Background()

			// Create a test user with random initial quota
			user := &models.User{
				Username:     "testuser",
				Email:        "test@example.com",
				PasswordHash: "hash",
				Quota:        initialQuota,
				UsedQuota:    0,
				Status:       "active",
			}
			if err := userRepo.Create(ctx, user); err != nil {
				return true // Skip if user creation fails
			}

			// Perform sign-in
			quotaAwarded, err := quotaService.SignIn(ctx, user.ID)
			if err != nil {
				return false // Sign-in should succeed
			}

			// Check that exactly 1000 tokens were awarded
			if quotaAwarded != 1000 {
				return false
			}

			// Verify user's quota increased by 1000
			updatedUser, err := userRepo.FindByID(ctx, user.ID)
			if err != nil {
				return false
			}

			return updatedUser.Quota == initialQuota+1000
		},
		initialQuotaGen,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Property 17: 重复签到限制
// Feature: api-aggregator, Property 17: For any user who has already signed in today, attempting to sign in again should return an error without increasing quota.
// Validates: Requirements 8.2
func TestProperty_DuplicateSignInRestriction(t *testing.T) {
	db := setupQuotaTestDB(t)
	defer cleanupQuotaTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	signInRepo := repository.NewSignInRepository(db)
	quotaService := NewQuotaService(userRepo, signInRepo)

	properties := gopter.NewProperties(nil)

	// Generator for initial quota (0 to 100000)
	initialQuotaGen := gen.Int64Range(0, 100000)

	properties.Property("Duplicate sign-in returns error without increasing quota", prop.ForAll(
		func(initialQuota int64) bool {
			// Clean up before each test
			db.Exec("TRUNCATE TABLE sign_in_records CASCADE")
			db.Exec("TRUNCATE TABLE users CASCADE")

			ctx := context.Background()

			// Create a test user
			user := &models.User{
				Username:     "testuser",
				Email:        "test@example.com",
				PasswordHash: "hash",
				Quota:        initialQuota,
				UsedQuota:    0,
				Status:       "active",
			}
			if err := userRepo.Create(ctx, user); err != nil {
				return true // Skip if user creation fails
			}

			// First sign-in should succeed
			_, err := quotaService.SignIn(ctx, user.ID)
			if err != nil {
				return false // First sign-in should succeed
			}

			// Get quota after first sign-in
			userAfterFirstSignIn, err := userRepo.FindByID(ctx, user.ID)
			if err != nil {
				return false
			}
			quotaAfterFirstSignIn := userAfterFirstSignIn.Quota

			// Second sign-in should fail
			_, err = quotaService.SignIn(ctx, user.ID)
			if err != ErrAlreadySignedIn {
				return false // Should return ErrAlreadySignedIn
			}

			// Verify quota didn't change
			userAfterSecondSignIn, err := userRepo.FindByID(ctx, user.ID)
			if err != nil {
				return false
			}

			return userAfterSecondSignIn.Quota == quotaAfterFirstSignIn
		},
		initialQuotaGen,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Unit test for GetQuotaInfo
func TestQuotaService_GetQuotaInfo(t *testing.T) {
	db := setupQuotaTestDB(t)
	defer cleanupQuotaTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	signInRepo := repository.NewSignInRepository(db)
	quotaService := NewQuotaService(userRepo, signInRepo)

	ctx := context.Background()

	// Create a test user
	user := &models.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hash",
		Quota:        10000,
		UsedQuota:    2000,
		Status:       "active",
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Get quota info
	quotaInfo, err := quotaService.GetQuotaInfo(ctx, user.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if quotaInfo.TotalQuota != 10000 {
		t.Errorf("Expected total quota 10000, got %d", quotaInfo.TotalQuota)
	}
	if quotaInfo.UsedQuota != 2000 {
		t.Errorf("Expected used quota 2000, got %d", quotaInfo.UsedQuota)
	}
	if quotaInfo.RemainingQuota != 8000 {
		t.Errorf("Expected remaining quota 8000, got %d", quotaInfo.RemainingQuota)
	}
}

// Unit test for SignIn
func TestQuotaService_SignIn(t *testing.T) {
	db := setupQuotaTestDB(t)
	defer cleanupQuotaTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	signInRepo := repository.NewSignInRepository(db)
	quotaService := NewQuotaService(userRepo, signInRepo)

	ctx := context.Background()

	// Create a test user
	user := &models.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hash",
		Quota:        10000,
		UsedQuota:    0,
		Status:       "active",
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// First sign-in should succeed
	quotaAwarded, err := quotaService.SignIn(ctx, user.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if quotaAwarded != 1000 {
		t.Errorf("Expected quota awarded 1000, got %d", quotaAwarded)
	}

	// Verify user quota increased
	updatedUser, err := userRepo.FindByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}
	if updatedUser.Quota != 11000 {
		t.Errorf("Expected quota 11000, got %d", updatedUser.Quota)
	}
	if updatedUser.LastSignIn == nil {
		t.Error("Expected last_sign_in to be set")
	}

	// Second sign-in should fail
	_, err = quotaService.SignIn(ctx, user.ID)
	if err != ErrAlreadySignedIn {
		t.Errorf("Expected ErrAlreadySignedIn, got %v", err)
	}

	// Verify quota didn't change
	updatedUser2, err := userRepo.FindByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}
	if updatedUser2.Quota != 11000 {
		t.Errorf("Expected quota to remain 11000, got %d", updatedUser2.Quota)
	}
}

// Unit test for DeductQuota
func TestQuotaService_DeductQuota(t *testing.T) {
	db := setupQuotaTestDB(t)
	defer cleanupQuotaTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	signInRepo := repository.NewSignInRepository(db)
	quotaService := NewQuotaService(userRepo, signInRepo)

	ctx := context.Background()

	// Create a test user
	user := &models.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hash",
		Quota:        10000,
		UsedQuota:    0,
		Status:       "active",
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Deduct quota
	err := quotaService.DeductQuota(ctx, user.ID, 1000)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify used quota increased
	updatedUser, err := userRepo.FindByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}
	if updatedUser.UsedQuota != 1000 {
		t.Errorf("Expected used quota 1000, got %d", updatedUser.UsedQuota)
	}

	// Try to deduct more than available
	err = quotaService.DeductQuota(ctx, user.ID, 10000)
	if err != ErrInsufficientQuota {
		t.Errorf("Expected ErrInsufficientQuota, got %v", err)
	}

	// Verify used quota didn't change
	updatedUser2, err := userRepo.FindByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}
	if updatedUser2.UsedQuota != 1000 {
		t.Errorf("Expected used quota to remain 1000, got %d", updatedUser2.UsedQuota)
	}
}

// Unit test for CheckQuota
func TestQuotaService_CheckQuota(t *testing.T) {
	db := setupQuotaTestDB(t)
	defer cleanupQuotaTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	signInRepo := repository.NewSignInRepository(db)
	quotaService := NewQuotaService(userRepo, signInRepo)

	ctx := context.Background()

	// Create a test user
	user := &models.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hash",
		Quota:        10000,
		UsedQuota:    2000,
		Status:       "active",
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Check sufficient quota
	hasSufficient, err := quotaService.CheckQuota(ctx, user.ID, 5000)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !hasSufficient {
		t.Error("Expected sufficient quota")
	}

	// Check insufficient quota
	hasInsufficient, err := quotaService.CheckQuota(ctx, user.ID, 10000)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if hasInsufficient {
		t.Error("Expected insufficient quota")
	}
}

// Test sign-in on different days
func TestQuotaService_SignInDifferentDays(t *testing.T) {
	db := setupQuotaTestDB(t)
	defer cleanupQuotaTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	signInRepo := repository.NewSignInRepository(db)
	quotaService := NewQuotaService(userRepo, signInRepo)

	ctx := context.Background()

	// Create a test user
	user := &models.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hash",
		Quota:        10000,
		UsedQuota:    0,
		Status:       "active",
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// First sign-in
	_, err := quotaService.SignIn(ctx, user.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Create a sign-in record from yesterday manually to simulate different day
	yesterday := time.Now().Add(-24 * time.Hour)
	oldRecord := &models.SignInRecord{
		UserID:       user.ID,
		QuotaAwarded: 1000,
	}
	db.Create(oldRecord)
	// Update the created_at to yesterday
	db.Model(oldRecord).Update("created_at", yesterday)

	// Clean today's record
	db.Exec("DELETE FROM sign_in_records WHERE user_id = ? AND created_at >= ?",
		user.ID, time.Now().Truncate(24*time.Hour))

	// Sign-in today should succeed
	quotaAwarded, err := quotaService.SignIn(ctx, user.ID)
	if err != nil {
		t.Fatalf("Expected no error for sign-in on different day, got %v", err)
	}
	if quotaAwarded != 1000 {
		t.Errorf("Expected quota awarded 1000, got %d", quotaAwarded)
	}
}
