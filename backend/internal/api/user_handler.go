package api

import (
	"api-aggregator/backend/internal/service"
	"api-aggregator/backend/pkg/response"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetUsers handles getting a paginated list of users (admin only)
func (h *UserHandler) GetUsers(c *gin.Context) {
	var req service.GetUsersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid request", err.Error())
		return
	}

	resp, err := h.userService.GetUsers(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidPage) {
			response.BadRequest(c, "Invalid page parameters", err.Error())
			return
		}
		response.InternalError(c, err)
		return
	}

	response.Success(c, resp)
}

// GetUserByID handles getting a user by ID (admin only)
func (h *UserHandler) GetUserByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid user ID", "User ID must be a valid number")
		return
	}

	user, err := h.userService.GetUserByID(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			response.NotFound(c, "User not found")
			return
		}
		response.InternalError(c, err)
		return
	}

	response.Success(c, gin.H{"user": user})
}

// UpdateUserStatus handles updating a user's status (admin only)
func (h *UserHandler) UpdateUserStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid user ID", "User ID must be a valid number")
		return
	}

	var req service.UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request", err.Error())
		return
	}

	if err := h.userService.UpdateUserStatus(c.Request.Context(), uint(id), req.Status); err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			response.NotFound(c, "User not found")
			return
		}
		response.InternalError(c, err)
		return
	}

	response.SuccessWithMessage(c, "User status updated successfully", nil)
}

// UpdateUserQuota handles updating a user's quota (admin only)
func (h *UserHandler) UpdateUserQuota(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid user ID", "User ID must be a valid number")
		return
	}

	var req service.UpdateUserQuotaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request", err.Error())
		return
	}

	if err := h.userService.UpdateUserQuota(c.Request.Context(), uint(id), req.Quota); err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			response.NotFound(c, "User not found")
			return
		}
		response.InternalError(c, err)
		return
	}

	response.SuccessWithMessage(c, "User quota updated successfully", nil)
}
