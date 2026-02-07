import { apiClient } from '../lib/api';
import { User } from '../types';

export interface UsersListResponse {
  users: User[];
  total: number;
  page: number;
  page_size: number;
}

export interface UpdateUserStatusRequest {
  status: string;
}

export interface UpdateUserQuotaRequest {
  quota: number;
}

export const userService = {
  // 获取用户列表
  getUsers: async (params?: {
    page?: number;
    page_size?: number;
    search?: string;
    status?: string;
  }): Promise<UsersListResponse> => {
    const response = await apiClient.get<UsersListResponse>('/admin/users', {
      params,
    });
    return response.data;
  },

  // 获取用户详情
  getUserById: async (id: number): Promise<User> => {
    const response = await apiClient.get<User>(`/admin/users/${id}`);
    return response.data;
  },

  // 更新用户状态
  updateUserStatus: async (
    id: number,
    data: UpdateUserStatusRequest
  ): Promise<User> => {
    const response = await apiClient.put<User>(
      `/admin/users/${id}/status`,
      data
    );
    return response.data;
  },

  // 更新用户额度
  updateUserQuota: async (
    id: number,
    data: UpdateUserQuotaRequest
  ): Promise<User> => {
    const response = await apiClient.put<User>(
      `/admin/users/${id}/quota`,
      data
    );
    return response.data;
  },

  // 禁用用户
  disableUser: async (id: number): Promise<User> => {
    return userService.updateUserStatus(id, { status: 'disabled' });
  },

  // 启用用户
  enableUser: async (id: number): Promise<User> => {
    return userService.updateUserStatus(id, { status: 'active' });
  },
};
