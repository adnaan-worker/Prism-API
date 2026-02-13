import { useState, useMemo } from 'react';
import {
  Typography,
  Card,
  Row,
  Col,
  Input,
  Select,
  Space,
  Tag,
  Empty,
  Spin,
  Badge,
  Tooltip,
  Alert,
  Statistic,
  Button,
} from 'antd';
import {
  SearchOutlined,
  ApiOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  ThunderboltOutlined,
  RobotOutlined,
  FilterOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import { useQuery } from '@tanstack/react-query';
import { apiClient } from '../../lib/api';
import type { Model } from '../../types';

const { Title, Paragraph, Text } = Typography;
const { Search } = Input;

interface GetModelsResponse {
  models: Model[];
  total: number;
}

// Provider colors and icons
const providerConfig: Record<string, { color: string; icon: React.ReactNode }> = {
  openai: { color: '#10a37f', icon: <RobotOutlined /> },
  anthropic: { color: '#d97757', icon: <RobotOutlined /> },
  gemini: { color: '#4285f4', icon: <RobotOutlined /> },
  custom: { color: '#8c8c8c', icon: <ApiOutlined /> },
};

const ModelsPage = () => {
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedProvider, setSelectedProvider] = useState<string>('all');

  // Fetch models
  const { data: modelsData, isLoading, refetch } = useQuery({
    queryKey: ['models'],
    queryFn: async () => {
      const response = await apiClient.get<GetModelsResponse>('/models');
      return response.data;
    },
  });

  const models = modelsData?.models || [];

  // Get unique providers for filter
  const providers = useMemo(() => {
    const uniqueProviders = new Set(models.map((model) => model.provider));
    return Array.from(uniqueProviders).sort();
  }, [models]);

  // Filter models based on search and provider
  const filteredModels = useMemo(() => {
    return models.filter((model) => {
      const matchesSearch =
        searchQuery === '' ||
        model.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        model.description.toLowerCase().includes(searchQuery.toLowerCase()) ||
        model.provider.toLowerCase().includes(searchQuery.toLowerCase());

      const matchesProvider =
        selectedProvider === 'all' || model.provider === selectedProvider;

      return matchesSearch && matchesProvider;
    });
  }, [models, searchQuery, selectedProvider]);

  // Statistics
  const activeModelsCount = models.filter((m) => m.status === 'active').length;
  const totalConfigsCount = models.reduce((sum, m) => sum + m.config_count, 0);

  const getProviderColor = (provider: string) => {
    return providerConfig[provider.toLowerCase()]?.color || '#8c8c8c';
  };

  const getProviderIcon = (provider: string) => {
    return providerConfig[provider.toLowerCase()]?.icon || <ApiOutlined />;
  };

  const getStatusTag = (status: string) => {
    if (status === 'active') {
      return (
        <Tag icon={<CheckCircleOutlined />} color="success">
          可用
        </Tag>
      );
    }
    return (
      <Tag icon={<CloseCircleOutlined />} color="error">
        不可用
      </Tag>
    );
  };

  return (
    <div style={{ padding: '0 24px' }}>
      <div style={{ marginBottom: 24 }}>
        <Title level={2}>模型列表</Title>
        <Paragraph type="secondary">
          浏览所有可用的AI模型，包括GPT-4、Claude、Gemini等。选择合适的模型来满足您的需求。
        </Paragraph>
      </div>

      {/* Statistics */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={8}>
          <Card>
            <Statistic
              title="总模型数"
              value={models.length}
              prefix={<RobotOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card>
            <Statistic
              title="可用模型"
              value={activeModelsCount}
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card>
            <Statistic
              title="API配置数"
              value={totalConfigsCount}
              prefix={<ApiOutlined />}
              valueStyle={{ color: '#722ed1' }}
            />
          </Card>
        </Col>
      </Row>

      {/* Search and Filter */}
      <Card style={{ marginBottom: 24 }}>
        <Row gutter={[16, 16]}>
          <Col xs={24} md={12}>
            <Search
              placeholder="搜索模型名称、描述或提供商..."
              allowClear
              size="large"
              prefix={<SearchOutlined />}
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              onSearch={setSearchQuery}
            />
          </Col>
          <Col xs={24} md={8}>
            <Select
              size="large"
              style={{ width: '100%' }}
              placeholder="筛选提供商"
              value={selectedProvider}
              onChange={setSelectedProvider}
              suffixIcon={<FilterOutlined />}
            >
              <Select.Option value="all">全部提供商</Select.Option>
              {providers.map((provider) => (
                <Select.Option key={provider} value={provider}>
                  <Space>
                    {getProviderIcon(provider)}
                    <span style={{ textTransform: 'capitalize' }}>{provider}</span>
                  </Space>
                </Select.Option>
              ))}
            </Select>
          </Col>
          <Col xs={24} md={4}>
            <Button
              size="large"
              icon={<ReloadOutlined />}
              onClick={() => refetch()}
              loading={isLoading}
              style={{ width: '100%' }}
            >
              刷新
            </Button>
          </Col>
        </Row>
      </Card>

      {/* Models Grid */}
      {isLoading ? (
        <div style={{ textAlign: 'center', padding: '60px 0' }}>
          <Spin size="large">
            <div style={{ paddingTop: 50 }}>加载模型列表...</div>
          </Spin>
        </div>
      ) : filteredModels.length === 0 ? (
        <Card>
          <Empty
            description={
              searchQuery || selectedProvider !== 'all'
                ? '没有找到匹配的模型'
                : '暂无可用模型'
            }
          />
        </Card>
      ) : (
        <>
          <Alert
            message={`找到 ${filteredModels.length} 个模型`}
            type="info"
            showIcon
            style={{ marginBottom: 16 }}
          />
          <Row gutter={[16, 16]}>
            {filteredModels.map((model) => (
              <Col xs={24} sm={12} lg={8} xl={6} key={model.name}>
                <Badge.Ribbon
                  text={
                    <Space size={4}>
                      {getProviderIcon(model.provider)}
                      <span style={{ textTransform: 'capitalize' }}>
                        {model.provider}
                      </span>
                    </Space>
                  }
                  color={getProviderColor(model.provider)}
                >
                  <Card
                    hoverable
                    style={{
                      height: '100%',
                      borderRadius: 8,
                    }}
                    styles={{
                      body: {
                        display: 'flex',
                        flexDirection: 'column',
                        height: '100%',
                      },
                    }}
                  >
                    <div style={{ flex: 1 }}>
                      <Space direction="vertical" size={12} style={{ width: '100%' }}>
                        {/* Model Name */}
                        <div>
                          <Text
                            strong
                            style={{
                              fontSize: 16,
                              display: 'block',
                              marginBottom: 8,
                            }}
                          >
                            {model.name}
                          </Text>
                          {getStatusTag(model.status)}
                        </div>

                        {/* Model Type */}
                        <div>
                          <Tag color="blue">{model.type}</Tag>
                        </div>

                        {/* Description */}
                        <Paragraph
                          type="secondary"
                          ellipsis={{ rows: 3 }}
                          style={{ marginBottom: 0, fontSize: 13 }}
                        >
                          {model.description || '暂无描述'}
                        </Paragraph>
                      </Space>
                    </div>

                    {/* Footer with config count */}
                    <div
                      style={{
                        marginTop: 16,
                        paddingTop: 16,
                        borderTop: '1px solid #f0f0f0',
                      }}
                    >
                      <Tooltip title="可用的API配置数量，用于负载均衡">
                        <Space>
                          <ThunderboltOutlined style={{ color: '#faad14' }} />
                          <Text type="secondary" style={{ fontSize: 13 }}>
                            {model.config_count} 个配置
                          </Text>
                        </Space>
                      </Tooltip>
                    </div>
                  </Card>
                </Badge.Ribbon>
              </Col>
            ))}
          </Row>
        </>
      )}

      {/* Help Section */}
      {!isLoading && models.length > 0 && (
        <Card
          style={{ marginTop: 24, background: '#fafafa' }}
          styles={{ body: { padding: 16 } }}
        >
          <Space direction="vertical" size={8}>
            <Text strong>
              <ApiOutlined /> 如何使用这些模型？
            </Text>
            <Text type="secondary" style={{ fontSize: 13 }}>
              1. 在【API密钥】页面创建您的API密钥
            </Text>
            <Text type="secondary" style={{ fontSize: 13 }}>
              2. 使用OpenAI兼容的接口格式调用任意模型
            </Text>
            <Text type="secondary" style={{ fontSize: 13 }}>
              3. 查看【使用文档】了解详细的调用示例
            </Text>
          </Space>
        </Card>
      )}
    </div>
  );
};

export default ModelsPage;
