package quota

import (
	"api-aggregator/backend/pkg/errors"
	"api-aggregator/backend/pkg/logger"
	"context"
	"time"
)

// Service 配额服务接口
type Service interface {
	GetQuotaInfo(ctx context.Context, userID uint) (*QuotaInfoResponse, error)
	SignIn(ctx context.Context, userID uint) (*SignInResponse, error)
	DeductQuota(ctx context.Context, userID uint, amount int64) error
	CheckQuota(ctx context.Context, userID uint, amount int64) (*CheckQuotaResponse, error)
	GetUsageHistory(ctx context.Context, userID uint, days int) (*UsageHistoryResponse, error)
}

// service 配额服务实现
type service struct {
	repo   Repository
	logger logger.Logger
}

// NewService 创建配额服务
func NewService(repo Repository, logger logger.Logger) Service {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

// GetQuotaInfo 获取配额信息
func (s *service) GetQuotaInfo(ctx context.Context, userID uint) (*QuotaInfoResponse, error) {
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to find user", logger.Uint("user_id", userID), logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to find user")
	}
	if user == nil {
		return nil, errors.ErrUserNotFound
	}

	remainingQuota := user.Quota - user.UsedQuota
	if remainingQuota < 0 {
		remainingQuota = 0
	}

	return &QuotaInfoResponse{
		TotalQuota:     user.Quota,
		UsedQuota:      user.UsedQuota,
		RemainingQuota: remainingQuota,
		LastSignIn:     user.LastSignIn,
	}, nil
}

// SignIn 每日签到
func (s *service) SignIn(ctx context.Context, userID uint) (*SignInResponse, error) {
	// 检查今天是否已签到
	hasSignedIn, err := s.repo.HasSignedInToday(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to check sign-in status",
			logger.Uint("user_id", userID),
			logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to check sign-in status")
	}
	if hasSignedIn {
		return nil, errors.New(409001, "Already signed in today")
	}

	// 获取用户
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to find user", logger.Uint("user_id", userID), logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to find user")
	}
	if user == nil {
		return nil, errors.ErrUserNotFound
	}

	// 增加配额
	user.Quota += DailySignInQuota
	now := time.Now()
	user.LastSignIn = &now

	if err := s.repo.UpdateUser(ctx, user); err != nil {
		s.logger.Error("Failed to update user quota",
			logger.Uint("user_id", userID),
			logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to update user quota")
	}

	// 创建签到记录
	record := &SignInRecord{
		UserID:       userID,
		QuotaAwarded: DailySignInQuota,
	}
	if err := s.repo.CreateSignInRecord(ctx, record); err != nil {
		s.logger.Error("Failed to create sign-in record",
			logger.Uint("user_id", userID),
			logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to create sign-in record")
	}

	s.logger.Info("User signed in successfully",
		logger.Uint("user_id", userID),
		logger.Int("quota_awarded", DailySignInQuota))

	remainingQuota := user.Quota - user.UsedQuota
	if remainingQuota < 0 {
		remainingQuota = 0
	}

	return &SignInResponse{
		QuotaAwarded:   DailySignInQuota,
		TotalQuota:     user.Quota,
		RemainingQuota: remainingQuota,
		SignInDate:     now,
	}, nil
}

// DeductQuota 扣除配额
func (s *service) DeductQuota(ctx context.Context, userID uint, amount int64) error {
	if amount < 0 {
		return errors.ErrInvalidParam.WithDetails("Amount must be non-negative")
	}

	// 获取用户
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to find user", logger.Uint("user_id", userID), logger.Error(err))
		return errors.Wrap(err, 500002, "Failed to find user")
	}
	if user == nil {
		return errors.ErrUserNotFound
	}

	// 检查配额是否充足
	if user.Quota-user.UsedQuota < amount {
		return errors.ErrQuotaExceeded
	}

	// 扣除配额
	if err := s.repo.IncrementUsedQuota(ctx, userID, amount); err != nil {
		s.logger.Error("Failed to deduct quota",
			logger.Uint("user_id", userID),
			logger.Int64("amount", amount),
			logger.Error(err))
		return errors.Wrap(err, 500002, "Failed to deduct quota")
	}

	s.logger.Info("Quota deducted successfully",
		logger.Uint("user_id", userID),
		logger.Int64("amount", amount))

	return nil
}

// CheckQuota 检查配额是否充足
func (s *service) CheckQuota(ctx context.Context, userID uint, amount int64) (*CheckQuotaResponse, error) {
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to find user", logger.Uint("user_id", userID), logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to find user")
	}
	if user == nil {
		return nil, errors.ErrUserNotFound
	}

	remainingQuota := user.Quota - user.UsedQuota
	if remainingQuota < 0 {
		remainingQuota = 0
	}

	return &CheckQuotaResponse{
		HasSufficientQuota: remainingQuota >= amount,
		RemainingQuota:     remainingQuota,
		RequiredAmount:     amount,
	}, nil
}

// GetUsageHistory 获取使用历史
func (s *service) GetUsageHistory(ctx context.Context, userID uint, days int) (*UsageHistoryResponse, error) {
	// 设置默认值
	if days == 0 {
		days = 7
	}
	if days > 90 {
		days = 90
	}

	// 计算日期范围
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days+1)

	// 获取使用统计
	usageMap, err := s.repo.GetDailyUsage(ctx, userID, startDate, endDate)
	if err != nil {
		s.logger.Error("Failed to get usage history",
			logger.Uint("user_id", userID),
			logger.Int("days", days),
			logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to get usage history")
	}

	// 填充所有日期（包括使用量为0的日期）
	history := make([]UsageHistoryItem, days)
	for i := 0; i < days; i++ {
		date := startDate.AddDate(0, 0, i)
		dateStr := date.Format("2006-01-02")
		history[i] = UsageHistoryItem{
			Date:   dateStr,
			Tokens: int64(usageMap[dateStr]),
		}
	}

	return &UsageHistoryResponse{
		History: history,
		Days:    days,
	}, nil
}
