package service

import (
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
