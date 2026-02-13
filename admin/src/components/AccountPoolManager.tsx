import React, { useState } from 'react';
import {
  Modal,
  Table,
  Button,
  Space,
  Tag,
  Badge,
  Tooltip,
  Popconfirm,
  message,
  Form,
  Input,
  Descriptions,
  Statistic,
  Row,
  Col,
  Progress,
  Card,
  InputNumber,
} from 'antd';
import {
  PlusOutlined,
  ReloadOutlined,
  DeleteOutlined,
  EditOutlined,
  EyeOutlined,
  SyncOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
} from '@ant-design/icons';
import { useQuery, useMutation } from '@tanstack/react-query';
import { accountPoolService } from '../services/accountPoolService';
import type { AccountCredential, CreateCredentialRequest } from '../types';
import type { ColumnsType } from 'antd/es/table';
import { formatDateTime } from '../utils/format';

const { TextArea } = Input;

interface AccountPoolManagerProps {
  visible: boolean;
  poolId: number;
  poolName: string;
  onClose: () => void;
}

const AccountPoolManager: React.FC<AccountPoolManagerProps> = ({
  visible,
  poolId,
  poolName,
  onClose,
}) => {
  const [accountModalVisible, setAccountModalVisible] = useState(false);
  const [accountDetailVisible, setAccountDetailVisible] = useState(false);
  const [batchImportModalVisible, setBatchImportModalVisible] = useState(false);
  const [editingAccount, setEditingAccount] = useState<AccountCredential | null>(null);
  const [viewingAccount, setViewingAccount] = useState<AccountCredential | null>(null);
  const [accountForm] = Form.useForm();
  const [batchImportForm] = Form.useForm();

  // 获取账号列表
  const { data: accounts, isLoading, refetch } = useQuery({
    queryKey: ['pool-accounts', poolId],
    queryFn: () => accountPoolService.getCredentials({ pool_id: poolId }),
    enabled: visible && !!poolId,
  });

  // 添加账号
  const addAccountMutation = useMutation({
    mutationFn: accountPoolService.createCredential,
    onSuccess: () => {
      message.success('账号添加成功');
      refetch();
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
      refetch();
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
      refetch();
    },
    onError: () => {
      message.error('账号删除失败');
    },
  });

  // 刷新令牌
  const refreshAccountMutation = useMutation({
    mutationFn: accountPoolService.refreshCredential,
    onSuccess: () => {
      message.success('令牌刷新成功');
      refetch();
    },
    onError: (error: any) => {
      message.error(error.response?.data?.error?.message || '令牌刷新失败');
    },
  });

  // 批量导入
  const batchImportMutation = useMutation({
    mutationFn: accountPoolService.batchImport,
    onSuccess: (result) => {
      message.success(`导入完成！成功: ${result.success}, 失败: ${result.failed}`);
      if (result.errors && result.errors.length > 0) {
        console.error('导入错误:', result.errors);
      }
      refetch();
      setBatchImportModalVisible(false);
      batchImportForm.resetFields();
    },
    onError: (error: any) => {
      message.error(error.response?.data?.error?.message || '批量导入失败');
    },
  });

  // 打开批量导入
  const handleBatchImport = () => {
    batchImportForm.setFieldsValue({
      weight: 1,
      rate_limit: 0,
    });
    setBatchImportModalVisible(true);
  };

  // 提交批量导入
  const handleBatchImportSubmit = () => {
    batchImportForm.validateFields().then((values) => {
      batchImportMutation.mutate({
        pool_id: poolId,
        json_data: values.json_data,
        weight: values.weight || 1,
        rate_limit: values.rate_limit || 0,
      });
    });
  };

  // 表格列配置
  const columns: ColumnsType<AccountCredential> = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 60,
    },
    {
      title: '账号名称',
      dataIndex: 'account_name',
      key: 'account_name',
      render: (text: string, record) => (
        <Space direction="vertical" size={0}>
          <Space>
            {record.is_active ? (
              <CheckCircleOutlined style={{ color: '#52c41a' }} />
            ) : (
              <CloseCircleOutlined style={{ color: '#ff4d4f' }} />
            )}
            <span>{text || `账号 ${record.id}`}</span>
          </Space>
          {record.subscription_type && (
            <Tag color="blue" style={{ fontSize: '11px' }}>
              {record.subscription_type}
            </Tag>
          )}
        </Space>
      ),
    },
    {
      title: '状态',
      key: 'status',
      width: 100,
      render: (_, record) => {
        const statusConfig: Record<string, { color: string; text: string }> = {
          active: { color: 'success', text: '正常' },
          expired: { color: 'error', text: '已过期' },
          error: { color: 'error', text: '错误' },
          refreshing: { color: 'processing', text: '刷新中' },
          unknown: { color: 'default', text: '未知' },
        };
        const config = statusConfig[record.status] || statusConfig.unknown;
        return <Tag color={config.color}>{config.text}</Tag>;
      },
    },
    {
      title: '订阅/使用量',
      key: 'subscription_usage',
      width: 200,
      render: (_, record) => (
        <Space direction="vertical" size={4} style={{ width: '100%' }}>
          {record.subscription_expires_at && (
            <Tooltip title={`到期时间: ${formatDateTime(record.subscription_expires_at)}`}>
              <div style={{ fontSize: '12px' }}>
                <span>剩余: </span>
                <Tag color={
                  (record.subscription_days_remaining || 0) > 30 ? 'green' :
                  (record.subscription_days_remaining || 0) > 7 ? 'orange' : 'red'
                }>
                  {record.subscription_days_remaining || 0} 天
                </Tag>
              </div>
            </Tooltip>
          )}
          {record.usage_limit > 0 && (
            <Tooltip title={`已用: ${record.usage_current} / 限额: ${record.usage_limit}`}>
              <Progress
                percent={record.usage_percent}
                size="small"
                status={
                  record.usage_percent > 90 ? 'exception' :
                  record.usage_percent > 70 ? 'normal' : 'success'
                }
                format={() => `${record.usage_percent.toFixed(0)}%`}
              />
            </Tooltip>
          )}
        </Space>
      ),
    },
    {
      title: '使用情况',
      key: 'usage',
      width: 200,
      render: (_, record) => {
        const total = record.total_requests || 0;
        const errors = record.total_errors || 0;
        const success = total - errors;
        const successRate = total > 0 ? ((success / total) * 100) : 0;
        return (
          <Tooltip title={`成功: ${success} / 总计: ${total}`}>
            <Progress
              percent={successRate}
              size="small"
              status={successRate > 80 ? 'success' : successRate > 50 ? 'normal' : 'exception'}
              format={() => `${success}/${total}`}
            />
          </Tooltip>
        );
      },
    },
    {
      title: '速率限制',
      key: 'rate_limit',
      width: 150,
      render: (_, record) => {
        const limit = record.rate_limit || 0;
        const used = record.current_usage || 0;
        
        if (limit === 0) {
          return <Tag color="blue">无限制</Tag>;
        }
        
        const percentage = limit > 0 ? (used / limit) * 100 : 0;
        return (
          <Tooltip title={`已用: ${used} / 限制: ${limit}/分钟`}>
            <Progress
              percent={percentage}
              size="small"
              status={percentage > 90 ? 'exception' : percentage > 70 ? 'normal' : 'success'}
              format={() => `${used}/${limit}`}
            />
          </Tooltip>
        );
      },
    },
    {
      title: '最后使用',
      dataIndex: 'last_used_at',
      key: 'last_used_at',
      width: 150,
      render: (date: string) => (date ? formatDateTime(date) : '-'),
    },
    {
      title: '操作',
      key: 'action',
      fixed: 'right',
      width: 280,
      render: (_, record) => (
        <Space size="small">
          <Tooltip title="查看详情">
            <Button
              type="link"
              size="small"
              icon={<EyeOutlined />}
              onClick={() => handleViewAccount(record)}
            />
          </Tooltip>
          <Tooltip title="刷新令牌">
            <Button
              type="link"
              size="small"
              icon={<SyncOutlined />}
              loading={refreshAccountMutation.isPending}
              onClick={() => refreshAccountMutation.mutate(record.id)}
            />
          </Tooltip>
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => handleEditAccount(record)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确定删除此账号？"
            onConfirm={() => deleteAccountMutation.mutate(record.id)}
          >
            <Button type="link" size="small" danger icon={<DeleteOutlined />}>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  // 打开添加账号
  const handleAddAccount = () => {
    setEditingAccount(null);
    accountForm.resetFields();
    accountForm.setFieldsValue({
      pool_id: poolId,
      provider: 'kiro',
      auth_type: 'oauth',
      weight: 1,
      rate_limit: 0,
    });
    setAccountModalVisible(true);
  };

  // 打开编辑账号
  const handleEditAccount = (account: AccountCredential) => {
    setEditingAccount(account);
    accountForm.setFieldsValue({
      account_name: account.account_name,
      account_email: account.account_email,
      weight: account.weight,
      rate_limit: account.rate_limit,
    });
    setAccountModalVisible(true);
  };

  // 查看账号详情
  const handleViewAccount = (account: AccountCredential) => {
    setViewingAccount(account);
    setAccountDetailVisible(true);
  };

  // 提交账号表单
  const handleAccountSubmit = () => {
    accountForm.validateFields().then((values) => {
      const data: CreateCredentialRequest = {
        pool_id: poolId,
        provider: 'kiro',
        auth_type: values.auth_type || 'oauth',
        access_token: values.access_token,
        refresh_token: values.refresh_token,
        session_token: values.session_token,
        account_name: values.account_name,
        account_email: values.account_email,
        weight: values.weight || 1,
        rate_limit: values.rate_limit || 0,
      };

      if (editingAccount) {
        // 编辑时只更新部分字段
        const updateData: any = {
          account_name: values.account_name,
          account_email: values.account_email,
          weight: values.weight,
          rate_limit: values.rate_limit,
        };
        // 如果提供了新的 token，也更新
        if (values.access_token) updateData.access_token = values.access_token;
        if (values.refresh_token) updateData.refresh_token = values.refresh_token;
        if (values.session_token) updateData.session_token = values.session_token;
        
        updateAccountMutation.mutate({ id: editingAccount.id, data: updateData });
      } else {
        addAccountMutation.mutate(data);
      }
    });
  };

  // 计算统计数据
  const activeCount = accounts?.filter((a) => a.is_active).length || 0;
  const totalRequests = accounts?.reduce((sum, a) => sum + (a.total_requests || 0), 0) || 0;
  const totalErrors = accounts?.reduce((sum, a) => sum + (a.total_errors || 0), 0) || 0;
  const totalSuccess = totalRequests - totalErrors;
  const successRate = totalRequests > 0 ? ((totalSuccess / totalRequests) * 100).toFixed(1) : '0';

  return (
    <>
      <Modal
        title={
          <Space>
            <span>账号池管理</span>
            <Tag color="purple">{poolName}</Tag>
          </Space>
        }
        open={visible}
        onCancel={onClose}
        footer={null}
        width={1200}
        style={{ top: 20 }}
      >
        {/* 统计卡片 */}
        <Row gutter={16} style={{ marginBottom: 16 }}>
          <Col span={6}>
            <Card>
              <Statistic
                title="总账号数"
                value={accounts?.length || 0}
                prefix={<CheckCircleOutlined />}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="活跃账号"
                value={activeCount}
                valueStyle={{ color: '#3f8600' }}
                prefix={<CheckCircleOutlined />}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="总请求数"
                value={totalRequests}
                prefix={<SyncOutlined />}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="成功率"
                value={successRate}
                suffix="%"
                valueStyle={{ color: parseFloat(successRate) > 80 ? '#3f8600' : '#cf1322' }}
              />
            </Card>
          </Col>
        </Row>

        {/* 操作按钮 */}
        <Space style={{ marginBottom: 16 }}>
          <Button type="primary" icon={<PlusOutlined />} onClick={handleAddAccount}>
            添加账号
          </Button>
          <Button icon={<PlusOutlined />} onClick={handleBatchImport}>
            批量导入
          </Button>
          <Button icon={<ReloadOutlined />} onClick={() => refetch()}>
            刷新
          </Button>
        </Space>

        {/* 账号列表 */}
        <Table
          columns={columns}
          dataSource={accounts || []}
          rowKey="id"
          loading={isLoading}
          scroll={{ x: 1000 }}
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showTotal: (total) => `共 ${total} 个账号`,
          }}
        />
      </Modal>

      {/* 添加/编辑账号模态框 */}
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
          <Form.Item name="auth_type" hidden initialValue="oauth">
            <Input />
          </Form.Item>

          <Form.Item
            label="账号名称"
            name="account_name"
            rules={[{ required: true, message: '请输入账号名称' }]}
          >
            <Input placeholder="例如: Kiro Account 1" />
          </Form.Item>

          <Form.Item
            label="账号邮箱"
            name="account_email"
          >
            <Input placeholder="例如: user@example.com" />
          </Form.Item>

          {!editingAccount && (
            <>
              <Form.Item
                label="Access Token"
                name="access_token"
                rules={[{ required: true, message: '请输入 Access Token' }]}
              >
                <TextArea
                  rows={3}
                  placeholder="从 Kiro 账号管理器导出的 accessToken"
                />
              </Form.Item>

              <Form.Item
                label="Refresh Token"
                name="refresh_token"
                rules={[{ required: true, message: '请输入 Refresh Token' }]}
              >
                <TextArea
                  rows={3}
                  placeholder="从 Kiro 账号管理器导出的 refreshToken"
                />
              </Form.Item>

              <Form.Item
                label="Session Token (可选)"
                name="session_token"
              >
                <TextArea
                  rows={2}
                  placeholder="Session Token（如果有）"
                />
              </Form.Item>
            </>
          )}

          {editingAccount && (
            <div style={{ marginBottom: 16, padding: 12, background: '#f0f0f0', borderRadius: 4 }}>
              <p style={{ margin: 0, fontSize: 12, color: '#666' }}>
                提示：编辑时如需更新 Token，请填写以下字段，否则保持原有 Token
              </p>
            </div>
          )}

          {editingAccount && (
            <>
              <Form.Item
                label="Access Token (可选)"
                name="access_token"
              >
                <TextArea
                  rows={3}
                  placeholder="留空则不更新"
                />
              </Form.Item>

              <Form.Item
                label="Refresh Token (可选)"
                name="refresh_token"
              >
                <TextArea
                  rows={3}
                  placeholder="留空则不更新"
                />
              </Form.Item>

              <Form.Item
                label="Session Token (可选)"
                name="session_token"
              >
                <TextArea
                  rows={2}
                  placeholder="留空则不更新"
                />
              </Form.Item>
            </>
          )}

          <Form.Item
            label="权重"
            name="weight"
            rules={[{ required: true, message: '请输入权重' }]}
            extra="用于加权轮询策略，权重越高被选中概率越大"
          >
            <InputNumber min={1} max={100} style={{ width: '100%' }} />
          </Form.Item>

          <Form.Item
            label="速率限制（每分钟）"
            name="rate_limit"
            extra="0 表示无限制"
          >
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Modal>

      {/* 账号详情模态框 */}
      <Modal
        title="账号详情"
        open={accountDetailVisible}
        onCancel={() => {
          setAccountDetailVisible(false);
          setViewingAccount(null);
        }}
        footer={[
          <Button key="close" onClick={() => setAccountDetailVisible(false)}>
            关闭
          </Button>,
        ]}
        width={700}
      >
        {viewingAccount && (
          <Descriptions bordered column={2}>
            <Descriptions.Item label="账号ID">{viewingAccount.id}</Descriptions.Item>
            <Descriptions.Item label="账号名称">{viewingAccount.account_name || '-'}</Descriptions.Item>
            <Descriptions.Item label="账号邮箱" span={2}>{viewingAccount.account_email || '-'}</Descriptions.Item>
            <Descriptions.Item label="提供商">{viewingAccount.provider}</Descriptions.Item>
            <Descriptions.Item label="认证类型">{viewingAccount.auth_type}</Descriptions.Item>
            <Descriptions.Item label="状态">
              <Space>
                <Badge
                  status={viewingAccount.is_active ? 'success' : 'error'}
                  text={viewingAccount.is_active ? '活跃' : '禁用'}
                />
                <Tag color={
                  viewingAccount.status === 'active' ? 'success' :
                  viewingAccount.status === 'expired' ? 'error' :
                  viewingAccount.status === 'error' ? 'error' :
                  viewingAccount.status === 'refreshing' ? 'processing' : 'default'
                }>
                  {viewingAccount.status}
                </Tag>
              </Space>
            </Descriptions.Item>
            <Descriptions.Item label="健康状态">
              <Tag color={viewingAccount.health_status === 'healthy' ? 'success' : viewingAccount.health_status === 'unhealthy' ? 'error' : 'default'}>
                {viewingAccount.health_status}
              </Tag>
            </Descriptions.Item>
            
            {/* 订阅信息 */}
            {viewingAccount.subscription_type && (
              <>
                <Descriptions.Item label="订阅类型" span={2}>
                  <Space>
                    <Tag color="blue">{viewingAccount.subscription_type}</Tag>
                    {viewingAccount.subscription_title && (
                      <span>{viewingAccount.subscription_title}</span>
                    )}
                  </Space>
                </Descriptions.Item>
                {viewingAccount.subscription_expires_at && (
                  <>
                    <Descriptions.Item label="订阅到期">
                      {formatDateTime(viewingAccount.subscription_expires_at)}
                    </Descriptions.Item>
                    <Descriptions.Item label="剩余天数">
                      <Tag color={
                        (viewingAccount.subscription_days_remaining || 0) > 30 ? 'green' :
                        (viewingAccount.subscription_days_remaining || 0) > 7 ? 'orange' : 'red'
                      }>
                        {viewingAccount.subscription_days_remaining || 0} 天
                      </Tag>
                    </Descriptions.Item>
                  </>
                )}
              </>
            )}
            
            {/* 使用量信息 */}
            {viewingAccount.usage_limit > 0 && (
              <>
                <Descriptions.Item label="使用量" span={2}>
                  <Space direction="vertical" style={{ width: '100%' }}>
                    <div>
                      <span>{viewingAccount.usage_current} / {viewingAccount.usage_limit}</span>
                      <span style={{ marginLeft: 8, color: '#999' }}>
                        ({viewingAccount.usage_percent.toFixed(1)}%)
                      </span>
                    </div>
                    <Progress
                      percent={viewingAccount.usage_percent}
                      status={
                        viewingAccount.usage_percent > 90 ? 'exception' :
                        viewingAccount.usage_percent > 70 ? 'normal' : 'success'
                      }
                    />
                  </Space>
                </Descriptions.Item>
                {viewingAccount.base_limit > 0 && (
                  <Descriptions.Item label="基础额度">
                    {viewingAccount.base_current} / {viewingAccount.base_limit}
                  </Descriptions.Item>
                )}
                {viewingAccount.free_trial_limit > 0 && (
                  <Descriptions.Item label="试用额度">
                    {viewingAccount.free_trial_current} / {viewingAccount.free_trial_limit}
                  </Descriptions.Item>
                )}
                {viewingAccount.next_reset_date && (
                  <Descriptions.Item label="下次重置" span={2}>
                    {formatDateTime(viewingAccount.next_reset_date)}
                  </Descriptions.Item>
                )}
              </>
            )}
            
            {viewingAccount.machine_id && (
              <Descriptions.Item label="机器码" span={2}>
                <code style={{ fontSize: '12px' }}>{viewingAccount.machine_id}</code>
              </Descriptions.Item>
            )}
            
            <Descriptions.Item label="权重">{viewingAccount.weight}</Descriptions.Item>
            <Descriptions.Item label="速率限制">
              {viewingAccount.rate_limit > 0 ? `${viewingAccount.rate_limit}/分钟` : '无限制'}
            </Descriptions.Item>
            <Descriptions.Item label="当前使用量">{viewingAccount.current_usage}</Descriptions.Item>
            <Descriptions.Item label="总请求数">{viewingAccount.total_requests || 0}</Descriptions.Item>
            <Descriptions.Item label="成功请求">{viewingAccount.total_requests - viewingAccount.total_errors || 0}</Descriptions.Item>
            <Descriptions.Item label="失败请求">{viewingAccount.total_errors || 0}</Descriptions.Item>
            <Descriptions.Item label="成功率">
              {viewingAccount.total_requests > 0
                ? `${((1 - viewingAccount.error_rate) * 100).toFixed(1)}%`
                : '0%'}
            </Descriptions.Item>
            <Descriptions.Item label="是否过期">
              <Tag color={viewingAccount.is_expired ? 'error' : 'success'}>
                {viewingAccount.is_expired ? '已过期' : '正常'}
              </Tag>
            </Descriptions.Item>
            {viewingAccount.expires_at && (
              <Descriptions.Item label="Token过期时间" span={2}>
                {formatDateTime(viewingAccount.expires_at)}
              </Descriptions.Item>
            )}
            {viewingAccount.last_error && (
              <Descriptions.Item label="最后错误" span={2}>
                <span style={{ color: '#ff4d4f', fontSize: '12px' }}>{viewingAccount.last_error}</span>
              </Descriptions.Item>
            )}
            <Descriptions.Item label="创建时间" span={2}>
              {viewingAccount.created_at ? formatDateTime(viewingAccount.created_at) : '-'}
            </Descriptions.Item>
            <Descriptions.Item label="最后使用" span={2}>
              {viewingAccount.last_used_at ? formatDateTime(viewingAccount.last_used_at) : '从未使用'}
            </Descriptions.Item>
            {viewingAccount.last_checked_at && (
              <Descriptions.Item label="最后检查" span={2}>
                {formatDateTime(viewingAccount.last_checked_at)}
              </Descriptions.Item>
            )}
          </Descriptions>
        )}
      </Modal>

      {/* 批量导入模态框 */}
      <Modal
        title="批量导入 Kiro 账号"
        open={batchImportModalVisible}
        onOk={handleBatchImportSubmit}
        onCancel={() => {
          setBatchImportModalVisible(false);
          batchImportForm.resetFields();
        }}
        confirmLoading={batchImportMutation.isPending}
        width={800}
      >
        <Form form={batchImportForm} layout="vertical">
          <Form.Item
            label="JSON 数据"
            name="json_data"
            rules={[{ required: true, message: '请粘贴 JSON 数据' }]}
            extra="从 Kiro Account Manager 导出的 JSON 数据，支持数组格式"
          >
            <TextArea
              rows={15}
              placeholder={`粘贴从 Kiro Account Manager 导出的 JSON 数据，例如:
[
  {
    "email": "user@example.com",
    "credentials": {
      "accessToken": "...",
      "refreshToken": "...",
      ...
    },
    ...
  }
]`}
              style={{ fontFamily: 'monospace', fontSize: '12px' }}
            />
          </Form.Item>

          <Form.Item
            label="默认权重"
            name="weight"
            extra="所有导入账号的默认权重"
          >
            <InputNumber min={1} max={100} style={{ width: '100%' }} />
          </Form.Item>

          <Form.Item
            label="默认速率限制（每分钟）"
            name="rate_limit"
            extra="0 表示无限制"
          >
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Modal>

    </>
  );
};

export default AccountPoolManager;
