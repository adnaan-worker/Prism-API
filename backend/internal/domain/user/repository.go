package user

import (
	"api-aggregator/backend/pkg/query"
	"context"
	"errors"

	"gorm.io/gorm"
)

// Repository 用户仓储接口
type Repository interface {
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uint) error
	FindByID(ctx context.Context, id uint) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindAll(ctx context.Context, page, pageSize int) ([]*User, int64, error)
	List(ctx context.Context, filters []query.Filter, sorts []query.Sort, pagination *query.Pagination) ([]*User, int64, error)
	UpdateStatus(ctx context.Context, id uint, status string) error
	UpdateQuota(ctx context.Context, id uint, quota int64) error
	CountAll(ctx context.Context) (int64, error)
	CountByStatus(ctx context.Context, status string) (int64, error)
}

// repository 用户仓储实现
type repository struct {
	db *gorm.DB
}

// NewRepository 创建用户仓储
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// Create 创建用户
func (r *repository) Create(ctx context.Context, user *User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// Update 更新用户
func (r *repository) Update(ctx context.Context, user *User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// Delete 删除用户
func (r *repository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&User{}, id).Error
}

// FindByID 根据ID查找用户
func (r *repository) FindByID(ctx context.Context, id uint) (*User, error) {
	var user User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindByUsername 根据用户名查找用户
func (r *repository) FindByUsername(ctx context.Context, username string) (*User, error) {
	var user User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindByEmail 根据邮箱查找用户
func (r *repository) FindByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindAll 查询所有用户（分页）
func (r *repository) FindAll(ctx context.Context, page, pageSize int) ([]*User, int64, error) {
	var users []*User
	var total int64

	// 统计总数
	if err := r.db.WithContext(ctx).Model(&User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := r.db.WithContext(ctx).
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&users).Error

	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// List 查询用户列表（支持过滤、排序、分页）
func (r *repository) List(ctx context.Context, filters []query.Filter, sorts []query.Sort, pagination *query.Pagination) ([]*User, int64, error) {
	builder := query.NewBuilder(r.db.WithContext(ctx).Model(&User{}))
	
	// 应用过滤
	builder.ApplyFilters(filters)
	
	// 统计总数
	var total int64
	builder.Count(&total)
	
	// 应用排序和分页
	builder.ApplySort(sorts).ApplyPagination(pagination)
	
	// 查询结果
	var users []*User
	err := builder.Find(&users)
	if err != nil {
		return nil, 0, err
	}
	
	return users, total, nil
}

// UpdateStatus 更新用户状态
func (r *repository) UpdateStatus(ctx context.Context, id uint, status string) error {
	return r.db.WithContext(ctx).Model(&User{}).Where("id = ?", id).Update("status", status).Error
}

// UpdateQuota 更新用户配额
func (r *repository) UpdateQuota(ctx context.Context, id uint, quota int64) error {
	return r.db.WithContext(ctx).Model(&User{}).Where("id = ?", id).Update("quota", quota).Error
}

// CountAll 统计所有用户数
func (r *repository) CountAll(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&User{}).Count(&count).Error
	return count, err
}

// CountByStatus 根据状态统计用户数
func (r *repository) CountByStatus(ctx context.Context, status string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&User{}).Where("status = ?", status).Count(&count).Error
	return count, err
}
