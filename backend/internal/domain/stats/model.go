package stats

// Stats 模块不需要独立的数据模型
// 统计数据从其他模块聚合而来

// StatsOverview 统计概览
type StatsOverview struct {
	TotalUsers      int64 `json:"total_users"`
	ActiveUsers     int64 `json:"active_users"`
	TotalRequests   int64 `json:"total_requests"`
	TodayRequests   int64 `json:"today_requests"`
	SuccessRequests int64 `json:"success_requests"`
	FailedRequests  int64 `json:"failed_requests"`
}

// RequestTrendItem 请求趋势项
type RequestTrendItem struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// ModelUsageItem 模型使用项
type ModelUsageItem struct {
	Model string `json:"model"`
	Count int64  `json:"count"`
}

// UserGrowthItem 用户增长项
type UserGrowthItem struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// TokenUsageItem Token使用项
type TokenUsageItem struct {
	Date   string `json:"date"`
	Tokens int64  `json:"tokens"`
}
