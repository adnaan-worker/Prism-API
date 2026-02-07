package adapter

import (
	"context"
	"net/http"
)

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest represents a unified chat completion request
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

// ChatResponse represents a unified chat completion response
type ChatResponse struct {
	ID      string       `json:"id"`
	Model   string       `json:"model"`
	Choices []ChatChoice `json:"choices"`
	Usage   UsageInfo    `json:"usage"`
}

// ChatChoice represents a single choice in the response
type ChatChoice struct {
	Index   int     `json:"index"`
	Message Message `json:"message"`
}

// UsageInfo represents token usage information
type UsageInfo struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Adapter is the interface that all API adapters must implement
type Adapter interface {
	// Call makes a request to the API and returns a unified response
	Call(ctx context.Context, req *ChatRequest) (*ChatResponse, error)

	// CallStream makes a streaming request to the API and returns the raw HTTP response
	// The caller is responsible for reading and closing the response body
	CallStream(ctx context.Context, req *ChatRequest) (*http.Response, error)

	// GetType returns the adapter type (openai, anthropic, gemini, custom)
	GetType() string
}

// Config represents the configuration for an adapter
type Config struct {
	BaseURL string
	APIKey  string
	Model   string
	Timeout int
	Client  *http.Client
}
