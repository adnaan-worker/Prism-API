import { useState } from 'react';
import { DEFAULT_PAGE_SIZE } from '../utils/constants';

/**
 * 表格状态管理 Hook
 * 统一管理表格的分页、筛选等状态
 */
export function useTable(initialPageSize: number = DEFAULT_PAGE_SIZE) {
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(initialPageSize);
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);

  const handlePageChange = (newPage: number, newPageSize: number) => {
    setPage(newPage);
    setPageSize(newPageSize);
  };

  const handleSelectionChange = (keys: React.Key[]) => {
    setSelectedRowKeys(keys);
  };

  const clearSelection = () => {
    setSelectedRowKeys([]);
  };

  const resetPagination = () => {
    setPage(1);
  };

  return {
    page,
    pageSize,
    selectedRowKeys,
    setPage,
    setPageSize,
    handlePageChange,
    handleSelectionChange,
    clearSelection,
    resetPagination,
  };
}
