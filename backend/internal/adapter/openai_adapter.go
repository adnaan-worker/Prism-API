package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OpenAIAdapter implements the Adapter interface for OpenAI API
type OpenAIAdapter struct {
	config *Config
}

// NewOpenAIAdapter creates a new OpenAI adapter
func NewOpenAIAdapter(config *Config) *OpenAIAdapter {
	if config.Client == nil {
		timeout := 30 * time.Second
		if config.Timeout > 0 {
			timeout = time.Duration(config.Timeout) * time.Second
		}
		config.Client = &http.Client{
			Timeout: timeout,
		}
	}
	return &OpenAIAdapter{
		config: config,
	}
}

// GetType returns the adapter type
func (a *OpenAIAdapter) GetType() string {
	return "openai"
}

// OpenAI request/response structures
type openAIRequest struct {
	Model              string                 `json:"model"`
	Messages           []Message              `json:"messages"`
	Temperature        float64                `json:"temperature,omitempty"`
	TopP               float64                `json:"top_p,omitempty"`
	MaxTokens          int                    `json:"max_tokens,omitempty"`
	Stream             bool                   `json:"stream,omitempty"`
	Stop               interface{}            `json:"stop,omitempty"` // string or []string
	N                  int                    `json:"n,omitempty"`
	PresencePenalty    float64                `json:"presence_penalty,omitempty"`
	FrequencyPenalty   float64                `json:"frequency_penalty,omitempty"`
	Tools              []Tool                 `json:"tools,omitempty"`
	ToolChoice         interface{}            `json:"tool_choice,omitempty"`
	User               string                 `json:"user,omitempty"`
	Seed               *int                   `json:"seed,omitempty"`
	LogitBias          map[string]int         `json:"logit_bias,omitempty"`
	Logprobs           bool                   `json:"logprobs,omitempty"`
	TopLogprobs        int                    `json:"top_logprobs,omitempty"`
	ResponseFormat     *ResponseFormat        `json:"response_format,omitempty"`
	ServiceTier        string                 `json:"service_tier,omitempty"`
	ParallelToolCalls  *bool                  `json:"parallel_tool_calls,omitempty"`
	StreamOptions      *StreamOptions         `json:"stream_options,omitempty"`
}

type openAIResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []openAIChoice `json:"choices"`
	Usage   openAIUsage    `json:"usage"`
}

type openAIChoice struct {
	Index        int            `json:"index"`
	Message      openAIMessage  `json:"message"`
	FinishReason string         `json:"finish_reason"`
}

type openAIMessage struct {
	Role      string     `json:"role"`
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

type openAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Call makes a request to OpenAI API
func (a *OpenAIAdapter) Call(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	// Convert unified request to OpenAI format
	openAIReq := &openAIRequest{
		Model:             req.Model,
		Messages:          req.Messages,
		Temperature:       req.Temperature,
		TopP:              req.TopP,
		MaxTokens:         req.MaxTokens,
		Stream:            req.Stream,
		Stop:              req.Stop,
		N:                 req.N,
		PresencePenalty:   req.PresencePenalty,
		FrequencyPenalty:  req.FrequencyPenalty,
		Tools:             req.Tools,
		ToolChoice:        req.ToolChoice,
		User:              req.User,
		Seed:              req.Seed,
		LogitBias:         req.LogitBias,
		Logprobs:          req.Logprobs,
		TopLogprobs:       req.TopLogprobs,
		ResponseFormat:    req.ResponseFormat,
		ServiceTier:       req.ServiceTier,
		ParallelToolCalls: req.ParallelToolCalls,
		StreamOptions:     req.StreamOptions,
	}

	// Marshal request
	reqBody, err := json.Marshal(openAIReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	// Handle base URLs that already include /v1
	baseURL := a.config.BaseURL
	if strings.HasSuffix(baseURL, "/v1") {
		baseURL = strings.TrimSuffix(baseURL, "/v1")
	}
	url := baseURL + "/v1/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+a.config.APIKey)

	// Make request
	resp, err := a.config.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Unmarshal response
	var openAIResp openAIResponse
	if err := json.Unmarshal(respBody, &openAIResp); err != nil {
		// Show first 200 chars of response for debugging
		preview := string(respBody)
		if len(preview) > 200 {
			preview = preview[:200] + "..."
		}
		return nil, fmt.Errorf("failed to unmarshal response: %w (response preview: %s)", err, preview)
	}

	// Convert to unified response
	return a.convertResponse(&openAIResp), nil
}

// convertResponse converts OpenAI response to unified format
func (a *OpenAIAdapter) convertResponse(resp *openAIResponse) *ChatResponse {
	choices := make([]ChatChoice, len(resp.Choices))
	for i, choice := range resp.Choices {
		msg := Message{
			Role:    choice.Message.Role,
			Content: choice.Message.Content,
		}
		
		// Convert tool calls if present
		if len(choice.Message.ToolCalls) > 0 {
			msg.ToolCalls = choice.Message.ToolCalls
		}
		
		choices[i] = ChatChoice{
			Index:        choice.Index,
			Message:      msg,
			FinishReason: choice.FinishReason,
		}
	}

	return &ChatResponse{
		ID:      resp.ID,
		Object:  resp.Object,  // Preserve OpenAI's "chat.completion"
		Created: resp.Created, // Preserve Unix timestamp
		Model:   resp.Model,
		Choices: choices,
		Usage: UsageInfo{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}
}

// CallStream makes a streaming request to OpenAI API
func (a *OpenAIAdapter) CallStream(ctx context.Context, req *ChatRequest) (*http.Response, error) {
	// Convert unified request to OpenAI format with stream enabled
	openAIReq := &openAIRequest{
		Model:             req.Model,
		Messages:          req.Messages,
		Temperature:       req.Temperature,
		TopP:              req.TopP,
		MaxTokens:         req.MaxTokens,
		Stream:            true,
		Stop:              req.Stop,
		N:                 req.N,
		PresencePenalty:   req.PresencePenalty,
		FrequencyPenalty:  req.FrequencyPenalty,
		Tools:             req.Tools,
		ToolChoice:        req.ToolChoice,
		User:              req.User,
		Seed:              req.Seed,
		LogitBias:         req.LogitBias,
		Logprobs:          req.Logprobs,
		TopLogprobs:       req.TopLogprobs,
		ResponseFormat:    req.ResponseFormat,
		ServiceTier:       req.ServiceTier,
		ParallelToolCalls: req.ParallelToolCalls,
		StreamOptions:     req.StreamOptions,
	}

	// Marshal request
	reqBody, err := json.Marshal(openAIReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	// Handle base URLs that already include /v1
	baseURL := a.config.BaseURL
	if strings.HasSuffix(baseURL, "/v1") {
		baseURL = strings.TrimSuffix(baseURL, "/v1")
	}
	url := baseURL + "/v1/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+a.config.APIKey)
	httpReq.Header.Set("Accept", "text/event-stream")

	// Make request and return response directly
	resp, err := a.config.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Return response for streaming (caller must close body)
	return resp, nil
}
