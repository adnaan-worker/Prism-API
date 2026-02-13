package settings

import (
	"api-aggregator/backend/internal/domain/user"
	"api-aggregator/backend/pkg/crypto"
	"api-aggregator/backend/pkg/errors"
	"api-aggregator/backend/pkg/utils"
	"context"
	"strconv"
)

// Service 设置服务接口
type Service interface {
	GetRuntimeConfig(ctx context.Context) (*RuntimeConfigResponse, error)
	UpdateRuntimeConfig(ctx context.Context, req *UpdateRuntimeConfigRequest) (*RuntimeConfigResponse, error)
	GetSystemConfig(ctx context.Context) (*SystemConfigResponse, error)
	UpdatePassword(ctx context.Context, userID uint, req *UpdatePasswordRequest) error
	GetDefaultQuota(ctx context.Context) (*DefaultQuotaResponse, error)
	UpdateDefaultQuota(ctx context.Context, req *UpdateDefaultQuotaRequest) (*DefaultQuotaResponse, error)
	GetDefaultRateLimit(ctx context.Context) (*DefaultRateLimitResponse, error)
	UpdateDefaultRateLimit(ctx context.Context, req *UpdateDefaultRateLimitRequest) (*DefaultRateLimitResponse, error)
}

type service struct {
	repo     Repository
	userRepo user.Repository
}

// NewService 创建设置服务实例
func NewService(repo Repository, userRepo user.Repository) Service {
	return &service{
		repo:     repo,
		userRepo: userRepo,
	}
}

// GetRuntimeConfig 获取运行时配置
func (s *service) GetRuntimeConfig(ctx context.Context) (*RuntimeConfigResponse, error) {
	keys := []string{
		KeyRuntimeCacheEnabled,
		KeyRuntimeCacheTTL,
		KeyRuntimeSemanticCacheEnabled,
		KeyRuntimeSemanticThreshold,
		KeyRuntimeEmbeddingEnabled,
	}

	settings, err := s.repo.GetMultiple(ctx, keys)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get runtime config")
	}

	// 将秒数转换为时间格式字符串
	cacheTTLSeconds := s.getInt(settings, KeyRuntimeCacheTTL, 3600)
	cacheTTL := utils.FormatDuration(cacheTTLSeconds)

	config := &RuntimeConfigResponse{
		CacheEnabled:         s.getBool(settings, KeyRuntimeCacheEnabled, true),
		CacheTTL:             cacheTTL,
		SemanticCacheEnabled: s.getBool(settings, KeyRuntimeSemanticCacheEnabled, false),
		SemanticThreshold:    s.getFloat(settings, KeyRuntimeSemanticThreshold, 0.85),
		EmbeddingEnabled:     s.getBool(settings, KeyRuntimeEmbeddingEnabled, false),
	}

	return config, nil
}

// UpdateRuntimeConfig 更新运行时配置
func (s *service) UpdateRuntimeConfig(ctx context.Context, req *UpdateRuntimeConfigRequest) (*RuntimeConfigResponse, error) {
	updates := make(map[string]string)

	if req.CacheEnabled != nil {
		updates[KeyRuntimeCacheEnabled] = strconv.FormatBool(*req.CacheEnabled)
	}
	if req.CacheTTL != nil {
		// 将时间格式字符串转换为秒数
		seconds, err := utils.ParseDuration(*req.CacheTTL)
		if err != nil {
			return nil, errors.NewValidationError("invalid cache_ttl format", map[string]string{
				"cache_ttl": "must be in format like '24h', '1h30m', '30m'",
			})
		}
		updates[KeyRuntimeCacheTTL] = strconv.Itoa(seconds)
	}
	if req.SemanticCacheEnabled != nil {
		updates[KeyRuntimeSemanticCacheEnabled] = strconv.FormatBool(*req.SemanticCacheEnabled)
	}
	if req.SemanticThreshold != nil {
		if *req.SemanticThreshold < 0 || *req.SemanticThreshold > 1 {
			return nil, errors.NewValidationError("invalid semantic_threshold", map[string]string{
				"semantic_threshold": "must be between 0.0 and 1.0",
			})
		}
		updates[KeyRuntimeSemanticThreshold] = strconv.FormatFloat(*req.SemanticThreshold, 'f', 2, 64)
	}
	if req.EmbeddingEnabled != nil {
		updates[KeyRuntimeEmbeddingEnabled] = strconv.FormatBool(*req.EmbeddingEnabled)
	}

	if len(updates) > 0 {
		if err := s.repo.SetMultiple(ctx, updates); err != nil {
			return nil, errors.Wrap(err, "failed to update runtime config")
		}
	}

	return s.GetRuntimeConfig(ctx)
}

// GetSystemConfig 获取系统运行信息
func (s *service) GetSystemConfig(ctx context.Context) (*SystemConfigResponse, error) {
	keys := []string{
		KeyRuntimeCacheEnabled,
		KeyRuntimeCacheTTL,
		KeyRuntimeSemanticCacheEnabled,
		KeyRuntimeSemanticThreshold,
		KeyRuntimeEmbeddingEnabled,
		KeyDefaultRateLimitPerMinute,
	}

	settings, err := s.repo.GetMultiple(ctx, keys)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get system config")
	}

	// 将秒数转换为时间格式字符串
	cacheTTLSeconds := s.getInt(settings, KeyRuntimeCacheTTL, 3600)
	cacheTTL := utils.FormatDuration(cacheTTLSeconds)

	// 获取速率限制配置
	rateLimitPerMinute := s.getInt(settings, KeyDefaultRateLimitPerMinute, 60)

	config := &SystemConfigResponse{
		// 缓存配置
		CacheEnabled:         s.getBool(settings, KeyRuntimeCacheEnabled, true),
		CacheTTL:             cacheTTL,
		SemanticCacheEnabled: s.getBool(settings, KeyRuntimeSemanticCacheEnabled, false),
		SemanticThreshold:    s.getFloat(settings, KeyRuntimeSemanticThreshold, 0.85),
		EmbeddingEnabled:     s.getBool(settings, KeyRuntimeEmbeddingEnabled, false),
		// 速率限制
		RateLimitEnabled:  true,
		RateLimitRequests: rateLimitPerMinute,
		RateLimitWindow:   "1m",
		// 服务信息
		Version:   utils.GetVersion(),
		Uptime:    utils.GetUptime(),
		GoVersion: utils.GetGoVersion(),
	}

	return config, nil
}

