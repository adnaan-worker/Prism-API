package api

import (
	"api-aggregator/backend/internal/models"
	"api-aggregator/backend/internal/service"
	"api-aggregator/backend/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PricingHandler struct {
	pricingService *service.PricingService
}

func NewPricingHandler(pricingService *service.PricingService) *PricingHandler {
	return &PricingHandler{
		pricingService: pricingService,
	}
}

// GetAllPricings handles getting all pricing configurations
func (h *PricingHandler) GetAllPricings(c *gin.Context) {
	pricings, err := h.pricingService.GetAllPricings(c.Request.Context())
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, gin.H{
		"pricings": pricings,
		"total":    len(pricings),
	})
}

// GetPricingByID handles getting a pricing configuration by ID
func (h *PricingHandler) GetPricingByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid pricing ID")
		return
	}

	pricing, err := h.pricingService.GetPricingByID(c.Request.Context(), uint(id))
	if err != nil {
		if err == service.ErrPricingNotFound {
			response.NotFound(c, "Pricing not found")
			return
		}
		response.InternalError(c, err)
		return
	}

	response.Success(c, pricing)
}

// CreatePricing handles creating a new pricing configuration
func (h *PricingHandler) CreatePricing(c *gin.Context) {
	var pricing models.Pricing
	if err := c.ShouldBindJSON(&pricing); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	if err := h.pricingService.CreatePricing(c.Request.Context(), &pricing); err != nil {
		if err == service.ErrPricingExists {
			response.Conflict(c, "Pricing already exists", "Pricing for this model and provider already exists")
			return
		}
		response.InternalError(c, err)
		return
	}

	response.Created(c, pricing)
}

// UpdatePricing handles updating a pricing configuration
func (h *PricingHandler) UpdatePricing(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid pricing ID")
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	if err := h.pricingService.UpdatePricing(c.Request.Context(), uint(id), updates); err != nil {
		if err == service.ErrPricingNotFound {
			response.NotFound(c, "Pricing not found")
			return
		}
		response.InternalError(c, err)
		return
	}

	response.SuccessWithMessage(c, "Pricing updated successfully", nil)
}

// DeletePricing handles deleting a pricing configuration
func (h *PricingHandler) DeletePricing(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid pricing ID")
		return
	}

	if err := h.pricingService.DeletePricing(c.Request.Context(), uint(id)); err != nil {
		response.InternalError(c, err)
		return
	}

	response.SuccessWithMessage(c, "Pricing deleted successfully", nil)
}

// GetPricingsByAPIConfig handles getting pricings by API config
func (h *PricingHandler) GetPricingsByAPIConfig(c *gin.Context) {
	apiConfigIDStr := c.Query("api_config_id")
	if apiConfigIDStr == "" {
		response.BadRequest(c, "API config ID parameter is required")
		return
	}

	apiConfigID, err := strconv.ParseUint(apiConfigIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid API config ID")
		return
	}

	pricings, err := h.pricingService.GetPricingsByAPIConfig(c.Request.Context(), uint(apiConfigID))
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

// InitializeDefaults handles initializing default pricing configurations
func (h *PricingHandler) InitializeDefaults(c *gin.Context) {
	response.Success(c, gin.H{
		"message": "Please set pricing for each API config and model individually",
	})
}
