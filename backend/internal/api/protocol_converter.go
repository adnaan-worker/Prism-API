package api

import (
	"api-aggregator/backend/internal/adapter"
	"encoding/json"
	"fmt"
)

// ProtocolConverter handles conversion between different API protocols
type ProtocolConverter struct{}

func NewProtocolConverter() *ProtocolConverter {
	return &ProtocolConverter{}
}

// ConvertToInternalFormat converts any protocol request to internal ChatRequest format
func (pc *ProtocolConverter) ConvertToInternalFormat(rawReq map[string]interface{}, protocol string) (*adapter.ChatRequest, error) {
	switch protocol {
	case "openai":
		return pc.convertOpenAIToInternal(rawReq)
	case "anthropic":
		return pc.convertAnthropicToInternal(rawReq)
	case "gemini":
		return pc.convertGeminiToInternal(rawReq)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", protocol)
	}
}

// ConvertFromInternalFormat converts internal ChatResponse to requested protocol format
func (pc *ProtocolConverter) ConvertFromInternalFormat(resp *adapter.ChatResponse, protocol string) (interface{}, error) {
	switch protocol {
	case "openai":
		return resp, nil // Already in OpenAI format
	case "anthropic":
		return pc.convertInternalToAnthropic(resp), nil
	case "gemini":
		return pc.convertInternalToGemini(resp), nil
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", protocol)
	}
}

// ===== OpenAI Protocol =====

func (pc *ProtocolConverter) convertOpenAIToInternal(rawReq map[string]interface{}) (*adapter.ChatRequest, error) {
	// OpenAI format is our internal format, direct conversion
	var req adapter.ChatRequest
	reqBytes, _ := json.Marshal(rawReq)
	if err := json.Unmarshal(reqBytes, &req); err != nil {
		return nil, err
	}
	return &req, nil
}

// ===== Anthropic Protocol =====

func (pc *ProtocolConverter) convertAnthropicToInternal(rawReq map[string]interface{}) (*adapter.ChatRequest, error) {
	var req adapter.ChatRequest
	
	// Extract basic fields
	if model, ok := rawReq["model"].(string); ok {
		req.Model = model
	}
	if stream, ok := rawReq["stream"].(bool); ok {
		req.Stream = stream
	}
	if maxTokens, ok := rawReq["max_tokens"].(float64); ok {
		req.MaxTokens = int(maxTokens)
	}
	if temp, ok := rawReq["temperature"].(float64); ok {
		req.Temperature = temp
	}
	if topP, ok := rawReq["top_p"].(float64); ok {
		req.TopP = topP
	}
	
	// Convert messages
	if messagesRaw, ok := rawReq["messages"].([]interface{}); ok {
		req.Messages = make([]adapter.Message, 0, len(messagesRaw))
		for _, msgRaw := range messagesRaw {
			msgMap, _ := msgRaw.(map[string]interface{})
			msg := adapter.Message{
				Role: msgMap["role"].(string),
			}
			
			// Handle content (can be string or array)
			if content, ok := msgMap["content"].(string); ok {
				msg.Content = content
			} else if contentArr, ok := msgMap["content"].([]interface{}); ok {
				// Convert content array to string (extract text parts)
				var textParts []string
				for _, part := range contentArr {
					partMap, _ := part.(map[string]interface{})
					if partType, _ := partMap["type"].(string); partType == "text" {
						if text, _ := partMap["text"].(string); text != "" {
							textParts = append(textParts, text)
						}
					}
				}
				msg.Content = joinStrings(textParts, "\n")
			}
			
			req.Messages = append(req.Messages, msg)
		}
	}
	
	// Convert system prompt
	if system, ok := rawReq["system"].(string); ok {
		// Prepend system message
		systemMsg := adapter.Message{
			Role:    "system",
			Content: system,
		}
		req.Messages = append([]adapter.Message{systemMsg}, req.Messages...)
	}
	
	// Convert tools (Anthropic format to OpenAI format)
	if toolsRaw, ok := rawReq["tools"].([]interface{}); ok {
		req.Tools = make([]adapter.Tool, 0, len(toolsRaw))
		for _, toolRaw := range toolsRaw {
			toolMap, _ := toolRaw.(map[string]interface{})
			
			// Anthropic format: {name, description, input_schema}
			tool := adapter.Tool{
				Type: "function",
				Function: adapter.ToolFunction{
					Name:        toolMap["name"].(string),
					Description: getString(toolMap, "description"),
					Parameters:  getMap(toolMap, "input_schema"),
				},
			}
			req.Tools = append(req.Tools, tool)
		}
	}
	
	return &req, nil
}

