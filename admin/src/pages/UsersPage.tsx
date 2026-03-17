import React from 'react';
import {
  Table,
  Button,
  Input,
  Space,
  Tag,
  Drawer,
  Descriptions,
  message,
  Modal,
  Form,
  InputNumber,
  Select,
  Popconfirm,
} from 'antd';
import {
  UserOutlined,
  EditOutlined,
  StopOutlined,
  CheckCircleOutlined,
  DownloadOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { userService } from '../services/userService';
import type { User } from '../types';
import type { ColumnsType } from 'antd/es/table';
import TableToolbar from '../components/TableToolbar';
import StatusTag from '../components/StatusTag';
import PageContainer from '../components/PageContainer';
import { useTable } from '../hooks/useTable';
import { useModal } from '../hooks/useModal';
import { formatNumber, formatDateTime, formatPercent } from '../utils/format';

const { Search } = Input;
const { Option } = Select;

const UsersPage: React.FC = () => {
  const { page, pageSize, handlePageChange, resetPagination } = useTable();
  const [search, setSearch] = React.useState('');
  const [statusFilter, setStatusFilter] = React.useState<string | undefined>();
  const [selectedUser, setSelectedUser] = React.useState<User | null>(null);
  const [drawerVisible, setDrawerVisible] = React.useState(false);
  const [selectedRowKeys, setSelectedRowKeys] = React.useState<React.Key[]>([]);
  const quotaModal = useModal<User>();

  const queryClient = useQueryClient();

  // 获取用户列表
  const { data, isLoading, refetch } = useQuery({
    queryKey: ['users', page, pageSize, search, statusFilter],
    queryFn: () =>
      userService.getUsers({
        page,
        page_size: pageSize,
        search: search || undefined,
        status: statusFilter,
      }),
  });

  // 更新用户状态
  const updateStatusMutation = useMutation({
    mutationFn: ({ id, status }: { id: number; status: string }) =>
      userService.updateUserStatus(id, { status }),
    onSuccess: () => {
      message.success('用户状态更新成功');
      queryClient.invalidateQueries({ queryKey: ['users'] });
      setDrawerVisible(false);
    },
    onError: () => {
      message.error('用户状态更新失败');
    },
  });

  // 更新用户额度
  const updateQuotaMutation = useMutation({
    mutationFn: ({ id, quota }: { id: number; quota: number }) =>
      userService.updateUserQuota(id, { quota }),
    onSuccess: () => {
      message.success('用户额度更新成功');
      queryClient.invalidateQueries({ queryKey: ['users'] });
      quotaModal.hideModal();
      setDrawerVisible(false);
    },
    onError: () => {
      message.error('用户额度更新失败');
    },
  });

  // 表格列配置
  const columns: ColumnsType<User> = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
      render: (id: number) => (
        <span className="text-text-secondary font-mono text-sm">#{id}</span>
      ),
    },
    {
      title: '用户信息',
      key: 'user_info',
      width: 280,
      render: (_, record) => (
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-full bg-primary/10 flex items-center justify-center text-primary font-semibold">
            {record.username.charAt(0).toUpperCase()}
          </div>
          <div>
            <div className="text-text-primary font-medium flex items-center gap-2">
              {record.username}
              {record.is_admin && (
                <Tag color="blue" className="text-xs">管理员</Tag>
              )}
            </div>
            <div className="text-text-secondary text-sm">{record.email}</div>
          </div>
        </div>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => <StatusTag status={status} />,
    },
    {
      title: '配额使用',
      key: 'quota_usage',
      width: 200,
      render: (_, record) => {
        const percent = record.quota > 0 ? (record.used_quota / record.quota) * 100 : 0;
        const isLow = percent > 80;
        return (
          <div>
            <div className="flex items-center justify-between mb-1">
              <span className="text-text-secondary text-sm">
                {formatNumber(record.used_quota)} / {formatNumber(record.quota)}
              </span>
              <span className={`text-sm font-medium ${isLow ? 'text-red-400' : 'text-text-secondary'}`}>
                {percent.toFixed(1)}%
              </span>
            </div>
            <div className="w-full h-1.5 bg-white/5 rounded-full overflow-hidden">
              <div
                className={`h-full rounded-full transition-all ${
                  isLow ? 'bg-red-500' : 'bg-primary'
                }`}
                style={{ width: `${Math.min(percent, 100)}%` }}
              />
            </div>
          </div>
        );
      },
    },
    {
      title: '注册时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (date: string) => (
        <span className="text-text-secondary text-sm">{formatDateTime(date)}</span>
      ),
    },
    {
      title: '操作',
      key: 'action',
      fixed: 'right',
      width: 200,
      render: (_, record) => (
        <Space size="small">
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => handleViewUser(record)}
            className="text-primary hover:text-primary/80"
          >
            详情
          </Button>
          {record.status === 'active' ? (
            <Popconfirm
              title="确定要禁用该用户吗？"
              description="禁用后用户将无法使用 API 服务"
              onConfirm={() => handleDisableUser(record.id)}
              okText="确定"
              cancelText="取消"
            >
              <Button
                type="link"
                size="small"
                danger
                icon={<StopOutlined />}
              >
                禁用
              </Button>
            </Popconfirm>
          ) : (
            <Popconfirm
              title="确定要启用该用户吗？"
              onConfirm={() => handleEnableUser(record.id)}
              okText="确定"
              cancelText="取消"
            >
              <Button
                type="link"
                size="small"
                icon={<CheckCircleOutlined />}
                className="text-green-400 hover:text-green-300"
              >
                启用
              </Button>
            </Popconfirm>
          )}
        </Space>
      ),
    },
  ];

  // 查看用户详情
  const handleViewUser = (user: User) => {
    setSelectedUser(user);
    setDrawerVisible(true);
  };

  // 禁用用户
  const handleDisableUser = (id: number) => {
    updateStatusMutation.mutate({ id, status: 'disabled' });
  };

  // 启用用户
  const handleEnableUser = (id: number) => {
    updateStatusMutation.mutate({ id, status: 'active' });
  };

  // 调整额度
  const handleAdjustQuota = () => {
    if (!selectedUser) return;
    quotaModal.showModal(selectedUser);
    quotaModal.form.setFieldsValue({ quota: selectedUser.quota });
  };

  // 提交额度调整
  const handleQuotaSubmit = () => {
    if (!selectedUser) return;
    quotaModal.form.validateFields().then((values) => {
      updateQuotaMutation.mutate({
        id: selectedUser.id,
        quota: values.quota,
      });
    });
  };

  // 搜索
  const handleSearch = (value: string) => {
    setSearch(value);
    resetPagination();
  };

  // 状态筛选
  const handleStatusFilter = (value: string | undefined) => {
    setStatusFilter(value);
    resetPagination();
  };

  // 导出用户数据
  const handleExport = () => {
    if (!data?.users) return;
    
    const csvContent = [
      ['ID', '用户名', '邮箱', '状态', '总配额', '已使用', '剩余配额', '注册时间'].join(','),
      ...data.users.map(user => [
        user.id,
        user.username,
        user.email,
        user.status,
        user.quota,
        user.used_quota,
        user.quota - user.used_quota,
        user.created_at,
      ].join(',')),
    ].join('\n');

    const blob = new Blob(['\ufeff' + csvContent], { type: 'text/csv;charset=utf-8;' });
    const link = document.createElement('a');
    link.href = URL.createObjectURL(blob);
    link.download = `users_${new Date().toISOString().split('T')[0]}.csv`;
    link.click();
    message.success('导出成功');
  };

  // 批量操作
  const handleBatchDisable = () => {
    if (selectedRowKeys.length === 0) {
      message.warning('请先选择要禁用的用户');
      return;
    }
    Modal.confirm({
      title: '批量禁用用户',
      content: `确定要禁用选中的 ${selectedRowKeys.length} 个用户吗？`,
      onOk: async () => {
        try {
          await Promise.all(
            selectedRowKeys.map(id => userService.updateUserStatus(Number(id), { status: 'disabled' }))
          );
          message.success('批量禁用成功');
          queryClient.invalidateQueries({ queryKey: ['users'] });
          setSelectedRowKeys([]);
        } catch (error) {
          message.error('批量禁用失败');
        }
      },
    });
  };

  const handleBatchEnable = () => {
    if (selectedRowKeys.length === 0) {
      message.warning('请先选择要启用的用户');
      return;
    }
    Modal.confirm({
      title: '批量启用用户',
      content: `确定要启用选中的 ${selectedRowKeys.length} 个用户吗？`,
      onOk: async () => {
        try {
          await Promise.all(
            selectedRowKeys.map(id => userService.updateUserStatus(Number(id), { status: 'active' }))
          );
          message.success('批量启用成功');
          queryClient.invalidateQueries({ queryKey: ['users'] });
          setSelectedRowKeys([]);
        } catch (error) {
          message.error('批量启用失败');
        }
      },
    });
  };

  return (
    <PageContainer title="用户管理" description="管理平台用户、配额和权限">
      {/* 统计卡片 */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
        <div className="glass-card p-5">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-text-secondary text-sm mb-1">总用户数</div>
              <div className="text-2xl font-bold text-white">{formatNumber(data?.total || 0)}</div>
            </div>
            <div className="w-12 h-12 rounded-xl bg-primary/10 flex items-center justify-center">
              <UserOutlined className="text-2xl text-primary" />
            </div>
          </div>
        </div>
        <div className="glass-card p-5">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-text-secondary text-sm mb-1">活跃用户</div>
              <div className="text-2xl font-bold text-green-400">
                {formatNumber(data?.users.filter(u => u.status === 'active').length || 0)}
              </div>
            </div>
            <div className="w-12 h-12 rounded-xl bg-green-500/10 flex items-center justify-center">
              <CheckCircleOutlined className="text-2xl text-green-400" />
            </div>
          </div>
        </div>
        <div className="glass-card p-5">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-text-secondary text-sm mb-1">禁用用户</div>
              <div className="text-2xl font-bold text-red-400">
                {formatNumber(data?.users.filter(u => u.status === 'disabled').length || 0)}
              </div>
            </div>
            <div className="w-12 h-12 rounded-xl bg-red-500/10 flex items-center justify-center">
              <StopOutlined className="text-2xl text-red-400" />
            </div>
          </div>
        </div>
        <div className="glass-card p-5">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-text-secondary text-sm mb-1">管理员</div>
              <div className="text-2xl font-bold text-blue-400">
                {formatNumber(data?.users.filter(u => u.is_admin).length || 0)}
              </div>
            </div>
            <div className="w-12 h-12 rounded-xl bg-blue-500/10 flex items-center justify-center">
              <UserOutlined className="text-2xl text-blue-400" />
            </div>
          </div>
        </div>
      </div>

      <div className="glass-card p-6">
        {/* 搜索和筛选栏 */}
        <div className="mb-6 flex flex-wrap gap-4 items-center justify-between">
          <Space wrap size="middle">
            <Search
              placeholder="搜索用户名或邮箱"
              allowClear
              className="w-full sm:w-80"
              onSearch={handleSearch}
              size="large"
            />
            <Select
              placeholder="筛选状态"
              allowClear
              className="w-40"
              onChange={handleStatusFilter}
              size="large"
            >
              <Option value="active">正常</Option>
              <Option value="inactive">未激活</Option>
              <Option value="banned">已封禁</Option>
            </Select>
          </Space>
          <Space>
            {selectedRowKeys.length > 0 && (
              <>
                <Button
                  icon={<CheckCircleOutlined />}
                  onClick={handleBatchEnable}
                  className="bg-green-500/10 border-green-500/30 text-green-400 hover:bg-green-500/20"
                >
                  批量启用 ({selectedRowKeys.length})
                </Button>
                <Button
                  danger
                  icon={<StopOutlined />}
                  onClick={handleBatchDisable}
                >
                  批量禁用 ({selectedRowKeys.length})
                </Button>
              </>
            )}
            <Button
              icon={<DownloadOutlined />}
              onClick={handleExport}
              disabled={!data?.users || data.users.length === 0}
            >
              导出
            </Button>
            <Button
              icon={<ReloadOutlined />}
              onClick={() => refetch()}
            >
              刷新
            </Button>
          </Space>
        </div>

        {/* 用户列表表格 */}
        <Table
          columns={columns}
          dataSource={data?.users || []}
          rowKey="id"
          loading={isLoading}
          scroll={{ x: 1200 }}
          pagination={{
            current: page,
            pageSize: pageSize,
            total: data?.total || 0,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${formatNumber(total)} 条记录`,
            onChange: handlePageChange,
            pageSizeOptions: ['10', '20', '50', '100'],
          }}
          size="middle"
          rowClassName="hover:bg-white/5 transition-colors cursor-pointer"
        />
      </div>

      {/* 用户详情抽屉 */}
      <Drawer
        title={
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-full bg-primary/10 flex items-center justify-center text-primary font-semibold text-lg">
              {selectedUser?.username.charAt(0).toUpperCase()}
            </div>
            <div>
              <div className="text-lg font-semibold">用户详情</div>
              <div className="text-sm text-text-secondary font-normal">{selectedUser?.username}</div>
            </div>
          </div>
        }
        placement="right"
        width={640}
        onClose={() => setDrawerVisible(false)}
        open={drawerVisible}
        extra={
          <Space>
            <Button 
              icon={<EditOutlined />}
              onClick={handleAdjustQuota}
            >
              调整配额
            </Button>
            {selectedUser?.status === 'active' ? (
              <Button
                danger
                icon={<StopOutlined />}
                onClick={() => handleDisableUser(selectedUser.id)}
              >
                禁用用户
              </Button>
            ) : (
              <Button
                type="primary"
                icon={<CheckCircleOutlined />}
                onClick={() => handleEnableUser(selectedUser!.id)}
              >
                启用用户
              </Button>
            )}
          </Space>
        }
      >
        {selectedUser && (
          <div className="space-y-6">
            {/* 基本信息 */}
            <div className="glass-card p-5">
              <h4 className="text-base font-semibold text-white mb-4">基本信息</h4>
              <Descriptions column={1} colon={false} className="custom-descriptions">
                <Descriptions.Item label="用户 ID">
                  <span className="font-mono text-text-secondary">#{selectedUser.id}</span>
                </Descriptions.Item>
                <Descriptions.Item label="用户名">
                  <span className="text-white font-medium">{selectedUser.username}</span>
                </Descriptions.Item>
                <Descriptions.Item label="邮箱">
                  <span className="text-text-secondary">{selectedUser.email}</span>
                </Descriptions.Item>
                <Descriptions.Item label="状态">
                  <StatusTag status={selectedUser.status} />
                </Descriptions.Item>
                <Descriptions.Item label="角色">
                  {selectedUser.is_admin ? (
                    <Tag color="blue" className="font-medium">管理员</Tag>
                  ) : (
                    <Tag className="font-medium">普通用户</Tag>
                  )}
                </Descriptions.Item>
              </Descriptions>
            </div>

            {/* 配额信息 */}
            <div className="glass-card p-5">
              <h4 className="text-base font-semibold text-white mb-4">配额信息</h4>
              <div className="space-y-4">
                <div>
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-text-secondary text-sm">总配额</span>
                    <span className="text-white font-semibold text-lg">
                      {formatNumber(selectedUser.quota)} tokens
                    </span>
                  </div>
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-text-secondary text-sm">已使用</span>
                    <span className="text-text-secondary font-medium">
                      {formatNumber(selectedUser.used_quota)} tokens
                    </span>
                  </div>
                  <div className="flex items-center justify-between mb-3">
                    <span className="text-text-secondary text-sm">剩余配额</span>
                    <span className="text-primary font-semibold text-lg">
                      {formatNumber(selectedUser.quota - selectedUser.used_quota)} tokens
                    </span>
                  </div>
                  
                  {/* 使用进度条 */}
                  <div className="mt-4">
                    <div className="flex items-center justify-between mb-2">
                      <span className="text-text-secondary text-sm">使用率</span>
                      <span className="text-white font-medium">
                        {formatPercent(selectedUser.used_quota, selectedUser.quota)}
                      </span>
                    </div>
                    <div className="w-full h-2 bg-white/5 rounded-full overflow-hidden">
                      <div
                        className={`h-full rounded-full transition-all ${
                          (selectedUser.used_quota / selectedUser.quota) * 100 > 80
                            ? 'bg-red-500'
                            : 'bg-gradient-to-r from-primary to-blue-400'
                        }`}
                        style={{
                          width: `${Math.min(
                            (selectedUser.used_quota / selectedUser.quota) * 100,
                            100
                          )}%`,
                        }}
                      />
                    </div>
                  </div>
                </div>
              </div>
            </div>

            {/* 时间信息 */}
            <div className="glass-card p-5">
              <h4 className="text-base font-semibold text-white mb-4">时间信息</h4>
              <Descriptions column={1} colon={false} className="custom-descriptions">
                <Descriptions.Item label="注册时间">
                  <span className="text-text-secondary">{formatDateTime(selectedUser.created_at)}</span>
                </Descriptions.Item>
                {selectedUser.last_sign_in && (
                  <Descriptions.Item label="最后登录">
                    <span className="text-text-secondary">{formatDateTime(selectedUser.last_sign_in)}</span>
                  </Descriptions.Item>
                )}
              </Descriptions>
            </div>
          </div>
        )}
      </Drawer>

      {/* 配额调整模态框 */}
      <Modal
        title={
          <div className="flex items-center gap-2">
            <EditOutlined className="text-primary" />
            <span>调整用户配额</span>
          </div>
        }
        open={quotaModal.visible}
        onOk={handleQuotaSubmit}
        onCancel={quotaModal.hideModal}
        confirmLoading={updateQuotaMutation.isPending}
        okText="确认调整"
        cancelText="取消"
        width={500}
      >
        <Form form={quotaModal.form} layout="vertical" className="mt-6">
          <Form.Item
            label={<span className="text-text-primary font-medium">新的总配额</span>}
            name="quota"
            rules={[
              { required: true, message: '请输入配额' },
              { type: 'number', min: 0, message: '配额不能为负数' },
            ]}
          >
            <InputNumber
              style={{ width: '100%' }}
              size="large"
              min={0}
              step={10000}
              formatter={(value) =>
                `${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',')
              }
              parser={(value) => value!.replace(/\$\s?|(,*)/g, '') as any}
              addonAfter="tokens"
              placeholder="请输入配额数量"
            />
          </Form.Item>
          {selectedUser && (
            <div className="glass-card p-4 space-y-2">
              <div className="flex items-center justify-between text-sm">
                <span className="text-text-secondary">当前总配额</span>
                <span className="text-white font-medium">{formatNumber(selectedUser.quota)} tokens</span>
              </div>
              <div className="flex items-center justify-between text-sm">
                <span className="text-text-secondary">已使用配额</span>
                <span className="text-text-secondary">{formatNumber(selectedUser.used_quota)} tokens</span>
              </div>
              <div className="flex items-center justify-between text-sm pt-2 border-t border-border/40">
                <span className="text-text-secondary">剩余配额</span>
                <span className="text-primary font-semibold">{formatNumber(selectedUser.quota - selectedUser.used_quota)} tokens</span>
              </div>
            </div>
          )}
        </Form>
      </Modal>
    </PageContainer>
  );
};

export default UsersPage;
