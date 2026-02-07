import { Button, Typography } from 'antd';
import { ArrowRightOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import HeroSection from '../components/landing/HeroSection';
import FeaturesSection from '../components/landing/FeaturesSection';
import StatsSection from '../components/landing/StatsSection';
import PricingSection from '../components/landing/PricingSection';
import Footer from '../components/landing/Footer';

const { Title } = Typography;

const LandingPage = () => {
  const navigate = useNavigate();
  const token = localStorage.getItem('token');
  const isLoggedIn = !!token;

  return (
    <div className="min-h-screen bg-white">
      {/* Navigation Bar */}
      <nav className="fixed top-0 left-0 right-0 z-50 bg-white/80 backdrop-blur-md border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center gap-3">
              <img src="/logo.svg" alt="Prism API" style={{ width: 32, height: 32 }} />
              <Title level={4} className="!mb-0 !text-primary-600">
                Prism API
              </Title>
              <span className="ml-2 text-xs text-gray-500">by Adnaan</span>
            </div>
            <div className="hidden md:flex items-center space-x-8">
              <a href="#features" className="text-gray-700 hover:text-primary-600 transition-colors">
                产品
              </a>
              <a href="#pricing" className="text-gray-700 hover:text-primary-600 transition-colors">
                定价
              </a>
              <a href="#docs" className="text-gray-700 hover:text-primary-600 transition-colors">
                文档
              </a>
              {isLoggedIn ? (
                <Button 
                  type="primary" 
                  onClick={() => navigate('/dashboard')}
                  icon={<ArrowRightOutlined />}
                >
                  进入控制台
                </Button>
              ) : (
                <>
                  <Button type="default" href="/login">
                    登录
                  </Button>
                  <Button type="primary" href="/register" icon={<ArrowRightOutlined />}>
                    开始使用
                  </Button>
                </>
              )}
            </div>
          </div>
        </div>
      </nav>

      {/* Main Content */}
      <main className="pt-16">
        <HeroSection />
        <FeaturesSection />
        <StatsSection />
        <PricingSection />
      </main>

      {/* Footer */}
      <Footer />
    </div>
  );
};

export default LandingPage;
