package cache

import (
	"api-aggregator/backend/pkg/errors"
	"api-aggregator/backend/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handler 缓存处理器
type Handler struct {
	service Service
}

// NewHandler 创建缓存处理器
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// GetCacheStats 获取缓存统计
// @Summary 获取缓存统计
// @Description 获取缓存统计信息
// @Tags Cache
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id query int false "用户ID（管理员可查询所有用户）"
// @Success 200 {object} CacheStatsResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/cache/stats [get]
func (h *Handler) GetCacheStats(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	// 检查是否是管理员查询其他用户
	var targetUserID *uint
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		// 这里应该检查是否是管理员，简化处理直接使用当前用户
		uid := userID.(uint)
		targetUserID = &uid
	} else {
		uid := userID.(uint)
		targetUserID = &uid
	}

	stats, err := h.service.GetCacheStats(c.Request.Context(), targetUserID)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, stats)
}

// GetCacheList 获取缓存列表
// @Summary 获取缓存列表
// @Description 获取缓存列表（管理员）
// @Tags Cache
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param user_id query int false "用户ID"
// @Param model query string false "模型名称"
// @Success 200 {object} CacheListResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/admin/cache/list [get]
func (h *Handler) GetCacheList(c *gin.Context) {
	var req GetCacheListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters", err.Error())
		return
	}

	caches, err := h.service.GetCacheList(c.Request.Context(), &req)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, caches)
}

// CleanExpiredCache 清理过期缓存
// @Summary 清理过期缓存
// @Description 清理所有过期的缓存（管理员）
// @Tags Cache
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} CleanExpiredCacheResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/admin/cache/clean [delete]
func (h *Handler) CleanExpiredCache(c *gin.Context) {
	result, err := h.service.CleanExpiredCache(c.Request.Context())
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, result)
}

// ClearUserCache 清除用户缓存
// @Summary 清除用户缓存
// @Description 清除指定用户的所有缓存（管理员）
// @Tags Cache
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} ClearUserCacheResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/admin/cache/user/{id} [delete]
func (h *Handler) ClearUserCache(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid user ID", "User ID must be a valid number")
		return
	}

	result, err := h.service.ClearUserCache(c.Request.Context(), uint(userID))
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, result)
}

// DeleteCache 删除缓存
// @Summary 删除缓存
// @Description 删除指定ID的缓存（管理员）
// @Tags Cache
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "缓存ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/admin/cache/{id} [delete]
func (h *Handler) DeleteCache(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid cache ID", "Cache ID must be a valid number")
		return
	}

	if err := h.service.DeleteCache(c.Request.Context(), uint(id)); err != nil {
		appErr, ok := err.(*errors.AppError)
		if ok && appErr.Code == 404001 {
			response.NotFound(c, "Cache not found")
			return
		}
		response.InternalError(c, err)
		return
	}

	response.SuccessWithMessage(c, "Cache deleted successfully", nil)
}
