package quota

import (
	"api-aggregator/backend/internal/domain/user"
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

// Repository 配额仓储接口
type Repository interface {
	// 用户配额相关
	FindUserByID(ctx context.Context, id uint) (*user.User, error)
	UpdateUser(ctx context.Context, user *user.User) error
	UpdateUserQuota(ctx context.Context, userID uint, quota int64) error
	UpdateUserUsedQuota(ctx context.Context, userID uint, usedQuota int64) error
	IncrementUsedQuota(ctx context.Context, userID uint, amount int64) error
	
	// 签到记录相关
	CreateSignInRecord(ctx context.Context, record *SignInRecord) error
	FindTodaySignIn(ctx context.Context, userID uint) (*SignInRecord, error)
	HasSignedInToday(ctx context.Context, userID uint) (bool, error)
	GetSignInHistory(ctx context.Context, userID uint, limit int) ([]*SignInRecord, error)
	
	// 使用统计相关
	GetDailyUsage(ctx context.Context, userID uint, startDate, endDate time.Time) (map[string]int, error)
}

// repository 配额仓储实现
type repository struct {
	db *gorm.DB
}

// NewRepository 创建配额仓储
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// FindUserByID 根据ID查找用户
func (r *repository) FindUserByID(ctx context.Context, id uint) (*user.User, error) {
	var u user.User
	err := r.db.WithContext(ctx).First(&u, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

// UpdateUser 更新用户
func (r *repository) UpdateUser(ctx context.Context, u *user.User) error {
	return r.db.WithContext(ctx).Save(u).Error
}

// UpdateUserQuota 更新用户总配额
func (r *repository) UpdateUserQuota(ctx context.Context, userID uint, quota int64) error {
	return r.db.WithContext(ctx).Model(&user.User{}).
		Where("id = ?", userID).
		Update("quota", quota).Error
}

// UpdateUserUsedQuota 更新用户已使用配额
func (r *repository) UpdateUserUsedQuota(ctx context.Context, userID uint, usedQuota int64) error {
	return r.db.WithContext(ctx).Model(&user.User{}).
		Where("id = ?", userID).
		Update("used_quota", usedQuota).Error
}

// IncrementUsedQuota 增加用户已使用配额
func (r *repository) IncrementUsedQuota(ctx context.Context, userID uint, amount int64) error {
	return r.db.WithContext(ctx).Model(&user.User{}).
		Where("id = ?", userID).
		UpdateColumn("used_quota", gorm.Expr("used_quota + ?", amount)).Error
}

// CreateSignInRecord 创建签到记录
func (r *repository) CreateSignInRecord(ctx context.Context, record *SignInRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

// FindTodaySignIn 查找今天的签到记录
func (r *repository) FindTodaySignIn(ctx context.Context, userID uint) (*SignInRecord, error) {
	var record SignInRecord
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, startOfDay, endOfDay).
		First(&record).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &record, nil
}

// HasSignedInToday 检查今天是否已签到
func (r *repository) HasSignedInToday(ctx context.Context, userID uint) (bool, error) {
	record, err := r.FindTodaySignIn(ctx, userID)
	if err != nil {
		return false, err
	}
	return record != nil, nil
}

// GetSignInHistory 获取签到历史
func (r *repository) GetSignInHistory(ctx context.Context, userID uint, limit int) ([]*SignInRecord, error) {
	var records []*SignInRecord
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&records).Error
	if err != nil {
		return nil, err
	}
	return records, nil
}

// GetDailyUsage 获取每日使用量统计
func (r *repository) GetDailyUsage(ctx context.Context, userID uint, startDate, endDate time.Time) (map[string]int, error) {
	// 这里需要从 request_logs 表查询
	// 由于 request_logs 在另一个模块，这里返回空数据
	// 实际实现时需要注入 RequestLogRepository 或直接查询
	usageMap := make(map[string]int)
	
	// 查询 request_logs 表的每日使用量
	type DailyUsage struct {
		Date   string
		Tokens int
	}
	
	var results []DailyUsage
	err := r.db.WithContext(ctx).
		Table("request_logs").
		Select("DATE(created_at) as date, SUM(tokens_used) as tokens").
		Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, startDate, endDate).
		Group("DATE(created_at)").
		Scan(&results).Error
	
	if err != nil {
		return usageMap, err
	}
	
	for _, result := range results {
		usageMap[result.Date] = result.Tokens
	}
	
	return usageMap, nil
}
