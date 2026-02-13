package main

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// migrate.go - 数据库迁移脚本
// 用途：初始化数据库，创建所有表、索引、默认设置和管理员用户
// 使用方法：go run migrate.go
// 环境变量：
//   - DATABASE_URL: 数据库连接字符串（默认：postgres://postgres:postgres@localhost:5432/api_aggregator?sslmode=disable）
//   - ADMIN_USERNAME: 管理员用户名（默认：admin）
//   - ADMIN_EMAIL: 管理员邮箱（默认：admin@example.com）
//   - ADMIN_PASSWORD: 管理员密码（默认：admin123）

func main() {
	// 从环境变量读取数据库配置
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/api_aggregator?sslmode=disable"
	}

	fmt.Printf("🔄 Connecting to database...\n")
	fmt.Printf("   Database: %s\n", databaseURL)

	// 连接数据库
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	fmt.Println("✅ Connected to database")
	fmt.Println("\n🔄 Creating tables...")

	// ==================== 核心业务表 ====================

	// 创建 users 表 - 用户表
	// 对应模型：backend/internal/domain/user/model.go - User
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP,
			username VARCHAR(255) NOT NULL UNIQUE,
			email VARCHAR(255) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			quota BIGINT NOT NULL DEFAULT 10000,
			used_quota BIGINT NOT NULL DEFAULT 0,
			is_admin BOOLEAN NOT NULL DEFAULT false,
			status VARCHAR(50) NOT NULL DEFAULT 'active',
			last_sign_in TIMESTAMP
		)
	`).Error
	if err != nil {
		log.Fatalf("❌ Failed to create users table: %v", err)
	}
	fmt.Println("  ✓ users")

	// 创建 api_keys 表 - API密钥表
	// 对应模型：backend/internal/domain/apikey/model.go - APIKey
	// 外键关系：user_id -> users(id) ON DELETE CASCADE
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS api_keys (
			id SERIAL PRIMARY KEY,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			key VARCHAR(255) NOT NULL UNIQUE,
			name VARCHAR(255) NOT NULL,
			is_active BOOLEAN NOT NULL DEFAULT true,
			rate_limit INTEGER NOT NULL DEFAULT 60,
			last_used_at TIMESTAMP
		)
	`).Error
	if err != nil {
		log.Fatalf("❌ Failed to create api_keys table: %v", err)
	}
	fmt.Println("  ✓ api_keys")

	// 创建 api_configs 表 - API配置表
	// 对应模型：backend/internal/domain/apiconfig/model.go - APIConfig
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS api_configs (
			id SERIAL PRIMARY KEY,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP,
			name VARCHAR(255) NOT NULL,
			type VARCHAR(50) NOT NULL,
			base_url TEXT NOT NULL,
			api_key TEXT,
			models JSONB NOT NULL DEFAULT '[]',
			headers JSONB,
			metadata JSONB,
			is_active BOOLEAN NOT NULL DEFAULT true,
			priority INTEGER NOT NULL DEFAULT 100,
			weight INTEGER NOT NULL DEFAULT 1,
			max_rps INTEGER NOT NULL DEFAULT 0,
			timeout INTEGER NOT NULL DEFAULT 30
		)
	`).Error
	if err != nil {
		log.Fatalf("❌ Failed to create api_configs table: %v", err)
	}
	fmt.Println("  ✓ api_configs")

	// 创建 settings 表 - 系统设置表
	// 对应模型：backend/internal/domain/settings/model.go - Setting
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS settings (
			id SERIAL PRIMARY KEY,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP,
			"key" VARCHAR(255) NOT NULL UNIQUE,
			value TEXT,
			type VARCHAR(50) NOT NULL DEFAULT 'string',
			description TEXT,
			is_system BOOLEAN NOT NULL DEFAULT false
		)
	`).Error
	if err != nil {
		log.Fatalf("❌ Failed to create settings table: %v", err)
	}
	fmt.Println("  ✓ settings")

	// ==================== 配额和定价表 ====================

	// 创建 sign_in_records 表 - 签到记录表
	// 对应模型：backend/internal/domain/quota/model.go - SignInRecord
	// 外键关系：user_id -> users(id) ON DELETE CASCADE
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS sign_in_records (
			id SERIAL PRIMARY KEY,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			quota_awarded INTEGER NOT NULL
		)
	`).Error
	if err != nil {
		log.Fatalf("❌ Failed to create sign_in_records table: %v", err)
	}
	fmt.Println("  ✓ sign_in_records")

	// 创建 pricings 表 - 定价表
	// 对应模型：backend/internal/domain/pricing/model.go - Pricing
	// 外键关系：api_config_id -> api_configs(id) ON DELETE CASCADE
	// 唯一约束：(api_config_id, model_name) - 每个配置的每个模型只能有一个定价
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS pricings (
			id SERIAL PRIMARY KEY,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP,
			api_config_id INTEGER NOT NULL REFERENCES api_configs(id) ON DELETE CASCADE,
			model_name VARCHAR(255) NOT NULL,
			input_price DOUBLE PRECISION NOT NULL DEFAULT 0,
			output_price DOUBLE PRECISION NOT NULL DEFAULT 0,
			currency VARCHAR(20) NOT NULL DEFAULT 'credits',
			unit INTEGER NOT NULL DEFAULT 1000,
			is_active BOOLEAN NOT NULL DEFAULT true,
			description VARCHAR(500),
			UNIQUE(api_config_id, model_name)
		)
	`).Error
	if err != nil {
		log.Fatalf("❌ Failed to create pricings table: %v", err)
	}
	fmt.Println("  ✓ pricings")

	// ==================== 日志和缓存表 ====================

	// 创建 request_logs 表 - 请求日志表
	// 对应模型：backend/internal/domain/log/model.go - RequestLog
	// 外键关系：user_id -> users(id) ON DELETE CASCADE
	// 注意：api_key_id 和 api_config_id 不设置外键，因为日志需要保留历史记录
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS request_logs (
			id SERIAL PRIMARY KEY,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			api_key_id INTEGER NOT NULL,
			api_config_id INTEGER NOT NULL,
			model VARCHAR(255) NOT NULL,
			method VARCHAR(10) NOT NULL,
			path TEXT NOT NULL,
			status_code INTEGER NOT NULL,
			response_time INTEGER NOT NULL,
			tokens_used INTEGER NOT NULL DEFAULT 0,
			error_msg TEXT
		)
	`).Error
	if err != nil {
		log.Fatalf("❌ Failed to create request_logs table: %v", err)
	}
	fmt.Println("  ✓ request_logs")

	// 创建 request_caches 表 - 请求缓存表
	// 对应模型：backend/internal/domain/cache/model.go - RequestCache
	// 外键关系：user_id -> users(id) ON DELETE CASCADE
	// 唯一约束：cache_key - 缓存键必须唯一
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS request_caches (
			id SERIAL PRIMARY KEY,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			cache_key VARCHAR(32) NOT NULL UNIQUE,
			query_text TEXT,
			embedding TEXT,
			model VARCHAR(100) NOT NULL,
			request TEXT NOT NULL,
			response TEXT NOT NULL,
			tokens_saved INTEGER NOT NULL DEFAULT 0,
			hit_count INTEGER NOT NULL DEFAULT 0,
			expires_at TIMESTAMP NOT NULL
		)
	`).Error
	if err != nil {
		log.Fatalf("❌ Failed to create request_caches table: %v", err)
	}
	fmt.Println("  ✓ request_caches")

	// ==================== 负载均衡表 ====================

	// 创建 load_balancer_configs 表 - 负载均衡配置表
	// 对应模型：backend/internal/domain/loadbalancer/model.go - LoadBalancerConfig
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS load_balancer_configs (
			id SERIAL PRIMARY KEY,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP,
			model_name VARCHAR(255) NOT NULL,
			strategy VARCHAR(50) NOT NULL DEFAULT 'round_robin',
			is_active BOOLEAN NOT NULL DEFAULT true
		)
	`).Error
	if err != nil {
		log.Fatalf("❌ Failed to create load_balancer_configs table: %v", err)
	}
	fmt.Println("  ✓ load_balancer_configs")

	// ==================== 账号池表 ====================

	// 创建 account_pools 表 - 账号池表
	// 对应模型：backend/internal/domain/accountpool/model.go - AccountPool
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS account_pools (
			id SERIAL PRIMARY KEY,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			provider_type VARCHAR(50) NOT NULL,
			strategy VARCHAR(50) NOT NULL DEFAULT 'round_robin',
			health_check_interval INTEGER NOT NULL DEFAULT 300,
			health_check_timeout INTEGER NOT NULL DEFAULT 10,
			max_retries INTEGER NOT NULL DEFAULT 3,
			is_active BOOLEAN NOT NULL DEFAULT true,
			total_requests BIGINT NOT NULL DEFAULT 0,
			total_errors BIGINT NOT NULL DEFAULT 0
		)
	`).Error
	if err != nil {
		log.Fatalf("❌ Failed to create account_pools table: %v", err)
	}
	fmt.Println("  ✓ account_pools")

	// 创建 account_credentials 表 - 账号凭据表
	// 对应模型：backend/internal/domain/accountpool/model.go - AccountCredential
	// 外键关系：pool_id -> account_pools(id) ON DELETE CASCADE
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS account_credentials (
			id SERIAL PRIMARY KEY,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP,
			pool_id INTEGER NOT NULL REFERENCES account_pools(id) ON DELETE CASCADE,
			provider_type VARCHAR(50) NOT NULL,
			auth_type VARCHAR(50) NOT NULL DEFAULT 'api_key',
			api_key TEXT,
			access_token TEXT,
			refresh_token TEXT,
			session_token TEXT,
			expires_at TIMESTAMP,
			account_name VARCHAR(255),
			account_email VARCHAR(255),
			weight INTEGER NOT NULL DEFAULT 1,
			is_active BOOLEAN NOT NULL DEFAULT true,
			health_status VARCHAR(50) DEFAULT 'unknown',
			last_checked_at TIMESTAMP,
			last_used_at TIMESTAMP,
			total_requests BIGINT NOT NULL DEFAULT 0,
			total_errors BIGINT NOT NULL DEFAULT 0,
			rate_limit INTEGER NOT NULL DEFAULT 0,
			current_usage INTEGER NOT NULL DEFAULT 0,
			rate_limit_reset_at TIMESTAMP
		)
	`).Error
	if err != nil {
		log.Fatalf("❌ Failed to create account_credentials table: %v", err)
	}
	fmt.Println("  ✓ account_credentials")

	// 创建 account_pool_request_logs 表 - 账号池请求日志表
	// 对应模型：backend/internal/domain/accountpool/model.go - AccountPoolRequestLog
	// 外键关系：
	//   - credential_id -> account_credentials(id) ON DELETE SET NULL (保留日志)
	//   - pool_id -> account_pools(id) ON DELETE SET NULL (保留日志)
	//   - request_log_id -> request_logs(id) ON DELETE SET NULL (保留日志)
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS account_pool_request_logs (
			id SERIAL PRIMARY KEY,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			credential_id INTEGER REFERENCES account_credentials(id) ON DELETE SET NULL,
			pool_id INTEGER REFERENCES account_pools(id) ON DELETE SET NULL,
			provider_type VARCHAR(50) NOT NULL,
			model VARCHAR(255) NOT NULL,
			method VARCHAR(10) NOT NULL,
			status_code INTEGER,
			response_time INTEGER,
			tokens_used INTEGER,
			error_message TEXT,
			request_log_id INTEGER REFERENCES request_logs(id) ON DELETE SET NULL
		)
	`).Error
	if err != nil {
		log.Fatalf("❌ Failed to create account_pool_request_logs table: %v", err)
	}
	fmt.Println("  ✓ account_pool_request_logs")

	fmt.Println("✅ All tables created successfully")

	// 创建索引
	fmt.Println("\n🔄 Creating indexes...")
	createIndexes(db)
	fmt.Println("✅ All indexes created successfully")

	// 插入默认设置
	fmt.Println("\n🔄 Inserting default settings...")
	insertDefaultSettings(db)
	fmt.Println("✅ Default settings inserted successfully")

	// 创建管理员用户
	fmt.Println("\n🔄 Creating admin user...")
	createAdminUser(db)
	fmt.Println("✅ Admin user setup completed")

	fmt.Println("\n🎉 Database migration completed successfully!")
	fmt.Println("\n📋 Next steps:")
	fmt.Println("   1. Start the backend server: go run cmd/server/main.go")
	fmt.Println("   2. Login with admin credentials from .env file")
	fmt.Println("   3. Configure API providers in the admin panel")
}

