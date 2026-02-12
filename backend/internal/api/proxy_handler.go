package api

import (
	"api-aggregator/backend/internal/adapter"
	"api-aggregator/backend/internal/service"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type ProxyHandler struct {
	proxyService       *service.ProxyService
	protocolConverter  *ProtocolConverter
}

func NewProxyHandler(proxyService *service.ProxyService) *ProxyHandler {
	return &ProxyHandler{
		proxyService:      proxyService,
		protocolConverter: NewProtocolConverter(),
	}
}

// ChatCompletions handles OpenAI-compatible chat completions
func (h *ProxyHandler) ChatCompletions(c *gin.Context) {
	// Extract API key
	apiKey := h.extractAPIKey(c, "Bearer ")
	if apiKey == "" {
		h.respondError(c, http.StatusUnauthorized, "openai", "Missing or invalid API key")
		return
	}

	// Parse raw request
	var rawReq map[string]interface{}
	if err := c.ShouldBindJSON(&rawReq); err != nil {
		h.respondError(c, http.StatusBadRequest, "openai", fmt.Sprintf("Invalid request: %v", err))
		return
	}

	// Convert to internal format
	req, err := h.protocolConverter.ConvertToInternalFormat(rawReq, "openai")
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "openai", fmt.Sprintf("Failed to parse request: %v", err))
		return
	}

	// Handle streaming vs non-streaming
	if req.Stream {
		h.handleStreamingRequest(c, apiKey, req, "openai")
		return
	}

	// Proxy the request
	resp, err := h.proxyService.ProxyRequest(c.Request.Context(), apiKey, req)
	if err != nil {
		h.handleProxyError(c, err, "openai")
		return
	}

	// Convert response to requested protocol format
	finalResp, err := h.protocolConverter.ConvertFromInternalFormat(resp, "openai")
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, "openai", fmt.Sprintf("Failed to convert response: %v", err))
		return
	}

	c.JSON(http.StatusOK, finalResp)
}

// AnthropicMessages handles Anthropic-compatible messages endpoint
func (h *ProxyHandler) AnthropicMessages(c *gin.Context) {
	// Extract API key (Anthropic uses x-api-key header)
	apiKey := c.GetHeader("x-api-key")
	if apiKey == "" {
		apiKey = h.extractAPIKey(c, "Bearer ")
	}
	if apiKey == "" {
		h.respondError(c, http.StatusUnauthorized, "anthropic", "Missing API key")
		return
	}

	// Parse raw request
	var rawReq map[string]interface{}
	if err := c.ShouldBindJSON(&rawReq); err != nil {
		h.respondError(c, http.StatusBadRequest, "anthropic", fmt.Sprintf("Invalid request: %v", err))
		return
	}

	// Convert to internal format
	req, err := h.protocolConverter.ConvertToInternalFormat(rawReq, "anthropic")
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "anthropic", fmt.Sprintf("Failed to parse request: %v", err))
		return
	}

	// Handle streaming vs non-streaming
	if req.Stream {
		h.handleStreamingRequest(c, apiKey, req, "anthropic")
		return
	}

	// Proxy the request
	resp, err := h.proxyService.ProxyRequest(c.Request.Context(), apiKey, req)
	if err != nil {
		h.handleProxyError(c, err, "anthropic")
		return
	}

	// Convert response to Anthropic format
	finalResp, err := h.protocolConverter.ConvertFromInternalFormat(resp, "anthropic")
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, "anthropic", fmt.Sprintf("Failed to convert response: %v", err))
		return
	}

	c.JSON(http.StatusOK, finalResp)
}

// extractAPIKey extracts API key from headers
func (h *ProxyHandler) extractAPIKey(c *gin.Context, prefix string) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}
	apiKey := strings.TrimPrefix(authHeader, prefix)
	if apiKey == authHeader {
		return ""
	}
	return apiKey
}

// respondError sends an error response in the appropriate format
func (h *ProxyHandler) respondError(c *gin.Context, statusCode int, protocol string, message string) {
	switch protocol {
	case "anthropic":
		c.JSON(statusCode, gin.H{
			"error": gin.H{
				"type":    "api_error",
				"message": message,
			},
		})
	case "gemini":
		c.JSON(statusCode, gin.H{
			"error": gin.H{
				"code":    statusCode,
				"message": message,
			},
		})
	default: // openai
		c.JSON(statusCode, gin.H{
			"error": gin.H{
				"code":    statusCode * 1000 + 1,
				"message": message,
			},
		})
	}
}

