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
		// 适配多种协议的API密钥请求头提取
		var key string

		// 1. OpenAI 格式: Authorization: Bearer <key>
		if authHeader := c.GetHeader("Authorization"); authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
				key = strings.TrimSpace(parts[1])
			}
		}

		// 2. Anthropic 格式: x-api-key: <key>
		if key == "" {
			key = strings.TrimSpace(c.GetHeader("x-api-key"))
		}

		// 3. Azure / 其他常见格式: api-key: <key>
		if key == "" {
			key = strings.TrimSpace(c.GetHeader("api-key"))
		}

		if key == "" {
			response.Unauthorized(c, "missing or invalid api key in request headers")
			c.Abort()
			return
		}

		// 验证API密钥
		userID, apiKeyID, err := m.apiKeyService.ValidateAPIKey(c.Request.Context(), key)
		if err != nil {
			response.Unauthorized(c, "invalid or inactive API key")
			c.Abort()
			return
		}

		// 设置用户信息到上下文
		c.Set("user_id", userID)
		c.Set("api_key_id", apiKeyID)
		c.Set("api_key", key)
		c.Next()
	}
}
