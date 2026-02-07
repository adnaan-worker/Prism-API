import {
  Typography,
  Card,
  Row,
  Col,
  Descriptions,
  Statistic,
  Progress,
  Tag,
  Space,
  Divider,
  Avatar,
} from 'antd';
import {
  UserOutlined,
  MailOutlined,
  CalendarOutlined,
  ThunderboltOutlined,
  ApiOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
} from '@ant-design/icons';
import { useQuery } from '@tanstack/react-query';
import { authService } from '../../services/authService';
import { quotaService } from '../../services/quotaService';
import { Column } from '@ant-design/charts';

const { Title, Text } = Typography;

const ProfilePage = () => {
  // Fetch user info
  const { data: user, isLoading: userLoading } = useQuery({
    queryKey: ['currentUser'],
    queryFn: authService.getCurrentUser,
  });

  // Fetch quota info
  const { data: quotaInfo, isLoading: quotaLoading } = useQuery({
    queryKey: ['quotaInfo'],
    queryFn: quotaService.getQuotaInfo,
  });

  // Calculate usage percentage
  const usagePercentage = quotaInfo
    ? Math.round((quotaInfo.used_quota / quotaInfo.total_quota) * 100)
    : 0;

  // Format date
  const formatDate = (dateString?: string) => {
    if (!dateString) return '从未';
    return new Date(dateString).toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  // Mock usage history data (7 days)
  // In a real implementation, this would come from an API endpoint
  const usageHistoryData = [
    { date: '周一', tokens: 1200 },
    { date: '周二', tokens: 1800 },
    { date: '周三', tokens: 2100 },
    { date: '周四', tokens: 1500 },
    { date: '周五', tokens: 2400 },
    { date: '周六', tokens: 1900 },
    { date: '周日', tokens: 1600 },
  ];

  const chartConfig = {
    data: usageHistoryData,
    xField: 'date',
    yField: 'tokens',
    color: '#1890ff',
    columnStyle: {
      radius: [8, 8, 0, 0],
    },
    label: {
      position: 'top' as const,
      style: {
        fill: '#000000',
        opacity: 0.6,
      },
    },
    xAxis: {
      label: {
        autoHide: false,
        autoRotate: false,
      },
    },
    yAxis: {
      label: {
        formatter: (v: string) => `${v}`,
      },
    },
    tooltip: {
      formatter: (datum: any) => {
        return { name: '使用量', value: `${datum.tokens} tokens` };
      },
    },
    meta: {
      tokens: {
        alias: '使用量 (tokens)',
      },
    },
  };

  return (
    <div style={{ padding: '0 24px' }}>
      <Title level={2}>个人信息</Title>
      <Text type="secondary">
        查看和管理您的账户信息、额度使用情况和使用历史。
      </Text>

      {/* Basic Information Card */}
      <Card
        style={{ marginTop: 24 }}
        title={
          <Space>
            <UserOutlined />
            <span>基本信息</span>
          </Space>
        }
        loading={userLoading}
      >
        <Row gutter={[24, 24]}>
          <Col xs={24} md={6} style={{ textAlign: 'center' }}>
            <Avatar size={120} icon={<UserOutlined />} style={{ backgroundColor: '#1890ff' }} />
            <div style={{ marginTop: 16 }}>
              <Text strong style={{ fontSize: 18 }}>
                {user?.username}
              </Text>
            </div>
            <div style={{ marginTop: 8 }}>
              <Tag color={user?.status === 'active' ? 'success' : 'default'}>
                {user?.status === 'active' ? '正常' : user?.status}
              </Tag>
              {user?.is_admin && <Tag color="gold">管理员</Tag>}
            </div>
          </Col>
          <Col xs={24} md={18}>
            <Descriptions column={{ xs: 1, sm: 2 }} bordered>
              <Descriptions.Item label={<><UserOutlined /> 用户名</>}>
                {user?.username}
              </Descriptions.Item>
              <Descriptions.Item label={<><MailOutlined /> 邮箱</>}>
                {user?.email}
              </Descriptions.Item>
              <Descriptions.Item label={<><CalendarOutlined /> 注册时间</>}>
                {formatDate(user?.created_at)}
              </Descriptions.Item>
              <Descriptions.Item label={<><ClockCircleOutlined /> 最后签到</>}>
                {formatDate(user?.last_sign_in)}
              </Descriptions.Item>
              <Descriptions.Item label="账户状态">
                <Tag color={user?.status === 'active' ? 'success' : 'default'}>
                  {user?.status === 'active' ? '正常' : user?.status}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="用户ID">
                {user?.id}
              </Descriptions.Item>
            </Descriptions>
          </Col>
        </Row>
      </Card>

      {/* Quota Statistics Card */}
      <Card
        style={{ marginTop: 16 }}
        title={
          <Space>
            <ThunderboltOutlined />
            <span>额度统计</span>
          </Space>
        }
        loading={quotaLoading}
      >
        <Row gutter={[16, 16]}>
          <Col xs={24} sm={8}>
            <Card>
              <Statistic
                title="总额度"
                value={quotaInfo?.total_quota || 0}
                suffix="tokens"
                prefix={<ThunderboltOutlined />}
                valueStyle={{ color: '#1890ff' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={8}>
            <Card>
              <Statistic
                title="已使用"
                value={quotaInfo?.used_quota || 0}
                suffix="tokens"
                prefix={<ApiOutlined />}
                valueStyle={{ color: '#faad14' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={8}>
            <Card>
              <Statistic
                title="剩余额度"
                value={quotaInfo?.remaining_quota || 0}
                suffix="tokens"
                prefix={<CheckCircleOutlined />}
                valueStyle={{ color: '#52c41a' }}
              />
            </Card>
          </Col>
        </Row>

        <Divider />

        <div>
          <Text strong>使用进度</Text>
          <Progress
            percent={usagePercentage}
            status={
              usagePercentage > 90
                ? 'exception'
                : usagePercentage > 70
                ? 'normal'
                : 'success'
            }
            strokeColor={{
              '0%': '#108ee9',
              '100%': '#87d068',
            }}
            style={{ marginTop: 8 }}
          />
          <div style={{ marginTop: 8, display: 'flex', justifyContent: 'space-between' }}>
            <Text type="secondary">
              已使用 {quotaInfo?.used_quota || 0} / {quotaInfo?.total_quota || 0} tokens
            </Text>
            <Text type="secondary">{usagePercentage}%</Text>
          </div>
        </div>
      </Card>

      {/* Usage History Chart */}
      <Card
        style={{ marginTop: 16 }}
        title={
          <Space>
            <ApiOutlined />
            <span>使用历史（近7天）</span>
          </Space>
        }
      >
        <Column {...chartConfig} height={300} />
        <Divider />
        <Row gutter={[16, 16]}>
          <Col xs={24} sm={8}>
            <Statistic
              title="7日总使用"
              value={usageHistoryData.reduce((sum, item) => sum + item.tokens, 0)}
              suffix="tokens"
            />
          </Col>
          <Col xs={24} sm={8}>
            <Statistic
              title="日均使用"
              value={Math.round(
                usageHistoryData.reduce((sum, item) => sum + item.tokens, 0) / 7
              )}
              suffix="tokens"
            />
          </Col>
          <Col xs={24} sm={8}>
            <Statistic
              title="最高单日"
              value={Math.max(...usageHistoryData.map((item) => item.tokens))}
              suffix="tokens"
            />
          </Col>
        </Row>
      </Card>
    </div>
  );
};

export default ProfilePage;
