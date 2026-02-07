package api

import (
	"api-aggregator/backend/internal/service"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type QuotaHandler struct {
	quotaService *service.QuotaService
}

func NewQuotaHandler(quotaService *service.QuotaService) *QuotaHandler {
	return &QuotaHandler{
		quotaService: quotaService,
	}
}

// GetQuotaInfo handles getting user quota information
func (h *QuotaHandler) GetQuotaInfo(c *gin.Context) {
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

	quotaInfo, err := h.quotaService.GetQuotaInfo(c.Request.Context(), userID.(uint))
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

	c.JSON(http.StatusOK, quotaInfo)
}

// SignIn handles daily sign-in
func (h *QuotaHandler) SignIn(c *gin.Context) {
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

	quotaAwarded, err := h.quotaService.SignIn(c.Request.Context(), userID.(uint))
	if err != nil {
		if errors.Is(err, service.ErrAlreadySignedIn) {
			c.JSON(http.StatusConflict, gin.H{
				"error": gin.H{
					"code":    409002,
					"message": "Already signed in",
					"details": "You have already signed in today",
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
		"message":       "Sign in successful",
		"quota_awarded": quotaAwarded,
	})
}

// GetUsageHistory handles getting user usage history
func (h *QuotaHandler) GetUsageHistory(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    401001,
				"message": "Unauthorized",
			},
		})
		return
	}

	// Get days parameter (default 7)
	days := 7
	if daysParam := c.Query("days"); daysParam != "" {
		if d, err := strconv.Atoi(daysParam); err == nil && d > 0 {
			days = d
		}
	}

	history, err := h.quotaService.GetUsageHistory(c.Request.Context(), userID.(uint), days)
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

	c.JSON(http.StatusOK, history)
}
