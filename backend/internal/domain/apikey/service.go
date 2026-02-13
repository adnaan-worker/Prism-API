package apikey

import (
	"api-aggregator/backend/pkg/crypto"
	"api-aggregator/backend/pkg/errors"
	"api-aggregator/backend/pkg/logger"
	"api-aggregator/backend/pkg/query"
	"context"
)

// Service API密钥服务接口
type Service interface {
	CreateAPIKey(ctx context.Context, userID uint, req *CreateAPIKeyRequest) (*CreateAPIKeyResponse, error)
	GetAPIKeys(ctx context.Context, userID uint, req *GetAPIKeysRequest) (*APIKeyListResponse, error)
	GetAPIKeyByID(ctx context.Context, userID uint, id uint) (*APIKeyResponse, error)
	UpdateAPIKey(ctx context.Context, userID uint, id uint, req *UpdateAPIKeyRequest) error
	DeleteAPIKey(ctx context.Context, userID uint, id uint) error
	ValidateAPIKey(ctx context.Context, key string) (userID uint, apiKeyID uint, err error)
}

// service API密钥服务实现
type service struct {
	repo   Repository
	logger logger.Logger
}

// NewService 创建API密钥服务
func NewService(repo Repository, logger logger.Logger) Service {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

// CreateAPIKey 创建API密钥
func (s *service) CreateAPIKey(ctx context.Context, userID uint, req *CreateAPIKeyRequest) (*CreateAPIKeyResponse, error) {
	// 生成唯一的API密钥
	key, err := crypto.GenerateAPIKey()
	if err != nil {
		s.logger.Error("Failed to generate API key", logger.Error(err))
		return nil, errors.Wrap(err, 500005, "Failed to generate API key")
	}

	// 设置默认速率限制
	rateLimit := req.RateLimit
	if rateLimit == 0 {
		rateLimit = 60
	}

	// 创建API密钥
	apiKey := &APIKey{
		UserID:    userID,
		Key:       key,
		Name:      req.Name,
		IsActive:  true,
		RateLimit: rateLimit,
	}

	if err := s.repo.Create(ctx, apiKey); err != nil {
		s.logger.Error("Failed to create API key",
			logger.Uint("user_id", userID),
			logger.String("name", req.Name),
			logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to create API key")
	}

	s.logger.Info("API key created successfully",
		logger.Uint("user_id", userID),
		logger.Uint("key_id", apiKey.ID),
		logger.String("name", apiKey.Name))

	return &CreateAPIKeyResponse{
		APIKey: apiKey.ToResponse(),
	}, nil
}

// GetAPIKeys 获取API密钥列表
func (s *service) GetAPIKeys(ctx context.Context, userID uint, req *GetAPIKeysRequest) (*APIKeyListResponse, error) {
	// 设置默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}

	// 构建过滤条件
	var filters []query.Filter
	if req.IsActive != nil {
		filters = append(filters, query.Filter{
			Field:    "is_active",
			Operator: "=",
			Value:    *req.IsActive,
		})
	}

	// 构建排序
	sorts := []query.Sort{
		{Field: "created_at", Desc: true},
	}

	// 构建分页
	pagination := &query.Pagination{
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	// 查询API密钥列表
	apiKeys, total, err := s.repo.List(ctx, userID, filters, sorts, pagination)
	if err != nil {
		s.logger.Error("Failed to get API keys", logger.Uint("user_id", userID), logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to get API keys")
	}

	return &APIKeyListResponse{
		Keys:     ToResponseList(apiKeys),
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// GetAPIKeyByID 根据ID获取API密钥
func (s *service) GetAPIKeyByID(ctx context.Context, userID uint, id uint) (*APIKeyResponse, error) {
	apiKey, err := s.repo.FindByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get API key", logger.Uint("key_id", id), logger.Error(err))
		return nil, errors.Wrap(err, 500002, "Failed to get API key")
	}
	if apiKey == nil {
		return nil, errors.ErrAPIKeyNotFound
	}

	// 验证所有权
	if apiKey.UserID != userID {
		return nil, errors.ErrForbidden.WithDetails("You don't have permission to access this API key")
	}

	return apiKey.ToResponse(), nil
}

// UpdateAPIKey 更新API密钥
func (s *service) UpdateAPIKey(ctx context.Context, userID uint, id uint, req *UpdateAPIKeyRequest) error {
	// 查找API密钥
	apiKey, err := s.repo.FindByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get API key", logger.Uint("key_id", id), logger.Error(err))
		return errors.Wrap(err, 500002, "Failed to get API key")
	}
	if apiKey == nil {
		return errors.ErrAPIKeyNotFound
	}

	// 验证所有权
	if apiKey.UserID != userID {
		return errors.ErrForbidden.WithDetails("You don't have permission to update this API key")
	}

	// 更新字段
	if req.Name != "" {
		apiKey.Name = req.Name
	}
	if req.RateLimit > 0 {
		apiKey.RateLimit = req.RateLimit
	}
	if req.IsActive != nil {
		apiKey.IsActive = *req.IsActive
	}

	// 保存更新
	if err := s.repo.Update(ctx, apiKey); err != nil {
		s.logger.Error("Failed to update API key",
			logger.Uint("key_id", id),
			logger.Error(err))
		return errors.Wrap(err, 500002, "Failed to update API key")
	}

	s.logger.Info("API key updated successfully",
		logger.Uint("user_id", userID),
		logger.Uint("key_id", id))

	return nil
}

// DeleteAPIKey 删除API密钥
func (s *service) DeleteAPIKey(ctx context.Context, userID uint, id uint) error {
	// 查找API密钥
	apiKey, err := s.repo.FindByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get API key", logger.Uint("key_id", id), logger.Error(err))
		return errors.Wrap(err, 500002, "Failed to get API key")
	}
	if apiKey == nil {
		return errors.ErrAPIKeyNotFound
	}

	// 验证所有权
	if apiKey.UserID != userID {
		return errors.ErrForbidden.WithDetails("You don't have permission to delete this API key")
	}

	// 删除API密钥
	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to delete API key",
			logger.Uint("key_id", id),
			logger.Error(err))
		return errors.Wrap(err, 500002, "Failed to delete API key")
	}

	s.logger.Info("API key deleted successfully",
		logger.Uint("user_id", userID),
		logger.Uint("key_id", id))

	return nil
}

// ValidateAPIKey 验证API密钥并返回用户ID和API Key ID
func (s *service) ValidateAPIKey(ctx context.Context, key string) (userID uint, apiKeyID uint, err error) {
	apiKey, err := s.repo.FindByKey(ctx, key)
	if err != nil {
		s.logger.Error("Failed to find API key", logger.Error(err))
		return 0, 0, errors.Wrap(err, 500002, "Failed to find API key")
	}
	if apiKey == nil {
		return 0, 0, errors.ErrAPIKeyNotFound
	}

	// 检查密钥是否有效
	if !apiKey.IsValid() {
		return 0, 0, errors.New(403001, "API key is inactive or deleted")
	}

	// 更新最后使用时间（异步，不影响主流程）
	go func() {
		if err := s.repo.UpdateLastUsedAt(context.Background(), apiKey.ID); err != nil {
			s.logger.Warn("Failed to update last used time",
				logger.Uint("key_id", apiKey.ID),
				logger.Error(err))
		}
	}()

	return apiKey.UserID, apiKey.ID, nil
}
