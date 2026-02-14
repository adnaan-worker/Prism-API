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
  config_type?: string; // 'direct' or 'account_pool'
  account_pool_id?: number;
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

// Pricing types
export interface Pricing {
  id: number;
  api_config_id: number;
  api_config?: APIConfig;
  model_name: string;
  input_price: number;
  output_price: number;
  currency: string;
  unit: number;
  is_active: boolean;
  description?: string;
  created_at: string;
  updated_at: string;
}

// Account Pool types
export interface AccountPool {
  id: number;
  name: string;
  description?: string;
  provider: string;
  strategy: string;
  health_check_interval: number;
  health_check_timeout: number;
  max_retries: number;
  is_active: boolean;
  total_requests: number;
  total_errors: number;
  created_at: string;
  updated_at: string;
}

export interface AccountCredential {
  id: number;
  created_at: string;
  updated_at: string;
  pool_id: number;
  provider: string;
  auth_type: string;
  account_name?: string;
  account_email?: string;
  weight: number;
  is_active: boolean;

  // 账号状态
  status: string; // active, expired, error, refreshing, unknown
  last_error?: string;

  // 健康状态
  health_status: string;
  last_checked_at?: string;
  last_used_at?: string;

  // 统计
  total_requests: number;
  total_errors: number;
  error_rate: number;

  // 订阅信息
  subscription_type?: string; // Free, Pro, Pro_Plus, Enterprise
  subscription_title?: string;
  subscription_expires_at?: string;
  subscription_days_remaining?: number;

  // 使用量详情
  usage_current: number;
  usage_limit: number;
  usage_percent: number;
  usage_last_updated?: string;
  base_limit: number;
  base_current: number;
  free_trial_limit: number;
  free_trial_current: number;
  free_trial_expiry?: string;
  next_reset_date?: string;

  // 机器码
  machine_id?: string;

  // OAuth 相关
  expires_at?: string;
  is_expired: boolean;

  // 速率限制
  rate_limit: number;
  current_usage: number;

  // 兼容字段
  name?: string;
  credentials_data?: any;
  request_count?: number;
}

export interface CreateCredentialRequest {
  pool_id: number;
  provider: string;
  auth_type: 'api_key' | 'oauth' | 'session_token';
  api_key?: string;
  access_token?: string;
  refresh_token?: string;
  session_token?: string;
  account_name?: string;
  account_email?: string;
  weight?: number;
  rate_limit?: number;
}

export interface PoolStats {
  pool_id: number;
  pool_name: string;
  provider: string;
  total_creds: number;
  active_creds: number;
  total_requests: number;
}

// List response types
export interface CredentialListResponse {
  credentials: AccountCredential[];
  total: number;
}

export interface PoolListResponse {
  pools: AccountPool[];
  total: number;
}
