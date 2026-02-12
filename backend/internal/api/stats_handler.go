package api

import (
	"api-aggregator/backend/internal/service"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type StatsHandler struct {
	statsService *service.StatsService
}

func NewStatsHandler(statsService *service.StatsService) *StatsHandler {
	return &StatsHandler{
		statsService: statsService,
	}
}

// GetStatsOverview handles getting the statistics overview (admin only)
func (h *StatsHandler) GetStatsOverview(c *gin.Context) {
	stats, err := h.statsService.GetStatsOverview(c.Request.Context())
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

	c.JSON(http.StatusOK, stats)
}

// GetRequestTrend handles getting request trend data
func (h *StatsHandler) GetRequestTrend(c *gin.Context) {
	// Get days parameter (default 7)
	days := 7
	if daysParam := c.Query("days"); daysParam != "" {
		var parsedDays int
		if _, err := fmt.Sscanf(daysParam, "%d", &parsedDays); err == nil && parsedDays > 0 {
			days = parsedDays
		}
	}
	
	trend, err := h.statsService.GetRequestTrend(c.Request.Context(), days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    500001,
				"message": "Failed to get request trend",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, trend)
}

// GetModelUsage handles getting model usage statistics
func (h *StatsHandler) GetModelUsage(c *gin.Context) {
	// Get limit parameter (default 10)
	limit := 10
	if limitParam := c.Query("limit"); limitParam != "" {
		var parsedLimit int
		if _, err := fmt.Sscanf(limitParam, "%d", &parsedLimit); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}
	
	usage, err := h.statsService.GetModelUsage(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    500001,
				"message": "Failed to get model usage",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, usage)
}
