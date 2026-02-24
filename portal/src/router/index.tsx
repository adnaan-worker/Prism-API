import { createBrowserRouter, Navigate } from 'react-router-dom';
import LandingPage from '../pages/LandingPage';
import LoginPage from '../pages/LoginPage';
import RegisterPage from '../pages/RegisterPage';
import DashboardLayout from '../layouts/DashboardLayout';
import ProtectedRoute from '../components/ProtectedRoute';
import OverviewPage from '../pages/dashboard/OverviewPage';
import ApiKeysPage from '../pages/dashboard/ApiKeysPage';
import ModelsPage from '../pages/dashboard/ModelsPage';
import DocsPage from '../pages/dashboard/DocsPage';
import ProfilePage from '../pages/dashboard/ProfilePage';
import PlaygroundPage from '../pages/dashboard/PlaygroundPage';

// Theme state management (will be passed from main.tsx)
let isDarkMode = false;
let themeToggleCallback: (() => void) | null = null;

export const setThemeState = (dark: boolean, callback: () => void) => {
  isDarkMode = dark;
  themeToggleCallback = callback;
};

export const router = createBrowserRouter([
  {
    path: '/',
    element: <LandingPage />,
  },
  {
    path: '/login',
    element: <LoginPage />,
  },
  {
    path: '/register',
    element: <RegisterPage />,
  },
  {
    path: '/dashboard',
    element: (
      <ProtectedRoute>
        <DashboardLayout
          isDarkMode={isDarkMode}
          onThemeToggle={() => themeToggleCallback?.()}
        />
      </ProtectedRoute>
    ),
    children: [
      {
        index: true,
        element: <OverviewPage />,
      },
      {
        path: 'api-keys',
        element: <ApiKeysPage />,
      },
      {
        path: 'models',
        element: <ModelsPage />,
      },
      {
        path: 'docs',
        element: <DocsPage />,
      },
      {
        path: 'playground',
        element: <PlaygroundPage />,
      },
      {
        path: 'profile',
        element: <ProfilePage />,
      },
    ],
  },
  {
    path: '*',
    element: <Navigate to="/" replace />,
  },
]);
