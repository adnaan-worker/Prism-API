package service

import (
	"api-aggregator/backend/internal/models"
	"api-aggregator/backend/internal/repository"
	"context"
	"errors"
	"fmt"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrInvalidPage  = errors.New("invalid page parameters")
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// GetUsersRequest represents a request to get users list
type GetUsersRequest struct {
	Page     int `form:"page" binding:"omitempty,min=1"`
	PageSize int `form:"page_size" binding:"omitempty,min=1,max=100"`
}

// GetUsersResponse represents a paginated users response
type GetUsersResponse struct {
	Users    []*models.User `json:"users"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

// UpdateUserStatusRequest represents a request to update user status
type UpdateUserStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=active inactive banned"`
}

// UpdateUserQuotaRequest represents a request to update user quota
type UpdateUserQuotaRequest struct {
	Quota int64 `json:"quota" binding:"required,min=0"`
}

// GetUsers returns a paginated list of users
func (s *UserService) GetUsers(ctx context.Context, req *GetUsersRequest) (*GetUsersResponse, error) {
	// Set default values
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}

	// Validate parameters
	if req.Page < 1 || req.PageSize < 1 || req.PageSize > 100 {
		return nil, ErrInvalidPage
	}

	users, total, err := s.userRepo.FindAll(ctx, req.Page, req.PageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	return &GetUsersResponse{
		Users:    users,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// GetUserByID returns a user by ID
func (s *UserService) GetUserByID(ctx context.Context, id uint) (*models.User, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// UpdateUserStatus updates a user's status
func (s *UserService) UpdateUserStatus(ctx context.Context, id uint, status string) error {
	// Check if user exists
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Update status
	if err := s.userRepo.UpdateStatus(ctx, id, status); err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}

	return nil
}

// UpdateUserQuota updates a user's quota
func (s *UserService) UpdateUserQuota(ctx context.Context, id uint, quota int64) error {
	// Check if user exists
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Update quota
	if err := s.userRepo.UpdateQuota(ctx, id, quota); err != nil {
		return fmt.Errorf("failed to update user quota: %w", err)
	}

	return nil
}
