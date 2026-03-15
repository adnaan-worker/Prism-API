<div align="center">

<img src="logo.svg" alt="Prism API Logo" width="120" height="120" />

# Prism API

**棱镜 —— 通用 AI API 网关**

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.23-blue.svg)](https://golang.org/)
[![React Version](https://img.shields.io/badge/React-18-blue.svg)](https://reactjs.org/)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](http://makeapullrequest.com)

*作者：adnaan*

[功能特性](#-功能特性) • [快速开始](#-快速开始) • [文档](#-文档) • [架构](#️-架构)

</div>

---

一个统一的 AI API 网关，通过单一优雅的接口聚合多个 AI 服务提供商（OpenAI、Anthropic、Gemini）。就像棱镜将光线折射成多种颜色一样，Prism API 无缝地转换和路由不同 AI 提供商之间的请求。

## ✨ 功能特性

- 🌈 **通用接口**：支持 OpenAI、Anthropic 和 Gemini 格式
- 🔄 **多提供商支持**：聚合多个 AI 服务提供商
- ⚖️ **智能负载均衡**：轮询、加权、最少连接和随机策略
- 🔑 **API 密钥管理**：为每个用户生成独立的 API 密钥
- 💰 **配额管理**：灵活的配额分配和使用跟踪
- 📊 **数据分析**：详细的使用统计和请求日志
- 🚀 **高性能**：Redis 缓存和连接池优化
- 🔒 **安全可靠**：JWT 认证、速率限制和管理员访问控制
- 📱 **现代化界面**：使用 React 和 Ant Design 构建的用户友好界面
- 🌊 **流式支持**：所有提供商的完整 SSE 流式支持

## 🏗️ 架构

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│ 用户门户    │     │ 管理面板    │     │   Nginx     │
│   (React)   │────▶│   (React)   │────▶│  反向代理   │
└─────────────┘     └─────────────┘     └──────┬──────┘
                                               │
                                               ▼
                                        ┌─────────────┐
                                        │ 后端 API    │
                                        │    (Go)     │
                                        └──────┬──────┘
                                               │
                        ┌──────────────────────┼──────────────────────┐
                        ▼                      ▼                      ▼
                 ┌─────────────┐       ┌─────────────┐       ┌─────────────┐
                 │ PostgreSQL  │       │    Redis    │       │  AI APIs    │
                 │   数据库    │       │    缓存     │       │  提供商     │
                 └─────────────┘       └─────────────┘       └─────────────┘
```

## 🎯 Prism API 的特别之处

**Prism API** 充当 AI API 的通用转换器和路由器。就像棱镜将白光折射成光谱一样，Prism API：

- **折射**任何格式的传入请求（OpenAI、Anthropic、Gemini）
- **转换**为统一的内部格式
- **路由**到适当的 AI 提供商
- **反射**以原始格式返回响应

这意味着您可以：
- 使用 OpenAI 的格式调用 Claude 模型
- 使用 Anthropic 的格式调用 GPT 模型
- 使用 Gemini 的格式调用任何提供商
- 无需更改代码即可切换提供商

## 📋 项目结构

```
prism-api/
├── backend/          # Go 后端服务
│   ├── cmd/         # 应用程序入口
│   ├── config/      # 配置管理
│   ├── internal/    # 内部代码
│   │   ├── adapter/     # API 适配器（OpenAI、Anthropic、Gemini）
│   │   ├── api/         # HTTP 处理器
│   │   ├── loadbalancer/# 负载均衡策略
│   │   ├── middleware/  # 中间件（认证、速率限制）
│   │   ├── models/      # 数据模型
│   │   ├── repository/  # 数据访问层
│   │   └── service/     # 业务逻辑
│   ├── pkg/         # 公共库
│   └── scripts/     # 数据库迁移
├── portal/           # 用户门户（React）
│   ├── src/
│   │   ├── components/  # 组件
│   │   ├── layouts/     # 布局
│   │   ├── pages/       # 页面
│   │   ├── services/    # API 服务
│   │   └── router/      # 路由
│   └── public/          # 静态资源
├── admin/            # 管理面板（React）
│   └── src/             # 与 portal 结构类似
├── docs/             # 文档
│   ├── API.md           # API 文档
│   ├── DEPLOYMENT.md    # 部署指南
│   └── DEVELOPMENT.md   # 开发指南
├── docker-compose.yml   # Docker 配置
├── nginx.conf           # Nginx 配置
└── README.md
```

## 🛠️ 技术栈

### 后端
- Go 1.23
- Gin 框架
- PostgreSQL 15
- Redis 7
- GORM

### 前端
- React 18
- Vite
- Ant Design
- TanStack Query
- TypeScript

## 🚀 快速开始

### 使用 Docker Compose（推荐）

1. **克隆仓库**
```bash
# GitHub
git clone https://github.com/adnaan-worker/prism-api.git
cd prism-api

# 或者使用 Gitee（国内推荐）
git clone https://gitee.com/adnaan/prism-api.git
cd prism-api
```

2. **配置环境变量**
```bash
cp .env.example .env
# 编辑 .env 文件，至少需要修改 JWT_SECRET
```

3. **启动所有服务**
```bash
docker-compose up -d
```

4. **初始化管理员账户**（首次部署时必需）
```bash
# 进入后端容器
docker-compose exec backend bash

# 运行初始化脚本
cd scripts
go run init_admin.go
```

或者直接在宿主机运行：
```bash
cd backend/scripts
go run init_admin.go
```

5. **访问应用**
- 用户门户：http://localhost:3000
- 管理面板：http://localhost:3001（需要管理员账户登录）
- 后端 API：http://localhost:8080/api

### 本地开发

#### 前置要求
- Go 1.23+
- Node.js 20+
- PostgreSQL 15+
- Redis 7+

#### 启动后端

```bash
cd backend

# 安装依赖
go mod download

# 启动数据库和 Redis
docker-compose up -d postgres redis

# 运行数据库迁移
go run scripts/migrate.go

# 初始化管理员账户（首次部署时必需）
go run scripts/init_admin.go

# 启动服务器
go run cmd/server/main.go
```

后端将运行在 http://localhost:8080

#### 初始化管理员账户

首次部署时，需要创建管理员账户才能访问管理面板：

```bash
cd backend/scripts

# 使用默认配置创建管理员（从 .env 文件读取）
go run init_admin.go

# 使用自定义凭据创建管理员
export ADMIN_USERNAME=myadmin
export ADMIN_EMAIL=myadmin@example.com
export ADMIN_PASSWORD=mypassword123
go run init_admin.go

# 强制更新已存在的管理员密码（非交互模式）
go run init_admin.go --force
```

**默认管理员凭据**（如果未配置）：
- 用户名：`admin`
- 邮箱：`admin@example.com`
- 密码：`admin123`

**注意**：
- 脚本会自动检查数据库表是否存在，如果不存在会自动运行迁移
- 如果管理员账户已存在，脚本会询问是否更新密码
- 使用 `--force` 标志可以在非交互模式下强制更新密码
- 生产环境建议使用强密码并修改默认凭据

#### 启动用户门户

```bash
cd portal

# 安装依赖
npm install

# 启动开发服务器
npm run dev
```

门户将运行在 http://localhost:3000

#### 启动管理面板

```bash
cd admin

# 安装依赖
npm install

# 启动开发服务器
npm run dev
```

管理面板将运行在 http://localhost:3001

## 📖 文档

- [API 文档](docs/API.md) - 完整的 API 参考
- [部署指南](docs/DEPLOYMENT.md) - 生产环境部署指南
- [开发指南](docs/DEVELOPMENT.md) - 开发设置和最佳实践

## 🔧 配置

### 环境变量

关键环境变量（参见 `.env.example`）：

```env
# 数据库
DATABASE_URL=postgres://postgres:postgres@localhost:5432/api_aggregator?sslmode=disable
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=5m

# Redis
REDIS_URL=redis://localhost:6379
REDIS_POOL_SIZE=10
REDIS_MIN_IDLE_CONN=2

# JWT（生产环境必须修改！）
JWT_SECRET=your-secret-key-change-in-production

# 服务器
PORT=8080
SERVER_READ_TIMEOUT=10s
SERVER_WRITE_TIMEOUT=10s
REQUEST_TIMEOUT=30s

# 初始管理员账户（可选）
# 配置后运行 go run backend/scripts/init_admin.go 创建管理员账户
ADMIN_USERNAME=admin
ADMIN_EMAIL=admin@example.com
ADMIN_PASSWORD=admin123
```

### 初始化管理员账户

首次部署时，需要创建管理员账户才能访问管理面板。有两种方式：

#### 方式 1：使用初始化脚本（推荐）

```bash
cd backend/scripts
go run init_admin.go
```

脚本会：
- 自动从 `.env` 文件读取管理员配置
- 检查数据库表是否存在，不存在则自动运行迁移
- 如果管理员已存在，询问是否更新密码
- 创建或更新管理员账户

#### 方式 2：服务器启动时自动创建

服务器启动时会自动检查并创建管理员账户（如果配置了且账户不存在）。这种方式适合开发环境。

**生产环境建议**：
- 使用初始化脚本手动创建管理员账户
- 使用强密码（至少 12 位，包含大小写字母、数字和特殊字符）
- 修改默认的管理员用户名和邮箱
- 不要在代码仓库中提交包含真实密码的 `.env` 文件

## 🧪 测试

### 后端测试

```bash
cd backend

# 运行所有测试
go test ./...

# 运行测试并生成覆盖率报告
go test -cover ./...

# 运行特定测试
go test -v ./internal/service -run TestAuthService
```

### 前端测试

```bash
cd portal  # 或 cd admin

# 运行测试
npm test

# 运行测试并生成覆盖率报告
npm test -- --coverage
```

## 📊 功能详情

### 用户门户

- ✅ 用户注册和登录
- ✅ API 密钥管理
- ✅ 配额跟踪和每日签到
- ✅ 模型浏览
- ✅ 文档和代码示例
- ✅ 个人统计和使用历史

### 管理面板

- ✅ 用户管理（查看、启用/禁用、调整配额）
- ✅ API 配置管理（增删改查、批量操作）
- ✅ 负载均衡器配置
- ✅ 统计概览（用户、请求、趋势）
- ✅ 请求日志和导出

### 后端 API

- ✅ 用户认证（JWT）
- ✅ API 密钥认证
- ✅ 速率限制（基于 Redis）
- ✅ 多提供商适配器（OpenAI、Anthropic、Gemini）
- ✅ 智能负载均衡
- ✅ 配额管理
- ✅ 请求日志记录
- ✅ 管理员访问控制
- ✅ 流式支持（SSE）

## 🔐 安全性

- **密码加密**：bcrypt 哈希
- **JWT 认证**：安全的基于令牌的认证
- **API 密钥**：安全的 sk- 前缀密钥
- **速率限制**：防止 API 滥用
- **管理员权限**：严格的访问控制
- **CORS 配置**：跨域保护

## 🚀 性能

- **连接池**：优化的数据库和 Redis 连接
- **缓存**：热数据的 Redis 缓存
- **负载均衡**：智能请求分发
- **异步处理**：使用 Goroutines 处理并发请求
- **索引优化**：数据库查询优化

## 🤝 贡献

欢迎提交 Issues 和 Pull Requests！

## 👨‍💻 作者

**Adnaan**
- GitHub: [@adnaan-worker](https://github.com/adnaan-worker)
- Gitee: [@adnaan](https://gitee.com/adnaan)

## 🔗 仓库地址

- GitHub: [https://github.com/adnaan-worker/prism-api](https://github.com/adnaan-worker/prism-api)
- Gitee: [https://gitee.com/adnaan/prism-api](https://gitee.com/adnaan/prism-api)（国内推荐）

## 📄 许可证

MIT
