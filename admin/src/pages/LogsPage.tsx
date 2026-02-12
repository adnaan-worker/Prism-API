import React, { useState } from 'react';
import {
  Card,
  Table,
  Space,
  Tag,
  DatePicker,
  Select,
  Input,
  Button,
  Drawer,
  Descriptions,
  message,
} from 'antd';
import {
  SearchOutlined,
  ReloadOutlined,
  DownloadOutlined,
  EyeOutlined,
} from '@ant-design/icons';
import { useQuery } from '@tanstack/react-query';
import { logService } from '../services/logService';
import type { RequestLog } from '../types';
import type { ColumnsType } from 'antd/es/table';
import dayjs, { Dayjs } from 'dayjs';
import PageContainer from '../components/PageContainer';

const { RangePicker } = DatePicker;
const { Option } = Select;

const LogsPage: React.FC = () => {
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [dateRange, setDateRange] = useState<[Dayjs, Dayjs] | null>(null);
  const [userIdFilter, setUserIdFilter] = useState<string>('');
  const [modelFilter, setModelFilter] = useState<string | undefined>();
  const [statusFilter, setStatusFilter] = useState<'success' | 'error' | undefined>();
  const [selectedLog, setSelectedLog] = useState<RequestLog | null>(null);
  const [drawerVisible, setDrawerVisible] = useState(false);

  // 获取请求日志列表
  const { data, isLoading, refetch } = useQuery({
    queryKey: [
      'logs',
      page,
      pageSize,
      dateRange,
      userIdFilter,
      modelFilter,
      statusFilter,
    ],
    queryFn: () =>
      logService.getLogs({
        page,
        page_size: pageSize,
        start_time: dateRange?.[0]?.toISOString(),
        end_time: dateRange?.[1]?.toISOString(),
        user_id: userIdFilter ? parseInt(userIdFilter) : undefined,
        model: modelFilter,
        status: statusFilter,
      }),
  });

  // 表格列配置
  const columns: ColumnsType<RequestLog> = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
    },
    {
      title: '时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (date: string) => new Date(date).toLocaleString('zh-CN'),
    },
    {
      title: '用户ID',
      dataIndex: 'user_id',
      key: 'user_id',
      width: 100,
    },
    {
      title: '模型',
      dataIndex: 'model',
      key: 'model',
      width: 150,
      render: (model: string) => <Tag color="blue">{model}</Tag>,
    },
    {
      title: '方法',
      dataIndex: 'method',
      key: 'method',
      width: 80,
      render: (method: string) => {
        const colorMap: Record<string, string> = {
          GET: 'green',
          POST: 'blue',
          PUT: 'orange',
          DELETE: 'red',
        };
        return <Tag color={colorMap[method] || 'default'}>{method}</Tag>;
      },
    },
    {
      title: '路径',
      dataIndex: 'path',
      key: 'path',
      ellipsis: true,
    },
    {
      title: '状态码',
      dataIndex: 'status_code',
      key: 'status_code',
      width: 100,
      render: (code: number) => {
        let color = 'default';
        if (code >= 200 && code < 300) color = 'success';
        else if (code >= 400 && code < 500) color = 'warning';
        else if (code >= 500) color = 'error';
        return <Tag color={color}>{code}</Tag>;
      },
    },
    {
      title: '响应时间',
      dataIndex: 'response_time',
      key: 'response_time',
      width: 120,
      render: (time: number) => `${time}ms`,
      sorter: (a, b) => a.response_time - b.response_time,
    },
    {
      title: 'Tokens',
      dataIndex: 'tokens_used',
      key: 'tokens_used',
      width: 100,
      render: (tokens: number) => tokens.toLocaleString(),
    },
    {
      title: '状态',
      key: 'status',
      width: 80,
      render: (_, record) => {
        const isSuccess = record.status_code >= 200 && record.status_code < 300;
        return (
          <Tag color={isSuccess ? 'success' : 'error'}>
            {isSuccess ? '成功' : '失败'}
          </Tag>
        );
      },
      filters: [
        { text: '成功', value: 'success' },
        { text: '失败', value: 'error' },
      ],
      onFilter: (value, record) => {
        const isSuccess = record.status_code >= 200 && record.status_code < 300;
        return value === 'success' ? isSuccess : !isSuccess;
      },
    },
    {
      title: '操作',
      key: 'action',
      fixed: 'right',
      width: 100,
      render: (_, record) => (
        <Button
          type="link"
          size="small"
          icon={<EyeOutlined />}
          onClick={() => handleViewLog(record)}
        >
          详情
        </Button>
      ),
    },
  ];

  // 查看日志详情
  const handleViewLog = (log: RequestLog) => {
    setSelectedLog(log);
    setDrawerVisible(true);
  };

  // 导出日志
  const handleExport = async () => {
    try {
      message.loading({ content: '正在导出...', key: 'export' });
      const blob = await logService.exportLogs({
        start_time: dateRange?.[0]?.toISOString(),
        end_time: dateRange?.[1]?.toISOString(),
        user_id: userIdFilter ? parseInt(userIdFilter) : undefined,
        model: modelFilter,
        status: statusFilter,
      });

      // 创建下载链接
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `logs_${dayjs().format('YYYY-MM-DD_HH-mm-ss')}.csv`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      window.URL.revokeObjectURL(url);

      message.success({ content: '导出成功', key: 'export' });
    } catch (error) {
      message.error({ content: '导出失败', key: 'export' });
    }
  };

  // 重置筛选
  const handleReset = () => {
    setDateRange(null);
    setUserIdFilter('');
    setModelFilter(undefined);
    setStatusFilter(undefined);
    setPage(1);
  };

  return (
    <PageContainer title="请求日志" description="查看和分析 API 请求记录">
      <Card>
        {/* 筛选栏 */}
        <Space style={{ marginBottom: 16 }} wrap>
          <RangePicker
            showTime
            value={dateRange}
            onChange={(dates) => {
              setDateRange(dates as [Dayjs, Dayjs] | null);
              setPage(1);
            }}
            placeholder={['开始时间', '结束时间']}
            style={{ width: 380 }}
          />

          <Input
            placeholder="用户ID"
            value={userIdFilter}
            onChange={(e) => {
              setUserIdFilter(e.target.value);
              setPage(1);
            }}
            style={{ width: 120 }}
            allowClear
          />

          <Select
            placeholder="筛选模型"
            value={modelFilter}
            onChange={(value) => {
              setModelFilter(value);
              setPage(1);
            }}
            style={{ width: 180 }}
            allowClear
          >
            <Option value="gpt-4">gpt-4</Option>
            <Option value="gpt-3.5-turbo">gpt-3.5-turbo</Option>
            <Option value="claude-3-opus">claude-3-opus</Option>
            <Option value="claude-3-sonnet">claude-3-sonnet</Option>
            <Option value="gemini-pro">gemini-pro</Option>
          </Select>

          <Select
            placeholder="筛选状态"
            value={statusFilter}
            onChange={(value) => {
              setStatusFilter(value);
              setPage(1);
            }}
            style={{ width: 120 }}
            allowClear
          >
            <Option value="success">成功</Option>
            <Option value="error">失败</Option>
          </Select>

          <Button icon={<SearchOutlined />} type="primary" onClick={() => refetch()}>
            查询
          </Button>

          <Button icon={<ReloadOutlined />} onClick={handleReset}>
            重置
          </Button>

          <Button icon={<DownloadOutlined />} onClick={handleExport}>
            导出
          </Button>
        </Space>

        {/* 日志列表表格 */}
        <Table
          columns={columns}
          dataSource={data?.logs || []}
          rowKey="id"
          loading={isLoading}
          scroll={{ x: 1400 }}
          pagination={{
            current: page,
            pageSize: pageSize,
            total: data?.total || 0,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 条记录`,
            onChange: (newPage, newPageSize) => {
              setPage(newPage);
              setPageSize(newPageSize);
            },
          }}
        />
      </Card>

      {/* 日志详情抽屉 */}
      <Drawer
        title="请求日志详情"
        placement="right"
        width={700}
        onClose={() => setDrawerVisible(false)}
        open={drawerVisible}
      >
        {selectedLog && (
          <Descriptions column={1} bordered>
            <Descriptions.Item label="日志ID">{selectedLog.id}</Descriptions.Item>
            <Descriptions.Item label="时间">
              {new Date(selectedLog.created_at).toLocaleString('zh-CN')}
            </Descriptions.Item>
            <Descriptions.Item label="用户ID">
              {selectedLog.user_id}
            </Descriptions.Item>
            <Descriptions.Item label="API Key ID">
              {selectedLog.api_key_id}
            </Descriptions.Item>
            <Descriptions.Item label="API配置ID">
              {selectedLog.api_config_id}
            </Descriptions.Item>
            <Descriptions.Item label="模型">
              <Tag color="blue">{selectedLog.model}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="请求方法">
              <Tag>{selectedLog.method}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="请求路径">
              {selectedLog.path}
            </Descriptions.Item>
            <Descriptions.Item label="状态码">
              <Tag
                color={
                  selectedLog.status_code >= 200 && selectedLog.status_code < 300
                    ? 'success'
                    : 'error'
                }
              >
                {selectedLog.status_code}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="响应时间">
              {selectedLog.response_time}ms
            </Descriptions.Item>
            <Descriptions.Item label="使用Tokens">
              {selectedLog.tokens_used.toLocaleString()}
            </Descriptions.Item>
            {selectedLog.error_msg && (
              <Descriptions.Item label="错误信息">
                <pre
                  style={{
                    background: '#f5f5f5',
                    padding: '8px',
                    borderRadius: '4px',
                    whiteSpace: 'pre-wrap',
                    wordBreak: 'break-word',
                  }}
                >
                  {selectedLog.error_msg}
                </pre>
              </Descriptions.Item>
            )}
          </Descriptions>
        )}
      </Drawer>
    </PageContainer>
  );
};

export default LogsPage;
