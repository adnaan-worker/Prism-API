package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"api-aggregator/backend/internal/models"
	"api-aggregator/backend/pkg/redis"

	"github.com/gin-gonic/gin"
)

// RateLimiterMiddleware implements rate limiting based on API key
func RateLimiterMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get API key from context (set by auth middleware)
		apiKeyInterface, exists := c.Get("api_key")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "API key not found in context",
			})
			c.Abort()
			return
		}

		apiKey, ok := apiKeyInterface.(*models.APIKey)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Invalid API key type in context",
			})
			c.Abort()
			return
		}

		// Check rate limit
		allowed, err := checkRateLimit(apiKey.ID, apiKey.RateLimit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to check rate limit",
			})
			c.Abort()
			return
		}

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"code":    429001,
				"message": fmt.Sprintf("Rate limit of %d requests per minute exceeded", apiKey.RateLimit),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// checkRateLimit checks if the API key has exceeded its rate limit
// Uses Redis sliding window algorithm
func checkRateLimit(apiKeyID uint, rateLimit int) (bool, error) {
	ctx := context.Background()

	// Create a key for this API key and current minute
	now := time.Now()
	key := fmt.Sprintf("rate_limit:%d:%d", apiKeyID, now.Unix()/60)

	// Increment the counter
	count, err := redis.Client.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}

	// Set expiration on first request (2 minutes to be safe)
	if count == 1 {
		redis.Client.Expire(ctx, key, 2*time.Minute)
	}

	// Check if limit exceeded
	return count <= int64(rateLimit), nil
}
