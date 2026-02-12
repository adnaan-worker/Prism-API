import { apiClient } from '../lib/api';
import { QuotaInfo, SignInResponse, UsageHistoryItem } from '../types';

export const quotaService = {
  // Get quota information
  async getQuotaInfo(): Promise<QuotaInfo> {
    const response = await apiClient.get<QuotaInfo>('/user/quota');
    return response.data;
  },

  // Daily sign-in
  async signIn(): Promise<SignInResponse> {
    const response = await apiClient.post<SignInResponse>('/user/signin');
    return response.data;
  },

  // Get usage history
  async getUsageHistory(days: number = 7): Promise<UsageHistoryItem[]> {
    const response = await apiClient.get<UsageHistoryItem[]>('/user/usage-history', {
      params: { days },
    });
    return response.data;
  },
};
