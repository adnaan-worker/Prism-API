package middleware

import (
	"api-aggregator/backend/internal/domain/apikey"
	"api-aggregator/backend/internal/domain/auth"
	"api-aggregator/backend/internal/domain/user"
	"api-aggregator/backend/pkg/cache"
	"api-aggregator/backend/pkg/logger"
	"time"
)

// Manager 中间件管理器
type Manager struct {
	// 认证相关
	Auth     *Auth
	APIKey   *APIKey
	Admin    *Admin
	
	// 限流相关
	RateLimit *RateLimit
	
	// 通用中间件
	CORS      *CORS
	Logger    *Logger
	Recovery  *Recovery
	RequestID *RequestID
	Timeout   *Timeout
}

// Config 中间件配置
type Config struct {
	// 服务依赖
	AuthService   auth.Service
	APIKeyService apikey.Service
	UserService   user.Service
	Cache         cache.Cache
	Logger        *logger.Logger
	
	// CORS配置
	CORSConfig *CORSConfig
	
	// 超时配置
	RequestTimeout time.Duration
}

// NewManager 创建中间件管理器实例
func NewManager(config *Config) *Manager {
	// 设置默认值
	if config.RequestTimeout == 0 {
		config.RequestTimeout = 30 * time.Second
	}
	
	return &Manager{
		// 认证相关
		Auth:   NewAuth(config.AuthService),
		APIKey: NewAPIKey(config.APIKeyService),
		Admin:  NewAdmin(config.UserService),
		
		// 限流相关
		RateLimit: NewRateLimit(config.Cache),
		
		// 通用中间件
		CORS:      NewCORS(config.CORSConfig),
		Logger:    NewLogger(config.Logger),
		Recovery:  NewRecovery(config.Logger),
		RequestID: NewRequestID(),
		Timeout:   NewTimeout(config.RequestTimeout),
	}
}
