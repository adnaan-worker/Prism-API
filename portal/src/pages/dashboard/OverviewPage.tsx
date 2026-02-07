import { useState, useEffect } from 'react';
import { 
  Typography, 
  Card, 
  Row, 
  Col, 
  Button, 
  Statistic, 
  Progress, 
  message,
  Alert,
  Steps,
  Space,
  Divider
} from 'antd';
import {
  GiftOutlined,
  ThunderboltOutlined,
  CheckCircleOutlined,
  ApiOutlined,
  CodeOutlined,
  RocketOutlined,
} from '@ant-design/icons';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { quotaService } from '../../services/quotaService';
import { Line } from '@ant-design/charts';

const { Title, Paragraph, Text } = Typography;

const OverviewPage = () => {
  const queryClient = useQueryClient();
  const [hasSignedInToday, setHasSignedInToday] = useState(false);

  // Fetch quota info
  const { data: quotaInfo, isLoading } = useQuery({
    queryKey: ['quotaInfo'],
    queryFn: quotaService.getQuotaInfo,
    refetchInterval: 30000, // Refetch every 30 seconds
  });

  // Check if user has signed in today
  useEffect(() => {
    if (quotaInfo?.last_sign_in) {
      const lastSignIn = new Date(quotaInfo.last_sign_in);
      const today = new Date();
      const isSameDay = 
        lastSignIn.getDate() === today.getDate() &&
        lastSignIn.getMonth() === today.getMonth() &&
        lastSignIn.getFullYear() === today.getFullYear();
      setHasSignedInToday(isSameDay);
    }
  }, [quotaInfo]);

  // Sign-in mutation
  const signInMutation = useMutation({
    mutationFn: quotaService.signIn,
    onSuccess: (data) => {
      message.success(`签到成功！获得 ${data.quota_awarded} tokens`);
      setHasSignedInToday(true);
      queryClient.invalidateQueries({ queryKey: ['quotaInfo'] });
    },
    onError: (error: any) => {
      if (error.response?.data?.error?.code === 409002) {
        message.warning('今日已签到');
        setHasSignedInToday(true);
      } else {
        message.error('签到失败，请稍后重试');
      }
    },
  });

  const handleSignIn = () => {
    signInMutation.mutate();
  };

  // Calculate usage percentage
  const usagePercentage = quotaInfo 
    ? Math.round((quotaInfo.used_quota / quotaInfo.total_quota) * 100)
    : 0;

  // Mock usage trend data (since we don't have historical data endpoint yet)
  const usageTrendData = [
    { date: '周一', usage: 1200 },
    { date: '周二', usage: 1800 },
    { date: '周三', usage: 2100 },
    { date: '周四', usage: 1500 },
    { date: '周五', usage: 2400 },
    { date: '周六', usage: 1900 },
    { date: '周日', usage: 1600 },
  ];

  const chartConfig = {
    data: usageTrendData,
    xField: 'date',
    yField: 'usage',
    smooth: true,
    color: '#1890ff',
    point: {
      size: 5,
      shape: 'circle',
    },
    label: {
      style: {
        fill: '#aaa',
      },
    },
    yAxis: {
      label: {
        formatter: (v: string) => `${v} tokens`,
      },
    },
    tooltip: {
      formatter: (datum: any) => {
        return { name: '使用量', value: `${datum.usage} tokens` };
      },
    },
  };

  return (
    <div style={{ padding: '0 24px' }}>
      <Title level={2}>概览</Title>
      <Paragraph type="secondary">
        欢迎来到 Prism API 控制台，这里是您的使用概览和快速开始指南。
      </Paragraph>

      {/* Quota Statistics Cards */}
      <Row gutter={[16, 16]} style={{ marginTop: 24 }}>
        <Col xs={24} sm={12} lg={6}>
          <Card loading={isLoading}>
            <Statistic
              title="总额度"
              value={quotaInfo?.total_quota || 0}
              suffix="tokens"
              prefix={<ThunderboltOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card loading={isLoading}>
            <Statistic
              title="已使用"
              value={quotaInfo?.used_quota || 0}
              suffix="tokens"
              prefix={<ApiOutlined />}
              valueStyle={{ color: '#faad14' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card loading={isLoading}>
            <Statistic
              title="剩余额度"
              value={quotaInfo?.remaining_quota || 0}
              suffix="tokens"
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <div style={{ textAlign: 'center' }}>
              <Text type="secondary" style={{ display: 'block', marginBottom: 8 }}>
                每日签到
              </Text>
              <Button
                type="primary"
                size="large"
                icon={<GiftOutlined />}
                onClick={handleSignIn}
                disabled={hasSignedInToday}
                loading={signInMutation.isPending}
                style={{ width: '100%' }}
              >
                {hasSignedInToday ? '今日已签到' : '签到领取 1000 tokens'}
              </Button>
              {hasSignedInToday && (
                <Text type="success" style={{ display: 'block', marginTop: 8, fontSize: 12 }}>
                  <CheckCircleOutlined /> 明天再来吧
                </Text>
              )}
            </div>
          </Card>
        </Col>
      </Row>

      {/* Usage Progress */}
      <Card style={{ marginTop: 16 }} title="额度使用情况">
        <Progress
          percent={usagePercentage}
          status={usagePercentage > 90 ? 'exception' : usagePercentage > 70 ? 'normal' : 'success'}
          strokeColor={{
            '0%': '#108ee9',
            '100%': '#87d068',
          }}
        />
        <div style={{ marginTop: 8, display: 'flex', justifyContent: 'space-between' }}>
          <Text type="secondary">
            已使用 {quotaInfo?.used_quota || 0} / {quotaInfo?.total_quota || 0} tokens
          </Text>
          <Text type="secondary">
            {usagePercentage}%
          </Text>
        </div>
        {usagePercentage > 80 && (
          <Alert
            message="额度即将用尽"
            description="您的额度使用已超过80%，请注意合理使用或联系管理员增加额度。"
            type="warning"
            showIcon
            style={{ marginTop: 16 }}
          />
        )}
      </Card>

      {/* Usage Trend Chart */}
      <Card style={{ marginTop: 16 }} title="使用趋势（近7天）">
        <Line {...chartConfig} height={300} />
      </Card>

      {/* Quick Start Guide */}
      <Card style={{ marginTop: 16 }} title={<><RocketOutlined /> 快速开始指南</>}>
        <Steps
          direction="vertical"
          current={-1}
          items={[
            {
              title: '创建 API 密钥',
              description: (
                <Space direction="vertical">
                  <Text>前往 API 密钥页面创建您的第一个密钥，用于调用平台 API。</Text>
                  <Button type="link" href="/dashboard/api-keys" style={{ padding: 0 }}>
                    前往创建 →
                  </Button>
                </Space>
              ),
              icon: <ApiOutlined />,
            },
            {
              title: '查看可用模型',
              description: (
                <Space direction="vertical">
                  <Text>浏览平台支持的所有 AI 模型，包括 GPT-4、Claude、Gemini 等。</Text>
                  <Button type="link" href="/dashboard/models" style={{ padding: 0 }}>
                    查看模型列表 →
                  </Button>
                </Space>
              ),
              icon: <ThunderboltOutlined />,
            },
            {
              title: '阅读 API 文档',
              description: (
                <Space direction="vertical">
                  <Text>学习如何使用统一的 API 接口调用不同提供商的模型。</Text>
                  <Button type="link" href="/dashboard/docs" style={{ padding: 0 }}>
                    查看文档 →
                  </Button>
                </Space>
              ),
              icon: <CodeOutlined />,
            },
            {
              title: '开始调用 API',
              description: (
                <div>
                  <Text>使用您的 API 密钥开始调用，示例代码：</Text>
                  <pre style={{ 
                    background: '#f5f5f5', 
                    padding: 12, 
                    borderRadius: 4, 
                    marginTop: 8,
                    overflow: 'auto'
                  }}>
{`curl https://api.example.com/v1/chat/completions \\
  -H "Authorization: Bearer YOUR_API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'`}
                  </pre>
                </div>
              ),
              icon: <RocketOutlined />,
            },
          ]}
        />
      </Card>

      <Divider />

      {/* Additional Tips */}
      <Card style={{ marginTop: 16 }} title="💡 使用提示">
        <Row gutter={[16, 16]}>
          <Col xs={24} md={12}>
            <Alert
              message="每日签到"
              description="每天签到可获得 1000 tokens，不要忘记哦！"
              type="info"
              showIcon
            />
          </Col>
          <Col xs={24} md={12}>
            <Alert
              message="合理使用"
              description="建议根据实际需求选择合适的模型，以节省额度。"
              type="success"
              showIcon
            />
          </Col>
        </Row>
      </Card>
    </div>
  );
};

export default OverviewPage;
