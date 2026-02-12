package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// CORS 跨域资源共享中间件
type CORS struct {
	allowOrigins     []string
	allowMethods     []string
	allowHeaders     []string
	exposeHeaders    []string
	allowCredentials bool
	maxAge           int
}

// CORSConfig CORS配置
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int
}

// NewCORS 创建CORS中间件实例
func NewCORS(config *CORSConfig) *CORS {
	if config == nil {
		config = DefaultCORSConfig()
	}
	return &CORS{
		allowOrigins:     config.AllowOrigins,
		allowMethods:     config.AllowMethods,
		allowHeaders:     config.AllowHeaders,
		exposeHeaders:    config.ExposeHeaders,
		allowCredentials: config.AllowCredentials,
		maxAge:           config.MaxAge,
	}
}

// DefaultCORSConfig 默认CORS配置
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           86400, // 24小时
	}
}

// Handle CORS处理
func (m *CORS) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// 设置CORS头
		if len(m.allowOrigins) > 0 {
			if m.allowOrigins[0] == "*" {
				c.Header("Access-Control-Allow-Origin", "*")
			} else {
				// 检查origin是否在允许列表中
				for _, allowOrigin := range m.allowOrigins {
					if origin == allowOrigin {
						c.Header("Access-Control-Allow-Origin", origin)
						break
					}
				}
			}
		}

		if len(m.allowMethods) > 0 {
			c.Header("Access-Control-Allow-Methods", joinStrings(m.allowMethods, ", "))
		}

		if len(m.allowHeaders) > 0 {
			c.Header("Access-Control-Allow-Headers", joinStrings(m.allowHeaders, ", "))
		}

		if len(m.exposeHeaders) > 0 {
			c.Header("Access-Control-Expose-Headers", joinStrings(m.exposeHeaders, ", "))
		}

		if m.allowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if m.maxAge > 0 {
			c.Header("Access-Control-Max-Age", fmt.Sprintf("%d", m.maxAge))
		}

		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// joinStrings 连接字符串数组
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
