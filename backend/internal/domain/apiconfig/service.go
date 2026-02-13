package apiconfig

import (
	"api-aggregator/backend/pkg/errors"
	"api-aggregator/backend/pkg/logger"
	"api-aggregator/backend/pkg/query"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Service API配置服务接口
type Service interface {
	CreateConfig(ctx context.Context, req *CreateConfigRequest) (*ConfigResponse, error)
	GetConfig(ctx context.Context, id uint) (*ConfigResponse, error)
	GetConfigs(ctx context.Context, req *GetConfigsRequest) (*ConfigListResponse, error)
	GetAllConfigs(ctx context.Context) ([]*ConfigResponse, error)
	GetActiveConfigs(ctx context.Context) ([]*ConfigResponse, error)
	GetConfigsByModel(ctx context.Context, model string) ([]*ConfigResponse, error)
	UpdateConfig(ctx context.Context, id uint, req *UpdateConfigRequest) (*ConfigResponse, error)
	DeleteConfig(ctx context.Context, id uint) error
	ActivateConfig(ctx context.Context, id uint) error
	DeactivateConfig(ctx context.Context, id uint) error
	BatchDeleteConfigs(ctx context.Context, ids []uint) (*BatchOperationResponse, error)
	BatchActivateConfigs(ctx context.Context, ids []uint) (*BatchOperationResponse, error)
	BatchDeactivateConfigs(ctx context.Context, ids []uint) (*BatchOperationResponse, error)
	FetchModels(ctx context.Context, req *FetchModelsRequest) (*FetchModelsResponse, error)
}

// service API配置服务实现
type service struct {
	repo   Repository
	logger logger.Logger
}

// NewService 创建API配置服务
func NewService(repo Repository, logger logger.Logger) Service {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

// CreateConfig 创建配置
func (s *service) CreateConfig(ctx context.Context, req *CreateConfigRequest) (*ConfigResponse, error) {
	// 验证 BaseURL
	baseURL := strings.TrimSpace(req.BaseURL)
	
	// 如果是账号池格式，验证格式
	if strings.HasPrefix(baseURL, "account_pool:") {
		parts := strings.Split(baseURL, ":")
		if len(parts) != 3 {
			return nil, errors.NewValidationError("invalid account pool format", map[string]string{
				"base_url": "account pool format should be: account_pool:provider:pool_id",
			})
		}
		// 验证 pool_id 是数字
		if _, err := strconv.ParseUint(parts[2], 10, 32); err != nil {
			return nil, errors.NewValidationError("invalid pool id", map[string]string{
				"base_url": "pool_id must be a valid number",
			})
		}
	} else {
		// 非账号池格式，验证是否为有效 URL（除了 kiro 类型）
		if req.Type != "kiro" {
			if baseURL == "" {
				return nil, errors.NewValidationError("base_url is required", map[string]string{
					"base_url": "required for openai, anthropic, gemini, and custom types",
				})
			}
			// 简单的 URL 验证
			if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
				return nil, errors.NewValidationError("invalid base_url", map[string]string{
					"base_url": "must be a valid HTTP or HTTPS URL",
				})
			}
		}
		
		// Kiro 类型如果没有 base_url，使用默认值
		if req.Type == "kiro" && baseURL == "" {
			baseURL = "https://q.us-east-1.amazonaws.com"
		}
	}

	// 设置默认值
	priority := req.Priority
	if priority == 0 {
		priority = 100
	}
	weight := req.Weight
	if weight == 0 {
		weight = 1
	}
	timeout := req.Timeout
	if timeout == 0 {
		timeout = 30
	}

	// 创建配置
	config := &APIConfig{
		Name:     req.Name,
		Type:     req.Type,
		BaseURL:  baseURL,
		APIKey:   req.APIKey,
		Models:   req.Models,
		Headers:  req.Headers,
		Metadata: req.Metadata,
		IsActive: true,
		Priority: priority,
		Weight:   weight,
		MaxRPS:   req.MaxRPS,
		Timeout:  timeout,
	}

	if err := s.repo.Create(ctx, config); err != nil {
		s.logger.Error("Failed to create config",
			logger.String("name", req.Name),
			logger.String("type", req.Type),
			logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to create config")
	}

	s.logger.Info("Config created successfully",
		logger.Uint("config_id", config.ID),
		logger.String("name", config.Name),
		logger.String("type", config.Type))

	return config.ToResponse(), nil
}

// GetConfig 获取配置
func (s *service) GetConfig(ctx context.Context, id uint) (*ConfigResponse, error) {
	config, err := s.repo.FindByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get config", logger.Uint("config_id", id), logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to get config")
	}
	if config == nil {
		return nil, errors.ErrAPIConfigNotFound
	}

	return config.ToResponse(), nil
}

