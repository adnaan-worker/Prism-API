package service

import (
	"api-aggregator/backend/internal/models"
	"api-aggregator/backend/internal/repository"
	"context"
	"errors"
	"fmt"
)

var (
	ErrPricingNotFound = errors.New("pricing not found")
	ErrPricingExists   = errors.New("pricing already exists for this model and API config")
)

type PricingService struct {
	pricingRepo *repository.PricingRepository
}

func NewPricingService(pricingRepo *repository.PricingRepository) *PricingService {
	return &PricingService{
		pricingRepo: pricingRepo,
	}
}

// GetAllPricings retrieves all pricing configurations
func (s *PricingService) GetAllPricings(ctx context.Context) ([]*models.Pricing, error) {
	return s.pricingRepo.FindAll(ctx)
}

// GetPricingByID retrieves a pricing configuration by ID
func (s *PricingService) GetPricingByID(ctx context.Context, id uint) (*models.Pricing, error) {
	pricing, err := s.pricingRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if pricing == nil {
		return nil, ErrPricingNotFound
	}
	return pricing, nil
}

// GetPricingByModelAndAPIConfig retrieves pricing by model and API config
func (s *PricingService) GetPricingByModelAndAPIConfig(ctx context.Context, modelName string, apiConfigID uint) (*models.Pricing, error) {
	return s.pricingRepo.FindByModelAndAPIConfig(ctx, modelName, apiConfigID)
}

// CreatePricing creates a new pricing configuration
func (s *PricingService) CreatePricing(ctx context.Context, pricing *models.Pricing) error {
	// Check if pricing already exists
	existing, err := s.pricingRepo.FindByModelAndAPIConfig(ctx, pricing.ModelName, pricing.APIConfigID)
	if err != nil {
		return fmt.Errorf("failed to check existing pricing: %w", err)
	}
	if existing != nil {
		return ErrPricingExists
	}

	return s.pricingRepo.Create(ctx, pricing)
}

// UpdatePricing updates a pricing configuration
func (s *PricingService) UpdatePricing(ctx context.Context, id uint, updates map[string]interface{}) error {
	pricing, err := s.pricingRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if pricing == nil {
		return ErrPricingNotFound
	}

	// Apply updates
	if inputPrice, ok := updates["input_price"].(float64); ok {
		pricing.InputPrice = inputPrice
	}
	if outputPrice, ok := updates["output_price"].(float64); ok {
		pricing.OutputPrice = outputPrice
	}
	if currency, ok := updates["currency"].(string); ok {
		pricing.Currency = currency
	}
	if isActive, ok := updates["is_active"].(bool); ok {
		pricing.IsActive = isActive
	}
	if description, ok := updates["description"].(string); ok {
		pricing.Description = description
	}

	return s.pricingRepo.Update(ctx, pricing)
}

// DeletePricing deletes a pricing configuration
func (s *PricingService) DeletePricing(ctx context.Context, id uint) error {
	return s.pricingRepo.Delete(ctx, id)
}

// CalculateCost calculates the cost for given token usage
func (s *PricingService) CalculateCost(ctx context.Context, modelName string, apiConfigID uint, inputTokens, outputTokens int) (float64, error) {
	pricing, err := s.pricingRepo.FindByModelAndAPIConfig(ctx, modelName, apiConfigID)
	if err != nil {
		return 0, err
	}
	
	// If no pricing found, use default (free)
	if pricing == nil {
		return 0, nil
	}

	// Calculate cost based on pricing unit (usually per 1000 tokens)
	inputCost := (float64(inputTokens) / float64(pricing.Unit)) * pricing.InputPrice
	outputCost := (float64(outputTokens) / float64(pricing.Unit)) * pricing.OutputPrice
	
	return inputCost + outputCost, nil
}

// GetPricingsByAPIConfig retrieves all pricings for an API config
func (s *PricingService) GetPricingsByAPIConfig(ctx context.Context, apiConfigID uint) ([]*models.Pricing, error) {
	return s.pricingRepo.FindByAPIConfig(ctx, apiConfigID)
}

// InitializeDefaultPricings creates default pricing configurations for existing API configs
func (s *PricingService) InitializeDefaultPricings(ctx context.Context) error {
	// This method is now deprecated as pricing should be set per API config
	// Instead, we'll just return success
	return nil
}
