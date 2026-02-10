import { apiClient } from '../lib/api';

export interface CacheStats {
  total_hits: number;
  tokens_saved: number;
  cache_entries: number;
}

export const cacheService = {
  // 获取缓存统计
  getCacheStats: () => apiClient.get<CacheStats>('/admin/cache/stats'),

  // 清理过期缓存（管理员）
  cleanExpiredCache: () => apiClient.post('/admin/cache/clean'),

  // 清除用户缓存（管理员）
  clearUserCache: (userId: number) =>
    apiClient.delete(`/admin/cache/users/${userId}`),
};
