package accountpool

import (
	"api-aggregator/backend/pkg/query"
	"context"

	"gorm.io/gorm"
)

// Repository 账号池仓储接口
type Repository interface {
	Create(ctx context.Context, pool *AccountPool) error
	Update(ctx context.Context, pool *AccountPool) error
	Delete(ctx context.Context, id uint) error
	FindByID(ctx context.Context, id uint) (*AccountPool, error)
	FindByProvider(ctx context.Context, provider string) ([]*AccountPool, error)
	List(ctx context.Context, filter *PoolFilter, opts *query.Options) ([]*AccountPool, int64, error)
	FindAll(ctx context.Context) ([]*AccountPool, error)
	IncrementRequests(ctx context.Context, id uint) error
	IncrementErrors(ctx context.Context, id uint) error
	
	// 凭据相关
	CreateCredential(ctx context.Context, cred *AccountCredential) error
	UpdateCredential(ctx context.Context, cred *AccountCredential) error
	DeleteCredential(ctx context.Context, id uint) error
	FindCredentialByID(ctx context.Context, id uint) (*AccountCredential, error)
	FindActiveCredentialsByPoolID(ctx context.Context, poolID uint) ([]*AccountCredential, error)
	ListCredentials(ctx context.Context, filter *CredentialFilter, opts *query.Options) ([]*AccountCredential, int64, error)
	UpdateCredentialStatus(ctx context.Context, id uint, isActive bool) error
	IncrementCredentialRequests(ctx context.Context, id uint) error
	IncrementCredentialErrors(ctx context.Context, id uint) error
	
	// 请求日志相关
	CreateRequestLog(ctx context.Context, log *AccountPoolRequestLog) error
	ListRequestLogs(ctx context.Context, filter *RequestLogFilter, opts *query.Options) ([]*AccountPoolRequestLog, int64, error)
	GetPoolRequestStats(ctx context.Context, poolID uint) (map[string]interface{}, error)
}

type repository struct {
	db *gorm.DB
}

// NewRepository 创建账号池仓储实例
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// Create 创建账号池
func (r *repository) Create(ctx context.Context, pool *AccountPool) error {
	return r.db.WithContext(ctx).Create(pool).Error
}

// Update 更新账号池
func (r *repository) Update(ctx context.Context, pool *AccountPool) error {
	return r.db.WithContext(ctx).Save(pool).Error
}

// Delete 删除账号池
func (r *repository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&AccountPool{}, id).Error
}

// FindByID 根据ID查找账号池
func (r *repository) FindByID(ctx context.Context, id uint) (*AccountPool, error) {
	var pool AccountPool
	err := r.db.WithContext(ctx).First(&pool, id).Error
	if err != nil {
		return nil, err
	}
	return &pool, nil
}

// FindByProvider 根据提供商查找账号池
func (r *repository) FindByProvider(ctx context.Context, provider string) ([]*AccountPool, error) {
	var pools []*AccountPool
	err := r.db.WithContext(ctx).
		Where("provider_type = ?", provider).
		Find(&pools).Error
	return pools, err
}

// List 查询账号池列表
func (r *repository) List(ctx context.Context, filter *PoolFilter, opts *query.Options) ([]*AccountPool, int64, error) {
	var pools []*AccountPool
	var total int64

	db := r.db.WithContext(ctx).Model(&AccountPool{})

	// 应用过滤器
	if filter != nil {
		if filter.Provider != nil {
			db = db.Where("provider_type = ?", *filter.Provider)
		}
		if filter.Strategy != nil {
			db = db.Where("strategy = ?", *filter.Strategy)
		}
		if filter.IsActive != nil {
			db = db.Where("is_active = ?", *filter.IsActive)
		}
	}

	// 计算总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用查询选项
	if opts != nil {
		db = query.ApplyOptions(db, opts)
	}

	// 查询数据
	if err := db.Find(&pools).Error; err != nil {
		return nil, 0, err
	}

	return pools, total, nil
}

// FindAll 查询所有账号池
func (r *repository) FindAll(ctx context.Context) ([]*AccountPool, error) {
	var pools []*AccountPool
	err := r.db.WithContext(ctx).Find(&pools).Error
	return pools, err
}

// IncrementRequests 增加请求计数
func (r *repository) IncrementRequests(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&AccountPool{}).
		Where("id = ?", id).
		UpdateColumn("total_requests", gorm.Expr("total_requests + ?", 1)).
		Error
}

// IncrementErrors 增加错误计数
func (r *repository) IncrementErrors(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&AccountPool{}).
		Where("id = ?", id).
		UpdateColumn("total_errors", gorm.Expr("total_errors + ?", 1)).
		Error
}

