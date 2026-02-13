package protocol

import (
	"api-aggregator/backend/internal/adapter"
)

// Protocol 协议类型
type Protocol string

const (
	ProtocolOpenAI    Protocol = "openai"
	ProtocolAnthropic Protocol = "anthropic"
	ProtocolGemini    Protocol = "gemini"
)

// Converter 协议转换器接口
// 负责在客户端协议格式和统一格式之间转换
type Converter interface {
	// GetProtocol 返回协议类型
	GetProtocol() Protocol

	// ParseRequest 解析客户端请求为统一格式
	// rawBody: 原始请求体 (JSON bytes)
	// model: 模型名称（某些协议需要从路径提取）
	// 返回: 统一的 ChatRequest
	ParseRequest(rawBody []byte, model string) (*adapter.ChatRequest, error)

	// FormatResponse 将统一响应格式化为客户端期望的格式
	// resp: 统一的 ChatResponse
	// 返回: 客户端协议格式的响应 (map 或 struct)
	FormatResponse(resp *adapter.ChatResponse) (interface{}, error)

	// FormatStreamChunk 格式化流式响应块
	// chunk: SSE 数据块
	// 返回: 客户端协议格式的 SSE 数据
	FormatStreamChunk(chunk []byte) ([]byte, error)
}

// ConverterFactory 转换器工厂
type ConverterFactory struct {
	converters map[Protocol]Converter
}

// NewConverterFactory 创建转换器工厂
func NewConverterFactory() *ConverterFactory {
	factory := &ConverterFactory{
		converters: make(map[Protocol]Converter),
	}

	// 注册所有转换器
	factory.Register(NewOpenAIConverter())
	factory.Register(NewAnthropicConverter())
	factory.Register(NewGeminiConverter())

	return factory
}

// Register 注册转换器
func (f *ConverterFactory) Register(converter Converter) {
	f.converters[converter.GetProtocol()] = converter
}

// GetConverter 获取转换器
func (f *ConverterFactory) GetConverter(protocol Protocol) Converter {
	return f.converters[protocol]
}
