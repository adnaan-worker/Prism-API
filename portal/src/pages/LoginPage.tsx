import { useState } from 'react';
import { Form, Input, Button, Checkbox, message, Alert } from 'antd';
import { UserOutlined, LockOutlined, GoogleOutlined, GithubOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

const LoginPage = () => {
  const navigate = useNavigate();
  const { login } = useAuth();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const onFinish = async (values: any) => {
    try {
      setLoading(true);
      setError('');
      await login(values.username, values.password);
      message.success('登录成功');
      navigate('/dashboard');
    } catch (err: any) {
      setError(err.message || '登录失败，请检查用户名和密码');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-slate-50 dark:bg-black relative overflow-hidden transition-colors duration-300">
      {/* Background Decoration */}
      <div className="absolute inset-0 pointer-events-none">
        <div className="absolute top-0 left-1/4 w-[500px] h-[500px] bg-primary/20 rounded-full blur-[100px] animate-pulse"></div>
        <div className="absolute bottom-0 right-1/4 w-[500px] h-[500px] bg-purple-500/20 rounded-full blur-[100px] animate-pulse delay-1000"></div>
      </div>

      <div className="w-full max-w-md px-4 relative z-10">
        <div className="glass-card p-8 sm:p-10 rounded-3xl border border-slate-200 dark:border-white/10 shadow-2xl">
          <div className="text-center mb-10">
            <div className="inline-flex items-center justify-center w-12 h-12 rounded-xl bg-primary/10 text-primary mb-4">
              <img src="/logo.svg" alt="Prism API" className="w-8 h-8" />
            </div>
            <h1 className="text-2xl font-bold text-slate-900 dark:text-white mb-2">
              欢迎回来
            </h1>
            <p className="text-slate-500 dark:text-slate-400">
              登录 Prism API 控制台
            </p>
          </div>

          {error && (
            <Alert
              message={error}
              type="error"
              showIcon
              className="mb-6 bg-red-50 dark:bg-red-900/10 border-red-100 dark:border-red-900/20 text-red-600 dark:text-red-400"
            />
          )}

          <Form
            name="login"
            initialValues={{ remember: true }}
            onFinish={onFinish}
            layout="vertical"
            size="large"
          >
            <Form.Item
              name="username"
              rules={[{ required: true, message: '请输入用户名' }]}
            >
              <Input
                prefix={<UserOutlined className="text-slate-400" />}
                placeholder="用户名"
                className="bg-slate-50 dark:bg-white/5 border-slate-200 dark:border-white/10 text-slate-900 dark:text-white placeholder-slate-400 hover:border-primary focus:border-primary hover:bg-white dark:hover:bg-white/10 focus:bg-white dark:focus:bg-white/10 !rounded-xl"
              />
            </Form.Item>

            <Form.Item
              name="password"
              rules={[{ required: true, message: '请输入密码' }]}
            >
              <Input.Password
                prefix={<LockOutlined className="text-slate-400" />}
                placeholder="密码"
                className="bg-slate-50 dark:bg-white/5 border-slate-200 dark:border-white/10 text-slate-900 dark:text-white placeholder-slate-400 hover:border-primary focus:border-primary hover:bg-white dark:hover:bg-white/10 focus:bg-white dark:focus:bg-white/10 !rounded-xl"
              />
            </Form.Item>

            <div className="flex justify-between items-center mb-6">
              <Form.Item name="remember" valuePropName="checked" noStyle>
                <Checkbox className="text-slate-500 dark:text-slate-400">记住我</Checkbox>
              </Form.Item>
              <a href="/forgot-password" className="text-primary hover:text-primary-600 font-medium transition-colors">
                忘记密码？
              </a>
            </div>

            <Form.Item>
              <Button
                type="primary"
                htmlType="submit"
                loading={loading}
                block
                className="!h-12 !rounded-xl !text-base !font-semibold bg-primary hover:bg-primary-600 border-none shadow-lg shadow-primary/25"
              >
                登录
              </Button>
            </Form.Item>

            <div className="relative my-8">
              <div className="absolute inset-0 flex items-center">
                <div className="w-full border-t border-slate-200 dark:border-white/10"></div>
              </div>
              <div className="relative flex justify-center text-sm">
                <span className="px-4 bg-white dark:bg-[#0d1117] text-slate-500">或是通过以下方式登录</span>
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <Button
                icon={<GithubOutlined />}
                block
                className="!h-10 !rounded-xl border-slate-200 dark:border-white/10 bg-white dark:bg-white/5 text-slate-700 dark:text-slate-300 hover:text-primary hover:border-primary dark:hover:bg-white/10"
              >
                Github
              </Button>
              <Button
                icon={<GoogleOutlined />}
                block
                className="!h-10 !rounded-xl border-slate-200 dark:border-white/10 bg-white dark:bg-white/5 text-slate-700 dark:text-slate-300 hover:text-primary hover:border-primary dark:hover:bg-white/10"
              >
                Google
              </Button>
            </div>
          </Form>

          <p className="mt-8 text-center text-slate-500 dark:text-slate-400">
            还没有账号？{' '}
            <a href="/register" className="text-primary hover:text-primary-600 font-medium transition-colors">
              立即注册
            </a>
          </p>
        </div>
      </div>
    </div>
  );
};

export default LoginPage;
