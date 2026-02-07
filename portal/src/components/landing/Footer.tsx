import { Typography } from 'antd';
import {
  GithubOutlined,
  TwitterOutlined,
  WechatOutlined,
  MailOutlined,
} from '@ant-design/icons';

const { Title, Paragraph } = Typography;

const Footer = () => {
  const currentYear = new Date().getFullYear();

  const footerLinks = {
    product: [
      { label: '功能特性', href: '#features' },
      { label: '定价方案', href: '#pricing' },
      { label: '更新日志', href: '/changelog' },
      { label: '路线图', href: '/roadmap' },
    ],
    developers: [
      { label: 'API文档', href: '/docs' },
      { label: '快速开始', href: '/docs/quickstart' },
      { label: 'SDK下载', href: '/docs/sdk' },
      { label: '示例代码', href: '/docs/examples' },
    ],
    company: [
      { label: '关于我们', href: '/about' },
      { label: '博客', href: '/blog' },
      { label: '联系我们', href: '/contact' },
      { label: '加入我们', href: '/careers' },
    ],
    legal: [
      { label: '服务条款', href: '/terms' },
      { label: '隐私政策', href: '/privacy' },
      { label: 'Cookie政策', href: '/cookies' },
      { label: '许可协议', href: '/license' },
    ],
  };

  return (
    <footer className="bg-gray-900 text-gray-300">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
        {/* Main Footer Content */}
        <div className="grid grid-cols-2 md:grid-cols-5 gap-8 mb-12">
          {/* Brand Column */}
          <div className="col-span-2 md:col-span-1">
            <Title level={4} className="!text-white !mb-4">
              🌈 Prism API
            </Title>
            <Paragraph className="text-gray-400 !mb-4">
              Universal AI API Gateway
              <br />
              <span className="text-sm">by Adnaan</span>
            </Paragraph>
            <div className="flex gap-4">
              <a
                href="https://github.com"
                target="_blank"
                rel="noopener noreferrer"
                className="text-gray-400 hover:text-white transition-colors"
              >
                <GithubOutlined className="text-2xl" />
              </a>
              <a
                href="https://twitter.com"
                target="_blank"
                rel="noopener noreferrer"
                className="text-gray-400 hover:text-white transition-colors"
              >
                <TwitterOutlined className="text-2xl" />
              </a>
              <a
                href="#wechat"
                className="text-gray-400 hover:text-white transition-colors"
              >
                <WechatOutlined className="text-2xl" />
              </a>
              <a
                href="mailto:support@example.com"
                className="text-gray-400 hover:text-white transition-colors"
              >
                <MailOutlined className="text-2xl" />
              </a>
            </div>
          </div>

          {/* Product Links */}
          <div>
            <Title level={5} className="!text-white !mb-4">
              产品
            </Title>
            <ul className="space-y-2">
              {footerLinks.product.map((link, index) => (
                <li key={index}>
                  <a
                    href={link.href}
                    className="text-gray-400 hover:text-white transition-colors"
                  >
                    {link.label}
                  </a>
                </li>
              ))}
            </ul>
          </div>

          {/* Developer Links */}
          <div>
            <Title level={5} className="!text-white !mb-4">
              开发者
            </Title>
            <ul className="space-y-2">
              {footerLinks.developers.map((link, index) => (
                <li key={index}>
                  <a
                    href={link.href}
                    className="text-gray-400 hover:text-white transition-colors"
                  >
                    {link.label}
                  </a>
                </li>
              ))}
            </ul>
          </div>

          {/* Company Links */}
          <div>
            <Title level={5} className="!text-white !mb-4">
              公司
            </Title>
            <ul className="space-y-2">
              {footerLinks.company.map((link, index) => (
                <li key={index}>
                  <a
                    href={link.href}
                    className="text-gray-400 hover:text-white transition-colors"
                  >
                    {link.label}
                  </a>
                </li>
              ))}
            </ul>
          </div>

          {/* Legal Links */}
          <div>
            <Title level={5} className="!text-white !mb-4">
              法律
            </Title>
            <ul className="space-y-2">
              {footerLinks.legal.map((link, index) => (
                <li key={index}>
                  <a
                    href={link.href}
                    className="text-gray-400 hover:text-white transition-colors"
                  >
                    {link.label}
                  </a>
                </li>
              ))}
            </ul>
          </div>
        </div>

        {/* Newsletter Section */}
        <div className="border-t border-gray-800 pt-8 mb-8">
          <div className="max-w-md">
            <Title level={5} className="!text-white !mb-2">
              订阅我们的新闻
            </Title>
            <Paragraph className="text-gray-400 !mb-4">
              获取最新的产品更新和技术文章
            </Paragraph>
            <div className="flex gap-2">
              <input
                type="email"
                placeholder="输入您的邮箱"
                className="flex-1 px-4 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-primary-500"
              />
              <button className="px-6 py-2 bg-primary-600 hover:bg-primary-700 text-white rounded-lg font-semibold transition-colors">
                订阅
              </button>
            </div>
          </div>
        </div>

        {/* Bottom Bar */}
        <div className="border-t border-gray-800 pt-8">
          <div className="flex flex-col md:flex-row justify-between items-center gap-4">
            <Paragraph className="text-gray-400 !mb-0">
              © {currentYear} Prism API by Adnaan. All rights reserved.
            </Paragraph>
            <div className="flex gap-6 text-sm">
              <a href="/status" className="text-gray-400 hover:text-white transition-colors">
                系统状态
              </a>
              <a href="/security" className="text-gray-400 hover:text-white transition-colors">
                安全中心
              </a>
              <a href="/sitemap" className="text-gray-400 hover:text-white transition-colors">
                网站地图
              </a>
            </div>
          </div>
        </div>
      </div>
    </footer>
  );
};

export default Footer;
