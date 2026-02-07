import { apiClient } from '../lib/api';
import { QuotaInfo, SignInResponse } from '../types';

export const quotaService = {
  // Get quota information
  async getQuotaInfo(): Promise<QuotaInfo> {
    const response = await apiClient.get<QuotaInfo>('/user/info');
    return response.data;
  },

  // Daily sign-in
  async signIn(): Promise<SignInResponse> {
    const response = await apiClient.post<SignInResponse>('/user/signin');
    return response.data;
  },
};
