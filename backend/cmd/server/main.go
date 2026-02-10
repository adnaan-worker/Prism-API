package main

import (
	"api-aggregator/backend/config"
	"api-aggregator/backend/internal/api"
	"api-aggregator/backend/internal/middleware"
	"api-aggregator/backend/internal/repository"
	"api-aggregator/backend/internal/service"
	"api-aggregator/backend/pkg/redis"
	"database/sql"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Connect to database with connection pooling
	db, err := gorm.Open(postgres.Open(cfg.Database.URL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get database instance:", err)
	}
	configureDBPool(sqlDB, cfg.Database)

	// Initialize Redis with connection pooling
	if err := redis.InitRedis(redis.Config{
		URL:         cfg.Redis.URL,
		PoolSize:    cfg.Redis.PoolSize,
		MinIdleConn: cfg.Redis.MinIdleConn,
	}); err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	defer redis.CloseRedis()

	log.Println("Successfully connected to database and Redis")

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	apiKeyRepo := repository.NewAPIKeyRepository(db)
	apiConfigRepo := repository.NewAPIConfigRepository(db)
	signInRepo := repository.NewSignInRepository(db)
	requestLogRepo := repository.NewRequestLogRepository(db)
	lbRepo := repository.NewLoadBalancerRepository(db)

	// Initialize services
	authService := service.NewAuthService(userRepo, cfg.JWT.Secret)
	apiKeyService := service.NewAPIKeyService(apiKeyRepo)
	apiConfigService := service.NewAPIConfigService(apiConfigRepo)
	modelService := service.NewModelService(apiConfigRepo)
	userService := service.NewUserService(userRepo)
	quotaService := service.NewQuotaService(userRepo, signInRepo)
	statsService := service.NewStatsService(userRepo, requestLogRepo)
	logService := service.NewLogService(requestLogRepo)
	proxyService := service.NewProxyService(apiKeyRepo, apiConfigRepo, userRepo, requestLogRepo, quotaService)
	lbService := service.NewLoadBalancerService(lbRepo)

	// Initialize handlers
	authHandler := api.NewAuthHandler(authService)
	apiKeyHandler := api.NewAPIKeyHandler(apiKeyService)
	apiConfigHandler := api.NewAPIConfigHandler(apiConfigService)
	modelHandler := api.NewModelHandler(modelService)
	providerHandler := api.NewProviderHandler()
	userHandler := api.NewUserHandler(userService)
	quotaHandler := api.NewQuotaHandler(quotaService)
	statsHandler := api.NewStatsHandler(statsService)
	logHandler := api.NewLogHandler(logService)
	proxyHandler := api.NewProxyHandler(proxyService)
	lbHandler := api.NewLoadBalancerHandler(lbService, apiConfigService, modelService)

	// Setup router
	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// Public routes (no authentication)
	r.POST("/api/auth/register", authHandler.Register)
	r.POST("/api/auth/login", authHandler.Login)

	// OpenAI-compatible API routes (API key authentication)
	v1Routes := r.Group("/v1")
	{
		// OpenAI format
		v1Routes.POST("/chat/completions", proxyHandler.ChatCompletions)

		// Anthropic format
		v1Routes.POST("/messages", proxyHandler.AnthropicMessages)

		// Gemini format - using wildcard to match the full path
		v1Routes.POST("/models/*action", proxyHandler.GeminiGenerateContent)
	}

	// User routes (authentication required)
	userRoutes := r.Group("/api/user")
	userRoutes.Use(middleware.AuthMiddleware(authService))
	{
		userRoutes.GET("/info", quotaHandler.GetQuotaInfo)
		userRoutes.POST("/signin", quotaHandler.SignIn)
		userRoutes.GET("/usage-history", quotaHandler.GetUsageHistory)
	}

	// API Key routes (authentication required)
	apiKeyRoutes := r.Group("/api/apikeys")
	apiKeyRoutes.Use(middleware.AuthMiddleware(authService))
	{
		apiKeyRoutes.GET("", apiKeyHandler.GetAPIKeys)
		apiKeyRoutes.POST("", apiKeyHandler.CreateAPIKey)
		apiKeyRoutes.DELETE("/:id", apiKeyHandler.DeleteAPIKey)
	}

	// Model routes (authentication required)
	modelRoutes := r.Group("/api/models")
	modelRoutes.Use(middleware.AuthMiddleware(authService))
	{
		modelRoutes.GET("", modelHandler.GetAllModels)
		modelRoutes.GET("/provider", modelHandler.GetModelsByProvider)
		modelRoutes.GET("/:name", modelHandler.GetModelInfo)
	}

	// Admin routes (authentication + admin privileges required)
	adminRoutes := r.Group("/api/admin")
	adminRoutes.Use(middleware.AuthMiddleware(authService))
	adminRoutes.Use(middleware.AdminMiddleware(userRepo))
	{
		// User management endpoints
		adminRoutes.GET("/users", userHandler.GetUsers)
		adminRoutes.GET("/users/:id", userHandler.GetUserByID)
		adminRoutes.PUT("/users/:id/status", userHandler.UpdateUserStatus)
		adminRoutes.PUT("/users/:id/quota", userHandler.UpdateUserQuota)

		// API configuration management endpoints
		adminRoutes.GET("/api-configs", apiConfigHandler.GetAllConfigs)
		adminRoutes.GET("/api-configs/:id", apiConfigHandler.GetConfig)
		adminRoutes.POST("/api-configs", apiConfigHandler.CreateConfig)
		adminRoutes.PUT("/api-configs/:id", apiConfigHandler.UpdateConfig)
		adminRoutes.DELETE("/api-configs/:id", apiConfigHandler.DeleteConfig)
		adminRoutes.PUT("/api-configs/:id/activate", apiConfigHandler.ActivateConfig)
		adminRoutes.PUT("/api-configs/:id/deactivate", apiConfigHandler.DeactivateConfig)

		// Batch operations for API configurations
		adminRoutes.POST("/api-configs/batch/delete", apiConfigHandler.BatchDeleteConfigs)
		adminRoutes.POST("/api-configs/batch/activate", apiConfigHandler.BatchActivateConfigs)
		adminRoutes.POST("/api-configs/batch/deactivate", apiConfigHandler.BatchDeactivateConfigs)

		// Statistics and logging endpoints
		adminRoutes.GET("/stats/overview", statsHandler.GetStatsOverview)
		adminRoutes.GET("/logs", logHandler.GetLogs)
		adminRoutes.GET("/logs/export", logHandler.ExportLogs)

		// Provider endpoints
		adminRoutes.POST("/providers/fetch-models", providerHandler.FetchModels)

		// Load balancer endpoints
		adminRoutes.GET("/load-balancer/configs", lbHandler.GetConfigs)
		adminRoutes.GET("/load-balancer/configs/:id", lbHandler.GetConfig)
		adminRoutes.POST("/load-balancer/configs", lbHandler.CreateConfig)
		adminRoutes.PUT("/load-balancer/configs/:id", lbHandler.UpdateConfig)
		adminRoutes.DELETE("/load-balancer/configs/:id", lbHandler.DeleteConfig)
		adminRoutes.GET("/load-balancer/models/:model/endpoints", lbHandler.GetModelEndpoints)
		adminRoutes.GET("/models", lbHandler.GetAvailableModels)
	}

	// Start server with timeouts
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Starting server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// configureDBPool configures database connection pool
func configureDBPool(sqlDB *sql.DB, cfg config.DatabaseConfig) {
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	log.Printf("Database pool configured: MaxOpen=%d, MaxIdle=%d, MaxLifetime=%v",
		cfg.MaxOpenConns, cfg.MaxIdleConns, cfg.ConnMaxLifetime)
}
