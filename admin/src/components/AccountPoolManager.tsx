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
  SettingOutlined,
} from '@ant-design/icons';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { accountPoolService } from '../services/accountPoolService';
import type { AccountCredential } from '../types';
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
  const [quotaModalVisible, setQuotaModalVisible] = useState(false);
  const [editingAccount, setEditingAccount] = useState<AccountCredential | null>(null);
  const [viewingAccount, setViewingAccount] = useState<AccountCredential | null>(null);
  const [quotaAccount, setQuotaAccount] = useState<AccountCredential | null>(null);
  const [accountForm] = Form.useForm();
  const [quotaForm] = Form.useForm();

  const queryClient = useQueryClient();

  // 获取账号列表
  const { data: accounts, isLoading, refetch } = useQuery({
    queryKey: ['pool-accounts', poolId],
    queryFn: () => accountPoolService.getCredentials({ pool_id: poolId }),
    enabled: visible && !!poolId,
  });

  // 获取池统计
  const { data: poolStats } = useQuery({
    queryKey: ['pool-stats', poolId],
    queryFn: () => accountPoolService.getPoolStats(poolId),
    enabled: visible && !!poolId,
  });

  // 设置配额
  const setQuotaMutation = useMutation({
    mutationFn: ({ id, quota }: { id: number; quota: number }) =>
      accountPoolService.updateCredential(id, { daily_quota: quota }),
    onSuccess: () => {
      message.success('配额设置成功');
      refetch();
      setQuotaModalVisible(false);
      quotaForm.resetFields();
      setQuotaAccount(null);
    },
    onError: () => {
      message.error('配额设置失败');
    },
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

  // 更新账号状态
  const toggleStatusMutation = useMutation({
    mutationFn: ({ id, isActive }: { id: number; isActive: boolean }) =>
      accountPoolService.updateCredentialStatus(id, isActive),
    onSuccess: () => {
      message.success('状态更新成功');
      refetch();
    },
    onError: () => {
      message.error('状态更新失败');
    },
  });

  // 打开配额设置
  const handleSetQuota = (account: AccountCredential) => {
    setQuotaAccount(account);
    quotaForm.setFieldsValue({
      daily_quota: account.daily_quota || 0,
    });
    setQuotaModalVisible(true);
  };

  // 提交配额设置
  const handleQuotaSubmit = () => {
    quotaForm.validateFields().then((values) => {
      if (quotaAccount) {
        setQuotaMutation.mutate({
          id: quotaAccount.id,
          quota: values.daily_quota,
        });
      }
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
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record) => (
        <Space>
          {record.is_active ? (
            <CheckCircleOutlined style={{ color: '#52c41a' }} />
          ) : (
            <CloseCircleOutlined style={{ color: '#ff4d4f' }} />
          )}
          <span>{text || `账号 ${record.id}`}</span>
        </Space>
      ),
    },
    {
      title: '状态',
      dataIndex: 'is_active',
      key: 'is_active',
      width: 100,
      render: (is_active: boolean, record) => (
        <Badge
          status={is_active ? 'success' : 'error'}
          text={is_active ? '活跃' : '禁用'}
        />
      ),
    },
    {
      title: '使用情况',
      key: 'usage',
      width: 200,
      render: (_, record) => {
        const total = record.request_count || 0;
        const success = record.success_count || 0;
        const successRate = total > 0 ? (success / total) * 100 : 0;
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
      title: '配额使用',
      key: 'quota',
      width: 150,
      render: (_, record) => {
        const quota = record.daily_quota || 0;
        const used = record.daily_used || 0;
        const percentage = quota > 0 ? (used / quota) * 100 : 0;
        
        if (quota === 0) {
          return <Tag color="blue">无限制</Tag>;
        }
        
        return (
          <Tooltip title={`已用: ${used} / 配额: ${quota}`}>
            <Progress
              percent={percentage}
              size="small"
              status={percentage > 90 ? 'exception' : percentage > 70 ? 'normal' : 'success'}
              format={() => `${used}/${quota}`}
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
          <Tooltip title="设置配额">
            <Button
              type="link"
              size="small"
              icon={<SettingOutlined />}
              onClick={() => handleSetQuota(record)}
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
      region: 'us-east-1',
    });
    setAccountModalVisible(true);
  };

  // 打开编辑账号
  const handleEditAccount = (account: AccountCredential) => {
    setEditingAccount(account);
    accountForm.setFieldsValue({
      name: account.name,
      credentials_data: JSON.stringify(account.credentials_data, null, 2),
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
      let credData = values.credentials_data;
      if (typeof credData === 'string') {
        try {
          credData = JSON.parse(credData);
        } catch (e) {
          message.error('凭据数据格式错误，请输入有效的 JSON');
          return;
        }
      }

      const data = {
        ...values,
        credentials_data: credData,
        pool_id: poolId,
        provider: 'kiro',
      };

      if (editingAccount) {
        updateAccountMutation.mutate({ id: editingAccount.id, data });
      } else {
        addAccountMutation.mutate(data);
      }
    });
  };

  // 计算统计数据
  const activeCount = accounts?.filter((a) => a.is_active).length || 0;
  const totalRequests = accounts?.reduce((sum, a) => sum + (a.request_count || 0), 0) || 0;
  const totalSuccess = accounts?.reduce((sum, a) => sum + (a.success_count || 0), 0) || 0;
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
            extra="JSON 格式，包含 accessToken, refreshToken, clientId, clientSecret 等字段"
          >
            <TextArea
              rows={12}
              placeholder={`{
  "accessToken": "...",
  "refreshToken": "...",
  "clientId": "...",
  "clientSecret": "...",
  "region": "us-east-1"
}`}
            />
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
            <Descriptions.Item label="账号名称">{viewingAccount.name}</Descriptions.Item>
            <Descriptions.Item label="提供商">{viewingAccount.provider}</Descriptions.Item>
            <Descriptions.Item label="状态">
              <Badge
                status={viewingAccount.is_active ? 'success' : 'error'}
                text={viewingAccount.is_active ? '活跃' : '禁用'}
              />
            </Descriptions.Item>
            <Descriptions.Item label="区域">
              {viewingAccount.credentials_data?.region || 'us-east-1'}
            </Descriptions.Item>
            <Descriptions.Item label="认证方式">
              {viewingAccount.credentials_data?.auth_method || '自动检测'}
            </Descriptions.Item>
            <Descriptions.Item label="总请求数">{viewingAccount.request_count || 0}</Descriptions.Item>
            <Descriptions.Item label="成功请求">{viewingAccount.success_count || 0}</Descriptions.Item>
            <Descriptions.Item label="失败请求">{viewingAccount.error_count || 0}</Descriptions.Item>
            <Descriptions.Item label="成功率">
              {viewingAccount.request_count > 0
                ? `${((viewingAccount.success_count / viewingAccount.request_count) * 100).toFixed(1)}%`
                : '0%'}
            </Descriptions.Item>
            <Descriptions.Item label="每日配额">
              {viewingAccount.daily_quota > 0 ? viewingAccount.daily_quota : '无限制'}
            </Descriptions.Item>
            <Descriptions.Item label="今日已用">{viewingAccount.daily_used || 0}</Descriptions.Item>
            <Descriptions.Item label="配额重置时间" span={2}>
              {viewingAccount.quota_reset_at ? formatDateTime(viewingAccount.quota_reset_at) : '每日 00:00 UTC'}
            </Descriptions.Item>
            <Descriptions.Item label="创建时间" span={2}>
              {viewingAccount.created_at ? formatDateTime(viewingAccount.created_at) : '-'}
            </Descriptions.Item>
            <Descriptions.Item label="最后使用" span={2}>
              {viewingAccount.last_used_at ? formatDateTime(viewingAccount.last_used_at) : '从未使用'}
            </Descriptions.Item>
            {viewingAccount.error_message && (
              <Descriptions.Item label="错误信息" span={2}>
                <Tag color="error">{viewingAccount.error_message}</Tag>
              </Descriptions.Item>
            )}
          </Descriptions>
        )}
      </Modal>

      {/* 配额设置模态框 */}
      <Modal
        title="设置每日配额"
        open={quotaModalVisible}
        onOk={handleQuotaSubmit}
        onCancel={() => {
          setQuotaModalVisible(false);
          quotaForm.resetFields();
          setQuotaAccount(null);
        }}
        confirmLoading={setQuotaMutation.isPending}
      >
        <Form form={quotaForm} layout="vertical">
          <Form.Item
            label="每日配额"
            name="daily_quota"
            rules={[{ required: true, message: '请输入每日配额' }]}
            extra="设置为 0 表示无限制"
          >
            <InputNumber
              min={0}
              style={{ width: '100%' }}
              placeholder="0 表示无限制"
            />
          </Form.Item>
          {quotaAccount && (
            <div style={{ marginTop: 16 }}>
              <p>当前账号: {quotaAccount.name}</p>
              <p>今日已用: {quotaAccount.daily_used || 0}</p>
              <p>
                配额重置时间:{' '}
                {quotaAccount.quota_reset_at
                  ? formatDateTime(quotaAccount.quota_reset_at)
                  : '每日 00:00 UTC'}
              </p>
            </div>
          )}
        </Form>
      </Modal>
    </>
  );
};

export default AccountPoolManager;
