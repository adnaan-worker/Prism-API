import React from 'react';
import {
  Card,
  Table,
  Button,
  Space,
  Modal,
  Form,
  Input,
  Select,
  InputNumber,
  message,
  Popconfirm,
  Switch,
} from 'antd';
import {
  EditOutlined,
  DeleteOutlined,
  ApiOutlined,
} from '@ant-design/icons';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiConfigService } from '../services/apiConfigService';
import type { APIConfig } from '../types';
import type { ColumnsType } from 'antd/es/table';
import TableToolbar from '../components/TableToolbar';
import ProviderTag from '../components/ProviderTag';
import { useTable } from '../hooks/useTable';
import { useModal } from '../hooks/useModal';
import { formatDateTime } from '../utils/format';

const { Option } = Select;
const { TextArea } = Input;

const ApiConfigsPage: React.FC = () => {
  const { page, pageSize, selectedRowKeys, handlePageChange, handleSelectionChange, clearSelection, resetPagination } = useTable();
  const [typeFilter, setTypeFilter] = React.useState<string | undefined>();
  const configModal = useModal<APIConfig>();
  const [fetchingModels, setFetchingModels] = React.useState(false);

  const queryClient = useQueryClient();

  // 获取API配置列表
  const { data, isLoading, refetch } = useQuery({
    queryKey: ['api-configs', page, pageSize, typeFilter],
    queryFn: () =>
      apiConfigService.getConfigs({
        page,
        page_size: pageSize,
        type: typeFilter,
      }),
  });

  // 创建API配置
  const createMutation = useMutation({
    mutationFn: apiConfigService.createConfig,
    onSuccess: () => {
      message.success('API配置创建成功');
      queryClient.invalidateQueries({ queryKey: ['api-configs'] });
      configModal.hideModal();
    },
    onError: () => {
      message.error('API配置创建失败');
    },
  });

  // 更新API配置
  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: any }) =>
      apiConfigService.updateConfig(id, data),
    onSuccess: () => {
      message.success('API配置更新成功');
      queryClient.invalidateQueries({ queryKey: ['api-configs'] });
      configModal.hideModal();
    },
    onError: () => {
      message.error('API配置更新失败');
    },
  });

  // 删除API配置
  const deleteMutation = useMutation({
    mutationFn: apiConfigService.deleteConfig,
    onSuccess: () => {
      message.success('API配置删除成功');
      queryClient.invalidateQueries({ queryKey: ['api-configs'] });
    },
    onError: () => {
      message.error('API配置删除失败');
    },
  });

  // 切换配置状态
  const toggleStatusMutation = useMutation({
    mutationFn: ({ id, is_active }: { id: number; is_active: boolean }) =>
      apiConfigService.toggleConfigStatus(id, is_active),
    onSuccess: () => {
      message.success('配置状态更新成功');
      queryClient.invalidateQueries({ queryKey: ['api-configs'] });
    },
    onError: () => {
      message.error('配置状态更新失败');
    },
  });

  // 批量删除
  const batchDeleteMutation = useMutation({
    mutationFn: apiConfigService.batchDeleteConfigs,
    onSuccess: () => {
      message.success('批量删除成功');
      queryClient.invalidateQueries({ queryKey: ['api-configs'] });
      clearSelection();
    },
    onError: () => {
      message.error('批量删除失败');
    },
  });

  // 表格列配置
  const columns: ColumnsType<APIConfig> = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
    },
    {
      title: '配置名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string) => (
        <Space>
          <ApiOutlined />
          {text}
        </Space>
      ),
    },
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      render: (type: string) => <ProviderTag provider={type} />,
      filters: [
        { text: 'OpenAI', value: 'openai' },
        { text: 'Anthropic', value: 'anthropic' },
        { text: 'Gemini', value: 'gemini' },
        { text: 'Custom', value: 'custom' },
      ],
      onFilter: (value, record) => record.type === value,
    },
    {
      title: 'Base URL',
      dataIndex: 'base_url',
      key: 'base_url',
      ellipsis: true,
    },
    {
      title: '支持模型',
      dataIndex: 'models',
      key: 'models',
      render: (models: string[]) => (
        <Space wrap>
          {models.slice(0, 3).map((model) => (
            <Tag key={model}>{model}</Tag>
          ))}
          {models.length > 3 && <Tag>+{models.length - 3}</Tag>}
        </Space>
      ),
    },
    {
      title: '优先级',
      dataIndex: 'priority',
      key: 'priority',
      width: 100,
      sorter: (a, b) => a.priority - b.priority,
    },
    {
      title: '权重',
      dataIndex: 'weight',
      key: 'weight',
      width: 80,
    },
    {
      title: '状态',
      dataIndex: 'is_active',
      key: 'is_active',
      width: 100,
      render: (is_active: boolean, record) => (
        <Switch
          checked={is_active}
          onChange={(checked) =>
            toggleStatusMutation.mutate({ id: record.id, is_active: checked })
          }
          checkedChildren="启用"
          unCheckedChildren="禁用"
        />
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date: string) => formatDateTime(date),
    },
    {
      title: '操作',
      key: 'action',
      fixed: 'right',
      width: 150,
      render: (_, record) => (
        <Space>
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确定要删除该配置吗？"
            onConfirm={() => deleteMutation.mutate(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Button type="link" size="small" danger icon={<DeleteOutlined />}>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  // 打开添加模态框
  const handleAdd = () => {
    configModal.showModal();
    configModal.form.setFieldsValue({
      priority: 100,
      weight: 1,
      max_rps: 0,
      timeout: 30,
    });
  };

  // 打开编辑模态框
  const handleEdit = (config: APIConfig) => {
    configModal.showModal(config);
    configModal.form.setFieldsValue({
      ...config,
      models: config.models.join('\n'),
    });
  };

  // 获取模型列表
  const handleFetchModels = async () => {
    const type = configModal.form.getFieldValue('type');
    const baseUrl = configModal.form.getFieldValue('base_url');
    const apiKey = configModal.form.getFieldValue('api_key');

    if (!type) {
      message.warning('请先选择API类型');
      return;
    }
    if (!baseUrl) {
      message.warning('请先输入Base URL');
      return;
    }
    if (!apiKey) {
      message.warning('请先输入API Key');
      return;
    }

    setFetchingModels(true);
    try {
      const response = await apiConfigService.fetchModels({
        type,
        base_url: baseUrl,
        api_key: apiKey,
      });
      
      if (response.models && response.models.length > 0) {
        configModal.form.setFieldsValue({
          models: response.models.join('\n'),
        });
        message.success(`成功获取 ${response.models.length} 个模型`);
      } else {
        message.info('未获取到模型列表');
      }
    } catch (error: any) {
      message.error(error.response?.data?.error?.message || '获取模型失败');
    } finally {
      setFetchingModels(false);
    }
  };

  // 提交表单
  const handleSubmit = () => {
    configModal.form.validateFields().then((values) => {
      const modelsArray =
        typeof values.models === 'string'
          ? values.models.split('\n').filter((m: string) => m.trim())
          : values.models;

      const data = {
        ...values,
        models: modelsArray,
      };

      if (configModal.editingItem) {
        updateMutation.mutate({ id: configModal.editingItem.id, data });
      } else {
        createMutation.mutate(data);
      }
    });
  };

  // 批量删除
  const handleBatchDelete = () => {
    if (selectedRowKeys.length === 0) {
      message.warning('请选择要删除的配置');
      return;
    }
    Modal.confirm({
      title: '批量删除确认',
      content: `确定要删除选中的 ${selectedRowKeys.length} 个配置吗？`,
      onOk: () => {
        batchDeleteMutation.mutate(selectedRowKeys as number[]);
      },
    });
  };

  // 行选择配置
  const rowSelection = {
    selectedRowKeys,
    onChange: handleSelectionChange,
  };

  return (
    <div>
      <Card>
        {/* 操作栏 */}
        <TableToolbar
          onAdd={handleAdd}
          addText="添加配置"
          onRefresh={() => refetch()}
          extra={
            <>
              <Button
                danger
                icon={<DeleteOutlined />}
                onClick={handleBatchDelete}
                disabled={selectedRowKeys.length === 0}
              >
                批量删除 ({selectedRowKeys.length})
              </Button>
              <Select
                placeholder="筛选类型"
                allowClear
                style={{ width: 150 }}
                onChange={(value) => {
                  setTypeFilter(value);
                  resetPagination();
                }}
              >
                <Option value="openai">OpenAI</Option>
                <Option value="anthropic">Anthropic</Option>
                <Option value="gemini">Gemini</Option>
                <Option value="custom">Custom</Option>
              </Select>
            </>
          }
        />

        {/* 配置列表表格 */}
        <Table
          columns={columns}
          dataSource={data?.configs || []}
          rowKey="id"
          loading={isLoading}
          rowSelection={rowSelection}
          scroll={{ x: 1400 }}
          pagination={{
            current: page,
            pageSize: pageSize,
            total: data?.total || 0,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 条记录`,
            onChange: handlePageChange,
          }}
        />
      </Card>

      {/* 添加/编辑模态框 */}
      <Modal
        title={configModal.isEditing ? '编辑API配置' : '添加API配置'}
        open={configModal.visible}
        onOk={handleSubmit}
        onCancel={configModal.hideModal}
        confirmLoading={createMutation.isPending || updateMutation.isPending}
        width={700}
      >
        <Form form={configModal.form} layout="vertical">
          <Form.Item
            label="配置名称"
            name="name"
            rules={[{ required: true, message: '请输入配置名称' }]}
          >
            <Input placeholder="例如: OpenAI Official" />
          </Form.Item>

          <Form.Item
            label="类型"
            name="type"
            rules={[{ required: true, message: '请选择类型' }]}
          >
            <Select placeholder="选择API类型">
              <Option value="openai">OpenAI</Option>
              <Option value="anthropic">Anthropic</Option>
              <Option value="gemini">Gemini</Option>
              <Option value="custom">Custom</Option>
            </Select>
          </Form.Item>

          <Form.Item
            label="Base URL"
            name="base_url"
            rules={[
              { required: true, message: '请输入Base URL' },
              { type: 'url', message: '请输入有效的URL' },
            ]}
          >
            <Input placeholder="https://api.openai.com" />
          </Form.Item>

          <Form.Item label="API Key" name="api_key">
            <Input.Password placeholder="sk-..." />
          </Form.Item>

          <Form.Item
            label={
              <Space>
                <span>支持的模型</span>
                <Button
                  type="link"
                  size="small"
                  icon={<ApiOutlined />}
                  loading={fetchingModels}
                  onClick={handleFetchModels}
                >
                  获取模型
                </Button>
              </Space>
            }
            name="models"
            rules={[{ required: true, message: '请输入支持的模型' }]}
            extra='每行一个模型名称，或点击"获取模型"自动获取'
          >
            <TextArea
              rows={6}
              placeholder={'gpt-4\ngpt-3.5-turbo\ngpt-4-turbo'}
            />
          </Form.Item>

          <Form.Item label="优先级" name="priority">
            <InputNumber
              min={0}
              max={1000}
              style={{ width: '100%' }}
              placeholder="数值越大优先级越高"
            />
          </Form.Item>

          <Form.Item label="权重" name="weight">
            <InputNumber
              min={1}
              max={100}
              style={{ width: '100%' }}
              placeholder="负载均衡权重"
            />
          </Form.Item>

          <Form.Item label="最大RPS" name="max_rps">
            <InputNumber
              min={0}
              style={{ width: '100%' }}
              placeholder="0表示不限制"
            />
          </Form.Item>

          <Form.Item label="超时时间(秒)" name="timeout">
            <InputNumber
              min={1}
              max={300}
              style={{ width: '100%' }}
              placeholder="请求超时时间"
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default ApiConfigsPage;
