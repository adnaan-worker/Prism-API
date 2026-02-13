package router

import (
	"api-aggregator/backend/internal/domain/accountpool"
	"api-aggregator/backend/internal/domain/apiconfig"
	"api-aggregator/backend/internal/domain/apikey"
	"api-aggregator/backend/internal/domain/auth"
	"api-aggregator/backend/internal/domain/cache"
	"api-aggregator/backend/internal/domain/loadbalancer"
	"api-aggregator/backend/internal/domain/log"
	"api-aggregator/backend/internal/domain/pricing"
	"api-aggregator/backend/internal/domain/proxy"
	"api-aggregator/backend/internal/domain/quota"
	"api-aggregator/backend/internal/domain/settings"
	"api-aggregator/backend/internal/domain/stats"
	"api-aggregator/backend/internal/domain/user"
	"api-aggregator/backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

// Router 路由管理器
type Router struct {
	engine *gin.Engine
	mw     *middleware.Manager
	
	// 处理器
	authHandler          *auth.Handler
	apiKeyHandler        *apikey.Handler
	apiConfigHandler     *apiconfig.Handler
	userHandler          *user.Handler
	quotaHandler         *quota.Handler
	statsHandler         *stats.Handler
	logHandler           *log.Handler
	pricingHandler       *pricing.Handler
	cacheHandler         *cache.Handler
	loadBalancerHandler  *loadbalancer.Handler
	accountPoolHandler   *accountpool.Handler
	settingsHandler      *settings.Handler
	proxyHandler         *proxy.Handler
}

// Config 路由配置
type Config struct {
	Engine *gin.Engine
	MW     *middleware.Manager
	
	// 处理器
	AuthHandler          *auth.Handler
	APIKeyHandler        *apikey.Handler
	APIConfigHandler     *apiconfig.Handler
	UserHandler          *user.Handler
	QuotaHandler         *quota.Handler
	StatsHandler         *stats.Handler
	LogHandler           *log.Handler
	PricingHandler       *pricing.Handler
	CacheHandler         *cache.Handler
	LoadBalancerHandler  *loadbalancer.Handler
	AccountPoolHandler   *accountpool.Handler
	SettingsHandler      *settings.Handler
	ProxyHandler         *proxy.Handler
}

// New 创建路由管理器实例
func New(config *Config) *Router {
	return &Router{
		engine:               config.Engine,
		mw:                   config.MW,
		authHandler:          config.AuthHandler,
		apiKeyHandler:        config.APIKeyHandler,
		apiConfigHandler:     config.APIConfigHandler,
		userHandler:          config.UserHandler,
		quotaHandler:         config.QuotaHandler,
		statsHandler:         config.StatsHandler,
		logHandler:           config.LogHandler,
		pricingHandler:       config.PricingHandler,
		cacheHandler:         config.CacheHandler,
		loadBalancerHandler:  config.LoadBalancerHandler,
		accountPoolHandler:   config.AccountPoolHandler,
		settingsHandler:      config.SettingsHandler,
		proxyHandler:         config.ProxyHandler,
	}
}

// Setup 设置所有路由
func (r *Router) Setup() {
	// 全局中间件
	r.engine.Use(r.mw.Recovery.Handle())
	r.engine.Use(r.mw.RequestID.Handle())
	r.engine.Use(r.mw.Logger.Handle())
	r.engine.Use(r.mw.CORS.Handle())

	// 健康检查
	r.engine.GET("/health", r.healthCheck)

	// 设置各模块路由
	r.setupAuthRoutes()
	r.setupUserRoutes()
	r.setupAPIKeyRoutes()
	r.setupProxyRoutes()
	r.setupAdminRoutes()
}

// healthCheck 健康检查
func (r *Router) healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok"})
}

