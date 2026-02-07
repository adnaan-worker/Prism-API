package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// GeminiAdapter implements the Adapter interface for Google Gemini API
type GeminiAdapter struct {
	config *Config
}

// NewGeminiAdapter creates a new Gemini adapter
func NewGeminiAdapter(config *Config) *GeminiAdapter {
	if config.Client == nil {
		timeout := 30 * time.Second
		if config.Timeout > 0 {
			timeout = time.Duration(config.Timeout) * time.Second
		}
		config.Client = &http.Client{
			Timeout: timeout,
		}
	}
	return &GeminiAdapter{
		config: config,
	}
}

// GetType returns the adapter type
func (a *GeminiAdapter) GetType() string {
	return "gemini"
}

// Gemini request/response structures
type geminiRequest struct {
	Contents         []geminiContent         `json:"contents"`
	GenerationConfig *geminiGenerationConfig `json:"generationConfig,omitempty"`
}

type geminiContent struct {
	Role  string       `json:"role"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenerationConfig struct {
	Temperature     float64 `json:"temperature,omitempty"`
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
}

type geminiResponse struct {
	Candidates    []geminiCandidate `json:"candidates"`
	UsageMetadata geminiUsage       `json:"usageMetadata"`
}

type geminiCandidate struct {
	Content      geminiContent `json:"content"`
	FinishReason string        `json:"finishReason"`
	Index        int           `json:"index"`
}

type geminiUsage struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

// Call makes a request to Gemini API
func (a *GeminiAdapter) Call(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	// Convert unified request to Gemini format
	contents := a.convertMessages(req.Messages)

	geminiReq := &geminiRequest{
		Contents: contents,
	}

	// Add generation config if specified
	if req.Temperature > 0 || req.MaxTokens > 0 {
		geminiReq.GenerationConfig = &geminiGenerationConfig{
			Temperature:     req.Temperature,
			MaxOutputTokens: req.MaxTokens,
		}
	}

	// Marshal request
	reqBody, err := json.Marshal(geminiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	// Gemini uses model in URL path
	url := fmt.Sprintf("%s/v1/models/%s:generateContent?key=%s",
		a.config.BaseURL, req.Model, a.config.APIKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")

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
	var geminiResp geminiResponse
	if err := json.Unmarshal(respBody, &geminiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Convert to unified response
	return a.convertResponse(&geminiResp, req.Model), nil
}

// convertMessages converts OpenAI-style messages to Gemini format
func (a *GeminiAdapter) convertMessages(messages []Message) []geminiContent {
	var contents []geminiContent

	for _, msg := range messages {
		// Gemini uses "user" and "model" roles
		role := msg.Role
		if role == "assistant" {
			role = "model"
		}
		// Skip system messages or prepend to first user message
		if role == "system" {
			continue
		}

		contents = append(contents, geminiContent{
			Role: role,
			Parts: []geminiPart{
				{Text: msg.Content},
			},
		})
	}

	return contents
}

// convertResponse converts Gemini response to unified format
func (a *GeminiAdapter) convertResponse(resp *geminiResponse, model string) *ChatResponse {
	choices := make([]ChatChoice, len(resp.Candidates))

	for i, candidate := range resp.Candidates {
		var content string
		if len(candidate.Content.Parts) > 0 {
			content = candidate.Content.Parts[0].Text
		}

		// Convert "model" role back to "assistant"
		role := candidate.Content.Role
		if role == "model" {
			role = "assistant"
		}

		choices[i] = ChatChoice{
			Index: candidate.Index,
			Message: Message{
				Role:    role,
				Content: content,
			},
		}
	}

	return &ChatResponse{
		ID:      fmt.Sprintf("gemini-%d", time.Now().Unix()),
		Model:   model,
		Choices: choices,
		Usage: UsageInfo{
			PromptTokens:     resp.UsageMetadata.PromptTokenCount,
			CompletionTokens: resp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      resp.UsageMetadata.TotalTokenCount,
		},
	}
}

// CallStream makes a streaming request to Gemini API
func (a *GeminiAdapter) CallStream(ctx context.Context, req *ChatRequest) (*http.Response, error) {
	// Convert unified request to Gemini format
	contents := a.convertMessages(req.Messages)

	geminiReq := &geminiRequest{
		Contents: contents,
	}

	// Add generation config if specified
	if req.Temperature > 0 || req.MaxTokens > 0 {
		geminiReq.GenerationConfig = &geminiGenerationConfig{
			Temperature:     req.Temperature,
			MaxOutputTokens: req.MaxTokens,
		}
	}

	// Marshal request
	reqBody, err := json.Marshal(geminiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	// Gemini uses streamGenerateContent for streaming
	url := fmt.Sprintf("%s/v1/models/%s:streamGenerateContent?key=%s&alt=sse",
		a.config.BaseURL, req.Model, a.config.APIKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")

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
