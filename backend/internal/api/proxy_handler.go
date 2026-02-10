package api

import (
	"api-aggregator/backend/internal/adapter"
	"api-aggregator/backend/internal/service"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type ProxyHandler struct {
	proxyService *service.ProxyService
}

func NewProxyHandler(proxyService *service.ProxyService) *ProxyHandler {
	return &ProxyHandler{
		proxyService: proxyService,
	}
}

// ChatCompletions handles OpenAI-compatible chat completions
func (h *ProxyHandler) ChatCompletions(c *gin.Context) {
	// Extract API key from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    401001,
				"message": "Unauthorized",
				"details": "Missing Authorization header",
			},
		})
		return
	}

	// Extract bearer token
	apiKey := strings.TrimPrefix(authHeader, "Bearer ")
	if apiKey == authHeader {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    401001,
				"message": "Unauthorized",
				"details": "Invalid Authorization header format",
			},
		})
		return
	}

	// Parse request - bind directly to adapter.ChatRequest
	var req adapter.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    400001,
				"message": "Invalid request",
				"details": err.Error(),
			},
		})
		return
	}

	// Handle streaming vs non-streaming requests
	if req.Stream {
		h.handleStreamingRequest(c, apiKey, &req)
		return
	}

	// Proxy the request (non-streaming)
	resp, err := h.proxyService.ProxyRequest(c.Request.Context(), apiKey, &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidAPIKey) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    401001,
					"message": "Unauthorized",
					"details": "Invalid API key",
				},
			})
			return
		}
		if errors.Is(err, service.ErrInsufficientQuota) {
			c.JSON(http.StatusPaymentRequired, gin.H{
				"error": gin.H{
					"code":    402001,
					"message": "Insufficient quota",
					"details": "Your quota has been exceeded",
				},
			})
			return
		}
		if errors.Is(err, service.ErrNoConfigAvailable) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    404001,
					"message": "Model not found",
					"details": "No configuration available for the requested model",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    500001,
				"message": "Internal server error",
				"details": err.Error(),
			},
		})
		return
	}

	// Return response directly without conversion (already in OpenAI format)
	c.JSON(http.StatusOK, resp)
}

// AnthropicMessages handles Anthropic-compatible messages endpoint
func (h *ProxyHandler) AnthropicMessages(c *gin.Context) {
	// Extract API key from x-api-key header (Anthropic style)
	apiKey := c.GetHeader("x-api-key")
	if apiKey == "" {
		// Also try Authorization header as fallback
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			apiKey = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	if apiKey == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"type":    "authentication_error",
				"message": "Missing API key",
			},
		})
		return
	}

	// Parse request directly to adapter.ChatRequest
	var req adapter.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"type":    "invalid_request_error",
				"message": err.Error(),
			},
		})
		return
	}

	// Handle streaming vs non-streaming requests
	if req.Stream {
		h.handleStreamingRequestAnthropic(c, apiKey, &req)
		return
	}

	// Proxy the request (non-streaming)
	resp, err := h.proxyService.ProxyRequest(c.Request.Context(), apiKey, &req)
	if err != nil {
		h.handleProxyError(c, err, "anthropic")
		return
	}

	// Return response directly (adapter already handles format conversion)
	c.JSON(http.StatusOK, resp)
}

// GeminiGenerateContent handles Gemini-compatible generateContent endpoint
func (h *ProxyHandler) GeminiGenerateContent(c *gin.Context) {
	// Extract model from URL path
	// Path format: /v1/models/{model}:generateContent
	// Gin wildcard gives us: /{model}:generateContent
	action := c.Param("action")
	if action == "" || !strings.Contains(action, ":generateContent") {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    400,
				"message": "Invalid path format. Expected /v1/models/{model}:generateContent",
			},
		})
		return
	}

	// Extract model name (remove leading slash and :generateContent suffix)
	model := strings.TrimPrefix(action, "/")
	model = strings.TrimSuffix(model, ":generateContent")

	if model == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    400,
				"message": "Model parameter is required",
			},
		})
		return
	}

	// Extract API key from query parameter (Gemini style)
	apiKey := c.Query("key")
	if apiKey == "" {
		// Also try Authorization header as fallback
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			apiKey = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	if apiKey == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    401,
				"message": "API key is required",
			},
		})
		return
	}

	// Parse request directly to adapter.ChatRequest
	var req adapter.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    400,
				"message": err.Error(),
			},
		})
		return
	}

	// Set model from URL
	req.Model = model

	// Proxy the request
	resp, err := h.proxyService.ProxyRequest(c.Request.Context(), apiKey, &req)
	if err != nil {
		h.handleProxyError(c, err, "gemini")
		return
	}

	// Return response directly (adapter already handles format conversion)
	c.JSON(http.StatusOK, resp)
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

// handleStreamingRequest handles streaming chat completion requests
func (h *ProxyHandler) handleStreamingRequest(c *gin.Context, apiKey string, req *adapter.ChatRequest) {
	// Get streaming response from proxy service
	config, resp, err := h.proxyService.ProxyStreamRequest(c.Request.Context(), apiKey, req)
	if err != nil {
		h.handleProxyError(c, err, "openai")
		return
	}
	defer resp.Body.Close()

	// Set headers for SSE streaming
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")
	c.Header("X-Accel-Buffering", "no")

	// Log the streaming request start
	fmt.Printf("Starting stream for model %s using config %d\n", req.Model, config.ID)

	// Stream the response directly to client
	c.Stream(func(w io.Writer) bool {
		buf := make([]byte, 4096)
		n, err := resp.Body.Read(buf)
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Stream read error: %v\n", err)
			}
			return false
		}
		if n > 0 {
			_, writeErr := w.Write(buf[:n])
			if writeErr != nil {
				fmt.Printf("Stream write error: %v\n", writeErr)
				return false
			}
		}
		return true
	})
}

// handleStreamingRequestAnthropic handles streaming requests in Anthropic format
func (h *ProxyHandler) handleStreamingRequestAnthropic(c *gin.Context, apiKey string, req *adapter.ChatRequest) {
	// Get streaming response from proxy service
	config, resp, err := h.proxyService.ProxyStreamRequest(c.Request.Context(), apiKey, req)
	if err != nil {
		h.handleProxyError(c, err, "anthropic")
		return
	}
	defer resp.Body.Close()

	// Set headers for SSE streaming
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	// Log the streaming request start
	fmt.Printf("Starting Anthropic stream for model %s using config %d\n", req.Model, config.ID)

	// Stream the response directly to client
	c.Stream(func(w io.Writer) bool {
		buf := make([]byte, 4096)
		n, err := resp.Body.Read(buf)
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Stream read error: %v\n", err)
			}
			return false
		}
		if n > 0 {
			_, writeErr := w.Write(buf[:n])
			if writeErr != nil {
				fmt.Printf("Stream write error: %v\n", writeErr)
				return false
			}
		}
		return true
	})
}
