import { apiClient } from '../lib/api';

export interface StatsOverview {
  total_users: number;
  active_users: number;
  total_requests: number;
  today_requests: number;
}

export interface RequestTrend {
  date: string;
  count: number;
}

export interface ModelUsage {
  model: string;
  count: number;
}

export const statsService = {
  // 获取统计概览
  getOverview: async (): Promise<StatsOverview> => {
    const response = await apiClient.get<StatsOverview>('/admin/stats/overview');
    return response.data;
  },

  // 获取请求趋势
  getRequestTrend: async (days: number = 7): Promise<RequestTrend[]> => {
    const response = await apiClient.get<{ trend: RequestTrend[]; days: number }>('/admin/stats/trend', {
      params: { days },
    });
    return response.data.trend || [];
  },

  // 获取模型使用排行
  getModelUsage: async (limit: number = 10): Promise<ModelUsage[]> => {
    const response = await apiClient.get<{ usage: ModelUsage[]; total: number }>('/admin/stats/models', {
      params: { limit },
    });
    return response.data.usage || [];
  },

  // 获取最近日志
  getRecentLogs: async (limit: number = 5): Promise<any> => {
    const response = await apiClient.get('/admin/logs', {
      params: {
        page: 1,
        page_size: limit,
      },
    });
    return response.data;
  },
};
