package log

import (
	"api-aggregator/backend/pkg/errors"
	"api-aggregator/backend/pkg/response"
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Handler 日志处理器
type Handler struct {
	service Service
}

// NewHandler 创建日志处理器
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// GetLogs 获取日志列表
// @Summary 获取日志列表
// @Description 获取请求日志列表（管理员）
// @Tags Log
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param user_id query int false "用户ID"
// @Param model query string false "模型名称"
// @Param status_code query int false "状态码"
// @Param start_date query string false "开始日期"
// @Param end_date query string false "结束日期"
// @Success 200 {object} LogListResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/admin/logs [get]
func (h *Handler) GetLogs(c *gin.Context) {
	var req GetLogsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters", err.Error())
		return
	}

	logs, err := h.service.GetLogs(c.Request.Context(), &req)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, logs)
}

// GetLogStats 获取日志统计
// @Summary 获取日志统计
// @Description 获取日志统计信息（管理员）
// @Tags Log
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param start_date query string false "开始日期"
// @Param end_date query string false "结束日期"
// @Success 200 {object} LogStatsResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/admin/logs/stats [get]
func (h *Handler) GetLogStats(c *gin.Context) {
	// 解析日期参数（可选）
	var startDate, endDate *time.Time
	if startStr := c.Query("start_date"); startStr != "" {
		if t, err := time.Parse("2006-01-02", startStr); err == nil {
			startDate = &t
		}
	}
	if endStr := c.Query("end_date"); endStr != "" {
		if t, err := time.Parse("2006-01-02", endStr); err == nil {
			endDate = &t
		}
	}

	stats, err := h.service.GetLogStats(c.Request.Context(), startDate, endDate)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, stats)
}

// ExportLogs 导出日志为CSV
// @Summary 导出日志
// @Description 导出日志为CSV文件（管理员）
// @Tags Log
// @Accept json
// @Produce text/csv
// @Security BearerAuth
// @Param user_id query int false "用户ID"
// @Param model query string false "模型名称"
// @Param status_code query int false "状态码"
// @Param start_date query string false "开始日期"
// @Param end_date query string false "结束日期"
// @Success 200 {file} file
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/admin/logs/export [get]
func (h *Handler) ExportLogs(c *gin.Context) {
	var req GetLogsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters", err.Error())
		return
	}

	// 设置大的页面大小用于导出（最多10000条）
	req.Page = 1
	req.PageSize = 10000

	logs, err := h.service.GetLogs(c.Request.Context(), &req)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	// 创建CSV缓冲区
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// 写入CSV头
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
		response.InternalErrorWithMessage(c, "Failed to write CSV header", err)
		return
	}

	// 写入CSV行
	for _, log := range logs.Logs {
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
			response.InternalErrorWithMessage(c, "Failed to write CSV row", err)
			return
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		response.InternalErrorWithMessage(c, "Failed to flush CSV writer", err)
		return
	}

	// 设置响应头用于CSV下载
	filename := "request_logs.csv"
	if c.Query("start_date") != "" {
		filename = fmt.Sprintf("request_logs_%s.csv", c.Query("start_date"))
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(200, "text/csv", buf.Bytes())
}

// DeleteOldLogs 删除旧日志
// @Summary 删除旧日志
// @Description 删除指定天数之前的日志（管理员）
// @Tags Log
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param days query int true "天数"
// @Success 200 {object} object{deleted=int64,message=string}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/admin/logs/cleanup [delete]
func (h *Handler) DeleteOldLogs(c *gin.Context) {
	daysStr := c.Query("days")
	if daysStr == "" {
		response.BadRequest(c, "Days parameter is required", "")
		return
	}

	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		response.BadRequest(c, "Invalid days parameter", "Days must be a positive integer")
		return
	}

	deleted, err := h.service.DeleteOldLogs(c.Request.Context(), days)
	if err != nil {
		if errors.Is(err, errors.ErrInvalidParam) {
			response.BadRequest(c, err.Error(), "")
			return
		}
		response.InternalError(c, err)
		return
	}

	response.Success(c, gin.H{
		"deleted": deleted,
		"message": fmt.Sprintf("Deleted %d logs older than %d days", deleted, days),
	})
}
