import { apiClient } from '../lib/api';
import { LoginRequest, LoginResponse } from '../types';

export const authService = {
  login: async (credentials: LoginRequest): Promise<LoginResponse> => {
    const response = await apiClient.post<LoginResponse>('/auth/login', credentials);
    return response.data;
  },

  logout: () => {
    localStorage.removeItem('admin_token');
    window.location.href = '/login';
  },

  isAuthenticated: (): boolean => {
    return !!localStorage.getItem('admin_token');
  },

  getToken: (): string | null => {
    return localStorage.getItem('admin_token');
  },

  setToken: (token: string): void => {
    localStorage.setItem('admin_token', token);
  },
};
