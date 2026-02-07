package middleware

// This file provides an example of how to set up admin routes with proper middleware protection
// This is for reference only and should be adapted to your actual application structure

/*
Example usage in your main.go or router setup:

import (
	"api-aggregator/backend/internal/api"
	"api-aggregator/backend/internal/middleware"
	"api-aggregator/backend/internal/repository"
	"api-aggregator/backend/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB, jwtSecret string) *gin.Engine {
	r := gin.Default()

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	apiKeyRepo := repository.NewAPIKeyRepository(db)
	apiConfigRepo := repository.NewAPIConfigRepository(db)

	// Initialize services
	authService := service.NewAuthService(userRepo, jwtSecret)
	apiKeyService := service.NewAPIKeyService(apiKeyRepo, userRepo)
	apiConfigService := service.NewAPIConfigService(apiConfigRepo)

	// Initialize handlers
	authHandler := api.NewAuthHandler(authService)
	apiKeyHandler := api.NewAPIKeyHandler(apiKeyService)
	apiConfigHandler := api.NewAPIConfigHandler(apiConfigService)

	// Public routes (no authentication)
	r.POST("/api/auth/register", authHandler.Register)
	r.POST("/api/auth/login", authHandler.Login)

	// User routes (authentication required)
	userRoutes := r.Group("/api/user")
	userRoutes.Use(middleware.AuthMiddleware(authService))
	{
		userRoutes.GET("/info", authHandler.GetUserInfo)
		userRoutes.POST("/signin", authHandler.SignIn)
	}

	// API Key routes (authentication required)
	apiKeyRoutes := r.Group("/api/apikeys")
	apiKeyRoutes.Use(middleware.AuthMiddleware(authService))
	{
		apiKeyRoutes.GET("", apiKeyHandler.GetAPIKeys)
		apiKeyRoutes.POST("", apiKeyHandler.CreateAPIKey)
		apiKeyRoutes.DELETE("/:id", apiKeyHandler.DeleteAPIKey)
	}

	// Admin routes (authentication + admin privileges required)
	adminRoutes := r.Group("/api/admin")
	adminRoutes.Use(middleware.AuthMiddleware(authService))
	adminRoutes.Use(middleware.AdminMiddleware(userRepo))
	{
		// User management endpoints
		adminRoutes.GET("/users", adminHandler.GetAllUsers)
		adminRoutes.GET("/users/:id", adminHandler.GetUser)
		adminRoutes.PUT("/users/:id", adminHandler.UpdateUser)
		adminRoutes.PUT("/users/:id/quota", adminHandler.AdjustUserQuota)
		adminRoutes.PUT("/users/:id/status", adminHandler.UpdateUserStatus)
		adminRoutes.DELETE("/users/:id", adminHandler.DeleteUser)

		// API configuration management endpoints
		adminRoutes.GET("/api-configs", apiConfigHandler.GetAllConfigs)
		adminRoutes.GET("/api-configs/:id", apiConfigHandler.GetConfig)
		adminRoutes.POST("/api-configs", apiConfigHandler.CreateConfig)
		adminRoutes.PUT("/api-configs/:id", apiConfigHandler.UpdateConfig)
		adminRoutes.DELETE("/api-configs/:id", apiConfigHandler.DeleteConfig)
		adminRoutes.PUT("/api-configs/:id/activate", apiConfigHandler.ActivateConfig)
		adminRoutes.PUT("/api-configs/:id/deactivate", apiConfigHandler.DeactivateConfig)

		// Statistics and monitoring endpoints
		adminRoutes.GET("/stats", adminHandler.GetStatistics)
		adminRoutes.GET("/logs", adminHandler.GetRequestLogs)
		adminRoutes.GET("/logs/export", adminHandler.ExportLogs)
	}

	return r
}

// Example of an admin handler that uses the user from context
func ExampleAdminHandler(c *gin.Context) {
	// The admin middleware sets the user object in context
	userValue, exists := c.Get("user")
	if !exists {
		// This should never happen if middleware is properly configured
		c.JSON(500, gin.H{"error": "User not found in context"})
		return
	}

	user, ok := userValue.(*models.User)
	if !ok {
		c.JSON(500, gin.H{"error": "Invalid user type in context"})
		return
	}

	// Now you can use the user object
	// user.ID, user.Username, user.IsAdmin, etc.

	c.JSON(200, gin.H{
		"message": "Admin access granted",
		"admin": gin.H{
			"id":       user.ID,
			"username": user.Username,
		},
	})
}
*/
