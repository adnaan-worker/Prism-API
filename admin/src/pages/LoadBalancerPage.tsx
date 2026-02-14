import React, { useState } from 'react';
import {
  Row,
  Col,
  Select,
  Table,
  Button,
  Space,
  Tag,
  Modal,
  Form,
  Input,
  message,
  Descriptions,
  Progress,
  Statistic,
} from 'antd';
import {
  ReloadOutlined,
  SettingOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  QuestionCircleOutlined,
} from '@ant-design/icons';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { loadBalancerService } from '../services/loadBalancerService';
import type { ModelEndpoint } from '../services/loadBalancerService';
import type { ColumnsType } from 'antd/es/table';
import PageContainer from '../components/PageContainer';

const { Option } = Select;

const LoadBalancerPage: React.FC = () => {
  const [selectedModel, setSelectedModel] = useState<string | undefined>();
  const [configModalVisible, setConfigModalVisible] = useState(false);
  const [form] = Form.useForm();

  const queryClient = useQueryClient();

  // 获取可用模型列表
  const { data: models } = useQuery({
    queryKey: ['available-models'],
    queryFn: loadBalancerService.getAvailableModels,
  });

  // 获取负载均衡配置列表
  const { data: configs } = useQuery({
    queryKey: ['load-balancer-configs'],
    queryFn: loadBalancerService.getConfigs,
  });

  // 获取选中模型的端点列表
  const { data: endpoints, isLoading: endpointsLoading, refetch: refetchEndpoints } = useQuery({
    queryKey: ['model-endpoints', selectedModel],
    queryFn: () => loadBalancerService.getModelEndpoints(selectedModel!),
    enabled: !!selectedModel,
  });

  // 创建负载均衡配置
  const createConfigMutation = useMutation({
    mutationFn: loadBalancerService.createConfig,
    onSuccess: () => {
      message.success('负载均衡配置创建成功');
      queryClient.invalidateQueries({ queryKey: ['load-balancer-configs'] });
      setConfigModalVisible(false);
      form.resetFields();
    },
    onError: () => {
      message.error('负载均衡配置创建失败');
    },
  });

  // 更新负载均衡配置
  const updateConfigMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: any }) =>
      loadBalancerService.updateConfig(id, data),
    onSuccess: () => {
      message.success('负载均衡配置更新成功');
      queryClient.invalidateQueries({ queryKey: ['load-balancer-configs'] });
    },
    onError: () => {
      message.error('负载均衡配置更新失败');
    },
  });

  // 端点列表表格列配置
  const endpointColumns: ColumnsType<ModelEndpoint> = [
    {
      title: '配置名称',
      dataIndex: 'config_name',
      key: 'config_name',
    },
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      render: (type: string) => {
        const colorMap: Record<string, string> = {
          openai: 'blue',
          anthropic: 'orange',
          gemini: 'green',
          custom: 'purple',
        };
        return <Tag color={colorMap[type] || 'default'}>{type.toUpperCase()}</Tag>;
      },
    },
    {
      title: 'Base URL',
      dataIndex: 'base_url',
      key: 'base_url',
      ellipsis: true,
    },
    {
      title: '优先级',
      dataIndex: 'priority',
      key: 'priority',
      sorter: (a, b) => a.priority - b.priority,
    },
    {
      title: '权重',
      dataIndex: 'weight',
      key: 'weight',
    },
    {
      title: '健康状态',
      dataIndex: 'health_status',
      key: 'health_status',
      render: (status: string) => {
        const statusConfig = {
          healthy: { color: 'success', icon: <CheckCircleOutlined />, text: '健康' },
          unhealthy: { color: 'error', icon: <CloseCircleOutlined />, text: '异常' },
          unknown: { color: 'default', icon: <QuestionCircleOutlined />, text: '未知' },
        };
        const config = statusConfig[status as keyof typeof statusConfig] || statusConfig.unknown;
        return (
          <Tag color={config.color} icon={config.icon}>
            {config.text}
          </Tag>
        );
      },
    },
    {
      title: '响应时间',
      dataIndex: 'response_time',
      key: 'response_time',
      render: (time?: number) => (time ? `${time}ms` : '-'),
    },
    {
      title: '成功率',
      dataIndex: 'success_rate',
      key: 'success_rate',
      render: (rate?: number) =>
        rate !== undefined ? (
          <Progress
            percent={rate}
            size="small"
            status={rate >= 95 ? 'success' : rate >= 80 ? 'normal' : 'exception'}
          />
        ) : (
          '-'
        ),
    },
    {
      title: '状态',
      dataIndex: 'is_active',
      key: 'is_active',
      render: (is_active: boolean) => (
        <Tag color={is_active ? 'green' : 'red'}>
          {is_active ? '启用' : '禁用'}
        </Tag>
      ),
    },
  ];

  // 获取当前模型的负载均衡配置
  const currentConfig = configs?.find((c) => c.model_name === selectedModel);

  // 策略选项
  const strategyOptions = [
    { value: 'round_robin', label: '轮询 (Round Robin)' },
    { value: 'weighted_round_robin', label: '加权轮询 (Weighted Round Robin)' },
    { value: 'least_connections', label: '最少连接 (Least Connections)' },
    { value: 'random', label: '随机 (Random)' },
  ];

  // 策略说明
  const strategyDescriptions: Record<string, string> = {
    round_robin: '按顺序依次分配请求到每个端点',
    weighted_round_robin: '根据权重值按比例分配请求',
    least_connections: '优先分配到当前连接数最少的端点',
    random: '随机选择一个端点处理请求',
  };

  // 打开配置模态框
  const handleOpenConfigModal = () => {
    if (!selectedModel) {
      message.warning('请先选择模型');
      return;
    }
    form.setFieldsValue({
      model_name: selectedModel,
      strategy: currentConfig?.strategy || 'round_robin',
    });
    setConfigModalVisible(true);
  };

  // 提交配置
  const handleSubmitConfig = () => {
    form.validateFields().then((values) => {
      if (currentConfig) {
        updateConfigMutation.mutate({
          id: currentConfig.id,
          data: { strategy: values.strategy },
        });
      } else {
        createConfigMutation.mutate(values);
      }
      setConfigModalVisible(false);
    });
  };

  // 计算统计数据
  const stats = {
    total: endpoints?.length || 0,
    healthy: endpoints?.filter((e) => e.health_status === 'healthy').length || 0,
    active: endpoints?.filter((e) => e.is_active).length || 0,
    avgResponseTime:
      endpoints && endpoints.length > 0
        ? Math.round(
          endpoints.reduce((sum, e) => sum + (e.response_time || 0), 0) /
          endpoints.length
        )
        : 0,
  };

  return (
    <PageContainer title="负载均衡" description="管理模型端点分配和均衡策略">
      {/* 模型选择和配置 */}
      <div className="glass-card p-6 mb-6">
        <Space size="large" wrap className="w-full justify-between">
          <div>
            <div className="mb-2 font-medium text-text-primary">选择模型</div>
            <Select
              style={{ width: 300 }}
              placeholder="选择要配置的模型"
              value={selectedModel}
              onChange={setSelectedModel}
              showSearch
              filterOption={(input, option) =>
                String(option?.children)
                  .toLowerCase()
                  .includes(input.toLowerCase())
              }
            >
              {models?.map((model) => (
                <Option key={model} value={model}>
                  {model}
                </Option>
              ))}
            </Select>
          </div>

          {selectedModel && (
            <div className="flex items-end gap-4">
              <div>
                <div className="mb-2 font-medium text-text-primary">负载均衡策略</div>
                <Space>
                  <Tag color="blue" className="mr-0">
                    {strategyOptions.find((s) => s.value === currentConfig?.strategy)
                      ?.label || '未配置'}
                  </Tag>
                  <Button
                    icon={<SettingOutlined />}
                    onClick={handleOpenConfigModal}
                  >
                    配置策略
                  </Button>
                </Space>
              </div>

              <Button
                icon={<ReloadOutlined />}
                onClick={() => refetchEndpoints()}
              >
                刷新状态
              </Button>
            </div>
          )}
        </Space>
      </div >

      {/* 统计卡片 */}
      {
        selectedModel && (
          <Row gutter={[16, 16]} className="mb-6">
            <Col xs={24} sm={12} lg={6}>
              <div className="glass-card p-6">
                <Statistic
                  title={<span className="text-text-secondary">总端点数</span>}
                  value={stats.total}
                  valueStyle={{ color: '#38bdf8', fontWeight: 600 }}
                />
              </div>
            </Col>
            <Col xs={24} sm={12} lg={6}>
              <div className="glass-card p-6">
                <Statistic
                  title={<span className="text-text-secondary">健康端点</span>}
                  value={stats.healthy}
                  suffix={<span className="text-text-tertiary text-sm">/ {stats.total}</span>}
                  valueStyle={{ color: '#4ade80', fontWeight: 600 }}
                />
              </div>
            </Col>
            <Col xs={24} sm={12} lg={6}>
              <div className="glass-card p-6">
                <Statistic
                  title={<span className="text-text-secondary">启用端点</span>}
                  value={stats.active}
                  suffix={<span className="text-text-tertiary text-sm">/ {stats.total}</span>}
                  valueStyle={{ color: '#a78bfa', fontWeight: 600 }}
                />
              </div>
            </Col>
            <Col xs={24} sm={12} lg={6}>
              <div className="glass-card p-6">
                <Statistic
                  title={<span className="text-text-secondary">平均响应时间</span>}
                  value={stats.avgResponseTime}
                  suffix={<span className="text-sm">ms</span>}
                  valueStyle={{ color: '#fbbf24', fontWeight: 600 }}
                />
              </div>
            </Col>
          </Row>
        )
      }

      {/* 端点列表 */}
      {
        selectedModel ? (
          <div className="glass-card p-6">
            <div className="mb-4 font-bold text-lg text-text-primary">{selectedModel} 的可用端点</div>
            <Table
              columns={endpointColumns}
              dataSource={endpoints || []}
              rowKey={(record) => `${record.config_id}-${record.base_url}`}
              loading={endpointsLoading}
              pagination={false}
              size="middle"
            />
          </div>
        ) : (
          <div className="glass-card p-12 text-center text-text-tertiary">
            <SettingOutlined style={{ fontSize: 48, marginBottom: 16 }} />
            <div>请选择一个模型以查看和配置负载均衡</div>
          </div>
        )
      }

      {/* 配置策略模态框 */}
      <Modal
        title="配置负载均衡策略"
        open={configModalVisible}
        onOk={handleSubmitConfig}
        onCancel={() => setConfigModalVisible(false)}
        confirmLoading={
          createConfigMutation.isPending || updateConfigMutation.isPending
        }
      >
        <Form form={form} layout="vertical">
          <Form.Item label="模型" name="model_name">
            <Input disabled />
          </Form.Item>

          <Form.Item
            label="负载均衡策略"
            name="strategy"
            rules={[{ required: true, message: '请选择策略' }]}
          >
            <Select>
              {strategyOptions.map((option) => (
                <Option key={option.value} value={option.value}>
                  {option.label}
                </Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item noStyle shouldUpdate={(prev, curr) => prev.strategy !== curr.strategy}>
            {({ getFieldValue }) => {
              const strategy = getFieldValue('strategy');
              return strategy ? (
                <Descriptions column={1} bordered size="small">
                  <Descriptions.Item label="策略说明">
                    {strategyDescriptions[strategy]}
                  </Descriptions.Item>
                </Descriptions>
              ) : null;
            }}
          </Form.Item>
        </Form>
      </Modal>
    </PageContainer >
  );
};

export default LoadBalancerPage;
