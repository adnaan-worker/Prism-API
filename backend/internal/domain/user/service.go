package user

import (
	"api-aggregator/backend/pkg/errors"
	"api-aggregator/backend/pkg/logger"
	"api-aggregator/backend/pkg/query"
	"context"
)

// Service 用户服务接口
type Service interface {
	GetUsers(ctx context.Context, req *GetUsersRequest) (*GetUsersResponse, error)
	GetUserByID(ctx context.Context, id uint) (*UserResponse, error)
	UpdateUserStatus(ctx context.Context, id uint, req *UpdateUserStatusRequest) error
	UpdateUserQuota(ctx context.Context, id uint, req *UpdateUserQuotaRequest) error
	DeleteUser(ctx context.Context, id uint) error
}

// service 用户服务实现
type service struct {
	repo   Repository
	logger logger.Logger
}

// NewService 创建用户服务
func NewService(repo Repository, logger logger.Logger) Service {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

// GetUsers 获取用户列表
func (s *service) GetUsers(ctx context.Context, req *GetUsersRequest) (*GetUsersResponse, error) {
	// 设置默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}

	// 验证参数
	if req.Page < 1 || req.PageSize < 1 || req.PageSize > 100 {
		return nil, errors.ErrInvalidParam.WithDetails("Invalid page parameters")
	}

	// 构建过滤条件
	var filters []query.Filter
	if req.Status != "" {
		filters = append(filters, query.Filter{
			Field:    "status",
			Operator: "=",
			Value:    req.Status,
		})
	}
	if req.Search != "" {
		// 搜索用户名或邮箱
		filters = append(filters, query.Filter{
			Field:    "username",
			Operator: "LIKE",
			Value:    "%" + req.Search + "%",
		})
	}

	// 构建排序
	sorts := []query.Sort{
		{Field: "created_at", Desc: true},
	}

	// 构建分页
	pagination := &query.Pagination{
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	// 查询用户列表
	users, total, err := s.repo.List(ctx, filters, sorts, pagination)
	if err != nil {
		s.logger.Error("Failed to get users", logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to get users")
	}

	return &GetUsersResponse{
		Users:    ToResponseList(users),
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// GetUserByID 根据ID获取用户
func (s *service) GetUserByID(ctx context.Context, id uint) (*UserResponse, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get user", logger.Uint("user_id", id), logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to get user")
	}
	if user == nil {
		return nil, errors.ErrUserNotFound
	}

	return user.ToResponse(), nil
}

// UpdateUserStatus 更新用户状态
func (s *service) UpdateUserStatus(ctx context.Context, id uint, req *UpdateUserStatusRequest) error {
	// 检查用户是否存在
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get user", logger.Uint("user_id", id), logger.Error(err))
		return errors.Wrap(err, 500002, "Failed to get user")
	}
	if user == nil {
		return errors.ErrUserNotFound
	}

	// 更新状态
	if err := s.repo.UpdateStatus(ctx, id, req.Status); err != nil {
		s.logger.Error("Failed to update user status",
			logger.Uint("user_id", id),
			logger.String("status", req.Status),
			logger.Error(err))
		return errors.Wrap(err, 500002, "Failed to update user status")
	}

	s.logger.Info("User status updated",
		logger.Uint("user_id", id),
		logger.String("status", req.Status))

	return nil
}

// UpdateUserQuota 更新用户配额
func (s *service) UpdateUserQuota(ctx context.Context, id uint, req *UpdateUserQuotaRequest) error {
	// 检查用户是否存在
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get user", logger.Uint("user_id", id), logger.Error(err))
		return errors.Wrap(err, 500002, "Failed to get user")
	}
	if user == nil {
		return errors.ErrUserNotFound
	}

	// 更新配额
	if err := s.repo.UpdateQuota(ctx, id, req.Quota); err != nil {
		s.logger.Error("Failed to update user quota",
			logger.Uint("user_id", id),
			logger.Int64("quota", req.Quota),
			logger.Error(err))
		return errors.Wrap(err, 500002, "Failed to update user quota")
	}

	s.logger.Info("User quota updated",
		logger.Uint("user_id", id),
		logger.Int64("quota", req.Quota))

	return nil
}

// DeleteUser 删除用户
func (s *service) DeleteUser(ctx context.Context, id uint) error {
	// 检查用户是否存在
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get user", logger.Uint("user_id", id), logger.Error(err))
		return errors.Wrap(err, 500002, "Failed to get user")
	}
	if user == nil {
		return errors.ErrUserNotFound
	}

	// 删除用户
	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to delete user", logger.Uint("user_id", id), logger.Error(err))
		return errors.Wrap(err, 500002, "Failed to delete user")
	}

	s.logger.Info("User deleted", logger.Uint("user_id", id))

	return nil
}
