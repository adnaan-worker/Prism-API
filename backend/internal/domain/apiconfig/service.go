package apiconfig

import (
	"api-aggregator/backend/pkg/errors"
	"api-aggregator/backend/pkg/logger"
	"api-aggregator/backend/pkg/query"
	"context"
)

// Service API配置服务接口
type Service interface {
	CreateConfig(ctx context.Context, req *CreateConfigRequest) (*ConfigResponse, error)
	GetConfig(ctx context.Context, id uint) (*ConfigResponse, error)
	GetConfigs(ctx context.Context, req *GetConfigsRequest) (*ConfigListResponse, error)
	GetAllConfigs(ctx context.Context) ([]*ConfigResponse, error)
	GetActiveConfigs(ctx context.Context) ([]*ConfigResponse, error)
	GetConfigsByModel(ctx context.Context, model string) ([]*ConfigResponse, error)
	UpdateConfig(ctx context.Context, id uint, req *UpdateConfigRequest) (*ConfigResponse, error)
	DeleteConfig(ctx context.Context, id uint) error
	ActivateConfig(ctx context.Context, id uint) error
	DeactivateConfig(ctx context.Context, id uint) error
	BatchDeleteConfigs(ctx context.Context, ids []uint) (*BatchOperationResponse, error)
	BatchActivateConfigs(ctx context.Context, ids []uint) (*BatchOperationResponse, error)
	BatchDeactivateConfigs(ctx context.Context, ids []uint) (*BatchOperationResponse, error)
}

// service API配置服务实现
type service struct {
	repo   Repository
	logger logger.Logger
}

// NewService 创建API配置服务
func NewService(repo Repository, logger logger.Logger) Service {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

// CreateConfig 创建配置
func (s *service) CreateConfig(ctx context.Context, req *CreateConfigRequest) (*ConfigResponse, error) {
	// 设置默认值
	priority := req.Priority
	if priority == 0 {
		priority = 100
	}
	weight := req.Weight
	if weight == 0 {
		weight = 1
	}
	timeout := req.Timeout
	if timeout == 0 {
		timeout = 30
	}

	// 创建配置
	config := &APIConfig{
		Name:     req.Name,
		Type:     req.Type,
		BaseURL:  req.BaseURL,
		APIKey:   req.APIKey,
		Models:   req.Models,
		Headers:  req.Headers,
		Metadata: req.Metadata,
		IsActive: true,
		Priority: priority,
		Weight:   weight,
		MaxRPS:   req.MaxRPS,
		Timeout:  timeout,
	}

	if err := s.repo.Create(ctx, config); err != nil {
		s.logger.Error("Failed to create config",
			logger.String("name", req.Name),
			logger.String("type", req.Type),
			logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to create config")
	}

	s.logger.Info("Config created successfully",
		logger.Uint("config_id", config.ID),
		logger.String("name", config.Name),
		logger.String("type", config.Type))

	return config.ToResponse(), nil
}

// GetConfig 获取配置
func (s *service) GetConfig(ctx context.Context, id uint) (*ConfigResponse, error) {
	config, err := s.repo.FindByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get config", logger.Uint("config_id", id), logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to get config")
	}
	if config == nil {
		return nil, errors.ErrAPIConfigNotFound
	}

	return config.ToResponse(), nil
}

