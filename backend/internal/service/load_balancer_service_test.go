package service

import (
	"api-aggregator/backend/internal/models"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLoadBalancerRepository is a mock implementation of LoadBalancerRepository
type MockLoadBalancerRepository struct {
	mock.Mock
}

func (m *MockLoadBalancerRepository) FindAll(ctx context.Context) ([]*models.LoadBalancerConfig, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.LoadBalancerConfig), args.Error(1)
}

func (m *MockLoadBalancerRepository) FindByID(ctx context.Context, id uint) (*models.LoadBalancerConfig, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LoadBalancerConfig), args.Error(1)
}

func (m *MockLoadBalancerRepository) FindByModel(ctx context.Context, modelName string) (*models.LoadBalancerConfig, error) {
	args := m.Called(ctx, modelName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LoadBalancerConfig), args.Error(1)
}

func (m *MockLoadBalancerRepository) Create(ctx context.Context, config *models.LoadBalancerConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockLoadBalancerRepository) Update(ctx context.Context, id uint, updates map[string]interface{}) error {
	args := m.Called(ctx, id, updates)
	return args.Error(0)
}

func (m *MockLoadBalancerRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestLoadBalancerService_CreateConfig(t *testing.T) {
	mockRepo := new(MockLoadBalancerRepository)
	service := NewLoadBalancerService(mockRepo)
	ctx := context.Background()

	config := &models.LoadBalancerConfig{
		ModelName: "gpt-4",
		Strategy:  "round_robin",
		IsActive:  true,
	}

	// Test successful creation
	mockRepo.On("FindByModel", ctx, "gpt-4").Return(nil, assert.AnError).Once()
	mockRepo.On("Create", ctx, config).Return(nil).Once()

	err := service.CreateConfig(ctx, config)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)

	// Test creation when config already exists
	mockRepo.On("FindByModel", ctx, "gpt-4").Return(config, nil).Once()

	err = service.CreateConfig(ctx, config)
	assert.Error(t, err)
	assert.Equal(t, ErrLBConfigExists, err)
	mockRepo.AssertExpectations(t)
}

func TestLoadBalancerService_GetConfigByModel(t *testing.T) {
	mockRepo := new(MockLoadBalancerRepository)
	service := NewLoadBalancerService(mockRepo)
	ctx := context.Background()

	expectedConfig := &models.LoadBalancerConfig{
		ID:        1,
		ModelName: "gpt-4",
		Strategy:  "round_robin",
		IsActive:  true,
	}

	// Test successful retrieval
	mockRepo.On("FindByModel", ctx, "gpt-4").Return(expectedConfig, nil).Once()

	config, err := service.GetConfigByModel(ctx, "gpt-4")
	assert.NoError(t, err)
	assert.Equal(t, expectedConfig, config)
	mockRepo.AssertExpectations(t)

	// Test config not found
	mockRepo.On("FindByModel", ctx, "gpt-5").Return(nil, assert.AnError).Once()

	config, err = service.GetConfigByModel(ctx, "gpt-5")
	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Equal(t, ErrLBConfigNotFound, err)
	mockRepo.AssertExpectations(t)
}

func TestLoadBalancerService_UpdateConfig(t *testing.T) {
	mockRepo := new(MockLoadBalancerRepository)
	service := NewLoadBalancerService(mockRepo)
	ctx := context.Background()

	config := &models.LoadBalancerConfig{
		ID:        1,
		ModelName: "gpt-4",
		Strategy:  "round_robin",
		IsActive:  true,
	}

	updates := map[string]interface{}{
		"strategy": "weighted_round_robin",
	}

	// Test successful update
	mockRepo.On("FindByID", ctx, uint(1)).Return(config, nil).Once()
	mockRepo.On("Update", ctx, uint(1), updates).Return(nil).Once()

	err := service.UpdateConfig(ctx, 1, updates)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)

	// Test update when config not found
	mockRepo.On("FindByID", ctx, uint(999)).Return(nil, assert.AnError).Once()

	err = service.UpdateConfig(ctx, 999, updates)
	assert.Error(t, err)
	assert.Equal(t, ErrLBConfigNotFound, err)
	mockRepo.AssertExpectations(t)
}

func TestLoadBalancerService_DeleteConfig(t *testing.T) {
	mockRepo := new(MockLoadBalancerRepository)
	service := NewLoadBalancerService(mockRepo)
	ctx := context.Background()

	config := &models.LoadBalancerConfig{
		ID:        1,
		ModelName: "gpt-4",
		Strategy:  "round_robin",
		IsActive:  true,
	}

	// Test successful deletion
	mockRepo.On("FindByID", ctx, uint(1)).Return(config, nil).Once()
	mockRepo.On("Delete", ctx, uint(1)).Return(nil).Once()

	err := service.DeleteConfig(ctx, 1)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)

	// Test deletion when config not found
	mockRepo.On("FindByID", ctx, uint(999)).Return(nil, assert.AnError).Once()

	err = service.DeleteConfig(ctx, 999)
	assert.Error(t, err)
	assert.Equal(t, ErrLBConfigNotFound, err)
	mockRepo.AssertExpectations(t)
}
