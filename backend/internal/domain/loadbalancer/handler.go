package loadbalancer

import (
	"api-aggregator/backend/pkg/query"
	"api-aggregator/backend/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handler 负载均衡配置处理器
type Handler struct {
	service Service
}

// NewHandler 创建负载均衡配置处理器实例
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// CreateConfig 创建负载均衡配置
// @Summary 创建负载均衡配置
// @Tags LoadBalancer
// @Accept json
// @Produce json
// @Param request body CreateConfigRequest true "创建请求"
// @Success 201 {object} ConfigResponse
// @Router /api/v1/load-balancer/configs [post]
func (h *Handler) CreateConfig(c *gin.Context) {
	var req CreateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	config, err := h.service.CreateConfig(c.Request.Context(), &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Created(c, config)
}

// UpdateConfig 更新负载均衡配置
// @Summary 更新负载均衡配置
// @Tags LoadBalancer
// @Accept json
// @Produce json
// @Param id path int true "配置ID"
// @Param request body UpdateConfigRequest true "更新请求"
// @Success 200 {object} ConfigResponse
// @Router /api/v1/load-balancer/configs/{id} [put]
func (h *Handler) UpdateConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid config id")
		return
	}

	var req UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	config, err := h.service.UpdateConfig(c.Request.Context(), uint(id), &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, config)
}

// DeleteConfig 删除负载均衡配置
// @Summary 删除负载均衡配置
// @Tags LoadBalancer
// @Produce json
// @Param id path int true "配置ID"
// @Success 200 {object} response.Response
// @Router /api/v1/load-balancer/configs/{id} [delete]
func (h *Handler) DeleteConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid config id")
		return
	}

	if err := h.service.DeleteConfig(c.Request.Context(), uint(id)); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "config deleted successfully"})
}

// GetConfig 获取负载均衡配置
// @Summary 获取负载均衡配置
// @Tags LoadBalancer
// @Produce json
// @Param id path int true "配置ID"
// @Success 200 {object} ConfigResponse
// @Router /api/v1/load-balancer/configs/{id} [get]
func (h *Handler) GetConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid config id")
		return
	}

	config, err := h.service.GetConfig(c.Request.Context(), uint(id))
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, config)
}

// GetConfigByModel 根据模型名称获取负载均衡配置
// @Summary 根据模型名称获取负载均衡配置
// @Tags LoadBalancer
// @Produce json
// @Param model path string true "模型名称"
// @Success 200 {object} ConfigResponse
// @Router /api/v1/load-balancer/models/{model}/config [get]
func (h *Handler) GetConfigByModel(c *gin.Context) {
	modelName := c.Param("model")
	if modelName == "" {
		response.BadRequest(c, "model name is required")
		return
	}

	config, err := h.service.GetConfigByModel(c.Request.Context(), modelName)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, config)
}

// ListConfigs 查询负载均衡配置列表
// @Summary 查询负载均衡配置列表
// @Tags LoadBalancer
// @Produce json
// @Param model_name query string false "模型名称"
// @Param strategy query string false "策略"
// @Param is_active query bool false "是否激活"
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Param sort_by query string false "排序字段"
// @Param sort_order query string false "排序方向"
// @Success 200 {object} ConfigListResponse
// @Router /api/v1/load-balancer/configs [get]
func (h *Handler) ListConfigs(c *gin.Context) {
	// 构建过滤器
	filter := &ConfigFilter{}
	if modelName := c.Query("model_name"); modelName != "" {
		filter.ModelName = &modelName
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

	configs, err := h.service.ListConfigs(c.Request.Context(), filter, opts)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, configs)
}

// ActivateConfig 激活负载均衡配置
// @Summary 激活负载均衡配置
// @Tags LoadBalancer
// @Produce json
// @Param id path int true "配置ID"
// @Success 200 {object} response.Response
// @Router /api/v1/load-balancer/configs/{id}/activate [post]
func (h *Handler) ActivateConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid config id")
		return
	}

	if err := h.service.ActivateConfig(c.Request.Context(), uint(id)); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "config activated successfully"})
}

// DeactivateConfig 停用负载均衡配置
// @Summary 停用负载均衡配置
// @Tags LoadBalancer
// @Produce json
// @Param id path int true "配置ID"
// @Success 200 {object} response.Response
// @Router /api/v1/load-balancer/configs/{id}/deactivate [post]
func (h *Handler) DeactivateConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid config id")
		return
	}

	if err := h.service.DeactivateConfig(c.Request.Context(), uint(id)); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "config deactivated successfully"})
}

// GetModelEndpoints 获取模型端点列表
// @Summary 获取模型端点列表
// @Tags LoadBalancer
// @Produce json
// @Param model path string true "模型名称"
// @Success 200 {object} ModelEndpointsResponse
// @Router /api/v1/admin/load-balancer/models/{model}/endpoints [get]
func (h *Handler) GetModelEndpoints(c *gin.Context) {
	modelName := c.Param("model")
	if modelName == "" {
		response.BadRequest(c, "model name is required")
		return
	}

	endpoints, err := h.service.GetModelEndpoints(c.Request.Context(), modelName)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, endpoints)
}
