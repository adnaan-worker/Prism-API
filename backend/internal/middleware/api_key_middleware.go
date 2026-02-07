package middleware

import (
	"api-aggregator/backend/internal/service"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// APIKeyMiddleware creates a middleware that validates API keys
func APIKeyMiddleware(apiKeyService *service.APIKeyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get API key from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    401001,
					"message": "Unauthorized",
					"details": "Missing authorization header",
				},
			})
			c.Abort()
			return
		}

		// Extract API key from "Bearer <key>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    401001,
					"message": "Unauthorized",
					"details": "Invalid authorization header format",
				},
			})
			c.Abort()
			return
		}

		apiKey := parts[1]

		// Validate API key
		apiKeyObj, err := apiKeyService.GetAPIKeyByKey(c.Request.Context(), apiKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    401001,
					"message": "Unauthorized",
					"details": "Invalid or inactive API key",
				},
			})
			c.Abort()
			return
		}

		// Check if API key is active
		if !apiKeyObj.IsActive {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    401001,
					"message": "Unauthorized",
					"details": "API key is inactive",
				},
			})
			c.Abort()
			return
		}

		// Set user ID and API key object in context
		c.Set("user_id", apiKeyObj.UserID)
		c.Set("api_key", apiKeyObj)
		c.Next()
	}
}
