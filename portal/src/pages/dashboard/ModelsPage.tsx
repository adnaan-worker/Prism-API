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
  openai: { color: '#00a67e', icon: <RobotOutlined /> }, // OpenAI Green
  anthropic: { color: '#d97757', icon: <RobotOutlined /> },
  gemini: { color: '#4285f4', icon: <RobotOutlined /> }, // Google Blue
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
        <div className="flex items-center gap-1 text-green-500 text-xs font-medium bg-green-500/10 px-2 py-0.5 rounded-full border border-green-500/20">
          <CheckCircleOutlined /> Active
        </div>
      );
    }
    return (
      <div className="flex items-center gap-1 text-red-500 text-xs font-medium bg-red-500/10 px-2 py-0.5 rounded-full border border-red-500/20">
        <CloseCircleOutlined /> Inactive
      </div>
    );
  };

  return (
    <div className="max-w-7xl mx-auto space-y-8 animate-fade-in pb-12">
      <div className="mb-8">
        <h2 className="text-3xl font-bold text-text-primary mb-2">模型列表</h2>
        <p className="text-text-secondary text-lg">
          浏览所有可用的AI模型，包括GPT-4、Claude、Gemini等。选择合适的模型来满足您的需求。
        </p>
      </div>

      {/* Statistics */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="glass-card p-6 rounded-2xl flex items-center gap-4">
          <div className="w-12 h-12 rounded-xl bg-blue-500/10 flex items-center justify-center text-blue-500 text-2xl">
            <RobotOutlined />
          </div>
          <div>
            <p className="text-slate-500 dark:text-gray-500 text-sm">总模型数</p>
            <p className="text-2xl font-bold text-slate-900 dark:text-white">{models.length}</p>
          </div>
        </div>
        <div className="glass-card p-6 rounded-2xl flex items-center gap-4">
          <div className="w-12 h-12 rounded-xl bg-green-500/10 flex items-center justify-center text-green-500 text-2xl">
            <CheckCircleOutlined />
          </div>
          <div>
            <p className="text-slate-500 dark:text-gray-500 text-sm">可用模型</p>
            <p className="text-2xl font-bold text-slate-900 dark:text-white">{activeModelsCount}</p>
          </div>
        </div>
        <div className="glass-card p-6 rounded-2xl flex items-center gap-4">
          <div className="w-12 h-12 rounded-xl bg-purple-500/10 flex items-center justify-center text-purple-500 text-2xl">
            <ApiOutlined />
          </div>
          <div>
            <p className="text-slate-500 dark:text-gray-500 text-sm">API配置数</p>
            <p className="text-2xl font-bold text-slate-900 dark:text-white">{totalConfigsCount}</p>
          </div>
        </div>
      </div>

      {/* Search and Filter */}
      <div className="glass-card p-4 rounded-2xl">
        <div className="flex flex-col md:flex-row gap-4">
          <div className="flex-1">
            <div className="bg-white dark:bg-black/40 border border-border rounded-xl px-4 py-3 flex items-center focus-within:ring-2 focus-within:ring-primary/50 transition-all">
              <SearchOutlined className="text-text-tertiary mr-3" />
              <input
                type="text"
                placeholder="搜索模型名称、描述或提供商..."
                className="bg-transparent border-none outline-none text-slate-900 dark:text-white w-full placeholder-text-tertiary"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
              />
            </div>
          </div>
          <div className="w-full md:w-64">
            <Select
              size="large"
              style={{ width: '100%', height: '50px' }}
              className="custom-select"
              placeholder="筛选提供商"
              value={selectedProvider}
              onChange={setSelectedProvider}
              suffixIcon={<FilterOutlined className="text-text-tertiary" />}
              variant="borderless"
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
          </div>
          <div>
            <Button
              size="large"
              icon={<ReloadOutlined />}
              onClick={() => refetch()}
              loading={isLoading}
              className="h-[50px] px-6 rounded-xl bg-white/5 border-border/40 text-text-primary hover:bg-white/10"
            >
              刷新
            </Button>
          </div>
        </div>
      </div>

      {/* Models Grid */}
      {isLoading ? (
        <div className="flex flex-col items-center justify-center py-20">
          <Spin size="large" />
          <p className="text-text-tertiary mt-4">加载模型列表...</p>
        </div>
      ) : filteredModels.length === 0 ? (
        <div className="glass-card p-12 rounded-2xl text-center">
          <Empty
            description={
              <span className="text-text-tertiary">
                {searchQuery || selectedProvider !== 'all'
                  ? '没有找到匹配的模型'
                  : '暂无可用模型'}
              </span>
            }
          />
        </div>
      ) : (
        <>
          <div className="flex items-center gap-2 mb-4 bg-blue-500/10 border border-blue-500/20 px-4 py-2 rounded-lg inline-flex">
            <CheckCircleOutlined className="text-blue-500" />
            <span className="text-blue-500 font-medium">找到 {filteredModels.length} 个模型</span>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            {filteredModels.map((model) => (
              <div key={model.name} className="group relative">
                {/* Ribbon */}
                <div className="absolute top-0 right-0 z-10">
                  <div
                    className="px-3 py-1 rounded-bl-xl rounded-tr-xl text-xs font-bold text-white shadow-sm flex items-center gap-1"
                    style={{ backgroundColor: getProviderColor(model.provider) }}
                  >
                    {getProviderIcon(model.provider)}
                    <span className="capitalize">{model.provider}</span>
                  </div>
                </div>

                <div className="glass-card h-full p-6 rounded-2xl hover:border-primary/50 transition-all duration-300 flex flex-col group-hover:shadow-[0_0_30px_rgba(14,165,233,0.15)] group-hover:-translate-y-1">
                  <div className="flex-1">
                    <div className="mb-4 pr-16">
                      <h3 className="text-lg font-bold text-slate-900 dark:text-white mb-2 line-clamp-1" title={model.name}>
                        {model.name}
                      </h3>
                      {getStatusTag(model.status)}
                    </div>

                    <div className="mb-4">
                      <span className="px-2 py-1 bg-slate-100 dark:bg-white/5 rounded text-xs font-mono text-slate-600 dark:text-text-secondary border border-slate-200 dark:border-border/40">
                        {model.type}
                      </span>
                    </div>

                    <p className="text-text-tertiary text-sm line-clamp-3 mb-4">
                      {model.description || '暂无描述'}
                    </p>
                  </div>

                  <div className="pt-4 mt-auto border-t border-border/40 flex items-center justify-between">
                    <Tooltip title="可用的API配置数量，用于负载均衡">
                      <div className="flex items-center gap-2 text-text-secondary font-medium">
                        <ThunderboltOutlined className="text-amber-500" />
                        <span>{model.config_count} 配置</span>
                      </div>
                    </Tooltip>
                    <Button type="text" size="small" className="text-primary hover:text-primary-400 p-0 h-auto">
                      调用 &rarr;
                    </Button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </>
      )}

      {/* Help Section */}
      {!isLoading && models.length > 0 && (
        <div className="glass-card p-6 rounded-2xl bg-gradient-to-r from-black/5 to-transparent dark:from-white/5 border-l-4 border-l-primary">
          <h4 className="font-bold text-text-primary mb-2 flex items-center gap-2">
            <ApiOutlined /> 如何使用这些模型？
          </h4>
          <div className="space-y-1 text-sm text-text-secondary">
            <p>1. 在【API密钥】页面创建您的API密钥</p>
            <p>2. 使用OpenAI兼容的接口格式调用任意模型</p>
            <p>3. 查看【使用文档】了解详细的调用示例</p>
          </div>
        </div>
      )}
    </div>
  );
};

export default ModelsPage;
