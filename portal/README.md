# API聚合平台 - 用户门户

用户门户前端应用，基于 React 18 + Vite 构建。

## 技术栈

- **React 18** - UI 框架
- **Vite** - 构建工具
- **TypeScript** - 类型安全
- **Ant Design 5** - UI 组件库
- **Tailwind CSS 3** - 样式框架
- **React Router 6** - 路由管理
- **TanStack Query 5** - 数据获取和状态管理
- **Axios** - HTTP 客户端

## 项目结构

```
portal/
├── public/              # 静态资源
├── src/
│   ├── components/      # 通用组件
│   ├── pages/          # 页面组件
│   ├── services/       # API 服务
│   ├── lib/            # 工具库
│   │   ├── api.ts      # Axios 配置
│   │   └── queryClient.ts  # TanStack Query 配置
│   ├── router/         # 路由配置
│   ├── styles/         # 样式文件
│   ├── types/          # TypeScript 类型定义
│   ├── App.tsx         # 根组件
│   └── main.tsx        # 入口文件
├── .env.example        # 环境变量示例
├── tailwind.config.js  # Tailwind 配置
├── vite.config.ts      # Vite 配置
└── package.json        # 依赖配置
```

## 开发指南

### 安装依赖

```bash
cd portal
npm install
```

### 启动开发服务器

```bash
npm run dev
```

应用将在 http://localhost:3000 启动

### 构建生产版本

```bash
npm run build
```

### 预览生产构建

```bash
npm run preview
```

## 配置说明

### 环境变量

复制 `.env.example` 为 `.env` 并配置：

```env
VITE_API_BASE_URL=http://localhost:8080/api
VITE_APP_NAME=API聚合平台
VITE_APP_VERSION=1.0.0
```

### Ant Design 主题

在 `src/main.tsx` 中配置主题：

```typescript
const theme = {
  token: {
    colorPrimary: '#1890ff',
    borderRadius: 8,
    fontFamily: 'Inter, -apple-system, BlinkMacSystemFont, Segoe UI, Roboto, sans-serif',
  },
};
```

### Tailwind CSS

Tailwind 配置在 `tailwind.config.js` 中，已禁用 preflight 以避免与 Ant Design 冲突。

### 路径别名

使用 `@/` 作为 `src/` 的别名：

```typescript
import { apiClient } from '@/lib/api';
import { User } from '@/types';
```

## API 集成

### 使用 Axios 客户端

```typescript
import { apiClient } from '@/lib/api';

const response = await apiClient.get('/user/info');
```

### 使用 TanStack Query

```typescript
import { useQuery } from '@tanstack/react-query';
import { apiClient } from '@/lib/api';

function useUserInfo() {
  return useQuery({
    queryKey: ['user', 'info'],
    queryFn: async () => {
      const { data } = await apiClient.get('/user/info');
      return data;
    },
  });
}
```

## 代码规范

- 使用 TypeScript 严格模式
- 组件使用函数式组件 + Hooks
- 使用 Ant Design 组件优先
- 使用 Tailwind CSS 进行样式定制
- API 调用统一使用 TanStack Query

## 下一步

- [ ] 实现 Landing Page
- [ ] 实现用户认证页面
- [ ] 实现 Dashboard 布局
- [ ] 实现各功能页面