// CreateRequestLog 创建请求日志
func (r *repository) CreateRequestLog(ctx context.Context, log *AccountPoolRequestLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// ListRequestLogs 查询请求日志列表
func (r *repository) ListRequestLogs(ctx context.Context, filter *RequestLogFilter, opts *query.Options) ([]*AccountPoolRequestLog, int64, error) {
	var logs []*AccountPoolRequestLog
	var total int64

	db := r.db.WithContext(ctx).Model(&AccountPoolRequestLog{})

	// 应用过滤器
	if filter != nil {
		if filter.PoolID != nil {
			db = db.Where("pool_id = ?", *filter.PoolID)
		}
		if filter.CredentialID != nil {
			db = db.Where("credential_id = ?", *filter.CredentialID)
		}
		if filter.Provider != nil {
			db = db.Where("provider_type = ?", *filter.Provider)
		}
		if filter.Model != nil {
			db = db.Where("model = ?", *filter.Model)
		}
		if filter.StatusCode != nil {
			db = db.Where("status_code = ?", *filter.StatusCode)
		}
	}

	// 计算总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用查询选项
	if opts != nil {
		db = query.ApplyOptions(db, opts)
	}

	// 查询数据
	if err := db.Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetPoolRequestStats 获取账号池请求统计
func (r *repository) GetPoolRequestStats(ctx context.Context, poolID uint) (map[string]interface{}, error) {
	var stats struct {
		TotalRequests int64
		SuccessCount  int64
		ErrorCount    int64
		AvgResponse   float64
		TotalTokens   int64
	}

	err := r.db.WithContext(ctx).
		Model(&AccountPoolRequestLog{}).
		Where("pool_id = ?", poolID).
		Select(`
			COUNT(*) as total_requests,
			SUM(CASE WHEN status_code >= 200 AND status_code < 300 THEN 1 ELSE 0 END) as success_count,
			SUM(CASE WHEN status_code >= 400 OR error_message != '' THEN 1 ELSE 0 END) as error_count,
			AVG(response_time) as avg_response,
			SUM(tokens_used) as total_tokens
		`).
		Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_requests": stats.TotalRequests,
		"success_count":  stats.SuccessCount,
		"error_count":    stats.ErrorCount,
		"avg_response":   stats.AvgResponse,
		"total_tokens":   stats.TotalTokens,
	}, nil
}

// CreateCredential 创建凭据
func (r *repository) CreateCredential(ctx context.Context, cred *AccountCredential) error {
	return r.db.WithContext(ctx).Create(cred).Error
}

// UpdateCredential 更新凭据
func (r *repository) UpdateCredential(ctx context.Context, cred *AccountCredential) error {
	return r.db.WithContext(ctx).Save(cred).Error
}

// DeleteCredential 删除凭据
func (r *repository) DeleteCredential(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&AccountCredential{}, id).Error
}

// FindCredentialByID 根据ID查找凭据
func (r *repository) FindCredentialByID(ctx context.Context, id uint) (*AccountCredential, error) {
	var cred AccountCredential
	err := r.db.WithContext(ctx).First(&cred, id).Error
	if err != nil {
		return nil, err
	}
	return &cred, nil
}

// ListCredentials 查询凭据列表
func (r *repository) ListCredentials(ctx context.Context, filter *CredentialFilter, opts *query.Options) ([]*AccountCredential, int64, error) {
	var creds []*AccountCredential
	var total int64

	db := r.db.WithContext(ctx).Model(&AccountCredential{})

	// 应用过滤器
	if filter != nil {
		if filter.PoolID != nil {
			db = db.Where("pool_id = ?", *filter.PoolID)
		}
		if filter.Provider != nil {
			db = db.Where("provider_type = ?", *filter.Provider)
		}
		if filter.AuthType != nil {
			db = db.Where("auth_type = ?", *filter.AuthType)
		}
		if filter.IsActive != nil {
			db = db.Where("is_active = ?", *filter.IsActive)
		}
		if filter.HealthStatus != nil {
			db = db.Where("health_status = ?", *filter.HealthStatus)
		}
	}

	// 计算总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用查询选项
	if opts != nil {
		db = query.ApplyOptions(db, opts)
	}

	// 查询数据
	if err := db.Find(&creds).Error; err != nil {
		return nil, 0, err
	}

	return creds, total, nil
}

// UpdateCredentialStatus 更新凭据状态
func (r *repository) UpdateCredentialStatus(ctx context.Context, id uint, isActive bool) error {
	return r.db.WithContext(ctx).
		Model(&AccountCredential{}).
		Where("id = ?", id).
		Update("is_active", isActive).
		Error
}

// IncrementCredentialRequests 增加凭据请求计数
func (r *repository) IncrementCredentialRequests(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&AccountCredential{}).
		Where("id = ?", id).
		UpdateColumn("total_requests", gorm.Expr("total_requests + ?", 1)).
		Error
}

// IncrementCredentialErrors 增加凭据错误计数
func (r *repository) IncrementCredentialErrors(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&AccountCredential{}).
		Where("id = ?", id).
		UpdateColumn("total_errors", gorm.Expr("total_errors + ?", 1)).
		Error
}

// FindActiveCredentialsByPoolID 查找账号池的所有活跃凭据
func (r *repository) FindActiveCredentialsByPoolID(ctx context.Context, poolID uint) ([]*AccountCredential, error) {
	var creds []*AccountCredential
	err := r.db.WithContext(ctx).
		Where("pool_id = ? AND is_active = ?", poolID, true).
		Order("weight DESC, total_requests ASC").
		Find(&creds).Error
	return creds, err
}
