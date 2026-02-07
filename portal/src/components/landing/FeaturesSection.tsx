import { Card, Typography } from 'antd';
import {
  ApiOutlined,
  ThunderboltOutlined,
  SafetyOutlined,
  RocketOutlined,
  GlobalOutlined,
  DashboardOutlined,
} from '@ant-design/icons';

const { Title, Paragraph } = Typography;

const features = [
  {
    icon: <ApiOutlined className="text-4xl text-primary-600" />,
    title: '统一接口',
    description: '兼容OpenAI、Anthropic、Gemini等主流API格式，一个接口调用所有模型。',
  },
  {
    icon: <ThunderboltOutlined className="text-4xl text-yellow-600" />,
    title: '智能负载均衡',
    description: '支持轮询、加权、最少连接等多种策略，自动故障转移，确保高可用。',
  },
  {
    icon: <SafetyOutlined className="text-4xl text-green-600" />,
    title: '企业级安全',
    description: 'JWT认证、API Key管理、限流保护，全方位保障您的数据安全。',
  },
  {
    icon: <RocketOutlined className="text-4xl text-purple-600" />,
    title: '高性能',
    description: 'Redis缓存、连接池优化、异步处理，毫秒级响应时间。',
  },
  {
    icon: <GlobalOutlined className="text-4xl text-blue-600" />,
    title: '多提供商支持',
    description: '支持OpenAI、Anthropic、Gemini及自定义中转站，灵活配置。',
  },
  {
    icon: <DashboardOutlined className="text-4xl text-red-600" />,
    title: '完善的监控',
    description: '实时统计、请求日志、使用分析，全面掌握API使用情况。',
  },
];

const FeaturesSection = () => {
  return (
    <section id="features" className="py-20 bg-white">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        {/* Section Header */}
        <div className="text-center mb-16">
          <div className="inline-block px-4 py-2 rounded-full bg-primary-100 text-primary-700 text-sm font-medium mb-4">
            核心功能
          </div>
          <Title level={2} className="!text-4xl !font-bold !mb-4">
            为什么选择我们
          </Title>
          <Paragraph className="text-xl text-gray-600 max-w-2xl mx-auto">
            专业的API聚合解决方案，让您的AI应用开发更简单、更高效
          </Paragraph>
        </div>

        {/* Features Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
          {features.map((feature, index) => (
            <Card
              key={index}
              className="border border-gray-200 hover:border-primary-400 hover:shadow-lg transition-all duration-300 cursor-pointer group"
              bodyStyle={{ padding: '32px' }}
            >
              <div className="mb-4 transform group-hover:scale-110 transition-transform duration-300">
                {feature.icon}
              </div>
              <Title level={4} className="!mb-3 !font-semibold">
                {feature.title}
              </Title>
              <Paragraph className="text-gray-600 !mb-0">
                {feature.description}
              </Paragraph>
            </Card>
          ))}
        </div>

        {/* Additional Info */}
        <div className="mt-16 text-center">
          <Card className="bg-gradient-to-r from-primary-50 to-purple-50 border-0">
            <div className="py-8">
              <Title level={3} className="!mb-4">
                还有更多强大功能等你探索
              </Title>
              <Paragraph className="text-lg text-gray-600 !mb-6">
                额度管理、签到奖励、详细文档、SDK支持...
              </Paragraph>
              <a
                href="/docs"
                className="text-primary-600 hover:text-primary-700 font-semibold text-lg"
              >
                查看完整功能列表 →
              </a>
            </div>
          </Card>
        </div>
      </div>
    </section>
  );
};

export default FeaturesSection;
