package loadbalancer

import (
	"api-aggregator/backend/pkg/errors"
	"api-aggregator/backend/pkg/query"
	"context"
)

// Service 负载均衡配置服务接口
type Service interface {
	CreateConfig(ctx context.Context, req *CreateConfigRequest) (*ConfigResponse, error)
	UpdateConfig(ctx context.Context, id uint, req *UpdateConfigRequest) (*ConfigResponse, error)
	DeleteConfig(ctx context.Context, id uint) error
	GetConfig(ctx context.Context, id uint) (*ConfigResponse, error)
	GetConfigByModel(ctx context.Context, modelName string) (*ConfigResponse, error)
	ListConfigs(ctx context.Context, filter *ConfigFilter, opts *query.Options) (*ConfigListResponse, error)
	ActivateConfig(ctx context.Context, id uint) error
	DeactivateConfig(ctx context.Context, id uint) error
}

type service struct {
	repo Repository
}

// NewService 创建负载均衡配置服务实例
func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

// CreateConfig 创建负载均衡配置
func (s *service) CreateConfig(ctx context.Context, req *CreateConfigRequest) (*ConfigResponse, error) {
	// 验证策略
	if !IsValidStrategy(req.Strategy) {
		return nil, errors.NewValidationError("invalid strategy", map[string]string{
			"strategy": "must be one of: round_robin, weighted_round_robin, least_connections, random",
		})
	}

	// 检查模型是否已存在配置
	exists, err := s.repo.ExistsByModel(ctx, req.ModelName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to check existing config")
	}
	if exists {
		return nil, errors.NewConflictError("load balancer configuration already exists for this model")
	}

	// 创建配置
	config := &LoadBalancerConfig{
		ModelName: req.ModelName,
		Strategy:  req.Strategy,
		IsActive:  true,
	}

	if err := s.repo.Create(ctx, config); err != nil {
		return nil, errors.Wrap(err, "failed to create config")
	}

	return ToConfigResponse(config), nil
}

// UpdateConfig 更新负载均衡配置
func (s *service) UpdateConfig(ctx context.Context, id uint, req *UpdateConfigRequest) (*ConfigResponse, error) {
	// 查找配置
	config, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.NewNotFoundError("load balancer configuration not found")
	}

	// 更新字段
	if req.Strategy != "" {
		if !IsValidStrategy(req.Strategy) {
			return nil, errors.NewValidationError("invalid strategy", map[string]string{
				"strategy": "must be one of: round_robin, weighted_round_robin, least_connections, random",
			})
		}
		config.Strategy = req.Strategy
	}
	if req.IsActive != nil {
		config.IsActive = *req.IsActive
	}

	// 保存更新
	if err := s.repo.Update(ctx, config); err != nil {
		return nil, errors.Wrap(err, "failed to update config")
	}

	return ToConfigResponse(config), nil
}

// DeleteConfig 删除负载均衡配置
func (s *service) DeleteConfig(ctx context.Context, id uint) error {
	// 检查配置是否存在
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return errors.NewNotFoundError("load balancer configuration not found")
	}

	// 删除配置
	if err := s.repo.Delete(ctx, id); err != nil {
		return errors.Wrap(err, "failed to delete config")
	}

	return nil
}

// GetConfig 获取负载均衡配置
func (s *service) GetConfig(ctx context.Context, id uint) (*ConfigResponse, error) {
	config, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.NewNotFoundError("load balancer configuration not found")
	}

	return ToConfigResponse(config), nil
}

// GetConfigByModel 根据模型名称获取负载均衡配置
func (s *service) GetConfigByModel(ctx context.Context, modelName string) (*ConfigResponse, error) {
	config, err := s.repo.FindByModel(ctx, modelName)
	if err != nil {
		return nil, errors.NewNotFoundError("load balancer configuration not found")
	}

	return ToConfigResponse(config), nil
}

// ListConfigs 查询负载均衡配置列表
func (s *service) ListConfigs(ctx context.Context, filter *ConfigFilter, opts *query.Options) (*ConfigListResponse, error) {
	configs, total, err := s.repo.List(ctx, filter, opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list configs")
	}

	return ToConfigListResponse(configs, total), nil
}

// ActivateConfig 激活负载均衡配置
func (s *service) ActivateConfig(ctx context.Context, id uint) error {
	config, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return errors.NewNotFoundError("load balancer configuration not found")
	}

	config.Activate()

	if err := s.repo.Update(ctx, config); err != nil {
		return errors.Wrap(err, "failed to activate config")
	}

	return nil
}

// DeactivateConfig 停用负载均衡配置
func (s *service) DeactivateConfig(ctx context.Context, id uint) error {
	config, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return errors.NewNotFoundError("load balancer configuration not found")
	}

	config.Deactivate()

	if err := s.repo.Update(ctx, config); err != nil {
		return errors.Wrap(err, "failed to deactivate config")
	}

	return nil
}
