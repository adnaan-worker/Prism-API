package service

import (
	"api-aggregator/backend/internal/models"
	"api-aggregator/backend/internal/repository"
	"context"
	"errors"
)

var (
	ErrLBConfigNotFound = errors.New("load balancer configuration not found")
	ErrLBConfigExists   = errors.New("load balancer configuration already exists for this model")
)

// LoadBalancerRepositoryInterface defines the interface for load balancer repository operations
type LoadBalancerRepositoryInterface interface {
	FindAll(ctx context.Context) ([]*models.LoadBalancerConfig, error)
	FindByID(ctx context.Context, id uint) (*models.LoadBalancerConfig, error)
	FindByModel(ctx context.Context, modelName string) (*models.LoadBalancerConfig, error)
	Create(ctx context.Context, config *models.LoadBalancerConfig) error
	Update(ctx context.Context, id uint, updates map[string]interface{}) error
	Delete(ctx context.Context, id uint) error
}

type LoadBalancerService struct {
	lbRepo LoadBalancerRepositoryInterface
}

func NewLoadBalancerService(lbRepo LoadBalancerRepositoryInterface) *LoadBalancerService {
	return &LoadBalancerService{
		lbRepo: lbRepo,
	}
}

// NewLoadBalancerServiceWithRepo creates a service with a concrete repository (for backward compatibility)
func NewLoadBalancerServiceWithRepo(lbRepo *repository.LoadBalancerRepository) *LoadBalancerService {
	return &LoadBalancerService{
		lbRepo: lbRepo,
	}
}

// GetAllConfigs retrieves all load balancer configurations
func (s *LoadBalancerService) GetAllConfigs(ctx context.Context) ([]*models.LoadBalancerConfig, error) {
	return s.lbRepo.FindAll(ctx)
}

// GetConfigByID retrieves a load balancer configuration by ID
func (s *LoadBalancerService) GetConfigByID(ctx context.Context, id uint) (*models.LoadBalancerConfig, error) {
	config, err := s.lbRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrLBConfigNotFound
	}
	return config, nil
}

// GetConfigByModel retrieves a load balancer configuration by model name
func (s *LoadBalancerService) GetConfigByModel(ctx context.Context, modelName string) (*models.LoadBalancerConfig, error) {
	config, err := s.lbRepo.FindByModel(ctx, modelName)
	if err != nil {
		return nil, ErrLBConfigNotFound
	}
	return config, nil
}

// CreateConfig creates a new load balancer configuration
func (s *LoadBalancerService) CreateConfig(ctx context.Context, config *models.LoadBalancerConfig) error {
	// Check if config already exists for this model
	existing, _ := s.lbRepo.FindByModel(ctx, config.ModelName)
	if existing != nil {
		return ErrLBConfigExists
	}

	return s.lbRepo.Create(ctx, config)
}

// UpdateConfig updates a load balancer configuration
func (s *LoadBalancerService) UpdateConfig(ctx context.Context, id uint, updates map[string]interface{}) error {
	// Check if config exists
	_, err := s.lbRepo.FindByID(ctx, id)
	if err != nil {
		return ErrLBConfigNotFound
	}

	return s.lbRepo.Update(ctx, id, updates)
}

// DeleteConfig deletes a load balancer configuration
func (s *LoadBalancerService) DeleteConfig(ctx context.Context, id uint) error {
	// Check if config exists
	_, err := s.lbRepo.FindByID(ctx, id)
	if err != nil {
		return ErrLBConfigNotFound
	}

	return s.lbRepo.Delete(ctx, id)
}

// ActivateConfig activates a load balancer configuration
func (s *LoadBalancerService) ActivateConfig(ctx context.Context, id uint) error {
	return s.UpdateConfig(ctx, id, map[string]interface{}{"is_active": true})
}

// DeactivateConfig deactivates a load balancer configuration
func (s *LoadBalancerService) DeactivateConfig(ctx context.Context, id uint) error {
	return s.UpdateConfig(ctx, id, map[string]interface{}{"is_active": false})
}