// createIndexes 创建所有索引以优化查询性能
func createIndexes(db *gorm.DB) {
	indexes := []string{
		// ==================== users 表索引 ====================
		"CREATE INDEX IF NOT EXISTS idx_users_email ON users(email) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_users_username ON users(username) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_users_status ON users(status) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_users_is_admin ON users(is_admin) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at)",

		// ==================== api_keys 表索引 ====================
		"CREATE INDEX IF NOT EXISTS idx_api_keys_key ON api_keys(key) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_api_keys_user_id ON api_keys(user_id) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_api_keys_user_active ON api_keys(user_id, is_active) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_api_keys_deleted_at ON api_keys(deleted_at)",

		// ==================== api_configs 表索引 ====================
		"CREATE INDEX IF NOT EXISTS idx_api_configs_type ON api_configs(type) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_api_configs_is_active ON api_configs(is_active) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_api_configs_priority ON api_configs(priority DESC) WHERE deleted_at IS NULL AND is_active = true",
		"CREATE INDEX IF NOT EXISTS idx_api_configs_type_active ON api_configs(type, is_active) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_api_configs_deleted_at ON api_configs(deleted_at)",

		// ==================== settings 表索引 ====================
		"CREATE INDEX IF NOT EXISTS idx_settings_key ON settings(\"key\") WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_settings_is_system ON settings(is_system) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_settings_deleted_at ON settings(deleted_at)",

		// ==================== pricings 表索引 ====================
		"CREATE INDEX IF NOT EXISTS idx_pricings_api_config_id ON pricings(api_config_id) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_pricings_model_name ON pricings(model_name) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_pricings_is_active ON pricings(is_active) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_pricings_config_model_active ON pricings(api_config_id, model_name, is_active) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_pricings_deleted_at ON pricings(deleted_at)",

		// ==================== request_logs 表索引 ====================
		// 时间范围查询优化
		"CREATE INDEX IF NOT EXISTS idx_request_logs_created_at ON request_logs(created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_request_logs_user_id ON request_logs(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_request_logs_user_created ON request_logs(user_id, created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_request_logs_model ON request_logs(model)",
		"CREATE INDEX IF NOT EXISTS idx_request_logs_model_created ON request_logs(model, created_at DESC)",
		// 统计查询优化
		"CREATE INDEX IF NOT EXISTS idx_request_logs_status_code ON request_logs(status_code)",
		"CREATE INDEX IF NOT EXISTS idx_request_logs_api_key_id ON request_logs(api_key_id)",
		"CREATE INDEX IF NOT EXISTS idx_request_logs_api_config_id ON request_logs(api_config_id)",
		// 复合索引优化常见查询
		"CREATE INDEX IF NOT EXISTS idx_request_logs_user_model_created ON request_logs(user_id, model, created_at DESC)",

		// ==================== sign_in_records 表索引 ====================
		"CREATE INDEX IF NOT EXISTS idx_sign_in_records_user_id ON sign_in_records(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_sign_in_records_user_created ON sign_in_records(user_id, created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_sign_in_records_created_at ON sign_in_records(created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_sign_in_records_deleted_at ON sign_in_records(deleted_at)",

		// ==================== request_caches 表索引 ====================
		"CREATE INDEX IF NOT EXISTS idx_request_caches_user_id ON request_caches(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_request_caches_cache_key ON request_caches(cache_key)",
		"CREATE INDEX IF NOT EXISTS idx_request_caches_model ON request_caches(model)",
		"CREATE INDEX IF NOT EXISTS idx_request_caches_expires_at ON request_caches(expires_at)",
		// 清理过期缓存优化
		"CREATE INDEX IF NOT EXISTS idx_request_caches_expires_user ON request_caches(expires_at, user_id)",
		// 缓存命中统计优化
		"CREATE INDEX IF NOT EXISTS idx_request_caches_user_model ON request_caches(user_id, model)",

		// ==================== load_balancer_configs 表索引 ====================
		"CREATE INDEX IF NOT EXISTS idx_load_balancer_configs_model_name ON load_balancer_configs(model_name)",
		"CREATE INDEX IF NOT EXISTS idx_load_balancer_configs_is_active ON load_balancer_configs(is_active) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_load_balancer_configs_model_active ON load_balancer_configs(model_name, is_active) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_load_balancer_configs_deleted_at ON load_balancer_configs(deleted_at)",

		// ==================== account_pools 表索引 ====================
		"CREATE INDEX IF NOT EXISTS idx_account_pools_provider_type ON account_pools(provider_type)",
		"CREATE INDEX IF NOT EXISTS idx_account_pools_is_active ON account_pools(is_active) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_account_pools_provider_active ON account_pools(provider_type, is_active) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_account_pools_deleted_at ON account_pools(deleted_at)",

		// ==================== account_credentials 表索引 ====================
		"CREATE INDEX IF NOT EXISTS idx_account_credentials_pool_id ON account_credentials(pool_id)",
		"CREATE INDEX IF NOT EXISTS idx_account_credentials_provider_type ON account_credentials(provider_type)",
		"CREATE INDEX IF NOT EXISTS idx_account_credentials_is_active ON account_credentials(is_active) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_account_credentials_health_status ON account_credentials(health_status)",
		"CREATE INDEX IF NOT EXISTS idx_account_credentials_expires_at ON account_credentials(expires_at)",
		// 账号池选择优化
		"CREATE INDEX IF NOT EXISTS idx_account_credentials_pool_active_health ON account_credentials(pool_id, is_active, health_status) WHERE deleted_at IS NULL",
		// 过期检查优化
		"CREATE INDEX IF NOT EXISTS idx_account_credentials_expires_active ON account_credentials(expires_at, is_active) WHERE deleted_at IS NULL AND expires_at IS NOT NULL",
		"CREATE INDEX IF NOT EXISTS idx_account_credentials_deleted_at ON account_credentials(deleted_at)",

		// ==================== account_pool_request_logs 表索引 ====================
		"CREATE INDEX IF NOT EXISTS idx_account_pool_request_logs_credential_id ON account_pool_request_logs(credential_id)",
		"CREATE INDEX IF NOT EXISTS idx_account_pool_request_logs_pool_id ON account_pool_request_logs(pool_id)",
		"CREATE INDEX IF NOT EXISTS idx_account_pool_request_logs_created_at ON account_pool_request_logs(created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_account_pool_request_logs_request_log_id ON account_pool_request_logs(request_log_id)",
		// 统计查询优化
		"CREATE INDEX IF NOT EXISTS idx_account_pool_request_logs_pool_created ON account_pool_request_logs(pool_id, created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_account_pool_request_logs_credential_created ON account_pool_request_logs(credential_id, created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_account_pool_request_logs_status ON account_pool_request_logs(status_code)",
	}

	for _, idx := range indexes {
		if err := db.Exec(idx).Error; err != nil {
			log.Printf("  ⚠️  Warning: Failed to create index: %v", err)
		}
	}
}

