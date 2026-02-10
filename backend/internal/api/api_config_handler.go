package api

import (
	"api-aggregator/backend/internal/service"
	"api-aggregator/backend/pkg/response"
	"errors"
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
		response.BadRequest(c, "Invalid request", err.Error())
		return
	}

	config, err := h.configService.CreateConfig(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidConfig) {
			response.BadRequest(c, "Invalid configuration", err.Error())
			return
		}
		response.InternalError(c, err)
		return
	}

	response.Created(c, gin.H{"config": config})
}

// GetConfig handles getting an API configuration by ID
func (h *APIConfigHandler) GetConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid ID", "ID must be a positive integer")
		return
	}

	config, err := h.configService.GetConfig(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, service.ErrConfigNotFound) {
			response.NotFound(c, "Configuration not found")
			return
		}
		response.InternalError(c, err)
		return
	}

	response.Success(c, gin.H{"config": config})
}

// GetAllConfigs handles getting all API configurations
func (h *APIConfigHandler) GetAllConfigs(c *gin.Context) {
	configs, err := h.configService.GetAllConfigs(c.Request.Context())
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, gin.H{
		"configs": configs,
		"total":   len(configs),
	})
}

// GetActiveConfigs handles getting all active API configurations
func (h *APIConfigHandler) GetActiveConfigs(c *gin.Context) {
	configs, err := h.configService.GetActiveConfigs(c.Request.Context())
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, gin.H{
		"configs": configs,
		"total":   len(configs),
	})
}

// UpdateConfig handles updating an API configuration
func (h *APIConfigHandler) UpdateConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid ID", "ID must be a positive integer")
		return
	}

	var req service.UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request", err.Error())
		return
	}

	config, err := h.configService.UpdateConfig(c.Request.Context(), uint(id), &req)
	if err != nil {
		if errors.Is(err, service.ErrConfigNotFound) {
			response.NotFound(c, "Configuration not found")
			return
		}
		if errors.Is(err, service.ErrInvalidConfig) {
			response.BadRequest(c, "Invalid configuration", err.Error())
			return
		}
		response.InternalError(c, err)
		return
	}

	response.Success(c, gin.H{"config": config})
}

// DeleteConfig handles deleting an API configuration
func (h *APIConfigHandler) DeleteConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid ID", "ID must be a positive integer")
		return
	}

	err = h.configService.DeleteConfig(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, service.ErrConfigNotFound) {
			response.NotFound(c, "Configuration not found")
			return
		}
		response.InternalError(c, err)
		return
	}

	response.SuccessWithMessage(c, "Configuration deleted successfully", nil)
}

// ActivateConfig handles activating an API configuration
func (h *APIConfigHandler) ActivateConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid ID", "ID must be a positive integer")
		return
	}

	err = h.configService.ActivateConfig(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, service.ErrConfigNotFound) {
			response.NotFound(c, "Configuration not found")
			return
		}
		response.InternalError(c, err)
		return
	}

	response.SuccessWithMessage(c, "Configuration activated successfully", nil)
}

// DeactivateConfig handles deactivating an API configuration
func (h *APIConfigHandler) DeactivateConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid ID", "ID must be a positive integer")
		return
	}

	err = h.configService.DeactivateConfig(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, service.ErrConfigNotFound) {
			response.NotFound(c, "Configuration not found")
			return
		}
		response.InternalError(c, err)
		return
	}

	response.SuccessWithMessage(c, "Configuration deactivated successfully", nil)
}

// BatchDeleteConfigs handles batch deletion of API configurations
func (h *APIConfigHandler) BatchDeleteConfigs(c *gin.Context) {
	var req struct {
		IDs []uint `json:"ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request", err.Error())
		return
	}

	if len(req.IDs) == 0 {
		response.BadRequest(c, "Invalid request", "IDs array cannot be empty")
		return
	}

	err := h.configService.BatchDeleteConfigs(c.Request.Context(), req.IDs)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, gin.H{
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
		response.BadRequest(c, "Invalid request", err.Error())
		return
	}

	if len(req.IDs) == 0 {
		response.BadRequest(c, "Invalid request", "IDs array cannot be empty")
		return
	}

	err := h.configService.BatchActivateConfigs(c.Request.Context(), req.IDs)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, gin.H{
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
		response.BadRequest(c, "Invalid request", err.Error())
		return
	}

	if len(req.IDs) == 0 {
		response.BadRequest(c, "Invalid request", "IDs array cannot be empty")
		return
	}

	err := h.configService.BatchDeactivateConfigs(c.Request.Context(), req.IDs)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, gin.H{
		"message": "Configurations deactivated successfully",
		"count":   len(req.IDs),
	})
}
