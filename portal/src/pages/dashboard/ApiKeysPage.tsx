import { useState } from 'react';
import {
  Typography,
  Card,
  Table,
  Button,
  Modal,
  Form,
  Input,
  InputNumber,
  Space,
  Tag,
  message,
  Popconfirm,
  Tooltip,
  Alert,
  Statistic,
  Row,
  Col,
} from 'antd';
import {
  PlusOutlined,
  CopyOutlined,
  DeleteOutlined,
  KeyOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  EyeOutlined,
  EyeInvisibleOutlined,
} from '@ant-design/icons';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../lib/api';
import type { APIKey } from '../../types';
import type { ColumnsType } from 'antd/es/table';

const { Title, Paragraph, Text } = Typography;

interface CreateAPIKeyRequest {
  name: string;
  rate_limit: number;
}

interface CreateAPIKeyResponse {
  api_key: APIKey;
}

interface GetAPIKeysResponse {
  keys: APIKey[];
}

const ApiKeysPage = () => {
  const queryClient = useQueryClient();
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [form] = Form.useForm();
  const [visibleKeys, setVisibleKeys] = useState<Set<number>>(new Set());

  // Fetch API keys
  const { data: apiKeysData, isLoading } = useQuery({
    queryKey: ['apiKeys'],
    queryFn: async () => {
      const response = await apiClient.get<GetAPIKeysResponse>('/apikeys');
      return response.data;
    },
  });

  // Create API key mutation
  const createMutation = useMutation({
    mutationFn: async (values: CreateAPIKeyRequest) => {
      const response = await apiClient.post<CreateAPIKeyResponse>('/apikeys', values);
      return response.data;
    },
    onSuccess: (data) => {
      message.success('API key created successfully');
      Modal.success({
        title: 'Key Created Successfully',
        width: 600,
        content: (
          <div>
            <Alert
              message="Save your key now"
              description="For security reasons, the key will only be displayed once. You won't be able to see it again after closing this window."
              type="warning"
              showIcon
              style={{ marginBottom: 16 }}
            />
            <div style={{ marginTop: 16 }}>
              <Text strong>Key Name:&nbsp;</Text>
              <Text>{data.api_key.name}</Text>
            </div>
            <div style={{ marginTop: 8 }}>
              <Text strong>API Key:</Text>
              <Input.TextArea
                value={data.api_key.key}
                readOnly
                autoSize={{ minRows: 2, maxRows: 4 }}
                style={{ marginTop: 8, fontFamily: 'monospace' }}
              />
              <Button
                icon={<CopyOutlined />}
                onClick={() => {
                  navigator.clipboard.writeText(data.api_key.key);
                  message.success('Key copied to clipboard');
                }}
                style={{ marginTop: 8 }}
              >
                Copy Key
              </Button>
            </div>
          </div>
        ),
      });
      queryClient.invalidateQueries({ queryKey: ['apiKeys'] });
      setIsModalVisible(false);
      form.resetFields();
    },
    onError: (error: any) => {
      message.error(error.response?.data?.error?.message || 'Failed to create key');
    },
  });

  // Delete API key mutation
  const deleteMutation = useMutation({
    mutationFn: async (keyId: number) => {
      await apiClient.delete(`/apikeys/${keyId}`);
    },
    onSuccess: () => {
      message.success('API key deleted');
      queryClient.invalidateQueries({ queryKey: ['apiKeys'] });
    },
    onError: (error: any) => {
      message.error(error.response?.data?.error?.message || 'Failed to delete key');
    },
  });

  const handleCreate = () => {
    setIsModalVisible(true);
  };

  const handleModalOk = () => {
    form.validateFields().then((values) => {
      createMutation.mutate(values);
    });
  };

  const handleModalCancel = () => {
    setIsModalVisible(false);
    form.resetFields();
  };

  const handleDelete = (keyId: number) => {
    deleteMutation.mutate(keyId);
  };

  const handleCopyKey = (key: string) => {
    navigator.clipboard.writeText(key);
    message.success('Key copied to clipboard');
  };

  const toggleKeyVisibility = (keyId: number) => {
    setVisibleKeys((prev) => {
      const newSet = new Set(prev);
      if (newSet.has(keyId)) {
        newSet.delete(keyId);
      } else {
        newSet.add(keyId);
      }
      return newSet;
    });
  };

  const maskKey = (key: string) => {
    if (key.length <= 12) return key;
    return `${key.substring(0, 8)}...${key.substring(key.length - 4)}`;
  };

  const formatDate = (dateString?: string) => {
    if (!dateString) return '-';
    return new Date(dateString).toLocaleString();
  };

  const columns: ColumnsType<APIKey> = [
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      render: (name: string) => (
        <Space>
          <KeyOutlined className="text-primary" />
          <span className="font-medium text-text-primary">{name}</span>
        </Space>
      ),
    },
    {
      title: 'API Key',
      dataIndex: 'key',
      key: 'key',
      render: (key: string, record: APIKey) => (
        <Space>
          <span className="font-mono text-text-secondary bg-black/5 dark:bg-white/5 px-2 py-0.5 rounded text-sm">
            {visibleKeys.has(record.id) ? key : maskKey(key)}
          </span>
          <Tooltip title={visibleKeys.has(record.id) ? 'Hide' : 'Show'}>
            <Button
              type="text"
              size="small"
              icon={visibleKeys.has(record.id) ? <EyeInvisibleOutlined /> : <EyeOutlined />}
              onClick={() => toggleKeyVisibility(record.id)}
              className="text-text-tertiary hover:text-primary"
            />
          </Tooltip>
          <Tooltip title="Copy">
            <Button
              type="text"
              size="small"
              icon={<CopyOutlined />}
              onClick={() => handleCopyKey(key)}
              className="text-text-tertiary hover:text-primary"
            />
          </Tooltip>
        </Space>
      ),
    },
    {
      title: 'Status',
      dataIndex: 'is_active',
      key: 'is_active',
      render: (isActive: boolean) =>
        isActive ? (
          <Tag icon={<CheckCircleOutlined />} color="success" className="border-none bg-green-500/10 text-green-500">
            Active
          </Tag>
        ) : (
          <Tag icon={<CloseCircleOutlined />} color="error" className="border-none bg-red-500/10 text-red-500">
            Inactive
          </Tag>
        ),
    },
    {
      title: 'Rate Limit (RPM)',
      dataIndex: 'rate_limit',
      key: 'rate_limit',
      render: (rateLimit: number) => <span className="text-text-primary">{rateLimit}</span>,
    },
    {
      title: 'Created',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date: string) => <span className="text-text-secondary">{formatDate(date)}</span>,
    },
    {
      title: 'Last Used',
      dataIndex: 'last_used_at',
      key: 'last_used_at',
      render: (date?: string) => <span className="text-text-secondary">{formatDate(date)}</span>,
    },
    {
      title: 'Action',
      key: 'action',
      render: (_: any, record: APIKey) => (
        <Space>
          <Popconfirm
            title="Delete Key"
            description="Are you sure you want to delete this key? This action cannot be undone."
            onConfirm={() => handleDelete(record.id)}
            okText="Delete"
            cancelText="Cancel"
            okButtonProps={{ danger: true }}
          >
            <Button type="text" danger icon={<DeleteOutlined />} size="small">
              Delete
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  const apiKeys = apiKeysData?.keys || [];
  const activeKeysCount = apiKeys.filter((key) => key.is_active).length;
  const totalKeysCount = apiKeys.length;

  return (
    <div className="max-w-7xl mx-auto space-y-8 animate-fade-in pb-12">
      <div className="mb-8">
        <h2 className="text-3xl font-bold text-text-primary mb-2">API密钥管理</h2>
        <p className="text-text-secondary text-lg">
          创建和管理您的API密钥，用于调用平台提供的API服务。每个密钥都有独立的限流配置。
        </p>
      </div>

      {/* Statistics */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="glass-card p-6 rounded-2xl flex items-center gap-4">
          <div className="w-12 h-12 rounded-xl bg-blue-500/10 flex items-center justify-center text-blue-500 text-2xl">
            <KeyOutlined />
          </div>
          <div>
            <p className="text-text-tertiary text-sm">总密钥数</p>
            <p className="text-2xl font-bold text-text-primary">{totalKeysCount}</p>
          </div>
        </div>
        <div className="glass-card p-6 rounded-2xl flex items-center gap-4">
          <div className="w-12 h-12 rounded-xl bg-green-500/10 flex items-center justify-center text-green-500 text-2xl">
            <CheckCircleOutlined />
          </div>
          <div>
            <p className="text-text-tertiary text-sm">激活密钥</p>
            <p className="text-2xl font-bold text-text-primary">{activeKeysCount}</p>
          </div>
        </div>
        <div className="glass-card p-6 rounded-2xl flex items-center gap-4">
          <div className="w-12 h-12 rounded-xl bg-red-500/10 flex items-center justify-center text-red-500 text-2xl">
            <CloseCircleOutlined />
          </div>
          <div>
            <p className="text-text-tertiary text-sm">禁用密钥</p>
            <p className="text-2xl font-bold text-text-primary">{totalKeysCount - activeKeysCount}</p>
          </div>
        </div>
      </div>

      {/* API Keys Table */}
      <div className="glass-card p-6 rounded-2xl">
        <div className="flex items-center justify-between mb-6">
          <h3 className="text-xl font-bold text-text-primary flex items-center gap-2">
            <KeyOutlined /> 我的API密钥
          </h3>
          <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
            创建密钥
          </Button>
        </div>

        {apiKeys.length === 0 && !isLoading && (
          <Alert
            message="还没有API密钥"
            description="点击右上角的【创建密钥】按钮来创建您的第一个API密钥。"
            type="info"
            showIcon
            className="mb-6 bg-blue-500/5 border-blue-500/20"
          />
        )}

        <Table
          columns={columns}
          dataSource={apiKeys}
          rowKey="id"
          loading={isLoading}
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showTotal: (total) => `共 ${total} 个密钥`,
          }}
          className="glass-table"
        />
      </div>

      {/* Create API Key Modal - Keeping Ant Design Modal for now, behaves well enough */}
      <Modal
        title={
          <Space>
            <PlusOutlined />
            <span>创建API密钥</span>
          </Space>
        }
        open={isModalVisible}
        onOk={handleModalOk}
        onCancel={handleModalCancel}
        confirmLoading={createMutation.isPending}
        okText="创建"
        cancelText="取消"
        width={500}
      >
        <Alert
          message="安全提示"
          description="创建后密钥只会显示一次，请务必妥善保管。"
          type="warning"
          showIcon
          style={{ marginBottom: 16 }}
        />
        <Form
          form={form}
          layout="vertical"
          initialValues={{
            rate_limit: 60,
          }}
        >
          <Form.Item
            label="密钥名称"
            name="name"
            rules={[
              { required: true, message: '请输入密钥名称' },
              { min: 2, message: '密钥名称至少2个字符' },
              { max: 50, message: '密钥名称最多50个字符' },
            ]}
          >
            <Input placeholder="例如：生产环境密钥" />
          </Form.Item>
          <Form.Item
            label="限流（次/分钟）"
            name="rate_limit"
            rules={[
              { required: true, message: '请输入限流值' },
              { type: 'number', min: 1, message: '限流值至少为1' },
              { type: 'number', max: 1000, message: '限流值最多为1000' },
            ]}
            tooltip="设置此密钥每分钟最多可以发起的请求次数"
          >
            <InputNumber
              style={{ width: '100%' }}
              placeholder="60"
              min={1}
              max={1000}
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default ApiKeysPage;
