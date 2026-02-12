package api

import (
	"api-aggregator/backend/internal/models"
	"api-aggregator/backend/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AccountPoolHandler struct {
	poolService *service.AccountPoolService
}

func NewAccountPoolHandler(poolService *service.AccountPoolService) *AccountPoolHandler {
	return &AccountPoolHandler{
		poolService: poolService,
	}
}

// GetPools 获取所有账号池
func (h *AccountPoolHandler) GetPools(c *gin.Context) {
	provider := c.Query("provider")
	
	pools, err := h.poolService.GetPools(c.Request.Context(), provider)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pools)
}

// GetPool 获取指定账号池
func (h *AccountPoolHandler) GetPool(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pool ID"})
		return
	}

	pool, err := h.poolService.GetPool(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if pool == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pool not found"})
		return
	}

	c.JSON(http.StatusOK, pool)
}

// CreatePool 创建账号池
func (h *AccountPoolHandler) CreatePool(c *gin.Context) {
	var req models.AccountPool
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pool, err := h.poolService.CreatePool(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, pool)
}

// UpdatePool 更新账号池
func (h *AccountPoolHandler) UpdatePool(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pool ID"})
		return
	}

	var req models.AccountPool
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.ID = uint(id)
	pool, err := h.poolService.UpdatePool(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pool)
}

// DeletePool 删除账号池
func (h *AccountPoolHandler) DeletePool(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pool ID"})
		return
	}

	if err := h.poolService.DeletePool(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// UpdatePoolStatus 更新账号池状态
func (h *AccountPoolHandler) UpdatePoolStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pool ID"})
		return
	}

	var req struct {
		IsActive bool `json:"is_active" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pool, err := h.poolService.UpdatePoolStatus(c.Request.Context(), uint(id), req.IsActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pool)
}

// GetPoolStats 获取账号池统计
func (h *AccountPoolHandler) GetPoolStats(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pool ID"})
		return
	}

	stats, err := h.poolService.GetPoolStats(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}
