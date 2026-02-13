package utils

import (
	"fmt"
	"runtime"
	"time"
)

var (
	startTime = time.Now()
	// 这些变量在编译时通过 -ldflags 注入
	// 构建命令示例: go build -ldflags "-X 'api-aggregator/backend/pkg/utils.Version=1.0.0' -X 'api-aggregator/backend/pkg/utils.BuildTime=2024-01-01T00:00:00Z' -X 'api-aggregator/backend/pkg/utils.GitCommit=abc123'"
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

// GetVersion 获取应用版本
func GetVersion() string {
	return Version
}

// GetBuildTime 获取构建时间
func GetBuildTime() string {
	return BuildTime
}

// GetGitCommit 获取 Git 提交哈希
func GetGitCommit() string {
	if len(GitCommit) > 7 {
		return GitCommit[:7]
	}
	return GitCommit
}

// GetFullVersion 获取完整版本信息
func GetFullVersion() string {
	return fmt.Sprintf("%s (commit: %s, built: %s)", Version, GetGitCommit(), BuildTime)
}

// GetUptime 获取运行时间
func GetUptime() string {
	duration := time.Since(startTime)
	
	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60
	
	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

// GetGoVersion 获取 Go 版本
func GetGoVersion() string {
	return runtime.Version()
}

// GetNumGoroutines 获取当前 goroutine 数量
func GetNumGoroutines() int {
	return runtime.NumGoroutine()
}

// GetMemStats 获取内存统计
func GetMemStats() *runtime.MemStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return &m
}
