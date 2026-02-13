package settings

import (
	"api-aggregator/backend/internal/domain/user"
	"api-aggregator/backend/pkg/crypto"
	"api-aggregator/backend/pkg/errors"
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
		KeyRuntimeMaxRetries,
		KeyRuntimeTimeout,
		KeyRuntimeEnableLoadBalance,
	}

	settings, err := s.repo.GetMultiple(ctx, keys)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get runtime config")
	}

	config := &RuntimeConfigResponse{
		CacheEnabled:      s.getBool(settings, KeyRuntimeCacheEnabled, true),
		CacheTTL:          s.getInt(settings, KeyRuntimeCacheTTL, 3600),
		MaxRetries:        s.getInt(settings, KeyRuntimeMaxRetries, 3),
		Timeout:           s.getInt(settings, KeyRuntimeTimeout, 30),
		EnableLoadBalance: s.getBool(settings, KeyRuntimeEnableLoadBalance, true),
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
		updates[KeyRuntimeCacheTTL] = strconv.Itoa(*req.CacheTTL)
	}
	if req.MaxRetries != nil {
		updates[KeyRuntimeMaxRetries] = strconv.Itoa(*req.MaxRetries)
	}
	if req.Timeout != nil {
		updates[KeyRuntimeTimeout] = strconv.Itoa(*req.Timeout)
	}
	if req.EnableLoadBalance != nil {
		updates[KeyRuntimeEnableLoadBalance] = strconv.FormatBool(*req.EnableLoadBalance)
	}

	if len(updates) > 0 {
		if err := s.repo.SetMultiple(ctx, updates); err != nil {
			return nil, errors.Wrap(err, "failed to update runtime config")
		}
	}

	return s.GetRuntimeConfig(ctx)
}

// GetSystemConfig 获取系统配置
func (s *service) GetSystemConfig(ctx context.Context) (*SystemConfigResponse, error) {
	keys := []string{
		KeySystemSiteName,
		KeySystemSiteDescription,
		KeySystemAdminEmail,
		KeySystemMaintenanceMode,
	}

	settings, err := s.repo.GetMultiple(ctx, keys)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get system config")
	}

	config := &SystemConfigResponse{
		SiteName:        s.getString(settings, KeySystemSiteName, "API Aggregator"),
		SiteDescription: s.getString(settings, KeySystemSiteDescription, "API Aggregator Platform"),
		AdminEmail:      s.getString(settings, KeySystemAdminEmail, "admin@example.com"),
		MaintenanceMode: s.getBool(settings, KeySystemMaintenanceMode, false),
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
		KeyDefaultQuotaMonthly,
		KeyDefaultQuotaTotal,
	}

	settings, err := s.repo.GetMultiple(ctx, keys)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get default quota")
	}

	quota := &DefaultQuotaResponse{
		Daily:   s.getInt64(settings, KeyDefaultQuotaDaily, 1000),
		Monthly: s.getInt64(settings, KeyDefaultQuotaMonthly, 30000),
		Total:   s.getInt64(settings, KeyDefaultQuotaTotal, 100000),
	}

	return quota, nil
}

// UpdateDefaultQuota 更新默认配额
func (s *service) UpdateDefaultQuota(ctx context.Context, req *UpdateDefaultQuotaRequest) (*DefaultQuotaResponse, error) {
	updates := make(map[string]string)

	if req.Daily != nil {
		updates[KeyDefaultQuotaDaily] = strconv.FormatInt(*req.Daily, 10)
	}
	if req.Monthly != nil {
		updates[KeyDefaultQuotaMonthly] = strconv.FormatInt(*req.Monthly, 10)
	}
	if req.Total != nil {
		updates[KeyDefaultQuotaTotal] = strconv.FormatInt(*req.Total, 10)
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
		KeyDefaultRateLimitPerHour,
		KeyDefaultRateLimitPerDay,
	}

	settings, err := s.repo.GetMultiple(ctx, keys)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get default rate limit")
	}

	rateLimit := &DefaultRateLimitResponse{
		PerMinute: s.getInt(settings, KeyDefaultRateLimitPerMinute, 60),
		PerHour:   s.getInt(settings, KeyDefaultRateLimitPerHour, 3600),
		PerDay:    s.getInt(settings, KeyDefaultRateLimitPerDay, 86400),
	}

	return rateLimit, nil
}

// UpdateDefaultRateLimit 更新默认速率限制
func (s *service) UpdateDefaultRateLimit(ctx context.Context, req *UpdateDefaultRateLimitRequest) (*DefaultRateLimitResponse, error) {
	updates := make(map[string]string)

	if req.PerMinute != nil {
		updates[KeyDefaultRateLimitPerMinute] = strconv.Itoa(*req.PerMinute)
	}
	if req.PerHour != nil {
		updates[KeyDefaultRateLimitPerHour] = strconv.Itoa(*req.PerHour)
	}
	if req.PerDay != nil {
		updates[KeyDefaultRateLimitPerDay] = strconv.Itoa(*req.PerDay)
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
