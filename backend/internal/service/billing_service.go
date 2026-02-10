package service

import (
	"api-aggregator/backend/internal/repository"
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
)

var (
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrInvalidAmount       = errors.New("invalid amount")
)

// BillingService handles all billing and quota operations with precision
// Uses micro-credits (1 credit = 1000 micro-credits) for 3 decimal precision
type BillingService struct {
	userRepo       *repository.UserRepository
	pricingService *PricingService
	mu             sync.RWMutex // Protects concurrent quota operations
}

func NewBillingService(userRepo *repository.UserRepository, pricingService *PricingService) *BillingService {
	return &BillingService{
		userRepo:       userRepo,
		pricingService: pricingService,
	}
}

// CostCalculation represents a detailed cost calculation
type CostCalculation struct {
	InputTokens       int     `json:"input_tokens"`
	OutputTokens      int     `json:"output_tokens"`
	TotalTokens       int     `json:"total_tokens"`
	InputCost         float64 `json:"input_cost"`          // In credits
	OutputCost        float64 `json:"output_cost"`         // In credits
	TotalCost         float64 `json:"total_cost"`          // In credits
	MicroCredits      int64   `json:"micro_credits"`       // In micro-credits (1/1000 credit)
	PricingConfigID   uint    `json:"pricing_config_id"`
	HasPricing        bool    `json:"has_pricing"`
	EstimatedCost     bool    `json:"estimated_cost"`      // True for streaming requests
}

// CalculateCost calculates the cost for a request with full precision
func (s *BillingService) CalculateCost(
	ctx context.Context,
	modelName string,
	apiConfigID uint,
	inputTokens int,
	outputTokens int,
	isEstimate bool,
) (*CostCalculation, error) {
	calc := &CostCalculation{
		InputTokens:   inputTokens,
		OutputTokens:  outputTokens,
		TotalTokens:   inputTokens + outputTokens,
		EstimatedCost: isEstimate,
	}

	// Get pricing configuration
	pricing, err := s.pricingService.GetPricingByModelAndAPIConfig(ctx, modelName, apiConfigID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pricing: %w", err)
	}

	if pricing != nil && pricing.IsActive {
		// Calculate cost based on pricing
		calc.HasPricing = true
		calc.PricingConfigID = pricing.ID
		
		// Calculate input cost
		calc.InputCost = (float64(inputTokens) / float64(pricing.Unit)) * pricing.InputPrice
		
		// Calculate output cost
		calc.OutputCost = (float64(outputTokens) / float64(pricing.Unit)) * pricing.OutputPrice
		
		// Total cost in credits
		calc.TotalCost = calc.InputCost + calc.OutputCost
	} else {
		// No pricing configured, use token count as cost (1 token = 1 credit)
		calc.HasPricing = false
		calc.InputCost = float64(inputTokens)
		calc.OutputCost = float64(outputTokens)
		calc.TotalCost = float64(calc.TotalTokens)
	}

	// Convert to micro-credits (1 credit = 1000 micro-credits)
	// This preserves 3 decimal places of precision
	calc.MicroCredits = int64(math.Round(calc.TotalCost * 1000))

	return calc, nil
}

// CheckBalance checks if user has sufficient balance
func (s *BillingService) CheckBalance(ctx context.Context, userID uint, microCredits int64) (bool, error) {
	if microCredits < 0 {
		return false, ErrInvalidAmount
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return false, fmt.Errorf("user not found")
	}

	// Convert to micro-credits for comparison
	availableMicroCredits := (user.Quota - user.UsedQuota) * 1000
	
	return availableMicroCredits >= microCredits, nil
}

// DeductBalance deducts balance from user account with transaction safety
func (s *BillingService) DeductBalance(ctx context.Context, userID uint, microCredits int64) error {
	if microCredits < 0 {
		return ErrInvalidAmount
	}
	if microCredits == 0 {
		return nil // Nothing to deduct
	}

	// Lock to prevent concurrent deductions
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get user with fresh data
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	// Convert micro-credits to credits (round up to ensure we never under-charge)
	creditsToDeduct := int64(math.Ceil(float64(microCredits) / 1000.0))

	// Check balance
	availableCredits := user.Quota - user.UsedQuota
	if availableCredits < creditsToDeduct {
		return ErrInsufficientBalance
	}

	// Deduct quota
	user.UsedQuota += creditsToDeduct

	// Update user
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user balance: %w", err)
	}

	return nil
}

// RefundBalance refunds balance to user account (for failed requests)
func (s *BillingService) RefundBalance(ctx context.Context, userID uint, microCredits int64) error {
	if microCredits < 0 {
		return ErrInvalidAmount
	}
	if microCredits == 0 {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	// Convert micro-credits to credits (round down for refunds)
	creditsToRefund := int64(math.Floor(float64(microCredits) / 1000.0))

	// Refund quota (ensure UsedQuota doesn't go negative)
	if user.UsedQuota >= creditsToRefund {
		user.UsedQuota -= creditsToRefund
	} else {
		user.UsedQuota = 0
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to refund user balance: %w", err)
	}

	return nil
}

// GetBalance returns user's current balance in credits
func (s *BillingService) GetBalance(ctx context.Context, userID uint) (totalCredits, usedCredits, availableCredits int64, err error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return 0, 0, 0, fmt.Errorf("user not found")
	}

	return user.Quota, user.UsedQuota, user.Quota - user.UsedQuota, nil
}

// AddBalance adds balance to user account (for admin operations or purchases)
func (s *BillingService) AddBalance(ctx context.Context, userID uint, credits int64) error {
	if credits < 0 {
		return ErrInvalidAmount
	}
	if credits == 0 {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	user.Quota += credits

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to add balance: %w", err)
	}

	return nil
}

// EstimateInputTokens estimates input tokens from request
func (s *BillingService) EstimateInputTokens(messages []struct{ Content string }, tools []interface{}) int {
	var totalChars int
	
	// Count all message content
	for _, msg := range messages {
		totalChars += len(msg.Content)
		totalChars += 10 // Overhead for role and structure
	}
	
	// Add overhead for tools if present
	if len(tools) > 0 {
		totalChars += len(tools) * 150 // Rough estimate per tool
	}
	
	// Conservative estimate: 1 token per 3 characters
	estimatedTokens := totalChars / 3
	if estimatedTokens < 10 {
		estimatedTokens = 10
	}
	
	return estimatedTokens
}

// EstimateOutputTokens estimates output tokens from request
func (s *BillingService) EstimateOutputTokens(maxTokens int, inputTokens int) int {
	if maxTokens > 0 {
		return maxTokens
	}
	
	// Default estimate: 50% of input, min 100, max 2000
	estimated := inputTokens / 2
	if estimated < 100 {
		estimated = 100
	}
	if estimated > 2000 {
		estimated = 2000
	}
	
	return estimated
}

// FormatCredits formats micro-credits as human-readable credits string
func FormatCredits(microCredits int64) string {
	credits := float64(microCredits) / 1000.0
	return fmt.Sprintf("%.3f", credits)
}

// ParseCredits parses credits string to micro-credits
func ParseCredits(creditsStr string) (int64, error) {
	var credits float64
	_, err := fmt.Sscanf(creditsStr, "%f", &credits)
	if err != nil {
		return 0, err
	}
	return int64(math.Round(credits * 1000)), nil
}
