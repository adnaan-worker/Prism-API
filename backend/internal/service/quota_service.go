package service

import (
	"api-aggregator/backend/internal/models"
	"api-aggregator/backend/internal/repository"
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrAlreadySignedIn = errors.New("already signed in today")
)

const (
	DailySignInQuota = 1000
)

type QuotaService struct {
	userRepo   *repository.UserRepository
	signInRepo *repository.SignInRepository
}

func NewQuotaService(userRepo *repository.UserRepository, signInRepo *repository.SignInRepository) *QuotaService {
	return &QuotaService{
		userRepo:   userRepo,
		signInRepo: signInRepo,
	}
}

// QuotaInfo represents user quota information
type QuotaInfo struct {
	TotalQuota     int64      `json:"total_quota"`
	UsedQuota      int64      `json:"used_quota"`
	RemainingQuota int64      `json:"remaining_quota"`
	LastSignIn     *time.Time `json:"last_sign_in,omitempty"`
}

// GetQuotaInfo returns the quota information for a user
func (s *QuotaService) GetQuotaInfo(ctx context.Context, userID uint) (*QuotaInfo, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	return &QuotaInfo{
		TotalQuota:     user.Quota,
		UsedQuota:      user.UsedQuota,
		RemainingQuota: user.Quota - user.UsedQuota,
		LastSignIn:     user.LastSignIn,
	}, nil
}

// SignIn performs daily sign-in and awards quota
func (s *QuotaService) SignIn(ctx context.Context, userID uint) (int, error) {
	// Check if user has already signed in today
	hasSignedIn, err := s.signInRepo.HasSignedInToday(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to check sign-in status: %w", err)
	}
	if hasSignedIn {
		return 0, ErrAlreadySignedIn
	}

	// Get user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return 0, fmt.Errorf("user not found")
	}

	// Update user quota
	user.Quota += DailySignInQuota
	now := time.Now()
	user.LastSignIn = &now

	if err := s.userRepo.Update(ctx, user); err != nil {
		return 0, fmt.Errorf("failed to update user quota: %w", err)
	}

	// Create sign-in record
	record := &models.SignInRecord{
		UserID:       userID,
		QuotaAwarded: DailySignInQuota,
	}
	if err := s.signInRepo.Create(ctx, record); err != nil {
		return 0, fmt.Errorf("failed to create sign-in record: %w", err)
	}

	return DailySignInQuota, nil
}

// DeductQuota deducts quota from a user
func (s *QuotaService) DeductQuota(ctx context.Context, userID uint, amount int64) error {
	if amount < 0 {
		return fmt.Errorf("amount must be non-negative")
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	// Check if user has sufficient quota
	if user.Quota-user.UsedQuota < amount {
		return ErrInsufficientQuota
	}

	// Deduct quota
	user.UsedQuota += amount

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user quota: %w", err)
	}

	return nil
}

// CheckQuota checks if a user has sufficient quota
func (s *QuotaService) CheckQuota(ctx context.Context, userID uint, amount int64) (bool, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return false, fmt.Errorf("user not found")
	}

	return user.Quota-user.UsedQuota >= amount, nil
}

// UsageHistoryItem represents a single day's usage
type UsageHistoryItem struct {
	Date   string `json:"date"`
	Tokens int64  `json:"tokens"`
}

// GetUsageHistory returns usage history for the past N days
func (s *QuotaService) GetUsageHistory(ctx context.Context, userID uint, days int) ([]UsageHistoryItem, error) {
	// TODO: Implement actual usage history from request_logs
	// For now, return mock data
	history := make([]UsageHistoryItem, days)
	now := time.Now()

	for i := 0; i < days; i++ {
		date := now.AddDate(0, 0, -days+i+1)
		history[i] = UsageHistoryItem{
			Date:   date.Format("2006-01-02"),
			Tokens: 0, // Would be calculated from request_logs
		}
	}

	return history, nil
}
