import { apiClient } from '../lib/api';
import { RequestLog } from '../types';

export interface LogsListResponse {
  logs: RequestLog[];
  total: number;
  page: number;
  page_size: number;
}

export interface LogsQueryParams {
  page?: number;
  page_size?: number;
  start_time?: string;
  end_time?: string;
  user_id?: number;
  model?: string;
  status?: 'success' | 'error';
}

export const logService = {
  // 获取请求日志列表
  getLogs: async (params?: LogsQueryParams): Promise<LogsListResponse> => {
    const response = await apiClient.get<LogsListResponse>('/admin/logs', {
      params,
    });
    return response.data;
  },

  // 导出日志为CSV
  exportLogs: async (params?: LogsQueryParams): Promise<Blob> => {
    const response = await apiClient.get('/admin/logs/export', {
      params,
      responseType: 'blob',
    });
    return response.data;
  },

  // 获取日志详情
  getLogById: async (id: number): Promise<RequestLog> => {
    const response = await apiClient.get<RequestLog>(`/admin/logs/${id}`);
    return response.data;
  },
};
