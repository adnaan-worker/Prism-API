package accountpool

import (
	"api-aggregator/backend/pkg/query"
	"api-aggregator/backend/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handler 账号池处理器
type Handler struct {
	service Service
}

// NewHandler 创建账号池处理器实例
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// CreatePool 创建账号池
// @Summary 创建账号池
// @Tags AccountPool
// @Accept json
// @Produce json
// @Param request body CreatePoolRequest true "创建请求"
// @Success 201 {object} PoolResponse
// @Router /api/v1/account-pools [post]
func (h *Handler) CreatePool(c *gin.Context) {
	var req CreatePoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	pool, err := h.service.CreatePool(c.Request.Context(), &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Created(c, pool)
}

// UpdatePool 更新账号池
// @Summary 更新账号池
// @Tags AccountPool
// @Accept json
// @Produce json
// @Param id path int true "账号池ID"
// @Param request body UpdatePoolRequest true "更新请求"
// @Success 200 {object} PoolResponse
// @Router /api/v1/account-pools/{id} [put]
func (h *Handler) UpdatePool(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid pool id")
		return
	}

	var req UpdatePoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	pool, err := h.service.UpdatePool(c.Request.Context(), uint(id), &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, pool)
}

// DeletePool 删除账号池
// @Summary 删除账号池
// @Tags AccountPool
// @Produce json
// @Param id path int true "账号池ID"
// @Success 200 {object} response.Response
// @Router /api/v1/account-pools/{id} [delete]
func (h *Handler) DeletePool(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid pool id")
		return
	}

	if err := h.service.DeletePool(c.Request.Context(), uint(id)); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "pool deleted successfully"})
}

// GetPool 获取账号池
// @Summary 获取账号池
// @Tags AccountPool
// @Produce json
// @Param id path int true "账号池ID"
// @Success 200 {object} PoolResponse
// @Router /api/v1/account-pools/{id} [get]
func (h *Handler) GetPool(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid pool id")
		return
	}

	pool, err := h.service.GetPool(c.Request.Context(), uint(id))
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, pool)
}

// ListPools 查询账号池列表
// @Summary 查询账号池列表
// @Tags AccountPool
// @Produce json
// @Param provider query string false "提供商"
// @Param strategy query string false "策略"
// @Param is_active query bool false "是否激活"
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Param sort_by query string false "排序字段"
// @Param sort_order query string false "排序方向"
// @Success 200 {object} PoolListResponse
// @Router /api/v1/account-pools [get]
func (h *Handler) ListPools(c *gin.Context) {
	// 构建过滤器
	filter := &PoolFilter{}
	if provider := c.Query("provider"); provider != "" {
		filter.Provider = &provider
	}
	if strategy := c.Query("strategy"); strategy != "" {
		filter.Strategy = &strategy
	}
	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		isActive := isActiveStr == "true"
		filter.IsActive = &isActive
	}

	// 构建查询选项
	opts := query.NewOptionsFromQuery(c)

	pools, err := h.service.ListPools(c.Request.Context(), filter, opts)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, pools)
}

// UpdatePoolStatus 更新账号池状态
// @Summary 更新账号池状态
// @Tags AccountPool
// @Accept json
// @Produce json
// @Param id path int true "账号池ID"
// @Param request body UpdatePoolStatusRequest true "状态更新请求"
// @Success 200 {object} PoolResponse
// @Router /api/v1/account-pools/{id}/status [put]
func (h *Handler) UpdatePoolStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid pool id")
		return
	}

	var req UpdatePoolStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	pool, err := h.service.UpdatePoolStatus(c.Request.Context(), uint(id), req.IsActive)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, pool)
}

// GetPoolStats 获取账号池统计
// @Summary 获取账号池统计
// @Tags AccountPool
// @Produce json
// @Param id path int true "账号池ID"
// @Success 200 {object} PoolStatsResponse
// @Router /api/v1/account-pools/{id}/stats [get]
func (h *Handler) GetPoolStats(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid pool id")
		return
	}

	stats, err := h.service.GetPoolStats(c.Request.Context(), uint(id))
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, stats)
}

// ListRequestLogs 查询请求日志列表
// @Summary 查询请求日志列表
// @Tags AccountPool
// @Produce json
// @Param pool_id query int false "账号池ID"
// @Param credential_id query int false "凭据ID"
// @Param provider query string false "提供商"
// @Param model query string false "模型"
// @Param status_code query int false "状态码"
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Param sort_by query string false "排序字段"
// @Param sort_order query string false "排序方向"
// @Success 200 {object} RequestLogListResponse
// @Router /api/v1/account-pools/request-logs [get]
func (h *Handler) ListRequestLogs(c *gin.Context) {
	// 构建过滤器
	filter := &RequestLogFilter{}
	if poolIDStr := c.Query("pool_id"); poolIDStr != "" {
		if poolID, err := strconv.ParseUint(poolIDStr, 10, 32); err == nil {
			id := uint(poolID)
			filter.PoolID = &id
		}
	}
	if credIDStr := c.Query("credential_id"); credIDStr != "" {
		if credID, err := strconv.ParseUint(credIDStr, 10, 32); err == nil {
			id := uint(credID)
			filter.CredentialID = &id
		}
	}
	if provider := c.Query("provider"); provider != "" {
		filter.Provider = &provider
	}
	if model := c.Query("model"); model != "" {
		filter.Model = &model
	}
	if statusStr := c.Query("status_code"); statusStr != "" {
		if status, err := strconv.Atoi(statusStr); err == nil {
			filter.StatusCode = &status
		}
	}

	// 构建查询选项
	opts := query.NewOptionsFromQuery(c)

	logs, err := h.service.ListRequestLogs(c.Request.Context(), filter, opts)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, logs)
}
