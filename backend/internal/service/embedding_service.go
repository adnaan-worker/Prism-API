package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"
)

// EmbeddingHTTPService 通过 HTTP 调用 Python 向量服务
type EmbeddingHTTPService struct {
	baseURL    string
	httpClient *http.Client
}

// NewEmbeddingHTTPService 创建 HTTP 嵌入服务实例
func NewEmbeddingHTTPService(baseURL string, timeout time.Duration) *EmbeddingHTTPService {
	if baseURL == "" {
		baseURL = "http://localhost:8765"
	}
	
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	
	return &EmbeddingHTTPService{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// GetEmbedding 获取文本的向量表示
func (s *EmbeddingHTTPService) GetEmbedding(ctx context.Context, text string) ([]float64, error) {
	reqBody := map[string]interface{}{
		"text": text,
	}
	
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/embed", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call embedding service: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("embedding service returned status %d: %s", resp.StatusCode, string(body))
	}
	
	var result struct {
		Embedding []float64 `json:"embedding"`
		Dimension int       `json:"dimension"`
		Error     string    `json:"error"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	if result.Error != "" {
		return nil, fmt.Errorf("embedding error: %s", result.Error)
	}
	
	return result.Embedding, nil
}

// BatchGetEmbeddings 批量获取向量
func (s *EmbeddingHTTPService) BatchGetEmbeddings(ctx context.Context, texts []string) ([][]float64, error) {
	reqBody := map[string]interface{}{
		"texts": texts,
	}
	
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/embed/batch", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call embedding service: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("embedding service returned status %d: %s", resp.StatusCode, string(body))
	}
	
	var result struct {
		Embeddings [][]float64 `json:"embeddings"`
		Count      int         `json:"count"`
		Dimension  int         `json:"dimension"`
		Error      string      `json:"error"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	if result.Error != "" {
		return nil, fmt.Errorf("embedding error: %s", result.Error)
	}
	
	return result.Embeddings, nil
}

// CosineSimilarity 计算两个向量的余弦相似度
func (s *EmbeddingHTTPService) CosineSimilarity(vec1, vec2 []float64) float64 {
	if len(vec1) != len(vec2) {
		return 0
	}

	var dotProduct, norm1, norm2 float64
	for i := range vec1 {
		dotProduct += vec1[i] * vec2[i]
		norm1 += vec1[i] * vec1[i]
		norm2 += vec2[i] * vec2[i]
	}

	if norm1 == 0 || norm2 == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(norm1) * math.Sqrt(norm2))
}

// HealthCheck 检查服务是否可用
func (s *EmbeddingHTTPService) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", s.baseURL+"/health", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call health endpoint: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status %d", resp.StatusCode)
	}
	
	return nil
}
