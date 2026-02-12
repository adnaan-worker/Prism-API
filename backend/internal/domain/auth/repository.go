package auth

import (
	"api-aggregator/backend/internal/domain/user"
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

// Repository 认证仓储接口
type Repository interface {
	// 用户相关
	FindUserByUsername(ctx context.Context, username string) (*user.User, error)
	FindUserByEmail(ctx context.Context, email string) (*user.User, error)
	FindUserByID(ctx context.Context, id uint) (*user.User, error)
	CreateUser(ctx context.Context, user *user.User) error
	UpdateUser(ctx context.Context, user *user.User) error
	UpdateLastSignIn(ctx context.Context, userID uint) error
	UpdatePassword(ctx context.Context, userID uint, passwordHash string) error
}

// repository 认证仓储实现
type repository struct {
	db *gorm.DB
}

// NewRepository 创建认证仓储
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// FindUserByUsername 根据用户名查找用户
func (r *repository) FindUserByUsername(ctx context.Context, username string) (*user.User, error) {
	var u user.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

// FindUserByEmail 根据邮箱查找用户
func (r *repository) FindUserByEmail(ctx context.Context, email string) (*user.User, error) {
	var u user.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
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

// CreateUser 创建用户
func (r *repository) CreateUser(ctx context.Context, u *user.User) error {
	return r.db.WithContext(ctx).Create(u).Error
}

// UpdateUser 更新用户
func (r *repository) UpdateUser(ctx context.Context, u *user.User) error {
	return r.db.WithContext(ctx).Save(u).Error
}

// UpdateLastSignIn 更新最后登录时间
func (r *repository) UpdateLastSignIn(ctx context.Context, userID uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&user.User{}).
		Where("id = ?", userID).
		Update("last_sign_in", now).Error
}

// UpdatePassword 更新密码
func (r *repository) UpdatePassword(ctx context.Context, userID uint, passwordHash string) error {
	return r.db.WithContext(ctx).Model(&user.User{}).
		Where("id = ?", userID).
		Update("password_hash", passwordHash).Error
}
