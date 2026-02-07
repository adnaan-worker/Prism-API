package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type ProviderHandler struct{}

func NewProviderHandler() *ProviderHandler {
	return &ProviderHandler{}
}

// FetchModelsRequest represents the request to fetch models from a provider
type FetchModelsRequest struct {
	Type    string `json:"type" binding:"required"`
	BaseURL string `json:"base_url" binding:"required"`
	APIKey  string `json:"api_key" binding:"required"`
}

// FetchModels fetches available models from the provider's API
func (h *ProviderHandler) FetchModels(c *gin.Context) {
	var req FetchModelsRequest
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

	var models []string
	var err error

	switch strings.ToLower(req.Type) {
	case "openai":
		models, err = h.fetchOpenAIModels(req.BaseURL, req.APIKey)
	case "anthropic":
		models, err = h.fetchAnthropicModels(req.BaseURL, req.APIKey)
	case "gemini":
		models, err = h.fetchGeminiModels(req.BaseURL, req.APIKey)
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    400002,
				"message": "Unsupported provider type",
				"details": fmt.Sprintf("Provider type '%s' is not supported", req.Type),
			},
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{
				"code":    502001,
				"message": "Failed to fetch models from provider",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"models": models,
		"total":  len(models),
	})
}

// fetchOpenAIModels fetches models from OpenAI API
func (h *ProviderHandler) fetchOpenAIModels(baseURL, apiKey string) ([]string, error) {
	url := strings.TrimSuffix(baseURL, "/") + "/models"

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	models := make([]string, 0, len(result.Data))
	for _, model := range result.Data {
		// Filter: include chat models and exclude embeddings, audio, etc.
		modelID := strings.ToLower(model.ID)

		// Exclude non-chat models
		if strings.Contains(modelID, "embed") ||
			strings.Contains(modelID, "whisper") ||
			strings.Contains(modelID, "tts") ||
			strings.Contains(modelID, "dall-e") ||
			strings.Contains(modelID, "babbage") ||
			strings.Contains(modelID, "davinci") ||
			strings.Contains(modelID, "curie") ||
			strings.Contains(modelID, "ada") {
			continue
		}

		models = append(models, model.ID)
	}

	return models, nil
}

// fetchAnthropicModels returns common Anthropic models
func (h *ProviderHandler) fetchAnthropicModels(baseURL, apiKey string) ([]string, error) {
	// Anthropic doesn't have a models list endpoint, return common models
	return []string{
		"claude-3-5-sonnet-20241022",
		"claude-3-5-haiku-20241022",
		"claude-3-opus-20240229",
		"claude-3-sonnet-20240229",
		"claude-3-haiku-20240307",
		"claude-2.1",
		"claude-2.0",
		"claude-instant-1.2",
	}, nil
}

// fetchGeminiModels returns common Gemini models
func (h *ProviderHandler) fetchGeminiModels(baseURL, apiKey string) ([]string, error) {
	// Gemini models list
	return []string{
		"gemini-2.0-flash-exp",
		"gemini-1.5-pro",
		"gemini-1.5-flash",
		"gemini-1.5-flash-8b",
		"gemini-pro",
		"gemini-pro-vision",
	}, nil
}
