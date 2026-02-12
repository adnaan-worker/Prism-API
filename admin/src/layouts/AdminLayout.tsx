import React, { useState } from 'react';
import { Layout, Menu, Breadcrumb, Avatar, Dropdown, Space, Typography } from 'antd';
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
} from '@ant-design/icons';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import type { MenuProps } from 'antd';

const { Header, Sider, Content } = Layout;
const { Text } = Typography;

// 菜单项配置 — 分组结构
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
      {
        key: '/users',
        icon: <UserOutlined />,
        label: '用户管理',
      },
      {
        key: '/api-configs',
        icon: <ApiOutlined />,
        label: 'API 配置',
      },
      {
        key: '/load-balancer',
        icon: <BarChartOutlined />,
        label: '负载均衡',
      },
      {
        key: '/pricing',
        icon: <DollarOutlined />,
        label: '定价管理',
      },
    ],
  },
  {
    key: 'system',
    type: 'group',
    label: '系统',
    children: [
      {
        key: '/logs',
        icon: <FileTextOutlined />,
        label: '请求日志',
      },
      {
        key: '/settings',
        icon: <SettingOutlined />,
        label: '系统设置',
      },
    ],
  },
];

// 面包屑映射
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

  // 用户下拉菜单
  const userMenuItems: MenuProps['items'] = [
    {
      key: 'settings',
      icon: <SettingOutlined />,
      label: '系统设置',
    },
    { type: 'divider' },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      danger: true,
    },
  ];

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

  // 面包屑 — 使用 items API（非废弃的 Breadcrumb.Item）
  const getBreadcrumbItems = () => {
    const segments = location.pathname.split('/').filter(Boolean);
    const items = [{ title: '首页', href: '/dashboard', onClick: (e: React.MouseEvent) => { e.preventDefault(); navigate('/dashboard'); } }];

    segments.forEach((seg) => {
      const name = breadcrumbMap[seg];
      if (name) {
        items.push({ title: name, href: '', onClick: undefined as any });
      }
    });

    return items;
  };

  const siderWidth = collapsed ? 80 : 220;

  return (
    <Layout style={{ minHeight: '100vh' }}>
      {/* 侧边栏 */}
      <Sider
        collapsible
        collapsed={collapsed}
        onCollapse={setCollapsed}
        trigger={null}
        width={220}
        style={{
          overflow: 'auto',
          height: '100vh',
          position: 'fixed',
          left: 0,
          top: 0,
          bottom: 0,
          zIndex: 10,
        }}
        theme="dark"
      >
        {/* Logo */}
        <div
          style={{
            height: 64,
            display: 'flex',
            alignItems: 'center',
            justifyContent: collapsed ? 'center' : 'flex-start',
            padding: collapsed ? 0 : '0 20px',
            color: '#fff',
            fontSize: collapsed ? 16 : 18,
            fontWeight: 600,
            borderBottom: '1px solid rgba(255, 255, 255, 0.06)',
            gap: 10,
            letterSpacing: collapsed ? 0 : 0.5,
            userSelect: 'none',
            overflow: 'hidden',
            whiteSpace: 'nowrap',
          }}
        >
          <img
            src="/logo-dark.svg"
            alt="Prism API"
            style={{
              width: collapsed ? 28 : 32,
              height: collapsed ? 28 : 32,
              flexShrink: 0,
            }}
            onError={(e) => {
              // Logo 加载失败时隐藏
              (e.target as HTMLImageElement).style.display = 'none';
            }}
          />
          {!collapsed && <span>Prism API</span>}
        </div>

        {/* 导航菜单 */}
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[location.pathname]}
          items={menuItems}
          onClick={handleMenuClick}
          style={{ borderRight: 0, marginTop: 4 }}
        />

        {/* 侧边栏底部 — 折叠按钮 */}
        <div
          style={{
            position: 'absolute',
            bottom: 0,
            left: 0,
            right: 0,
            height: 48,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            borderTop: '1px solid rgba(255, 255, 255, 0.06)',
            cursor: 'pointer',
            color: 'rgba(255, 255, 255, 0.45)',
            transition: 'color 0.2s',
          }}
          onClick={() => setCollapsed(!collapsed)}
          onMouseEnter={(e) => (e.currentTarget.style.color = 'rgba(255, 255, 255, 0.85)')}
          onMouseLeave={(e) => (e.currentTarget.style.color = 'rgba(255, 255, 255, 0.45)')}
        >
          {collapsed ? <MenuUnfoldOutlined style={{ fontSize: 16 }} /> : <MenuFoldOutlined style={{ fontSize: 16 }} />}
        </div>
      </Sider>

      {/* 主内容区 */}
      <Layout style={{ marginLeft: siderWidth, transition: 'margin-left 0.2s' }}>
        {/* Header */}
        <Header
          style={{
            padding: '0 24px',
            background: '#fff',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            boxShadow: '0 1px 2px 0 rgba(0, 0, 0, 0.03), 0 1px 6px -1px rgba(0, 0, 0, 0.02), 0 2px 4px 0 rgba(0, 0, 0, 0.02)',
            position: 'sticky',
            top: 0,
            zIndex: 9,
            height: 56,
            lineHeight: '56px',
          }}
        >
          {/* 面包屑 */}
          <Breadcrumb items={getBreadcrumbItems()} />

          {/* 右侧 */}
          <Dropdown
            menu={{ items: userMenuItems, onClick: handleUserMenuClick }}
            placement="bottomRight"
            trigger={['click']}
          >
            <Space style={{ cursor: 'pointer', padding: '0 8px', borderRadius: 6, transition: 'background 0.2s' }}>
              <Avatar
                size={32}
                icon={<UserOutlined />}
                style={{ background: '#1677ff' }}
              />
              <Text style={{ fontSize: 14 }}>管理员</Text>
            </Space>
          </Dropdown>
        </Header>

        {/* 内容区域 */}
        <Content
          className="page-content"
          style={{
            margin: 24,
            minHeight: 'calc(100vh - 56px - 48px)',
          }}
        >
          <Outlet />
        </Content>

        {/* Footer */}
        <div
          style={{
            textAlign: 'center',
            padding: '12px 0',
            color: 'rgba(0, 0, 0, 0.25)',
            fontSize: 12,
          }}
        >
          Prism API Admin &copy; {new Date().getFullYear()}
        </div>
      </Layout>
    </Layout>
  );
};

export default AdminLayout;
