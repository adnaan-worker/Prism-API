import { Button } from 'antd';
import { ArrowRightOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import HeroSection from '../components/landing/HeroSection';
import FeaturesSection from '../components/landing/FeaturesSection';
import StatsSection from '../components/landing/StatsSection';
import PricingSection from '../components/landing/PricingSection';
import Footer from '../components/landing/Footer';

const LandingPage = () => {
  const navigate = useNavigate();
  const token = localStorage.getItem('token');
  const isLoggedIn = !!token;

  return (
    <div className="min-h-screen bg-slate-50 dark:bg-black text-slate-900 dark:text-white transition-colors duration-300">
      {/* Navigation Bar */}
      <nav className="fixed top-0 left-0 right-0 z-50 glass border-b border-slate-200 dark:border-white/10">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-20">
            <div className="flex items-center gap-3">
              <div className="relative">
                <div className="absolute inset-0 bg-primary/20 blur-lg rounded-full animate-active-pulse"></div>
                <img src="/logo.svg" alt="Prism API" className="w-8 h-8 relative z-10" />
              </div>
              <div className="flex flex-col">
                <span className="text-xl font-bold tracking-tight text-slate-900 dark:text-white leading-none">
                  Prism <span className="text-primary">API</span>
                </span>
                <span className="text-xs text-slate-500 dark:text-slate-400">Universal Gateway</span>
              </div>
            </div>
            <div className="hidden md:flex items-center space-x-8">
              {['Features', 'Pricing', 'Docs'].map((item) => (
                <a
                  key={item}
                  href={`#${item.toLowerCase()} `}
                  className="text-sm font-medium text-slate-600 dark:text-slate-300 hover:text-primary dark:hover:text-primary transition-colors"
                >
                  {item}
                </a>
              ))}
              {isLoggedIn ? (
                <Button
                  type="primary"
                  onClick={() => navigate('/dashboard')}
                  icon={<ArrowRightOutlined />}
                  className="bg-primary hover:bg-primary-600 border-none shadow-lg shadow-primary/20"
                >
                  Console
                </Button>
              ) : (
                <div className="flex items-center gap-4">
                  <Button
                    type="text"
                    href="/login"
                    className="text-slate-600 dark:text-slate-300 hover:text-slate-900 dark:hover:text-white"
                  >
                    Login
                  </Button>
                  <Button
                    type="primary"
                    href="/register"
                    icon={<ArrowRightOutlined />}
                    className="bg-primary hover:bg-primary-600 border-none shadow-lg shadow-primary/20"
                  >
                    Get Started
                  </Button>
                </div>
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