// setupAuthRoutes 设置认证路由
func (r *Router) setupAuthRoutes() {
	auth := r.engine.Group("/api/v1/auth")
	{
		auth.POST("/register", r.authHandler.Register)
		auth.POST("/login", r.authHandler.Login)
		
		// 需要认证的路由
		authProtected := auth.Group("")
		authProtected.Use(r.mw.Auth.Handle())
		{
			authProtected.GET("/profile", r.authHandler.GetProfile)
			authProtected.POST("/change-password", r.authHandler.ChangePassword)
		}
	}
}

// setupUserRoutes 设置用户路由
func (r *Router) setupUserRoutes() {
	user := r.engine.Group("/api/v1/user")
	user.Use(r.mw.Auth.Handle())
	{
		// 配额相关
		user.GET("/quota", r.quotaHandler.GetQuotaInfo)
		user.POST("/signin", r.quotaHandler.SignIn)
		user.GET("/usage-history", r.quotaHandler.GetUsageHistory)
		
		// 缓存统计
		user.GET("/cache/stats", r.cacheHandler.GetCacheStats)
	}
	
	// 模型列表（不需要认证，但需要在 /api/v1 下）
	models := r.engine.Group("/api/v1")
	models.Use(r.mw.Auth.Handle())
	{
		models.GET("/models", r.apiConfigHandler.GetAvailableModels)
	}
}

// setupAPIKeyRoutes 设置API密钥路由
func (r *Router) setupAPIKeyRoutes() {
	apikeys := r.engine.Group("/api/v1/apikeys")
	apikeys.Use(r.mw.Auth.Handle())
	{
		apikeys.GET("", r.apiKeyHandler.GetAPIKeys)
		apikeys.POST("", r.apiKeyHandler.CreateAPIKey)
		apikeys.GET("/:id", r.apiKeyHandler.GetAPIKeyByID)
		apikeys.PUT("/:id", r.apiKeyHandler.UpdateAPIKey)
		apikeys.DELETE("/:id", r.apiKeyHandler.DeleteAPIKey)
	}
}

// setupAdminRoutes 设置管理员路由
func (r *Router) setupAdminRoutes() {
	admin := r.engine.Group("/api/v1/admin")
	admin.Use(r.mw.Auth.Handle())
	admin.Use(r.mw.Admin.Handle())
	{
		// 用户管理
		r.setupAdminUserRoutes(admin)
		
		// API配置管理
		r.setupAdminAPIConfigRoutes(admin)
		
		// 统计和日志
		r.setupAdminStatsRoutes(admin)
		r.setupAdminLogRoutes(admin)
		
		// 定价管理
		r.setupAdminPricingRoutes(admin)
		
		// 缓存管理
		r.setupAdminCacheRoutes(admin)
		
		// 负载均衡管理
		r.setupAdminLoadBalancerRoutes(admin)
		
		// 账号池管理
		r.setupAdminAccountPoolRoutes(admin)
		
		// 系统设置
		r.setupAdminSettingsRoutes(admin)
	}
}

// setupAdminUserRoutes 设置管理员用户路由
func (r *Router) setupAdminUserRoutes(group *gin.RouterGroup) {
	users := group.Group("/users")
	{
		users.GET("", r.userHandler.GetUsers)
		users.GET("/:id", r.userHandler.GetUserByID)
		users.PUT("/:id/status", r.userHandler.UpdateUserStatus)
		users.PUT("/:id/quota", r.userHandler.UpdateUserQuota)
		users.DELETE("/:id", r.userHandler.DeleteUser)
	}
}

