package auth

import (
	"api-aggregator/backend/internal/domain/user"
	"api-aggregator/backend/pkg/crypto"
	"api-aggregator/backend/pkg/errors"
	"api-aggregator/backend/pkg/logger"
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Service 认证服务接口
type Service interface {
	Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error)
	Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error)
	ValidateToken(ctx context.Context, tokenString string) (*Claims, error)
	ChangePassword(ctx context.Context, userID uint, req *ChangePasswordRequest) error
	GetUserInfo(ctx context.Context, userID uint) (*UserInfo, error)
}

// service 认证服务实现
type service struct {
	repo      Repository
	jwtSecret string
	logger    logger.Logger
}

// NewService 创建认证服务
func NewService(repo Repository, jwtSecret string, logger logger.Logger) Service {
	return &service{
		repo:      repo,
		jwtSecret: jwtSecret,
		logger:    logger,
	}
}

// Register 用户注册
func (s *service) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	// 检查用户名是否已存在
	existingUser, err := s.repo.FindUserByUsername(ctx, req.Username)
	if err != nil {
		s.logger.Error("Failed to check username", logger.String("username", req.Username), logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to check username")
	}
	if existingUser != nil {
		return nil, errors.ErrUsernameExists
	}

	// 检查邮箱是否已存在
	existingUser, err = s.repo.FindUserByEmail(ctx, req.Email)
	if err != nil {
		s.logger.Error("Failed to check email", logger.String("email", req.Email), logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to check email")
	}
	if existingUser != nil {
		return nil, errors.ErrEmailExists
	}

	// 哈希密码
	hashedPassword, err := crypto.HashPassword(req.Password)
	if err != nil {
		s.logger.Error("Failed to hash password", logger.Error(err))
		return nil, errors.Wrap(err, 500005, "Failed to hash password")
	}

	// 创建用户
	newUser := &user.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Quota:        10000, // 默认配额
		UsedQuota:    0,
		IsAdmin:      false,
		Status:       "active",
	}

	if err := s.repo.CreateUser(ctx, newUser); err != nil {
		s.logger.Error("Failed to create user",
			logger.String("username", req.Username),
			logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to create user")
	}

	s.logger.Info("User registered successfully",
		logger.Uint("user_id", newUser.ID),
		logger.String("username", newUser.Username))

	return &RegisterResponse{
		User: s.toUserInfo(newUser),
	}, nil
}

// Login 用户登录
func (s *service) Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error) {
	// 查找用户
	u, err := s.repo.FindUserByUsername(ctx, req.Username)
	if err != nil {
		s.logger.Error("Failed to find user", logger.String("username", req.Username), logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to find user")
	}
	if u == nil {
		return nil, errors.ErrInvalidPassword
	}

	// 检查用户状态
	if !u.IsActive() {
		return nil, errors.New(403001, "User account is not active")
	}

	// 验证密码
	if !crypto.CheckPassword(req.Password, u.PasswordHash) {
		return nil, errors.ErrInvalidPassword
	}

	// 生成 JWT token
	token, expiresAt, err := s.generateToken(u)
	if err != nil {
		s.logger.Error("Failed to generate token", logger.Uint("user_id", u.ID), logger.Error(err))
		return nil, errors.Wrap(err, 500001, "Failed to generate token")
	}

	// 更新最后登录时间
	if err := s.repo.UpdateLastSignIn(ctx, u.ID); err != nil {
		s.logger.Warn("Failed to update last sign in", logger.Uint("user_id", u.ID), logger.Error(err))
		// 不影响登录流程，只记录警告
	}

	s.logger.Info("User logged in successfully",
		logger.Uint("user_id", u.ID),
		logger.String("username", u.Username))

	return &AuthResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      s.toUserInfo(u),
	}, nil
}

// ValidateToken 验证 JWT token
func (s *service) ValidateToken(ctx context.Context, tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New(401002, "Invalid token signing method")
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		s.logger.Warn("Failed to parse token", logger.Error(err))
		return nil, errors.ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.ErrInvalidToken
	}

	// 检查是否过期
	if claims.IsExpired() {
		return nil, errors.ErrTokenExpired
	}

	return claims, nil
}

// ChangePassword 修改密码
func (s *service) ChangePassword(ctx context.Context, userID uint, req *ChangePasswordRequest) error {
	// 查找用户
	u, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to find user", logger.Uint("user_id", userID), logger.Error(err))
		return errors.Wrap(err, 500002, "Failed to find user")
	}
	if u == nil {
		return errors.ErrUserNotFound
	}

	// 验证旧密码
	if !crypto.CheckPassword(req.OldPassword, u.PasswordHash) {
		return errors.ErrInvalidPassword
	}

	// 哈希新密码
	hashedPassword, err := crypto.HashPassword(req.NewPassword)
	if err != nil {
		s.logger.Error("Failed to hash password", logger.Error(err))
		return errors.Wrap(err, 500005, "Failed to hash password")
	}

	// 更新密码
	if err := s.repo.UpdatePassword(ctx, userID, hashedPassword); err != nil {
		s.logger.Error("Failed to update password", logger.Uint("user_id", userID), logger.Error(err))
		return errors.Wrap(err, 500002, "Failed to update password")
	}

	s.logger.Info("Password changed successfully", logger.Uint("user_id", userID))

	return nil
}

// GetUserInfo 获取用户信息
func (s *service) GetUserInfo(ctx context.Context, userID uint) (*UserInfo, error) {
	u, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to find user", logger.Uint("user_id", userID), logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to find user")
	}
	if u == nil {
		return nil, errors.ErrUserNotFound
	}

	return s.toUserInfo(u), nil
}

// generateToken 生成 JWT token
func (s *service) generateToken(u *user.User) (string, time.Time, error) {
	duration := 24 * time.Hour
	claims := NewClaims(u.ID, u.Username, u.IsAdmin, duration)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, claims.ExpiresAt.Time, nil
}

// toUserInfo 转换为用户信息
func (s *service) toUserInfo(u *user.User) *UserInfo {
	return &UserInfo{
		ID:         u.ID,
		Username:   u.Username,
		Email:      u.Email,
		IsAdmin:    u.IsAdmin,
		Status:     u.Status,
		Quota:      u.Quota,
		UsedQuota:  u.UsedQuota,
		LastSignIn: u.LastSignIn,
		CreatedAt:  u.CreatedAt,
	}
}
