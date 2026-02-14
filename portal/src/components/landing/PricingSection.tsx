import { Button } from 'antd';
import { CheckOutlined, StarOutlined, CrownOutlined, RocketOutlined } from '@ant-design/icons';

interface PricingPlan {
  name: string;
  icon: React.ReactNode;
  price: string;
  period: string;
  description: string;
  features: string[];
  highlighted?: boolean;
  buttonText: string;
  buttonType: 'default' | 'primary';
}

const pricingPlans: PricingPlan[] = [
  {
    name: '免费版',
    icon: <StarOutlined className="text-3xl" />,
    price: '¥0',
    period: '永久免费',
    description: '适合个人开发者和小型项目',
    features: [
      '10,000 tokens 初始额度',
      '每日签到获得 1,000 tokens',
      '支持所有主流模型',
      '基础限流保护',
      '社区技术支持',
      'API文档访问',
    ],
    buttonText: '免费开始',
    buttonType: 'default',
  },
  {
    name: '专业版',
    icon: <RocketOutlined className="text-3xl" />,
    price: '¥99',
    period: '/月',
    description: '适合中小型团队和商业项目',
    features: [
      '500,000 tokens/月',
      '优先级负载均衡',
      '更高的请求限流',
      '详细使用统计',
      '邮件技术支持',
      '99.9% SLA保障',
      '自定义API配置',
      '团队协作功能',
    ],
    highlighted: true,
    buttonText: '立即订阅',
    buttonType: 'primary',
  },
  {
    name: '企业版',
    icon: <CrownOutlined className="text-3xl" />,
    price: '定制',
    period: '联系我们',
    description: '适合大型企业和高并发场景',
    features: [
      '无限 tokens 额度',
      '专属负载均衡策略',
      '私有化部署支持',
      '定制化开发',
      '7x24 专属技术支持',
      '99.99% SLA保障',
      '数据安全审计',
      '专属客户经理',
    ],
    buttonText: '联系销售',
    buttonType: 'default',
  },
];

const PricingSection = () => {
  return (
    <section id="pricing" className="py-24 bg-slate-50 dark:bg-black relative overflow-hidden">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 relative z-10">
        {/* Section Header */}
        <div className="text-center mb-16">
          <div className="inline-block px-4 py-2 rounded-full bg-primary/10 text-primary text-sm font-medium mb-4">
            定价方案
          </div>
          <h2 className="text-4xl font-bold mb-6 text-slate-900 dark:text-white">
            选择适合您的方案
          </h2>
          <p className="text-xl text-slate-600 dark:text-slate-400 max-w-2xl mx-auto">
            灵活的定价，透明的收费，无隐藏费用
          </p>
        </div>

        {/* Pricing Cards */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-8 max-w-6xl mx-auto">
          {pricingPlans.map((plan, index) => (
            <div
              key={index}
              className={`relative glass-card p-8 rounded-2xl transition-all duration-300 ${plan.highlighted
                ? 'border-2 border-primary shadow-2xl shadow-primary/20 transform md:scale-105 z-10'
                : 'border border-slate-200 dark:border-white/5 hover:border-primary/50 hover:shadow-xl'
                }`}
            >
              {plan.highlighted && (
                <div className="absolute -top-4 left-1/2 transform -translate-x-1/2">
                  <div className="bg-gradient-to-r from-primary to-purple-600 text-white px-4 py-1 rounded-full text-sm font-semibold shadow-lg">
                    最受欢迎
                  </div>
                </div>
              )}

              <div className="text-center mb-8">
                <div className={`inline-flex items-center justify-center w-16 h-16 rounded-2xl mb-6 ${plan.highlighted ? 'bg-primary/10 text-primary' : 'bg-slate-100 dark:bg-white/5 text-slate-600 dark:text-slate-400'
                  }`}>
                  {plan.icon}
                </div>
                <h3 className="text-2xl font-bold mb-2 text-slate-900 dark:text-white">
                  {plan.name}
                </h3>
                <p className="text-slate-500 dark:text-slate-400 mb-6 h-10">
                  {plan.description}
                </p>
                <div className="flex items-baseline justify-center">
                  <span className="text-5xl font-bold text-slate-900 dark:text-white tracking-tight">{plan.price}</span>
                  {plan.period !== '联系我们' && (
                    <span className="text-slate-500 dark:text-slate-400 ml-2">{plan.period}</span>
                  )}
                </div>
                {plan.period === '联系我们' && (
                  <div className="text-slate-500 dark:text-slate-400 ml-2 opacity-0 h-0">{plan.period}</div>
                )}
              </div>

              <div className="mb-8 space-y-4">
                {plan.features.map((feature, featureIndex) => (
                  <div key={featureIndex} className="flex items-start">
                    <CheckOutlined className={`${plan.highlighted ? 'text-primary' : 'text-green-500'
                      } mt-1 mr-3 flex-shrink-0`} />
                    <span className="text-slate-700 dark:text-slate-300">{feature}</span>
                  </div>
                ))}
              </div>

              <Button
                type={plan.buttonType}
                size="large"
                block
                className={`!h-12 !font-semibold !rounded-xl ${plan.highlighted
                  ? 'bg-primary hover:bg-primary-600 border-none shadow-lg shadow-primary/20 text-white'
                  : 'bg-white dark:bg-white/5 border-slate-200 dark:border-white/10 text-slate-900 dark:text-white hover:border-primary hover:text-primary'
                  }`}
              >
                {plan.buttonText}
              </Button>
            </div>
          ))}
        </div>

        {/* Additional Info */}
        <div className="mt-20 text-center">
          <div className="glass-card p-10 rounded-3xl bg-gradient-to-r from-blue-500/5 to-indigo-500/5 border border-blue-200 dark:border-blue-500/10 max-w-4xl mx-auto">
            <h3 className="text-2xl font-bold mb-4 text-slate-900 dark:text-white">
              需要更多信息？
            </h3>
            <p className="text-lg text-slate-600 dark:text-slate-400 mb-8">
              我们的销售团队随时为您提供帮助，解答您的任何问题
            </p>
            <div className="flex flex-col sm:flex-row gap-4 justify-center">
              <Button size="large" className="!h-12 !px-8 !rounded-xl" href="/contact">
                联系销售
              </Button>
              <Button type="primary" size="large" className="!h-12 !px-8 !rounded-xl" href="/docs/pricing">
                查看详细定价
              </Button>
            </div>
          </div>
        </div>

        {/* FAQ Teaser */}
        <div className="mt-16 text-center">
          <p className="text-slate-500 dark:text-slate-400">
            常见问题？查看我们的{' '}
            <a href="/faq" className="text-primary hover:text-primary-600 font-semibold transition-colors">
              FAQ页面
            </a>
          </p>
        </div>
      </div>
    </section>
  );
};

export default PricingSection;
