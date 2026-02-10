import React from 'react';
import { Tag } from 'antd';

interface ProviderTagProps {
  provider: string;
}

const PROVIDER_COLORS: Record<string, string> = {
  openai: 'blue',
  anthropic: 'orange',
  gemini: 'green',
  custom: 'purple',
};

/**
 * 厂商标签组件
 * 统一的厂商类型显示标签
 */
const ProviderTag: React.FC<ProviderTagProps> = ({ provider }) => {
  const color = PROVIDER_COLORS[provider.toLowerCase()] || 'default';
  
  return (
    <Tag color={color}>
      {provider.toUpperCase()}
    </Tag>
  );
};

export default ProviderTag;
