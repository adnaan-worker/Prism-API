package proxy

import (
	"api-aggregator/backend/internal/adapter"
	"api-aggregator/backend/internal/protocol"
	"api-aggregator/backend/pkg/logger"
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"time"
)

// StreamWrapper 包装流式响应，用于拦截和解析 token 使用信息
type StreamWrapper struct {
	reader       io.ReadCloser
	buffer       *bytes.Buffer
	usage        *adapter.UsageInfo
	startTime    time.Time
	logger       logger.Logger
	ctx          context.Context
	service      *service
	req          *ProxyRequest
	apiConfigID  uint
	credentialID uint
	proto        protocol.Protocol
}

// NewStreamWrapper 创建流式响应包装器
func NewStreamWrapper(
	reader io.ReadCloser,
	ctx context.Context,
	service *service,
	req *ProxyRequest,
	apiConfigID uint,
	credentialID uint,
	proto protocol.Protocol,
) *StreamWrapper {
	return &StreamWrapper{
		reader:       reader,
		buffer:       &bytes.Buffer{},
		usage:        &adapter.UsageInfo{},
		startTime:    time.Now(),
		logger:       service.logger,
		ctx:          ctx,
		service:      service,
		req:          req,
		apiConfigID:  apiConfigID,
		credentialID: credentialID,
		proto:        proto,
	}
}

// Read 实现 io.Reader 接口，拦截并解析流数据
func (w *StreamWrapper) Read(p []byte) (n int, err error) {
	n, err = w.reader.Read(p)
	if n > 0 {
		// 将读取的数据写入缓冲区用于解析
		w.buffer.Write(p[:n])
	}

	// 如果读取完成（EOF），解析 token 使用信息并记录日志
	if err == io.EOF {
		w.parseUsageAndLog()
	}

	return n, err
}

// Close 实现 io.Closer 接口
func (w *StreamWrapper) Close() error {
	// 确保在关闭时也解析和记录（防止 Read 没有返回 EOF）
	if w.usage.TotalTokens == 0 {
		w.parseUsageAndLog()
	}
	return w.reader.Close()
}

// parseUsageAndLog 解析 token 使用信息并记录日志
func (w *StreamWrapper) parseUsageAndLog() {
	defer func() {
		if r := recover(); r != nil {
			w.logger.Error("Panic in parseUsageAndLog", logger.Any("panic", r))
		}
	}()

	// 解析缓冲区中的所有 SSE 数据块
	scanner := bufio.NewScanner(w.buffer)
	for scanner.Scan() {
		line := scanner.Text()

		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}

		// 解析 SSE 格式: data: {...}
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")

			// 跳过 [DONE] 标记
			if data == "[DONE]" {
				continue
			}

			// 根据协议解析数据
			if w.proto == protocol.ProtocolOpenAI || w.proto == protocol.ProtocolAnthropic {
				w.parseOpenAIChunk(data)
			} else if w.proto == protocol.ProtocolGemini {
				w.parseGeminiChunk(data)
			}
		}
	}

	// 如果没有解析到 token 使用信息，使用默认值
	if w.usage.TotalTokens == 0 {
		w.logger.Warn("No token usage found in stream, using default values")
		// 估算 token 数量（简单估算：每个字符约 0.25 个 token）
		w.usage.PromptTokens = len(w.req.ChatRequest.Messages) * 100 // 粗略估算
		w.usage.CompletionTokens = 100                                // 默认值
		w.usage.TotalTokens = w.usage.PromptTokens + w.usage.CompletionTokens
	}

	// 计算响应时间
	responseTime := time.Since(w.startTime)

	// 计算并扣除费用
	cost, err := w.service.calculateAndDeductCost(
		w.ctx,
		w.req.UserID,
		w.apiConfigID,
		w.req.Model,
		*w.usage,
	)
	if err != nil {
		w.logger.Error("✗ Failed to calculate and deduct cost",
			logger.Uint("user_id", w.req.UserID),
			logger.String("model", w.req.Model),
			logger.Error(err))
	} else {
		w.logger.Info("✓ Cost calculated and deducted",
			logger.Uint("user_id", w.req.UserID),
			logger.Int("cost", cost),
			logger.Int("total_tokens", w.usage.TotalTokens))
	}

	// 记录成功（如果使用账号池）
	if w.credentialID > 0 {
		w.service.poolManager.RecordSuccess(w.ctx, w.credentialID)
		w.logger.Info("✓ Credential success recorded", logger.Uint("credential_id", w.credentialID))
	}

	// 记录请求日志
	w.service.logRequest(
		w.ctx,
		w.req,
		w.apiConfigID,
		w.usage.TotalTokens,
		responseTime,
		nil,
	)

	w.logger.Info("✓ Stream request completed",
		logger.Uint("user_id", w.req.UserID),
		logger.String("model", w.req.Model),
		logger.Int("total_tokens", w.usage.TotalTokens),
		logger.Duration("response_time", responseTime))
}

// parseOpenAIChunk 解析 OpenAI/Anthropic 格式的流式数据块
func (w *StreamWrapper) parseOpenAIChunk(data string) {
	var chunk struct {
		Usage *adapter.UsageInfo `json:"usage,omitempty"`
	}

	if err := json.Unmarshal([]byte(data), &chunk); err != nil {
		// 解析失败，跳过
		return
	}

	// 如果包含 usage 信息，更新累计值
	if chunk.Usage != nil {
		if chunk.Usage.PromptTokens > 0 {
			w.usage.PromptTokens = chunk.Usage.PromptTokens
		}
		if chunk.Usage.CompletionTokens > 0 {
			w.usage.CompletionTokens = chunk.Usage.CompletionTokens
		}
		if chunk.Usage.TotalTokens > 0 {
			w.usage.TotalTokens = chunk.Usage.TotalTokens
		}
	}
}

// parseGeminiChunk 解析 Gemini 格式的流式数据块
func (w *StreamWrapper) parseGeminiChunk(data string) {
	var chunk struct {
		UsageMetadata *struct {
			PromptTokenCount     int `json:"promptTokenCount"`
			CandidatesTokenCount int `json:"candidatesTokenCount"`
			TotalTokenCount      int `json:"totalTokenCount"`
		} `json:"usageMetadata,omitempty"`
	}

	if err := json.Unmarshal([]byte(data), &chunk); err != nil {
		// 解析失败，跳过
		return
	}

	// 如果包含 usage 信息，更新累计值
	if chunk.UsageMetadata != nil {
		if chunk.UsageMetadata.PromptTokenCount > 0 {
			w.usage.PromptTokens = chunk.UsageMetadata.PromptTokenCount
		}
		if chunk.UsageMetadata.CandidatesTokenCount > 0 {
			w.usage.CompletionTokens = chunk.UsageMetadata.CandidatesTokenCount
		}
		if chunk.UsageMetadata.TotalTokenCount > 0 {
			w.usage.TotalTokens = chunk.UsageMetadata.TotalTokenCount
		}
	}
}
