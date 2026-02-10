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

// AnthropicAdapter implements the Adapter interface for Anthropic API
type AnthropicAdapter struct {
	config *Config
}

// NewAnthropicAdapter creates a new Anthropic adapter
func NewAnthropicAdapter(config *Config) *AnthropicAdapter {
	if config.Client == nil {
		timeout := 30 * time.Second
		if config.Timeout > 0 {
			timeout = time.Duration(config.Timeout) * time.Second
		}
		config.Client = &http.Client{
			Timeout: timeout,
		}
	}
	return &AnthropicAdapter{
		config: config,
	}
}

// GetType returns the adapter type
func (a *AnthropicAdapter) GetType() string {
	return "anthropic"
}

// Anthropic request/response structures
type anthropicRequest struct {
	Model       string             `json:"model"`
	Messages    []anthropicMessage `json:"messages"`
	MaxTokens   int                `json:"max_tokens"`
	Temperature float64            `json:"temperature,omitempty"`
	TopP        float64            `json:"top_p,omitempty"`
	TopK        int                `json:"top_k,omitempty"`
	System      string             `json:"system,omitempty"`
	Tools       []anthropicTool    `json:"tools,omitempty"`
	Stream      bool               `json:"stream,omitempty"`
}

type anthropicMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"` // string or []anthropicContent
}

type anthropicTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

type anthropicResponse struct {
	ID         string              `json:"id"`
	Type       string              `json:"type"`
	Role       string              `json:"role"`
	Content    []anthropicContent  `json:"content"`
	Model      string              `json:"model"`
	StopReason string              `json:"stop_reason"`
	Usage      anthropicUsage      `json:"usage"`
}

type anthropicContent struct {
	Type  string                 `json:"type"` // text, tool_use, tool_result
	Text  string                 `json:"text,omitempty"`
	ID    string                 `json:"id,omitempty"`
	Name  string                 `json:"name,omitempty"`
	Input map[string]interface{} `json:"input,omitempty"`
}

type anthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// Call makes a request to Anthropic API
func (a *AnthropicAdapter) Call(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	// Convert unified request to Anthropic format
	messages, system := a.convertMessages(req.Messages)

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 1024 // Anthropic requires max_tokens
	}

	anthropicReq := &anthropicRequest{
		Model:       req.Model,
		Messages:    messages,
		MaxTokens:   maxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		TopK:        req.TopK,
		System:      system,
	}

	// Convert tools if present
	if len(req.Tools) > 0 {
		anthropicReq.Tools = a.convertTools(req.Tools)
	}

	// Marshal request
	reqBody, err := json.Marshal(anthropicReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := a.config.BaseURL + "/v1/messages"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", a.config.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

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
	var anthropicResp anthropicResponse
	if err := json.Unmarshal(respBody, &anthropicResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Convert to unified response
	return a.convertResponse(&anthropicResp), nil
}

// convertMessages converts OpenAI-style messages to Anthropic format
// Extracts system message separately as Anthropic uses a separate system field
func (a *AnthropicAdapter) convertMessages(messages []Message) ([]anthropicMessage, string) {
	var anthropicMessages []anthropicMessage
	var system string

	for _, msg := range messages {
		if msg.Role == "system" {
			// Anthropic uses a separate system field
			system = msg.Content
		} else if msg.Role == "tool" {
			// Tool result message
			anthropicMessages = append(anthropicMessages, anthropicMessage{
				Role: "user",
				Content: []anthropicContent{
					{
						Type: "tool_result",
						ID:   msg.ToolCallID,
						Text: msg.Content,
					},
				},
			})
		} else {
			// Check if message has tool calls
			if len(msg.ToolCalls) > 0 {
				// Convert tool calls to Anthropic format
				contents := make([]anthropicContent, 0, len(msg.ToolCalls)+1)
				
				// Add text content if present
				if msg.Content != "" {
					contents = append(contents, anthropicContent{
						Type: "text",
						Text: msg.Content,
					})
				}
				
				// Add tool use contents
				for _, tc := range msg.ToolCalls {
					var input map[string]interface{}
					json.Unmarshal([]byte(tc.Function.Arguments), &input)
					
					contents = append(contents, anthropicContent{
						Type:  "tool_use",
						ID:    tc.ID,
						Name:  tc.Function.Name,
						Input: input,
					})
				}
				
				anthropicMessages = append(anthropicMessages, anthropicMessage{
					Role:    msg.Role,
					Content: contents,
				})
			} else {
				// Regular text message
				anthropicMessages = append(anthropicMessages, anthropicMessage{
					Role:    msg.Role,
					Content: msg.Content,
				})
			}
		}
	}

	return anthropicMessages, system
}

// convertTools converts OpenAI-style tools to Anthropic format
func (a *AnthropicAdapter) convertTools(tools []Tool) []anthropicTool {
	anthropicTools := make([]anthropicTool, len(tools))
	for i, tool := range tools {
		anthropicTools[i] = anthropicTool{
			Name:        tool.Function.Name,
			Description: tool.Function.Description,
			InputSchema: tool.Function.Parameters,
		}
	}
	return anthropicTools
}

// convertResponse converts Anthropic response to unified format
func (a *AnthropicAdapter) convertResponse(resp *anthropicResponse) *ChatResponse {
	// Extract text and tool calls from content array
	var textContent string
	var toolCalls []ToolCall
	
	for _, content := range resp.Content {
		switch content.Type {
		case "text":
			textContent += content.Text
		case "tool_use":
			// Convert to OpenAI-style tool call
			argsJSON, _ := json.Marshal(content.Input)
			toolCalls = append(toolCalls, ToolCall{
				ID:   content.ID,
				Type: "function",
				Function: FunctionCall{
					Name:      content.Name,
					Arguments: string(argsJSON),
				},
			})
		}
	}

	msg := Message{
		Role:    resp.Role,
		Content: textContent,
	}
	
	if len(toolCalls) > 0 {
		msg.ToolCalls = toolCalls
	}

	return &ChatResponse{
		ID:    resp.ID,
		Model: resp.Model,
		Choices: []ChatChoice{
			{
				Index:        0,
				Message:      msg,
				FinishReason: resp.StopReason,
			},
		},
		Usage: UsageInfo{
			PromptTokens:     resp.Usage.InputTokens,
			CompletionTokens: resp.Usage.OutputTokens,
			TotalTokens:      resp.Usage.InputTokens + resp.Usage.OutputTokens,
		},
	}
}

// CallStream makes a streaming request to Anthropic API
func (a *AnthropicAdapter) CallStream(ctx context.Context, req *ChatRequest) (*http.Response, error) {
	// Convert unified request to Anthropic format
	messages, system := a.convertMessages(req.Messages)

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 1024 // Anthropic requires max_tokens
	}

	anthropicReq := &anthropicRequest{
		Model:       req.Model,
		Messages:    messages,
		MaxTokens:   maxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		TopK:        req.TopK,
		System:      system,
		Stream:      true,
	}

	// Convert tools if present
	if len(req.Tools) > 0 {
		anthropicReq.Tools = a.convertTools(req.Tools)
	}

	// Marshal request
	reqBody, err := json.Marshal(anthropicReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := a.config.BaseURL + "/v1/messages"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", a.config.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
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
