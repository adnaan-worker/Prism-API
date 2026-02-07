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
      message.success('API密钥创建成功！');
      // Show the full key in a modal
      Modal.success({
        title: '密钥创建成功',
        width: 600,
        content: (
          <div>
            <Alert
              message="请立即复制并保存您的密钥"
              description="出于安全考虑，密钥只会显示一次。关闭此窗口后将无法再次查看完整密钥。"
              type="warning"
              showIcon
              style={{ marginBottom: 16 }}
            />
            <div style={{ marginTop: 16 }}>
              <Text strong>密钥名称：</Text>
              <Text>{data.api_key.name}</Text>
            </div>
            <div style={{ marginTop: 8 }}>
              <Text strong>API密钥：</Text>
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
                  message.success('密钥已复制到剪贴板');
                }}
                style={{ marginTop: 8 }}
              >
                复制密钥
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
      message.error(error.response?.data?.error?.message || '创建失败，请稍后重试');
    },
  });

  // Delete API key mutation
  const deleteMutation = useMutation({
    mutationFn: async (keyId: number) => {
      await apiClient.delete(`/apikeys/${keyId}`);
    },
    onSuccess: () => {
      message.success('API密钥已删除');
      queryClient.invalidateQueries({ queryKey: ['apiKeys'] });
    },
    onError: (error: any) => {
      message.error(error.response?.data?.error?.message || '删除失败，请稍后重试');
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
    message.success('密钥已复制到剪贴板');
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
    return new Date(dateString).toLocaleString('zh-CN');
  };

  const columns: ColumnsType<APIKey> = [
    {
      title: '密钥名称',
      dataIndex: 'name',
      key: 'name',
      render: (name: string) => (
        <Space>
          <KeyOutlined />
          <Text strong>{name}</Text>
        </Space>
      ),
    },
    {
      title: 'API密钥',
      dataIndex: 'key',
      key: 'key',
      render: (key: string, record: APIKey) => (
        <Space>
          <Text code style={{ fontFamily: 'monospace' }}>
            {visibleKeys.has(record.id) ? key : maskKey(key)}
          </Text>
          <Tooltip title={visibleKeys.has(record.id) ? '隐藏' : '显示'}>
            <Button
              type="text"
              size="small"
              icon={visibleKeys.has(record.id) ? <EyeInvisibleOutlined /> : <EyeOutlined />}
              onClick={() => toggleKeyVisibility(record.id)}
            />
          </Tooltip>
          <Tooltip title="复制密钥">
            <Button
              type="text"
              size="small"
              icon={<CopyOutlined />}
              onClick={() => handleCopyKey(key)}
            />
          </Tooltip>
        </Space>
      ),
    },
    {
      title: '状态',
      dataIndex: 'is_active',
      key: 'is_active',
      render: (isActive: boolean) =>
        isActive ? (
          <Tag icon={<CheckCircleOutlined />} color="success">
            激活
          </Tag>
        ) : (
          <Tag icon={<CloseCircleOutlined />} color="error">
            禁用
          </Tag>
        ),
    },
    {
      title: '限流（次/分钟）',
      dataIndex: 'rate_limit',
      key: 'rate_limit',
      render: (rateLimit: number) => <Text>{rateLimit}</Text>,
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date: string) => <Text type="secondary">{formatDate(date)}</Text>,
    },
    {
      title: '最后使用',
      dataIndex: 'last_used_at',
      key: 'last_used_at',
      render: (date?: string) => <Text type="secondary">{formatDate(date)}</Text>,
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: APIKey) => (
        <Space>
          <Popconfirm
            title="确认删除"
            description="删除后将无法恢复，确定要删除此密钥吗？"
            onConfirm={() => handleDelete(record.id)}
            okText="确认"
            cancelText="取消"
          >
            <Button type="text" danger icon={<DeleteOutlined />} size="small">
              删除
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
    <div style={{ padding: '0 24px' }}>
      <div style={{ marginBottom: 24 }}>
        <Title level={2}>API密钥管理</Title>
        <Paragraph type="secondary">
          创建和管理您的API密钥，用于调用平台提供的API服务。每个密钥都有独立的限流配置。
        </Paragraph>
      </div>

      {/* Statistics */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={12} lg={8}>
          <Card>
            <Statistic
              title="总密钥数"
              value={totalKeysCount}
              prefix={<KeyOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={8}>
          <Card>
            <Statistic
              title="激活密钥"
              value={activeKeysCount}
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={8}>
          <Card>
            <Statistic
              title="禁用密钥"
              value={totalKeysCount - activeKeysCount}
              prefix={<CloseCircleOutlined />}
              valueStyle={{ color: '#ff4d4f' }}
            />
          </Card>
        </Col>
      </Row>

      {/* API Keys Table */}
      <Card
        title={
          <Space>
            <KeyOutlined />
            <span>我的API密钥</span>
          </Space>
        }
        extra={
          <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
            创建密钥
          </Button>
        }
      >
        {apiKeys.length === 0 && !isLoading && (
          <Alert
            message="还没有API密钥"
            description="点击右上角的【创建密钥】按钮来创建您的第一个API密钥。"
            type="info"
            showIcon
            style={{ marginBottom: 16 }}
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
        />
      </Card>

      {/* Create API Key Modal */}
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
