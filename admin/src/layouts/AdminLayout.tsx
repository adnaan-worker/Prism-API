import React, { useState, useEffect } from 'react';
import { Breadcrumb, Avatar, Dropdown, theme as antTheme, Button, Drawer } from 'antd';
import {
  DashboardOutlined,
  UserOutlined,
  ApiOutlined,
  BarChartOutlined,
  FileTextOutlined,
  SettingOutlined,
  LogoutOutlined,
  DollarOutlined,
  BellOutlined,
  SearchOutlined,
  MenuOutlined,
  SunOutlined,
  MoonOutlined
} from '@ant-design/icons';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import type { MenuProps } from 'antd';

// Menu Configuration
const menuItems = [
  { key: '/dashboard', icon: <DashboardOutlined />, label: '统计概览', group: 'main' },
  { key: '/users', icon: <UserOutlined />, label: '用户管理', group: 'business' },
  { key: '/api-configs', icon: <ApiOutlined />, label: 'API 配置', group: 'business' },
  { key: '/load-balancer', icon: <BarChartOutlined />, label: '负载均衡', group: 'business' },
  { key: '/pricing', icon: <DollarOutlined />, label: '定价管理', group: 'business' },
  { key: '/logs', icon: <FileTextOutlined />, label: '请求日志', group: 'system' },
  { key: '/settings', icon: <SettingOutlined />, label: '系统设置', group: 'system' },
];

const breadcrumbMap: Record<string, string> = {
  dashboard: '统计概览',
  users: '用户管理',
  'api-configs': 'API 配置',
  'load-balancer': '负载均衡',
  pricing: '定价管理',
  logs: '请求日志',
  settings: '系统设置',
};

