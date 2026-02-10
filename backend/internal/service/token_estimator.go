package service

import "api-aggregator/backend/internal/adapter"

// TokenEstimator estimates token usage for requests
type TokenEstimator struct{}

// NewTokenEstimator creates a new token estimator
func NewTokenEstimator() *TokenEstimator {
	return &TokenEstimator{}
}

// EstimateTotal estimates total tokens (input + output) for a request
func (e *TokenEstimator) EstimateTotal(req *adapter.ChatRequest) int64 {
	input := e.EstimateInput(req)
	output := e.EstimateOutput(req, input)
	return input + output
}

// EstimateInput estimates input tokens from request messages and tools
func (e *TokenEstimator) EstimateInput(req *adapter.ChatRequest) int64 {
	var totalChars int
	
	// Count all message content
	for _, msg := range req.Messages {
		totalChars += len(msg.Content)
		totalChars += 10 // Overhead for role and structure
	}
	
	// Add overhead for tools if present
	if len(req.Tools) > 0 {
		totalChars += len(req.Tools) * 150 // Rough estimate per tool
	}
	
	// Conservative estimate: 1 token per 3 characters
	estimated := int64(totalChars / 3)
	if estimated < 10 {
		estimated = 10
	}
	
	return estimated
}

// EstimateOutput estimates output tokens based on max_tokens or input size
func (e *TokenEstimator) EstimateOutput(req *adapter.ChatRequest, inputTokens int64) int64 {
	// If max_tokens is specified, use it
	if req.MaxTokens > 0 {
		return int64(req.MaxTokens)
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
