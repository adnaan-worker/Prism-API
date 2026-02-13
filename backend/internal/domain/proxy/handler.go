package proxy

import (
	"api-aggregator/backend/internal/protocol"
	"api-aggregator/backend/pkg/response"
	"bufio"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Handler 代理处理器
type Handler struct {
	service          Service
	converterFactory *protocol.ConverterFactory
}

// NewHandler 创建代理处理器
func NewHandler(service Service) *Handler {
	return &Handler{
		service:          service,
		converterFactory: protocol.NewConverterFactory(),
	}
}

// ChatCompletions 处理聊天补全请求 (OpenAI 协议)
// @Summary 聊天补全
// @Description 处理 OpenAI 兼容的聊天补全请求
// @Tags Proxy
// @Accept json
// @Produce json
// @Param request body adapter.ChatRequest true "聊天请求"
// @Success 200 {object} adapter.ChatResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/chat/completions [post]
func (h *Handler) ChatCompletions(c *gin.Context) {
	h.handleRequest(c, protocol.ProtocolOpenAI, "")
}

// ChatCompletionsOpenAI OpenAI 格式的聊天补全（别名）
func (h *Handler) ChatCompletionsOpenAI(c *gin.Context) {
	h.ChatCompletions(c)
}

// ChatCompletionsAnthropic Anthropic 格式的聊天补全
// @Summary Anthropic 聊天补全
// @Description 处理 Anthropic 格式的聊天补全请求
// @Tags Proxy
// @Accept json
// @Produce json
// @Param request body protocol.AnthropicRequest true "聊天请求"
// @Success 200 {object} protocol.AnthropicResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/messages [post]
func (h *Handler) ChatCompletionsAnthropic(c *gin.Context) {
	h.handleRequest(c, protocol.ProtocolAnthropic, "")
}

// ChatCompletionsGemini Gemini 格式的聊天补全
// @Summary Gemini 聊天补全
// @Description 处理 Gemini 格式的聊天补全请求
// @Tags Proxy
// @Accept json
// @Produce json
// @Param action path string true "模型和操作，格式: model:generateContent"
// @Param request body protocol.GeminiRequest true "聊天请求"
// @Success 200 {object} protocol.GeminiResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/models/{model}:generateContent [post]
func (h *Handler) ChatCompletionsGemini(c *gin.Context) {
	// 从路径参数获取完整路径，格式: /model:generateContent 或 /model:streamGenerateContent
	action := c.Param("action")

	// 移除开头的斜杠
	if len(action) > 0 && action[0] == '/' {
		action = action[1:]
	}

	// 分割模型名和操作
	parts := strings.Split(action, ":")
	if len(parts) != 2 {
		response.Error(c, http.StatusBadRequest, 400001, "Invalid Gemini API path format, expected: /models/{model}:generateContent", nil)
		return
	}

	model := parts[0]
	h.handleRequest(c, protocol.ProtocolGemini, model)
}

// handleRequest 统一处理请求的核心方法
func (h *Handler) handleRequest(c *gin.Context, proto protocol.Protocol, model string) {
	// 1. 获取协议转换器
	converter := h.converterFactory.GetConverter(proto)
	if converter == nil {
		response.Error(c, http.StatusInternalServerError, 500001, "Protocol converter not found", nil)
		return
	}

	// 2. 读取原始请求体
	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 400001, "Failed to read request body", err)
		return
	}

	// 3. 使用转换器解析请求
	chatReq, err := converter.ParseRequest(rawBody, model)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 400001, "Failed to parse request", err)
		return
	}

	// 4. 从上下文获取用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, 401001, "User ID not found in context", nil)
		return
	}

	apiKeyID, exists := c.Get("api_key_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, 401001, "API Key ID not found in context", nil)
		return
	}

	// 5. 构建代理请求
	proxyReq := &ProxyRequest{
		UserID:      userID.(uint),
		APIKeyID:    apiKeyID.(uint),
		Model:       chatReq.Model,
		Stream:      chatReq.Stream,
		ChatRequest: chatReq,
	}

	// 6. 处理流式请求
	if chatReq.Stream {
		h.handleStream(c, proxyReq, converter)
		return
	}

	// 7. 处理非流式请求
	resp, err := h.service.ChatCompletions(c.Request.Context(), proxyReq)
	if err != nil {
		response.ErrorFromError(c, err)
		return
	}

	// 8. 使用转换器格式化响应
	formattedResp, err := converter.FormatResponse(resp)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500001, "Failed to format response", err)
		return
	}

	// 9. 返回响应 - 所有协议都直接返回原始格式，不使用包装器
	c.JSON(http.StatusOK, formattedResp)
}

// handleStream 处理流式请求
func (h *Handler) handleStream(c *gin.Context, req *ProxyRequest, converter protocol.Converter) {
	// 调用服务
	resp, err := h.service.ChatCompletionsStream(c.Request.Context(), req)
	if err != nil {
		response.ErrorFromError(c, err)
		return
	}
	defer resp.Body.Close()

	// 根据协议设置不同的响应头
	proto := converter.GetProtocol()
	if proto == protocol.ProtocolGemini {
		// Gemini 使用普通 JSON 流，不是 SSE
		c.Header("Content-Type", "application/json")
	} else {
		// OpenAI 和 Anthropic 使用 SSE
		c.Header("Content-Type", "text/event-stream")
	}
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	// 复制响应流
	c.Stream(func(w io.Writer) bool {
		// 使用 bufio.Reader 逐行读取
		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				if err != io.EOF {
					// 记录错误但不中断流
					if proto != protocol.ProtocolGemini {
						c.SSEvent("error", err.Error())
					}
				}
				return false
			}

			// 使用转换器格式化流式数据块
			formattedChunk, err := converter.FormatStreamChunk(line)
			if err != nil {
				// 格式化失败，跳过这个块
				continue
			}

			// 跳过空块
			if len(formattedChunk) == 0 {
				continue
			}

			// 写入响应
			if _, err := w.Write(formattedChunk); err != nil {
				return false
			}

			// 刷新缓冲区
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	})
}
