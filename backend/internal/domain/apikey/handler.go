package apikey

import (
	"api-aggregator/backend/pkg/errors"
	"api-aggregator/backend/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handler API密钥处理器
type Handler struct {
	service Service
}

// NewHandler 创建API密钥处理器
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// CreateAPIKey 创建API密钥
// @Summary 创建API密钥
// @Description 为当前用户创建新的API密钥
// @Tags APIKey
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateAPIKeyRequest true "创建请求"
// @Success 201 {object} CreateAPIKeyResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/apikeys [post]
func (h *Handler) CreateAPIKey(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	resp, err := h.service.CreateAPIKey(c.Request.Context(), userID.(uint), &req)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Created(c, resp)
}

// GetAPIKeys 获取API密钥列表
// @Summary 获取API密钥列表
// @Description 获取当前用户的所有API密钥
// @Tags APIKey
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param is_active query bool false "是否激活"
// @Success 200 {object} APIKeyListResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/apikeys [get]
func (h *Handler) GetAPIKeys(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req GetAPIKeysRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters", err.Error())
		return
	}

	resp, err := h.service.GetAPIKeys(c.Request.Context(), userID.(uint), &req)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, resp)
}

// GetAPIKeyByID 根据ID获取API密钥
// @Summary 根据ID获取API密钥
// @Description 获取指定ID的API密钥详情
// @Tags APIKey
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "密钥ID"
// @Success 200 {object} APIKeyResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/apikeys/{id} [get]
func (h *Handler) GetAPIKeyByID(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid API key ID", "API key ID must be a valid number")
		return
	}

	apiKey, err := h.service.GetAPIKeyByID(c.Request.Context(), userID.(uint), uint(id))
	if err != nil {
		if errors.Is(err, errors.ErrAPIKeyNotFound) {
			response.NotFound(c, "API key not found")
			return
		}
		if errors.Is(err, errors.ErrForbidden) {
			response.Forbidden(c, err.Error())
			return
		}
		response.InternalError(c, err)
		return
	}

	response.Success(c, gin.H{"api_key": apiKey})
}

// UpdateAPIKey 更新API密钥
// @Summary 更新API密钥
// @Description 更新指定ID的API密钥信息
// @Tags APIKey
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "密钥ID"
// @Param request body UpdateAPIKeyRequest true "更新请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/apikeys/{id} [put]
func (h *Handler) UpdateAPIKey(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid API key ID", "API key ID must be a valid number")
		return
	}

	var req UpdateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	if err := h.service.UpdateAPIKey(c.Request.Context(), userID.(uint), uint(id), &req); err != nil {
		if errors.Is(err, errors.ErrAPIKeyNotFound) {
			response.NotFound(c, "API key not found")
			return
		}
		if errors.Is(err, errors.ErrForbidden) {
			response.Forbidden(c, err.Error())
			return
		}
		response.InternalError(c, err)
		return
	}

	response.SuccessWithMessage(c, "API key updated successfully", nil)
}

// DeleteAPIKey 删除API密钥
// @Summary 删除API密钥
// @Description 删除指定ID的API密钥
// @Tags APIKey
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "密钥ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/apikeys/{id} [delete]
func (h *Handler) DeleteAPIKey(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid API key ID", "API key ID must be a valid number")
		return
	}

	if err := h.service.DeleteAPIKey(c.Request.Context(), userID.(uint), uint(id)); err != nil {
		if errors.Is(err, errors.ErrAPIKeyNotFound) {
			response.NotFound(c, "API key not found")
			return
		}
		if errors.Is(err, errors.ErrForbidden) {
			response.Forbidden(c, err.Error())
			return
		}
		response.InternalError(c, err)
		return
	}

	response.SuccessWithMessage(c, "API key deleted successfully", nil)
}
