import React, { useState } from 'react';
import {
  Card,
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
import type { ModelEndpoint, LoadBalancerConfig } from '../services/loadBalancerService';
import type { ColumnsType } from 'antd/es/table';

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
    <div>
      {/* 模型选择和配置 */}
      <Card style={{ marginBottom: 16 }}>
        <Space size="large" wrap>
          <div>
            <div style={{ marginBottom: 8, fontWeight: 500 }}>选择模型</div>
            <Select
              style={{ width: 300 }}
              placeholder="选择要配置的模型"
              value={selectedModel}
              onChange={setSelectedModel}
              showSearch
              filterOption={(input, option) =>
                (option?.children as string)
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
            <>
              <div>
                <div style={{ marginBottom: 8, fontWeight: 500 }}>负载均衡策略</div>
                <Space>
                  <Tag color="blue">
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
            </>
          )}
        </Space>
      </Card>

      {/* 统计卡片 */}
      {selectedModel && (
        <Row gutter={16} style={{ marginBottom: 16 }}>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="总端点数"
                value={stats.total}
                valueStyle={{ color: '#1890ff' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="健康端点"
                value={stats.healthy}
                suffix={`/ ${stats.total}`}
                valueStyle={{ color: '#52c41a' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="启用端点"
                value={stats.active}
                suffix={`/ ${stats.total}`}
                valueStyle={{ color: '#722ed1' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="平均响应时间"
                value={stats.avgResponseTime}
                suffix="ms"
                valueStyle={{ color: '#fa8c16' }}
              />
            </Card>
          </Col>
        </Row>
      )}

      {/* 端点列表 */}
      {selectedModel ? (
        <Card title={`${selectedModel} 的可用端点`}>
          <Table
            columns={endpointColumns}
            dataSource={endpoints || []}
            rowKey="config_id"
            loading={endpointsLoading}
            pagination={false}
          />
        </Card>
      ) : (
        <Card>
          <div style={{ textAlign: 'center', padding: '60px 0', color: '#999' }}>
            <SettingOutlined style={{ fontSize: 48, marginBottom: 16 }} />
            <div>请选择一个模型以查看和配置负载均衡</div>
          </div>
        </Card>
      )}

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
    </div>
  );
};

export default LoadBalancerPage;
