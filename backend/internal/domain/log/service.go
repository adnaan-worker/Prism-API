package log

import (
	"api-aggregator/backend/pkg/errors"
	"api-aggregator/backend/pkg/logger"
	"api-aggregator/backend/pkg/query"
	"context"
	"time"
)

// Service 日志服务接口
type Service interface {
	CreateLog(ctx context.Context, req *CreateLogRequest) error
	GetLogs(ctx context.Context, req *GetLogsRequest) (*LogListResponse, error)
	GetLogStats(ctx context.Context, startDate, endDate *time.Time) (*LogStatsResponse, error)
	DeleteOldLogs(ctx context.Context, days int) (int64, error)
}

// service 日志服务实现
type service struct {
	repo   Repository
	logger logger.Logger
}

// NewService 创建日志服务
func NewService(repo Repository, logger logger.Logger) Service {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

// CreateLog 创建日志
func (s *service) CreateLog(ctx context.Context, req *CreateLogRequest) error {
	log := &RequestLog{
		UserID:       req.UserID,
		APIKeyID:     req.APIKeyID,
		APIConfigID:  req.APIConfigID,
		Model:        req.Model,
		Method:       req.Method,
		Path:         req.Path,
		StatusCode:   req.StatusCode,
		ResponseTime: req.ResponseTime,
		TokensUsed:   req.TokensUsed,
		ErrorMsg:     req.ErrorMsg,
	}

	if err := s.repo.Create(ctx, log); err != nil {
		s.logger.Error("Failed to create log",
			logger.Uint("user_id", req.UserID),
			logger.String("model", req.Model),
			logger.Error(err))
		return errors.Wrap(err, 500002, "Failed to create log")
	}

	return nil
}

// GetLogs 获取日志列表
func (s *service) GetLogs(ctx context.Context, req *GetLogsRequest) (*LogListResponse, error) {
	// 设置默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	// 构建过滤条件
	var filters []query.Filter
	if req.UserID != nil {
		filters = append(filters, query.Filter{
			Field:    "user_id",
			Operator: "=",
			Value:    *req.UserID,
		})
	}
	if req.Model != "" {
		filters = append(filters, query.Filter{
			Field:    "model",
			Operator: "=",
			Value:    req.Model,
		})
	}
	if req.StatusCode != nil {
		filters = append(filters, query.Filter{
			Field:    "status_code",
			Operator: "=",
			Value:    *req.StatusCode,
		})
	}
	if req.StartDate != nil {
		filters = append(filters, query.Filter{
			Field:    "created_at",
			Operator: ">=",
			Value:    *req.StartDate,
		})
	}
	if req.EndDate != nil {
		filters = append(filters, query.Filter{
			Field:    "created_at",
			Operator: "<=",
			Value:    *req.EndDate,
		})
	}

	// 构建排序
	sorts := []query.Sort{
		{Field: "created_at", Desc: true},
	}

	// 构建分页
	pagination := &query.Pagination{
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	// 查询日志列表
	logs, total, err := s.repo.List(ctx, filters, sorts, pagination)
	if err != nil {
		s.logger.Error("Failed to get logs", logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to get logs")
	}

	return &LogListResponse{
		Logs:     ToResponseList(logs),
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// GetLogStats 获取日志统计
func (s *service) GetLogStats(ctx context.Context, startDate, endDate *time.Time) (*LogStatsResponse, error) {
	stats, err := s.repo.GetStats(ctx, startDate, endDate)
	if err != nil {
		s.logger.Error("Failed to get log stats", logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to get log stats")
	}

	return stats, nil
}

// DeleteOldLogs 删除旧日志
func (s *service) DeleteOldLogs(ctx context.Context, days int) (int64, error) {
	if days <= 0 {
		return 0, errors.ErrInvalidParam.WithDetails("Days must be positive")
	}

	before := time.Now().AddDate(0, 0, -days)
	deleted, err := s.repo.DeleteOldLogs(ctx, before)
	if err != nil {
		s.logger.Error("Failed to delete old logs",
			logger.Int("days", days),
			logger.Error(err))
		return 0, errors.Wrap(err, 500002, "Failed to delete old logs")
	}

	s.logger.Info("Old logs deleted successfully",
		logger.Int("days", days),
		logger.Int64("deleted", deleted))

	return deleted, nil
}
