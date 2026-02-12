package middleware

import (
	"api-aggregator/backend/internal/domain/apikey"
	"api-aggregator/backend/pkg/response"
	"strings"

	"github.com/gin-gonic/gin"
)

// APIKey API密钥认证中间件
type APIKey struct {
	apiKeyService apikey.Service
}

// NewAPIKey 创建API密钥认证中间件实例
func NewAPIKey(apiKeyService apikey.Service) *APIKey {
	return &APIKey{
		apiKeyService: apiKeyService,
	}
}

// Handle API密钥认证处理
func (m *APIKey) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "missing authorization header")
			c.Abort()
			return
		}

		// 提取API密钥 "Bearer <key>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "invalid authorization header format")
			c.Abort()
			return
		}

		key := parts[1]

		// 验证API密钥
		userID, err := m.apiKeyService.ValidateAPIKey(c.Request.Context(), key)
		if err != nil {
			response.Unauthorized(c, "invalid or inactive API key")
			c.Abort()
			return
		}

		// 设置用户信息到上下文
		c.Set("user_id", userID)
		c.Set("api_key", key)
		c.Next()
	}
}
