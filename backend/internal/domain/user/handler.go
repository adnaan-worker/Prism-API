package user

import (
	"api-aggregator/backend/pkg/errors"
	"api-aggregator/backend/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handler 用户处理器
type Handler struct {
	service Service
}

// NewHandler 创建用户处理器
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// GetUsers 获取用户列表
// @Summary 获取用户列表
// @Description 获取用户列表（管理员）
// @Tags User
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param status query string false "状态过滤" Enums(active, inactive, banned)
// @Param search query string false "搜索关键词"
// @Success 200 {object} GetUsersResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/admin/users [get]
func (h *Handler) GetUsers(c *gin.Context) {
	var req GetUsersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters", err.Error())
		return
	}

	resp, err := h.service.GetUsers(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, errors.ErrInvalidParam) {
			response.BadRequest(c, err.Error(), "")
			return
		}
		response.InternalError(c, err)
		return
	}

	response.Success(c, resp)
}

// GetUserByID 根据ID获取用户
// @Summary 根据ID获取用户
// @Description 根据ID获取用户详情（管理员）
// @Tags User
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} UserResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/admin/users/{id} [get]
func (h *Handler) GetUserByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid user ID", "User ID must be a valid number")
		return
	}

	user, err := h.service.GetUserByID(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, errors.ErrUserNotFound) {
			response.NotFound(c, "User not found")
			return
		}
		response.InternalError(c, err)
		return
	}

	response.Success(c, gin.H{"user": user})
}

// UpdateUserStatus 更新用户状态
// @Summary 更新用户状态
// @Description 更新用户状态（管理员）
// @Tags User
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Param request body UpdateUserStatusRequest true "更新请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/admin/users/{id}/status [put]
func (h *Handler) UpdateUserStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid user ID", "User ID must be a valid number")
		return
	}

	var req UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	if err := h.service.UpdateUserStatus(c.Request.Context(), uint(id), &req); err != nil {
		if errors.Is(err, errors.ErrUserNotFound) {
			response.NotFound(c, "User not found")
			return
		}
		response.InternalError(c, err)
		return
	}

	response.SuccessWithMessage(c, "User status updated successfully", nil)
}

// UpdateUserQuota 更新用户配额
// @Summary 更新用户配额
// @Description 更新用户配额（管理员）
// @Tags User
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Param request body UpdateUserQuotaRequest true "更新请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/admin/users/{id}/quota [put]
func (h *Handler) UpdateUserQuota(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid user ID", "User ID must be a valid number")
		return
	}

	var req UpdateUserQuotaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	if err := h.service.UpdateUserQuota(c.Request.Context(), uint(id), &req); err != nil {
		if errors.Is(err, errors.ErrUserNotFound) {
			response.NotFound(c, "User not found")
			return
		}
		response.InternalError(c, err)
		return
	}

	response.SuccessWithMessage(c, "User quota updated successfully", nil)
}

// DeleteUser 删除用户
// @Summary 删除用户
// @Description 删除用户（管理员）
// @Tags User
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/admin/users/{id} [delete]
func (h *Handler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid user ID", "User ID must be a valid number")
		return
	}

	if err := h.service.DeleteUser(c.Request.Context(), uint(id)); err != nil {
		if errors.Is(err, errors.ErrUserNotFound) {
			response.NotFound(c, "User not found")
			return
		}
		response.InternalError(c, err)
		return
	}

	response.SuccessWithMessage(c, "User deleted successfully", nil)
}
