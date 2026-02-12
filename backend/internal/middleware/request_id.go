package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestID 请求ID中间件
type RequestID struct{}

// NewRequestID 创建请求ID中间件实例
func NewRequestID() *RequestID {
	return &RequestID{}
}

// Handle 请求ID处理
func (m *RequestID) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试从header获取请求ID
		requestID := c.GetHeader("X-Request-ID")
		
		// 如果没有，生成新的请求ID
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// 设置请求ID到上下文和响应头
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}
