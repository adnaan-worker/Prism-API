package proxy

import (
	"api-aggregator/backend/internal/adapter"
	"api-aggregator/backend/pkg/response"
	"bufio"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Handler 代理处理器
type Handler struct {
	service Service
}

// NewHandler 创建代理处理器
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// ChatCompletions 处理聊天补全请求
// @Summary 聊天补全
// @Description 处理 OpenAI 兼容的聊天补全请求，支持 OpenAI、Anthropic、Gemini 格式
// @Tags Proxy
// @Accept json
// @Produce json
// @Param request body adapter.ChatRequest true "聊天请求"
// @Success 200 {object} adapter.ChatResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/chat/completions [post]
func (h *Handler) ChatCompletions(c *gin.Context) {
	// 解析请求
	var chatReq adapter.ChatRequest
	if err := c.ShouldBindJSON(&chatReq); err != nil {
		response.Error(c, http.StatusBadRequest, 400001, "Invalid request format", err)
		return
	}

	// 从上下文获取用户信息
	userID, _ := c.Get("user_id")
	apiKeyID, _ := c.Get("api_key_id")

	// 构建代理请求
	proxyReq := &ProxyRequest{
		UserID:      userID.(uint),
		APIKeyID:    apiKeyID.(uint),
		Model:       chatReq.Model,
		Stream:      chatReq.Stream,
		ChatRequest: &chatReq,
	}

	// 处理流式请求
	if chatReq.Stream {
		h.handleStream(c, proxyReq)
		return
	}

	// 处理非流式请求
	resp, err := h.service.ChatCompletions(c.Request.Context(), proxyReq)
	if err != nil {
		response.ErrorFromError(c, err)
		return
	}

	response.Success(c, resp)
}

// handleStream 处理流式请求
func (h *Handler) handleStream(c *gin.Context, req *ProxyRequest) {
	// 调用服务
	resp, err := h.service.ChatCompletionsStream(c.Request.Context(), req)
	if err != nil {
		response.ErrorFromError(c, err)
		return
	}
	defer resp.Body.Close()

	// 设置响应头
	c.Header("Content-Type", "text/event-stream")
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
					c.SSEvent("error", err.Error())
				}
				return false
			}

			// 写入响应
			if _, err := w.Write(line); err != nil {
				return false
			}

			// 刷新缓冲区
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	})
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
// @Param request body adapter.ChatRequest true "聊天请求"
// @Success 200 {object} adapter.ChatResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/messages [post]
func (h *Handler) ChatCompletionsAnthropic(c *gin.Context) {
	// Anthropic 使用相同的处理逻辑，adapter 会自动转换格式
	h.ChatCompletions(c)
}

// ChatCompletionsGemini Gemini 格式的聊天补全
// @Summary Gemini 聊天补全
// @Description 处理 Gemini 格式的聊天补全请求
// @Tags Proxy
// @Accept json
// @Produce json
// @Param action path string true "模型和操作，格式: model:generateContent"
// @Param request body adapter.ChatRequest true "聊天请求"
// @Success 200 {object} adapter.ChatResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/models/{model}:generateContent [post]
func (h *Handler) ChatCompletionsGemini(c *gin.Context) {
	// 从路径参数获取完整路径，格式: /model:generateContent 或 /model:streamGenerateContent
	action := c.Param("action")
	
	// 解析模型名称和操作
	// action 格式: /model-name:generateContent
	var model string
	var isStream bool
	
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
	
	model = parts[0]
	operation := parts[1]
	
	// 判断是否为流式请求
	isStream = operation == "streamGenerateContent"
	
	// 解析请求
	var chatReq adapter.ChatRequest
	if err := c.ShouldBindJSON(&chatReq); err != nil {
		response.Error(c, http.StatusBadRequest, 400001, "Invalid request format", err)
		return
	}

	// 设置模型和流式标志
	chatReq.Model = model
	chatReq.Stream = isStream

	// 从上下文获取用户信息
	userID, _ := c.Get("user_id")
	apiKeyID, _ := c.Get("api_key_id")

	// 构建代理请求
	proxyReq := &ProxyRequest{
		UserID:      userID.(uint),
		APIKeyID:    apiKeyID.(uint),
		Model:       model,
		Stream:      isStream,
		ChatRequest: &chatReq,
	}

	// 处理流式请求
	if isStream {
		h.handleStream(c, proxyReq)
		return
	}

	// 处理非流式请求
	resp, err := h.service.ChatCompletions(c.Request.Context(), proxyReq)
	if err != nil {
		response.ErrorFromError(c, err)
		return
	}

	response.Success(c, resp)
}
