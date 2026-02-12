import React from 'react';
import { Card, Row, Col, Statistic, Table, Skeleton, Empty } from 'antd';
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
import { formatNumber } from '../utils/format';

// 统计卡片颜色配置
const statCards = [
  { key: 'total_users', title: '总用户数', icon: <UserOutlined />, color: '#1677ff', bg: '#e6f4ff' },
  { key: 'active_users', title: '活跃用户', icon: <TeamOutlined />, color: '#52c41a', bg: '#f6ffed' },
  { key: 'total_requests', title: '总请求数', icon: <ApiOutlined />, color: '#722ed1', bg: '#f9f0ff' },
  { key: 'today_requests', title: '今日请求', icon: <RiseOutlined />, color: '#fa8c16', bg: '#fff7e6' },
] as const;

const DashboardPage: React.FC = () => {
  // 数据查询
  const { data: overview, isLoading: overviewLoading } = useQuery({
    queryKey: ['stats', 'overview'],
    queryFn: statsService.getOverview,
  });

  const { data: trendData, isLoading: trendLoading } = useQuery({
    queryKey: ['stats', 'trend'],
    queryFn: () => statsService.getRequestTrend(7),
  });

  const { data: modelUsage, isLoading: modelLoading } = useQuery({
    queryKey: ['stats', 'models'],
    queryFn: statsService.getModelUsage,
  });

  const { data: recentLogs, isLoading: logsLoading } = useQuery({
    queryKey: ['logs', 'recent'],
    queryFn: () => statsService.getRecentLogs(5),
  });

  // 请求趋势图
  const trendChartOption: EChartsOption = {
    tooltip: { trigger: 'axis' },
    xAxis: {
      type: 'category',
      data: trendData?.map((item) => item.date) || [],
      axisLine: { lineStyle: { color: '#e5e7eb' } },
      axisLabel: { color: '#6b7280' },
    },
    yAxis: {
      type: 'value',
      name: '请求数',
      splitLine: { lineStyle: { type: 'dashed', color: '#f0f0f0' } },
      axisLabel: { color: '#6b7280' },
    },
    series: [
      {
        name: '请求数',
        type: 'line',
        smooth: true,
        symbol: 'circle',
        symbolSize: 6,
        data: trendData?.map((item) => item.count) || [],
        areaStyle: {
          color: {
            type: 'linear',
            x: 0, y: 0, x2: 0, y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(22, 119, 255, 0.15)' },
              { offset: 1, color: 'rgba(22, 119, 255, 0.01)' },
            ],
          },
        },
        lineStyle: { width: 2, color: '#1677ff' },
        itemStyle: { color: '#1677ff' },
      },
    ],
    grid: { left: 12, right: 12, bottom: 0, top: 8, containLabel: true },
  };

  // 模型使用排行
  const modelChartOption: EChartsOption = {
    tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
    xAxis: {
      type: 'value',
      name: '调用次数',
      splitLine: { lineStyle: { type: 'dashed', color: '#f0f0f0' } },
      axisLabel: { color: '#6b7280' },
    },
    yAxis: {
      type: 'category',
      data: modelUsage?.map((item) => item.model).reverse() || [],
      axisLine: { lineStyle: { color: '#e5e7eb' } },
      axisLabel: { color: '#6b7280', fontSize: 12 },
    },
    series: [
      {
        name: '调用次数',
        type: 'bar',
        barMaxWidth: 24,
        data: modelUsage?.map((item) => item.count).reverse() || [],
        itemStyle: {
          borderRadius: [0, 4, 4, 0],
          color: {
            type: 'linear',
            x: 0, y: 0, x2: 1, y2: 0,
            colorStops: [
              { offset: 0, color: '#1677ff' },
              { offset: 1, color: '#69b1ff' },
            ],
          },
        },
      },
    ],
    grid: { left: 12, right: 24, bottom: 0, top: 8, containLabel: true },
  };

  // 最近活动表格
  const activityColumns = [
    { title: '时间', dataIndex: 'time', key: 'time', width: 180 },
    { title: '用户', dataIndex: 'user', key: 'user' },
    { title: '操作', dataIndex: 'action', key: 'action' },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (status: string) => (
        <span style={{ color: status === '成功' ? '#52c41a' : '#ff4d4f', fontWeight: 500 }}>
          {status}
        </span>
      ),
    },
  ];

  const activityData =
    recentLogs?.logs?.map((log: any, index: number) => ({
      key: index + 1,
      time: new Date(log.created_at).toLocaleString('zh-CN'),
      user: log.username || `用户${log.user_id}`,
      action: `调用 ${log.model}`,
      status: log.status === 'success' ? '成功' : '失败',
    })) || [];

  return (
    <div>
      {/* 页面头部 */}
      <div className="page-header">
        <div className="page-title">统计概览</div>
        <div className="page-desc">查看系统运行状态和关键指标</div>
      </div>

      {/* 统计卡片 */}
      <Row gutter={[16, 16]}>
        {statCards.map((item) => (
          <Col xs={24} sm={12} lg={6} key={item.key}>
            <Card className="stat-card" styles={{ body: { padding: '20px 24px' } }}>
              {overviewLoading ? (
                <Skeleton active paragraph={{ rows: 1 }} title={false} />
              ) : (
                <Statistic
                  title={<span style={{ color: 'rgba(0,0,0,0.45)', fontSize: 14 }}>{item.title}</span>}
                  value={item.key === 'total_requests' || item.key === 'today_requests'
                    ? formatNumber((overview as any)?.[item.key] || 0)
                    : (overview as any)?.[item.key] || 0
                  }
                  prefix={
                    <span
                      style={{
                        display: 'inline-flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        width: 40,
                        height: 40,
                        borderRadius: 8,
                        background: item.bg,
                        color: item.color,
                        fontSize: 18,
                        marginRight: 12,
                      }}
                    >
                      {item.icon}
                    </span>
                  }
                  valueStyle={{ color: 'rgba(0,0,0,0.88)', fontSize: 28, fontWeight: 600 }}
                />
              )}
            </Card>
          </Col>
        ))}
      </Row>

      {/* 图表 */}
      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} lg={12}>
          <Card
            title="请求趋势"
            styles={{ header: { borderBottom: '1px solid #f0f0f0' }, body: { padding: '12px 16px' } }}
          >
            {trendLoading ? (
              <Skeleton active paragraph={{ rows: 6 }} title={false} />
            ) : trendData?.length ? (
              <ReactECharts option={trendChartOption} style={{ height: 320 }} />
            ) : (
              <Empty description="暂无数据" style={{ padding: '60px 0' }} />
            )}
          </Card>
        </Col>
        <Col xs={24} lg={12}>
          <Card
            title="模型使用排行"
            styles={{ header: { borderBottom: '1px solid #f0f0f0' }, body: { padding: '12px 16px' } }}
          >
            {modelLoading ? (
              <Skeleton active paragraph={{ rows: 6 }} title={false} />
            ) : modelUsage?.length ? (
              <ReactECharts option={modelChartOption} style={{ height: 320 }} />
            ) : (
              <Empty description="暂无数据" style={{ padding: '60px 0' }} />
            )}
          </Card>
        </Col>
      </Row>

      {/* 最近活动 */}
      <Card
        title="最近活动"
        style={{ marginTop: 16 }}
        styles={{ header: { borderBottom: '1px solid #f0f0f0' } }}
      >
        <Table
          columns={activityColumns}
          dataSource={activityData}
          pagination={false}
          loading={logsLoading}
          size="middle"
          locale={{ emptyText: <Empty description="暂无活动记录" /> }}
        />
      </Card>
    </div>
  );
};

export default DashboardPage;
