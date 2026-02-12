package middleware

import (
	"api-aggregator/backend/internal/domain/auth"
	"api-aggregator/backend/pkg/response"
	"strings"

	"github.com/gin-gonic/gin"
)

// Auth JWT认证中间件
type Auth struct {
	authService auth.Service
}

// NewAuth 创建JWT认证中间件实例
func NewAuth(authService auth.Service) *Auth {
	return &Auth{
		authService: authService,
	}
}

// Handle JWT认证处理
func (m *Auth) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "missing authorization header")
			c.Abort()
			return
		}

		// 提取token "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "invalid authorization header format")
			c.Abort()
			return
		}

		token := parts[1]

		// 验证token
		claims, err := m.authService.ValidateToken(c.Request.Context(), token)
		if err != nil {
			response.Unauthorized(c, "invalid or expired token")
			c.Abort()
			return
		}

		// 设置用户信息到上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}
