package apiconfig

import (
	"api-aggregator/backend/pkg/errors"
	"api-aggregator/backend/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handler API配置处理器
type Handler struct {
	service Service
}

// NewHandler 创建API配置处理器
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// CreateConfig 创建配置
// @Summary 创建API配置
// @Description 创建新的API配置（管理员）
// @Tags APIConfig
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateConfigRequest true "创建请求"
// @Success 201 {object} ConfigResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/admin/configs [post]
func (h *Handler) CreateConfig(c *gin.Context) {
	var req CreateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	config, err := h.service.CreateConfig(c.Request.Context(), &req)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Created(c, gin.H{"config": config})
}

// GetConfig 获取配置
// @Summary 获取API配置
// @Description 根据ID获取API配置详情（管理员）
// @Tags APIConfig
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "配置ID"
// @Success 200 {object} ConfigResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/admin/configs/{id} [get]
func (h *Handler) GetConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid config ID", "Config ID must be a valid number")
		return
	}

	config, err := h.service.GetConfig(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, errors.ErrAPIConfigNotFound) {
			response.NotFound(c, "Configuration not found")
			return
		}
		response.InternalError(c, err)
		return
	}

	response.Success(c, gin.H{"config": config})
}

// GetConfigs 获取配置列表
// @Summary 获取API配置列表
// @Description 获取API配置列表（管理员）
// @Tags APIConfig
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param type query string false "配置类型" Enums(openai, anthropic, gemini, kiro, custom)
// @Param is_active query bool false "是否激活"
// @Param model query string false "模型名称"
// @Success 200 {object} ConfigListResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/admin/configs [get]
func (h *Handler) GetConfigs(c *gin.Context) {
	var req GetConfigsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters", err.Error())
		return
	}

	configs, err := h.service.GetConfigs(c.Request.Context(), &req)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, configs)
}

// GetAllConfigs 获取所有配置
// @Summary 获取所有API配置
// @Description 获取所有API配置（管理员）
// @Tags APIConfig
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object{configs=[]ConfigResponse,total=int}
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/admin/configs/all [get]
func (h *Handler) GetAllConfigs(c *gin.Context) {
	configs, err := h.service.GetAllConfigs(c.Request.Context())
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, gin.H{
		"configs": configs,
		"total":   len(configs),
	})
}

// GetActiveConfigs 获取激活的配置
// @Summary 获取激活的API配置
// @Description 获取所有激活的API配置（管理员）
// @Tags APIConfig
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object{configs=[]ConfigResponse,total=int}
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/admin/configs/active [get]
func (h *Handler) GetActiveConfigs(c *gin.Context) {
	configs, err := h.service.GetActiveConfigs(c.Request.Context())
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, gin.H{
		"configs": configs,
		"total":   len(configs),
	})
}

// UpdateConfig 更新配置
// @Summary 更新API配置
// @Description 更新指定ID的API配置（管理员）
// @Tags APIConfig
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "配置ID"
// @Param request body UpdateConfigRequest true "更新请求"
// @Success 200 {object} ConfigResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/admin/configs/{id} [put]
func (h *Handler) UpdateConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid config ID", "Config ID must be a valid number")
		return
	}

	var req UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	config, err := h.service.UpdateConfig(c.Request.Context(), uint(id), &req)
	if err != nil {
		if errors.Is(err, errors.ErrAPIConfigNotFound) {
			response.NotFound(c, "Configuration not found")
			return
		}
		response.InternalError(c, err)
		return
	}

	response.Success(c, gin.H{"config": config})
}

// DeleteConfig 删除配置
// @Summary 删除API配置
// @Description 删除指定ID的API配置（管理员）
// @Tags APIConfig
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "配置ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/admin/configs/{id} [delete]
func (h *Handler) DeleteConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid config ID", "Config ID must be a valid number")
		return
	}

	if err := h.service.DeleteConfig(c.Request.Context(), uint(id)); err != nil {
		if errors.Is(err, errors.ErrAPIConfigNotFound) {
			response.NotFound(c, "Configuration not found")
			return
		}
		response.InternalError(c, err)
		return
	}

	response.SuccessWithMessage(c, "Configuration deleted successfully", nil)
}

