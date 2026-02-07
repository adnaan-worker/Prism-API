import { useState } from 'react';
import { Layout, Menu, Avatar, Dropdown, Button, Space, Typography, theme as antTheme } from 'antd';
import {
  DashboardOutlined,
  KeyOutlined,
  AppstoreOutlined,
  FileTextOutlined,
  UserOutlined,
  LogoutOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  ApiOutlined,
  SunOutlined,
  MoonOutlined,
} from '@ant-design/icons';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { authService } from '../services/authService';

const { Header, Sider, Content } = Layout;
const { Text } = Typography;

interface DashboardLayoutProps {
  isDarkMode: boolean;
  onThemeToggle: () => void;
}

const DashboardLayout = ({ isDarkMode, onThemeToggle }: DashboardLayoutProps) => {
  const [collapsed, setCollapsed] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();
  const { token } = antTheme.useToken();

  const user = authService.getSavedUser();

  const handleLogout = () => {
    authService.logout();
    navigate('/login');
  };

  const menuItems = [
    {
      key: '/dashboard',
      icon: <DashboardOutlined />,
      label: '概览',
      onClick: () => navigate('/dashboard'),
    },
    {
      key: '/dashboard/api-keys',
      icon: <KeyOutlined />,
      label: 'API密钥',
      onClick: () => navigate('/dashboard/api-keys'),
    },
    {
      key: '/dashboard/models',
      icon: <AppstoreOutlined />,
      label: '模型列表',
      onClick: () => navigate('/dashboard/models'),
    },
    {
      key: '/dashboard/docs',
      icon: <FileTextOutlined />,
      label: '使用文档',
      onClick: () => navigate('/dashboard/docs'),
    },
    {
      key: '/dashboard/profile',
      icon: <UserOutlined />,
      label: '个人信息',
      onClick: () => navigate('/dashboard/profile'),
    },
  ];

  const userMenuItems = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: '个人信息',
      onClick: () => navigate('/dashboard/profile'),
    },
    {
      type: 'divider' as const,
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      onClick: handleLogout,
      danger: true,
    },
  ];

  return (
    <Layout className="min-h-screen">
      {/* Sidebar */}
      <Sider
        trigger={null}
        collapsible
        collapsed={collapsed}
        width={240}
        style={{
          overflow: 'auto',
          height: '100vh',
          position: 'fixed',
          left: 0,
          top: 0,
          bottom: 0,
          background: isDarkMode ? token.colorBgContainer : '#001529',
        }}
      >
        {/* Logo */}
        <div
          className="flex items-center justify-center h-16 border-b"
          style={{
            borderColor: isDarkMode ? token.colorBorder : 'rgba(255, 255, 255, 0.1)',
          }}
        >
          <img 
            src={isDarkMode ? "/logo.svg" : "/logo-dark.svg"} 
            alt="Prism API" 
            style={{
              width: collapsed ? 32 : 40,
              height: collapsed ? 32 : 40,
              transition: 'all 0.3s',
            }}
          />
          {!collapsed && (
            <Text
              strong
              style={{
                marginLeft: 12,
                fontSize: 18,
                color: isDarkMode ? token.colorText : '#fff',
              }}
            >
              🌈 Prism API
            </Text>
          )}
        </div>

        {/* Navigation Menu */}
        <Menu
          theme={isDarkMode ? 'light' : 'dark'}
          mode="inline"
          selectedKeys={[location.pathname]}
          items={menuItems}
          style={{ borderRight: 0 }}
        />
      </Sider>

      {/* Main Layout */}
      <Layout style={{ marginLeft: collapsed ? 80 : 240, transition: 'all 0.2s' }}>
        {/* Header */}
        <Header
          style={{
            padding: '0 24px',
            background: token.colorBgContainer,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            borderBottom: `1px solid ${token.colorBorder}`,
            position: 'sticky',
            top: 0,
            zIndex: 1,
          }}
        >
          <Space>
            <Button
              type="text"
              icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
              onClick={() => setCollapsed(!collapsed)}
              style={{
                fontSize: '16px',
                width: 48,
                height: 48,
              }}
            />
          </Space>

          <Space size="middle">
            {/* Theme Toggle */}
            <Button
              type="text"
              icon={isDarkMode ? <SunOutlined /> : <MoonOutlined />}
              onClick={onThemeToggle}
              style={{
                fontSize: '16px',
                width: 40,
                height: 40,
              }}
            />

            {/* User Dropdown */}
            <Dropdown menu={{ items: userMenuItems }} placement="bottomRight">
              <Space className="cursor-pointer hover:opacity-80 transition-opacity">
                <Avatar
                  style={{
                    backgroundColor: token.colorPrimary,
                  }}
                  icon={<UserOutlined />}
                />
                <div className="hidden md:block">
                  <Text strong>{user?.username || 'User'}</Text>
                  <br />
                  <Text type="secondary" style={{ fontSize: 12 }}>
                    额度: {user?.quota ? (user.quota - (user.used_quota || 0)).toLocaleString() : 0}
                  </Text>
                </div>
              </Space>
            </Dropdown>
          </Space>
        </Header>

        {/* Content */}
        <Content
          style={{
            margin: '24px',
            padding: 24,
            minHeight: 280,
            background: token.colorBgContainer,
            borderRadius: token.borderRadius,
          }}
        >
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
};

export default DashboardLayout;
