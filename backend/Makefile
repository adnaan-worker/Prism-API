.PHONY: help build run migrate clean test version

# 默认版本号
VERSION ?= 1.0.0

# 构建信息
BUILD_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -X 'api-aggregator/backend/pkg/utils.Version=$(VERSION)' \
           -X 'api-aggregator/backend/pkg/utils.BuildTime=$(BUILD_TIME)' \
           -X 'api-aggregator/backend/pkg/utils.GitCommit=$(GIT_COMMIT)'

# 输出目录
OUTPUT_DIR := bin

help: ## 显示帮助信息
	@echo "Prism API - 构建命令"
	@echo ""
	@echo "使用方法: make [target]"
	@echo ""
	@echo "可用命令:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## 构建所有二进制文件
	@echo "Building Prism API..."
	@echo "  Version: $(VERSION)"
	@echo "  Build Time: $(BUILD_TIME)"
	@echo "  Git Commit: $(GIT_COMMIT)"
	@echo ""
	@mkdir -p $(OUTPUT_DIR)
	@echo "Building server..."
	@go build -ldflags "$(LDFLAGS)" -o $(OUTPUT_DIR)/server ./cmd/server
	@echo "Building migrate tool..."
	@go build -ldflags "$(LDFLAGS)" -o $(OUTPUT_DIR)/migrate ./scripts/migrate.go
	@echo "Building version tool..."
	@go build -ldflags "$(LDFLAGS)" -o $(OUTPUT_DIR)/version ./cmd/version
	@echo ""
	@echo "✅ Build completed successfully!"
	@echo "Binaries are in $(OUTPUT_DIR)/"

run: ## 运行服务器（开发模式）
	@echo "Starting server in development mode..."
	@go run -ldflags "$(LDFLAGS)" ./cmd/server/main.go

migrate: ## 运行数据库迁移
	@echo "Running database migration..."
	@go run ./scripts/migrate.go

clean: ## 清理构建文件
	@echo "Cleaning build artifacts..."
	@rm -rf $(OUTPUT_DIR)
	@echo "✅ Clean completed!"

test: ## 运行测试
	@echo "Running tests..."
	@go test -v ./...

version: ## 显示版本信息
	@echo "Prism API"
	@echo "========="
	@echo "Version:    $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"

install-deps: ## 安装依赖
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "✅ Dependencies installed!"

dev: migrate run ## 运行迁移并启动开发服务器

.DEFAULT_GOAL := help
