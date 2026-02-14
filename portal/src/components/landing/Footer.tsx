import {
  GithubOutlined,
  TwitterOutlined,
  WechatOutlined,
  MailOutlined,
} from '@ant-design/icons';

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
    <footer className="bg-slate-950 text-slate-400 border-t border-slate-800">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-16">
        {/* Main Footer Content */}
        <div className="grid grid-cols-2 md:grid-cols-5 gap-8 mb-16">
          {/* Brand Column */}
          <div className="col-span-2 md:col-span-1">
            <div className="flex items-center gap-2 mb-6">
              <div className="relative">
                <div className="absolute inset-0 bg-primary/20 blur-lg rounded-full"></div>
                <img src="/logo.svg" alt="Prism API" className="w-8 h-8 relative z-10" />
              </div>
              <span className="text-xl font-bold text-white tracking-tight">
                Prism API
              </span>
            </div>
            <p className="text-slate-500 mb-6 text-sm leading-relaxed">
              Universal AI API Gateway
              <br />
              <span className="opacity-75">Designed by Adnaan</span>
            </p>
            <div className="flex gap-4">
              {[
                { icon: <GithubOutlined />, href: "https://github.com" },
                { icon: <TwitterOutlined />, href: "https://twitter.com" },
                { icon: <WechatOutlined />, href: "#wechat" },
                { icon: <MailOutlined />, href: "mailto:support@example.com" }
              ].map((social, index) => (
                <a
                  key={index}
                  href={social.href}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="w-10 h-10 rounded-full bg-white/5 flex items-center justify-center text-slate-400 hover:bg-white/10 hover:text-white transition-all duration-300"
                >
                  {social.icon}
                </a>
              ))}
            </div>
          </div>

          {/* Product Links */}
          <div>
            <h4 className="text-white font-semibold mb-6">产品</h4>
            <ul className="space-y-3 text-sm">
              {footerLinks.product.map((link, index) => (
                <li key={index}>
                  <a href={link.href} className="text-slate-500 hover:text-primary transition-colors">
                    {link.label}
                  </a>
                </li>
              ))}
            </ul>
          </div>

          {/* Developer Links */}
          <div>
            <h4 className="text-white font-semibold mb-6">开发者</h4>
            <ul className="space-y-3 text-sm">
              {footerLinks.developers.map((link, index) => (
                <li key={index}>
                  <a href={link.href} className="text-slate-500 hover:text-primary transition-colors">
                    {link.label}
                  </a>
                </li>
              ))}
            </ul>
          </div>

          {/* Company Links */}
          <div>
            <h4 className="text-white font-semibold mb-6">公司</h4>
            <ul className="space-y-3 text-sm">
              {footerLinks.company.map((link, index) => (
                <li key={index}>
                  <a href={link.href} className="text-slate-500 hover:text-primary transition-colors">
                    {link.label}
                  </a>
                </li>
              ))}
            </ul>
          </div>

          {/* Legal Links */}
          <div>
            <h4 className="text-white font-semibold mb-6">法律</h4>
            <ul className="space-y-3 text-sm">
              {footerLinks.legal.map((link, index) => (
                <li key={index}>
                  <a href={link.href} className="text-slate-500 hover:text-primary transition-colors">
                    {link.label}
                  </a>
                </li>
              ))}
            </ul>
          </div>
        </div>

        {/* Newsletter Section */}
        <div className="border-t border-slate-800 pt-12 pb-12">
          <div className="max-w-md">
            <h4 className="text-white font-semibold mb-2">订阅我们的新闻</h4>
            <p className="text-slate-500 text-sm mb-4">
              获取最新的产品更新和技术文章，随时取消订阅。
            </p>
            <div className="flex gap-2">
              <input
                type="email"
                placeholder="输入您的邮箱"
                className="flex-1 px-4 py-2.5 bg-white/5 border border-white/10 rounded-xl text-white placeholder-slate-600 focus:outline-none focus:border-primary/50 focus:ring-1 focus:ring-primary/50 transition-all text-sm"
              />
              <button className="px-6 py-2.5 bg-primary hover:bg-primary-600 text-white rounded-xl font-semibold text-sm transition-colors shadow-lg shadow-primary/20">
                订阅
              </button>
            </div>
          </div>
        </div>

        {/* Bottom Bar */}
        <div className="border-t border-slate-800 pt-8 flex flex-col md:flex-row justify-between items-center gap-4 text-sm">
          <p className="text-slate-600 mb-0">
            © {currentYear} Prism API by Adnaan. All rights reserved.
          </p>
          <div className="flex gap-6">
            <a href="/status" className="text-slate-500 hover:text-white transition-colors flex items-center gap-2">
              <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"></span>
              系统状态
            </a>
            <a href="/security" className="text-slate-500 hover:text-white transition-colors">
              安全中心
            </a>
            <a href="/sitemap" className="text-slate-500 hover:text-white transition-colors">
              网站地图
            </a>
          </div>
        </div>
      </div>
    </footer>
  );
};

export default Footer;
