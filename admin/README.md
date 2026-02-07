# API聚合平台 - 管理后台

管理后台是一个独立的React应用，用于管理API聚合平台的用户、API配置、负载均衡和请求日志。

## 技术栈

- **React 18** - UI框架
- **Vite** - 构建工具
- **TypeScript** - 类型安全
- **Ant Design** - UI组件库
- **React Router** - 路由管理
- **TanStack Query** - 数据获取和缓存
- **ECharts** - 图表库
- **Axios** - HTTP客户端

## 项目结构

```
admin/
├── src/
│   ├── components/      # 通用组件
│   ├── pages/          # 页面组件
│   ├── services/       # API服务
│   ├── lib/            # 工具库
│   │   ├── api.ts      # Axios配置
│   │   └── queryClient.ts  # TanStack Query配置
│   ├── router/         # 路由配置
│   ├── styles/         # 样式文件
│   ├── types/          # TypeScript类型定义
│   ├── App.tsx         # 根组件
│   └── main.tsx        # 入口文件
├── public/             # 静态资源
├── .env                # 环境变量
├── .env.example        # 环境变量示例
├── index.html          # HTML模板
├── package.json        # 依赖配置
├── tsconfig.json       # TypeScript配置
└── vite.config.ts      # Vite配置
```

## 开发指南

### 安装依赖

```bash
cd admin
npm install
```

### 启动开发服务器

```bash
npm run dev
```

应用将在 http://localhost:3001 启动

### 构建生产版本

```bash
npm run build
```

### 预览生产构建

```bash
npm run preview
```

## 环境变量

复制 `.env.example` 到 `.env` 并配置：

```
VITE_API_BASE_URL=http://localhost:8080/api
```

## 核心配置

### React Router

路由配置在 `src/router/index.tsx`，使用 `createBrowserRouter` 创建路由。

### TanStack Query

Query客户端配置在 `src/lib/queryClient.ts`，默认配置：
- 缓存时间：5分钟
- 重试次数：1次
- 窗口聚焦时不自动刷新

### Axios

API客户端配置在 `src/lib/api.ts`，包含：
- 自动添加认证token
- 401错误自动跳转登录
- 30秒超时

### Ant Design

使用中文语言包，在 `main.tsx` 中配置 `ConfigProvider`。

## 待实现功能

根据任务列表，以下功能将在后续任务中实现：

- [ ] 管理后台布局（侧边栏、顶部栏）
- [ ] 统计概览页面
- [ ] 用户管理页面
- [ ] API配置管理页面
- [ ] 负载均衡配置页面
- [ ] 请求日志页面

## 注意事项

1. 管理后台使用独立的token存储（`admin_token`），与用户门户分离
2. 所有API请求都会自动添加认证token
3. 401错误会自动清除token并跳转到登录页
4. 使用Ant Design的中文语言包
