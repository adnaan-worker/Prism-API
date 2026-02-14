import React, { useState } from 'react';
import { Form, Input, Button, Typography, message } from 'antd';
import { UserOutlined, LockOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { authService } from '../services/authService';

const { Title, Text } = Typography;

const LoginPage: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  const onFinish = async (values: { username: string; password: string }) => {
    setLoading(true);
    try {
      const response = await authService.login(values);

      if (!response.user.is_admin) {
        message.error('您没有管理员权限');
        return;
      }

      authService.setToken(response.token);
      message.success('登录成功');
      navigate('/dashboard');
    } catch (error: any) {
      message.error(error.response?.data?.error?.message || '登录失败，请检查用户名和密码');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-page flex flex-col items-center justify-center p-4 relative overflow-hidden">
      {/* Background Gradients */}
      <div className="absolute top-[-10%] left-[-10%] w-[40%] h-[40%] bg-primary-500/20 rounded-full blur-[120px] pointer-events-none" />
      <div className="absolute bottom-[-10%] right-[-10%] w-[40%] h-[40%] bg-purple-500/20 rounded-full blur-[120px] pointer-events-none" />

      <div className="glass-card w-full max-w-md p-8 sm:p-10 relative z-10">
        {/* Logo & Title */}
        <div className="text-center mb-8">
          <div className="flex justify-center mb-4">
            <img
              src="/logo.svg"
              alt="Prism API"
              className="w-12 h-12"
              onError={(e) => {
                (e.target as HTMLImageElement).style.display = 'none';
              }}
            />
          </div>
          <Title level={3} className="!mb-1 !font-semibold !text-text-primary">
            Prism API
          </Title>
          <Text className="text-text-secondary text-sm">
            管理后台
          </Text>
        </div>

        {/* Login Form */}
        <Form
          name="login"
          onFinish={onFinish}
          autoComplete="off"
          size="large"
          layout="vertical"
          initialValues={{ remember: true }}
        >
          <Form.Item
            name="username"
            rules={[{ required: true, message: '请输入用户名' }]}
          >
            <Input
              prefix={<UserOutlined className="text-text-tertiary" />}
              placeholder="用户名"
              className="!bg-page-subtle !border-border !text-text-primary hover:!border-primary focus:!border-primary"
              autoFocus
            />
          </Form.Item>

          <Form.Item
            name="password"
            rules={[{ required: true, message: '请输入密码' }]}
            className="mb-8"
          >
            <Input.Password
              prefix={<LockOutlined className="text-text-tertiary" />}
              placeholder="密码"
              className="!bg-page-subtle !border-border !text-text-primary hover:!border-primary focus:!border-primary"
            />
          </Form.Item>

          <Form.Item className="mb-0">
            <Button
              type="primary"
              htmlType="submit"
              loading={loading}
              block
              className="!h-11 !font-medium !text-base shadow-lg shadow-primary/20 hover:!shadow-primary/40 transition-all duration-300"
            >
              登 录
            </Button>
          </Form.Item>
        </Form>
      </div>

      {/* Footer */}
      <div className="mt-8 text-center relative z-10">
        <Text className="text-text-tertiary text-xs">
          仅限授权管理员访问 &middot; v1.0.0
        </Text>
      </div>
    </div>
  );
};

export default LoginPage;
