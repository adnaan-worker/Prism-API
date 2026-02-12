package main

import (
	"fmt"
	"log"
	"os"

	"api-aggregator/backend/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Get database URL from environment variable
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "host=localhost user=postgres password=postgres dbname=api_aggregator port=5432 sslmode=disable"
	}

	// Connect to database
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("Running database migrations...")

	// Auto migrate all models
	err = db.AutoMigrate(
		&models.User{},
		&models.APIKey{},
		&models.APIConfig{},
		&models.LoadBalancerConfig{},
		&models.RequestLog{},
		&models.SignInRecord{},
		&models.Pricing{},
		&models.BillingTransaction{},
		&models.RequestCache{},
	)
	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Create additional indexes for performance optimization
	fmt.Println("Creating additional indexes...")

	// Index on users table
	db.Exec("CREATE INDEX IF NOT EXISTS idx_users_status ON users(status) WHERE deleted_at IS NULL")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_users_is_admin ON users(is_admin) WHERE deleted_at IS NULL")

	// Index on api_keys table
	db.Exec("CREATE INDEX IF NOT EXISTS idx_api_keys_user_active ON api_keys(user_id, is_active) WHERE deleted_at IS NULL")

	// Index on api_configs table
	db.Exec("CREATE INDEX IF NOT EXISTS idx_api_configs_type_active ON api_configs(type, is_active) WHERE deleted_at IS NULL")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_api_configs_priority ON api_configs(priority DESC) WHERE deleted_at IS NULL AND is_active = true")

	// Index on request_logs table for analytics queries
	db.Exec("CREATE INDEX IF NOT EXISTS idx_request_logs_created_at ON request_logs(created_at DESC)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_request_logs_user_created ON request_logs(user_id, created_at DESC)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_request_logs_model_created ON request_logs(model, created_at DESC)")

	// Index on sign_in_records for daily check
	db.Exec("CREATE INDEX IF NOT EXISTS idx_sign_in_records_user_created ON sign_in_records(user_id, created_at DESC)")

	fmt.Println("Migrations completed successfully!")
}
