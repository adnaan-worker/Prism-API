import { StrictMode, useState, useEffect } from 'react';
import ReactDOM from 'react-dom/client';
import { RouterProvider } from 'react-router-dom';
import { QueryClientProvider } from '@tanstack/react-query';
import { ConfigProvider } from 'antd';
import { queryClient } from './lib/queryClient';
import { getAntdTheme } from './theme/antdTheme';
import { router, setThemeState } from './router';
import { AuthProvider } from './context/AuthContext';
import './i18n/config';
import './styles/index.css';

const App = () => {
  const [isDarkMode, setIsDarkMode] = useState(() => {
    const saved = localStorage.getItem('theme');
    return saved === 'dark' || (!saved && window.matchMedia('(prefers-color-scheme: dark)').matches);
  });

  useEffect(() => {
    const root = window.document.documentElement;
    if (isDarkMode) {
      root.classList.add('dark');
      localStorage.setItem('theme', 'dark');
    } else {
      root.classList.remove('dark');
      localStorage.setItem('theme', 'light');
    }

    // Update router with current theme state (legacy support for router props if needed)
    setThemeState(isDarkMode, () => setIsDarkMode((prev) => !prev));
  }, [isDarkMode]);

  const toggleTheme = () => {
    setIsDarkMode((prev) => !prev);
  };

  // Update router with theme toggle callback
  useEffect(() => {
    setThemeState(isDarkMode, toggleTheme);
  }, [isDarkMode]);

  // Ant Design theme configuration
  const themeConfig = getAntdTheme(isDarkMode);

  return (
    <QueryClientProvider client={queryClient}>
      <ConfigProvider theme={themeConfig}>
        <AuthProvider>
          <RouterProvider router={router} />
        </AuthProvider>
      </ConfigProvider>
    </QueryClientProvider>
  );
};

ReactDOM.createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
);
