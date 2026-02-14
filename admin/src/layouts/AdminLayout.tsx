import React, { useState } from 'react';
import { Layout, Menu, Breadcrumb, Avatar, Dropdown, Space, Typography, theme as antTheme, Button } from 'antd';
import {
  DashboardOutlined,
  UserOutlined,
  ApiOutlined,
  BarChartOutlined,
  FileTextOutlined,
  SettingOutlined,
  LogoutOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  DollarOutlined,
  BellOutlined,
  SearchOutlined
} from '@ant-design/icons';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import type { MenuProps } from 'antd';

const { Text } = Typography;

// Menu Configuration
const menuItems: MenuProps['items'] = [
  {
    key: '/dashboard',
    icon: <DashboardOutlined />,
    label: '统计概览',
  },
  {
    type: 'divider',
    style: { margin: '8px 16px', borderColor: 'rgba(255,255,255,0.08)' },
  },
  {
    key: 'business',
    type: 'group',
    label: '业务管理',
    children: [
      { key: '/users', icon: <UserOutlined />, label: '用户管理' },
      { key: '/api-configs', icon: <ApiOutlined />, label: 'API 配置' },
      { key: '/load-balancer', icon: <BarChartOutlined />, label: '负载均衡' },
      { key: '/pricing', icon: <DollarOutlined />, label: '定价管理' },
    ],
  },
  {
    key: 'system',
    type: 'group',
    label: '系统',
    children: [
      { key: '/logs', icon: <FileTextOutlined />, label: '请求日志' },
      { key: '/settings', icon: <SettingOutlined />, label: '系统设置' },
    ],
  },
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
  const [collapsed, setCollapsed] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();
  const { token } = antTheme.useToken();

  const handleMenuClick = ({ key }: { key: string }) => {
    navigate(key);
  };

  const handleUserMenuClick = ({ key }: { key: string }) => {
    if (key === 'logout') {
      localStorage.removeItem('admin_token');
      navigate('/login');
    } else if (key === 'settings') {
      navigate('/settings');
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

  return (
    <div className="min-h-screen bg-page flex text-text-primary font-sans selection:bg-primary/30">

      {/* Sidebar */}
      <aside
        className={`fixed h-screen z-50 glass border-r-0 border-border/40 transition-all duration-300 ease-in-out flex flex-col
          ${collapsed ? 'w-20' : 'w-64'}
        `}
      >
        {/* Logo */}
        <div className={`h-20 flex items-center ${collapsed ? 'justify-center' : 'px-6'} border-b border-border/40 transition-all`}>
          <div className="relative flex items-center gap-3">
            <div className="absolute inset-0 bg-primary/20 blur-lg rounded-full animate-active-pulse"></div>
            <img src="/logo-dark.svg" alt="Prism" className="w-8 h-8 relative z-10" />
            {!collapsed && (
              <span className="text-xl font-bold tracking-tight text-white animate-fade-in relative z-10">
                Prism <span className="text-primary">Admin</span>
              </span>
            )}
          </div>
        </div>

        {/* Menu */}
        <div className="flex-1 overflow-y-auto py-4 px-2 custom-scrollbar">
          <Menu
            mode="inline"
            theme="dark"
            selectedKeys={[location.pathname]}
            items={menuItems}
            onClick={handleMenuClick}
            inlineCollapsed={collapsed}
            style={{
              background: 'transparent',
              border: 'none',
              fontSize: '15px'
            }}
          />
        </div>

        {/* Collapse Trigger */}
        <div
          onClick={() => setCollapsed(!collapsed)}
          className="h-12 border-t border-border/40 flex items-center justify-center cursor-pointer hover:bg-white/5 transition-colors text-text-secondary hover:text-white"
        >
          {collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
        </div>
      </aside>

      {/* Main Content */}
      <div
        className={`flex-1 flex flex-col min-h-screen transition-all duration-300 relative
          ${collapsed ? 'ml-20' : 'ml-64'}
        `}
      >
        {/* Sticky Header */}
        <header className="sticky top-0 z-40 h-20 px-8 flex items-center justify-between glass-header">
          <Breadcrumb
            items={getBreadcrumbItems()}
            separator={<span className="text-text-tertiary">/</span>}
            className="text-text-secondary"
          />

          <div className="flex items-center gap-6">
            <div className="hidden md:flex items-center bg-page-subtle/50 rounded-full px-4 py-2 border border-border/40 focus-within:border-primary/50 transition-colors">
              <SearchOutlined className="text-text-tertiary mr-2" />
              <input
                type="text"
                placeholder="Search..."
                className="bg-transparent border-none outline-none text-sm text-white placeholder-text-tertiary w-32 focus:w-48 transition-all"
              />
            </div>
            <Button type="text" shape="circle" icon={<BellOutlined className="text-text-secondary hover:text-white" />} />

            <Dropdown menu={{ items: userMenuItems, onClick: handleUserMenuClick }} placement="bottomRight" trigger={['click']}>
              <div className="flex items-center gap-3 cursor-pointer group">
                {!collapsed && (
                  <div className="text-right hidden sm:block">
                    <div className="text-sm font-medium text-white group-hover:text-primary transition-colors">Administrator</div>
                    <div className="text-xs text-text-tertiary">System Admin</div>
                  </div>
                )}
                <Avatar
                  size={36}
                  style={{ background: token.colorPrimary }}
                  icon={<UserOutlined />}
                  className="ring-2 ring-background border border-white/10 group-hover:ring-primary/50 transition-all"
                />
              </div>
            </Dropdown>
          </div>
        </header>

        {/* Page Content */}
        <main className="flex-1 p-8 overflow-y-auto animate-fade-in relative z-0">
          {/* Background Ambient Light */}
          <div className="fixed top-20 right-0 w-[500px] h-[500px] bg-primary/5 rounded-full blur-3xl pointer-events-none -z-10"></div>

          <Outlet />
        </main>

        <footer className="py-6 text-center text-xs text-text-tertiary border-t border-border/20 mx-8">
          Prism API Admin System &copy; {new Date().getFullYear()}
        </footer>
      </div>
    </div>
  );
};

export default AdminLayout;
