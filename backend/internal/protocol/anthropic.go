package protocol

import (
	"api-aggregator/backend/internal/adapter"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// AnthropicConverter Anthropic 协议转换器
type AnthropicConverter struct{}

// NewAnthropicConverter 创建 Anthropic 转换器
func NewAnthropicConverter() *AnthropicConverter {
	return &AnthropicConverter{}
}

// GetProtocol 返回协议类型
func (c *AnthropicConverter) GetProtocol() Protocol {
	return ProtocolAnthropic
}

// ParseRequest 解析 Anthropic 请求为统一格式
func (c *AnthropicConverter) ParseRequest(rawBody []byte, model string) (*adapter.ChatRequest, error) {
	var anthropicReq AnthropicRequest
	if err := json.Unmarshal(rawBody, &anthropicReq); err != nil {
		return nil, fmt.Errorf("failed to parse anthropic request: %w", err)
	}

	req := &adapter.ChatRequest{
		Model:     anthropicReq.Model,
		MaxTokens: anthropicReq.MaxTokens,
	}

	// 如果提供了 model 参数，覆盖请求中的 model
	if model != "" {
		req.Model = model
	}

	// 设置可选参数
	if anthropicReq.Temperature != nil {
		req.Temperature = *anthropicReq.Temperature
	}
	if anthropicReq.TopP != nil {
		req.TopP = *anthropicReq.TopP
	}
	if anthropicReq.TopK != nil {
		req.TopK = *anthropicReq.TopK
	}
	if anthropicReq.Stream != nil {
		req.Stream = *anthropicReq.Stream
	}

	// 转换消息
	messages := make([]adapter.Message, 0, len(anthropicReq.Messages)+1)

	// 添加系统消息（Anthropic 的 system 是顶层字段）
	if anthropicReq.System != "" {
		messages = append(messages, adapter.Message{
			Role:    "system",
			Content: anthropicReq.System,
		})
	}

	// 转换用户和助手消息
	for _, msg := range anthropicReq.Messages {
		message := adapter.Message{
			Role: msg.Role,
		}

		// 处理 content - 可能是 string 或 []interface{}
		switch content := msg.Content.(type) {
		case string:
			message.Content = content
		case []interface{}:
			// 多部分内容，提取文本和工具调用
			var textParts []string
			var toolCalls []adapter.ToolCall

			for _, part := range content {
				if partMap, ok := part.(map[string]interface{}); ok {
					partType, _ := partMap["type"].(string)

					switch partType {
					case "text":
						if text, ok := partMap["text"].(string); ok {
							textParts = append(textParts, text)
						}
					case "tool_use":
						// 工具调用
						toolCall := adapter.ToolCall{
							Type: "function",
						}
						if id, ok := partMap["id"].(string); ok {
							toolCall.ID = id
						}
						if name, ok := partMap["name"].(string); ok {
							toolCall.Function.Name = name
						}
						if input, ok := partMap["input"].(map[string]interface{}); ok {
							inputBytes, _ := json.Marshal(input)
							toolCall.Function.Arguments = string(inputBytes)
						}
						toolCalls = append(toolCalls, toolCall)
					}
				}
			}

			if len(textParts) > 0 {
				message.Content = strings.Join(textParts, "\n")
			}
			if len(toolCalls) > 0 {
				message.ToolCalls = toolCalls
			}
		}

		messages = append(messages, message)
	}

	req.Messages = messages

	// 转换工具
	if len(anthropicReq.Tools) > 0 {
		tools := make([]adapter.Tool, 0, len(anthropicReq.Tools))
		for _, tool := range anthropicReq.Tools {
			tools = append(tools, adapter.Tool{
				Type: "function",
				Function: adapter.ToolFunction{
					Name:        tool.Name,
					Description: tool.Description,
					Parameters:  tool.InputSchema,
				},
			})
		}
		req.Tools = tools
	}

	return req, nil
}

// FormatResponse 将统一响应格式化为 Anthropic 格式
func (c *AnthropicConverter) FormatResponse(resp *adapter.ChatResponse) (interface{}, error) {
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	choice := resp.Choices[0]
	content := []AnthropicContent{}

	// 添加文本内容
	if choice.Message.Content != "" {
		content = append(content, AnthropicContent{
			Type: "text",
			Text: choice.Message.Content,
		})
	}

	// 添加工具调用
	if len(choice.Message.ToolCalls) > 0 {
		for _, toolCall := range choice.Message.ToolCalls {
			var input map[string]interface{}
			json.Unmarshal([]byte(toolCall.Function.Arguments), &input)

			content = append(content, AnthropicContent{
				Type:  "tool_use",
				ID:    toolCall.ID,
				Name:  toolCall.Function.Name,
				Input: input,
			})
		}
	}

	// 映射 finish_reason
	stopReason := ""
	switch choice.FinishReason {
	case "stop":
		stopReason = "end_turn"
	case "length":
		stopReason = "max_tokens"
	case "tool_calls":
		stopReason = "tool_use"
	default:
		stopReason = choice.FinishReason
	}

	anthropicResp := &AnthropicResponse{
		ID:         resp.ID,
		Type:       "message",
		Role:       "assistant",
		Content:    content,
		Model:      resp.Model,
		StopReason: stopReason,
		Usage: AnthropicUsage{
			InputTokens:  resp.Usage.PromptTokens,
			OutputTokens: resp.Usage.CompletionTokens,
		},
	}

	return anthropicResp, nil
}

// FormatStreamChunk 格式化流式响应块
// 将 OpenAI SSE 格式转换为 Anthropic SSE 格式
func (c *AnthropicConverter) FormatStreamChunk(chunk []byte) ([]byte, error) {
	// 跳过空行
	if len(bytes.TrimSpace(chunk)) == 0 {
		return []byte(""), nil
	}

	// 解析 OpenAI SSE 格式
	line := string(chunk)
	
	// 处理 data: 行
	if !strings.HasPrefix(line, "data: ") {
		return []byte(""), nil
	}

	data := strings.TrimPrefix(line, "data: ")
	data = strings.TrimSpace(data)

	// 处理 [DONE] 标记
	if data == "[DONE]" {
		// Anthropic 使用 message_stop 事件
		return []byte("event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n"), nil
	}

	// 解析 JSON
	var openaiChunk map[string]interface{}
	if err := json.Unmarshal([]byte(data), &openaiChunk); err != nil {
		return []byte(""), nil // 解析失败，返回空
	}

	// 提取 delta 内容
	choices, ok := openaiChunk["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return []byte(""), nil
	}

	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		return []byte(""), nil
	}

	delta, ok := choice["delta"].(map[string]interface{})
	if !ok {
		return []byte(""), nil
	}

	// 转换为 Anthropic 格式
	var anthropicEvents []string

	// 处理文本内容
	if content, ok := delta["content"].(string); ok && content != "" {
		event := map[string]interface{}{
			"type":  "content_block_delta",
			"index": 0,
			"delta": map[string]interface{}{
				"type": "text_delta",
				"text": content,
			},
		}
		eventJSON, _ := json.Marshal(event)
		anthropicEvents = append(anthropicEvents, fmt.Sprintf("event: content_block_delta\ndata: %s\n\n", string(eventJSON)))
	}

	// 处理工具调用
	if toolCalls, ok := delta["tool_calls"].([]interface{}); ok && len(toolCalls) > 0 {
		for _, tc := range toolCalls {
			toolCall, ok := tc.(map[string]interface{})
			if !ok {
				continue
			}

			function, ok := toolCall["function"].(map[string]interface{})
			if !ok {
				continue
			}

			event := map[string]interface{}{
				"type":  "content_block_delta",
				"index": 0,
				"delta": map[string]interface{}{
					"type":         "input_json_delta",
					"partial_json": function["arguments"],
				},
			}
			eventJSON, _ := json.Marshal(event)
			anthropicEvents = append(anthropicEvents, fmt.Sprintf("event: content_block_delta\ndata: %s\n\n", string(eventJSON)))
		}
	}

	// 处理 finish_reason
	if finishReason, ok := choice["finish_reason"].(string); ok && finishReason != "" {
		// 映射 finish_reason
		stopReason := ""
		switch finishReason {
		case "stop":
			stopReason = "end_turn"
		case "length":
			stopReason = "max_tokens"
		case "tool_calls":
			stopReason = "tool_use"
		default:
			stopReason = finishReason
		}

		event := map[string]interface{}{
			"type": "message_delta",
			"delta": map[string]interface{}{
				"stop_reason": stopReason,
			},
		}
		eventJSON, _ := json.Marshal(event)
		anthropicEvents = append(anthropicEvents, fmt.Sprintf("event: message_delta\ndata: %s\n\n", string(eventJSON)))
	}

	if len(anthropicEvents) > 0 {
		return []byte(strings.Join(anthropicEvents, "")), nil
	}

	return []byte(""), nil
}
