import { apiClient } from '../lib/api';
import { Pricing } from '../types';

export interface PricingsListResponse {
  pricings: Pricing[];
  total: number;
}

export interface CreatePricingRequest {
  api_config_id: number;
  model_name: string;
  input_price: number;
  output_price: number;
  currency?: string;
  unit?: number;
  is_active?: boolean;
  description?: string;
}

export interface UpdatePricingRequest extends Partial<CreatePricingRequest> {}

export const pricingService = {
  // 获取所有定价配置
  getAllPricings: async (): Promise<PricingsListResponse> => {
    const response = await apiClient.get<PricingsListResponse>('/admin/pricings');
    return response.data;
  },

  // 获取定价配置详情
  getPricingById: async (id: number): Promise<Pricing> => {
    const response = await apiClient.get<Pricing>(`/admin/pricings/${id}`);
    return response.data;
  },

  // 按 API 配置获取定价
  getPricingsByAPIConfig: async (apiConfigId: number): Promise<PricingsListResponse> => {
    const response = await apiClient.get<PricingsListResponse>(
      '/admin/pricings/by-config',
      { params: { api_config_id: apiConfigId } }
    );
    return response.data;
  },

  // 创建定价配置
  createPricing: async (data: CreatePricingRequest): Promise<Pricing> => {
    const response = await apiClient.post<Pricing>('/admin/pricings', data);
    return response.data;
  },

  // 更新定价配置
  updatePricing: async (id: number, data: UpdatePricingRequest): Promise<void> => {
    await apiClient.put(`/admin/pricings/${id}`, data);
  },

  // 删除定价配置
  deletePricing: async (id: number): Promise<void> => {
    await apiClient.delete(`/admin/pricings/${id}`);
  },
};
