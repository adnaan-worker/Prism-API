package api

import (
	"api-aggregator/backend/internal/service"
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type LogHandler struct {
	logService *service.LogService
}

func NewLogHandler(logService *service.LogService) *LogHandler {
	return &LogHandler{
		logService: logService,
	}
}

// GetLogs handles getting request logs with filters (admin only)
func (h *LogHandler) GetLogs(c *gin.Context) {
	var req service.GetLogsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    400001,
				"message": "Invalid request",
				"details": err.Error(),
			},
		})
		return
	}

	resp, err := h.logService.GetLogs(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidPage) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    400001,
					"message": "Invalid page parameters",
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

	c.JSON(http.StatusOK, resp)
}

// ExportLogs handles exporting request logs to CSV (admin only)
func (h *LogHandler) ExportLogs(c *gin.Context) {
	var req service.GetLogsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    400001,
				"message": "Invalid request",
				"details": err.Error(),
			},
		})
		return
	}

	// Set a large page size for export (max 10000 records)
	req.Page = 1
	req.PageSize = 10000

	resp, err := h.logService.GetLogs(c.Request.Context(), &req)
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

	// Create CSV buffer
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write CSV header
	header := []string{
		"ID",
		"Created At",
		"User ID",
		"API Key ID",
		"API Config ID",
		"Model",
		"Method",
		"Path",
		"Status Code",
		"Response Time (ms)",
		"Tokens Used",
		"Error Message",
	}
	if err := writer.Write(header); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    500001,
				"message": "Failed to write CSV header",
				"details": err.Error(),
			},
		})
		return
	}

	// Write CSV rows
	for _, log := range resp.Logs {
		row := []string{
			strconv.FormatUint(uint64(log.ID), 10),
			log.CreatedAt.Format("2006-01-02 15:04:05"),
			strconv.FormatUint(uint64(log.UserID), 10),
			strconv.FormatUint(uint64(log.APIKeyID), 10),
			strconv.FormatUint(uint64(log.APIConfigID), 10),
			log.Model,
			log.Method,
			log.Path,
			strconv.Itoa(log.StatusCode),
			strconv.Itoa(log.ResponseTime),
			strconv.Itoa(log.TokensUsed),
			log.ErrorMsg,
		}
		if err := writer.Write(row); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    500001,
					"message": "Failed to write CSV row",
					"details": err.Error(),
				},
			})
			return
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    500001,
				"message": "Failed to flush CSV writer",
				"details": err.Error(),
			},
		})
		return
	}

	// Set response headers for CSV download
	filename := fmt.Sprintf("request_logs_%s.csv", c.Query("start_date"))
	if filename == "request_logs_.csv" {
		filename = "request_logs.csv"
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(http.StatusOK, "text/csv", buf.Bytes())
}
