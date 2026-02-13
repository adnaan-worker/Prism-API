package adapter

import (
	"context"
	"net/http"
)

// Message represents a chat message
type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	Name       string     `json:"name,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

// ToolCall represents a tool/function call
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"` // function
	Function FunctionCall `json:"function"`
}

// FunctionCall represents a function call
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // JSON string
}

// Tool represents a tool definition
type Tool struct {
	Type     string       `json:"type"` // function
	Function ToolFunction `json:"function"`
}

// ToolFunction represents a function definition
type ToolFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// ChatRequest represents a unified chat completion request
type ChatRequest struct {
	// 基础参数
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`

	// 采样参数
	Temperature float64 `json:"temperature,omitempty"`
	TopP        float64 `json:"top_p,omitempty"`
	TopK        int     `json:"top_k,omitempty"`

	// 输出控制
	MaxTokens int         `json:"max_tokens,omitempty"`
	Stop      interface{} `json:"stop,omitempty"` // string or []string
	N         int         `json:"n,omitempty"`

	// 惩罚参数
	PresencePenalty  float64 `json:"presence_penalty,omitempty"`
	FrequencyPenalty float64 `json:"frequency_penalty,omitempty"`

	// 工具调用
	Tools      []Tool      `json:"tools,omitempty"`
	ToolChoice interface{} `json:"tool_choice,omitempty"`

	// 流式输出
	Stream bool `json:"stream,omitempty"`

	// OpenAI 特有参数
	User              string         `json:"user,omitempty"`
	Seed              *int           `json:"seed,omitempty"`
	LogitBias         map[string]int `json:"logit_bias,omitempty"`
	Logprobs          bool           `json:"logprobs,omitempty"`
	TopLogprobs       int            `json:"top_logprobs,omitempty"`
	ResponseFormat    *ResponseFormat `json:"response_format,omitempty"`
	ServiceTier       string         `json:"service_tier,omitempty"`
	ParallelToolCalls *bool          `json:"parallel_tool_calls,omitempty"`
	StreamOptions     *StreamOptions `json:"stream_options,omitempty"`

	// Anthropic 特有参数
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	StopSequences []string               `json:"stop_sequences,omitempty"`

	// Gemini 特有参数
	SafetySettings []SafetySetting `json:"safety_settings,omitempty"`
	CachedContent  string          `json:"cached_content,omitempty"`
}

// ResponseFormat 响应格式配置
type ResponseFormat struct {
	Type       string      `json:"type"` // "text" or "json_object" or "json_schema"
	JSONSchema interface{} `json:"json_schema,omitempty"`
}

// StreamOptions 流式选项
type StreamOptions struct {
	IncludeUsage bool `json:"include_usage,omitempty"`
}

// SafetySetting Gemini 安全设置
type SafetySetting struct {
	Category  string `json:"category"`
	Threshold string `json:"threshold"`
}

// ChatResponse represents a unified chat completion response
type ChatResponse struct {
	ID      string       `json:"id"`
	Object  string       `json:"object,omitempty"`  // OpenAI: "chat.completion"
	Created int64        `json:"created,omitempty"` // OpenAI: Unix timestamp
	Model   string       `json:"model"`
	Choices []ChatChoice `json:"choices"`
	Usage   UsageInfo    `json:"usage"`
	Cached  bool         `json:"cached,omitempty"` // 标记是否来自缓存
}

// ChatChoice represents a single choice in the response
type ChatChoice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason,omitempty"` // stop, length, content_filter, tool_calls
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
