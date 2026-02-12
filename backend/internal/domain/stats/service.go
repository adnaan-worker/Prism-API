package stats

import (
	"api-aggregator/backend/pkg/errors"
	"api-aggregator/backend/pkg/logger"
	"context"
	"time"
)

// Service 统计服务接口
type Service interface {
	GetStatsOverview(ctx context.Context) (*GetStatsOverviewResponse, error)
	GetRequestTrend(ctx context.Context, req *GetRequestTrendRequest) (*GetRequestTrendResponse, error)
	GetModelUsage(ctx context.Context, req *GetModelUsageRequest) (*GetModelUsageResponse, error)
	GetUserGrowth(ctx context.Context, req *GetUserGrowthRequest) (*GetUserGrowthResponse, error)
	GetTokenUsage(ctx context.Context, req *GetTokenUsageRequest) (*GetTokenUsageResponse, error)
}

// service 统计服务实现
type service struct {
	repo   Repository
	logger logger.Logger
}

// NewService 创建统计服务
func NewService(repo Repository, logger logger.Logger) Service {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

// GetStatsOverview 获取统计概览
func (s *service) GetStatsOverview(ctx context.Context) (*GetStatsOverviewResponse, error) {
	// 获取总用户数
	totalUsers, err := s.repo.CountTotalUsers(ctx)
	if err != nil {
		s.logger.Error("Failed to count total users", logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to count total users")
	}

	// 获取活跃用户数
	activeUsers, err := s.repo.CountActiveUsers(ctx)
	if err != nil {
		s.logger.Error("Failed to count active users", logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to count active users")
	}

	// 获取总请求数
	totalRequests, err := s.repo.CountTotalRequests(ctx)
	if err != nil {
		s.logger.Error("Failed to count total requests", logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to count total requests")
	}

	// 获取今日请求数
	todayRequests, err := s.repo.CountTodayRequests(ctx)
	if err != nil {
		s.logger.Error("Failed to count today requests", logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to count today requests")
	}

	// 获取成功请求数
	successRequests, err := s.repo.CountSuccessRequests(ctx)
	if err != nil {
		s.logger.Error("Failed to count success requests", logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to count success requests")
	}

	// 获取失败请求数
	failedRequests, err := s.repo.CountFailedRequests(ctx)
	if err != nil {
		s.logger.Error("Failed to count failed requests", logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to count failed requests")
	}

	return &GetStatsOverviewResponse{
		TotalUsers:      totalUsers,
		ActiveUsers:     activeUsers,
		TotalRequests:   totalRequests,
		TodayRequests:   todayRequests,
		SuccessRequests: successRequests,
		FailedRequests:  failedRequests,
	}, nil
}

// GetRequestTrend 获取请求趋势
func (s *service) GetRequestTrend(ctx context.Context, req *GetRequestTrendRequest) (*GetRequestTrendResponse, error) {
	// 设置默认值
	days := req.Days
	if days == 0 {
		days = 7
	}

	// 计算日期范围
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days+1).Truncate(24 * time.Hour)

	// 获取趋势数据
	trendData, err := s.repo.GetRequestTrend(ctx, startDate, endDate)
	if err != nil {
		s.logger.Error("Failed to get request trend", logger.Int("days", days), logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to get request trend")
	}

	// 创建映射以便快速查找
	trendMap := make(map[string]int64)
	for _, item := range trendData {
		trendMap[item.Date] = item.Count
	}

	// 填充所有日期（包括请求数为0的日期）
	trend := make([]RequestTrendItem, days)
	for i := 0; i < days; i++ {
		date := startDate.AddDate(0, 0, i)
		dateStr := date.Format("2006-01-02")
		trend[i] = RequestTrendItem{
			Date:  dateStr,
			Count: trendMap[dateStr],
		}
	}

	return &GetRequestTrendResponse{
		Trend: trend,
		Days:  days,
	}, nil
}

// GetModelUsage 获取模型使用统计
func (s *service) GetModelUsage(ctx context.Context, req *GetModelUsageRequest) (*GetModelUsageResponse, error) {
	// 设置默认值
	limit := req.Limit
	if limit == 0 {
		limit = 10
	}

	// 获取模型使用数据
	usage, err := s.repo.GetModelUsage(ctx, limit)
	if err != nil {
		s.logger.Error("Failed to get model usage", logger.Int("limit", limit), logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to get model usage")
	}

	return &GetModelUsageResponse{
		Usage: usage,
		Total: len(usage),
	}, nil
}

// GetUserGrowth 获取用户增长趋势
func (s *service) GetUserGrowth(ctx context.Context, req *GetUserGrowthRequest) (*GetUserGrowthResponse, error) {
	// 设置默认值
	days := req.Days
	if days == 0 {
		days = 30
	}

	// 计算日期范围
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days+1).Truncate(24 * time.Hour)

	// 获取用户增长数据
	growthData, err := s.repo.GetUserGrowth(ctx, startDate, endDate)
	if err != nil {
		s.logger.Error("Failed to get user growth", logger.Int("days", days), logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to get user growth")
	}

	// 创建映射以便快速查找
	growthMap := make(map[string]int64)
	for _, item := range growthData {
		growthMap[item.Date] = item.Count
	}

	// 填充所有日期
	growth := make([]UserGrowthItem, days)
	for i := 0; i < days; i++ {
		date := startDate.AddDate(0, 0, i)
		dateStr := date.Format("2006-01-02")
		growth[i] = UserGrowthItem{
			Date:  dateStr,
			Count: growthMap[dateStr],
		}
	}

	return &GetUserGrowthResponse{
		Growth: growth,
		Days:   days,
	}, nil
}

// GetTokenUsage 获取Token使用统计
func (s *service) GetTokenUsage(ctx context.Context, req *GetTokenUsageRequest) (*GetTokenUsageResponse, error) {
	// 设置默认值
	days := req.Days
	if days == 0 {
		days = 7
	}

	// 计算日期范围
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days+1).Truncate(24 * time.Hour)

	// 获取Token使用数据
	usageData, err := s.repo.GetTokenUsage(ctx, startDate, endDate)
	if err != nil {
		s.logger.Error("Failed to get token usage", logger.Int("days", days), logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to get token usage")
	}

	// 创建映射以便快速查找
	usageMap := make(map[string]int64)
	var total int64
	for _, item := range usageData {
		usageMap[item.Date] = item.Tokens
		total += item.Tokens
	}

	// 填充所有日期
	usage := make([]TokenUsageItem, days)
	for i := 0; i < days; i++ {
		date := startDate.AddDate(0, 0, i)
		dateStr := date.Format("2006-01-02")
		usage[i] = TokenUsageItem{
			Date:   dateStr,
			Tokens: usageMap[dateStr],
		}
	}

	return &GetTokenUsageResponse{
		Usage: usage,
		Days:  days,
		Total: total,
	}, nil
}
