import { QueryClient } from '@tanstack/react-query';

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 0, // 数据立即过期，确保总是获取最新数据
      retry: 1,
      refetchOnWindowFocus: true, // 窗口聚焦时重新获取
      refetchOnMount: true, // 组件挂载时重新获取
    },
  },
});
