import React from 'react';
import ReactDOM from 'react-dom/client';
import { RouterProvider } from 'react-router-dom';
import { QueryClientProvider } from '@tanstack/react-query';
import { ConfigProvider, App } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import { router } from './router';
import { queryClient } from './lib/queryClient';
import './styles/index.css';

// Ant Design 5 主题 Token — 统一设计语言
const theme = {
  token: {
    colorPrimary: '#1677ff',
    borderRadius: 6,
    colorBgContainer: '#ffffff',
    fontFamily:
      "-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, 'Noto Sans', sans-serif",
  },
  components: {
    Layout: {
      siderBg: '#001529',
      headerBg: '#ffffff',
      bodyBg: '#f5f7fa',
    },
    Menu: {
      darkItemBg: '#001529',
      darkSubMenuItemBg: '#000c17',
    },
    Card: {
      paddingLG: 24,
    },
    Table: {
      headerBg: '#fafafa',
    },
  },
};

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={zhCN} theme={theme}>
        {/* App 包裹层：为 message/notification/modal 提供全局上下文 */}
        <App>
          <RouterProvider router={router} />
        </App>
      </ConfigProvider>
    </QueryClientProvider>
  </React.StrictMode>,
);
