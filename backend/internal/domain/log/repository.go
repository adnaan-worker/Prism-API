package log

import (
	"api-aggregator/backend/pkg/query"
	"context"
	"time"

	"gorm.io/gorm"
)

// Repository 日志仓储接口
type Repository interface {
	Create(ctx context.Context, log *RequestLog) error
	FindByID(ctx context.Context, id uint) (*RequestLog, error)
	FindByUserID(ctx context.Context, userID uint, limit, offset int) ([]*RequestLog, error)
	List(ctx context.Context, filters []query.Filter, sorts []query.Sort, pagination *query.Pagination) ([]*RequestLog, int64, error)
	CountAll(ctx context.Context) (int64, error)
	CountByUserID(ctx context.Context, userID uint) (int64, error)
	CountByDateRange(ctx context.Context, start, end time.Time) (int64, error)
	CountByStatusCodeRange(ctx context.Context, minCode, maxCode int) (int64, error)
	GetDailyUsage(ctx context.Context, userID uint, startDate, endDate time.Time) (map[string]int, error)
	GetStats(ctx context.Context, startDate, endDate *time.Time) (*LogStatsResponse, error)
	DeleteOldLogs(ctx context.Context, before time.Time) (int64, error)
}

// repository 日志仓储实现
type repository struct {
	db *gorm.DB
}

// NewRepository 创建日志仓储
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// Create 创建日志
func (r *repository) Create(ctx context.Context, log *RequestLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// FindByID 根据ID查找日志
func (r *repository) FindByID(ctx context.Context, id uint) (*RequestLog, error) {
	var log RequestLog
	err := r.db.WithContext(ctx).First(&log, id).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

// FindByUserID 根据用户ID查找日志
func (r *repository) FindByUserID(ctx context.Context, userID uint, limit, offset int) ([]*RequestLog, error) {
	var logs []*RequestLog
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

// List 查询日志列表（支持过滤、排序、分页）
func (r *repository) List(ctx context.Context, filters []query.Filter, sorts []query.Sort, pagination *query.Pagination) ([]*RequestLog, int64, error) {
	builder := query.NewBuilder(r.db.WithContext(ctx).Model(&RequestLog{}))
	
	// 应用过滤
	builder.ApplyFilters(filters)
	
	// 统计总数
	var total int64
	builder.Count(&total)
	
	// 应用排序和分页
	if len(sorts) == 0 {
		sorts = []query.Sort{
			{Field: "created_at", Desc: true},
		}
	}
	builder.ApplySort(sorts).ApplyPagination(pagination)
	
	// 查询结果
	var logs []*RequestLog
	err := builder.Find(&logs)
	if err != nil {
		return nil, 0, err
	}
	
	return logs, total, nil
}

// CountAll 统计所有日志数量
func (r *repository) CountAll(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&RequestLog{}).Count(&count).Error
	return count, err
}

// CountByUserID 根据用户ID统计日志数量
func (r *repository) CountByUserID(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&RequestLog{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}

// CountByDateRange 根据日期范围统计日志数量
func (r *repository) CountByDateRange(ctx context.Context, start, end time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&RequestLog{}).
		Where("created_at >= ? AND created_at <= ?", start, end).
		Count(&count).Error
	return count, err
}

// CountByStatusCodeRange 根据状态码范围统计日志数量
func (r *repository) CountByStatusCodeRange(ctx context.Context, minCode, maxCode int) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&RequestLog{}).
		Where("status_code >= ? AND status_code <= ?", minCode, maxCode).
		Count(&count).Error
	return count, err
}

// GetDailyUsage 获取每日使用量统计
func (r *repository) GetDailyUsage(ctx context.Context, userID uint, startDate, endDate time.Time) (map[string]int, error) {
	var results []struct {
		Date   string
		Tokens int
	}
	
	err := r.db.WithContext(ctx).
		Table("request_logs").
		Select("DATE(created_at) as date, SUM(tokens_used) as tokens").
		Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, startDate, endDate).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&results).Error
	
	if err != nil {
		return nil, err
	}
	
	usageMap := make(map[string]int)
	for _, r := range results {
		usageMap[r.Date] = r.Tokens
	}
	
	return usageMap, nil
}

// GetStats 获取日志统计
func (r *repository) GetStats(ctx context.Context, startDate, endDate *time.Time) (*LogStatsResponse, error) {
	query := r.db.WithContext(ctx).Model(&RequestLog{})
	
	if startDate != nil {
		query = query.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", *endDate)
	}
	
	var stats struct {
		TotalRequests   int64
		SuccessRequests int64
		ErrorRequests   int64
		AvgResponseTime float64
		TotalTokens     int64
	}
	
	// 总请求数
	query.Count(&stats.TotalRequests)
	
	// 成功请求数
	r.db.WithContext(ctx).Model(&RequestLog{}).
		Where("status_code >= ? AND status_code < ?", 200, 300).
		Count(&stats.SuccessRequests)
	
	// 错误请求数
	r.db.WithContext(ctx).Model(&RequestLog{}).
		Where("status_code >= ?", 400).
		Count(&stats.ErrorRequests)
	
	// 平均响应时间
	r.db.WithContext(ctx).Model(&RequestLog{}).
		Select("AVG(response_time)").
		Scan(&stats.AvgResponseTime)
	
	// 总token数
	r.db.WithContext(ctx).Model(&RequestLog{}).
		Select("SUM(tokens_used)").
		Scan(&stats.TotalTokens)
	
	return &LogStatsResponse{
		TotalRequests:   stats.TotalRequests,
		SuccessRequests: stats.SuccessRequests,
		ErrorRequests:   stats.ErrorRequests,
		AvgResponseTime: stats.AvgResponseTime,
		TotalTokens:     stats.TotalTokens,
	}, nil
}

// DeleteOldLogs 删除旧日志
func (r *repository) DeleteOldLogs(ctx context.Context, before time.Time) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("created_at < ?", before).
		Delete(&RequestLog{})
	return result.RowsAffected, result.Error
}
