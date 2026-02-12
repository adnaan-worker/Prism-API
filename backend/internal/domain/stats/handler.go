package stats

import (
	"api-aggregator/backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// Handler 统计处理器
type Handler struct {
	service Service
}

// NewHandler 创建统计处理器
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// GetStatsOverview 获取统计概览
// @Summary 获取统计概览
// @Description 获取系统统计概览（管理员）
// @Tags Stats
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} GetStatsOverviewResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/admin/stats/overview [get]
func (h *Handler) GetStatsOverview(c *gin.Context) {
	stats, err := h.service.GetStatsOverview(c.Request.Context())
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, stats)
}

// GetRequestTrend 获取请求趋势
// @Summary 获取请求趋势
// @Description 获取请求趋势数据（管理员）
// @Tags Stats
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param days query int false "天数" default(7)
// @Success 200 {object} GetRequestTrendResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/admin/stats/trend [get]
func (h *Handler) GetRequestTrend(c *gin.Context) {
	var req GetRequestTrendRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters", err.Error())
		return
	}

	trend, err := h.service.GetRequestTrend(c.Request.Context(), &req)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, trend)
}

// GetModelUsage 获取模型使用统计
// @Summary 获取模型使用统计
// @Description 获取模型使用统计数据（管理员）
// @Tags Stats
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "限制数量" default(10)
// @Success 200 {object} GetModelUsageResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/admin/stats/models [get]
func (h *Handler) GetModelUsage(c *gin.Context) {
	var req GetModelUsageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters", err.Error())
		return
	}

	usage, err := h.service.GetModelUsage(c.Request.Context(), &req)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, usage)
}

// GetUserGrowth 获取用户增长趋势
// @Summary 获取用户增长趋势
// @Description 获取用户增长趋势数据（管理员）
// @Tags Stats
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param days query int false "天数" default(30)
// @Success 200 {object} GetUserGrowthResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/admin/stats/user-growth [get]
func (h *Handler) GetUserGrowth(c *gin.Context) {
	var req GetUserGrowthRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters", err.Error())
		return
	}

	growth, err := h.service.GetUserGrowth(c.Request.Context(), &req)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, growth)
}

// GetTokenUsage 获取Token使用统计
// @Summary 获取Token使用统计
// @Description 获取Token使用统计数据（管理员）
// @Tags Stats
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param days query int false "天数" default(7)
// @Success 200 {object} GetTokenUsageResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/admin/stats/token-usage [get]
func (h *Handler) GetTokenUsage(c *gin.Context) {
	var req GetTokenUsageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters", err.Error())
		return
	}

	usage, err := h.service.GetTokenUsage(c.Request.Context(), &req)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, usage)
}