// GeminiGenerateContent handles Gemini-compatible generateContent endpoint
func (h *ProxyHandler) GeminiGenerateContent(c *gin.Context) {
	// Extract model from URL path
	// Path format: /v1/models/{model}:generateContent or /v1/models/{model}:streamGenerateContent
	// Gin wildcard gives us: /{model}:generateContent or /{model}:streamGenerateContent
	action := c.Param("action")
	if action == "" || (!strings.Contains(action, ":generateContent") && !strings.Contains(action, ":streamGenerateContent")) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    400,
				"message": "Invalid path format. Expected /v1/models/{model}:generateContent or :streamGenerateContent",
			},
		})
		return
	}

	// Extract model name (remove leading slash and suffix)
	model := strings.TrimPrefix(action, "/")
	isStreaming := strings.HasSuffix(model, ":streamGenerateContent")
	model = strings.TrimSuffix(model, ":generateContent")
	model = strings.TrimSuffix(model, ":streamGenerateContent")

	if model == "" {
		h.respondError(c, http.StatusBadRequest, "gemini", "Model parameter is required")
		return
	}

	// Extract API key from query parameter (Gemini style)
	apiKey := c.Query("key")
	if apiKey == "" {
		apiKey = h.extractAPIKey(c, "Bearer ")
	}

	if apiKey == "" {
		h.respondError(c, http.StatusUnauthorized, "gemini", "API key is required")
		return
	}

	// Parse raw request
	var rawReq map[string]interface{}
	if err := c.ShouldBindJSON(&rawReq); err != nil {
		h.respondError(c, http.StatusBadRequest, "gemini", fmt.Sprintf("Invalid request: %v", err))
		return
	}

	// Convert to internal format
	req, err := h.protocolConverter.ConvertToInternalFormat(rawReq, "gemini")
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "gemini", fmt.Sprintf("Failed to parse request: %v", err))
		return
	}

	// Set model from URL
	req.Model = model

	// Set streaming flag from URL path
	if isStreaming {
		req.Stream = true
	}

	// Handle streaming vs non-streaming
	if req.Stream {
		h.handleStreamingRequest(c, apiKey, req, "gemini")
		return
	}

	// Proxy the request
	resp, err := h.proxyService.ProxyRequest(c.Request.Context(), apiKey, req)
	if err != nil {
		h.handleProxyError(c, err, "gemini")
		return
	}

	// Convert response to Gemini format
	finalResp, err := h.protocolConverter.ConvertFromInternalFormat(resp, "gemini")
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, "gemini", fmt.Sprintf("Failed to convert response: %v", err))
		return
	}

	c.JSON(http.StatusOK, finalResp)
}

func (h *ProxyHandler) handleProxyError(c *gin.Context, err error, format string) {
	if errors.Is(err, service.ErrInvalidAPIKey) {
		switch format {
		case "anthropic":
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"type":    "authentication_error",
					"message": "Invalid API key",
				},
			})
		case "gemini":
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    401,
					"message": "Invalid API key",
				},
			})
		default:
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    401001,
					"message": "Invalid API key",
				},
			})
		}
		return
	}

	if errors.Is(err, service.ErrInsufficientQuota) {
		switch format {
		case "anthropic":
			c.JSON(http.StatusPaymentRequired, gin.H{
				"error": gin.H{
					"type":    "insufficient_quota",
					"message": "Your quota has been exceeded",
				},
			})
		case "gemini":
			c.JSON(http.StatusPaymentRequired, gin.H{
				"error": gin.H{
					"code":    402,
					"message": "Insufficient quota",
				},
			})
		default:
			c.JSON(http.StatusPaymentRequired, gin.H{
				"error": gin.H{
					"code":    402001,
					"message": "Insufficient quota",
				},
			})
		}
		return
	}

	if errors.Is(err, service.ErrNoConfigAvailable) {
		switch format {
		case "anthropic":
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"type":    "not_found_error",
					"message": "No configuration available for the requested model",
				},
			})
		case "gemini":
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    404,
					"message": "Model not found",
				},
			})
		default:
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    404001,
					"message": "Model not found",
				},
			})
		}
		return
	}

	// Generic error
	switch format {
	case "anthropic":
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"type":    "api_error",
				"message": err.Error(),
			},
		})
	case "gemini":
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    500,
				"message": err.Error(),
			},
		})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    500001,
				"message": err.Error(),
			},
		})
	}
}

// handleStreamingRequest handles streaming chat completion requests with protocol conversion
func (h *ProxyHandler) handleStreamingRequest(c *gin.Context, apiKey string, req *adapter.ChatRequest, protocol string) {
	// Get streaming response from proxy service
	config, resp, err := h.proxyService.ProxyStreamRequest(c.Request.Context(), apiKey, req)
	if err != nil {
		h.handleProxyError(c, err, protocol)
		return
	}
	defer resp.Body.Close()

	// Set headers for SSE streaming
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")
	c.Header("X-Accel-Buffering", "no")

	// Avoid unused variable warning
	_ = config

	// Stream with protocol conversion
	if protocol == "anthropic" {
		h.streamAnthropicFormat(c, resp.Body)
	} else if protocol == "gemini" {
		h.streamGeminiFormat(c, resp.Body)
	} else {
		// OpenAI format - direct passthrough
		h.streamOpenAIFormat(c, resp.Body)
	}
}

