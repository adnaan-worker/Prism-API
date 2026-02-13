# 构建指南

## 快速开始

### 使用 Makefile（推荐 - Linux/Mac）

```bash
# 查看所有可用命令
make help

# 构建所有二进制文件
make build

# 运行数据库迁移
make migrate

# 启动开发服务器
make run

# 运行迁移并启动服务器
make dev

# 清理构建文件
make clean

# 查看版本信息
make version
```

### 使用构建脚本

#### Windows
```cmd
cd backend\scripts
build.bat
```

#### Linux/Mac
```bash
cd backend/scripts
chmod +x build.sh
./build.sh
```

### 手动构建

```bash
# 设置版本号
export VERSION=1.0.0

# 获取构建信息
BUILD_TIME=$(date -u '+%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT=$(git rev-parse HEAD)

# 构建服务器
go build -ldflags "\
  -X 'api-aggregator/backend/pkg/utils.Version=${VERSION}' \
  -X 'api-aggregator/backend/pkg/utils.BuildTime=${BUILD_TIME}' \
  -X 'api-aggregator/backend/pkg/utils.GitCommit=${GIT_COMMIT}'" \
  -o bin/server ./cmd/server

# 构建迁移工具
go build -ldflags "\
  -X 'api-aggregator/backend/pkg/utils.Version=${VERSION}' \
  -X 'api-aggregator/backend/pkg/utils.BuildTime=${BUILD_TIME}' \
  -X 'api-aggregator/backend/pkg/utils.GitCommit=${GIT_COMMIT}'" \
  -o bin/migrate ./scripts/migrate.go
```

## 版本管理

### 查看版本信息

```bash
# 使用 make
make version

# 使用构建的版本工具
./bin/version

# 在代码中使用
import "api-aggregator/backend/pkg/utils"

version := utils.GetVersion()        // "1.0.0"
buildTime := utils.GetBuildTime()    // "2024-01-01T00:00:00Z"
gitCommit := utils.GetGitCommit()    // "abc123"
fullVersion := utils.GetFullVersion() // "1.0.0 (commit: abc123, built: 2024-01-01T00:00:00Z)"
```

### 设置版本号

```bash
# 使用环境变量
export VERSION=1.2.3
make build

# 或直接在命令中指定
make build VERSION=1.2.3
```

## 开发模式

开发时不需要构建，直接运行：

```bash
# 运行服务器
go run ./cmd/server/main.go

# 运行迁移
go run ./scripts/migrate.go

# 查看版本
go run ./cmd/version/main.go
```

开发模式下版本信息会显示为 "dev"。

## 生产构建

```bash
# 设置正式版本号
export VERSION=1.0.0

# 构建
make build

# 生成的二进制文件在 bin/ 目录
ls -lh bin/
# server   - API 服务器
# migrate  - 数据库迁移工具
# version  - 版本信息工具
```

## 交叉编译

```bash
# 为 Linux 构建
GOOS=linux GOARCH=amd64 make build

# 为 Windows 构建
GOOS=windows GOARCH=amd64 make build

# 为 Mac 构建
GOOS=darwin GOARCH=amd64 make build
```

## Docker 构建

```bash
# 构建 Docker 镜像（自动注入版本信息）
docker build \
  --build-arg VERSION=1.0.0 \
  --build-arg BUILD_TIME=$(date -u '+%Y-%m-%dT%H:%M:%SZ') \
  --build-arg GIT_COMMIT=$(git rev-parse HEAD) \
  -t prism-api:1.0.0 .
```

## 持续集成

在 CI/CD 流程中，可以使用 Git 标签作为版本号：

```bash
# 获取最新的 Git 标签作为版本号
VERSION=$(git describe --tags --always)
make build VERSION=$VERSION
```

## 故障排查

### 版本信息显示为 "dev"

这是正常的，说明你在开发模式下运行。使用 `make build` 构建后版本信息会正确显示。

### Git 提交哈希显示为 "unknown"

确保你在 Git 仓库中，并且已经有提交记录：

```bash
git init
git add .
git commit -m "Initial commit"
```

### 构建时间格式错误

确保系统有 `date` 命令，并且支持 `-u` 参数。Windows 用户建议使用 Git Bash 或 WSL。
