package middleware

import (
	"api-aggregator/backend/pkg/logger"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logger 日志中间件
type Logger struct {
	logger *logger.Logger
}

// NewLogger 创建日志中间件实例
func NewLogger(log *logger.Logger) *Logger {
	return &Logger{
		logger: log,
	}
}

// Handle 日志处理
func (m *Logger) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录请求开始时间
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// 处理请求
		c.Next()

		// 计算请求耗时
		latency := time.Since(start)

		// 获取响应状态码
		statusCode := c.Writer.Status()

		// 构建日志字段
		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		}

		// 如果有错误，添加错误信息
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.String()))
		}

		// 根据状态码选择日志级别
		if statusCode >= 500 {
			m.logger.Error("Server error", fields...)
		} else if statusCode >= 400 {
			m.logger.Warn("Client error", fields...)
		} else {
			m.logger.Info("Request completed", fields...)
		}
	}
}
