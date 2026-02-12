package middleware

import (
	"api-aggregator/backend/internal/domain/user"
	"api-aggregator/backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// Admin 管理员权限中间件
type Admin struct {
	userService user.Service
}

// NewAdmin 创建管理员权限中间件实例
func NewAdmin(userService user.Service) *Admin {
	return &Admin{
		userService: userService,
	}
}

// Handle 管理员权限检查处理
// 此中间件应在Auth或APIKey中间件之后使用
func (m *Admin) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文获取用户ID（由Auth或APIKey中间件设置）
		userIDValue, exists := c.Get("user_id")
		if !exists {
			response.Unauthorized(c, "user not authenticated")
			c.Abort()
			return
		}

		userID, ok := userIDValue.(uint)
		if !ok {
			response.InternalError(c, "invalid user id format")
			c.Abort()
			return
		}

		// 获取用户信息
		userObj, err := m.userService.GetUserByID(c.Request.Context(), userID)
		if err != nil {
			response.HandleError(c, err)
			c.Abort()
			return
		}

		// 检查是否为管理员
		if !userObj.IsAdmin {
			response.Forbidden(c, "admin privileges required")
			c.Abort()
			return
		}

		// 设置用户对象到上下文供后续使用
		c.Set("user", userObj)
		c.Next()
	}
}
