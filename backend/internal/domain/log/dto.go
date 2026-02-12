package log

import "time"

// CreateLogRequest 创建日志请求
type CreateLogRequest struct {
	UserID       uint   `json:"user_id" binding:"required"`
	APIKeyID     uint   `json:"api_key_id" binding:"required"`
	APIConfigID  uint   `json:"api_config_id" binding:"required"`
	Model        string `json:"model" binding:"required"`
	Method       string `json:"method" binding:"required"`
	Path         string `json:"path" binding:"required"`
	StatusCode   int    `json:"status_code" binding:"required"`
	ResponseTime int    `json:"response_time" binding:"required,min=0"`
	TokensUsed   int    `json:"tokens_used" binding:"omitempty,min=0"`
	ErrorMsg     string `json:"error_msg" binding:"omitempty"`
}

// GetLogsRequest 获取日志列表请求
type GetLogsRequest struct {
	Page       int        `form:"page" binding:"omitempty,min=1"`
	PageSize   int        `form:"page_size" binding:"omitempty,min=1,max=100"`
	UserID     *uint      `form:"user_id" binding:"omitempty"`
	Model      string     `form:"model" binding:"omitempty"`
	StatusCode *int       `form:"status_code" binding:"omitempty"`
	StartDate  *time.Time `form:"start_date" binding:"omitempty" time_format:"2006-01-02T15:04:05Z07:00"`
	EndDate    *time.Time `form:"end_date" binding:"omitempty" time_format:"2006-01-02T15:04:05Z07:00"`
}

// LogResponse 日志响应
type LogResponse struct {
	ID           uint      `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UserID       uint      `json:"user_id"`
	APIKeyID     uint      `json:"api_key_id"`
	APIConfigID  uint      `json:"api_config_id"`
	Model        string    `json:"model"`
	Method       string    `json:"method"`
	Path         string    `json:"path"`
	StatusCode   int       `json:"status_code"`
	ResponseTime int       `json:"response_time"`
	TokensUsed   int       `json:"tokens_used"`
	ErrorMsg     string    `json:"error_msg,omitempty"`
}

// LogListResponse 日志列表响应
type LogListResponse struct {
	Logs     []*LogResponse `json:"logs"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

// LogStatsResponse 日志统计响应
type LogStatsResponse struct {
	TotalRequests   int64   `json:"total_requests"`
	SuccessRequests int64   `json:"success_requests"`
	ErrorRequests   int64   `json:"error_requests"`
	AvgResponseTime float64 `json:"avg_response_time"`
	TotalTokens     int64   `json:"total_tokens"`
}

// ToResponse 转换为响应对象
func (l *RequestLog) ToResponse() *LogResponse {
	return &LogResponse{
		ID:           l.ID,
		CreatedAt:    l.CreatedAt,
		UserID:       l.UserID,
		APIKeyID:     l.APIKeyID,
		APIConfigID:  l.APIConfigID,
		Model:        l.Model,
		Method:       l.Method,
		Path:         l.Path,
		StatusCode:   l.StatusCode,
		ResponseTime: l.ResponseTime,
		TokensUsed:   l.TokensUsed,
		ErrorMsg:     l.ErrorMsg,
	}
}

// ToResponseList 批量转换为响应对象
func ToResponseList(logs []*RequestLog) []*LogResponse {
	responses := make([]*LogResponse, len(logs))
	for i, log := range logs {
		responses[i] = log.ToResponse()
	}
	return responses
}