// insertDefaultSettings 插入默认系统设置
func insertDefaultSettings(db *gorm.DB) {
	err := db.Exec(`
		INSERT INTO settings ("key", value, type, description, is_system, created_at, updated_at)
		VALUES
			-- 运行时配置
			('runtime.cache_enabled', 'true', 'bool', 'Enable request caching', true, NOW(), NOW()),
			('runtime.cache_ttl', '3600', 'int', 'Cache TTL in seconds', true, NOW(), NOW()),
			('runtime.max_retries', '3', 'int', 'Maximum retry attempts', true, NOW(), NOW()),
			('runtime.timeout', '30', 'int', 'Request timeout in seconds', true, NOW(), NOW()),
			('runtime.enable_load_balance', 'true', 'bool', 'Enable load balancing', true, NOW(), NOW()),
			
			-- 系统配置
			('system.site_name', 'Prism API', 'string', 'Site name', false, NOW(), NOW()),
			('system.site_description', 'AI API Aggregator', 'string', 'Site description', false, NOW(), NOW()),
			('system.admin_email', 'admin@example.com', 'string', 'Admin email', false, NOW(), NOW()),
			('system.maintenance_mode', 'false', 'bool', 'Maintenance mode', false, NOW(), NOW()),
			
			-- 默认配额
			('default_quota.daily', '1000', 'int', 'Default daily quota', false, NOW(), NOW()),
			('default_quota.monthly', '30000', 'int', 'Default monthly quota', false, NOW(), NOW()),
			('default_quota.total', '0', 'int', 'Default total quota (0 = unlimited)', false, NOW(), NOW()),
			
			-- 默认速率限制
			('default_rate_limit.per_minute', '60', 'int', 'Default rate limit per minute', false, NOW(), NOW()),
			('default_rate_limit.per_hour', '1000', 'int', 'Default rate limit per hour', false, NOW(), NOW()),
			('default_rate_limit.per_day', '10000', 'int', 'Default rate limit per day', false, NOW(), NOW())
		ON CONFLICT ("key") DO NOTHING
	`).Error
	
	if err != nil {
		log.Printf("  ⚠️  Warning: Failed to insert default settings: %v", err)
	}
}

