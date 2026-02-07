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

func getLogTestDB(t *testing.T) *gorm.DB {
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

func setupLogTestDB(t *testing.T) *gorm.DB {
	db := getLogTestDB(t)

	// Clean up existing data (ignore errors if tables don't exist)
	db.Exec("TRUNCATE TABLE request_logs CASCADE")
	db.Exec("TRUNCATE TABLE api_keys CASCADE")
	db.Exec("TRUNCATE TABLE users CASCADE")

	// Schema is already created by schema.sql, no need to migrate

	return db
}

func cleanupLogTestDB(t *testing.T, db *gorm.DB) {
	db.Exec("TRUNCATE TABLE request_logs CASCADE")
	db.Exec("TRUNCATE TABLE api_keys CASCADE")
	db.Exec("TRUNCATE TABLE users CASCADE")
}

func TestGetLogs(t *testing.T) {
	db := setupLogTestDB(t)
	defer cleanupLogTestDB(t, db)

	requestLogRepo := repository.NewRequestLogRepository(db)
	logService := NewLogService(requestLogRepo)

	ctx := context.Background()

	// Create test users
	user1 := &models.User{
		Username:     "user1",
		Email:        "user1@test.com",
		PasswordHash: "hash",
		Status:       "active",
		Quota:        10000,
	}
	if err := db.Create(user1).Error; err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}

	user2 := &models.User{
		Username:     "user2",
		Email:        "user2@test.com",
		PasswordHash: "hash",
		Status:       "active",
		Quota:        10000,
	}
	if err := db.Create(user2).Error; err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}

	// Create API keys
	apiKey1 := &models.APIKey{
		UserID:    user1.ID,
		Key:       "sk-test-key-1",
		Name:      "Test Key 1",
		IsActive:  true,
		RateLimit: 60,
	}
	if err := db.Create(apiKey1).Error; err != nil {
		t.Fatalf("Failed to create apiKey1: %v", err)
	}

	apiKey2 := &models.APIKey{
		UserID:    user2.ID,
		Key:       "sk-test-key-2",
		Name:      "Test Key 2",
		IsActive:  true,
		RateLimit: 60,
	}
	if err := db.Create(apiKey2).Error; err != nil {
		t.Fatalf("Failed to create apiKey2: %v", err)
	}

	// Create test request logs
	log1 := &models.RequestLog{
		UserID:       user1.ID,
		APIKeyID:     apiKey1.ID,
		APIConfigID:  1,
		Model:        "gpt-4",
		Method:       "POST",
		Path:         "/v1/chat/completions",
		StatusCode:   200,
		ResponseTime: 1000,
		TokensUsed:   100,
		CreatedAt:    time.Now(),
	}
	if err := requestLogRepo.Create(ctx, log1); err != nil {
		t.Fatalf("Failed to create log1: %v", err)
	}

	log2 := &models.RequestLog{
		UserID:       user2.ID,
		APIKeyID:     apiKey2.ID,
		APIConfigID:  1,
		Model:        "claude-3-opus",
		Method:       "POST",
		Path:         "/v1/messages",
		StatusCode:   200,
		ResponseTime: 1500,
		TokensUsed:   150,
		CreatedAt:    time.Now(),
	}
	if err := requestLogRepo.Create(ctx, log2); err != nil {
		t.Fatalf("Failed to create log2: %v", err)
	}

	log3 := &models.RequestLog{
		UserID:       user1.ID,
		APIKeyID:     apiKey1.ID,
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
	if err := requestLogRepo.Create(ctx, log3); err != nil {
		t.Fatalf("Failed to create log3: %v", err)
	}

	// Test: Get all logs
	t.Run("GetAllLogs", func(t *testing.T) {
		req := &GetLogsRequest{
			Page:     1,
			PageSize: 10,
		}

		resp, err := logService.GetLogs(ctx, req)
		if err != nil {
			t.Fatalf("Failed to get logs: %v", err)
		}

		if resp.Total != 3 {
			t.Errorf("Expected 3 total logs, got %d", resp.Total)
		}

		if len(resp.Logs) != 3 {
			t.Errorf("Expected 3 logs, got %d", len(resp.Logs))
		}
	})

	// Test: Filter by user ID
	t.Run("FilterByUserID", func(t *testing.T) {
		userID := user1.ID
		req := &GetLogsRequest{
			UserID:   &userID,
			Page:     1,
			PageSize: 10,
		}

		resp, err := logService.GetLogs(ctx, req)
		if err != nil {
			t.Fatalf("Failed to get logs: %v", err)
		}

		if resp.Total != 2 {
			t.Errorf("Expected 2 logs for user 1, got %d", resp.Total)
		}

		for _, log := range resp.Logs {
			if log.UserID != user1.ID {
				t.Errorf("Expected user ID %d, got %d", user1.ID, log.UserID)
			}
		}
	})

	// Test: Filter by model
	t.Run("FilterByModel", func(t *testing.T) {
		req := &GetLogsRequest{
			Model:    "gpt-4",
			Page:     1,
			PageSize: 10,
		}

		resp, err := logService.GetLogs(ctx, req)
		if err != nil {
			t.Fatalf("Failed to get logs: %v", err)
		}

		if resp.Total != 2 {
			t.Errorf("Expected 2 logs for gpt-4, got %d", resp.Total)
		}

		for _, log := range resp.Logs {
			if log.Model != "gpt-4" {
				t.Errorf("Expected model gpt-4, got %s", log.Model)
			}
		}
	})

	// Test: Filter by status code
	t.Run("FilterByStatusCode", func(t *testing.T) {
		statusCode := 200
		req := &GetLogsRequest{
			StatusCode: &statusCode,
			Page:       1,
			PageSize:   10,
		}

		resp, err := logService.GetLogs(ctx, req)
		if err != nil {
			t.Fatalf("Failed to get logs: %v", err)
		}

		if resp.Total != 2 {
			t.Errorf("Expected 2 logs with status 200, got %d", resp.Total)
		}

		for _, log := range resp.Logs {
			if log.StatusCode != 200 {
				t.Errorf("Expected status code 200, got %d", log.StatusCode)
			}
		}
	})

	// Test: Pagination
	t.Run("Pagination", func(t *testing.T) {
		req := &GetLogsRequest{
			Page:     1,
			PageSize: 2,
		}

		resp, err := logService.GetLogs(ctx, req)
		if err != nil {
			t.Fatalf("Failed to get logs: %v", err)
		}

		if resp.Total != 3 {
			t.Errorf("Expected 3 total logs, got %d", resp.Total)
		}

		if len(resp.Logs) != 2 {
			t.Errorf("Expected 2 logs on page 1, got %d", len(resp.Logs))
		}

		// Get page 2
		req.Page = 2
		resp, err = logService.GetLogs(ctx, req)
		if err != nil {
			t.Fatalf("Failed to get logs page 2: %v", err)
		}

		if len(resp.Logs) != 1 {
			t.Errorf("Expected 1 log on page 2, got %d", len(resp.Logs))
		}
	})
}
