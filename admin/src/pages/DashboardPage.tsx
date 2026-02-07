import React from 'react';
import { Card, Row, Col, Statistic, Table, Typography, Spin } from 'antd';
import {
  UserOutlined,
  TeamOutlined,
  ApiOutlined,
  RiseOutlined,
} from '@ant-design/icons';
import { useQuery } from '@tanstack/react-query';
import ReactECharts from 'echarts-for-react';
import { statsService } from '../services/statsService';
import type { EChartsOption } from 'echarts';

const { Title } = Typography;

const DashboardPage: React.FC = () => {
  // 获取统计概览数据
  const { data: overview, isLoading: overviewLoading } = useQuery({
    queryKey: ['stats', 'overview'],
    queryFn: statsService.getOverview,
  });

  // 获取请求趋势数据
  const { data: trendData, isLoading: trendLoading } = useQuery({
    queryKey: ['stats', 'trend'],
    queryFn: () => statsService.getRequestTrend(7),
  });

  // 获取模型使用排行
  const { data: modelUsage, isLoading: modelLoading } = useQuery({
    queryKey: ['stats', 'models'],
    queryFn: statsService.getModelUsage,
  });

  // 请求趋势图表配置
  const trendChartOption: EChartsOption = {
    title: {
      text: '请求趋势（最近7天）',
      left: 'center',
    },
    tooltip: {
      trigger: 'axis',
    },
    xAxis: {
      type: 'category',
      data: trendData?.map((item) => item.date) || [],
    },
    yAxis: {
      type: 'value',
      name: '请求数',
    },
    series: [
      {
        name: '请求数',
        type: 'line',
        smooth: true,
        data: trendData?.map((item) => item.count) || [],
        areaStyle: {
          color: {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(24, 144, 255, 0.3)' },
              { offset: 1, color: 'rgba(24, 144, 255, 0.05)' },
            ],
          },
        },
        itemStyle: {
          color: '#1890ff',
        },
      },
    ],
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true,
    },
  };

  // 模型使用排行图表配置
  const modelChartOption: EChartsOption = {
    title: {
      text: '模型使用排行',
      left: 'center',
    },
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow',
      },
    },
    xAxis: {
      type: 'value',
      name: '调用次数',
    },
    yAxis: {
      type: 'category',
      data: modelUsage?.map((item) => item.model).reverse() || [],
    },
    series: [
      {
        name: '调用次数',
        type: 'bar',
        data: modelUsage?.map((item) => item.count).reverse() || [],
        itemStyle: {
          color: {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 1,
            y2: 0,
            colorStops: [
              { offset: 0, color: '#667eea' },
              { offset: 1, color: '#764ba2' },
            ],
          },
        },
      },
    ],
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true,
    },
  };

  // 获取最近活动（请求日志）
  const { data: recentLogs, isLoading: logsLoading } = useQuery({
    queryKey: ['logs', 'recent'],
    queryFn: () => statsService.getRecentLogs(5),
  });

  // 最近活动表格列配置
  const activityColumns = [
    {
      title: '时间',
      dataIndex: 'time',
      key: 'time',
    },
    {
      title: '用户',
      dataIndex: 'user',
      key: 'user',
    },
    {
      title: '操作',
      dataIndex: 'action',
      key: 'action',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <span style={{ color: status === '成功' ? '#52c41a' : '#ff4d4f' }}>
          {status}
        </span>
      ),
    },
  ];

  // 格式化日志数据为活动数据
  const activityData = recentLogs?.logs?.map((log: any, index: number) => ({
    key: index + 1,
    time: new Date(log.created_at).toLocaleString('zh-CN'),
    user: log.username || `用户${log.user_id}`,
    action: `调用 ${log.model} 模型`,
    status: log.status === 'success' ? '成功' : '失败',
  })) || [];

  if (overviewLoading) {
    return (
      <div style={{ textAlign: 'center', padding: '100px 0' }}>
        <Spin size="large" />
      </div>
    );
  }

  return (
    <div>
      {/* 统计卡片行 */}
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="总用户数"
              value={overview?.total_users || 0}
              prefix={<UserOutlined />}
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="活跃用户"
              value={overview?.active_users || 0}
              prefix={<TeamOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="总请求数"
              value={overview?.total_requests || 0}
              prefix={<ApiOutlined />}
              valueStyle={{ color: '#cf1322' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="今日请求"
              value={overview?.today_requests || 0}
              prefix={<RiseOutlined />}
              valueStyle={{ color: '#722ed1' }}
            />
          </Card>
        </Col>
      </Row>

      {/* 图表行 */}
      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} lg={12}>
          <Card>
            {trendLoading ? (
              <div style={{ textAlign: 'center', padding: '50px 0' }}>
                <Spin />
              </div>
            ) : (
              <ReactECharts option={trendChartOption} style={{ height: 350 }} />
            )}
          </Card>
        </Col>
        <Col xs={24} lg={12}>
          <Card>
            {modelLoading ? (
              <div style={{ textAlign: 'center', padding: '50px 0' }}>
                <Spin />
              </div>
            ) : (
              <ReactECharts option={modelChartOption} style={{ height: 350 }} />
            )}
          </Card>
        </Col>
      </Row>

      {/* 最近活动 */}
      <Row style={{ marginTop: 16 }}>
        <Col span={24}>
          <Card title={<Title level={4}>最近活动</Title>}>
            <Table
              columns={activityColumns}
              dataSource={activityData}
              pagination={false}
              size="small"
            />
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default DashboardPage;