// createAdminUser 创建管理员用户
func createAdminUser(db *gorm.DB) {
	// 从环境变量读取管理员信息
	adminUsername := os.Getenv("ADMIN_USERNAME")
	if adminUsername == "" {
		adminUsername = "admin"
	}

	adminEmail := os.Getenv("ADMIN_EMAIL")
	if adminEmail == "" {
		adminEmail = "admin@example.com"
	}

	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		adminPassword = "admin123"
	}

	// 检查管理员用户是否已存在
	var count int64
	db.Raw("SELECT COUNT(*) FROM users WHERE username = ?", adminUsername).Scan(&count)

	if count > 0 {
		fmt.Printf("  ⚠️  Admin user '%s' already exists, skipping creation\n", adminUsername)
		return
	}

	// 使用 bcrypt 生成密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("  ⚠️  Failed to hash password: %v", err)
		return
	}

	// 创建管理员用户
	err = db.Exec(`
		INSERT INTO users (username, email, password_hash, quota, used_quota, is_admin, status, created_at, updated_at)
		VALUES (?, ?, ?, 100000, 0, true, 'active', NOW(), NOW())
	`, adminUsername, adminEmail, string(hashedPassword)).Error
	
	if err != nil {
		log.Printf("  ⚠️  Failed to create admin user: %v", err)
		log.Println("  Please create admin user manually or register through the application")
		return
	}

	fmt.Printf("  ✓ Username: %s\n", adminUsername)
	fmt.Printf("  ✓ Email: %s\n", adminEmail)
	fmt.Printf("  ✓ Password: %s (from .env file)\n", adminPassword)
}
