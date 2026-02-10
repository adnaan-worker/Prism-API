package api

import (
	"api-aggregator/backend/internal/models"
	"api-aggregator/backend/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type LoadBalancerHandler struct {
	lbService      *service.LoadBalancerService
	apiConfigService *service.APIConfigService
	modelService   *service.ModelService
}

func NewLoadBalancerHandler(
	lbService *service.LoadBalancerService,
	apiConfigService *service.APIConfigService,
	modelService *service.ModelService,
) *LoadBalancerHandler {
	return &LoadBalancerHandler{
		lbService:      lbService,
		apiConfigService: apiConfigService,
		modelService:   modelService,
	}
}

// GetConfigs handles getting all load balancer configurations
func (h *LoadBalancerHandler) GetConfigs(c *gin.Context) {
	configs, err := h.lbService.GetAllConfigs(c.Request.Context())
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

	c.JSON(http.StatusOK, configs)
}

// GetConfig handles getting a specific load balancer configuration
func (h *LoadBalancerHandler) GetConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    400001,
				"message": "Invalid request",
				"details": "Invalid configuration ID",
			},
		})
		return
	}

	config, err := h.lbService.GetConfigByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    404001,
				"message": "Configuration not found",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, config)
}

// CreateConfig handles creating a new load balancer configuration
func (h *LoadBalancerHandler) CreateConfig(c *gin.Context) {
	var req struct {
		ModelName string `json:"model_name" binding:"required"`
		Strategy  string `json:"strategy" binding:"required"`
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

	// Validate strategy
	validStrategies := map[string]bool{
		"round_robin":          true,
		"weighted_round_robin": true,
		"least_connections":    true,
		"random":               true,
	}
	if !validStrategies[req.Strategy] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    400002,
				"message": "Invalid strategy",
				"details": "Strategy must be one of: round_robin, weighted_round_robin, least_connections, random",
			},
		})
		return
	}

	config := &models.LoadBalancerConfig{
		ModelName: req.ModelName,
		Strategy:  req.Strategy,
		IsActive:  true,
	}

	if err := h.lbService.CreateConfig(c.Request.Context(), config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    500001,
				"message": "Failed to create configuration",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusCreated, config)
}

// UpdateConfig handles updating a load balancer configuration
func (h *LoadBalancerHandler) UpdateConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    400001,
				"message": "Invalid request",
				"details": "Invalid configuration ID",
			},
		})
		return
	}

	var req struct {
		Strategy string `json:"strategy"`
		IsActive *bool  `json:"is_active"`
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

	updates := make(map[string]interface{})
	if req.Strategy != "" {
		validStrategies := map[string]bool{
			"round_robin":          true,
			"weighted_round_robin": true,
			"least_connections":    true,
			"random":               true,
		}
		if !validStrategies[req.Strategy] {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    400002,
					"message": "Invalid strategy",
					"details": "Strategy must be one of: round_robin, weighted_round_robin, least_connections, random",
				},
			})
			return
		}
		updates["strategy"] = req.Strategy
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if err := h.lbService.UpdateConfig(c.Request.Context(), uint(id), updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    500001,
				"message": "Failed to update configuration",
				"details": err.Error(),
			},
		})
		return
	}

	config, _ := h.lbService.GetConfigByID(c.Request.Context(), uint(id))
	c.JSON(http.StatusOK, config)
}

// DeleteConfig handles deleting a load balancer configuration
func (h *LoadBalancerHandler) DeleteConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    400001,
				"message": "Invalid request",
				"details": "Invalid configuration ID",
			},
		})
		return
	}

	if err := h.lbService.DeleteConfig(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    500001,
				"message": "Failed to delete configuration",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Configuration deleted successfully",
	})
}

// GetModelEndpoints handles getting endpoints for a specific model
func (h *LoadBalancerHandler) GetModelEndpoints(c *gin.Context) {
	modelName := c.Param("model")
	if modelName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    400001,
				"message": "Invalid request",
				"details": "Model name is required",
			},
		})
		return
	}

	// Get all API configs for this model
	configs, err := h.apiConfigService.GetConfigsByModel(c.Request.Context(), modelName)
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

	// Convert to endpoint format
	endpoints := make([]gin.H, 0, len(configs))
	for _, config := range configs {
		// Determine health status based on is_active
		healthStatus := "healthy"
		if !config.IsActive {
			healthStatus = "inactive"
		}

		endpoints = append(endpoints, gin.H{
			"config_id":     config.ID,
			"config_name":   config.Name,
			"type":          config.Type,
			"base_url":      config.BaseURL,
			"priority":      config.Priority,
			"weight":        config.Weight,
			"is_active":     config.IsActive,
			"health_status": healthStatus,
			"response_time": nil, // Can be implemented with metrics collection
			"success_rate":  nil, // Can be implemented with metrics collection
		})
	}

	c.JSON(http.StatusOK, endpoints)
}

// GetAvailableModels handles getting all available models
func (h *LoadBalancerHandler) GetAvailableModels(c *gin.Context) {
	models, err := h.modelService.GetAllModels(c.Request.Context())
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

	// Extract unique model names
	modelNames := make([]string, 0, len(models))
	for _, model := range models {
		modelNames = append(modelNames, model.Name)
	}

	c.JSON(http.StatusOK, modelNames)
}
