#!/bin/bash

# 构建脚本 - 自动注入版本信息

set -e

# 获取版本信息
VERSION=${VERSION:-"1.0.0"}
BUILD_TIME=$(date -u '+%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT=$(git rev-parse HEAD 2>/dev/null || echo "unknown")

# 输出目录
OUTPUT_DIR="bin"
mkdir -p $OUTPUT_DIR

# 构建标志
LDFLAGS="-X 'api-aggregator/backend/pkg/utils.Version=${VERSION}' \
         -X 'api-aggregator/backend/pkg/utils.BuildTime=${BUILD_TIME}' \
         -X 'api-aggregator/backend/pkg/utils.GitCommit=${GIT_COMMIT}'"

echo "Building Prism API..."
echo "  Version: ${VERSION}"
echo "  Build Time: ${BUILD_TIME}"
echo "  Git Commit: ${GIT_COMMIT}"
echo ""

# 构建服务器
echo "Building server..."
go build -ldflags "${LDFLAGS}" -o ${OUTPUT_DIR}/server ./cmd/server

# 构建迁移工具
echo "Building migrate tool..."
go build -ldflags "${LDFLAGS}" -o ${OUTPUT_DIR}/migrate ./scripts/migrate.go

echo ""
echo "Build completed successfully!"
echo "Binaries are in ${OUTPUT_DIR}/"