// ActivateConfig 激活配置
// @Summary 激活API配置
// @Description 激活指定ID的API配置（管理员）
// @Tags APIConfig
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "配置ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/admin/configs/{id}/activate [post]
func (h *Handler) ActivateConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid config ID", "Config ID must be a valid number")
		return
	}

	if err := h.service.ActivateConfig(c.Request.Context(), uint(id)); err != nil {
		if errors.Is(err, errors.ErrAPIConfigNotFound) {
			response.NotFound(c, "Configuration not found")
			return
		}
		response.InternalError(c, err)
		return
	}

	response.SuccessWithMessage(c, "Configuration activated successfully", nil)
}

// DeactivateConfig 停用配置
// @Summary 停用API配置
// @Description 停用指定ID的API配置（管理员）
// @Tags APIConfig
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "配置ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/admin/configs/{id}/deactivate [post]
func (h *Handler) DeactivateConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid config ID", "Config ID must be a valid number")
		return
	}

	if err := h.service.DeactivateConfig(c.Request.Context(), uint(id)); err != nil {
		if errors.Is(err, errors.ErrAPIConfigNotFound) {
			response.NotFound(c, "Configuration not found")
			return
		}
		response.InternalError(c, err)
		return
	}

	response.SuccessWithMessage(c, "Configuration deactivated successfully", nil)
}

// BatchDeleteConfigs 批量删除配置
// @Summary 批量删除API配置
// @Description 批量删除API配置（管理员）
// @Tags APIConfig
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body BatchOperationRequest true "批量操作请求"
// @Success 200 {object} BatchOperationResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/admin/configs/batch/delete [post]
func (h *Handler) BatchDeleteConfigs(c *gin.Context) {
	var req BatchOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	result, err := h.service.BatchDeleteConfigs(c.Request.Context(), req.IDs)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, result)
}

// BatchActivateConfigs 批量激活配置
// @Summary 批量激活API配置
// @Description 批量激活API配置（管理员）
// @Tags APIConfig
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body BatchOperationRequest true "批量操作请求"
// @Success 200 {object} BatchOperationResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/admin/configs/batch/activate [post]
func (h *Handler) BatchActivateConfigs(c *gin.Context) {
	var req BatchOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	result, err := h.service.BatchActivateConfigs(c.Request.Context(), req.IDs)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, result)
}

// BatchDeactivateConfigs 批量停用配置
// @Summary 批量停用API配置
// @Description 批量停用API配置（管理员）
// @Tags APIConfig
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body BatchOperationRequest true "批量操作请求"
// @Success 200 {object} BatchOperationResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/admin/configs/batch/deactivate [post]
func (h *Handler) BatchDeactivateConfigs(c *gin.Context) {
	var req BatchOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	result, err := h.service.BatchDeactivateConfigs(c.Request.Context(), req.IDs)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, result)
}

// FetchModels 从提供商动态获取模型列表
// @Summary 从提供商动态获取模型列表
// @Tags APIConfig
// @Accept json
// @Produce json
// @Param request body FetchModelsRequest true "获取模型请求"
// @Success 200 {object} FetchModelsResponse
// @Router /api/v1/admin/providers/fetch-models [post]
func (h *Handler) FetchModels(c *gin.Context) {
	var req FetchModelsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	result, err := h.service.FetchModels(c.Request.Context(), &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, result)
}

// GetAvailableModels 获取所有可用的模型列表（用于用户端）
// @Summary 获取可用模型列表
// @Description 获取所有激活配置中的可用模型
// @Tags Models
// @Produce json
// @Security BearerAuth
// @Success 200 {object} AvailableModelsResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/models [get]
func (h *Handler) GetAvailableModels(c *gin.Context) {
	models, err := h.service.GetAvailableModels(c.Request.Context())
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, models)
}