// UpdatePassword 修改密码
func (s *service) UpdatePassword(ctx context.Context, userID uint, req *UpdatePasswordRequest) error {
	// 获取用户
	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return errors.NewNotFoundError("user not found")
	}

	// 验证旧密码
	if !crypto.CheckPassword(req.OldPassword, u.PasswordHash) {
		return errors.NewValidationError("invalid old password", map[string]string{
			"old_password": "incorrect password",
		})
	}

	// 加密新密码
	hashedPassword, err := crypto.HashPassword(req.NewPassword)
	if err != nil {
		return errors.Wrap(err, "failed to hash password")
	}

	// 更新密码
	u.PasswordHash = hashedPassword
	if err := s.userRepo.Update(ctx, u); err != nil {
		return errors.Wrap(err, "failed to update password")
	}

	return nil
}

// GetDefaultQuota 获取默认配额
func (s *service) GetDefaultQuota(ctx context.Context) (*DefaultQuotaResponse, error) {
	keys := []string{
		KeyDefaultQuotaDaily,
	}

	settings, err := s.repo.GetMultiple(ctx, keys)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get default quota")
	}

	quota := &DefaultQuotaResponse{
		DefaultQuota: s.getInt64(settings, KeyDefaultQuotaDaily, 100000),
	}

	return quota, nil
}

// UpdateDefaultQuota 更新默认配额
func (s *service) UpdateDefaultQuota(ctx context.Context, req *UpdateDefaultQuotaRequest) (*DefaultQuotaResponse, error) {
	updates := make(map[string]string)

	if req.DefaultQuota != nil {
		updates[KeyDefaultQuotaDaily] = strconv.FormatInt(*req.DefaultQuota, 10)
	}

	if len(updates) > 0 {
		if err := s.repo.SetMultiple(ctx, updates); err != nil {
			return nil, errors.Wrap(err, "failed to update default quota")
		}
	}

	return s.GetDefaultQuota(ctx)
}

// GetDefaultRateLimit 获取默认速率限制
func (s *service) GetDefaultRateLimit(ctx context.Context) (*DefaultRateLimitResponse, error) {
	keys := []string{
		KeyDefaultRateLimitPerMinute,
		KeyDefaultRateLimitPerDay,
	}

	settings, err := s.repo.GetMultiple(ctx, keys)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get default rate limit")
	}

	rateLimit := &DefaultRateLimitResponse{
		RequestsPerMinute: s.getInt(settings, KeyDefaultRateLimitPerMinute, 60),
		RequestsPerDay:    s.getInt(settings, KeyDefaultRateLimitPerDay, 10000),
	}

	return rateLimit, nil
}

// UpdateDefaultRateLimit 更新默认速率限制
func (s *service) UpdateDefaultRateLimit(ctx context.Context, req *UpdateDefaultRateLimitRequest) (*DefaultRateLimitResponse, error) {
	updates := make(map[string]string)

	if req.RequestsPerMinute != nil {
		updates[KeyDefaultRateLimitPerMinute] = strconv.Itoa(*req.RequestsPerMinute)
	}
	if req.RequestsPerDay != nil {
		updates[KeyDefaultRateLimitPerDay] = strconv.Itoa(*req.RequestsPerDay)
	}

	if len(updates) > 0 {
		if err := s.repo.SetMultiple(ctx, updates); err != nil {
			return nil, errors.Wrap(err, "failed to update default rate limit")
		}
	}

	return s.GetDefaultRateLimit(ctx)
}

// 辅助方法

func (s *service) getString(settings map[string]*Setting, key, defaultValue string) string {
	if setting, ok := settings[key]; ok {
		return setting.Value
	}
	return defaultValue
}

func (s *service) getInt(settings map[string]*Setting, key string, defaultValue int) int {
	if setting, ok := settings[key]; ok {
		if val, err := strconv.Atoi(setting.Value); err == nil {
			return val
		}
	}
	return defaultValue
}

func (s *service) getInt64(settings map[string]*Setting, key string, defaultValue int64) int64 {
	if setting, ok := settings[key]; ok {
		if val, err := strconv.ParseInt(setting.Value, 10, 64); err == nil {
			return val
		}
	}
	return defaultValue
}

func (s *service) getBool(settings map[string]*Setting, key string, defaultValue bool) bool {
	if setting, ok := settings[key]; ok {
		if val, err := strconv.ParseBool(setting.Value); err == nil {
			return val
		}
	}
	return defaultValue
}

func (s *service) getFloat(settings map[string]*Setting, key string, defaultValue float64) float64 {
	if setting, ok := settings[key]; ok {
		if val, err := strconv.ParseFloat(setting.Value, 64); err == nil {
			return val
		}
	}
	return defaultValue
}
