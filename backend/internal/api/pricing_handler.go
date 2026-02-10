package api

import (
	"api-aggregator/backend/internal/models"
	"api-aggregator/backend/internal/service"
	"net/http"
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    500001,
				"message": "Internal server error",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pricings": pricings,
		"total":    len(pricings),
	})
}

// GetPricingByID handles getting a pricing configuration by ID
func (h *PricingHandler) GetPricingByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    400001,
				"message": "Invalid pricing ID",
			},
		})
		return
	}

	pricing, err := h.pricingService.GetPricingByID(c.Request.Context(), uint(id))
	if err != nil {
		if err == service.ErrPricingNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    404001,
					"message": "Pricing not found",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    500001,
				"message": "Internal server error",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, pricing)
}

// CreatePricing handles creating a new pricing configuration
func (h *PricingHandler) CreatePricing(c *gin.Context) {
	var pricing models.Pricing
	if err := c.ShouldBindJSON(&pricing); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    400001,
				"message": "Invalid request body",
				"details": err.Error(),
			},
		})
		return
	}

	if err := h.pricingService.CreatePricing(c.Request.Context(), &pricing); err != nil {
		if err == service.ErrPricingExists {
			c.JSON(http.StatusConflict, gin.H{
				"error": gin.H{
					"code":    409001,
					"message": "Pricing already exists",
					"details": "Pricing for this model and provider already exists",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    500001,
				"message": "Internal server error",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusCreated, pricing)
}

// UpdatePricing handles updating a pricing configuration
func (h *PricingHandler) UpdatePricing(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    400001,
				"message": "Invalid pricing ID",
			},
		})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    400001,
				"message": "Invalid request body",
				"details": err.Error(),
			},
		})
		return
	}

	if err := h.pricingService.UpdatePricing(c.Request.Context(), uint(id), updates); err != nil {
		if err == service.ErrPricingNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    404001,
					"message": "Pricing not found",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    500001,
				"message": "Internal server error",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Pricing updated successfully",
	})
}

// DeletePricing handles deleting a pricing configuration
func (h *PricingHandler) DeletePricing(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    400001,
				"message": "Invalid pricing ID",
			},
		})
		return
	}

	if err := h.pricingService.DeletePricing(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    500001,
				"message": "Internal server error",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Pricing deleted successfully",
	})
}

// GetPricingsByAPIConfig handles getting pricings by API config
func (h *PricingHandler) GetPricingsByAPIConfig(c *gin.Context) {
	apiConfigIDStr := c.Query("api_config_id")
	if apiConfigIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    400001,
				"message": "API config ID parameter is required",
			},
		})
		return
	}

	apiConfigID, err := strconv.ParseUint(apiConfigIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    400001,
				"message": "Invalid API config ID",
			},
		})
		return
	}

	pricings, err := h.pricingService.GetPricingsByAPIConfig(c.Request.Context(), uint(apiConfigID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    500001,
				"message": "Internal server error",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pricings":      pricings,
		"total":         len(pricings),
		"api_config_id": uint(apiConfigID),
	})
}

// InitializeDefaults handles initializing default pricing configurations
func (h *PricingHandler) InitializeDefaults(c *gin.Context) {
	// This endpoint is deprecated as pricing should be set per API config
	c.JSON(http.StatusOK, gin.H{
		"message": "Please set pricing for each API config and model individually",
	})
}