// GetConfigs 获取配置列表
func (s *service) GetConfigs(ctx context.Context, req *GetConfigsRequest) (*ConfigListResponse, error) {
	// 设置默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}

	// 构建过滤条件
	var filters []query.Filter
	if req.Type != "" {
		filters = append(filters, query.Filter{
			Field:    "type",
			Operator: "=",
			Value:    req.Type,
		})
	}
	if req.IsActive != nil {
		filters = append(filters, query.Filter{
			Field:    "is_active",
			Operator: "=",
			Value:    *req.IsActive,
		})
	}
	if req.Model != "" {
		// 使用 PostgreSQL JSONB 查询
		filters = append(filters, query.Filter{
			Field:    "models",
			Operator: "@>",
			Value:    `["` + req.Model + `"]`,
		})
	}

	// 构建排序
	sorts := []query.Sort{
		{Field: "priority", Desc: false},
		{Field: "created_at", Desc: true},
	}

	// 构建分页
	pagination := &query.Pagination{
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	// 查询配置列表
	configs, total, err := s.repo.List(ctx, filters, sorts, pagination)
	if err != nil {
		s.logger.Error("Failed to get configs", logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to get configs")
	}

	return &ConfigListResponse{
		Configs:  ToResponseList(configs),
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// GetAllConfigs 获取所有配置
func (s *service) GetAllConfigs(ctx context.Context) ([]*ConfigResponse, error) {
	configs, err := s.repo.FindAll(ctx)
	if err != nil {
		s.logger.Error("Failed to get all configs", logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to get all configs")
	}

	return ToResponseList(configs), nil
}

// GetActiveConfigs 获取所有激活的配置
func (s *service) GetActiveConfigs(ctx context.Context) ([]*ConfigResponse, error) {
	configs, err := s.repo.FindActive(ctx)
	if err != nil {
		s.logger.Error("Failed to get active configs", logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to get active configs")
	}

	return ToResponseList(configs), nil
}

// GetConfigsByModel 根据模型获取配置
func (s *service) GetConfigsByModel(ctx context.Context, model string) ([]*ConfigResponse, error) {
	configs, err := s.repo.FindByModel(ctx, model)
	if err != nil {
		s.logger.Error("Failed to get configs by model",
			logger.String("model", model),
			logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to get configs by model")
	}

	return ToResponseList(configs), nil
}

// UpdateConfig 更新配置
func (s *service) UpdateConfig(ctx context.Context, id uint, req *UpdateConfigRequest) (*ConfigResponse, error) {
	// 查找配置
	config, err := s.repo.FindByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get config", logger.Uint("config_id", id), logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to get config")
	}
	if config == nil {
		return nil, errors.ErrAPIConfigNotFound
	}

	// 更新字段
	if req.Name != "" {
		config.Name = req.Name
	}
	if req.Type != "" {
		config.Type = req.Type
	}
	if req.BaseURL != "" {
		config.BaseURL = req.BaseURL
	}
	if req.APIKey != "" {
		config.APIKey = req.APIKey
	}
	if len(req.Models) > 0 {
		config.Models = req.Models
	}
	if req.Headers != nil {
		config.Headers = req.Headers
	}
	if req.Metadata != nil {
		config.Metadata = req.Metadata
	}
	if req.Priority != nil {
		config.Priority = *req.Priority
	}
	if req.Weight != nil {
		config.Weight = *req.Weight
	}
	if req.MaxRPS != nil {
		config.MaxRPS = *req.MaxRPS
	}
	if req.Timeout != nil {
		config.Timeout = *req.Timeout
	}
	if req.IsActive != nil {
		config.IsActive = *req.IsActive
	}

	// 保存更新
	if err := s.repo.Update(ctx, config); err != nil {
		s.logger.Error("Failed to update config",
			logger.Uint("config_id", id),
			logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to update config")
	}

	s.logger.Info("Config updated successfully",
		logger.Uint("config_id", id),
		logger.String("name", config.Name))

	return config.ToResponse(), nil
}

// DeleteConfig 删除配置
func (s *service) DeleteConfig(ctx context.Context, id uint) error {
	// 检查配置是否存在
	config, err := s.repo.FindByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get config", logger.Uint("config_id", id), logger.Error(err))
		return errors.Wrap(err, 500002, "Failed to get config")
	}
	if config == nil {
		return errors.ErrAPIConfigNotFound
	}

	// 删除配置
	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to delete config",
			logger.Uint("config_id", id),
			logger.Error(err))
		return errors.Wrap(err, 500002, "Failed to delete config")
	}

	s.logger.Info("Config deleted successfully",
		logger.Uint("config_id", id),
		logger.String("name", config.Name))

	return nil
}

// ActivateConfig 激活配置
func (s *service) ActivateConfig(ctx context.Context, id uint) error {
	// 检查配置是否存在
	config, err := s.repo.FindByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get config", logger.Uint("config_id", id), logger.Error(err))
		return errors.Wrap(err, 500002, "Failed to get config")
	}
	if config == nil {
		return errors.ErrAPIConfigNotFound
	}

	// 激活配置
	if err := s.repo.UpdateStatus(ctx, id, true); err != nil {
		s.logger.Error("Failed to activate config",
			logger.Uint("config_id", id),
			logger.Error(err))
		return errors.Wrap(err, 500002, "Failed to activate config")
	}

	s.logger.Info("Config activated successfully",
		logger.Uint("config_id", id),
		logger.String("name", config.Name))

	return nil
}

// DeactivateConfig 停用配置
func (s *service) DeactivateConfig(ctx context.Context, id uint) error {
	// 检查配置是否存在
	config, err := s.repo.FindByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get config", logger.Uint("config_id", id), logger.Error(err))
		return errors.Wrap(err, 500002, "Failed to get config")
	}
	if config == nil {
		return errors.ErrAPIConfigNotFound
	}

	// 停用配置
	if err := s.repo.UpdateStatus(ctx, id, false); err != nil {
		s.logger.Error("Failed to deactivate config",
			logger.Uint("config_id", id),
			logger.Error(err))
		return errors.Wrap(err, 500002, "Failed to deactivate config")
	}

	s.logger.Info("Config deactivated successfully",
		logger.Uint("config_id", id),
		logger.String("name", config.Name))

	return nil
}

// BatchDeleteConfigs 批量删除配置
func (s *service) BatchDeleteConfigs(ctx context.Context, ids []uint) (*BatchOperationResponse, error) {
	if len(ids) == 0 {
		return nil, errors.ErrInvalidParam.WithDetails("IDs array cannot be empty")
	}

	if err := s.repo.BatchDelete(ctx, ids); err != nil {
		s.logger.Error("Failed to batch delete configs",
			logger.Int("count", len(ids)),
			logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to batch delete configs")
	}

	s.logger.Info("Configs deleted successfully", logger.Int("count", len(ids)))

	return &BatchOperationResponse{
		Message: "Configurations deleted successfully",
		Count:   len(ids),
	}, nil
}

// BatchActivateConfigs 批量激活配置
func (s *service) BatchActivateConfigs(ctx context.Context, ids []uint) (*BatchOperationResponse, error) {
	if len(ids) == 0 {
		return nil, errors.ErrInvalidParam.WithDetails("IDs array cannot be empty")
	}

	if err := s.repo.BatchUpdateStatus(ctx, ids, true); err != nil {
		s.logger.Error("Failed to batch activate configs",
			logger.Int("count", len(ids)),
			logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to batch activate configs")
	}

	s.logger.Info("Configs activated successfully", logger.Int("count", len(ids)))

	return &BatchOperationResponse{
		Message: "Configurations activated successfully",
		Count:   len(ids),
	}, nil
}

// BatchDeactivateConfigs 批量停用配置
func (s *service) BatchDeactivateConfigs(ctx context.Context, ids []uint) (*BatchOperationResponse, error) {
	if len(ids) == 0 {
		return nil, errors.ErrInvalidParam.WithDetails("IDs array cannot be empty")
	}

	if err := s.repo.BatchUpdateStatus(ctx, ids, false); err != nil {
		s.logger.Error("Failed to batch deactivate configs",
			logger.Int("count", len(ids)),
			logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to batch deactivate configs")
	}

	s.logger.Info("Configs deactivated successfully", logger.Int("count", len(ids)))

	return &BatchOperationResponse{
		Message: "Configurations deactivated successfully",
		Count:   len(ids),
	}, nil
}

// FetchModels 从提供商动态获取模型列表
func (s *service) FetchModels(ctx context.Context, req *FetchModelsRequest) (*FetchModelsResponse, error) {
	var models []*ModelInfo
	var err error

	// 根据不同的提供商调用相应的 API
	switch req.Provider {
	case "openai":
		models, err = s.fetchOpenAIModels(ctx, req)
	case "anthropic":
		models, err = s.fetchAnthropicModels(ctx, req)
	case "gemini":
		models, err = s.fetchGeminiModels(ctx, req)
	default:
		return nil, errors.NewValidationError("unsupported provider", map[string]string{
			"provider": "must be one of: openai, anthropic, gemini",
		})
	}

	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch models from provider")
	}

	return &FetchModelsResponse{
		Provider: req.Provider,
		Models:   models,
		Count:    len(models),
	}, nil
}

// fetchOpenAIModels 从 OpenAI API 获取模型列表
func (s *service) fetchOpenAIModels(ctx context.Context, req *FetchModelsRequest) ([]*ModelInfo, error) {
	baseURL := req.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	// 创建 HTTP 客户端
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 构建请求
	reqURL := baseURL + "/models"
	httpReq, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+req.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var result struct {
		Data []struct {
			ID      string `json:"id"`
			Object  string `json:"object"`
			Created int64  `json:"created"`
			OwnedBy string `json:"owned_by"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// 转换为 ModelInfo
	models := make([]*ModelInfo, 0, len(result.Data))
	for _, m := range result.Data {
		// 只返回 GPT 和 O1 模型
		if strings.HasPrefix(m.ID, "gpt-") || strings.HasPrefix(m.ID, "o1") {
			capabilities := []string{"chat", "completion"}
			if strings.Contains(m.ID, "vision") || strings.Contains(m.ID, "4o") {
				capabilities = append(capabilities, "vision")
			}
			if strings.HasPrefix(m.ID, "o1") {
				capabilities = append(capabilities, "reasoning")
			}

			models = append(models, &ModelInfo{
				ID:           m.ID,
				Name:         formatModelName(m.ID),
				Provider:     "openai",
				Capabilities: capabilities,
			})
		}
	}

	return models, nil
}

// fetchAnthropicModels 从 Anthropic API 获取模型列表
func (s *service) fetchAnthropicModels(ctx context.Context, req *FetchModelsRequest) ([]*ModelInfo, error) {
	// Anthropic 没有公开的 list models API，返回已知的模型列表
	// 但我们可以通过测试 API 连接来验证 API Key 是否有效
	baseURL := req.BaseURL
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}

	// 验证 API Key（通过发送一个最小的请求）
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	testBody := map[string]interface{}{
		"model":      "claude-3-haiku-20240307",
		"max_tokens": 1,
		"messages": []map[string]string{
			{"role": "user", "content": "Hi"},
		},
	}

	bodyBytes, _ := json.Marshal(testBody)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/v1/messages", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("x-api-key", req.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 如果返回 401，说明 API Key 无效
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("invalid API key")
	}

	// 返回已知的 Claude 模型列表
	return []*ModelInfo{
		{ID: "claude-3-opus-20240229", Name: "Claude 3 Opus", Provider: "anthropic", Capabilities: []string{"chat", "completion", "vision"}},
		{ID: "claude-3-sonnet-20240229", Name: "Claude 3 Sonnet", Provider: "anthropic", Capabilities: []string{"chat", "completion", "vision"}},
		{ID: "claude-3-haiku-20240307", Name: "Claude 3 Haiku", Provider: "anthropic", Capabilities: []string{"chat", "completion", "vision"}},
		{ID: "claude-3-5-sonnet-20241022", Name: "Claude 3.5 Sonnet", Provider: "anthropic", Capabilities: []string{"chat", "completion", "vision"}},
		{ID: "claude-3-5-sonnet-20240620", Name: "Claude 3.5 Sonnet (Legacy)", Provider: "anthropic", Capabilities: []string{"chat", "completion", "vision"}},
		{ID: "claude-3-5-haiku-20241022", Name: "Claude 3.5 Haiku", Provider: "anthropic", Capabilities: []string{"chat", "completion"}},
		{ID: "claude-sonnet-4-5", Name: "Claude Sonnet 4.5", Provider: "anthropic", Capabilities: []string{"chat", "completion", "vision"}},
	}, nil
}

// fetchGeminiModels 从 Google Gemini API 获取模型列表
func (s *service) fetchGeminiModels(ctx context.Context, req *FetchModelsRequest) ([]*ModelInfo, error) {
	baseURL := req.BaseURL
	if baseURL == "" {
		baseURL = "https://generativelanguage.googleapis.com"
	}

	// 创建 HTTP 客户端
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 构建请求 - Gemini 使用 API Key 作为查询参数
	reqURL := fmt.Sprintf("%s/v1beta/models?key=%s", baseURL, req.APIKey)
	httpReq, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Gemini API error (status %d): %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var result struct {
		Models []struct {
			Name                       string   `json:"name"`
			DisplayName                string   `json:"displayName"`
			Description                string   `json:"description"`
			SupportedGenerationMethods []string `json:"supportedGenerationMethods"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// 转换为 ModelInfo
	models := make([]*ModelInfo, 0, len(result.Models))
	for _, m := range result.Models {
		// 提取模型 ID（去掉 "models/" 前缀）
		modelID := strings.TrimPrefix(m.Name, "models/")

		// 只返回 gemini 模型
		if strings.HasPrefix(modelID, "gemini-") {
			capabilities := []string{"chat", "completion"}
			if strings.Contains(modelID, "vision") || strings.Contains(modelID, "pro-vision") {
				capabilities = append(capabilities, "vision")
			}

			models = append(models, &ModelInfo{
				ID:           modelID,
				Name:         m.DisplayName,
				Description:  m.Description,
				Provider:     "gemini",
				Capabilities: capabilities,
			})
		}
	}

	return models, nil
}

// formatModelName 格式化模型名称
func formatModelName(id string) string {
	// 将 gpt-4-turbo 转换为 GPT-4 Turbo
	parts := strings.Split(id, "-")
	formatted := make([]string, len(parts))
	for i, part := range parts {
		if part == "gpt" || part == "o1" {
			formatted[i] = strings.ToUpper(part)
		} else {
			formatted[i] = strings.Title(part)
		}
	}
	return strings.Join(formatted, " ")
}