// setupAdminAPIConfigRoutes 设置管理员API配置路由
func (r *Router) setupAdminAPIConfigRoutes(group *gin.RouterGroup) {
	configs := group.Group("/api-configs")
	{
		configs.GET("", r.apiConfigHandler.GetConfigs)
		configs.POST("", r.apiConfigHandler.CreateConfig)
		configs.GET("/:id", r.apiConfigHandler.GetConfig)
		configs.PUT("/:id", r.apiConfigHandler.UpdateConfig)
		configs.DELETE("/:id", r.apiConfigHandler.DeleteConfig)
		configs.POST("/:id/activate", r.apiConfigHandler.ActivateConfig)
		configs.POST("/:id/deactivate", r.apiConfigHandler.DeactivateConfig)
		
		// 批量操作
		configs.POST("/batch/delete", r.apiConfigHandler.BatchDeleteConfigs)
		configs.POST("/batch/activate", r.apiConfigHandler.BatchActivateConfigs)
		configs.POST("/batch/deactivate", r.apiConfigHandler.BatchDeactivateConfigs)
	}
	
	// 提供商相关
	providers := group.Group("/providers")
	{
		providers.POST("/fetch-models", r.apiConfigHandler.FetchModels)
	}
}

// setupAdminStatsRoutes 设置管理员统计路由
func (r *Router) setupAdminStatsRoutes(group *gin.RouterGroup) {
	stats := group.Group("/stats")
	{
		stats.GET("/overview", r.statsHandler.GetStatsOverview)
		stats.GET("/trend", r.statsHandler.GetRequestTrend)
		stats.GET("/models", r.statsHandler.GetModelUsage)
		stats.GET("/users", r.statsHandler.GetUserGrowth)
		stats.GET("/tokens", r.statsHandler.GetTokenUsage)
	}
}

// setupAdminLogRoutes 设置管理员日志路由
func (r *Router) setupAdminLogRoutes(group *gin.RouterGroup) {
	logs := group.Group("/logs")
	{
		logs.GET("", r.logHandler.GetLogs)
		logs.GET("/export", r.logHandler.ExportLogs)
		logs.GET("/stats", r.logHandler.GetLogStats)
		logs.DELETE("/cleanup", r.logHandler.DeleteOldLogs)
	}
}

// setupAdminPricingRoutes 设置管理员定价路由
func (r *Router) setupAdminPricingRoutes(group *gin.RouterGroup) {
	pricings := group.Group("/pricings")
	{
		pricings.GET("", r.pricingHandler.GetPricings)
		pricings.POST("", r.pricingHandler.CreatePricing)
		pricings.POST("/batch", r.pricingHandler.BatchCreatePricings)
		pricings.GET("/:id", r.pricingHandler.GetPricing)
		pricings.PUT("/:id", r.pricingHandler.UpdatePricing)
		pricings.DELETE("/:id", r.pricingHandler.DeletePricing)
		pricings.POST("/calculate", r.pricingHandler.CalculateCost)
	}
}

// setupAdminCacheRoutes 设置管理员缓存路由
func (r *Router) setupAdminCacheRoutes(group *gin.RouterGroup) {
	cache := group.Group("/cache")
	{
		cache.GET("/stats", r.cacheHandler.GetCacheStats)
		cache.GET("/list", r.cacheHandler.GetCacheList)
		cache.DELETE("/clean", r.cacheHandler.CleanExpiredCache)
		cache.DELETE("/:id", r.cacheHandler.DeleteCache)
		cache.DELETE("/user/:id", r.cacheHandler.ClearUserCache)
	}
}

// setupAdminLoadBalancerRoutes 设置管理员负载均衡路由
func (r *Router) setupAdminLoadBalancerRoutes(group *gin.RouterGroup) {
	lb := group.Group("/load-balancer")
	{
		lb.GET("/configs", r.loadBalancerHandler.ListConfigs)
		lb.POST("/configs", r.loadBalancerHandler.CreateConfig)
		lb.GET("/configs/:id", r.loadBalancerHandler.GetConfig)
		lb.PUT("/configs/:id", r.loadBalancerHandler.UpdateConfig)
		lb.DELETE("/configs/:id", r.loadBalancerHandler.DeleteConfig)
		lb.POST("/configs/:id/activate", r.loadBalancerHandler.ActivateConfig)
		lb.POST("/configs/:id/deactivate", r.loadBalancerHandler.DeactivateConfig)
		lb.GET("/models", r.loadBalancerHandler.GetAvailableModels)
		lb.GET("/models/:model/config", r.loadBalancerHandler.GetConfigByModel)
		lb.GET("/models/:model/endpoints", r.loadBalancerHandler.GetModelEndpoints)
	}
}

