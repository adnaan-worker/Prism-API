package main

import (
	"api-aggregator/backend/internal/models"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// 从环境变量读取数据库配置
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=api_aggregator port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	}

	// 连接数据库
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("🔄 Starting database migration...")

	// 自动迁移所有模型
	err = db.AutoMigrate(
		// 核心表
		&models.User{},
		&models.APIKey{},
		&models.APIConfig{},
		&models.RequestLog{},
		&models.SignInRecord{},
		
		// 负载均衡和定价
		&models.LoadBalancerConfig{},
		&models.Pricing{},
		
		// 计费
		&models.BillingTransaction{},
		
		// 缓存
		&models.RequestCache{},
		
		// 账号池
		&models.AccountPool{},
		&models.AccountCredential{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	fmt.Println("✅ Core tables migrated successfully")

	// 创建账号池关联表（多对多关系）
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS account_pool_credentials (
			pool_id INTEGER NOT NULL REFERENCES account_pools(id) ON DELETE CASCADE,
			credential_id INTEGER NOT NULL REFERENCES account_credentials(id) ON DELETE CASCADE,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (pool_id, credential_id)
		)
	`).Error
	if err != nil {
		log.Fatal("Failed to create account_pool_credentials table:", err)
	}

	fmt.Println("✅ Account pool credentials table created")

	// 创建 Kiro 模型映射表
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS kiro_model_mappings (
			id SERIAL PRIMARY KEY,
			model_name VARCHAR(255) NOT NULL UNIQUE,
			kiro_model_id VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
	if err != nil {
		log.Fatal("Failed to create kiro_model_mappings table:", err)
	}

	fmt.Println("✅ Kiro model mappings table created")

	// 初始化 Kiro 模型映射
	err = initKiroModels(db)
	if err != nil {
		log.Printf("⚠️  Warning: Failed to initialize Kiro models: %v", err)
	} else {
		fmt.Println("✅ Kiro models initialized")
	}

	// 创建索引
	err = createIndexes(db)
	if err != nil {
		log.Printf("⚠️  Warning: Failed to create some indexes: %v", err)
	} else {
		fmt.Println("✅ Indexes created")
	}

	fmt.Println("🎉 Database migration completed successfully!")
}

// initKiroModels 初始化 Kiro 模型映射
func initKiroModels(db *gorm.DB) error {
	models := []struct {
		ModelName    string
		KiroModelID  string
	}{
		// Claude 4.5 系列
		{"claude-sonnet-4-5", "claude-sonnet-4.5"},
		{"claude-sonnet-4.5", "claude-sonnet-4.5"},
		{"claude-haiku-4-5", "claude-haiku-4.5"},
		{"claude-haiku-4.5", "claude-haiku-4.5"},
		{"claude-opus-4-5", "claude-opus-4.5"},
		{"claude-opus-4.5", "claude-opus-4.5"},
		
		// Claude 4 系列
		{"claude-sonnet-4", "claude-sonnet-4"},
		{"claude-sonnet-4-20250514", "claude-sonnet-4"},
		
		// Claude 3.5 系列（映射到 4.5）
		{"claude-3-5-sonnet", "claude-sonnet-4.5"},
		{"claude-3-5-sonnet-20241022", "claude-sonnet-4.5"},
		{"claude-3-5-sonnet-20240620", "claude-sonnet-4.5"},
		
		// Claude 3 系列
		{"claude-3-opus", "claude-sonnet-4.5"},
		{"claude-3-sonnet", "claude-sonnet-4"},
		{"claude-3-haiku", "claude-haiku-4.5"},
		
		// GPT 兼容映射
		{"gpt-4", "claude-sonnet-4.5"},
		{"gpt-4o", "claude-sonnet-4.5"},
		{"gpt-4-turbo", "claude-sonnet-4.5"},
		{"gpt-3.5-turbo", "claude-sonnet-4.5"},
	}

	for _, m := range models {
		err := db.Exec(`
			INSERT INTO kiro_model_mappings (model_name, kiro_model_id, created_at, updated_at)
			VALUES (?, ?, NOW(), NOW())
			ON CONFLICT (model_name) DO UPDATE SET
				kiro_model_id = EXCLUDED.kiro_model_id,
				updated_at = NOW()
		`, m.ModelName, m.KiroModelID).Error
		if err != nil {
			return fmt.Errorf("failed to insert model %s: %w", m.ModelName, err)
		}
	}

	return nil
}

// createIndexes 创建额外的索引以提升性能
func createIndexes(db *gorm.DB) error {
	indexes := []string{
		// Request logs 索引
		"CREATE INDEX IF NOT EXISTS idx_request_logs_created_at ON request_logs(created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_request_logs_user_created ON request_logs(user_id, created_at DESC)",
		
		// Billing transactions 索引
		"CREATE INDEX IF NOT EXISTS idx_billing_transactions_user_created ON billing_transactions(user_id, created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_billing_transactions_type ON billing_transactions(type)",
		
		// Request cache 索引
		"CREATE INDEX IF NOT EXISTS idx_request_caches_expires_at ON request_caches(expires_at)",
		"CREATE INDEX IF NOT EXISTS idx_request_caches_user_model ON request_caches(user_id, model)",
		
		// Account credentials 索引
		"CREATE INDEX IF NOT EXISTS idx_account_credentials_provider ON account_credentials(provider_type)",
		"CREATE INDEX IF NOT EXISTS idx_account_credentials_active ON account_credentials(is_active)",
		"CREATE INDEX IF NOT EXISTS idx_account_credentials_expires ON account_credentials(expires_at)",
		
		// Account pools 索引
		"CREATE INDEX IF NOT EXISTS idx_account_pools_provider ON account_pools(provider_type)",
		"CREATE INDEX IF NOT EXISTS idx_account_pools_active ON account_pools(is_active)",
	}

	for _, idx := range indexes {
		if err := db.Exec(idx).Error; err != nil {
			log.Printf("Warning: Failed to create index: %v", err)
		}
	}

	return nil
}
