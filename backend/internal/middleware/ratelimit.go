package middleware

import (
	"api-aggregator/backend/internal/domain/apikey"
	"api-aggregator/backend/pkg/cache"
	"api-aggregator/backend/pkg/response"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimit 速率限制中间件
type RateLimit struct {
	cache cache.Cache
}

// NewRateLimit 创建速率限制中间件实例
func NewRateLimit(cache cache.Cache) *RateLimit {
	return &RateLimit{
		cache: cache,
	}
}

// Handle 速率限制处理
// 此中间件应在APIKey中间件之后使用
func (m *RateLimit) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文获取API密钥（由APIKey中间件设置）
		apiKeyInterface, exists := c.Get("api_key")
		if !exists {
			response.Unauthorized(c, "API key not found in context")
			c.Abort()
			return
		}

		apiKeyObj, ok := apiKeyInterface.(*apikey.APIKeyResponse)
		if !ok {
			response.InternalError(c, "invalid API key type in context")
			c.Abort()
			return
		}

		// 检查速率限制
		allowed, err := m.checkRateLimit(apiKeyObj.ID, apiKeyObj.RateLimit)
		if err != nil {
			response.InternalError(c, "failed to check rate limit")
			c.Abort()
			return
		}

		if !allowed {
			response.TooManyRequests(c, fmt.Sprintf("rate limit of %d requests per minute exceeded", apiKeyObj.RateLimit))
			c.Abort()
			return
		}

		c.Next()
	}
}

// checkRateLimit 检查API密钥是否超过速率限制
// 使用Redis滑动窗口算法
func (m *RateLimit) checkRateLimit(apiKeyID uint, rateLimit int) (bool, error) {
	// 创建当前分钟的key
	now := time.Now()
	key := fmt.Sprintf("rate_limit:%d:%d", apiKeyID, now.Unix()/60)

	// 获取当前计数
	var count int64
	err := m.cache.Get(key, &count)
	if err != nil {
		// 如果key不存在，初始化为0
		count = 0
	}

	// 增加计数
	count++

	// 保存计数，过期时间2分钟
	if err := m.cache.Set(key, count, 2*time.Minute); err != nil {
		return false, err
	}

	// 检查是否超过限制
	return count <= int64(rateLimit), nil
}
