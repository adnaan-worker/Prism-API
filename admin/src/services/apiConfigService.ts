import { apiClient } from '../lib/api';
import { APIConfig } from '../types';

export interface APIConfigsListResponse {
  configs: APIConfig[];
  total: number;
  page: number;
  page_size: number;
}

export interface CreateAPIConfigRequest {
  name: string;
  type: string;
  base_url: string;
  api_key?: string;
  models: string[];
  priority?: number;
  weight?: number;
  max_rps?: number;
  timeout?: number;
}

export interface UpdateAPIConfigRequest extends Partial<CreateAPIConfigRequest> {}

export const apiConfigService = {
  // 获取API配置列表
  getConfigs: async (params?: {
    page?: number;
    page_size?: number;
    type?: string;
    is_active?: boolean;
  }): Promise<APIConfigsListResponse> => {
    const response = await apiClient.get<APIConfigsListResponse>(
      '/admin/api-configs',
      { params }
    );
    return response.data;
  },

  // 获取API配置详情
  getConfigById: async (id: number): Promise<APIConfig> => {
    const response = await apiClient.get<APIConfig>(`/admin/api-configs/${id}`);
    return response.data;
  },

  // 创建API配置
  createConfig: async (data: CreateAPIConfigRequest): Promise<APIConfig> => {
    const response = await apiClient.post<APIConfig>('/admin/api-configs', data);
    return response.data;
  },

  // 更新API配置
  updateConfig: async (
    id: number,
    data: UpdateAPIConfigRequest
  ): Promise<APIConfig> => {
    const response = await apiClient.put<APIConfig>(
      `/admin/api-configs/${id}`,
      data
    );
    return response.data;
  },

  // 删除API配置
  deleteConfig: async (id: number): Promise<void> => {
    await apiClient.delete(`/admin/api-configs/${id}`);
  },

  // 启用/禁用API配置
  toggleConfigStatus: async (id: number, is_active: boolean): Promise<APIConfig> => {
    const endpoint = is_active ? 'activate' : 'deactivate';
    const response = await apiClient.put<APIConfig>(
      `/admin/api-configs/${id}/${endpoint}`
    );
    return response.data;
  },

  // 批量删除API配置
  batchDeleteConfigs: async (ids: number[]): Promise<void> => {
    await apiClient.post('/admin/api-configs/batch-delete', { ids });
  },

  // 从提供商获取模型列表
  fetchModels: async (params: {
    type: string;
    base_url: string;
    api_key: string;
  }): Promise<{ models: string[]; total: number }> => {
    const response = await apiClient.post<{ models: string[]; total: number }>(
      '/admin/providers/fetch-models',
      params
    );
    return response.data;
  },
};