// streamOpenAIFormat streams OpenAI format SSE
func (h *ProxyHandler) streamOpenAIFormat(c *gin.Context, body io.Reader) {
	c.Stream(func(w io.Writer) bool {
		buf := make([]byte, 4096)
		n, err := body.Read(buf)
		if err != nil {
			return false
		}
		if n > 0 {
			_, writeErr := w.Write(buf[:n])
			if writeErr != nil {
				return false
			}
		}
		return true
	})
}

// streamAnthropicFormat converts OpenAI SSE to Anthropic SSE format
func (h *ProxyHandler) streamAnthropicFormat(c *gin.Context, body io.Reader) {
	c.Stream(func(w io.Writer) bool {
		buf := make([]byte, 4096)
		n, err := body.Read(buf)
		if err != nil {
			return false
		}
		if n > 0 {
			line := string(buf[:n])
			
			if strings.HasPrefix(line, "data: ") {
				dataStr := strings.TrimPrefix(line, "data: ")
				dataStr = strings.TrimSpace(dataStr)
				
				if dataStr == "[DONE]" {
					fmt.Fprintf(w, "event: message_stop\n")
					fmt.Fprintf(w, "data: {\"type\":\"message_stop\"}\n\n")
					return false
				}
				
				var openaiChunk map[string]interface{}
				if err := json.Unmarshal([]byte(dataStr), &openaiChunk); err == nil {
					choices, _ := openaiChunk["choices"].([]interface{})
					if len(choices) > 0 {
						choice, _ := choices[0].(map[string]interface{})
						delta, _ := choice["delta"].(map[string]interface{})
						
						if content, hasContent := delta["content"].(string); hasContent && content != "" {
							anthropicEvent := map[string]interface{}{
								"type":  "content_block_delta",
								"index": 0,
								"delta": map[string]interface{}{
									"type": "text_delta",
									"text": content,
								},
							}
							eventJSON, _ := json.Marshal(anthropicEvent)
							fmt.Fprintf(w, "event: content_block_delta\n")
							fmt.Fprintf(w, "data: %s\n\n", string(eventJSON))
						}
						
						if toolCalls, hasToolCalls := delta["tool_calls"].([]interface{}); hasToolCalls && len(toolCalls) > 0 {
							for _, tc := range toolCalls {
								toolCall, _ := tc.(map[string]interface{})
								function, _ := toolCall["function"].(map[string]interface{})
								
								if name, hasName := function["name"].(string); hasName {
									anthropicEvent := map[string]interface{}{
										"type":  "content_block_start",
										"index": 0,
										"content_block": map[string]interface{}{
											"type": "tool_use",
											"id":   toolCall["id"],
											"name": name,
										},
									}
									eventJSON, _ := json.Marshal(anthropicEvent)
									fmt.Fprintf(w, "event: content_block_start\n")
									fmt.Fprintf(w, "data: %s\n\n", string(eventJSON))
								}
							}
						}
					}
				}
			}
		}
		return true
	})
}

// streamGeminiFormat converts OpenAI SSE to Gemini SSE format
func (h *ProxyHandler) streamGeminiFormat(c *gin.Context, body io.Reader) {
	c.Stream(func(w io.Writer) bool {
		buf := make([]byte, 4096)
		n, err := body.Read(buf)
		if err != nil {
			return false
		}
		if n > 0 {
			line := string(buf[:n])
			
			if strings.HasPrefix(line, "data: ") {
				dataStr := strings.TrimPrefix(line, "data: ")
				dataStr = strings.TrimSpace(dataStr)
				
				if dataStr == "[DONE]" {
					return false
				}
				
				var openaiChunk map[string]interface{}
				if err := json.Unmarshal([]byte(dataStr), &openaiChunk); err == nil {
					choices, _ := openaiChunk["choices"].([]interface{})
					if len(choices) > 0 {
						choice, _ := choices[0].(map[string]interface{})
						delta, _ := choice["delta"].(map[string]interface{})
						
						if content, hasContent := delta["content"].(string); hasContent && content != "" {
							geminiChunk := map[string]interface{}{
								"candidates": []interface{}{
									map[string]interface{}{
										"content": map[string]interface{}{
											"parts": []interface{}{
												map[string]interface{}{
													"text": content,
												},
											},
											"role": "model",
										},
										"index": 0,
									},
								},
							}
							chunkJSON, _ := json.Marshal(geminiChunk)
							fmt.Fprintf(w, "data: %s\n\n", string(chunkJSON))
						}
					}
				}
			}
		}
		return true
	})
}
