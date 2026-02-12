package auth

import (
	"api-aggregator/backend/pkg/errors"
	"api-aggregator/backend/pkg/response"
	"strings"

	"github.com/gin-gonic/gin"
)

// Handler 认证处理器
type Handler struct {
	service Service
}

// NewHandler 创建认证处理器
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// Register 用户注册
// @Summary 用户注册
// @Description 创建新用户账号
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "注册请求"
// @Success 201 {object} RegisterResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	resp, err := h.service.Register(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, errors.ErrUsernameExists) {
			response.Conflict(c, "Username already exists", "")
			return
		}
		if errors.Is(err, errors.ErrEmailExists) {
			response.Conflict(c, "Email already exists", "")
			return
		}
		response.InternalError(c, err)
		return
	}

	response.Created(c, resp)
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录并获取 JWT token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "登录请求"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	resp, err := h.service.Login(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, errors.ErrInvalidPassword) {
			response.Unauthorized(c, "Invalid username or password")
			return
		}
		appErr, ok := err.(*errors.AppError)
		if ok && appErr.Code == 403001 {
			response.Forbidden(c, appErr.Message)
			return
		}
		response.InternalError(c, err)
		return
	}

	response.Success(c, resp)
}

// GetProfile 获取当前用户信息
// @Summary 获取当前用户信息
// @Description 获取当前登录用户的详细信息
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UserInfo
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/auth/profile [get]
func (h *Handler) GetProfile(c *gin.Context) {
	// 从上下文获取用户ID（由中间件设置）
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	userInfo, err := h.service.GetUserInfo(c.Request.Context(), userID.(uint))
	if err != nil {
		if errors.Is(err, errors.ErrUserNotFound) {
			response.NotFound(c, "User not found")
			return
		}
		response.InternalError(c, err)
		return
	}

	response.Success(c, gin.H{"user": userInfo})
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Description 修改当前用户的密码
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ChangePasswordRequest true "修改密码请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/auth/change-password [post]
func (h *Handler) ChangePassword(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	if err := h.service.ChangePassword(c.Request.Context(), userID.(uint), &req); err != nil {
		if errors.Is(err, errors.ErrUserNotFound) {
			response.NotFound(c, "User not found")
			return
		}
		if errors.Is(err, errors.ErrInvalidPassword) {
			response.BadRequest(c, "Invalid old password", "")
			return
		}
		response.InternalError(c, err)
		return
	}

	response.SuccessWithMessage(c, "Password changed successfully", nil)
}

// ValidateToken 验证 token（用于中间件）
func (h *Handler) ValidateToken(c *gin.Context) (*Claims, error) {
	// 从 Header 获取 token
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return nil, errors.ErrUnauthorized
	}

	// 解析 Bearer token
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, errors.ErrInvalidToken
	}

	// 验证 token
	claims, err := h.service.ValidateToken(c.Request.Context(), parts[1])
	if err != nil {
		return nil, err
	}

	return claims, nil
}
