package protocol

import (
	"api-aggregator/backend/internal/adapter"
	"encoding/json"
	"fmt"
)

// OpenAIConverter OpenAI 协议转换器
type OpenAIConverter struct{}

// NewOpenAIConverter 创建 OpenAI 转换器
func NewOpenAIConverter() *OpenAIConverter {
	return &OpenAIConverter{}
}

// GetProtocol 返回协议类型
func (c *OpenAIConverter) GetProtocol() Protocol {
	return ProtocolOpenAI
}

// ParseRequest 解析 OpenAI 请求为统一格式
// OpenAI 格式就是内部统一格式，直接解析即可
func (c *OpenAIConverter) ParseRequest(rawBody []byte, model string) (*adapter.ChatRequest, error) {
	var req adapter.ChatRequest
	if err := json.Unmarshal(rawBody, &req); err != nil {
		return nil, fmt.Errorf("failed to parse openai request: %w", err)
	}

	// 如果提供了 model 参数，覆盖请求中的 model
	if model != "" {
		req.Model = model
	}

	return &req, nil
}

// FormatResponse 将统一响应格式化为 OpenAI 格式
// OpenAI 格式就是内部统一格式，直接返回即可
func (c *OpenAIConverter) FormatResponse(resp *adapter.ChatResponse) (interface{}, error) {
	return resp, nil
}

// FormatStreamChunk 格式化流式响应块
// OpenAI 格式就是内部统一格式，直接透传
func (c *OpenAIConverter) FormatStreamChunk(chunk []byte) ([]byte, error) {
	return chunk, nil
}
