package api

import (
	"api-aggregator/backend/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ModelHandler struct {
	modelService *service.ModelService
}

func NewModelHandler(modelService *service.ModelService) *ModelHandler {
	return &ModelHandler{
		modelService: modelService,
	}
}

// GetAllModels handles getting all available models
func (h *ModelHandler) GetAllModels(c *gin.Context) {
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

	c.JSON(http.StatusOK, gin.H{
		"models": models,
		"total":  len(models),
	})
}

// GetModelsByProvider handles getting models filtered by provider
func (h *ModelHandler) GetModelsByProvider(c *gin.Context) {
	provider := c.Query("provider")
	if provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    400001,
				"message": "Invalid request",
				"details": "Provider parameter is required",
			},
		})
		return
	}

	models, err := h.modelService.GetModelsByProvider(c.Request.Context(), provider)
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
		"models":   models,
		"total":    len(models),
		"provider": provider,
	})
}

// GetModelInfo handles getting information about a specific model
func (h *ModelHandler) GetModelInfo(c *gin.Context) {
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

	modelInfo, err := h.modelService.GetModelInfo(c.Request.Context(), modelName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    404001,
				"message": "Model not found",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"model": modelInfo,
	})
}