// setupAdminAccountPoolRoutes 设置管理员账号池路由
func (r *Router) setupAdminAccountPoolRoutes(group *gin.RouterGroup) {
	pools := group.Group("/account-pools")
	{
		pools.GET("", r.accountPoolHandler.ListPools)
		pools.POST("", r.accountPoolHandler.CreatePool)
		pools.GET("/:id", r.accountPoolHandler.GetPool)
		pools.PUT("/:id", r.accountPoolHandler.UpdatePool)
		pools.DELETE("/:id", r.accountPoolHandler.DeletePool)
		pools.PUT("/:id/status", r.accountPoolHandler.UpdatePoolStatus)
		pools.GET("/:id/stats", r.accountPoolHandler.GetPoolStats)
		
		// 凭据管理
		pools.GET("/credentials", r.accountPoolHandler.ListCredentials)
		pools.POST("/credentials", r.accountPoolHandler.CreateCredential)
		pools.GET("/credentials/:id", r.accountPoolHandler.GetCredential)
		pools.PUT("/credentials/:id", r.accountPoolHandler.UpdateCredential)
		pools.DELETE("/credentials/:id", r.accountPoolHandler.DeleteCredential)
		pools.PUT("/credentials/:id/status", r.accountPoolHandler.UpdateCredentialStatus)
		pools.POST("/credentials/:id/refresh", r.accountPoolHandler.RefreshCredential)
		
		// 批量导入
		pools.POST("/batch-import", r.accountPoolHandler.BatchImportCredentials)
		pools.POST("/batch-import-json", r.accountPoolHandler.BatchImportCredentialsFromJSON)
		
		// 请求日志
		pools.GET("/request-logs", r.accountPoolHandler.ListRequestLogs)
	}
}

// setupAdminSettingsRoutes 设置管理员设置路由
func (r *Router) setupAdminSettingsRoutes(group *gin.RouterGroup) {
	settings := group.Group("/settings")
	{
		// 运行时配置
		settings.GET("/runtime", r.settingsHandler.GetRuntimeConfig)
		settings.PUT("/runtime", r.settingsHandler.UpdateRuntimeConfig)
		
		// 系统配置
		settings.GET("/system", r.settingsHandler.GetSystemConfig)
		
		// 密码管理
		settings.PUT("/password", r.settingsHandler.UpdatePassword)
		
		// 默认配额
		settings.GET("/default-quota", r.settingsHandler.GetDefaultQuota)
		settings.PUT("/default-quota", r.settingsHandler.UpdateDefaultQuota)
		
		// 默认速率限制
		settings.GET("/default-rate-limit", r.settingsHandler.GetDefaultRateLimit)
		settings.PUT("/default-rate-limit", r.settingsHandler.UpdateDefaultRateLimit)
	}
}



// setupProxyRoutes 设置代理路由（OpenAI 兼容接口）
func (r *Router) setupProxyRoutes() {
	// OpenAI 兼容接口
	v1 := r.engine.Group("/v1")
	v1.Use(r.mw.APIKey.Handle()) // API Key 验证
	{
		// OpenAI 格式
		v1.POST("/chat/completions", r.proxyHandler.ChatCompletionsOpenAI)
		
		// Anthropic 格式
		v1.POST("/messages", r.proxyHandler.ChatCompletionsAnthropic)
		
		// Gemini 格式 - 使用通配符匹配
		v1.POST("/models/*action", r.proxyHandler.ChatCompletionsGemini)
	}
}
