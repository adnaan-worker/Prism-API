package app

import (
	"api-aggregator/backend/config"
	"api-aggregator/backend/internal/adapter"
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
	"api-aggregator/backend/internal/router"
	pkgCache "api-aggregator/backend/pkg/cache"
	"api-aggregator/backend/pkg/embedding"
	"api-aggregator/backend/pkg/logger"
	"api-aggregator/backend/pkg/runtime"
	"context"
	"database/sql"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// App 应用实例
type App struct {
	Config        *config.Config
	DB            *gorm.DB
	Cache         pkgCache.Cache
	Logger        *logger.Logger
	Engine        *gin.Engine
	RuntimeConfig *runtime.Manager
}

// New 创建应用实例
func New(cfg *config.Config) (*App, error) {
	app := &App{
		Config: cfg,
	}

	// 初始化日志
	if err := app.initLogger(); err != nil {
		return nil, err
	}

	// 初始化数据库
	if err := app.initDatabase(); err != nil {
		return nil, err
	}

	// 初始化缓存
	if err := app.initCache(); err != nil {
		return nil, err
	}

	// 初始化运行时配置管理器
	if err := app.initRuntimeConfig(); err != nil {
		return nil, err
	}

	// 初始化Gin引擎
	app.initEngine()

	// 初始化路由
	if err := app.initRouter(); err != nil {
		return nil, err
	}

	return app, nil
}

// initLogger 初始化日志
func (app *App) initLogger() error {
	log, err := logger.New(&logger.Config{
		Level:      "info",
		OutputPath: "logs/app.log",
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     7,
		Compress:   true,
	})
	if err != nil {
		return err
	}
	app.Logger = log
	app.Logger.Info("Logger initialized")
	return nil
}

// initDatabase 初始化数据库
func (app *App) initDatabase() error {
	db, err := gorm.Open(postgres.Open(app.Config.Database.URL), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Info),
	})
	if err != nil {
		return err
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	app.configureDBPool(sqlDB)

	app.DB = db
	app.Logger.Info("Database connected and pool configured")
	return nil
}