// GetConfigs 获取配置列表
func (s *service) GetConfigs(ctx context.Context, req *GetConfigsRequest) (*ConfigListResponse, error) {
	// 设置默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}

	// 构建过滤条件
	var filters []query.Filter
	if req.Type != "" {
		filters = append(filters, query.Filter{
			Field:    "type",
			Operator: "=",
			Value:    req.Type,
		})
	}
	if req.IsActive != nil {
		filters = append(filters, query.Filter{
			Field:    "is_active",
			Operator: "=",
			Value:    *req.IsActive,
		})
	}
	if req.Model != "" {
		// 使用 PostgreSQL JSONB 查询
		filters = append(filters, query.Filter{
			Field:    "models",
			Operator: "@>",
			Value:    `["` + req.Model + `"]`,
		})
	}

	// 构建排序
	sorts := []query.Sort{
		{Field: "priority", Desc: false},
		{Field: "created_at", Desc: true},
	}

	// 构建分页
	pagination := &query.Pagination{
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	// 查询配置列表
	configs, total, err := s.repo.List(ctx, filters, sorts, pagination)
	if err != nil {
		s.logger.Error("Failed to get configs", logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to get configs")
	}

	return &ConfigListResponse{
		Configs:  ToResponseList(configs),
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// GetAllConfigs 获取所有配置
func (s *service) GetAllConfigs(ctx context.Context) ([]*ConfigResponse, error) {
	configs, err := s.repo.FindAll(ctx)
	if err != nil {
		s.logger.Error("Failed to get all configs", logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to get all configs")
	}

	return ToResponseList(configs), nil
}

// GetActiveConfigs 获取所有激活的配置
func (s *service) GetActiveConfigs(ctx context.Context) ([]*ConfigResponse, error) {
	configs, err := s.repo.FindActive(ctx)
	if err != nil {
		s.logger.Error("Failed to get active configs", logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to get active configs")
	}

	return ToResponseList(configs), nil
}

// GetConfigsByModel 根据模型获取配置
func (s *service) GetConfigsByModel(ctx context.Context, model string) ([]*ConfigResponse, error) {
	configs, err := s.repo.FindByModel(ctx, model)
	if err != nil {
		s.logger.Error("Failed to get configs by model",
			logger.String("model", model),
			logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to get configs by model")
	}

	return ToResponseList(configs), nil
}

// UpdateConfig 更新配置
func (s *service) UpdateConfig(ctx context.Context, id uint, req *UpdateConfigRequest) (*ConfigResponse, error) {
	// 查找配置
	config, err := s.repo.FindByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get config", logger.Uint("config_id", id), logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to get config")
	}
	if config == nil {
		return nil, errors.ErrAPIConfigNotFound
	}

	// 更新字段
	if req.Name != "" {
		config.Name = req.Name
	}
	if req.Type != "" {
		config.Type = req.Type
	}
	if req.BaseURL != "" {
		config.BaseURL = req.BaseURL
	}
	if req.APIKey != "" {
		config.APIKey = req.APIKey
	}
	if len(req.Models) > 0 {
		config.Models = req.Models
	}
	if req.Headers != nil {
		config.Headers = req.Headers
	}
	if req.Metadata != nil {
		config.Metadata = req.Metadata
	}
	if req.Priority != nil {
		config.Priority = *req.Priority
	}
	if req.Weight != nil {
		config.Weight = *req.Weight
	}
	if req.MaxRPS != nil {
		config.MaxRPS = *req.MaxRPS
	}
	if req.Timeout != nil {
		config.Timeout = *req.Timeout
	}
	if req.IsActive != nil {
		config.IsActive = *req.IsActive
	}

	// 保存更新
	if err := s.repo.Update(ctx, config); err != nil {
		s.logger.Error("Failed to update config",
			logger.Uint("config_id", id),
			logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to update config")
	}

	s.logger.Info("Config updated successfully",
		logger.Uint("config_id", id),
		logger.String("name", config.Name))

	return config.ToResponse(), nil
}

// DeleteConfig 删除配置
func (s *service) DeleteConfig(ctx context.Context, id uint) error {
	// 检查配置是否存在
	config, err := s.repo.FindByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get config", logger.Uint("config_id", id), logger.Error(err))
		return errors.Wrap(err, 500002, "Failed to get config")
	}
	if config == nil {
		return errors.ErrAPIConfigNotFound
	}

	// 删除配置
	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to delete config",
			logger.Uint("config_id", id),
			logger.Error(err))
		return errors.Wrap(err, 500002, "Failed to delete config")
	}

	s.logger.Info("Config deleted successfully",
		logger.Uint("config_id", id),
		logger.String("name", config.Name))

	return nil
}

// ActivateConfig 激活配置
func (s *service) ActivateConfig(ctx context.Context, id uint) error {
	// 检查配置是否存在
	config, err := s.repo.FindByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get config", logger.Uint("config_id", id), logger.Error(err))
		return errors.Wrap(err, 500002, "Failed to get config")
	}
	if config == nil {
		return errors.ErrAPIConfigNotFound
	}

	// 激活配置
	if err := s.repo.UpdateStatus(ctx, id, true); err != nil {
		s.logger.Error("Failed to activate config",
			logger.Uint("config_id", id),
			logger.Error(err))
		return errors.Wrap(err, 500002, "Failed to activate config")
	}

	s.logger.Info("Config activated successfully",
		logger.Uint("config_id", id),
		logger.String("name", config.Name))

	return nil
}

// DeactivateConfig 停用配置
func (s *service) DeactivateConfig(ctx context.Context, id uint) error {
	// 检查配置是否存在
	config, err := s.repo.FindByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get config", logger.Uint("config_id", id), logger.Error(err))
		return errors.Wrap(err, 500002, "Failed to get config")
	}
	if config == nil {
		return errors.ErrAPIConfigNotFound
	}

	// 停用配置
	if err := s.repo.UpdateStatus(ctx, id, false); err != nil {
		s.logger.Error("Failed to deactivate config",
			logger.Uint("config_id", id),
			logger.Error(err))
		return errors.Wrap(err, 500002, "Failed to deactivate config")
	}

	s.logger.Info("Config deactivated successfully",
		logger.Uint("config_id", id),
		logger.String("name", config.Name))

	return nil
}

// BatchDeleteConfigs 批量删除配置
func (s *service) BatchDeleteConfigs(ctx context.Context, ids []uint) (*BatchOperationResponse, error) {
	if len(ids) == 0 {
		return nil, errors.ErrInvalidParam.WithDetails("IDs array cannot be empty")
	}

	if err := s.repo.BatchDelete(ctx, ids); err != nil {
		s.logger.Error("Failed to batch delete configs",
			logger.Int("count", len(ids)),
			logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to batch delete configs")
	}

	s.logger.Info("Configs deleted successfully", logger.Int("count", len(ids)))

	return &BatchOperationResponse{
		Message: "Configurations deleted successfully",
		Count:   len(ids),
	}, nil
}

// BatchActivateConfigs 批量激活配置
func (s *service) BatchActivateConfigs(ctx context.Context, ids []uint) (*BatchOperationResponse, error) {
	if len(ids) == 0 {
		return nil, errors.ErrInvalidParam.WithDetails("IDs array cannot be empty")
	}

	if err := s.repo.BatchUpdateStatus(ctx, ids, true); err != nil {
		s.logger.Error("Failed to batch activate configs",
			logger.Int("count", len(ids)),
			logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to batch activate configs")
	}

	s.logger.Info("Configs activated successfully", logger.Int("count", len(ids)))

	return &BatchOperationResponse{
		Message: "Configurations activated successfully",
		Count:   len(ids),
	}, nil
}

// BatchDeactivateConfigs 批量停用配置
func (s *service) BatchDeactivateConfigs(ctx context.Context, ids []uint) (*BatchOperationResponse, error) {
	if len(ids) == 0 {
		return nil, errors.ErrInvalidParam.WithDetails("IDs array cannot be empty")
	}

	if err := s.repo.BatchUpdateStatus(ctx, ids, false); err != nil {
		s.logger.Error("Failed to batch deactivate configs",
			logger.Int("count", len(ids)),
			logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to batch deactivate configs")
	}

	s.logger.Info("Configs deactivated successfully", logger.Int("count", len(ids)))

	return &BatchOperationResponse{
		Message: "Configurations deactivated successfully",
		Count:   len(ids),
	}, nil
}
