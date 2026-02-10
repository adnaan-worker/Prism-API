package adapter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Test OpenAI adapter request/response conversion
func TestOpenAIAdapter_RequestResponseConversion(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request format
		var req openAIRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		// Verify OpenAI format
		if req.Model != "gpt-4" {
			t.Errorf("Expected model gpt-4, got %s", req.Model)
		}
		if len(req.Messages) != 2 {
			t.Errorf("Expected 2 messages, got %d", len(req.Messages))
		}

		// Return mock response
		resp := openAIResponse{
			ID:      "chatcmpl-123",
			Object:  "chat.completion",
			Created: 1234567890,
			Model:   "gpt-4",
			Choices: []openAIChoice{
				{
					Index: 0,
					Message: openAIMessage{
						Role:    "assistant",
						Content: "Hello! How can I help you?",
					},
					FinishReason: "stop",
				},
			},
			Usage: openAIUsage{
				PromptTokens:     10,
				CompletionTokens: 20,
				TotalTokens:      30,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create adapter
	config := &Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
		Timeout: 30,
	}
	adapter := NewOpenAIAdapter(config)

	// Make request
	req := &ChatRequest{
		Model: "gpt-4",
		Messages: []Message{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "user", Content: "Hello"},
		},
		Temperature: 0.7,
		MaxTokens:   100,
	}

	resp, err := adapter.Call(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify response
	if resp.ID != "chatcmpl-123" {
		t.Errorf("Expected ID chatcmpl-123, got %s", resp.ID)
	}
	if resp.Model != "gpt-4" {
		t.Errorf("Expected model gpt-4, got %s", resp.Model)
	}
	if len(resp.Choices) != 1 {
		t.Errorf("Expected 1 choice, got %d", len(resp.Choices))
	}
	if resp.Choices[0].Message.Content != "Hello! How can I help you?" {
		t.Errorf("Unexpected message content: %s", resp.Choices[0].Message.Content)
	}
	if resp.Usage.TotalTokens != 30 {
		t.Errorf("Expected 30 total tokens, got %d", resp.Usage.TotalTokens)
	}
}

// Test Anthropic adapter request/response conversion
func TestAnthropicAdapter_RequestResponseConversion(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request format
		var req anthropicRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		// Verify Anthropic format
		if req.Model != "claude-3-opus" {
			t.Errorf("Expected model claude-3-opus, got %s", req.Model)
		}
		if req.System != "You are a helpful assistant." {
			t.Errorf("Expected system message, got %s", req.System)
		}
		if len(req.Messages) != 1 {
			t.Errorf("Expected 1 message (system extracted), got %d", len(req.Messages))
		}

		// Return mock response
		resp := anthropicResponse{
			ID:   "msg-123",
			Type: "message",
			Role: "assistant",
			Content: []anthropicContent{
				{Type: "text", Text: "Hello! How can I help you?"},
			},
			Model:      "claude-3-opus",
			StopReason: "end_turn",
			Usage: anthropicUsage{
				InputTokens:  10,
				OutputTokens: 20,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create adapter
	config := &Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
		Timeout: 30,
	}
	adapter := NewAnthropicAdapter(config)

	// Make request
	req := &ChatRequest{
		Model: "claude-3-opus",
		Messages: []Message{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "user", Content: "Hello"},
		},
		Temperature: 0.7,
		MaxTokens:   100,
	}

	resp, err := adapter.Call(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify response
	if resp.ID != "msg-123" {
		t.Errorf("Expected ID msg-123, got %s", resp.ID)
	}
	if resp.Model != "claude-3-opus" {
		t.Errorf("Expected model claude-3-opus, got %s", resp.Model)
	}
	if len(resp.Choices) != 1 {
		t.Errorf("Expected 1 choice, got %d", len(resp.Choices))
	}
	if resp.Choices[0].Message.Content != "Hello! How can I help you?" {
		t.Errorf("Unexpected message content: %s", resp.Choices[0].Message.Content)
	}
	if resp.Usage.TotalTokens != 30 {
		t.Errorf("Expected 30 total tokens, got %d", resp.Usage.TotalTokens)
	}
}

// Test Gemini adapter request/response conversion
func TestGeminiAdapter_RequestResponseConversion(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request format
		var req geminiRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		// Verify Gemini format
		if len(req.Contents) != 1 {
			t.Errorf("Expected 1 content (system skipped), got %d", len(req.Contents))
		}
		if req.Contents[0].Role != "user" {
			t.Errorf("Expected role user, got %s", req.Contents[0].Role)
		}

		// Return mock response
		resp := geminiResponse{
			Candidates: []geminiCandidate{
				{
					Content: geminiContent{
						Role: "model",
						Parts: []geminiPart{
							{Text: "Hello! How can I help you?"},
						},
					},
					FinishReason: "STOP",
					Index:        0,
				},
			},
			UsageMetadata: geminiUsage{
				PromptTokenCount:     10,
				CandidatesTokenCount: 20,
				TotalTokenCount:      30,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create adapter
	config := &Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
		Timeout: 30,
	}
	adapter := NewGeminiAdapter(config)

	// Make request
	req := &ChatRequest{
		Model: "gemini-pro",
		Messages: []Message{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "user", Content: "Hello"},
		},
		Temperature: 0.7,
		MaxTokens:   100,
	}

	resp, err := adapter.Call(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify response
	if resp.Model != "gemini-pro" {
		t.Errorf("Expected model gemini-pro, got %s", resp.Model)
	}
	if len(resp.Choices) != 1 {
		t.Errorf("Expected 1 choice, got %d", len(resp.Choices))
	}
	if resp.Choices[0].Message.Role != "assistant" {
		t.Errorf("Expected role assistant, got %s", resp.Choices[0].Message.Role)
	}
	if resp.Choices[0].Message.Content != "Hello! How can I help you?" {
		t.Errorf("Unexpected message content: %s", resp.Choices[0].Message.Content)
	}
	if resp.Usage.TotalTokens != 30 {
		t.Errorf("Expected 30 total tokens, got %d", resp.Usage.TotalTokens)
	}
}

// Test message conversion for Anthropic (system message extraction)
func TestAnthropicAdapter_MessageConversion(t *testing.T) {
	adapter := &AnthropicAdapter{}

	messages := []Message{
		{Role: "system", Content: "You are a helpful assistant."},
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there!"},
		{Role: "user", Content: "How are you?"},
	}

	convertedMessages, system := adapter.convertMessages(messages)

	// Verify system message extracted
	if system != "You are a helpful assistant." {
		t.Errorf("Expected system message extracted, got %s", system)
	}

	// Verify other messages converted
	if len(convertedMessages) != 3 {
		t.Errorf("Expected 3 messages, got %d", len(convertedMessages))
	}

	if convertedMessages[0].Role != "user" || convertedMessages[0].Content != "Hello" {
		t.Error("First message not converted correctly")
	}
}

// Test message conversion for Gemini (role mapping)
func TestGeminiAdapter_MessageConversion(t *testing.T) {
	adapter := &GeminiAdapter{}

	messages := []Message{
		{Role: "system", Content: "You are a helpful assistant."},
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there!"},
		{Role: "user", Content: "How are you?"},
	}

	convertedMessages := adapter.convertMessages(messages)

	// Verify system message skipped
	if len(convertedMessages) != 3 {
		t.Errorf("Expected 3 messages (system skipped), got %d", len(convertedMessages))
	}

	// Verify role mapping (assistant -> model)
	if convertedMessages[1].Role != "model" {
		t.Errorf("Expected role 'model', got %s", convertedMessages[1].Role)
	}
}

// Test adapter factory
func TestFactory_CreateAdapter(t *testing.T) {
	factory := NewFactory()

	tests := []struct {
		adapterType string
		expectError bool
	}{
		{"openai", false},
		{"anthropic", false},
		{"gemini", false},
		{"custom", false},
		{"invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.adapterType, func(t *testing.T) {
			adapter, err := factory.CreateAdapterByType(tt.adapterType, "https://api.test.com", "test-key", 30)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if adapter == nil {
					t.Error("Expected adapter, got nil")
				}
			}
		})
	}
}

// Test adapter type methods
func TestAdapter_GetType(t *testing.T) {
	config := &Config{
		BaseURL: "https://api.test.com",
		APIKey:  "test-key",
		Timeout: 30,
	}

	tests := []struct {
		name         string
		adapter      Adapter
		expectedType string
	}{
		{"OpenAI", NewOpenAIAdapter(config), "openai"},
		{"Anthropic", NewAnthropicAdapter(config), "anthropic"},
		{"Gemini", NewGeminiAdapter(config), "gemini"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.adapter.GetType() != tt.expectedType {
				t.Errorf("Expected type %s, got %s", tt.expectedType, tt.adapter.GetType())
			}
		})
	}
}
