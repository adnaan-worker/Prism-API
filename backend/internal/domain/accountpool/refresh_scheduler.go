package accountpool

import (
	"api-aggregator/backend/pkg/logger"
	"context"
	"fmt"
	"time"
)

// RefreshScheduler Token 刷新调度器
type RefreshScheduler struct {
	repo           Repository
	refreshService *KiroRefreshService
	interval       time.Duration
	stopCh         chan struct{}
	logger         logger.Logger
}

// NewRefreshScheduler 创建刷新调度器
func NewRefreshScheduler(
	repo Repository,
	refreshService *KiroRefreshService,
	interval time.Duration,
	logger logger.Logger,
) *RefreshScheduler {
	return &RefreshScheduler{
		repo:           repo,
		refreshService: refreshService,
		interval:       interval,
		stopCh:         make(chan struct{}),
		logger:         logger,
	}
}

// Start 启动定时刷新任务
func (s *RefreshScheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	
	s.logger.Info("Token refresh scheduler started", logger.Duration("interval", s.interval))
	
	for {
		select {
		case <-ticker.C:
			s.refreshExpiredTokens(ctx)
		case <-s.stopCh:
			s.logger.Info("Token refresh scheduler stopped")
			return
		case <-ctx.Done():
			s.logger.Info("Token refresh scheduler context cancelled")
			return
		}
	}
}

// Stop 停止定时刷新任务
func (s *RefreshScheduler) Stop() {
	close(s.stopCh)
}

// refreshExpiredTokens 刷新即将过期的 token
func (s *RefreshScheduler) refreshExpiredTokens(ctx context.Context) {
	// 查找即将过期的 Kiro 凭据（30分钟内过期）
	threshold := time.Now().Add(30 * time.Minute)
	
	creds, err := s.repo.FindExpiringCredentials(ctx, "kiro", threshold)
	if err != nil {
		s.logger.Error("Failed to find expiring credentials", logger.Error(err))
		return
	}
	
	if len(creds) == 0 {
		return
	}
	
	s.logger.Info("Found expiring credentials", logger.Int("count", len(creds)))
	
	// 刷新每个凭据
	for _, cred := range creds {
		if err := s.refreshService.RefreshKiroToken(ctx, cred); err != nil {
			s.logger.Error("Failed to refresh token",
				logger.Uint("credential_id", cred.ID),
				logger.Error(err))
			
			// 标记为不健康
			cred.UpdateHealthStatus(HealthStatusUnhealthy)
			cred.LastError = fmt.Sprintf("auto refresh failed: %v", err)
		} else {
			s.logger.Info("Token refreshed successfully",
				logger.Uint("credential_id", cred.ID))
			
			// 标记为健康
			cred.UpdateHealthStatus(HealthStatusHealthy)
			cred.LastError = ""
		}
		
		// 保存更新
		s.repo.UpdateCredential(ctx, cred)
	}
}