func (pc *ProtocolConverter) convertInternalToAnthropic(openaiResp *adapter.ChatResponse) map[string]interface{} {
	if openaiResp == nil || len(openaiResp.Choices) == 0 {
		return map[string]interface{}{
			"id":      openaiResp.ID,
			"type":    "message",
			"role":    "assistant",
			"content": []interface{}{},
			"model":   openaiResp.Model,
			"usage": map[string]interface{}{
				"input_tokens":  openaiResp.Usage.PromptTokens,
				"output_tokens": openaiResp.Usage.CompletionTokens,
			},
		}
	}

	choice := openaiResp.Choices[0]
	message := choice.Message

	// Build content array
	content := []interface{}{}
	
	// Add text content if present
	if message.Content != "" {
		content = append(content, map[string]interface{}{
			"type": "text",
			"text": message.Content,
		})
	}

	// Add tool calls if present
	if len(message.ToolCalls) > 0 {
		for _, tc := range message.ToolCalls {
			var input map[string]interface{}
			json.Unmarshal([]byte(tc.Function.Arguments), &input)
			
			content = append(content, map[string]interface{}{
				"type":  "tool_use",
				"id":    tc.ID,
				"name":  tc.Function.Name,
				"input": input,
			})
		}
	}

	// Determine stop reason
	stopReason := "end_turn"
	if choice.FinishReason == "tool_calls" {
		stopReason = "tool_use"
	} else if choice.FinishReason == "length" {
		stopReason = "max_tokens"
	}

	return map[string]interface{}{
		"id":          openaiResp.ID,
		"type":        "message",
		"role":        "assistant",
		"content":     content,
		"model":       openaiResp.Model,
		"stop_reason": stopReason,
		"usage": map[string]interface{}{
			"input_tokens":  openaiResp.Usage.PromptTokens,
			"output_tokens": openaiResp.Usage.CompletionTokens,
		},
	}
}

// ===== Gemini Protocol =====

func (pc *ProtocolConverter) convertGeminiToInternal(rawReq map[string]interface{}) (*adapter.ChatRequest, error) {
	var req adapter.ChatRequest
	
	// Gemini doesn't have model in request body, it's in URL
	// Will be set by handler
	
	// Convert contents to messages
	if contentsRaw, ok := rawReq["contents"].([]interface{}); ok {
		req.Messages = make([]adapter.Message, 0, len(contentsRaw))
		for _, contentRaw := range contentsRaw {
			contentMap, _ := contentRaw.(map[string]interface{})
			msg := adapter.Message{
				Role: contentMap["role"].(string),
			}
			
			// Convert role (Gemini uses "user"/"model", we use "user"/"assistant")
			if msg.Role == "model" {
				msg.Role = "assistant"
			}
			
			// Extract text from parts
			if partsRaw, ok := contentMap["parts"].([]interface{}); ok {
				var textParts []string
				for _, partRaw := range partsRaw {
					partMap, _ := partRaw.(map[string]interface{})
					if text, ok := partMap["text"].(string); ok {
						textParts = append(textParts, text)
					}
				}
				msg.Content = joinStrings(textParts, "\n")
			}
			
			req.Messages = append(req.Messages, msg)
		}
	}
	
	// Convert system instruction
	if systemInst, ok := rawReq["systemInstruction"].(map[string]interface{}); ok {
		if partsRaw, ok := systemInst["parts"].([]interface{}); ok {
			var textParts []string
			for _, partRaw := range partsRaw {
				partMap, _ := partRaw.(map[string]interface{})
				if text, ok := partMap["text"].(string); ok {
					textParts = append(textParts, text)
				}
			}
			systemMsg := adapter.Message{
				Role:    "system",
				Content: joinStrings(textParts, "\n"),
			}
			req.Messages = append([]adapter.Message{systemMsg}, req.Messages...)
		}
	}
	
	// Convert generation config
	if genConfig, ok := rawReq["generationConfig"].(map[string]interface{}); ok {
		if temp, ok := genConfig["temperature"].(float64); ok {
			req.Temperature = temp
		}
		if topP, ok := genConfig["topP"].(float64); ok {
			req.TopP = topP
		}
		if maxTokens, ok := genConfig["maxOutputTokens"].(float64); ok {
			req.MaxTokens = int(maxTokens)
		}
	}
	
	return &req, nil
}

func (pc *ProtocolConverter) convertInternalToGemini(openaiResp *adapter.ChatResponse) map[string]interface{} {
	if openaiResp == nil || len(openaiResp.Choices) == 0 {
		return map[string]interface{}{
			"candidates": []interface{}{},
			"usageMetadata": map[string]interface{}{
				"promptTokenCount":     openaiResp.Usage.PromptTokens,
				"candidatesTokenCount": openaiResp.Usage.CompletionTokens,
				"totalTokenCount":      openaiResp.Usage.TotalTokens,
			},
		}
	}

	choice := openaiResp.Choices[0]
	message := choice.Message

	// Build parts array
	parts := []interface{}{}
	if message.Content != "" {
		parts = append(parts, map[string]interface{}{
			"text": message.Content,
		})
	}

	// Determine finish reason
	finishReason := "STOP"
	if choice.FinishReason == "length" {
		finishReason = "MAX_TOKENS"
	}

	return map[string]interface{}{
		"candidates": []interface{}{
			map[string]interface{}{
				"content": map[string]interface{}{
					"parts": parts,
					"role":  "model",
				},
				"finishReason": finishReason,
				"index":        0,
			},
		},
		"usageMetadata": map[string]interface{}{
			"promptTokenCount":     openaiResp.Usage.PromptTokens,
			"candidatesTokenCount": openaiResp.Usage.CompletionTokens,
			"totalTokenCount":      openaiResp.Usage.TotalTokens,
		},
	}
}

// ===== Helper Functions =====

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getMap(m map[string]interface{}, key string) map[string]interface{} {
	if v, ok := m[key].(map[string]interface{}); ok {
		return v
	}
	return make(map[string]interface{})
}

func joinStrings(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += sep + parts[i]
	}
	return result
}
