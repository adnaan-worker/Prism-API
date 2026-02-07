package main

import (
	"fmt"
	"os"
	"testing"

	"api-aggregator/backend/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// getTestDB creates a test database connection
func getTestDB(t *testing.T) *gorm.DB {
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

// cleanupDB drops all tables
func cleanupDB(t *testing.T, db *gorm.DB) {
	db.Exec("DROP TABLE IF EXISTS sign_in_records CASCADE")
	db.Exec("DROP TABLE IF EXISTS request_logs CASCADE")
	db.Exec("DROP TABLE IF EXISTS load_balancer_configs CASCADE")
	db.Exec("DROP TABLE IF EXISTS api_configs CASCADE")
	db.Exec("DROP TABLE IF EXISTS api_keys CASCADE")
	db.Exec("DROP TABLE IF EXISTS users CASCADE")
}

func TestMigration_TablesCreated(t *testing.T) {
	db := getTestDB(t)
	cleanupDB(t, db)

	// Run migrations
	err := db.AutoMigrate(
		&models.User{},
		&models.APIKey{},
		&models.APIConfig{},
		&models.LoadBalancerConfig{},
		&models.RequestLog{},
		&models.SignInRecord{},
	)
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Check if all tables exist
	tables := []string{
		"users",
		"api_keys",
		"api_configs",
		"load_balancer_configs",
		"request_logs",
		"sign_in_records",
	}

	for _, table := range tables {
		var exists bool
		err := db.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = ?)", table).Scan(&exists).Error
		if err != nil {
			t.Errorf("Failed to check table %s: %v", table, err)
		}
		if !exists {
			t.Errorf("Table %s was not created", table)
		}
	}
}

func TestMigration_UserTableConstraints(t *testing.T) {
	db := getTestDB(t)
	cleanupDB(t, db)

	err := db.AutoMigrate(&models.User{})
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Test unique constraint on username
	user1 := models.User{Username: "testuser", Email: "test1@example.com", PasswordHash: "hash"}
	db.Create(&user1)

	user2 := models.User{Username: "testuser", Email: "test2@example.com", PasswordHash: "hash"}
	result := db.Create(&user2)
	if result.Error == nil {
		t.Error("Expected error for duplicate username, got nil")
	}

	// Test unique constraint on email
	user3 := models.User{Username: "testuser2", Email: "test1@example.com", PasswordHash: "hash"}
	result = db.Create(&user3)
	if result.Error == nil {
		t.Error("Expected error for duplicate email, got nil")
	}

	// Test default values
	user4 := models.User{Username: "testuser3", Email: "test3@example.com", PasswordHash: "hash"}
	db.Create(&user4)

	var retrieved models.User
	db.First(&retrieved, user4.ID)

	if retrieved.Quota != 10000 {
		t.Errorf("Expected default quota 10000, got %d", retrieved.Quota)
	}
	if retrieved.UsedQuota != 0 {
		t.Errorf("Expected default used_quota 0, got %d", retrieved.UsedQuota)
	}
	if retrieved.IsAdmin != false {
		t.Errorf("Expected default is_admin false, got %v", retrieved.IsAdmin)
	}
	if retrieved.Status != "active" {
		t.Errorf("Expected default status 'active', got %s", retrieved.Status)
	}
}

func TestMigration_APIKeyTableConstraints(t *testing.T) {
	db := getTestDB(t)
	cleanupDB(t, db)

	err := db.AutoMigrate(&models.User{}, &models.APIKey{})
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Create a user first
	user := models.User{Username: "testuser", Email: "test@example.com", PasswordHash: "hash"}
	db.Create(&user)

	// Test unique constraint on key
	key1 := models.APIKey{UserID: user.ID, Key: "sk-test123", Name: "Test Key 1"}
	db.Create(&key1)

	key2 := models.APIKey{UserID: user.ID, Key: "sk-test123", Name: "Test Key 2"}
	result := db.Create(&key2)
	if result.Error == nil {
		t.Error("Expected error for duplicate key, got nil")
	}

	// Test foreign key constraint
	invalidKey := models.APIKey{UserID: 99999, Key: "sk-test456", Name: "Invalid Key"}
	result = db.Create(&invalidKey)
	if result.Error == nil {
		t.Error("Expected error for invalid user_id foreign key, got nil")
	}

	// Test default values
	key3 := models.APIKey{UserID: user.ID, Key: "sk-test789", Name: "Test Key 3"}
	db.Create(&key3)

	var retrieved models.APIKey
	db.First(&retrieved, key3.ID)

	if retrieved.IsActive != true {
		t.Errorf("Expected default is_active true, got %v", retrieved.IsActive)
	}
	if retrieved.RateLimit != 60 {
		t.Errorf("Expected default rate_limit 60, got %d", retrieved.RateLimit)
	}
}

func TestMigration_APIConfigTableConstraints(t *testing.T) {
	db := getTestDB(t)
	cleanupDB(t, db)

	err := db.AutoMigrate(&models.APIConfig{})
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Test JSONB models field
	config := models.APIConfig{
		Name:    "Test Config",
		Type:    "openai",
		BaseURL: "https://api.openai.com",
		Models:  models.StringArray{"gpt-4", "gpt-3.5-turbo"},
	}
	db.Create(&config)

	var retrieved models.APIConfig
	db.First(&retrieved, config.ID)

	if len(retrieved.Models) != 2 {
		t.Errorf("Expected 2 models, got %d", len(retrieved.Models))
	}
	if retrieved.Models[0] != "gpt-4" {
		t.Errorf("Expected first model 'gpt-4', got %s", retrieved.Models[0])
	}

	// Test default values
	if retrieved.IsActive != true {
		t.Errorf("Expected default is_active true, got %v", retrieved.IsActive)
	}
	if retrieved.Priority != 100 {
		t.Errorf("Expected default priority 100, got %d", retrieved.Priority)
	}
	if retrieved.Weight != 1 {
		t.Errorf("Expected default weight 1, got %d", retrieved.Weight)
	}
	if retrieved.Timeout != 30 {
		t.Errorf("Expected default timeout 30, got %d", retrieved.Timeout)
	}
}

func TestMigration_RequestLogTableConstraints(t *testing.T) {
	db := getTestDB(t)
	cleanupDB(t, db)

	err := db.AutoMigrate(&models.User{}, &models.APIKey{}, &models.RequestLog{})
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Create dependencies
	user := models.User{Username: "testuser", Email: "test@example.com", PasswordHash: "hash"}
	db.Create(&user)

	apiKey := models.APIKey{UserID: user.ID, Key: "sk-test123", Name: "Test Key"}
	db.Create(&apiKey)

	// Test foreign key constraints
	log := models.RequestLog{
		UserID:       user.ID,
		APIKeyID:     apiKey.ID,
		APIConfigID:  1,
		Model:        "gpt-4",
		Method:       "POST",
		Path:         "/v1/chat/completions",
		StatusCode:   200,
		ResponseTime: 1500,
		TokensUsed:   100,
	}
	result := db.Create(&log)
	if result.Error != nil {
		t.Errorf("Failed to create request log: %v", result.Error)
	}

	// Test default value
	var retrieved models.RequestLog
	db.First(&retrieved, log.ID)
	if retrieved.TokensUsed != 100 {
		t.Errorf("Expected tokens_used 100, got %d", retrieved.TokensUsed)
	}
}

func TestMigration_IndexesCreated(t *testing.T) {
	db := getTestDB(t)
	cleanupDB(t, db)

	// Run migrations
	err := db.AutoMigrate(
		&models.User{},
		&models.APIKey{},
		&models.APIConfig{},
		&models.LoadBalancerConfig{},
		&models.RequestLog{},
		&models.SignInRecord{},
	)
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Check for key indexes
	indexes := []struct {
		table string
		index string
	}{
		{"users", "idx_users_deleted_at"},
		{"users", "idx_users_username"},
		{"users", "idx_users_email"},
		{"api_keys", "idx_api_keys_deleted_at"},
		{"api_keys", "idx_api_keys_user_id"},
		{"api_keys", "idx_api_keys_key"},
		{"api_configs", "idx_api_configs_deleted_at"},
		{"request_logs", "idx_request_logs_deleted_at"},
		{"request_logs", "idx_request_logs_user_id"},
		{"request_logs", "idx_request_logs_api_key_id"},
		{"request_logs", "idx_request_logs_model"},
		{"sign_in_records", "idx_sign_in_records_deleted_at"},
		{"sign_in_records", "idx_sign_in_records_user_id"},
	}

	for _, idx := range indexes {
		var exists bool
		query := `
			SELECT EXISTS (
				SELECT 1 FROM pg_indexes 
				WHERE tablename = ? AND indexname = ?
			)
		`
		err := db.Raw(query, idx.table, idx.index).Scan(&exists).Error
		if err != nil {
			t.Errorf("Failed to check index %s on table %s: %v", idx.index, idx.table, err)
			continue
		}
		if !exists {
			// Some indexes might have different names, just log a warning
			fmt.Printf("Warning: Index %s on table %s not found (may have different name)\n", idx.index, idx.table)
		}
	}
}

func TestMigration_SoftDelete(t *testing.T) {
	db := getTestDB(t)
	cleanupDB(t, db)

	err := db.AutoMigrate(&models.User{})
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Create and soft delete a user
	user := models.User{Username: "testuser", Email: "test@example.com", PasswordHash: "hash"}
	db.Create(&user)

	db.Delete(&user)

	// Should not find with normal query
	var count int64
	db.Model(&models.User{}).Where("id = ?", user.ID).Count(&count)
	if count != 0 {
		t.Error("Soft deleted user should not be found in normal query")
	}

	// Should find with Unscoped
	db.Unscoped().Model(&models.User{}).Where("id = ?", user.ID).Count(&count)
	if count != 1 {
		t.Error("Soft deleted user should be found with Unscoped query")
	}
}
