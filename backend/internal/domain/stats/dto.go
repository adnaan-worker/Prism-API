package stats

// GetStatsOverviewResponse 统计概览响应
type GetStatsOverviewResponse struct {
	TotalUsers      int64 `json:"total_users"`
	ActiveUsers     int64 `json:"active_users"`
	TotalRequests   int64 `json:"total_requests"`
	TodayRequests   int64 `json:"today_requests"`
	SuccessRequests int64 `json:"success_requests"`
	FailedRequests  int64 `json:"failed_requests"`
}

// GetRequestTrendRequest 获取请求趋势请求
type GetRequestTrendRequest struct {
	Days int `form:"days" binding:"omitempty,min=1,max=90"`
}

// GetRequestTrendResponse 请求趋势响应
type GetRequestTrendResponse struct {
	Trend []RequestTrendItem `json:"trend"`
	Days  int                `json:"days"`
}

// GetModelUsageRequest 获取模型使用请求
type GetModelUsageRequest struct {
	Limit int `form:"limit" binding:"omitempty,min=1,max=100"`
}

// GetModelUsageResponse 模型使用响应
type GetModelUsageResponse struct {
	Usage []ModelUsageItem `json:"usage"`
	Total int              `json:"total"`
}

// GetUserGrowthRequest 获取用户增长请求
type GetUserGrowthRequest struct {
	Days int `form:"days" binding:"omitempty,min=1,max=90"`
}

// GetUserGrowthResponse 用户增长响应
type GetUserGrowthResponse struct {
	Growth []UserGrowthItem `json:"growth"`
	Days   int              `json:"days"`
}

// GetTokenUsageRequest 获取Token使用请求
type GetTokenUsageRequest struct {
	Days int `form:"days" binding:"omitempty,min=1,max=90"`
}

// GetTokenUsageResponse Token使用响应
type GetTokenUsageResponse struct {
	Usage []TokenUsageItem `json:"usage"`
	Days  int              `json:"days"`
	Total int64            `json:"total"`
}
