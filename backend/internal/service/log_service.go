package service

import (
	"api-aggregator/backend/internal/models"
	"api-aggregator/backend/internal/repository"
	"context"
	"fmt"
	"time"
)

type LogService struct {
	requestLogRepo *repository.RequestLogRepository
}

func NewLogService(requestLogRepo *repository.RequestLogRepository) *LogService {
	return &LogService{
		requestLogRepo: requestLogRepo,
	}
}

// GetLogsRequest represents a request to get request logs
type GetLogsRequest struct {
	UserID     *uint      `form:"user_id"`
	Model      string     `form:"model"`
	StatusCode *int       `form:"status_code"`
	StartDate  *time.Time `form:"start_date" time_format:"2006-01-02T15:04:05Z07:00"`
	EndDate    *time.Time `form:"end_date" time_format:"2006-01-02T15:04:05Z07:00"`
	Page       int        `form:"page" binding:"omitempty,min=1"`
	PageSize   int        `form:"page_size" binding:"omitempty,min=1,max=100"`
}

// GetLogsResponse represents a paginated logs response
type GetLogsResponse struct {
	Logs     []*models.RequestLog `json:"logs"`
	Total    int64                `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"page_size"`
}

// GetLogs returns a paginated list of request logs with filters
func (s *LogService) GetLogs(ctx context.Context, req *GetLogsRequest) (*GetLogsResponse, error) {
	// Set default values
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	// Validate parameters
	if req.Page < 1 || req.PageSize < 1 || req.PageSize > 100 {
		return nil, ErrInvalidPage
	}

	// Build filter
	filter := &repository.LogFilter{
		UserID:     req.UserID,
		Model:      req.Model,
		StatusCode: req.StatusCode,
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
		Page:       req.Page,
		PageSize:   req.PageSize,
	}

	// Query logs
	logs, total, err := s.requestLogRepo.FindWithFilters(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs: %w", err)
	}

	return &GetLogsResponse{
		Logs:     logs,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}
