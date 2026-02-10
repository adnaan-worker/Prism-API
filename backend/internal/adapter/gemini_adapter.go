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
	Tools            []geminiToolConfig      `json:"tools,omitempty"`
}

type geminiContent struct {
	Role  string       `json:"role"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text             string                 `json:"text,omitempty"`
	FunctionCall     *geminiFunctionCall    `json:"functionCall,omitempty"`
	FunctionResponse *geminiFunctionResponse `json:"functionResponse,omitempty"`
}

type geminiFunctionCall struct {
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args"`
}

type geminiFunctionResponse struct {
	Name     string                 `json:"name"`
	Response map[string]interface{} `json:"response"`
}

type geminiToolConfig struct {
	FunctionDeclarations []geminiFunctionDeclaration `json:"functionDeclarations"`
}

type geminiFunctionDeclaration struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type geminiGenerationConfig struct {
	Temperature     float64  `json:"temperature,omitempty"`
	TopP            float64  `json:"topP,omitempty"`
	TopK            int      `json:"topK,omitempty"`
	MaxOutputTokens int      `json:"maxOutputTokens,omitempty"`
	StopSequences   []string `json:"stopSequences,omitempty"`
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
	if req.Temperature > 0 || req.TopP > 0 || req.TopK > 0 || req.MaxTokens > 0 {
		geminiReq.GenerationConfig = &geminiGenerationConfig{
			Temperature:     req.Temperature,
			TopP:            req.TopP,
			TopK:            req.TopK,
			MaxOutputTokens: req.MaxTokens,
		}
		
		// Convert stop sequences
		if req.Stop != nil {
			switch v := req.Stop.(type) {
			case string:
				geminiReq.GenerationConfig.StopSequences = []string{v}
			case []string:
				geminiReq.GenerationConfig.StopSequences = v
			case []interface{}:
				stops := make([]string, len(v))
				for i, s := range v {
					if str, ok := s.(string); ok {
						stops[i] = str
					}
				}
				geminiReq.GenerationConfig.StopSequences = stops
			}
		}
	}

	// Convert tools if present
	if len(req.Tools) > 0 {
		geminiReq.Tools = a.convertTools(req.Tools)
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

		parts := []geminiPart{}

		// Add text content if present
		if msg.Content != "" {
			parts = append(parts, geminiPart{Text: msg.Content})
		}

		// Add tool calls if present
		if len(msg.ToolCalls) > 0 {
			for _, tc := range msg.ToolCalls {
				var args map[string]interface{}
				json.Unmarshal([]byte(tc.Function.Arguments), &args)
				
				parts = append(parts, geminiPart{
					FunctionCall: &geminiFunctionCall{
						Name: tc.Function.Name,
						Args: args,
					},
				})
			}
		}

		// Add tool result if present
		if msg.ToolCallID != "" {
			// This is a tool result message
			var response map[string]interface{}
			json.Unmarshal([]byte(msg.Content), &response)
			if response == nil {
				response = map[string]interface{}{"result": msg.Content}
			}
			
			parts = append(parts, geminiPart{
				FunctionResponse: &geminiFunctionResponse{
					Name:     msg.Name, // Tool name should be in Name field
					Response: response,
				},
			})
		}

		if len(parts) > 0 {
			contents = append(contents, geminiContent{
				Role:  role,
				Parts: parts,
			})
		}
	}

	return contents
}

// convertTools converts OpenAI-style tools to Gemini format
func (a *GeminiAdapter) convertTools(tools []Tool) []geminiToolConfig {
	declarations := make([]geminiFunctionDeclaration, len(tools))
	for i, tool := range tools {
		declarations[i] = geminiFunctionDeclaration{
			Name:        tool.Function.Name,
			Description: tool.Function.Description,
			Parameters:  tool.Function.Parameters,
		}
	}
	
	return []geminiToolConfig{
		{FunctionDeclarations: declarations},
	}
}

// convertResponse converts Gemini response to unified format
func (a *GeminiAdapter) convertResponse(resp *geminiResponse, model string) *ChatResponse {
	choices := make([]ChatChoice, len(resp.Candidates))

	for i, candidate := range resp.Candidates {
		var textContent string
		var toolCalls []ToolCall
		
		// Extract text and function calls from parts
		for _, part := range candidate.Content.Parts {
			if part.Text != "" {
				textContent += part.Text
			}
			
			if part.FunctionCall != nil {
				// Convert to OpenAI-style tool call
				argsJSON, _ := json.Marshal(part.FunctionCall.Args)
				toolCalls = append(toolCalls, ToolCall{
					ID:   fmt.Sprintf("call_%d", time.Now().UnixNano()),
					Type: "function",
					Function: FunctionCall{
						Name:      part.FunctionCall.Name,
						Arguments: string(argsJSON),
					},
				})
			}
		}

		// Convert "model" role back to "assistant"
		role := candidate.Content.Role
		if role == "model" {
			role = "assistant"
		}

		msg := Message{
			Role:    role,
			Content: textContent,
		}
		
		if len(toolCalls) > 0 {
			msg.ToolCalls = toolCalls
		}

		choices[i] = ChatChoice{
			Index:        candidate.Index,
			Message:      msg,
			FinishReason: convertGeminiFinishReason(candidate.FinishReason),
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

// convertGeminiFinishReason converts Gemini finish reason to OpenAI format
func convertGeminiFinishReason(reason string) string {
	switch reason {
	case "STOP":
		return "stop"
	case "MAX_TOKENS":
		return "length"
	case "SAFETY", "RECITATION":
		return "content_filter"
	default:
		return "stop"
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
	if req.Temperature > 0 || req.TopP > 0 || req.TopK > 0 || req.MaxTokens > 0 {
		geminiReq.GenerationConfig = &geminiGenerationConfig{
			Temperature:     req.Temperature,
			TopP:            req.TopP,
			TopK:            req.TopK,
			MaxOutputTokens: req.MaxTokens,
		}
		
		// Convert stop sequences
		if req.Stop != nil {
			switch v := req.Stop.(type) {
			case string:
				geminiReq.GenerationConfig.StopSequences = []string{v}
			case []string:
				geminiReq.GenerationConfig.StopSequences = v
			case []interface{}:
				stops := make([]string, len(v))
				for i, s := range v {
					if str, ok := s.(string); ok {
						stops[i] = str
					}
				}
				geminiReq.GenerationConfig.StopSequences = stops
			}
		}
	}

	// Convert tools if present
	if len(req.Tools) > 0 {
		geminiReq.Tools = a.convertTools(req.Tools)
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
