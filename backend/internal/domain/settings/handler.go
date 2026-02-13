package settings

import (
	"api-aggregator/backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// Handler 设置处理器
type Handler struct {
	service Service
}

// NewHandler 创建设置处理器实例
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// GetRuntimeConfig 获取运行时配置
// @Summary 获取运行时配置
// @Tags Settings
// @Produce json
// @Success 200 {object} RuntimeConfigResponse
// @Router /api/v1/admin/settings/runtime [get]
func (h *Handler) GetRuntimeConfig(c *gin.Context) {
	config, err := h.service.GetRuntimeConfig(c.Request.Context())
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, config)
}

// UpdateRuntimeConfig 更新运行时配置
// @Summary 更新运行时配置
// @Tags Settings
// @Accept json
// @Produce json
// @Param request body UpdateRuntimeConfigRequest true "更新请求"
// @Success 200 {object} RuntimeConfigResponse
// @Router /api/v1/admin/settings/runtime [put]
func (h *Handler) UpdateRuntimeConfig(c *gin.Context) {
	var req UpdateRuntimeConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	config, err := h.service.UpdateRuntimeConfig(c.Request.Context(), &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, config)
}

// GetSystemConfig 获取系统配置
// @Summary 获取系统配置
// @Tags Settings
// @Produce json
// @Success 200 {object} SystemConfigResponse
// @Router /api/v1/admin/settings/system [get]
func (h *Handler) GetSystemConfig(c *gin.Context) {
	config, err := h.service.GetSystemConfig(c.Request.Context())
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, config)
}

// UpdatePassword 修改密码
// @Summary 修改管理员密码
// @Tags Settings
// @Accept json
// @Produce json
// @Param request body UpdatePasswordRequest true "修改密码请求"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/settings/password [put]
func (h *Handler) UpdatePassword(c *gin.Context) {
	var req UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	if err := h.service.UpdatePassword(c.Request.Context(), userID.(uint), &req); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "password updated successfully"})
}

// GetDefaultQuota 获取默认配额
// @Summary 获取默认用户配额
// @Tags Settings
// @Produce json
// @Success 200 {object} DefaultQuotaResponse
// @Router /api/v1/admin/settings/default-quota [get]
func (h *Handler) GetDefaultQuota(c *gin.Context) {
	quota, err := h.service.GetDefaultQuota(c.Request.Context())
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, quota)
}

// UpdateDefaultQuota 更新默认配额
// @Summary 更新默认用户配额
// @Tags Settings
// @Accept json
// @Produce json
// @Param request body UpdateDefaultQuotaRequest true "更新请求"
// @Success 200 {object} DefaultQuotaResponse
// @Router /api/v1/admin/settings/default-quota [put]
func (h *Handler) UpdateDefaultQuota(c *gin.Context) {
	var req UpdateDefaultQuotaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	quota, err := h.service.UpdateDefaultQuota(c.Request.Context(), &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, quota)
}

// GetDefaultRateLimit 获取默认速率限制
// @Summary 获取默认速率限制
// @Tags Settings
// @Produce json
// @Success 200 {object} DefaultRateLimitResponse
// @Router /api/v1/admin/settings/default-rate-limit [get]
func (h *Handler) GetDefaultRateLimit(c *gin.Context) {
	rateLimit, err := h.service.GetDefaultRateLimit(c.Request.Context())
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, rateLimit)
}

// UpdateDefaultRateLimit 更新默认速率限制
// @Summary 更新默认速率限制
// @Tags Settings
// @Accept json
// @Produce json
// @Param request body UpdateDefaultRateLimitRequest true "更新请求"
// @Success 200 {object} DefaultRateLimitResponse
// @Router /api/v1/admin/settings/default-rate-limit [put]
func (h *Handler) UpdateDefaultRateLimit(c *gin.Context) {
	var req UpdateDefaultRateLimitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	rateLimit, err := h.service.UpdateDefaultRateLimit(c.Request.Context(), &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, rateLimit)
}
