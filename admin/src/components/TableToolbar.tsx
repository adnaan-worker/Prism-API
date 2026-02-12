import React from 'react';
import { Space, Button } from 'antd';
import { PlusOutlined, ReloadOutlined } from '@ant-design/icons';

interface TableToolbarProps {
  onAdd?: () => void;
  onRefresh?: () => void;
  addText?: string;
  showAdd?: boolean;
  showRefresh?: boolean;
  extra?: React.ReactNode;
}

/**
 * 表格工具栏组件
 * 统一的表格操作栏，包含添加、刷新等常用操作
 */
const TableToolbar: React.FC<TableToolbarProps> = ({
  onAdd,
  onRefresh,
  addText = '添加',
  showAdd = true,
  showRefresh = true,
  extra,
}) => {
  return (
    <Space  wrap>
      {showAdd && onAdd && (
        <Button type="primary" icon={<PlusOutlined />} onClick={onAdd}>
          {addText}
        </Button>
      )}
      {extra}
      {showRefresh && onRefresh && (
        <Button icon={<ReloadOutlined />} onClick={onRefresh}>
          刷新
        </Button>
      )}
    </Space>
  );
};

export default TableToolbar;
