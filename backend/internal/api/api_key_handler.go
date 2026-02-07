package api

import (
	"api-aggregator/backend/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type APIKeyHandler struct {
	apiKeyService *service.APIKeyService
}

func NewAPIKeyHandler(apiKeyService *service.APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{
		apiKeyService: apiKeyService,
	}
}

// CreateAPIKey handles POST /api/apikeys
func (h *APIKeyHandler) CreateAPIKey(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    401001,
				"message": "Unauthorized",
				"details": "User ID not found in context",
			},
		})
		return
	}

	var req service.CreateAPIKeyRequest
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

	apiKey, err := h.apiKeyService.CreateAPIKey(c.Request.Context(), userID.(uint), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    500001,
				"message": "Internal error",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"api_key": apiKey,
	})
}

// GetAPIKeys handles GET /api/apikeys
func (h *APIKeyHandler) GetAPIKeys(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    401001,
				"message": "Unauthorized",
				"details": "User ID not found in context",
			},
		})
		return
	}

	apiKeys, err := h.apiKeyService.GetAPIKeysByUserID(c.Request.Context(), userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    500001,
				"message": "Internal error",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"keys": apiKeys,
	})
}

// DeleteAPIKey handles DELETE /api/apikeys/:id
func (h *APIKeyHandler) DeleteAPIKey(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    401001,
				"message": "Unauthorized",
				"details": "User ID not found in context",
			},
		})
		return
	}

	// Get key ID from URL parameter
	keyIDStr := c.Param("id")
	keyID, err := strconv.ParseUint(keyIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    400001,
				"message": "Invalid request",
				"details": "Invalid key ID",
			},
		})
		return
	}

	err = h.apiKeyService.DeleteAPIKey(c.Request.Context(), userID.(uint), uint(keyID))
	if err != nil {
		if err == service.ErrAPIKeyNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    404001,
					"message": "Not found",
					"details": "API key not found",
				},
			})
			return
		}
		if err == service.ErrUnauthorizedAccess {
			c.JSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    403001,
					"message": "Forbidden",
					"details": "Unauthorized access to API key",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    500001,
				"message": "Internal error",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "API key deleted successfully",
	})
}