// configureDBPool 配置数据库连接池
func (app *App) configureDBPool(sqlDB *sql.DB) {
	sqlDB.SetMaxOpenConns(app.Config.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(app.Config.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(app.Config.Database.ConnMaxLifetime)
	app.Logger.Info("Database pool configured",
		logger.Int("max_open", app.Config.Database.MaxOpenConns),
		logger.Int("max_idle", app.Config.Database.MaxIdleConns),
		logger.Duration("max_lifetime", app.Config.Database.ConnMaxLifetime),
	)
}

// initCache 初始化缓存
func (app *App) initCache() error {
	// 处理 Redis URL，去掉 redis:// 前缀
	redisAddr := app.Config.Redis.URL
	if len(redisAddr) > 8 && redisAddr[:8] == "redis://" {
		redisAddr = redisAddr[8:]
	}
	
	cache, err := pkgCache.NewRedis(&pkgCache.RedisConfig{
		Addr:         redisAddr,
		Password:     "",
		DB:           0,
		PoolSize:     app.Config.Redis.PoolSize,
		MinIdleConns: app.Config.Redis.MinIdleConn,
	})
	if err != nil {
		return err
	}
	app.Cache = cache
	app.Logger.Info("Redis cache connected")
	return nil
}

// initEngine 初始化Gin引擎
func (app *App) initEngine() {
	if gin.Mode() == gin.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}
	app.Engine = gin.New()
	app.Logger.Info("Gin engine initialized")
}

// initRouter 初始化路由
func (app *App) initRouter() error {
	// 初始化仓储层
	userRepo := user.NewRepository(app.DB)
	authRepo := auth.NewRepository(app.DB)
	apiKeyRepo := apikey.NewRepository(app.DB)
	apiConfigRepo := apiconfig.NewRepository(app.DB)
	quotaRepo := quota.NewRepository(app.DB)
	pricingRepo := pricing.NewRepository(app.DB)
	logRepo := log.NewRepository(app.DB)
	statsRepo := stats.NewRepository(app.DB)
	cacheRepo := cache.NewRepository(app.DB)
	loadBalancerRepo := loadbalancer.NewRepository(app.DB)
	accountPoolRepo := accountpool.NewRepository(app.DB)
	settingsRepo := settings.NewRepository(app.DB)

	// 初始化服务层
	userService := user.NewService(userRepo, *app.Logger)
	authService := auth.NewService(authRepo, app.Config.JWT.Secret, *app.Logger)
	apiKeyService := apikey.NewService(apiKeyRepo, *app.Logger)
	apiConfigService := apiconfig.NewService(apiConfigRepo, *app.Logger)
	quotaService := quota.NewService(quotaRepo, *app.Logger)
	pricingService := pricing.NewService(pricingRepo, apiConfigRepo, *app.Logger)
	logService := log.NewService(logRepo, userRepo, *app.Logger)
	statsService := stats.NewService(statsRepo, *app.Logger)
	cacheService := cache.NewService(cacheRepo, *app.Logger)
	loadBalancerService := loadbalancer.NewService(loadBalancerRepo, apiConfigRepo)
	accountPoolService := accountpool.NewService(accountPoolRepo)
	settingsService := settings.NewService(settingsRepo, userRepo)

	// 初始化 Adapter Factory
	adapterFactory := adapter.NewFactory()

	// 初始化模型映射器（用于 Kiro）
	modelMapper := apiconfig.NewModelMapper(apiConfigRepo)

	// 初始化账号池管理器
	poolManager := accountpool.NewPoolManager(accountPoolRepo, modelMapper)
	
	// 初始化 Token 刷新调度器
	refreshScheduler := accountpool.NewRefreshScheduler(
		accountPoolRepo,
		accountpool.NewKiroRefreshService(),
		5*time.Minute, // 每5分钟检查一次
		*app.Logger,
	)
	
	// 启动刷新调度器
	go refreshScheduler.Start(context.Background())

	// 初始化 Embedding 客户端（如果启用）
	var embeddingClient *embedding.Client
	if app.Config.Embedding.Enabled {
		embeddingClient = embedding.NewClient(
			app.Config.Embedding.URL,
			app.Config.Embedding.Timeout,
		)
	}

	// 初始化 Proxy 服务
	proxyService := proxy.NewService(
		adapterFactory,
		apiConfigRepo,
		poolManager,
		loadBalancerService,
		cacheService,
		quotaService,
		pricingService,
		logService,
		app.RuntimeConfig,
		*app.Logger,
	)
	if embeddingClient != nil {
		proxyService.SetEmbeddingClient(embeddingClient)
	}

	// 初始化处理器层
	authHandler := auth.NewHandler(authService)
	apiKeyHandler := apikey.NewHandler(apiKeyService)
	apiConfigHandler := apiconfig.NewHandler(apiConfigService)
	userHandler := user.NewHandler(userService)
	quotaHandler := quota.NewHandler(quotaService)
	statsHandler := stats.NewHandler(statsService)
	logHandler := log.NewHandler(logService)
	pricingHandler := pricing.NewHandler(pricingService)
	cacheHandler := cache.NewHandler(cacheService)
	loadBalancerHandler := loadbalancer.NewHandler(loadBalancerService)
	accountPoolHandler := accountpool.NewHandler(accountPoolService)
	settingsHandler := settings.NewHandler(settingsService)
	proxyHandler := proxy.NewHandler(proxyService)

	// 初始化中间件管理器
	mw := middleware.NewManager(&middleware.Config{
		AuthService:   authService,
		APIKeyService: apiKeyService,
		UserService:   userService,
		Cache:         app.Cache,
		Logger:        app.Logger,
		CORSConfig: &middleware.CORSConfig{
			AllowOrigins:     []string{"*"},
			AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: false,
			MaxAge:           86400,
		},
		RequestTimeout: app.Config.Server.RequestTimeout,
	})

	// 初始化路由管理器
	r := router.New(&router.Config{
		Engine:              app.Engine,
		MW:                  mw,
		AuthHandler:         authHandler,
		APIKeyHandler:       apiKeyHandler,
		APIConfigHandler:    apiConfigHandler,
		UserHandler:         userHandler,
		QuotaHandler:        quotaHandler,
		StatsHandler:        statsHandler,
		LogHandler:          logHandler,
		PricingHandler:      pricingHandler,
		CacheHandler:        cacheHandler,
		LoadBalancerHandler: loadBalancerHandler,
		AccountPoolHandler:  accountPoolHandler,
		SettingsHandler:     settingsHandler,
		ProxyHandler:        proxyHandler,
	})

	// 设置路由
	r.Setup()
	app.Logger.Info("Router initialized")

	return nil
}

// Run 运行应用
func (app *App) Run(addr string) error {
	app.Logger.Info("Server starting", logger.String("addr", addr))
	return app.Engine.Run(addr)
}

// Close 关闭应用
func (app *App) Close() error {
	// 停止运行时配置管理器
	if app.RuntimeConfig != nil {
		app.RuntimeConfig.Stop()
	}
	
	if app.Logger != nil {
		app.Logger.Sync()
	}
	return nil
}


// initRuntimeConfig 初始化运行时配置管理器
func (app *App) initRuntimeConfig() error {
	manager := runtime.NewManager(app.DB)
	
	// 加载配置
	if err := manager.Load(); err != nil {
		app.Logger.Warn("Failed to load runtime config, using defaults", logger.Error(err))
	}
	
	// 启动自动重载（每分钟检查一次）
	manager.StartAutoReload(1 * time.Minute)
	
	app.RuntimeConfig = manager
	app.Logger.Info("Runtime config manager initialized")
	return nil
}
