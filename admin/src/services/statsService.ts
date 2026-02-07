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

  // 获取请求趋势（模拟数据，实际应该从后端获取）
  getRequestTrend: async (days: number = 7): Promise<RequestTrend[]> => {
    // TODO: 实际应该调用后端API
    // const response = await apiClient.get<RequestTrend[]>(`/admin/stats/trend?days=${days}`);
    // return response.data;
    
    // 模拟数据
    const today = new Date();
    const trend: RequestTrend[] = [];
    for (let i = days - 1; i >= 0; i--) {
      const date = new Date(today);
      date.setDate(date.getDate() - i);
      trend.push({
        date: date.toISOString().split('T')[0],
        count: Math.floor(Math.random() * 1000) + 500,
      });
    }
    return trend;
  },

  // 获取模型使用排行（模拟数据，实际应该从后端获取）
  getModelUsage: async (): Promise<ModelUsage[]> => {
    // TODO: 实际应该调用后端API
    // const response = await apiClient.get<ModelUsage[]>('/admin/stats/models');
    // return response.data;
    
    // 模拟数据
    return [
      { model: 'gpt-4', count: 1250 },
      { model: 'gpt-3.5-turbo', count: 980 },
      { model: 'claude-3-opus', count: 750 },
      { model: 'claude-3-sonnet', count: 620 },
      { model: 'gemini-pro', count: 450 },
    ];
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
