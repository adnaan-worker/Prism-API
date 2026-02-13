package adapter

// Extended types to support advanced features like multimodal input, vision, etc.
// These types complement the core types defined in adapter.go

// ContentPart represents a part of message content (for multimodal support)
type ContentPart struct {
	Type       string       `json:"type"` // text, image_url, image, inline_data, tool_use, tool_result
	Text       string       `json:"text,omitempty"`
	ImageURL   *ImageURL    `json:"image_url,omitempty"`
	Source     *ImageSource `json:"source,omitempty"`      // Anthropic
	InlineData *InlineData  `json:"inline_data,omitempty"` // Gemini

	// Tool use (Anthropic)
	ID    string                 `json:"id,omitempty"`
	Name  string                 `json:"name,omitempty"`
	Input map[string]interface{} `json:"input,omitempty"`

	// Tool result (Anthropic)
	ToolUseID string `json:"tool_use_id,omitempty"`
}

// ImageURL represents an image URL (OpenAI)
type ImageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"` // low, high, auto
}

// ImageSource represents an image source (Anthropic)
type ImageSource struct {
	Type      string `json:"type"` // base64, url
	MediaType string `json:"media_type"`
	Data      string `json:"data,omitempty"`
	URL       string `json:"url,omitempty"`
}

// InlineData represents inline data (Gemini)
type InlineData struct {
	MimeType string `json:"mime_type"`
	Data     string `json:"data"`
}
