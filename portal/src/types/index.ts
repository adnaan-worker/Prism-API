// User types
export interface User {
  id: number;
  username: string;
  email: string;
  quota: number;
  used_quota: number;
  is_admin: boolean;
  status: string;
  last_sign_in?: string;
  created_at: string;
  updated_at: string;
}

// API Key types
export interface APIKey {
  id: number;
  user_id: number;
  key: string;
  name: string;
  is_active: boolean;
  rate_limit: number;
  last_used_at?: string;
  created_at: string;
  updated_at: string;
}

// Model types
export interface Model {
  name: string;
  provider: string;
  type: string;
  status: string;
  description: string;
  config_count: number;
}

// Auth types
export interface LoginRequest {
  username: string;
  password: string;
}

export interface RegisterRequest {
  username: string;
  email: string;
  password: string;
}

export interface AuthResponse {
  token: string;
  user: User;
}

// Quota types
export interface QuotaInfo {
  total_quota: number;
  used_quota: number;
  remaining_quota: number;
  last_sign_in?: string;
}

export interface SignInResponse {
  message: string;
  quota_awarded: number;
}

// API Response wrapper
export interface ApiResponse<T> {
  data?: T;
  error?: {
    code: number;
    message: string;
    details?: string;
  };
}
