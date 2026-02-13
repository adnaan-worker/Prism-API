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
  Tag,
  Tabs,
  List,
  Badge,
  Tooltip,
  Descriptions,
  Statistic,
  Row,
  Col,
  Progress,
} from 'antd';
import {
  EditOutlined,
  DeleteOutlined,
  ApiOutlined,
  PlusOutlined,
  ReloadOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  DatabaseOutlined,
  UserOutlined,
  SyncOutlined,
  EyeOutlined,
} from '@ant-design/icons';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiConfigService } from '../services/apiConfigService';
import { accountPoolService } from '../services/accountPoolService';
import type { APIConfig, AccountCredential } from '../types';
import type { ColumnsType } from 'antd/es/table';
import TableToolbar from '../components/TableToolbar';
import ProviderTag from '../components/ProviderTag';
import AccountPoolManager from '../components/AccountPoolManager';
import PageContainer from '../components/PageContainer';
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
  const [selectedType, setSelectedType] = React.useState<string>('');
  const [currentPoolId, setCurrentPoolId] = React.useState<number | null>(null);
  
  // 账号池管理弹窗
  const [poolManageVisible, setPoolManageVisible] = React.useState(false);
  const [managingPoolId, setManagingPoolId] = React.useState<number | null>(null);
  const [managingConfig, setManagingConfig] = React.useState<APIConfig | null>(null);
  
  // 账号添加/编辑
  const [accountModalVisible, setAccountModalVisible] = React.useState(false);
  const [accountForm] = Form.useForm();
  const [editingAccount, setEditingAccount] = React.useState<AccountCredential | null>(null);
  const [viewingAccount, setViewingAccount] = React.useState<AccountCredential | null>(null);
  const [accountDetailVisible, setAccountDetailVisible] = React.useState(false);

  const queryClient = useQueryClient();

  // 获取当前池的账号列表
  const { data: poolAccounts, refetch: refetchAccounts } = useQuery({
    queryKey: ['pool-accounts', currentPoolId],
    queryFn: () => accountPoolService.getCredentials({ pool_id: currentPoolId! }),
    enabled: !!currentPoolId && selectedType === 'kiro',
  });
  
  // 获取管理池的账号列表
  const { data: managingPoolAccounts, refetch: refetchManagingAccounts } = useQuery({
    queryKey: ['managing-pool-accounts', managingPoolId],
    queryFn: () => accountPoolService.getCredentials({ pool_id: managingPoolId! }),
    enabled: !!managingPoolId && poolManageVisible,
  });
  
  // 获取管理池的详情
  const { data: managingPoolData } = useQuery({
    queryKey: ['managing-pool', managingPoolId],
    queryFn: () => accountPoolService.getPool(managingPoolId!),
    enabled: !!managingPoolId && poolManageVisible,
  });

  // 创建账号池（用于 Kiro）
  const createPoolMutation = useMutation({
    mutationFn: accountPoolService.createPool,
    onSuccess: (pool) => {
      setCurrentPoolId(pool.id);
      message.success('账号池创建成功');
    },
    onError: () => {
      message.error('账号池创建失败');
    },
  });

  // 添加账号
  const addAccountMutation = useMutation({
    mutationFn: accountPoolService.createCredential,
    onSuccess: () => {
      message.success('账号添加成功');
      refetchAccounts();
      setAccountModalVisible(false);
      accountForm.resetFields();
    },
    onError: () => {
      message.error('账号添加失败');
    },
  });

  // 更新账号
  const updateAccountMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: any }) =>
      accountPoolService.updateCredential(id, data),
    onSuccess: () => {
      message.success('账号更新成功');
      refetchAccounts();
      setAccountModalVisible(false);
      accountForm.resetFields();
      setEditingAccount(null);
    },
    onError: () => {
      message.error('账号更新失败');
    },
  });

  // 删除账号
  const deleteAccountMutation = useMutation({
    mutationFn: accountPoolService.deleteCredential,
    onSuccess: () => {
      message.success('账号删除成功');
      refetchAccounts();
    },
    onError: () => {
      message.error('账号删除失败');
    },
  });

  // 刷新账号令牌
  const refreshAccountMutation = useMutation({
    mutationFn: accountPoolService.refreshCredential,
    onSuccess: () => {
      message.success('令牌刷新成功');
      refetchAccounts();
    },
    onError: () => {
      message.error('令牌刷新失败');
    },
  });

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
        { text: 'Kiro', value: 'kiro' },
        { text: 'Custom', value: 'custom' },
      ],
      onFilter: (value, record) => record.type === value,
    },
    {
      title: '配置类型',
      dataIndex: 'config_type',
      key: 'config_type',
      width: 120,
      render: (configType: string, record) => {
        if (configType === 'account_pool') {
          return (
            <Tooltip title={`账号池 ID: ${record.account_pool_id}`}>
              <Tag color="purple" icon={<DatabaseOutlined />}>
                账号池
              </Tag>
            </Tooltip>
          );
        }
        return <Tag color="blue">直接调用</Tag>;
      },
      filters: [
        { text: '直接调用', value: 'direct' },
        { text: '账号池', value: 'account_pool' },
      ],
      onFilter: (value, record) => record.config_type === value,
    },
    {
      title: 'Base URL',
      dataIndex: 'base_url',
      key: 'base_url',
      ellipsis: true,
      render: (url: string, record) => {
        if (record.config_type === 'account_pool') {
          return <Tag color="purple">使用账号池</Tag>;
        }
        return url;
      },
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
      width: 200,
      render: (_, record) => {
        // 检查是否是账号池类型
        const isAccountPool = record.config_type === 'account_pool' && record.account_pool_id;
        
        return (
          <Space>
            {isAccountPool && (
              <Tooltip title="管理账号池">
                <Button
                  type="link"
                  size="small"
                  icon={<DatabaseOutlined />}
                  onClick={() => {
                    setManagingPoolId(record.account_pool_id!);
                    setManagingConfig(record);
                    setPoolManageVisible(true);
                  }}
                >
                  账号
                </Button>
              </Tooltip>
            )}
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
        );
      },
    },
  ];

  // 打开添加模态框
  const handleAdd = async () => {
    configModal.showModal();
    setSelectedType('');
    setCurrentPoolId(null);
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
    setSelectedType(config.type);
    
    // 如果使用账号池，设置 pool_id
    if (config.config_type === 'account_pool' && config.account_pool_id) {
      setCurrentPoolId(config.account_pool_id);
    } else {
      setCurrentPoolId(null);
    }
    
    configModal.form.setFieldsValue({
      ...config,
      models: config.models.join('\n'),
    });
  };

  // 添加账号
  const handleAddAccount = () => {
    setEditingAccount(null);
    accountForm.resetFields();
    accountForm.setFieldsValue({
      pool_id: currentPoolId,
      provider: 'kiro',
      region: 'us-east-1',
    });
    setAccountModalVisible(true);
  };

  // 编辑账号
  const handleEditAccount = (account: AccountCredential) => {
    setEditingAccount(account);
    accountForm.setFieldsValue({
      name: account.name,
      credentials_data: JSON.stringify(account.credentials_data, null, 2),
    });
    setAccountModalVisible(true);
  };

  // 提交账号表单
  const handleAccountSubmit = () => {
    accountForm.validateFields().then((values) => {
      let credData = values.credentials_data;
      if (typeof credData === 'string') {
        try {
          credData = JSON.parse(credData);
        } catch (e) {
          message.error('凭据数据格式错误');
          return;
        }
      }

      const data = {
        ...values,
        credentials_data: credData,
        pool_id: currentPoolId,
        provider: 'kiro',
      };

      if (editingAccount) {
        updateAccountMutation.mutate({ id: editingAccount.id, data });
      } else {
        addAccountMutation.mutate(data);
      }
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
  const handleSubmit = async () => {
    try {
      const values = await configModal.form.validateFields();
      
      const modelsArray =
        typeof values.models === 'string'
          ? values.models.split('\n').filter((m: string) => m.trim())
          : values.models;

      let configType = 'direct';
      let accountPoolId = null;
      let baseUrl = values.base_url;
      
      // 如果是 Kiro 类型，使用账号池
      if (values.type === 'kiro') {
        configType = 'account_pool';
        let poolId = currentPoolId;
        
        if (!poolId) {
          // 创建新的账号池
          const pool = await createPoolMutation.mutateAsync({
            name: `${values.name} - Account Pool`,
            provider: 'kiro',
            description: `Auto-created pool for ${values.name}`,
            is_active: true,
            strategy: 'weighted_round_robin',
            health_check_interval: 300,
            health_check_timeout: 30,
            max_retries: 3,
          });
          poolId = pool.id;
        }
        
        accountPoolId = poolId;
        baseUrl = ''; // 账号池类型不需要 base_url
      }

      const data = {
        ...values,
        config_type: configType,
        account_pool_id: accountPoolId,
        base_url: baseUrl,
        models: modelsArray,
      };

      if (configModal.editingItem) {
        updateMutation.mutate({ id: configModal.editingItem.id, data });
      } else {
        createMutation.mutate(data);
      }
    } catch (error) {
      console.error('Form validation failed:', error);
    }
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
    <PageContainer title="API 配置" description="管理 AI 服务提供商的接入配置">
      <Card>
        {/* 操作栏 */}
        <Space style={{ marginBottom: 16 }} wrap><TableToolbar
          onAdd={handleAdd}
          addText="添加配置"
          onRefresh={() => refetch()}
          extra={
            <Space wrap>
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
                <Option value="kiro">Kiro</Option>
                <Option value="custom">Custom</Option>
              </Select>
            </Space>
          }
        /></Space>
        

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
            <Select 
              placeholder="选择API类型"
              onChange={(value) => {
                setSelectedType(value);
                // 清空相关字段
                if (value === 'kiro') {
                  configModal.form.setFieldsValue({
                    base_url: '',
                    api_key: '',
                  });
                } else {
                  configModal.form.setFieldsValue({
                    pool_id: undefined,
                  });
                }
              }}
            >
              <Option value="openai">OpenAI</Option>
              <Option value="anthropic">Anthropic</Option>
              <Option value="gemini">Gemini</Option>
              <Option value="kiro">Kiro (账号池)</Option>
              <Option value="custom">Custom</Option>
            </Select>
          </Form.Item>

          {selectedType === 'kiro' ? (
            <>
              {/* Kiro 账号管理 */}
              <Form.Item label="Kiro 账号管理">
                <Card size="small">
                  <Space direction="vertical" style={{ width: '100%' }}>
                    <Space>
                      <Button
                        type="primary"
                        icon={<PlusOutlined />}
                        onClick={handleAddAccount}
                        disabled={!currentPoolId && !configModal.editingItem}
                      >
                        添加账号
                      </Button>
                      <Button
                        icon={<ReloadOutlined />}
                        onClick={() => refetchAccounts()}
                        disabled={!currentPoolId}
                      >
                        刷新
                      </Button>
                      {!currentPoolId && !configModal.editingItem && (
                        <Tag color="orange">保存配置后可添加账号</Tag>
                      )}
                    </Space>

                    <List
                      dataSource={poolAccounts || []}
                      locale={{ emptyText: '暂无账号，请添加' }}
                      renderItem={(account: AccountCredential) => (
                        <List.Item
                          actions={[
                            <Tooltip title="刷新令牌">
                              <Button
                                type="link"
                                size="small"
                                icon={<ReloadOutlined />}
                                onClick={() => refreshAccountMutation.mutate(account.id)}
                              />
                            </Tooltip>,
                            <Button
                              type="link"
                              size="small"
                              icon={<EditOutlined />}
                              onClick={() => handleEditAccount(account)}
                            >
                              编辑
                            </Button>,
                            <Popconfirm
                              title="确定删除此账号？"
                              onConfirm={() => deleteAccountMutation.mutate(account.id)}
                            >
                              <Button type="link" size="small" danger icon={<DeleteOutlined />}>
                                删除
                              </Button>
                            </Popconfirm>,
                          ]}
                        >
                          <List.Item.Meta
                            avatar={
                              account.is_active ? (
                                <CheckCircleOutlined style={{ color: '#52c41a', fontSize: 20 }} />
                              ) : (
                                <CloseCircleOutlined style={{ color: '#ff4d4f', fontSize: 20 }} />
                              )
                            }
                            title={
                              <Space>
                                {account.name || `账号 ${account.id}`}
                                <Badge
                                  status={account.is_active ? 'success' : 'error'}
                                  text={account.is_active ? '活跃' : '禁用'}
                                />
                              </Space>
                            }
                            description={
                              <Space direction="vertical" size="small">
                                <span>区域: {account.credentials_data?.region || 'us-east-1'}</span>
                                <span>请求次数: {account.request_count || 0}</span>
                                {account.last_used_at && (
                                  <span>最后使用: {formatDateTime(account.last_used_at)}</span>
                                )}
                              </Space>
                            }
                          />
                        </List.Item>
                      )}
                    />
                  </Space>
                </Card>
              </Form.Item>
              
              <Form.Item
                label="支持的模型"
                name="models"
                rules={[{ required: true, message: '请输入支持的模型' }]}
                extra="Kiro 支持的 Claude 模型，每行一个"
              >
                <TextArea
                  rows={4}
                  placeholder={'claude-sonnet-4-5\nclaude-haiku-4-5\nclaude-opus-4-5\nclaude-opus-4-6'}
                />
              </Form.Item>

              {/* 隐藏字段 */}
              <Form.Item name="base_url" hidden>
                <Input />
              </Form.Item>
            </>
          ) : (
            <>
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
            </>
          )}

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

      {/* 账号添加/编辑模态框 */}
      <Modal
        title={editingAccount ? '编辑账号' : '添加账号'}
        open={accountModalVisible}
        onOk={handleAccountSubmit}
        onCancel={() => {
          setAccountModalVisible(false);
          accountForm.resetFields();
          setEditingAccount(null);
        }}
        confirmLoading={addAccountMutation.isPending || updateAccountMutation.isPending}
        width={600}
      >
        <Form form={accountForm} layout="vertical">
          <Form.Item
            label="账号名称"
            name="name"
            rules={[{ required: true, message: '请输入账号名称' }]}
          >
            <Input placeholder="例如: Kiro Account 1" />
          </Form.Item>

          <Form.Item
            label="凭据数据"
            name="credentials_data"
            rules={[{ required: true, message: '请输入凭据数据' }]}
            extra="JSON 格式，包含 access_token, refresh_token, profile_arn, region 等字段"
          >
            <TextArea
              rows={10}
              placeholder={`{
  "access_token": "...",
  "refresh_token": "...",
  "profile_arn": "...",
  "region": "us-east-1",
  "auth_method": "social"
}`}
            />
          </Form.Item>
        </Form>
      </Modal>
      
      {/* 账号池管理弹窗 */}
      {managingPoolId && managingConfig && (
        <AccountPoolManager
          visible={poolManageVisible}
          poolId={managingPoolId}
          poolName={managingConfig.name}
          onClose={() => {
            setPoolManageVisible(false);
            setManagingPoolId(null);
            setManagingConfig(null);
          }}
        />
      )}
    </PageContainer>
  );
};

export default ApiConfigsPage;
