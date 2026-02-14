import React from 'react';
import ReactDOM from 'react-dom/client';
import { RouterProvider } from 'react-router-dom';
import { QueryClientProvider } from '@tanstack/react-query';
import { ConfigProvider, App } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import { router } from './router';
import { queryClient } from './lib/queryClient';
import { getAdminTheme } from './theme/adminTheme';
import './styles/index.css';

// Ant Design Theme
const theme = getAdminTheme();

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
