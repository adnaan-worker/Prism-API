import { useState, useEffect } from 'react';
import { Dropdown, Avatar, theme as antTheme, Button, Drawer } from 'antd';
import {
  DashboardOutlined,
  KeyOutlined,
  AppstoreOutlined,
  FileTextOutlined,
  UserOutlined,
  LogoutOutlined,
  MenuOutlined,
  BellOutlined,
  SearchOutlined,
  GlobalOutlined,
  SunOutlined,
  MoonOutlined
} from '@ant-design/icons';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { authService } from '../services/authService';

interface DashboardLayoutProps {
  isDarkMode: boolean;
  onThemeToggle: () => void;
}

const DashboardLayout = ({ isDarkMode, onThemeToggle }: DashboardLayoutProps) => {
  const [scrolled, setScrolled] = useState(false);
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();
  const { token } = antTheme.useToken();
  const { t, i18n } = useTranslation();

  const user = authService.getSavedUser();

  // Handle scroll effect for header
  useEffect(() => {
    const handleScroll = () => {
      setScrolled(window.scrollY > 20);
    };
    window.addEventListener('scroll', handleScroll);
    return () => window.removeEventListener('scroll', handleScroll);
  }, []);

  const handleLogout = () => {
    authService.logout();
    navigate('/login');
  };

  const changeLanguage = (lng: string) => {
    i18n.changeLanguage(lng);
  };

  const menuItems = [
    { key: '/dashboard', icon: <DashboardOutlined />, label: t('menu.dashboard') },
    { key: '/dashboard/api-keys', icon: <KeyOutlined />, label: t('menu.apiKeys') },
    { key: '/dashboard/models', icon: <AppstoreOutlined />, label: t('menu.models') },
    { key: '/dashboard/docs', icon: <FileTextOutlined />, label: t('menu.docs') },
    { key: '/dashboard/profile', icon: <UserOutlined />, label: t('menu.profile') },
  ];

  const userMenuItems = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: t('menu.profile'),
      onClick: () => navigate('/dashboard/profile'),
    },
    { type: 'divider' as const },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: t('menu.logout'),
      onClick: handleLogout,
      danger: true,
    },
  ];

  const languageItems = [
    {
      key: 'en',
      label: 'English',
      onClick: () => changeLanguage('en'),
    },
    {
      key: 'zh',
      label: '中文',
      onClick: () => changeLanguage('zh'),
    },
  ];

  const SidebarContent = () => (
    <div className="flex flex-col h-full">
      {/* Logo Area */}
      <div className="h-20 flex items-center px-8 border-b border-border bg-white dark:bg-transparent">
        <div className="flex items-center gap-3">
          <div className="relative">
            <div className="absolute inset-0 bg-primary/20 blur-lg rounded-full animate-active-pulse"></div>
            <img src="/logo-dark.svg" alt="Prism" className="w-8 h-8 relative z-10" />
          </div>
          <span className="text-xl font-bold tracking-tight text-text-primary">
            Prism <span className="text-primary">API</span>
          </span>
        </div>
      </div>

      {/* Navigation */}
      <div className="flex-1 py-6 px-4 space-y-1 overflow-y-auto custom-scrollbar">
        <div className="mb-2 px-4 text-xs font-semibold text-text-tertiary uppercase tracking-wider">
          {t('menu.menu')}
        </div>
        {menuItems.map((item) => {
          const isActive = location.pathname === item.key;
          return (
            <div
              key={item.key}
              onClick={() => {
                navigate(item.key);
                setMobileMenuOpen(false);
              }}
              className={`
                group flex items-center px-4 py-3 rounded-xl cursor-pointer transition-all duration-300
                ${isActive
                  ? 'bg-primary/10 text-primary shadow-sm font-medium'
                  : 'text-text-secondary hover:text-text-primary hover:bg-black/5 dark:hover:bg-white/5'
                }
              `}
            >
              <span
                className={`text-lg mr-3 transition-transform duration-300 ${isActive ? 'scale-110' : 'group-hover:scale-110'
                  }`}
              >
                {item.icon}
              </span>
              <span className="font-medium">{item.label}</span>
              {isActive && (
                <div className="ml-auto w-1.5 h-1.5 rounded-full bg-primary shadow-glow"></div>
              )}
            </div>
          );
        })}
      </div>

      {/* User Status Card */}
      <div className="p-4 border-t border-border bg-white dark:bg-transparent">
        <div className="bg-slate-50 dark:bg-white/5 rounded-xl p-4 border border-border shadow-sm group hover:shadow-md transition-all">
          <div className="flex justify-between items-center mb-2">
            <span className="text-xs text-text-tertiary">{t('dashboard.quota.remaining')}</span>
            <span className="text-xs text-primary font-mono bg-primary/10 px-2 py-0.5 rounded">
              PRO
            </span>
          </div>
          <div className="text-xl font-bold text-text-primary mb-1">
            ${user?.quota ? (user.quota - (user.used_quota || 0)).toLocaleString() : '0.00'}
          </div>
          <div className="w-full bg-background-subtle h-1.5 rounded-full overflow-hidden">
            <div
              className="bg-primary h-full rounded-full"
              style={{ width: '45%' }} // This would be dynamic in real app
            ></div>
          </div>
        </div>
      </div>
    </div>
  );

  return (
    <div className="min-h-screen bg-slate-50 dark:bg-black flex text-text-primary font-sans selection:bg-primary/30 transition-colors duration-300">
      {/* Desktop Sidebar */}
      <aside className="hidden md:block w-72 fixed h-screen z-50 bg-white dark:bg-black border-r border-border shadow-xl dark:shadow-none">
        <SidebarContent />
      </aside>

      {/* Mobile Drawer */}
      <Drawer
        placement="left"
        onClose={() => setMobileMenuOpen(false)}
        open={mobileMenuOpen}
        styles={{ body: { padding: 0 } }}
        width={280}
        closeIcon={null}
        className="dark:bg-black bg-white"
      >
        <SidebarContent />
      </Drawer>

      {/* Main Content */}
      <div className="flex-1 md:ml-72 flex flex-col min-h-screen relative transition-all duration-300">

        {/* Top Header */}
        <header
          className={`
            sticky top-0 z-40 h-20 px-8 flex items-center justify-between transition-all duration-300
            ${scrolled ? 'glass border-b border-border shadow-sm' : 'bg-transparent'}
          `}
        >
          <div className="flex items-center gap-4">
            <Button
              type="text"
              icon={<MenuOutlined className="text-text-primary text-lg" />}
              className="md:hidden"
              onClick={() => setMobileMenuOpen(true)}
            />
            {/* Breadcrumb-like title or Page Title */}
            <h1 className="text-xl font-bold text-text-primary capitalize">
              {menuItems.find(i => i.key === location.pathname)?.label || t('menu.dashboard')}
            </h1>
          </div>

          <div className="flex items-center gap-4">
            <div className="hidden md:flex items-center bg-white/50 dark:bg-white/5 rounded-full px-4 py-2 border border-border focus-within:border-primary/50 transition-colors">
              <SearchOutlined className="text-text-tertiary mr-2" />
              <input
                type="text"
                placeholder={t('common.search')}
                className="bg-transparent border-none outline-none text-sm text-text-primary placeholder-text-tertiary w-32 focus:w-48 transition-all"
              />
            </div>

            {/* Language Switcher */}
            <Dropdown menu={{ items: languageItems }} placement="bottomRight" trigger={['click']}>
              <Button type="text" shape="circle" icon={<GlobalOutlined className="text-text-secondary hover:text-text-primary" />} />
            </Dropdown>

            {/* Theme Toggle */}
            <Button
              type="text"
              shape="circle"
              onClick={onThemeToggle}
              icon={isDarkMode ? <SunOutlined className="text-text-secondary hover:text-text-primary" /> : <MoonOutlined className="text-text-secondary hover:text-text-primary" />}
            />

            <Button type="text" shape="circle" icon={<BellOutlined className="text-text-secondary hover:text-text-primary" />} />

            <Dropdown menu={{ items: userMenuItems }} placement="bottomRight" trigger={['click']}>
              <div className="flex items-center gap-3 cursor-pointer group">
                <div className="text-right hidden sm:block">
                  <div className="text-sm font-medium text-text-primary group-hover:text-primary transition-colors">{user?.username || 'User'}</div>
                  <div className="text-xs text-text-tertiary">{t(`profile.role.${user?.is_admin ? 'admin' : 'user'}`) || 'Developer'}</div>
                </div>
                <Avatar
                  size={40}
                  style={{ background: token.colorPrimary }}
                  icon={<UserOutlined />}
                  className="ring-2 ring-white dark:ring-black border border-black/10 dark:border-white/10 group-hover:ring-primary/50 transition-all"
                />
              </div>
            </Dropdown>
          </div>
        </header>

        {/* Page Content */}
        <main className="flex-1 p-8 relative z-10 animate-fade-in">
          <Outlet context={{ isDarkMode }} />
        </main>

      </div>
    </div>
  );
};

export default DashboardLayout;
