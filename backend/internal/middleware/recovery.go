package middleware

import (
	"api-aggregator/backend/pkg/logger"
	"api-aggregator/backend/pkg/response"
	"fmt"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Recovery 恢复中间件
type Recovery struct {
	logger *logger.Logger
}

// NewRecovery 创建恢复中间件实例
func NewRecovery(log *logger.Logger) *Recovery {
	return &Recovery{
		logger: log,
	}
}

// Handle 恢复处理
func (m *Recovery) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 获取堆栈信息
				stack := debug.Stack()

				// 记录panic信息
				m.logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("stack", string(stack)),
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
					zap.String("ip", c.ClientIP()),
				)

				// 返回错误响应
				response.InternalError(c, fmt.Sprintf("internal server error: %v", err))
				c.Abort()
			}
		}()

		c.Next()
	}
}
