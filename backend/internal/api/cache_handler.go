package api

import (
	"api-aggregator/backend/internal/service"
	"api-aggregator/backend/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CacheHandler struct {
	cacheService *service.CacheService
}

func NewCacheHandler(cacheService *service.CacheService) *CacheHandler {
	return &CacheHandler{
		cacheService: cacheService,
	}
}

// GetCacheStats 获取缓存统计信息
func (h *CacheHandler) GetCacheStats(c *gin.Context) {
	// 从上下文获取用户 ID（由认证中间件设置）
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	stats, err := h.cacheService.GetCacheStats(c.Request.Context(), userID.(uint))
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, stats)
}

// CleanExpiredCache 清理过期缓存（管理员接口）
func (h *CacheHandler) CleanExpiredCache(c *gin.Context) {
	err := h.cacheService.CleanExpiredCache(c.Request.Context())
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.SuccessWithMessage(c, "Expired cache cleaned successfully", nil)
}

// ClearUserCache 清除用户的所有缓存
func (h *CacheHandler) ClearUserCache(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid user ID", "User ID must be a valid number")
		return
	}

	err = h.cacheService.ClearUserCache(c.Request.Context(), uint(userID))
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.SuccessWithMessage(c, "User cache cleared successfully", nil)
}
