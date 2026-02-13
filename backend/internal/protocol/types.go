package protocol

// AnthropicRequest Anthropic 原生请求格式
type AnthropicRequest struct {
	Model         string             `json:"model"`
	Messages      []AnthropicMessage `json:"messages"`
	System        string             `json:"system,omitempty"`
	MaxTokens     int                `json:"max_tokens"`
	Temperature   *float64           `json:"temperature,omitempty"`
	TopP          *float64           `json:"top_p,omitempty"`
	TopK          *int               `json:"top_k,omitempty"`
	StopSequences []string           `json:"stop_sequences,omitempty"`
	Stream        *bool              `json:"stream,omitempty"`
	Tools         []AnthropicTool    `json:"tools,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

type AnthropicMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"` // string or []AnthropicContent
}

type AnthropicContent struct {
	Type      string                 `json:"type"` // text, image, tool_use, tool_result
	Text      string                 `json:"text,omitempty"`
	Source    *AnthropicImageSource  `json:"source,omitempty"`
	ID        string                 `json:"id,omitempty"`
	Name      string                 `json:"name,omitempty"`
	Input     map[string]interface{} `json:"input,omitempty"`
	ToolUseID string                 `json:"tool_use_id,omitempty"`
	Content   interface{}            `json:"content,omitempty"`
}

type AnthropicImageSource struct {
	Type      string `json:"type"` // base64, url
	MediaType string `json:"media_type"`
	Data      string `json:"data,omitempty"`
	URL       string `json:"url,omitempty"`
}

type AnthropicTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

// AnthropicResponse Anthropic 原生响应格式
type AnthropicResponse struct {
	ID           string                   `json:"id"`
	Type         string                   `json:"type"` // "message"
	Role         string                   `json:"role"` // "assistant"
	Content      []AnthropicContent       `json:"content"`
	Model        string                   `json:"model"`
	StopReason   string                   `json:"stop_reason,omitempty"`
	StopSequence string                   `json:"stop_sequence,omitempty"`
	Usage        AnthropicUsage           `json:"usage"`
}

type AnthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// GeminiRequest Gemini 原生请求格式
type GeminiRequest struct {
	Contents          []GeminiContent         `json:"contents"`
	SystemInstruction *GeminiContent          `json:"systemInstruction,omitempty"`
	Tools             []GeminiToolDeclaration `json:"tools,omitempty"`
	SafetySettings    []GeminiSafetySetting   `json:"safetySettings,omitempty"`
	GenerationConfig  *GeminiGenerationConfig `json:"generationConfig,omitempty"`
}

type GeminiContent struct {
	Role  string       `json:"role,omitempty"` // user, model
	Parts []GeminiPart `json:"parts"`
}

type GeminiPart struct {
	Text             string                  `json:"text,omitempty"`
	InlineData       *GeminiInlineData       `json:"inlineData,omitempty"`
	FunctionCall     *GeminiFunctionCall     `json:"functionCall,omitempty"`
	FunctionResponse *GeminiFunctionResponse `json:"functionResponse,omitempty"`
}

type GeminiInlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

type GeminiFunctionCall struct {
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args"`
}

type GeminiFunctionResponse struct {
	Name     string                 `json:"name"`
	Response map[string]interface{} `json:"response"`
}

type GeminiToolDeclaration struct {
	FunctionDeclarations []GeminiFunctionDeclaration `json:"functionDeclarations"`
}

type GeminiFunctionDeclaration struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type GeminiSafetySetting struct {
	Category  string `json:"category"`
	Threshold string `json:"threshold"`
}

type GeminiGenerationConfig struct {
	Temperature     *float64 `json:"temperature,omitempty"`
	TopP            *float64 `json:"topP,omitempty"`
	TopK            *int     `json:"topK,omitempty"`
	MaxOutputTokens *int     `json:"maxOutputTokens,omitempty"`
	StopSequences   []string `json:"stopSequences,omitempty"`
}

// GeminiResponse Gemini 原生响应格式
type GeminiResponse struct {
	Candidates    []GeminiCandidate `json:"candidates"`
	UsageMetadata GeminiUsage       `json:"usageMetadata,omitempty"`
}

type GeminiCandidate struct {
	Content      GeminiContent `json:"content"`
	FinishReason string        `json:"finishReason,omitempty"`
	Index        int           `json:"index"`
}

type GeminiUsage struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}
