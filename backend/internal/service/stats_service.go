package service

import (
	"api-aggregator/backend/internal/models"
	"api-aggregator/backend/internal/repository"
	"context"
	"fmt"
	"time"
)

type StatsService struct {
	userRepo       *repository.UserRepository
	requestLogRepo *repository.RequestLogRepository
}

func NewStatsService(userRepo *repository.UserRepository, requestLogRepo *repository.RequestLogRepository) *StatsService {
	return &StatsService{
		userRepo:       userRepo,
		requestLogRepo: requestLogRepo,
	}
}

// StatsOverview represents the statistics overview
type StatsOverview struct {
	TotalUsers      int64 `json:"total_users"`
	ActiveUsers     int64 `json:"active_users"`
	TotalRequests   int64 `json:"total_requests"`
	TodayRequests   int64 `json:"today_requests"`
	SuccessRequests int64 `json:"success_requests"`
	FailedRequests  int64 `json:"failed_requests"`
}

// GetStatsOverview returns the statistics overview
func (s *StatsService) GetStatsOverview(ctx context.Context) (*StatsOverview, error) {
	// Get total users count
	totalUsers, err := s.userRepo.CountAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count total users: %w", err)
	}

	// Get active users count (status = 'active')
	activeUsers, err := s.userRepo.CountByStatus(ctx, "active")
	if err != nil {
		return nil, fmt.Errorf("failed to count active users: %w", err)
	}

	// Get total requests count
	totalRequests, err := s.requestLogRepo.CountAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count total requests: %w", err)
	}

	// Get today's requests count
	todayStart := time.Now().Truncate(24 * time.Hour)
	todayRequests, err := s.requestLogRepo.CountByDateRange(ctx, todayStart, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to count today requests: %w", err)
	}

	// Get success requests count (status code 200-299)
	successRequests, err := s.requestLogRepo.CountByStatusCodeRange(ctx, 200, 299)
	if err != nil {
		return nil, fmt.Errorf("failed to count success requests: %w", err)
	}

	// Get failed requests count (status code >= 400)
	failedRequests, err := s.requestLogRepo.CountByStatusCodeRange(ctx, 400, 599)
	if err != nil {
		return nil, fmt.Errorf("failed to count failed requests: %w", err)
	}

	return &StatsOverview{
		TotalUsers:      totalUsers,
		ActiveUsers:     activeUsers,
		TotalRequests:   totalRequests,
		TodayRequests:   todayRequests,
		SuccessRequests: successRequests,
		FailedRequests:  failedRequests,
	}, nil
}

// RequestTrendItem represents a single day's request count
type RequestTrendItem struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// GetRequestTrend returns request trend for the past N days
func (s *StatsService) GetRequestTrend(ctx context.Context, days int) ([]RequestTrendItem, error) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days+1).Truncate(24 * time.Hour)
	
	// Query daily request counts
	var results []struct {
		Date  string
		Count int64
	}
	
	err := s.requestLogRepo.GetDB().WithContext(ctx).
		Model(&models.RequestLog{}).
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("created_at >= ? AND created_at < ?", 
			startDate.Format("2006-01-02"), 
			endDate.AddDate(0, 0, 1).Format("2006-01-02")).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&results).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to query request trend: %w", err)
	}
	
	// Create a map for quick lookup
	countMap := make(map[string]int64)
	for _, r := range results {
		countMap[r.Date] = r.Count
	}
	
	// Fill in all days (including days with 0 requests)
	trend := make([]RequestTrendItem, days)
	for i := 0; i < days; i++ {
		date := startDate.AddDate(0, 0, i)
		dateStr := date.Format("2006-01-02")
		trend[i] = RequestTrendItem{
			Date:  dateStr,
			Count: countMap[dateStr],
		}
	}
	
	return trend, nil
}

// ModelUsageItem represents model usage statistics
type ModelUsageItem struct {
	Model string `json:"model"`
	Count int64  `json:"count"`
}

// GetModelUsage returns model usage statistics
func (s *StatsService) GetModelUsage(ctx context.Context, limit int) ([]ModelUsageItem, error) {
	if limit <= 0 {
		limit = 10
	}
	
	var results []ModelUsageItem
	
	err := s.requestLogRepo.GetDB().WithContext(ctx).
		Model(&models.RequestLog{}).
		Select("model, COUNT(*) as count").
		Where("model != ''").
		Group("model").
		Order("count DESC").
		Limit(limit).
		Scan(&results).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to query model usage: %w", err)
	}
	
	return results, nil
}
