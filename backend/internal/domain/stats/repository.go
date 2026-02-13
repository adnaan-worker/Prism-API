package stats

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Repository 统计仓储接口
type Repository interface {
	// 用户统计
	CountTotalUsers(ctx context.Context) (int64, error)
	CountActiveUsers(ctx context.Context) (int64, error)
	GetUserGrowth(ctx context.Context, startDate, endDate time.Time) ([]UserGrowthItem, error)
	
	// 请求统计
	CountTotalRequests(ctx context.Context) (int64, error)
	CountTodayRequests(ctx context.Context) (int64, error)
	CountSuccessRequests(ctx context.Context) (int64, error)
	CountFailedRequests(ctx context.Context) (int64, error)
	GetRequestTrend(ctx context.Context, startDate, endDate time.Time) ([]RequestTrendItem, error)
	
	// 模型统计
	GetModelUsage(ctx context.Context, limit int) ([]ModelUsageItem, error)
	
	// Token统计
	GetTokenUsage(ctx context.Context, startDate, endDate time.Time) ([]TokenUsageItem, error)
}

// repository 统计仓储实现
type repository struct {
	db *gorm.DB
}

// NewRepository 创建统计仓储
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// CountTotalUsers 统计总用户数
func (r *repository) CountTotalUsers(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("users").Count(&count).Error
	return count, err
}

// CountActiveUsers 统计活跃用户数
func (r *repository) CountActiveUsers(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("users").
		Where("status = ?", "active").
		Count(&count).Error
	return count, err
}

// GetUserGrowth 获取用户增长趋势
func (r *repository) GetUserGrowth(ctx context.Context, startDate, endDate time.Time) ([]UserGrowthItem, error) {
	var results []UserGrowthItem
	err := r.db.WithContext(ctx).
		Table("users").
		Select("TO_CHAR(DATE(created_at), 'YYYY-MM-DD') as date, COUNT(*) as count").
		Where("created_at >= ? AND created_at < ?", startDate, endDate).
		Group("DATE(created_at)").
		Order("DATE(created_at) ASC").
		Scan(&results).Error
	return results, err
}

// CountTotalRequests 统计总请求数
func (r *repository) CountTotalRequests(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("request_logs").Count(&count).Error
	return count, err
}

// CountTodayRequests 统计今日请求数
func (r *repository) CountTodayRequests(ctx context.Context) (int64, error) {
	var count int64
	todayStart := time.Now().Truncate(24 * time.Hour)
	err := r.db.WithContext(ctx).Table("request_logs").
		Where("created_at >= ?", todayStart).
		Count(&count).Error
	return count, err
}

// CountSuccessRequests 统计成功请求数
func (r *repository) CountSuccessRequests(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("request_logs").
		Where("status_code >= ? AND status_code < ?", 200, 300).
		Count(&count).Error
	return count, err
}

// CountFailedRequests 统计失败请求数
func (r *repository) CountFailedRequests(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("request_logs").
		Where("status_code >= ?", 400).
		Count(&count).Error
	return count, err
}

// GetRequestTrend 获取请求趋势
func (r *repository) GetRequestTrend(ctx context.Context, startDate, endDate time.Time) ([]RequestTrendItem, error) {
	var results []RequestTrendItem
	err := r.db.WithContext(ctx).
		Table("request_logs").
		Select("TO_CHAR(DATE(created_at), 'YYYY-MM-DD') as date, COUNT(*) as count").
		Where("created_at >= ? AND created_at < ?", startDate, endDate).
		Group("DATE(created_at)").
		Order("DATE(created_at) ASC").
		Scan(&results).Error
	return results, err
}

// GetModelUsage 获取模型使用统计
func (r *repository) GetModelUsage(ctx context.Context, limit int) ([]ModelUsageItem, error) {
	var results []ModelUsageItem
	err := r.db.WithContext(ctx).
		Table("request_logs").
		Select("model, COUNT(*) as count").
		Where("model != ''").
		Group("model").
		Order("count DESC").
		Limit(limit).
		Scan(&results).Error
	return results, err
}

// GetTokenUsage 获取Token使用统计
func (r *repository) GetTokenUsage(ctx context.Context, startDate, endDate time.Time) ([]TokenUsageItem, error) {
	var results []TokenUsageItem
	err := r.db.WithContext(ctx).
		Table("request_logs").
		Select("TO_CHAR(DATE(created_at), 'YYYY-MM-DD') as date, SUM(tokens_used) as tokens").
		Where("created_at >= ? AND created_at < ?", startDate, endDate).
		Group("DATE(created_at)").
		Order("DATE(created_at) ASC").
		Scan(&results).Error
	return results, err
}
