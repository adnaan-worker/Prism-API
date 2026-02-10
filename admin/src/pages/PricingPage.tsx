import React, { useState } from 'react';
import {
  Card,
  Table,
  Button,
  Space,
  Tag,
  Modal,
  Form,
  Input,
  InputNumber,
  message,
  Popconfirm,
  Switch,
  Select,
  Tooltip,
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  ReloadOutlined,
  DollarOutlined,
  ApiOutlined,
} from '@ant-design/icons';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { pricingService } from '../services/pricingService';
import { apiConfigService } from '../services/apiConfigService';
import type { Pricing } from '../types';
import type { ColumnsType } from 'antd/es/table';

const { Option } = Select;
const { TextArea } = Input;

const PricingPage: React.FC = () => {
  const [modalVisible, setModalVisible] = useState(false);
  const [editingPricing, setEditingPricing] = useState<Pricing | null>(null);
  const [apiConfigFilter, setApiConfigFilter] = useState<number | undefined>();
  const [form] = Form.useForm();

  const queryClient = useQueryClient();

  // 获取 API 配置列表
  const { data: apiConfigsData } = useQuery({
    queryKey: ['api-configs'],
    queryFn: () => apiConfigService.getConfigs({}),
  });

  // 获取定价列表
  const { data, isLoading, refetch } = useQuery({
    queryKey: ['pricings'],
    queryFn: pricingService.getAllPricings,
  });

  // 创建定价
  const createMutation = useMutation({
    mutationFn: pricingService.createPricing,
    onSuccess: () => {
      message.success('定价配置创建成功');
      queryClient.invalidateQueries({ queryKey: ['pricings'] });
      handleModalClose();
    },
    onError: (error: any) => {
      message.error(error.response?.data?.error?.message || '定价配置创建失败');
    },
  });

  // 更新定价
  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: any }) =>
      pricingService.updatePricing(id, data),
    onSuccess: () => {
      message.success('定价配置更新成功');
      queryClient.invalidateQueries({ queryKey: ['pricings'] });
      handleModalClose();
    },
    onError: (error: any) => {
      message.error(error.response?.data?.error?.message || '定价配置更新失败');
    },
  });

  // 删除定价
  const deleteMutation = useMutation({
    mutationFn: pricingService.deletePricing,
    onSuccess: () => {
      message.success('定价配置删除成功');
      queryClient.invalidateQueries({ queryKey: ['pricings'] });
    },
    onError: () => {
      message.error('定价配置删除失败');
    },
  });

  // 表格列配置
  const columns: ColumnsType<Pricing> = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
    },
    {
      title: 'API 配置',
      dataIndex: 'api_config',
      key: 'api_config',
      render: (apiConfig: any) => (
        <Space>
          <ApiOutlined />
          <span>{apiConfig?.name || '-'}</span>
        </Space>
      ),
    },
    {
      title: '模型名称',
      dataIndex: 'model_name',
      key: 'model_name',
      render: (text: string) => (
        <Space>
          <DollarOutlined />
          <strong>{text}</strong>
        </Space>
      ),
    },
    {
      title: '输入价格',
      dataIndex: 'input_price',
      key: 'input_price',
      render: (price: number, record) => (
        <Tooltip title={`每 ${record.unit} tokens`}>
          <span>{price.toFixed(4)} {record.currency}</span>
        </Tooltip>
      ),
      sorter: (a, b) => a.input_price - b.input_price,
    },
    {
      title: '输出价格',
      dataIndex: 'output_price',
      key: 'output_price',
      render: (price: number, record) => (
        <Tooltip title={`每 ${record.unit} tokens`}>
          <span>{price.toFixed(4)} {record.currency}</span>
        </Tooltip>
      ),
      sorter: (a, b) => a.output_price - b.output_price,
    },
    {
      title: '计价单位',
      dataIndex: 'unit',
      key: 'unit',
      width: 100,
      render: (unit: number) => `${unit} tokens`,
    },
    {
      title: '货币',
      dataIndex: 'currency',
      key: 'currency',
      width: 100,
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
            updateMutation.mutate({
              id: record.id,
              data: { is_active: checked },
            })
          }
          checkedChildren="启用"
          unCheckedChildren="禁用"
        />
      ),
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date: string) => new Date(date).toLocaleString('zh-CN'),
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
            title="确定要删除该定价配置吗？"
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
    setEditingPricing(null);
    form.resetFields();
    form.setFieldsValue({
      currency: 'credits',
      unit: 1000,
      is_active: true,
    });
    setModalVisible(true);
  };

  // 打开编辑模态框
  const handleEdit = (pricing: Pricing) => {
    setEditingPricing(pricing);
    form.setFieldsValue({
      ...pricing,
      api_config_id: pricing.api_config_id,
    });
    setModalVisible(true);
  };

  // 关闭模态框
  const handleModalClose = () => {
    setModalVisible(false);
    setEditingPricing(null);
    form.resetFields();
  };

  // 提交表单
  const handleSubmit = () => {
    form.validateFields().then((values) => {
      if (editingPricing) {
        updateMutation.mutate({ id: editingPricing.id, data: values });
      } else {
        createMutation.mutate(values);
      }
    });
  };

  // 过滤数据
  const filteredData = apiConfigFilter
    ? data?.pricings.filter((p) => p.api_config_id === apiConfigFilter)
    : data?.pricings;

  return (
    <div>
      <Card>
        {/* 操作栏 */}
        <Space style={{ marginBottom: 16 }} wrap>
          <Button type="primary" icon={<PlusOutlined />} onClick={handleAdd}>
            添加定价
          </Button>
          <Select
            placeholder="筛选 API 配置"
            allowClear
            style={{ width: 200 }}
            onChange={setApiConfigFilter}
          >
            {apiConfigsData?.configs.map((config) => (
              <Option key={config.id} value={config.id}>
                {config.name}
              </Option>
            ))}
          </Select>
          <Button icon={<ReloadOutlined />} onClick={() => refetch()}>
            刷新
          </Button>
        </Space>

        {/* 定价列表表格 */}
        <Table
          columns={columns}
          dataSource={filteredData || []}
          rowKey="id"
          loading={isLoading}
          scroll={{ x: 1400 }}
          pagination={{
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 条记录`,
          }}
        />
      </Card>

      {/* 添加/编辑模态框 */}
      <Modal
        title={editingPricing ? '编辑定价配置' : '添加定价配置'}
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={handleModalClose}
        confirmLoading={createMutation.isPending || updateMutation.isPending}
        width={600}
      >
        <Form form={form} layout="vertical">
          <Form.Item
            label="API 配置"
            name="api_config_id"
            rules={[{ required: true, message: '请选择 API 配置' }]}
          >
            <Select
              placeholder="选择 API 配置"
              showSearch
              optionFilterProp="children"
              disabled={!!editingPricing}
            >
              {apiConfigsData?.configs.map((config) => (
                <Option key={config.id} value={config.id}>
                  <Space>
                    <Tag color={
                      config.type === 'openai' ? 'blue' :
                      config.type === 'anthropic' ? 'orange' :
                      config.type === 'gemini' ? 'green' : 'purple'
                    }>
                      {config.type.toUpperCase()}
                    </Tag>
                    {config.name}
                  </Space>
                </Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item
            label="模型名称"
            name="model_name"
            rules={[{ required: true, message: '请输入模型名称' }]}
            extra="请输入该 API 配置支持的模型名称"
          >
            <Input placeholder="例如: gpt-4, claude-3-opus" disabled={!!editingPricing} />
          </Form.Item>

          <Form.Item
            label="输入价格"
            name="input_price"
            rules={[{ required: true, message: '请输入输入价格' }]}
            extra="每1000个输入tokens的价格"
          >
            <InputNumber
              min={0}
              step={0.01}
              precision={4}
              style={{ width: '100%' }}
              placeholder="0.0000"
            />
          </Form.Item>

          <Form.Item
            label="输出价格"
            name="output_price"
            rules={[{ required: true, message: '请输入输出价格' }]}
            extra="每1000个输出tokens的价格"
          >
            <InputNumber
              min={0}
              step={0.01}
              precision={4}
              style={{ width: '100%' }}
              placeholder="0.0000"
            />
          </Form.Item>

          <Form.Item label="货币单位" name="currency">
            <Select>
              <Option value="credits">积分 (Credits)</Option>
              <Option value="usd">美元 (USD)</Option>
              <Option value="cny">人民币 (CNY)</Option>
            </Select>
          </Form.Item>

          <Form.Item label="计价单位" name="unit" extra="通常为1000 tokens">
            <InputNumber
              min={1}
              style={{ width: '100%' }}
              placeholder="1000"
            />
          </Form.Item>

          <Form.Item label="状态" name="is_active" valuePropName="checked">
            <Switch checkedChildren="启用" unCheckedChildren="禁用" />
          </Form.Item>

          <Form.Item label="描述" name="description">
            <TextArea rows={3} placeholder="定价配置的描述信息" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default PricingPage;
