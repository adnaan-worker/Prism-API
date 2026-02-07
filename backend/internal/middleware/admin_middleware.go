package middleware

import (
	"api-aggregator/backend/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AdminMiddleware creates a middleware that validates admin privileges
// This middleware should be used after AuthMiddleware
func AdminMiddleware(userRepo *repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context (set by AuthMiddleware)
		userIDValue, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    401001,
					"message": "Unauthorized",
					"details": "User not authenticated",
				},
			})
			c.Abort()
			return
		}

		userID, ok := userIDValue.(uint)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    500001,
					"message": "Internal Error",
					"details": "Invalid user ID format",
				},
			})
			c.Abort()
			return
		}

		// Fetch user from database
		user, err := userRepo.FindByID(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    500001,
					"message": "Internal Error",
					"details": "Failed to fetch user information",
				},
			})
			c.Abort()
			return
		}

		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    401001,
					"message": "Unauthorized",
					"details": "User not found",
				},
			})
			c.Abort()
			return
		}

		// Check if user is admin
		if !user.IsAdmin {
			c.JSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    403001,
					"message": "Forbidden",
					"details": "Admin privileges required",
				},
			})
			c.Abort()
			return
		}

		// Set user object in context for later use
		c.Set("user", user)
		c.Next()
	}
}
