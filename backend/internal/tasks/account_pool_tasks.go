package tasks

import (
	"api-aggregator/backend/internal/accountpool"
	"api-aggregator/backend/internal/repository"
	"context"
	"fmt"
	"log"
	"time"
)

// AccountPoolTasks 账号池后台任务管理器
type AccountPoolTasks struct {
	credRepo *repository.AccountCredentialRepository
	stopChan chan struct{}
}

// NewAccountPoolTasks 创建任务管理器
func NewAccountPoolTasks(credRepo *repository.AccountCredentialRepository) *AccountPoolTasks {
	return &AccountPoolTasks{
		credRepo: credRepo,
		stopChan: make(chan struct{}),
	}
}

// Start 启动后台任务
func (t *AccountPoolTasks) Start() {
	log.Println("Starting account pool tasks...")
	
	// 健康检查任务（每 5 分钟）
	go t.runTask("health-check", 5*time.Minute, t.healthCheck)
	
	// 令牌刷新任务（每 30 分钟）
	go t.runTask("token-refresh", 30*time.Minute, t.refreshTokens)
	
	log.Println("Account pool tasks started")
}

// Stop 停止后台任务
func (t *AccountPoolTasks) Stop() {
	log.Println("Stopping account pool tasks...")
	close(t.stopChan)
}

// runTask 运行定时任务
func (t *AccountPoolTasks) runTask(name string, interval time.Duration, fn func()) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fn()
		case <-t.stopChan:
			log.Printf("Task %s stopped", name)
			return
		}
	}
}

// healthCheck 健康检查
func (t *AccountPoolTasks) healthCheck() {
	ctx := context.Background()
	log.Println("Running health check...")

	creds, err := t.credRepo.FindByStatus(ctx, "active")
	if err != nil {
		log.Printf("Failed to get credentials: %v", err)
		return
	}

	checked, healthy, unhealthy := 0, 0, 0

	for _, cred := range creds {
		provider, err := accountpool.Get(cred.Provider)
		if err != nil {
			continue
		}

		checked++
		if err := provider.CheckHealth(ctx, cred); err != nil {
			log.Printf("Credential %d unhealthy: %v", cred.ID, err)
			cred.IsActive = false
			cred.ErrorMessage = err.Error()
			t.credRepo.Update(ctx, cred)
			unhealthy++
		} else {
			if cred.ErrorMessage != "" {
				cred.ErrorMessage = ""
				t.credRepo.Update(ctx, cred)
			}
			healthy++
		}
	}

	log.Printf("Health check complete: %d checked, %d healthy, %d unhealthy", checked, healthy, unhealthy)
}

// refreshTokens 刷新令牌
func (t *AccountPoolTasks) refreshTokens() {
	ctx := context.Background()
	log.Println("Running token refresh...")

	creds, err := t.credRepo.FindByStatus(ctx, "active")
	if err != nil {
		log.Printf("Failed to get credentials: %v", err)
		return
	}

	refreshed, failed := 0, 0

	for _, cred := range creds {
		// 检查是否需要刷新（1小时内过期）
		if cred.ExpiresAt != nil && time.Until(*cred.ExpiresAt) > time.Hour {
			continue
		}

		provider, err := accountpool.Get(cred.Provider)
		if err != nil {
			continue
		}

		log.Printf("Refreshing credential %d", cred.ID)
		if err := provider.RefreshToken(ctx, cred); err != nil {
			log.Printf("Failed to refresh credential %d: %v", cred.ID, err)
			cred.IsActive = false
			cred.ErrorMessage = fmt.Sprintf("Refresh failed: %v", err)
			t.credRepo.Update(ctx, cred)
			failed++
		} else {
			t.credRepo.Update(ctx, cred)
			refreshed++
			log.Printf("Refreshed credential %d", cred.ID)
		}
	}

	log.Printf("Token refresh complete: %d refreshed, %d failed", refreshed, failed)
}
