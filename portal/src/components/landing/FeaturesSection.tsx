import {
  ApiOutlined,
  ThunderboltOutlined,
  SafetyOutlined,
  RocketOutlined,
  GlobalOutlined,
  DashboardOutlined,
} from '@ant-design/icons';

const FeaturesSection = () => {
  const features = [
    {
      icon: <ApiOutlined className="text-4xl text-primary" />,
      title: '统一接口',
      description: '兼容OpenAI、Anthropic、Gemini等主流API格式，一个接口调用所有模型。',
    },
    {
      icon: <ThunderboltOutlined className="text-4xl text-yellow-500" />,
      title: '智能负载均衡',
      description: '支持轮询、加权、最少连接等多种策略，自动故障转移，确保高可用。',
    },
    {
      icon: <SafetyOutlined className="text-4xl text-green-500" />,
      title: '企业级安全',
      description: 'JWT认证、API Key管理、限流保护，全方位保障您的数据安全。',
    },
    {
      icon: <RocketOutlined className="text-4xl text-purple-500" />,
      title: '高性能',
      description: 'Redis缓存、连接池优化、异步处理，毫秒级响应时间。',
    },
    {
      icon: <GlobalOutlined className="text-4xl text-blue-500" />,
      title: '多提供商支持',
      description: '支持OpenAI、Anthropic、Gemini及自定义中转站，灵活配置。',
    },
    {
      icon: <DashboardOutlined className="text-4xl text-red-500" />,
      title: '完善的监控',
      description: '实时统计、请求日志、使用分析，全面掌握API使用情况。',
    },
  ];

  return (
    <section id="features" className="py-24 bg-slate-50 dark:bg-black relative overflow-hidden">
      {/* Background Glow */}
      <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[800px] h-[800px] bg-primary/5 rounded-full blur-[100px] pointer-events-none"></div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 relative z-10">
        {/* Section Header */}
        <div className="text-center mb-16">
          <div className="inline-block px-4 py-2 rounded-full bg-primary/10 text-primary text-sm font-medium mb-4">
            核心功能
          </div>
          <h2 className="text-4xl font-bold mb-6 text-slate-900 dark:text-white">
            为什么选择我们
          </h2>
          <p className="text-xl text-slate-600 dark:text-slate-400 max-w-2xl mx-auto">
            专业的API聚合解决方案，让您的AI应用开发更简单、更高效
          </p>
        </div>

        {/* Features Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
          {features.map((feature, index) => (
            <div
              key={index}
              className="glass-card p-8 rounded-2xl border border-slate-200 dark:border-white/5 hover:border-primary/50 transition-all duration-300 group hover:-translate-y-1"
            >
              <div className="mb-6 transform group-hover:scale-110 transition-transform duration-300 w-14 h-14 rounded-xl bg-slate-50 dark:bg-white/5 flex items-center justify-center">
                {feature.icon}
              </div>
              <h3 className="text-xl font-bold mb-3 text-slate-900 dark:text-white group-hover:text-primary transition-colors">
                {feature.title}
              </h3>
              <p className="text-slate-600 dark:text-slate-400 leading-relaxed">
                {feature.description}
              </p>
            </div>
          ))}
        </div>

        {/* Additional Info */}
        <div className="mt-20 text-center">
          <div className="glass-card p-10 rounded-3xl bg-gradient-to-r from-primary/5 to-purple-500/5 border border-primary/10">
            <h3 className="text-2xl font-bold mb-4 text-slate-900 dark:text-white">
              还有更多强大功能等你探索
            </h3>
            <p className="text-lg text-slate-600 dark:text-slate-400 mb-8 max-w-2xl mx-auto">
              额度管理、签到奖励、详细文档、SDK支持...
            </p>
            <a
              href="/docs"
              className="inline-flex items-center text-primary hover:text-primary-600 font-semibold text-lg group transition-colors"
            >
              查看完整功能列表
              <span className="ml-2 transform group-hover:translate-x-1 transition-transform">→</span>
            </a>
          </div>
        </div>
      </div>
    </section>
  );
};

export default FeaturesSection;
