import { Link, useNavigate } from 'react-router-dom';
import { Form, Input, Button, Card, Typography, message, Space } from 'antd';
import { UserOutlined, LockOutlined } from '@ant-design/icons';
import { useMutation } from '@tanstack/react-query';
import { authService } from '../services/authService';
import { LoginRequest } from '../types';

const { Title, Text } = Typography;

export default function LoginPage() {
  const navigate = useNavigate();
  const [form] = Form.useForm();

  const loginMutation = useMutation({
    mutationFn: (credentials: LoginRequest) => authService.login(credentials),
    onSuccess: (data) => {
      authService.saveAuthData(data.token, data.user);
      message.success('登录成功！');
      navigate('/dashboard');
    },
    onError: (error: any) => {
      const errorMessage = error.response?.data?.error?.message || '登录失败，请检查用户名和密码';
      message.error(errorMessage);
    },
  });

  const handleSubmit = (values: LoginRequest) => {
    loginMutation.mutate(values);
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center p-4">
      <Card className="w-full max-w-md shadow-lg">
        <div className="text-center mb-8">
          <Title level={2} className="!mb-2">
            登录
          </Title>
          <Text type="secondary">欢迎回到 Prism API</Text>
        </div>

        <Form
          form={form}
          name="login"
          onFinish={handleSubmit}
          layout="vertical"
          size="large"
          autoComplete="off"
        >
          <Form.Item
            name="username"
            rules={[
              { required: true, message: '请输入用户名' },
              { min: 3, message: '用户名至少3个字符' },
            ]}
          >
            <Input
              prefix={<UserOutlined />}
              placeholder="用户名"
              autoComplete="username"
            />
          </Form.Item>

          <Form.Item
            name="password"
            rules={[
              { required: true, message: '请输入密码' },
              { min: 6, message: '密码至少6个字符' },
            ]}
          >
            <Input.Password
              prefix={<LockOutlined />}
              placeholder="密码"
              autoComplete="current-password"
            />
          </Form.Item>

          <Form.Item>
            <Button
              type="primary"
              htmlType="submit"
              className="w-full"
              loading={loginMutation.isPending}
            >
              登录
            </Button>
          </Form.Item>

          <div className="text-center">
            <Space split="|">
              <Text type="secondary">还没有账号？</Text>
              <Link to="/register">
                <Button type="link" className="p-0 h-auto">立即注册</Button>
              </Link>
            </Space>
          </div>

          <div className="text-center mt-4">
            <Link to="/">
              <Text type="secondary">返回首页</Text>
            </Link>
          </div>
        </Form>
      </Card>
    </div>
  );
}
