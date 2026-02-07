package api

import (
	"api-aggregator/backend/internal/service"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type APIConfigHandler struct {
	configService *service.APIConfigService
}

func NewAPIConfigHandler(configService *service.APIConfigService) *APIConfigHandler {
	return &APIConfigHandler{
		configService: configService,
	}
}

// CreateConfig handles creating a new API configuration
func (h *APIConfigHandler) CreateConfig(c *gin.Context) {
	var req service.CreateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    400001,
				"message": "Invalid request",
				"details": err.Error(),
			},
		})
		return
	}

	config, err := h.configService.CreateConfig(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidConfig) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    400001,
					"message": "Invalid configuration",
					"details": err.Error(),
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

	c.JSON(http.StatusCreated, gin.H{
		"config": config,
	})
}

// GetConfig handles getting an API configuration by ID
func (h *APIConfigHandler) GetConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    400001,
				"message": "Invalid ID",
				"details": "ID must be a positive integer",
			},
		})
		return
	}

	config, err := h.configService.GetConfig(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, service.ErrConfigNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    404001,
					"message": "Configuration not found",
					"details": "The requested API configuration does not exist",
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
		"config": config,
	})
}

// GetAllConfigs handles getting all API configurations
func (h *APIConfigHandler) GetAllConfigs(c *gin.Context) {
	configs, err := h.configService.GetAllConfigs(c.Request.Context())
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
		"configs": configs,
		"total":   len(configs),
	})
}

// GetActiveConfigs handles getting all active API configurations
func (h *APIConfigHandler) GetActiveConfigs(c *gin.Context) {
	configs, err := h.configService.GetActiveConfigs(c.Request.Context())
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
		"configs": configs,
		"total":   len(configs),
	})
}

// UpdateConfig handles updating an API configuration
func (h *APIConfigHandler) UpdateConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    400001,
				"message": "Invalid ID",
				"details": "ID must be a positive integer",
			},
		})
		return
	}

	var req service.UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    400001,
				"message": "Invalid request",
				"details": err.Error(),
			},
		})
		return
	}

	config, err := h.configService.UpdateConfig(c.Request.Context(), uint(id), &req)
	if err != nil {
		if errors.Is(err, service.ErrConfigNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    404001,
					"message": "Configuration not found",
					"details": "The requested API configuration does not exist",
				},
			})
			return
		}
		if errors.Is(err, service.ErrInvalidConfig) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    400001,
					"message": "Invalid configuration",
					"details": err.Error(),
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
		"config": config,
	})
}

// DeleteConfig handles deleting an API configuration
func (h *APIConfigHandler) DeleteConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    400001,
				"message": "Invalid ID",
				"details": "ID must be a positive integer",
			},
		})
		return
	}

	err = h.configService.DeleteConfig(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, service.ErrConfigNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    404001,
					"message": "Configuration not found",
					"details": "The requested API configuration does not exist",
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
		"message": "Configuration deleted successfully",
	})
}

// ActivateConfig handles activating an API configuration
func (h *APIConfigHandler) ActivateConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    400001,
				"message": "Invalid ID",
				"details": "ID must be a positive integer",
			},
		})
		return
	}

	err = h.configService.ActivateConfig(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, service.ErrConfigNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    404001,
					"message": "Configuration not found",
					"details": "The requested API configuration does not exist",
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
		"message": "Configuration activated successfully",
	})
}

// DeactivateConfig handles deactivating an API configuration
func (h *APIConfigHandler) DeactivateConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    400001,
				"message": "Invalid ID",
				"details": "ID must be a positive integer",
			},
		})
		return
	}

	err = h.configService.DeactivateConfig(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, service.ErrConfigNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    404001,
					"message": "Configuration not found",
					"details": "The requested API configuration does not exist",
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
		"message": "Configuration deactivated successfully",
	})
}

// BatchDeleteConfigs handles batch deletion of API configurations
func (h *APIConfigHandler) BatchDeleteConfigs(c *gin.Context) {
	var req struct {
		IDs []uint `json:"ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    400001,
				"message": "Invalid request",
				"details": err.Error(),
			},
		})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    400001,
				"message": "Invalid request",
				"details": "IDs array cannot be empty",
			},
		})
		return
	}

	err := h.configService.BatchDeleteConfigs(c.Request.Context(), req.IDs)
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
		"message": "Configurations deleted successfully",
		"count":   len(req.IDs),
	})
}

// BatchActivateConfigs handles batch activation of API configurations
func (h *APIConfigHandler) BatchActivateConfigs(c *gin.Context) {
	var req struct {
		IDs []uint `json:"ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    400001,
				"message": "Invalid request",
				"details": err.Error(),
			},
		})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    400001,
				"message": "Invalid request",
				"details": "IDs array cannot be empty",
			},
		})
		return
	}

	err := h.configService.BatchActivateConfigs(c.Request.Context(), req.IDs)
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
		"message": "Configurations activated successfully",
		"count":   len(req.IDs),
	})
}

// BatchDeactivateConfigs handles batch deactivation of API configurations
func (h *APIConfigHandler) BatchDeactivateConfigs(c *gin.Context) {
	var req struct {
		IDs []uint `json:"ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    400001,
				"message": "Invalid request",
				"details": err.Error(),
			},
		})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    400001,
				"message": "Invalid request",
				"details": "IDs array cannot be empty",
			},
		})
		return
	}

	err := h.configService.BatchDeactivateConfigs(c.Request.Context(), req.IDs)
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
		"message": "Configurations deactivated successfully",
		"count":   len(req.IDs),
	})
}
