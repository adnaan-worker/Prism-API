import React from 'react';
import { Tag } from 'antd';

interface StatusTagProps {
  status: string;
  activeText?: string;
  inactiveText?: string;
}

/**
 * 状态标签组件
 * 统一的状态显示标签
 */
const StatusTag: React.FC<StatusTagProps> = ({
  status,
  activeText = '启用',
  inactiveText = '禁用',
}) => {
  const isActive = status === 'active' || status === true;
  
  return (
    <Tag color={isActive ? 'success' : 'default'}>
      {isActive ? activeText : inactiveText}
    </Tag>
  );
};

export default StatusTag;
