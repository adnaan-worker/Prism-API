import { apiClient as api } from '../lib/api';
import type { AccountPool, AccountCredential, PoolStats } from '../types';

export const accountPoolService = {
  // 账号池管理
  getPools: async (params?: { provider?: string }) => {
    const { data } = await api.get<AccountPool[]>('/admin/account-pools', { params });
    return data;
  },

  getPool: async (id: number) => {
    const { data } = await api.get<AccountPool>(`/admin/account-pools/${id}`);
    return data;
  },

  createPool: async (pool: Partial<AccountPool>) => {
    const { data } = await api.post<AccountPool>('/admin/account-pools', pool);
    return data;
  },

  updatePool: async (id: number, pool: Partial<AccountPool>) => {
    const { data } = await api.put<AccountPool>(`/admin/account-pools/${id}`, pool);
    return data;
  },

  deletePool: async (id: number) => {
    await api.delete(`/admin/account-pools/${id}`);
  },

  updatePoolStatus: async (id: number, isActive: boolean) => {
    const { data } = await api.put<AccountPool>(`/admin/account-pools/${id}/status`, { is_active: isActive });
    return data;
  },

  getPoolStats: async (id: number) => {
    const { data } = await api.get<PoolStats>(`/admin/account-pools/${id}/stats`);
    return data;
  },

  // 凭据管理
  getCredentials: async (params?: { pool_id?: number; provider?: string; status?: string }) => {
    const { data } = await api.get<AccountCredential[]>('/admin/account-credentials', { params });
    return data;
  },

  getCredential: async (id: number) => {
    const { data } = await api.get<AccountCredential>(`/admin/account-credentials/${id}`);
    return data;
  },

  createCredential: async (credential: Partial<AccountCredential>) => {
    const { data } = await api.post<AccountCredential>('/admin/account-credentials', credential);
    return data;
  },

  updateCredential: async (id: number, credential: Partial<AccountCredential>) => {
    const { data } = await api.put<AccountCredential>(`/admin/account-credentials/${id}`, credential);
    return data;
  },

  deleteCredential: async (id: number) => {
    await api.delete(`/admin/account-credentials/${id}`);
  },

  refreshCredential: async (id: number) => {
    const { data } = await api.post<AccountCredential>(`/admin/account-credentials/${id}/refresh`);
    return data;
  },

  updateCredentialStatus: async (id: number, isActive: boolean) => {
    const { data } = await api.put<AccountCredential>(`/admin/account-credentials/${id}/status`, { is_active: isActive });
    return data;
  },

  // OAuth
  initiateOAuth: async (params: { pool_id: number; provider: string; auth_type?: string }) => {
    const { data } = await api.post('/admin/oauth/initiate', params);
    return data;
  },

  pollDeviceCode: async (params: { pool_id: number; provider: string; device_code: string }) => {
    const { data } = await api.post('/admin/oauth/poll-device-code', params);
    return data;
  },
};
