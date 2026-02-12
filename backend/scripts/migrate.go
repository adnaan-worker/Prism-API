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
		databaseURL = "postgres://postgres:postgres@localhost:5432/prism_api?sslmode=disable"
	}

	// Connect to database
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("Running database migrations...")

	// Auto migrate all models in correct order (respecting foreign keys)
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

	// Index on pricings table
	db.Exec("CREATE INDEX IF NOT EXISTS idx_pricings_api_config_id ON pricings(api_config_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_pricings_is_active ON pricings(is_active) WHERE deleted_at IS NULL")

	// Index on billing_transactions table
	db.Exec("CREATE INDEX IF NOT EXISTS idx_billing_transactions_user_id ON billing_transactions(user_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_billing_transactions_request_log_id ON billing_transactions(request_log_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_billing_transactions_pricing_id ON billing_transactions(pricing_id)")

	// Index on request_caches table
	db.Exec("CREATE INDEX IF NOT EXISTS idx_request_caches_user_id ON request_caches(user_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_request_caches_model ON request_caches(model)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_request_caches_expires_at ON request_caches(expires_at)")

	fmt.Println("Migrations completed successfully!")
	fmt.Println("\nNext steps:")
	fmt.Println("1. Create an admin user via the registration API")
	fmt.Println("2. Update the user's is_admin field to true in the database")
	fmt.Println("3. Configure API providers in the admin panel")
}
