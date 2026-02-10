# 项目开发规则

## 架构原则

### 1. 单一职责原则 (SRP)
- 每个服务/模块只负责一个功能领域
- 避免"上帝类"，功能过于庞大的类要拆分

### 2. 依赖倒置原则 (DIP)
- 高层模块不依赖低层模块，都依赖抽象
- 使用接口定义契约，便于测试和替换实现

### 3. 开闭原则 (OCP)
- 对扩展开放，对修改关闭
- 通过接口和抽象类实现扩展

## 后端架构规范

### 分层架构
```
cmd/server/          # 应用入口
internal/
  ├── models/        # 数据模型（纯数据结构）
  ├── repository/    # 数据访问层（只负责数据库操作）
  ├── service/       # 业务逻辑层（核心业务逻辑）
  ├── api/           # API 处理层（HTTP 请求处理）
  ├── middleware/    # 中间件（认证、日志等）
  ├── adapter/       # 外部服务适配器
  └── pkg/           # 可复用的工具包
pkg/                 # 公共包（可被外部项目使用）
  ├── response/      # 统一响应格式
  ├── validator/     # 验证工具
  └── utils/         # 通用工具函数
```

### 服务职责划分

#### Repository 层
- **职责**：纯数据访问，CRUD 操作
- **禁止**：包含业务逻辑、调用其他 service
- **命名**：`Find*`, `Create`, `Update`, `Delete`

#### Service 层
- **职责**：业务逻辑、数据组装、事务管理
- **可以**：调用 repository、调用其他 service
- **禁止**：直接处理 HTTP 请求/响应

#### API/Handler 层
- **职责**：HTTP 请求解析、响应格式化、参数验证
- **可以**：调用 service
- **禁止**：包含业务逻辑

### 代码复用规范

#### 1. 避免重复代码
**错误示例**：
```go
// 每个 Handler 都重复写错误响应
c.JSON(http.StatusInternalServerError, gin.H{
    "error": gin.H{
        "code": 500001,
        "message": "Internal server error",
        "details": err.Error(),
    },
})
```

**正确做法**：
```go
// 使用统一的响应工具
response.InternalError(c, err)
```

#### 2. 提取公共组件和工具

**后端公共包**：
- `pkg/response`: 统一的 HTTP 响应格式
- `pkg/validator`: 通用验证函数
- `pkg/utils`: 工具函数（字符串、数学等）

**前端公共模块**：
- `components/`: 可复用 UI 组件
- `hooks/`: 自定义 Hooks
- `utils/`: 工具函数
- `utils/constants.ts`: 全局常量

#### 3. 统一的错误处理
```go
// 定义统一的错误类型
var (
    ErrNotFound = errors.New("resource not found")
    ErrInvalidInput = errors.New("invalid input")
)

// 使用 errors.Is 判断
if errors.Is(err, ErrNotFound) { ... }
```

#### 4. 统一的响应格式
```go
// 使用 response 包
response.Success(c, data)
response.BadRequest(c, "Invalid input", err.Error())
response.InternalError(c, err)
```

### 服务依赖规范

#### 避免循环依赖
```
✅ 正确：
Handler -> Service -> Repository
Service A -> Service B -> Repository

❌ 错误：
Service A -> Service B -> Service A
```

#### 依赖注入
```go
// 通过构造函数注入依赖
func NewProxyService(
    apiKeyRepo *repository.APIKeyRepository,
    billingService *BillingService,
) *ProxyService {
    return &ProxyService{
        apiKeyRepo:     apiKeyRepo,
        billingService: billingService,
    }
}
```

### 命名规范

#### 服务命名
- `*Service`: 业务逻辑服务
- `*Repository`: 数据访问
- `*Handler`: HTTP 处理器
- `*Middleware`: 中间件

#### 方法命名
- Repository: `Find*`, `Create`, `Update`, `Delete`, `Count`
- Service: 业务动词，如 `CalculateCost`, `ProcessPayment`, `ValidateUser`
- Handler: `Handle*`, 或直接用 HTTP 方法名

### 配置管理
```go
// 集中管理配置
type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Redis    RedisConfig
}

// 避免硬编码
const (
    DefaultTimeout = 30 * time.Second
    MaxRetries     = 3
)
```

## 前端架构规范

### 目录结构
```
src/
  ├── components/    # 可复用组件
  │   ├── TableToolbar.tsx    # 表格工具栏
  │   ├── StatusTag.tsx       # 状态标签
  │   └── ProviderTag.tsx     # 厂商标签
  ├── pages/         # 页面组件
  ├── services/      # API 调用封装
  ├── types/         # TypeScript 类型定义
  ├── hooks/         # 自定义 Hooks
  │   ├── useTable.ts         # 表格状态管理
  │   └── useModal.ts         # 模态框状态管理
  ├── utils/         # 工具函数
  │   ├── format.ts           # 格式化函数
  │   └── constants.ts        # 全局常量
  └── layouts/       # 布局组件
```

### 组件设计原则

