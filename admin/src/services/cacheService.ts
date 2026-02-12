import { apiClient } from '../lib/api';

export interface CacheStats {
  total_hits: number;
  tokens_saved: number;
  cache_entries: number;
}

export const cacheService = {
  // 获取缓存统计
  getCacheStats: async (): Promise<CacheStats> => {
    const response = await apiClient.get<CacheStats>('/admin/cache/stats');
    return response.data;
  },

  // 清理过期缓存（管理员）
  cleanExpiredCache: async (): Promise<any> => {
    const response = await apiClient.post('/admin/cache/clean-expired');
    return response.data;
  },

  // 清除用户缓存（管理员）
  clearUserCache: async (userId: number): Promise<any> => {
    const response = await apiClient.delete(`/admin/cache/users/${userId}`);
    return response.data;
  },
  cleanExpiredCache: () => apiClient.delete('/admin/cache/clean'),

  // 清除用户缓存（管理员）
  clearUserCache: (userId: number) =>
    apiClient.delete(`/admin/cache/user/${userId}`),
};
