import { Button, Card, Typography } from 'antd';
import { CheckOutlined, StarOutlined, CrownOutlined, RocketOutlined } from '@ant-design/icons';

const { Title, Paragraph } = Typography;

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
    <section id="pricing" className="py-20 bg-gray-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        {/* Section Header */}
        <div className="text-center mb-16">
          <div className="inline-block px-4 py-2 rounded-full bg-primary-100 text-primary-700 text-sm font-medium mb-4">
            定价方案
          </div>
          <Title level={2} className="!text-4xl !font-bold !mb-4">
            选择适合您的方案
          </Title>
          <Paragraph className="text-xl text-gray-600 max-w-2xl mx-auto">
            灵活的定价，透明的收费，无隐藏费用
          </Paragraph>
        </div>

        {/* Pricing Cards */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-8 max-w-6xl mx-auto">
          {pricingPlans.map((plan, index) => (
            <Card
              key={index}
              className={`relative ${
                plan.highlighted
                  ? 'border-2 border-primary-500 shadow-2xl transform md:scale-105'
                  : 'border border-gray-200'
              } hover:shadow-xl transition-all duration-300`}
              bodyStyle={{ padding: '32px' }}
            >
              {plan.highlighted && (
                <div className="absolute -top-4 left-1/2 transform -translate-x-1/2">
                  <div className="bg-gradient-to-r from-primary-500 to-purple-500 text-white px-4 py-1 rounded-full text-sm font-semibold">
                    最受欢迎
                  </div>
                </div>
              )}

              <div className="text-center mb-6">
                <div className={`inline-flex items-center justify-center w-16 h-16 rounded-full mb-4 ${
                  plan.highlighted ? 'bg-primary-100 text-primary-600' : 'bg-gray-100 text-gray-600'
                }`}>
                  {plan.icon}
                </div>
                <Title level={3} className="!mb-2">
                  {plan.name}
                </Title>
                <Paragraph className="text-gray-600 !mb-4">
                  {plan.description}
                </Paragraph>
                <div className="mb-2">
                  <span className="text-5xl font-bold text-gray-900">{plan.price}</span>
                  {plan.period !== '联系我们' && (
                    <span className="text-gray-600 ml-2">{plan.period}</span>
                  )}
                </div>
                {plan.period === '联系我们' && (
                  <div className="text-gray-600">{plan.period}</div>
                )}
              </div>

              <div className="mb-8">
                <ul className="space-y-3">
                  {plan.features.map((feature, featureIndex) => (
                    <li key={featureIndex} className="flex items-start">
                      <CheckOutlined className={`${
                        plan.highlighted ? 'text-primary-600' : 'text-green-600'
                      } mt-1 mr-3 flex-shrink-0`} />
                      <span className="text-gray-700">{feature}</span>
                    </li>
                  ))}
                </ul>
              </div>

              <Button
                type={plan.buttonType}
                size="large"
                block
                className={`!h-12 !font-semibold ${
                  plan.highlighted ? 'shadow-lg hover:shadow-xl' : ''
                }`}
              >
                {plan.buttonText}
              </Button>
            </Card>
          ))}
        </div>

        {/* Additional Info */}
        <div className="mt-16 text-center">
          <Card className="bg-gradient-to-r from-blue-50 to-indigo-50 border-0 max-w-4xl mx-auto">
            <div className="py-6">
              <Title level={4} className="!mb-3">
                需要更多信息？
              </Title>
              <Paragraph className="text-gray-600 !mb-4">
                我们的销售团队随时为您提供帮助，解答您的任何问题
              </Paragraph>
              <div className="flex flex-col sm:flex-row gap-4 justify-center">
                <Button size="large" href="/contact">
                  联系销售
                </Button>
                <Button size="large" href="/docs/pricing">
                  查看详细定价
                </Button>
              </div>
            </div>
          </Card>
        </div>

        {/* FAQ Teaser */}
        <div className="mt-12 text-center">
          <Paragraph className="text-gray-600">
            常见问题？查看我们的{' '}
            <a href="/faq" className="text-primary-600 hover:text-primary-700 font-semibold">
              FAQ页面
            </a>
          </Paragraph>
        </div>
      </div>
    </section>
  );
};

export default PricingSection;
