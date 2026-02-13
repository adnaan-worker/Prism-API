package proxy

import "api-aggregator/backend/internal/adapter"

// ProxyRequest 代理请求
type ProxyRequest struct {
	UserID      uint                  `json:"-"` // 从上下文获取
	APIKeyID    uint                  `json:"-"` // 从上下文获取
	Model       string                `json:"model" binding:"required"`
	Stream      bool                  `json:"stream"`
	ChatRequest *adapter.ChatRequest  `json:"-"` // 完整的请求对象
}
