package repository

import (
	"api-aggregator/backend/internal/models"
	"context"
	"time"

	"gorm.io/gorm"
)

type RequestLogRepository struct {
	db *gorm.DB
}

func NewRequestLogRepository(db *gorm.DB) *RequestLogRepository {
	return &RequestLogRepository{db: db}
}

// Create creates a new request log
func (r *RequestLogRepository) Create(ctx context.Context, log *models.RequestLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// FindByUserID finds request logs by user ID with pagination
func (r *RequestLogRepository) FindByUserID(ctx context.Context, userID uint, limit, offset int) ([]*models.RequestLog, error) {
	var logs []*models.RequestLog
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error
	if err != nil {
		return nil, err
	}
	return logs, nil
}

// CountByUserID counts request logs by user ID
func (r *RequestLogRepository) CountByUserID(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.RequestLog{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}

// CountAll counts all request logs
func (r *RequestLogRepository) CountAll(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.RequestLog{}).Count(&count).Error
	return count, err
}

// CountByDateRange counts request logs within a date range
func (r *RequestLogRepository) CountByDateRange(ctx context.Context, start, end time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.RequestLog{}).
		Where("created_at >= ? AND created_at <= ?", start, end).
		Count(&count).Error
	return count, err
}

// CountByStatusCodeRange counts request logs by status code range
func (r *RequestLogRepository) CountByStatusCodeRange(ctx context.Context, minCode, maxCode int) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.RequestLog{}).
		Where("status_code >= ? AND status_code <= ?", minCode, maxCode).
		Count(&count).Error
	return count, err
}

// LogFilter represents filters for querying request logs
type LogFilter struct {
	UserID     *uint
	Model      string
	StatusCode *int
	StartDate  *time.Time
	EndDate    *time.Time
	Page       int
	PageSize   int
}

// FindWithFilters finds request logs with filters and pagination
func (r *RequestLogRepository) FindWithFilters(ctx context.Context, filter *LogFilter) ([]*models.RequestLog, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.RequestLog{})

	// Apply filters
	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}
	if filter.Model != "" {
		query = query.Where("model = ?", filter.Model)
	}
	if filter.StatusCode != nil {
		query = query.Where("status_code = ?", *filter.StatusCode)
	}
	if filter.StartDate != nil {
		query = query.Where("created_at >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("created_at <= ?", *filter.EndDate)
	}

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	var logs []*models.RequestLog
	offset := (filter.Page - 1) * filter.PageSize
	err := query.
		Order("created_at DESC").
		Limit(filter.PageSize).
		Offset(offset).
		Find(&logs).Error

	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetDailyUsage returns daily token usage for a user within a date range
func (r *RequestLogRepository) GetDailyUsage(ctx context.Context, userID uint, startDate, endDate time.Time) (map[string]int, error) {
	var results []struct {
		Date   string
		Tokens int
	}
	
	err := r.db.WithContext(ctx).
		Table("request_logs").
		Select("DATE(created_at) as date, SUM(tokens_used) as tokens").
		Where("user_id = ? AND created_at >= ? AND created_at < ?", 
			userID, 
			startDate.Format("2006-01-02"), 
			endDate.AddDate(0, 0, 1).Format("2006-01-02")).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&results).Error
	
	if err != nil {
		return nil, err
	}
	
	// Convert to map
	usageMap := make(map[string]int)
	for _, r := range results {
		usageMap[r.Date] = r.Tokens
	}
	
	return usageMap, nil
}
