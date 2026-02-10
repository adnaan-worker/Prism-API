import React from 'react';
import {
  Card,
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
} from '@ant-design/icons';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { userService } from '../services/userService';
import type { User } from '../types';
import type { ColumnsType } from 'antd/es/table';
import TableToolbar from '../components/TableToolbar';
import StatusTag from '../components/StatusTag';
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
    },
    {
      title: '用户名',
      dataIndex: 'username',
      key: 'username',
      render: (text: string) => (
        <Space>
          <UserOutlined />
          {text}
        </Space>
      ),
    },
    {
      title: '邮箱',
      dataIndex: 'email',
      key: 'email',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => <StatusTag status={status} />,
      filters: [
        { text: '正常', value: 'active' },
        { text: '禁用', value: 'disabled' },
      ],
      onFilter: (value, record) => record.status === value,
    },
    {
      title: '总额度',
      dataIndex: 'quota',
      key: 'quota',
      render: (quota: number) => formatNumber(quota),
    },
    {
      title: '已使用',
      dataIndex: 'used_quota',
      key: 'used_quota',
      render: (used: number) => formatNumber(used),
    },
    {
      title: '剩余额度',
      key: 'remaining',
      render: (_, record) => formatNumber(record.quota - record.used_quota),
    },
    {
      title: '管理员',
      dataIndex: 'is_admin',
      key: 'is_admin',
      render: (isAdmin: boolean) =>
        isAdmin ? <Tag color="blue">是</Tag> : <Tag>否</Tag>,
    },
    {
      title: '注册时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date: string) => formatDateTime(date),
    },
    {
      title: '操作',
      key: 'action',
      fixed: 'right',
      width: 200,
      render: (_, record) => (
        <Space>
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => handleViewUser(record)}
          >
            详情
          </Button>
          {record.status === 'active' ? (
            <Popconfirm
              title="确定要禁用该用户吗？"
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

  return (
    <div>
      <Card>
        {/* 搜索和筛选栏 */}
        <Space style={{ marginBottom: 16 }} wrap>
          <Search
            placeholder="搜索用户名或邮箱"
            allowClear
            style={{ width: 300 }}
            onSearch={handleSearch}
          />
          <Select
            placeholder="筛选状态"
            allowClear
            style={{ width: 150 }}
            onChange={handleStatusFilter}
          >
            <Option value="active">正常</Option>
            <Option value="disabled">禁用</Option>
          </Select>
          <TableToolbar showAdd={false} onRefresh={() => refetch()} />
        </Space>

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
            showTotal: (total) => `共 ${total} 条记录`,
            onChange: handlePageChange,
          }}
        />
      </Card>

      {/* 用户详情抽屉 */}
      <Drawer
        title="用户详情"
        placement="right"
        width={600}
        onClose={() => setDrawerVisible(false)}
        open={drawerVisible}
        extra={
          <Space>
            <Button onClick={handleAdjustQuota}>调整额度</Button>
            {selectedUser?.status === 'active' ? (
              <Button
                danger
                onClick={() => handleDisableUser(selectedUser.id)}
              >
                禁用用户
              </Button>
            ) : (
              <Button
                type="primary"
                onClick={() => handleEnableUser(selectedUser!.id)}
              >
                启用用户
              </Button>
            )}
          </Space>
        }
      >
        {selectedUser && (
          <Descriptions column={1} bordered>
            <Descriptions.Item label="用户ID">
              {selectedUser.id}
            </Descriptions.Item>
            <Descriptions.Item label="用户名">
              {selectedUser.username}
            </Descriptions.Item>
            <Descriptions.Item label="邮箱">
              {selectedUser.email}
            </Descriptions.Item>
            <Descriptions.Item label="状态">
              <StatusTag status={selectedUser.status} />
            </Descriptions.Item>
            <Descriptions.Item label="管理员">
              {selectedUser.is_admin ? (
                <Tag color="blue">是</Tag>
              ) : (
                <Tag>否</Tag>
              )}
            </Descriptions.Item>
            <Descriptions.Item label="总额度">
              {formatNumber(selectedUser.quota)} tokens
            </Descriptions.Item>
            <Descriptions.Item label="已使用额度">
              {formatNumber(selectedUser.used_quota)} tokens
            </Descriptions.Item>
            <Descriptions.Item label="剩余额度">
              {formatNumber(selectedUser.quota - selectedUser.used_quota)} tokens
            </Descriptions.Item>
            <Descriptions.Item label="使用率">
              {formatPercent(selectedUser.used_quota, selectedUser.quota)}
            </Descriptions.Item>
            <Descriptions.Item label="注册时间">
              {formatDateTime(selectedUser.created_at)}
            </Descriptions.Item>
            {selectedUser.last_sign_in && (
              <Descriptions.Item label="最后登录">
                {formatDateTime(selectedUser.last_sign_in)}
              </Descriptions.Item>
            )}
          </Descriptions>
        )}
      </Drawer>

      {/* 调整额度模态框 */}
      <Modal
        title="调整用户额度"
        open={quotaModal.visible}
        onOk={handleQuotaSubmit}
        onCancel={quotaModal.hideModal}
        confirmLoading={updateQuotaMutation.isPending}
      >
        <Form form={quotaModal.form} layout="vertical">
          <Form.Item
            label="新的总额度"
            name="quota"
            rules={[
              { required: true, message: '请输入额度' },
              { type: 'number', min: 0, message: '额度不能为负数' },
            ]}
          >
            <InputNumber
              style={{ width: '100%' }}
              min={0}
              step={1000}
              formatter={(value) =>
                `${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',')
              }
              parser={(value) => value!.replace(/\$\s?|(,*)/g, '') as any}
              addonAfter="tokens"
            />
          </Form.Item>
          {selectedUser && (
            <div style={{ color: '#666', fontSize: 12 }}>
              <p>当前总额度: {formatNumber(selectedUser.quota)} tokens</p>
              <p>已使用额度: {formatNumber(selectedUser.used_quota)} tokens</p>
              <p>剩余额度: {formatNumber(selectedUser.quota - selectedUser.used_quota)} tokens</p>
            </div>
          )}
        </Form>
      </Modal>
    </div>
  );
};

export default UsersPage;