const AdminLayout: React.FC = () => {
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const [scrolled, setScrolled] = useState(false);
  // Re-render trigger for dark mode
  const [isDarkMode, setIsDarkMode] = useState((window as any).__isDarkMode);

  const navigate = useNavigate();
  const location = useLocation();
  const { token } = antTheme.useToken();

  useEffect(() => {
    const handleScroll = () => {
      setScrolled(window.scrollY > 20);
    };
    window.addEventListener('scroll', handleScroll);
    return () => window.removeEventListener('scroll', handleScroll);
  }, []);

  const handleUserMenuClick = ({ key }: { key: string }) => {
    if (key === 'logout') {
      localStorage.removeItem('admin_token');
      navigate('/login');
    } else if (key === 'settings') {
      navigate('/settings');
    }
  };

  const handleThemeToggle = () => {
    const toggle = (window as any).__toggleTheme;
    if (toggle) {
      toggle();
      setIsDarkMode((window as any).__isDarkMode);
    }
  };

  const userMenuItems: MenuProps['items'] = [
    { key: 'settings', icon: <SettingOutlined />, label: '系统设置' },
    { type: 'divider' },
    { key: 'logout', icon: <LogoutOutlined />, label: '退出登录', danger: true },
  ];

  const getBreadcrumbItems = () => {
    const segments = location.pathname.split('/').filter(Boolean);
    const items = [{ title: '首页', href: '/dashboard', onClick: (e: React.MouseEvent) => { e.preventDefault(); navigate('/dashboard'); } }];
    segments.forEach((seg) => {
      const name = breadcrumbMap[seg];
      if (name) items.push({ title: name, href: '', onClick: undefined as any });
    });
    return items;
  };

  const SidebarContent = () => (
    <div className="flex flex-col h-full bg-background dark:bg-black">
      {/* Logo */}
      <div className="h-20 flex items-center px-6 border-b border-border">
        <div className="relative flex items-center gap-3">
          <div className="absolute inset-0 bg-primary/20 blur-lg rounded-full animate-active-pulse"></div>
          {/* Automatically adjust logo color based on light/dark if possible, else standard */}
          <img src="/logo-dark.svg" alt="Prism" className="w-8 h-8 relative z-10 hidden dark:block" />
          <img src="/logo.svg" alt="Prism" className="w-8 h-8 relative z-10 block dark:hidden" />
          <span className="text-xl font-bold tracking-tight text-text-primary animate-fade-in relative z-10">
            Prism <span className="text-primary">Admin</span>
          </span>
        </div>
      </div>

      {/* Menu */}
      <div className="flex-1 overflow-y-auto py-6 px-4 space-y-1 custom-scrollbar">
        {['main', 'business', 'system'].map((group) => {
          const groupItems = menuItems.filter(item => item.group === group);
          if (groupItems.length === 0) return null;

          return (
            <div key={group} className="mb-6">
              {group !== 'main' && (
                <div className="mb-2 px-4 text-xs font-semibold text-text-tertiary uppercase tracking-wider">
                  {group === 'business' ? '业务管理' : '系统'}
                </div>
              )}
              {groupItems.map((item) => {
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
                    <span className={`text-lg mr-3 transition-transform duration-300 ${isActive ? 'scale-110' : 'group-hover:scale-110'}`}>
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
          );
        })}
      </div>
    </div>
  );

  return (
    <div className="min-h-screen bg-background flex text-text-primary font-sans selection:bg-primary/30 transition-colors duration-300">

      {/* Desktop Sidebar */}
      <aside className="hidden lg:block fixed w-64 h-screen z-50 border-r border-border bg-background dark:bg-black transition-all">
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
      >
        <SidebarContent />
      </Drawer>

      {/* Main Content */}
      <div className="flex-1 lg:ml-64 flex flex-col min-h-screen transition-all duration-300 relative">
        {/* Sticky Header */}
        <header
          className={`sticky top-0 z-40 h-20 px-4 md:px-8 flex items-center justify-between transition-all duration-300 ${scrolled ? 'glass-header' : 'bg-transparent'
            }`}
        >
          <div className="flex items-center gap-4">
            <Button
              type="text"
              icon={<MenuOutlined className="text-text-primary text-lg" />}
              className="lg:hidden"
              onClick={() => setMobileMenuOpen(true)}
            />
            <Breadcrumb
              items={getBreadcrumbItems()}
              separator={<span className="text-text-tertiary">/</span>}
              className="text-text-secondary hidden sm:block"
            />
          </div>

          <div className="flex items-center gap-2 sm:gap-6">
            <div className="hidden md:flex items-center bg-white/50 dark:bg-white/5 rounded-full px-4 py-2 border border-border focus-within:border-primary/50 transition-colors">
              <SearchOutlined className="text-text-tertiary mr-2" />
              <input
                type="text"
                placeholder="搜索..."
                className="bg-transparent border-none outline-none text-sm text-text-primary placeholder-text-tertiary w-32 focus:w-48 transition-all"
              />
            </div>

            <Button
              type="text"
              shape="circle"
              onClick={handleThemeToggle}
              icon={isDarkMode ? <SunOutlined className="text-text-secondary hover:text-text-primary" /> : <MoonOutlined className="text-text-secondary hover:text-text-primary" />}
            />

            <Button type="text" shape="circle" icon={<BellOutlined className="text-text-secondary hover:text-text-primary" />} />

            <Dropdown menu={{ items: userMenuItems, onClick: handleUserMenuClick }} placement="bottomRight" trigger={['click']}>
              <div className="flex items-center gap-3 cursor-pointer group">
                <div className="text-right hidden sm:block">
                  <div className="text-sm font-medium text-text-primary group-hover:text-primary transition-colors">Administrator</div>
                  <div className="text-xs text-text-tertiary">System Admin</div>
                </div>
                <Avatar
                  size={36}
                  style={{ background: token.colorPrimary }}
                  icon={<UserOutlined />}
                  className="ring-2 ring-white dark:ring-black border border-black/10 dark:border-white/10 group-hover:ring-primary/50 transition-all"
                />
              </div>
            </Dropdown>
          </div>
        </header>

        {/* Page Content */}
        <main className="flex-1 p-4 md:p-8 overflow-y-auto animate-fade-in relative z-0">
          {/* Background Ambient Light */}
          <div className="fixed top-0 left-1/2 -translate-x-1/2 w-[800px] h-[400px] bg-primary/10 rounded-full blur-[100px] pointer-events-none -z-10 animate-breathe"></div>

          <Outlet />
        </main>

        <footer className="py-6 text-center text-xs text-text-tertiary border-t border-border mx-8">
          Prism API Admin System &copy; {new Date().getFullYear()}
        </footer>
      </div>
    </div>
  );
};

export default AdminLayout;
