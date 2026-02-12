package auth

import "time"

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=100"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=100"`
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// AuthResponse 认证响应
type AuthResponse struct {
	Token     string       `json:"token"`
	ExpiresAt time.Time    `json:"expires_at"`
	User      *UserInfo    `json:"user"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID         uint       `json:"id"`
	Username   string     `json:"username"`
	Email      string     `json:"email"`
	IsAdmin    bool       `json:"is_admin"`
	Status     string     `json:"status"`
	Quota      int64      `json:"quota"`
	UsedQuota  int64      `json:"used_quota"`
	LastSignIn *time.Time `json:"last_sign_in,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// RegisterResponse 注册响应
type RegisterResponse struct {
	User *UserInfo `json:"user"`
}
