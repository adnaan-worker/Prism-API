package service

import (
	"api-aggregator/backend/internal/models"
	"api-aggregator/backend/internal/repository"
	"context"
	"os"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func getStatsTestDB(t *testing.T) *gorm.DB {
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

func setupStatsTestDB(t *testing.T) *gorm.DB {
	db := getStatsTestDB(t)

	// Clean up existing data (ignore errors if tables don't exist)
	db.Exec("TRUNCATE TABLE request_logs CASCADE")
	db.Exec("TRUNCATE TABLE api_keys CASCADE")
	db.Exec("TRUNCATE TABLE users CASCADE")

	// Schema is already created by schema.sql, no need to migrate

	return db
}

func cleanupStatsTestDB(t *testing.T, db *gorm.DB) {
	db.Exec("TRUNCATE TABLE request_logs CASCADE")
	db.Exec("TRUNCATE TABLE api_keys CASCADE")
	db.Exec("TRUNCATE TABLE users CASCADE")
}

func TestGetStatsOverview(t *testing.T) {
	db := setupStatsTestDB(t)
	defer cleanupStatsTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	requestLogRepo := repository.NewRequestLogRepository(db)
	statsService := NewStatsService(userRepo, requestLogRepo)

	ctx := context.Background()

	// Create test users
	activeUser := &models.User{
		Username:     "active_user",
		Email:        "active@test.com",
		PasswordHash: "hash",
		Status:       "active",
		Quota:        10000,
	}
	if err := userRepo.Create(ctx, activeUser); err != nil {
		t.Fatalf("Failed to create active user: %v", err)
	}

	inactiveUser := &models.User{
		Username:     "inactive_user",
		Email:        "inactive@test.com",
		PasswordHash: "hash",
		Status:       "inactive",
		Quota:        10000,
	}
	if err := userRepo.Create(ctx, inactiveUser); err != nil {
		t.Fatalf("Failed to create inactive user: %v", err)
	}

	// Create test request logs
	todayStart := time.Now().Truncate(24 * time.Hour)
	yesterday := todayStart.Add(-24 * time.Hour)

	// Create an API key for the active user
	apiKey := &models.APIKey{
		UserID:    activeUser.ID,
		Key:       "sk-test-key",
		Name:      "Test Key",
		IsActive:  true,
		RateLimit: 60,
	}
	if err := db.Create(apiKey).Error; err != nil {
		t.Fatalf("Failed to create API key: %v", err)
	}

	// Today's successful request
	successLog := &models.RequestLog{
		UserID:       activeUser.ID,
		APIKeyID:     apiKey.ID,
		APIConfigID:  1,
		Model:        "gpt-4",
		Method:       "POST",
		Path:         "/v1/chat/completions",
		StatusCode:   200,
		ResponseTime: 1000,
		TokensUsed:   100,
		CreatedAt:    time.Now(),
	}
	if err := requestLogRepo.Create(ctx, successLog); err != nil {
		t.Fatalf("Failed to create success log: %v", err)
	}

	// Today's failed request
	failedLog := &models.RequestLog{
		UserID:       activeUser.ID,
		APIKeyID:     apiKey.ID,
		APIConfigID:  1,
		Model:        "gpt-4",
		Method:       "POST",
		Path:         "/v1/chat/completions",
		StatusCode:   500,
		ResponseTime: 500,
		TokensUsed:   0,
		ErrorMsg:     "Internal error",
		CreatedAt:    time.Now(),
	}
	if err := requestLogRepo.Create(ctx, failedLog); err != nil {
		t.Fatalf("Failed to create failed log: %v", err)
	}

	// Yesterday's request
	oldLog := &models.RequestLog{
		UserID:       activeUser.ID,
		APIKeyID:     apiKey.ID,
		APIConfigID:  1,
		Model:        "gpt-4",
		Method:       "POST",
		Path:         "/v1/chat/completions",
		StatusCode:   200,
		ResponseTime: 1000,
		TokensUsed:   100,
		CreatedAt:    yesterday,
	}
	if err := requestLogRepo.Create(ctx, oldLog); err != nil {
		t.Fatalf("Failed to create old log: %v", err)
	}

	// Get stats overview
	stats, err := statsService.GetStatsOverview(ctx)
	if err != nil {
		t.Fatalf("Failed to get stats overview: %v", err)
	}

	// Verify stats
	if stats.TotalUsers != 2 {
		t.Errorf("Expected 2 total users, got %d", stats.TotalUsers)
	}

	if stats.ActiveUsers != 1 {
		t.Errorf("Expected 1 active user, got %d", stats.ActiveUsers)
	}

	if stats.TotalRequests != 3 {
		t.Errorf("Expected 3 total requests, got %d", stats.TotalRequests)
	}

	if stats.TodayRequests != 2 {
		t.Errorf("Expected 2 today requests, got %d", stats.TodayRequests)
	}

	if stats.SuccessRequests != 2 {
		t.Errorf("Expected 2 success requests, got %d", stats.SuccessRequests)
	}

	if stats.FailedRequests != 1 {
		t.Errorf("Expected 1 failed request, got %d", stats.FailedRequests)
	}
}
