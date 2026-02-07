// User types
export interface User {
  id: number;
  username: string;
  email: string;
  quota: number;
  used_quota: number;
  is_admin: boolean;
  status: string;
  created_at: string;
  last_sign_in?: string;
}

// API Config types
export interface APIConfig {
  id: number;
  name: string;
  type: string;
  base_url: string;
  api_key?: string;
  models: string[];
  is_active: boolean;
  priority: number;
  weight: number;
  created_at: string;
  updated_at: string;
}

// Request Log types
export interface RequestLog {
  id: number;
  user_id: number;
  api_key_id: number;
  api_config_id: number;
  model: string;
  method: string;
  path: string;
  status_code: number;
  response_time: number;
  tokens_used: number;
  error_msg?: string;
  created_at: string;
}

// Stats types
export interface Stats {
  total_users: number;
  active_users: number;
  total_requests: number;
  today_requests: number;
}

// Auth types
export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  user: User;
}
