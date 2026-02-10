/**
 * 格式化工具函数
 */

/**
 * 格式化数字，添加千分位分隔符
 */
export const formatNumber = (num: number): string => {
  return num.toLocaleString();
};

/**
 * 格式化日期时间
 */
export const formatDateTime = (date: string | Date): string => {
  if (!date) return '-';
  return new Date(date).toLocaleString('zh-CN');
};

/**
 * 格式化日期
 */
export const formatDate = (date: string | Date): string => {
  if (!date) return '-';
  return new Date(date).toLocaleDateString('zh-CN');
};

/**
 * 格式化文件大小
 */
export const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
};

/**
 * 格式化百分比
 */
export const formatPercent = (value: number, total: number): string => {
  if (total === 0) return '0%';
  return ((value / total) * 100).toFixed(2) + '%';
};

/**
 * 截断文本
 */
export const truncateText = (text: string, maxLength: number): string => {
  if (text.length <= maxLength) return text;
  return text.substring(0, maxLength) + '...';
};

/**
 * 格式化 credits（积分）
 */
export const formatCredits = (credits: number): string => {
  return credits.toLocaleString() + ' credits';
};
