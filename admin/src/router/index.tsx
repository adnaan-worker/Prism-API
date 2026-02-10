import { createBrowserRouter, Navigate } from 'react-router-dom';
import AdminLayout from '../layouts/AdminLayout';
import LoginPage from '../pages/LoginPage';
import DashboardPage from '../pages/DashboardPage';
import UsersPage from '../pages/UsersPage';
import ApiConfigsPage from '../pages/ApiConfigsPage';
import LoadBalancerPage from '../pages/LoadBalancerPage';
import PricingPage from '../pages/PricingPage';
import LogsPage from '../pages/LogsPage';
import SettingsPage from '../pages/SettingsPage';
import ProtectedRoute from '../components/ProtectedRoute';

export const router = createBrowserRouter([
  {
    path: '/login',
    element: <LoginPage />,
  },
  {
    path: '/',
    element: (
      <ProtectedRoute>
        <AdminLayout />
      </ProtectedRoute>
    ),
    children: [
      {
        index: true,
        element: <Navigate to="/dashboard" replace />,
      },
      {
        path: 'dashboard',
        element: <DashboardPage />,
      },
      {
        path: 'users',
        element: <UsersPage />,
      },
      {
        path: 'api-configs',
        element: <ApiConfigsPage />,
      },
      {
        path: 'load-balancer',
        element: <LoadBalancerPage />,
      },
      {
        path: 'pricing',
        element: <PricingPage />,
      },
      {
        path: 'logs',
        element: <LogsPage />,
      },
      {
        path: 'settings',
        element: <SettingsPage />,
      },
    ],
  },
]);
