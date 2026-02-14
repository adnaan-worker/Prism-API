import { useState } from 'react';
import { Form, Input, Button, Checkbox, message, Alert } from 'antd';
import { UserOutlined, LockOutlined, MailOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

const RegisterPage = () => {
  const navigate = useNavigate();
  const { register } = useAuth();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const onFinish = async (values: any) => {
    try {
      setLoading(true);
      setError('');
      await register(values.username, values.password, values.email);
      message.success('注册成功，请登录');
      navigate('/login');
    } catch (err: any) {
      setError(err.message || '注册失败，请稍后重试');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-slate-50 dark:bg-black relative overflow-hidden transition-colors duration-300">
      {/* Background Decoration */}
      <div className="absolute inset-0 pointer-events-none">
        <div className="absolute top-0 right-1/4 w-[500px] h-[500px] bg-primary/20 rounded-full blur-[100px] animate-pulse"></div>
        <div className="absolute bottom-0 left-1/4 w-[500px] h-[500px] bg-purple-500/20 rounded-full blur-[100px] animate-pulse delay-1000"></div>
      </div>

      <div className="w-full max-w-md px-4 relative z-10">
        <div className="glass-card p-8 sm:p-10 rounded-3xl border border-slate-200 dark:border-white/10 shadow-2xl">
          <div className="text-center mb-10">
            <div className="inline-flex items-center justify-center w-12 h-12 rounded-xl bg-primary/10 text-primary mb-4">
              <img src="/logo.svg" alt="Prism API" className="w-8 h-8" />
            </div>
            <h1 className="text-2xl font-bold text-slate-900 dark:text-white mb-2">
              创建账号
            </h1>
            <p className="text-slate-500 dark:text-slate-400">
              开始您的 AI 开发之旅
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
            name="register"
            initialValues={{ agreement: true }}
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
              name="email"
              rules={[
                { required: true, message: '请输入邮箱' },
                { type: 'email', message: '请输入有效的邮箱地址' }
              ]}
            >
              <Input
                prefix={<MailOutlined className="text-slate-400" />}
                placeholder="邮箱地址"
                className="bg-slate-50 dark:bg-white/5 border-slate-200 dark:border-white/10 text-slate-900 dark:text-white placeholder-slate-400 hover:border-primary focus:border-primary hover:bg-white dark:hover:bg-white/10 focus:bg-white dark:focus:bg-white/10 !rounded-xl"
              />
            </Form.Item>

            <Form.Item
              name="password"
              rules={[
                { required: true, message: '请输入密码' },
                { min: 6, message: '密码至少6个字符' }
              ]}
            >
              <Input.Password
                prefix={<LockOutlined className="text-slate-400" />}
                placeholder="密码"
                className="bg-slate-50 dark:bg-white/5 border-slate-200 dark:border-white/10 text-slate-900 dark:text-white placeholder-slate-400 hover:border-primary focus:border-primary hover:bg-white dark:hover:bg-white/10 focus:bg-white dark:focus:bg-white/10 !rounded-xl"
              />
            </Form.Item>

            <Form.Item
              name="confirm"
              dependencies={['password']}
              rules={[
                { required: true, message: '请确认密码' },
                ({ getFieldValue }) => ({
                  validator(_, value) {
                    if (!value || getFieldValue('password') === value) {
                      return Promise.resolve();
                    }
                    return Promise.reject(new Error('两次输入的密码不一致'));
                  },
                }),
              ]}
            >
              <Input.Password
                prefix={<LockOutlined className="text-slate-400" />}
                placeholder="确认密码"
                className="bg-slate-50 dark:bg-white/5 border-slate-200 dark:border-white/10 text-slate-900 dark:text-white placeholder-slate-400 hover:border-primary focus:border-primary hover:bg-white dark:hover:bg-white/10 focus:bg-white dark:focus:bg-white/10 !rounded-xl"
              />
            </Form.Item>

            <Form.Item
              name="agreement"
              valuePropName="checked"
              rules={[
                {
                  validator: (_, value) =>
                    value ? Promise.resolve() : Promise.reject(new Error('请阅读并同意服务条款')),
                },
              ]}
            >
              <Checkbox className="text-slate-500 dark:text-slate-400">
                我已阅读并同意 <a href="/terms" className="text-primary hover:text-primary-600">服务条款</a> 和 <a href="/privacy" className="text-primary hover:text-primary-600">隐私政策</a>
              </Checkbox>
            </Form.Item>

            <Form.Item>
              <Button
                type="primary"
                htmlType="submit"
                loading={loading}
                block
                className="!h-12 !rounded-xl !text-base !font-semibold bg-primary hover:bg-primary-600 border-none shadow-lg shadow-primary/25"
              >
                注册账号
              </Button>
            </Form.Item>
          </Form>

          <p className="mt-8 text-center text-slate-500 dark:text-slate-400">
            已有账号？{' '}
            <a href="/login" className="text-primary hover:text-primary-600 font-medium transition-colors">
              立即登录
            </a>
          </p>
        </div>
      </div>
    </div>
  );
};

export default RegisterPage;
