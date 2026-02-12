import { apiClient } from '../lib/api';
import { LoginRequest, RegisterRequest, AuthResponse, User } from '../types';

export const authService = {
  // User login
  async login(credentials: LoginRequest): Promise<AuthResponse> {
    const response = await apiClient.post<AuthResponse>('/auth/login', credentials);
    return response.data;
  },

  // User registration
  async register(data: RegisterRequest): Promise<AuthResponse> {
    const response = await apiClient.post<AuthResponse>('/auth/register', data);
    return response.data;
  },

  // Get current user info
  async getCurrentUser(): Promise<User> {
    const response = await apiClient.get<{ user: User }>('/auth/profile');
    return response.data.user;
  },

  // Logout
  logout(): void {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
  },

  // Save auth data to localStorage
  saveAuthData(token: string, user: User): void {
    localStorage.setItem('token', token);
    localStorage.setItem('user', JSON.stringify(user));
  },

  // Get saved user from localStorage
  getSavedUser(): User | null {
    const userStr = localStorage.getItem('user');
    if (userStr) {
      try {
        return JSON.parse(userStr);
      } catch {
        return null;
      }
    }
    return null;
  },

  // Check if user is authenticated
  isAuthenticated(): boolean {
    return !!localStorage.getItem('token');
  },
};
