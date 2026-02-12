package pricing

import (
	"api-aggregator/backend/pkg/errors"
	"api-aggregator/backend/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handler 定价处理器
type Handler struct {
	service Service
}

// NewHandler 创建定价处理器
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// CreatePricing 创建定价
// @Summary 创建定价
// @Description 创建新的定价配置（管理员）
// @Tags Pricing
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreatePricingRequest true "创建请求"
// @Success 201 {object} PricingResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/admin/pricings [post]
func (h *Handler) CreatePricing(c *gin.Context) {
	var req CreatePricingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	pricing, err := h.service.CreatePricing(c.Request.Context(), &req)
	if err != nil {
		appErr, ok := err.(*errors.AppError)
		if ok && appErr.Code == 409001 {
			response.Conflict(c, appErr.Message, "")
			return
		}
		response.InternalError(c, err)
		return
	}

	response.Created(c, pricing)
}

// GetPricing 获取定价
// @Summary 获取定价
// @Description 根据ID获取定价详情（管理员）
// @Tags Pricing
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "定价ID"
// @Success 200 {object} PricingResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/admin/pricings/{id} [get]
func (h *Handler) GetPricing(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid pricing ID", "Pricing ID must be a valid number")
		return
	}

	pricing, err := h.service.GetPricing(c.Request.Context(), uint(id))
	if err != nil {
		appErr, ok := err.(*errors.AppError)
		if ok && appErr.Code == 404001 {
			response.NotFound(c, "Pricing not found")
			return
		}
		response.InternalError(c, err)
		return
	}

	response.Success(c, gin.H{"pricing": pricing})
}

// GetPricings 获取定价列表
// @Summary 获取定价列表
// @Description 获取定价列表（管理员）
// @Tags Pricing
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param api_config_id query int false "API配置ID"
// @Param model_name query string false "模型名称"
// @Param is_active query bool false "是否激活"
// @Success 200 {object} PricingListResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/admin/pricings [get]
func (h *Handler) GetPricings(c *gin.Context) {
	var req GetPricingsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters", err.Error())
		return
	}

	pricings, err := h.service.GetPricings(c.Request.Context(), &req)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, pricings)
}

// GetAllPricings 获取所有定价
// @Summary 获取所有定价
// @Description 获取所有定价配置（管理员）
// @Tags Pricing
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object{pricings=[]PricingResponse,total=int}
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/admin/pricings/all [get]
func (h *Handler) GetAllPricings(c *gin.Context) {
	pricings, err := h.service.GetAllPricings(c.Request.Context())
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, gin.H{
		"pricings": pricings,
		"total":    len(pricings),
	})
}

// GetActivePricings 获取激活的定价
// @Summary 获取激活的定价
// @Description 获取所有激活的定价配置（管理员）
// @Tags Pricing
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object{pricings=[]PricingResponse,total=int}
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/admin/pricings/active [get]
func (h *Handler) GetActivePricings(c *gin.Context) {
	pricings, err := h.service.GetActivePricings(c.Request.Context())
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, gin.H{
		"pricings": pricings,
		"total":    len(pricings),
	})
}

// GetPricingsByAPIConfig 根据API配置获取定价
// @Summary 根据API配置获取定价
// @Description 获取指定API配置的所有定价（管理员）
// @Tags Pricing
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param api_config_id query int true "API配置ID"
// @Success 200 {object} object{pricings=[]PricingResponse,total=int}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/admin/pricings/by-config [get]
func (h *Handler) GetPricingsByAPIConfig(c *gin.Context) {
	apiConfigIDStr := c.Query("api_config_id")
	if apiConfigIDStr == "" {
		response.BadRequest(c, "API config ID is required", "")
		return
	}

	apiConfigID, err := strconv.ParseUint(apiConfigIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid API config ID", "API config ID must be a valid number")
		return
	}

	pricings, err := h.service.GetPricingsByAPIConfig(c.Request.Context(), uint(apiConfigID))
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, gin.H{
		"pricings":      pricings,
		"total":         len(pricings),
		"api_config_id": uint(apiConfigID),
	})
}

// UpdatePricing 更新定价
// @Summary 更新定价
// @Description 更新指定ID的定价配置（管理员）
// @Tags Pricing
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "定价ID"
// @Param request body UpdatePricingRequest true "更新请求"
// @Success 200 {object} PricingResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/admin/pricings/{id} [put]
func (h *Handler) UpdatePricing(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid pricing ID", "Pricing ID must be a valid number")
		return
	}

	var req UpdatePricingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	pricing, err := h.service.UpdatePricing(c.Request.Context(), uint(id), &req)
	if err != nil {
		appErr, ok := err.(*errors.AppError)
		if ok && appErr.Code == 404001 {
			response.NotFound(c, "Pricing not found")
			return
		}
		response.InternalError(c, err)
		return
	}

	response.Success(c, gin.H{"pricing": pricing})
}

// DeletePricing 删除定价
// @Summary 删除定价
// @Description 删除指定ID的定价配置（管理员）
// @Tags Pricing
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "定价ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/admin/pricings/{id} [delete]
func (h *Handler) DeletePricing(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid pricing ID", "Pricing ID must be a valid number")
		return
	}

	if err := h.service.DeletePricing(c.Request.Context(), uint(id)); err != nil {
		appErr, ok := err.(*errors.AppError)
		if ok && appErr.Code == 404001 {
			response.NotFound(c, "Pricing not found")
			return
		}
		response.InternalError(c, err)
		return
	}

	response.SuccessWithMessage(c, "Pricing deleted successfully", nil)
}

// CalculateCost 计算成本
// @Summary 计算成本
// @Description 根据模型和token数量计算成本
// @Tags Pricing
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CalculateCostRequest true "计算请求"
// @Success 200 {object} CostCalculationResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/pricing/calculate [post]
func (h *Handler) CalculateCost(c *gin.Context) {
	var req CalculateCostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	result, err := h.service.CalculateCost(c.Request.Context(), &req)
	if err != nil {
		appErr, ok := err.(*errors.AppError)
		if ok && appErr.Code == 404001 {
			response.NotFound(c, "Pricing not found")
			return
		}
		response.InternalError(c, err)
		return
	}

	response.Success(c, result)
}

// BatchCreatePricings 批量创建定价
// @Summary 批量创建定价
// @Description 批量创建定价配置（管理员）
// @Tags Pricing
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body BatchCreatePricingRequest true "批量创建请求"
// @Success 200 {object} BatchCreatePricingResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/admin/pricings/batch [post]
func (h *Handler) BatchCreatePricings(c *gin.Context) {
	var req BatchCreatePricingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	result, err := h.service.BatchCreatePricings(c.Request.Context(), &req)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, result)
}
