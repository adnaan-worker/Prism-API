/**
 * 全局常量定义
 */

// API 厂商类型
export const PROVIDER_TYPES = {
  OPENAI: 'openai',
  ANTHROPIC: 'anthropic',
  GEMINI: 'gemini',
  CUSTOM: 'custom',
} as const;

// 厂商颜色映射
export const PROVIDER_COLORS: Record<string, string> = {
  [PROVIDER_TYPES.OPENAI]: 'blue',
  [PROVIDER_TYPES.ANTHROPIC]: 'orange',
  [PROVIDER_TYPES.GEMINI]: 'green',
  [PROVIDER_TYPES.CUSTOM]: 'purple',
};

// 厂商选项
export const PROVIDER_OPTIONS = [
  { label: 'OpenAI', value: PROVIDER_TYPES.OPENAI },
  { label: 'Anthropic', value: PROVIDER_TYPES.ANTHROPIC },
  { label: 'Gemini', value: PROVIDER_TYPES.GEMINI },
  { label: 'Custom', value: PROVIDER_TYPES.CUSTOM },
];

// 用户状态
export const USER_STATUS = {
  ACTIVE: 'active',
  DISABLED: 'disabled',
  BANNED: 'banned',
} as const;

// 用户状态选项
export const USER_STATUS_OPTIONS = [
  { label: '正常', value: USER_STATUS.ACTIVE },
  { label: '禁用', value: USER_STATUS.DISABLED },
  { label: '封禁', value: USER_STATUS.BANNED },
];

// 负载均衡策略
export const LB_STRATEGIES = {
  ROUND_ROBIN: 'round_robin',
  RANDOM: 'random',
  WEIGHTED: 'weighted',
  LEAST_LOAD: 'least_load',
} as const;

// 负载均衡策略选项
export const LB_STRATEGY_OPTIONS = [
  { label: '轮询', value: LB_STRATEGIES.ROUND_ROBIN },
  { label: '随机', value: LB_STRATEGIES.RANDOM },
  { label: '加权轮询', value: LB_STRATEGIES.WEIGHTED },
  { label: '最少负载', value: LB_STRATEGIES.LEAST_LOAD },
];

// 货币单位
export const CURRENCY_OPTIONS = [
  { label: '积分 (Credits)', value: 'credits' },
  { label: '美元 (USD)', value: 'usd' },
  { label: '人民币 (CNY)', value: 'cny' },
];

// 分页默认值
export const DEFAULT_PAGE_SIZE = 10;
export const PAGE_SIZE_OPTIONS = ['10', '20', '50', '100'];

// HTTP 状态码颜色
export const STATUS_CODE_COLORS: Record<number, string> = {
  200: 'success',
  201: 'success',
  400: 'warning',
  401: 'error',
  403: 'error',
  404: 'warning',
  429: 'warning',
  500: 'error',
};