#### 1. 组件职责单一
```tsx
// ❌ 错误：一个组件做太多事
function UserManagementPage() {
  // 包含：列表、表单、弹窗、权限检查...
}

// ✅ 正确：拆分组件
function UserManagementPage() {
  return (
    <>
      <UserList />
      <UserFormModal />
    </>
  )
}
```

#### 2. 提取可复用组件
```tsx
// 发现重复的表格操作栏
// ❌ 在每个页面重复写
<Space>
  <Button icon={<PlusOutlined />}>添加</Button>
  <Button icon={<ReloadOutlined />}>刷新</Button>
</Space>

// ✅ 提取为组件
<TableToolbar
  onAdd={handleAdd}
  onRefresh={refetch}
  extra={<Select>...</Select>}
/>
```

#### 3. 自定义 Hooks 复用逻辑
```tsx
// ❌ 在每个页面重复写
const { data, isLoading } = useQuery(...)
const createMutation = useMutation(...)
const updateMutation = useMutation(...)

// ✅ 提取为 Hook
function useUsers() {
  const { data, isLoading, refetch } = useQuery(...)
  const createMutation = useMutation(...)
  const updateMutation = useMutation(...)
  
  return { users: data, isLoading, create, update, refetch }
}
```

### API 服务封装

#### 统一的 API 客户端
```typescript
// lib/api.ts
export const apiClient = axios.create({
  baseURL: '/api',
  timeout: 30000,
})

// 统一的错误处理
apiClient.interceptors.response.use(
  response => response,
  error => {
    // 统一处理错误
    return Promise.reject(error)
  }
)
```

#### 服务层封装
```typescript
// services/userService.ts
export const userService = {
  getUsers: (params) => apiClient.get('/users', { params }),
  createUser: (data) => apiClient.post('/users', data),
  updateUser: (id, data) => apiClient.put(`/users/${id}`, data),
}
```

### 类型定义规范

#### 1. 集中管理类型
```typescript
// types/index.ts
export interface User {
  id: number
  username: string
  email: string
}

export interface APIResponse<T> {
  data: T
  total?: number
}
```

#### 2. 避免类型重复
```typescript
// ❌ 错误
interface CreateUserRequest { username: string, email: string }
interface UpdateUserRequest { username: string, email: string }

// ✅ 正确
interface UserFormData { username: string, email: string }
type CreateUserRequest = UserFormData
type UpdateUserRequest = Partial<UserFormData>
```

## 代码质量规范

### 1. 代码审查清单
- [ ] 是否有重复代码？
- [ ] 是否遵循单一职责？
- [ ] 是否有适当的错误处理？
- [ ] 是否有必要的注释？
- [ ] 是否有单元测试？
- [ ] 命名是否清晰？

### 2. 性能优化
- 避免 N+1 查询
- 使用数据库索引
- 合理使用缓存
- 前端使用虚拟滚动处理大列表
- 使用 React.memo 避免不必要的重渲染

### 3. 安全规范
- 所有用户输入必须验证
- 使用参数化查询防止 SQL 注入
- 敏感信息不记录日志
- API 必须有认证和授权
- 使用 HTTPS

### 4. 测试规范
- 单元测试覆盖核心业务逻辑
- 集成测试覆盖关键流程
- 使用 mock 隔离外部依赖

## 重构指南

### 何时重构
1. 发现重复代码（DRY 原则）
2. 函数/类过长（超过 200 行）
3. 职责不清晰
4. 难以测试
5. 难以理解

### 重构步骤
1. 确保有测试覆盖
2. 小步重构，每次只改一个地方
3. 重构后运行测试
4. 提交代码

### 重构技巧
- **提取方法**：将长函数拆分
- **提取类**：将相关方法组合
- **引入参数对象**：减少参数数量
- **用多态替换条件**：消除 if-else
- **提取接口**：解耦依赖

## 文档规范

### 代码注释
```go
// CalculateCost 计算请求的费用
// 参数：
//   - modelName: 模型名称
//   - inputTokens: 输入 token 数量
//   - outputTokens: 输出 token 数量
// 返回：
//   - cost: 费用（单位：credits）
//   - error: 错误信息
func (s *BillingService) CalculateCost(...) (float64, error)
```

### API 文档
- 使用 Swagger/OpenAPI
- 包含请求/响应示例
- 说明错误码

### README
- 项目介绍
- 快速开始
- 架构说明
- 开发指南

## 版本控制规范

### Commit 消息格式
```
<type>(<scope>): <subject>

<body>

<footer>
```

类型：
- `feat`: 新功能
- `fix`: 修复 bug
- `refactor`: 重构
- `docs`: 文档
- `test`: 测试
- `chore`: 构建/工具

示例：
```
feat(billing): 添加微积分精度支持

- 使用 micro-credits 保留 3 位小数
- 添加 BillingService 统一管理计费
- 重构 ProxyService 消除重复代码

Closes #123
```

## 持续改进

### 定期审查
- 每周代码审查会议
- 每月架构评审
- 季度技术债务清理

### 技术债务管理
- 使用 TODO 标记待优化代码
- 维护技术债务清单
- 优先级排序，逐步解决

### 学习分享
- 团队技术分享会
- 代码审查中学习最佳实践
- 记录踩坑经验
