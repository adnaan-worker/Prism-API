package middleware

import (
	"api-aggregator/backend/pkg/response"
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

// Timeout 超时中间件
type Timeout struct {
	timeout time.Duration
}

// NewTimeout 创建超时中间件实例
func NewTimeout(timeout time.Duration) *Timeout {
	return &Timeout{
		timeout: timeout,
	}
}

// Handle 超时处理
func (m *Timeout) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 创建带超时的上下文
		ctx, cancel := context.WithTimeout(c.Request.Context(), m.timeout)
		defer cancel()

		// 替换请求的上下文
		c.Request = c.Request.WithContext(ctx)

		// 使用channel来检测请求是否完成
		finished := make(chan struct{})
		go func() {
			c.Next()
			close(finished)
		}()

		// 等待请求完成或超时
		select {
		case <-finished:
			// 请求正常完成
			return
		case <-ctx.Done():
			// 请求超时
			if ctx.Err() == context.DeadlineExceeded {
				response.RequestTimeout(c, "request timeout")
				c.Abort()
			}
		}
	}
}
