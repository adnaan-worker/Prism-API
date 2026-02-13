import { apiClient } from '../lib/api';

export interface LoadBalancerConfig {
  id: number;
  model_name: string;
  strategy: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface ModelEndpoint {
  config_id: number;
  config_name: string;
  type: string;
  base_url: string;
  priority: number;
  weight: number;
  is_active: boolean;
  health_status: 'healthy' | 'unhealthy' | 'unknown';
  response_time?: number;
  success_rate?: number;
}

export interface CreateLoadBalancerConfigRequest {
  model_name: string;
  strategy: string;
}

export interface UpdateLoadBalancerConfigRequest {
  strategy?: string;
  is_active?: boolean;
}

export const loadBalancerService = {
  // 获取负载均衡配置列表
  getConfigs: async (): Promise<LoadBalancerConfig[]> => {
    const response = await apiClient.get<{ configs: LoadBalancerConfig[]; total: number }>(
      '/admin/load-balancer/configs'
    );
    // 后端返回 { configs: [...], total: 10 }，拦截器已提取 data
    return response.data?.configs || [];
  },

  // 获取指定模型的端点列表
  getModelEndpoints: async (modelName: string): Promise<ModelEndpoint[]> => {
    const response = await apiClient.get<{ endpoints: ModelEndpoint[]; total: number }>(
      `/admin/load-balancer/models/${modelName}/endpoints`
    );
    return response.data?.endpoints || [];
  },

  // 创建负载均衡配置
  createConfig: async (
    data: CreateLoadBalancerConfigRequest
  ): Promise<LoadBalancerConfig> => {
    const response = await apiClient.post<LoadBalancerConfig>(
      '/admin/load-balancer/configs',
      data
    );
    return response.data;
  },

  // 更新负载均衡配置
  updateConfig: async (
    id: number,
    data: UpdateLoadBalancerConfigRequest
  ): Promise<LoadBalancerConfig> => {
    const response = await apiClient.put<LoadBalancerConfig>(
      `/admin/load-balancer/configs/${id}`,
      data
    );
    return response.data;
  },

  // 删除负载均衡配置
  deleteConfig: async (id: number): Promise<void> => {
    await apiClient.delete(`/admin/load-balancer/configs/${id}`);
  },

  // 获取可用模型列表（从配置中提取）
  getAvailableModels: async (): Promise<string[]> => {
    const response = await apiClient.get<{ configs: LoadBalancerConfig[]; total: number }>(
      '/admin/load-balancer/configs'
    );
    const configs = response.data?.configs || [];
    // 提取所有唯一的模型名称
    const models = Array.from(new Set(configs.map(c => c.model_name)));
    return models;
  },
};
