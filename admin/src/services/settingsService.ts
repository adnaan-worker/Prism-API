import { apiClient } from '../lib/api';

// ========== 类型定义 ==========

/** 运行时配置（第一优先级） */
export interface RuntimeConfig {
  cache_enabled: boolean;
  cache_ttl: string; // e.g. "24h", "1h30m"
  semantic_cache_enabled: boolean;
  semantic_threshold: number; // 0.0 ~ 1.0
  embedding_enabled: boolean;
}

/** 运行时配置更新请求 */
export type RuntimeConfigUpdate = Partial<RuntimeConfig>;

/** 系统运行信息（第二优先级） */
export interface SystemRunningConfig {
  // 缓存配置
  cache_enabled: boolean;
  cache_ttl: string;
  semantic_cache_enabled: boolean;
  semantic_threshold: number;
  embedding_enabled: boolean;
  // 速率限制
  rate_limit_enabled: boolean;
  rate_limit_requests: number;
  rate_limit_window: string;
  // 服务信息
  version: string;
  uptime: string;
  go_version: string;
}

/** 修改密码请求（第三优先级） */
export interface ChangePasswordRequest {
  old_password: string;
  new_password: string;
}

/** 默认配额设置 */
export interface DefaultQuotaConfig {
  default_quota: number;
}

/** 默认速率限制设置 */
export interface DefaultRateLimitConfig {
  requests_per_minute: number;
  requests_per_day: number;
}

// ========== API 调用 ==========

export const settingsService = {
  // ---- 第一优先级：运行时配置 ----

  /** 获取当前运行时配置 */
  getRuntimeConfig: async (): Promise<RuntimeConfig> => {
    const response = await apiClient.get<RuntimeConfig>('/admin/settings/runtime');
    return response.data;
  },

  /** 更新运行时配置 */
  updateRuntimeConfig: async (data: RuntimeConfigUpdate): Promise<RuntimeConfig> => {
    const response = await apiClient.put<RuntimeConfig>('/admin/settings/runtime', data);
    return response.data;
  },

  // ---- 第二优先级：运维可见性 ----

  /** 获取系统完整运行配置 */
  getSystemConfig: async (): Promise<SystemRunningConfig> => {
    const response = await apiClient.get<SystemRunningConfig>('/admin/settings/system');
    return response.data;
  },

  // ---- 第三优先级：安全管理 ----

  /** 修改管理员密码 */
  changePassword: async (data: ChangePasswordRequest): Promise<void> => {
    await apiClient.put('/admin/settings/password', data);
  },

  /** 获取默认用户配额 */
  getDefaultQuota: async (): Promise<DefaultQuotaConfig> => {
    const response = await apiClient.get<DefaultQuotaConfig>('/admin/settings/default-quota');
    return response.data;
  },

  /** 更新默认用户配额 */
  updateDefaultQuota: async (data: DefaultQuotaConfig): Promise<DefaultQuotaConfig> => {
    const response = await apiClient.put<DefaultQuotaConfig>('/admin/settings/default-quota', data);
    return response.data;
  },

  /** 获取默认速率限制 */
  getDefaultRateLimit: async (): Promise<DefaultRateLimitConfig> => {
    const response = await apiClient.get<DefaultRateLimitConfig>('/admin/settings/default-rate-limit');
    return response.data;
  },

  /** 更新默认速率限制 */
  updateDefaultRateLimit: async (data: DefaultRateLimitConfig): Promise<DefaultRateLimitConfig> => {
    const response = await apiClient.put<DefaultRateLimitConfig>('/admin/settings/default-rate-limit', data);
    return response.data;
  },
};
