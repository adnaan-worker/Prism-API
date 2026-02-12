package quota

import (
	"api-aggregator/backend/pkg/errors"
	"api-aggregator/backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// Handler 配额处理器
type Handler struct {
	service Service
}

// NewHandler 创建配额处理器
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// GetQuotaInfo 获取配额信息
// @Summary 获取配额信息
// @Description 获取当前用户的配额信息
// @Tags Quota
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} QuotaInfoResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/quota/info [get]
func (h *Handler) GetQuotaInfo(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	quotaInfo, err := h.service.GetQuotaInfo(c.Request.Context(), userID.(uint))
	if err != nil {
		if errors.Is(err, errors.ErrUserNotFound) {
			response.NotFound(c, "User not found")
			return
		}
		response.InternalError(c, err)
		return
	}

	response.Success(c, quotaInfo)
}

// SignIn 每日签到
// @Summary 每日签到
// @Description 用户每日签到获取配额奖励
// @Tags Quota
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SignInResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/quota/sign-in [post]
func (h *Handler) SignIn(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	signInResp, err := h.service.SignIn(c.Request.Context(), userID.(uint))
	if err != nil {
		if errors.Is(err, errors.ErrUserNotFound) {
			response.NotFound(c, "User not found")
			return
		}
		appErr, ok := err.(*errors.AppError)
		if ok && appErr.Code == 409001 {
			response.Conflict(c, appErr.Message, "")
			return
		}
		response.InternalError(c, err)
		return
	}

	response.Success(c, signInResp)
}

// CheckQuota 检查配额
// @Summary 检查配额
// @Description 检查用户配额是否充足
// @Tags Quota
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param amount query int true "需要的配额数量"
// @Success 200 {object} CheckQuotaResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/quota/check [get]
func (h *Handler) CheckQuota(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req CheckQuotaRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters", err.Error())
		return
	}

	checkResp, err := h.service.CheckQuota(c.Request.Context(), userID.(uint), req.Amount)
	if err != nil {
		if errors.Is(err, errors.ErrUserNotFound) {
			response.NotFound(c, "User not found")
			return
		}
		response.InternalError(c, err)
		return
	}

	response.Success(c, checkResp)
}

// GetUsageHistory 获取使用历史
// @Summary 获取使用历史
// @Description 获取用户的配额使用历史
// @Tags Quota
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param days query int false "天数" default(7)
// @Success 200 {object} UsageHistoryResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/quota/usage-history [get]
func (h *Handler) GetUsageHistory(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req UsageHistoryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters", err.Error())
		return
	}

	// 设置默认值
	if req.Days == 0 {
		req.Days = 7
	}

	historyResp, err := h.service.GetUsageHistory(c.Request.Context(), userID.(uint), req.Days)
	if err != nil {
		if errors.Is(err, errors.ErrUserNotFound) {
			response.NotFound(c, "User not found")
			return
		}
		response.InternalError(c, err)
		return
	}

	response.Success(c, historyResp)
}

// DeductQuota 扣除配额（内部API，仅供系统调用）
// @Summary 扣除配额
// @Description 扣除用户配额（内部API）
// @Tags Quota
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body DeductQuotaRequest true "扣除请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 429 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/quota/deduct [post]
func (h *Handler) DeductQuota(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req DeductQuotaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	if err := h.service.DeductQuota(c.Request.Context(), userID.(uint), req.Amount); err != nil {
		if errors.Is(err, errors.ErrUserNotFound) {
			response.NotFound(c, "User not found")
			return
		}
		if errors.Is(err, errors.ErrQuotaExceeded) {
			response.TooManyRequests(c, "Quota exceeded")
			return
		}
		if errors.Is(err, errors.ErrInvalidParam) {
			response.BadRequest(c, err.Error(), "")
			return
		}
		response.InternalError(c, err)
		return
	}

	response.SuccessWithMessage(c, "Quota deducted successfully", nil)
}
