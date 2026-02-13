package protocol

import (
	"api-aggregator/backend/internal/adapter"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// GeminiConverter Gemini 协议转换器
type GeminiConverter struct{}

// NewGeminiConverter 创建 Gemini 转换器
func NewGeminiConverter() *GeminiConverter {
	return &GeminiConverter{}
}

// GetProtocol 返回协议类型
func (c *GeminiConverter) GetProtocol() Protocol {
	return ProtocolGemini
}

// ParseRequest 解析 Gemini 请求为统一格式
func (c *GeminiConverter) ParseRequest(rawBody []byte, model string) (*adapter.ChatRequest, error) {
	var geminiReq GeminiRequest
	if err := json.Unmarshal(rawBody, &geminiReq); err != nil {
		return nil, fmt.Errorf("failed to parse gemini request: %w", err)
	}

	req := &adapter.ChatRequest{
		Model: model,
	}

	// 处理 generation config
	if geminiReq.GenerationConfig != nil {
		if geminiReq.GenerationConfig.Temperature != nil {
			req.Temperature = *geminiReq.GenerationConfig.Temperature
		}
		if geminiReq.GenerationConfig.TopP != nil {
			req.TopP = *geminiReq.GenerationConfig.TopP
		}
		if geminiReq.GenerationConfig.TopK != nil {
			req.TopK = *geminiReq.GenerationConfig.TopK
		}
		if geminiReq.GenerationConfig.MaxOutputTokens != nil {
			req.MaxTokens = *geminiReq.GenerationConfig.MaxOutputTokens
		}
	}

	// 转换消息
	messages := make([]adapter.Message, 0, len(geminiReq.Contents)+1)

	// 添加系统指令
	if geminiReq.SystemInstruction != nil {
		var systemText strings.Builder
		for _, part := range geminiReq.SystemInstruction.Parts {
			if part.Text != "" {
				systemText.WriteString(part.Text)
			}
		}
		if systemText.Len() > 0 {
			messages = append(messages, adapter.Message{
				Role:    "system",
				Content: systemText.String(),
			})
		}
	}

	// 转换内容
	for _, content := range geminiReq.Contents {
		role := content.Role
		// Gemini 使用 "model" 作为助手角色
		if role == "model" {
			role = "assistant"
		}

		var textBuilder strings.Builder
		var toolCalls []adapter.ToolCall

		for _, part := range content.Parts {
			if part.Text != "" {
				textBuilder.WriteString(part.Text)
			}

			// 处理函数调用
			if part.FunctionCall != nil {
				argsBytes, _ := json.Marshal(part.FunctionCall.Args)
				toolCalls = append(toolCalls, adapter.ToolCall{
					ID:   fmt.Sprintf("call_%s", part.FunctionCall.Name),
					Type: "function",
					Function: adapter.FunctionCall{
						Name:      part.FunctionCall.Name,
						Arguments: string(argsBytes),
					},
				})
			}
		}

		message := adapter.Message{
			Role:    role,
			Content: textBuilder.String(),
		}

		if len(toolCalls) > 0 {
			message.ToolCalls = toolCalls
		}

		if message.Content != "" || len(message.ToolCalls) > 0 {
			messages = append(messages, message)
		}
	}

	req.Messages = messages

	// 转换工具
	if len(geminiReq.Tools) > 0 {
		tools := make([]adapter.Tool, 0)
		for _, toolDecl := range geminiReq.Tools {
			for _, funcDecl := range toolDecl.FunctionDeclarations {
				tools = append(tools, adapter.Tool{
					Type: "function",
					Function: adapter.ToolFunction{
						Name:        funcDecl.Name,
						Description: funcDecl.Description,
						Parameters:  funcDecl.Parameters,
					},
				})
			}
		}
		req.Tools = tools
	}

	return req, nil
}

// FormatResponse 将统一响应格式化为 Gemini 格式
func (c *GeminiConverter) FormatResponse(resp *adapter.ChatResponse) (interface{}, error) {
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	choice := resp.Choices[0]
	parts := []GeminiPart{}

	// 添加文本内容
	if choice.Message.Content != "" {
		parts = append(parts, GeminiPart{
			Text: choice.Message.Content,
		})
	}

	// 添加函数调用
	if len(choice.Message.ToolCalls) > 0 {
		for _, toolCall := range choice.Message.ToolCalls {
			var args map[string]interface{}
			json.Unmarshal([]byte(toolCall.Function.Arguments), &args)

			parts = append(parts, GeminiPart{
				FunctionCall: &GeminiFunctionCall{
					Name: toolCall.Function.Name,
					Args: args,
				},
			})
		}
	}

	// 映射 finish_reason（Gemini 使用大写）
	finishReason := ""
	switch choice.FinishReason {
	case "stop":
		finishReason = "STOP"
	case "length":
		finishReason = "MAX_TOKENS"
	case "tool_calls":
		finishReason = "STOP" // Gemini 没有专门的 tool_calls finish reason
	case "content_filter":
		finishReason = "SAFETY"
	default:
		finishReason = "OTHER"
	}

	geminiResp := &GeminiResponse{
		Candidates: []GeminiCandidate{
			{
				Content: GeminiContent{
					Parts: parts,
					Role:  "model",
				},
				FinishReason: finishReason,
				Index:        choice.Index,
			},
		},
		UsageMetadata: GeminiUsage{
			PromptTokenCount:     resp.Usage.PromptTokens,
			CandidatesTokenCount: resp.Usage.CompletionTokens,
			TotalTokenCount:      resp.Usage.TotalTokens,
		},
	}

	return geminiResp, nil
}

// FormatStreamChunk 格式化流式响应块
// 将 OpenAI SSE 格式转换为 Gemini 流式格式（纯 JSON，非 SSE）
func (c *GeminiConverter) FormatStreamChunk(chunk []byte) ([]byte, error) {
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
		// Gemini 流式结束不需要特殊标记
		return []byte(""), nil
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

	// 转换为 Gemini 格式
	parts := []GeminiPart{}

	// 处理文本内容
	if content, ok := delta["content"].(string); ok && content != "" {
		parts = append(parts, GeminiPart{
			Text: content,
		})
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

			name, _ := function["name"].(string)
			argsStr, _ := function["arguments"].(string)
			
			var args map[string]interface{}
			if argsStr != "" {
				json.Unmarshal([]byte(argsStr), &args)
			}

			parts = append(parts, GeminiPart{
				FunctionCall: &GeminiFunctionCall{
					Name: name,
					Args: args,
				},
			})
		}
	}

	// 如果有内容，构建 Gemini 响应
	if len(parts) > 0 {
		geminiChunk := map[string]interface{}{
			"candidates": []map[string]interface{}{
				{
					"content": map[string]interface{}{
						"parts": parts,
						"role":  "model",
					},
					"index": 0,
				},
			},
		}

		// 处理 finish_reason
		if finishReason, ok := choice["finish_reason"].(string); ok && finishReason != "" {
			// 映射 finish_reason（Gemini 使用大写）
			mappedReason := ""
			switch finishReason {
			case "stop":
				mappedReason = "STOP"
			case "length":
				mappedReason = "MAX_TOKENS"
			case "content_filter":
				mappedReason = "SAFETY"
			default:
				mappedReason = "OTHER"
			}
			geminiChunk["candidates"].([]map[string]interface{})[0]["finishReason"] = mappedReason
		}

		chunkJSON, err := json.Marshal(geminiChunk)
		if err != nil {
			return []byte(""), nil
		}

		// Gemini 流式响应是纯 JSON，每行一个对象，不使用 SSE 格式
		return append(chunkJSON, '\n'), nil
	}

	return []byte(""), nil
}
