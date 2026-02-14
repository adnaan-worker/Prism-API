import React from 'react';
import { Row, Col, Table, Skeleton, Empty } from 'antd';
import {
  UserOutlined,
  TeamOutlined,
  ApiOutlined,
  RiseOutlined,
} from '@ant-design/icons';
import { useQuery } from '@tanstack/react-query';
import ReactECharts from 'echarts-for-react';
import { statsService, type StatsOverview, type RequestTrend, type ModelUsage } from '../services/statsService';
import type { EChartsOption } from 'echarts';
import { formatNumber } from '../utils/format';

// 统计卡片颜色配置
const statCards = [
  { key: 'total_users', title: '总用户数', icon: <UserOutlined />, color: '#38bdf8', bg: 'rgba(56, 189, 248, 0.1)' },
  { key: 'active_users', title: '活跃用户', icon: <TeamOutlined />, color: '#4ade80', bg: 'rgba(74, 222, 128, 0.1)' },
  { key: 'total_requests', title: '总请求数', icon: <ApiOutlined />, color: '#a78bfa', bg: 'rgba(167, 139, 250, 0.1)' },
  { key: 'today_requests', title: '今日请求', icon: <RiseOutlined />, color: '#fbbf24', bg: 'rgba(251, 191, 36, 0.1)' },
] as const;

const DashboardPage: React.FC = () => {
  // 数据查询
  const { data: overview, isLoading: overviewLoading } = useQuery<StatsOverview>({
    queryKey: ['stats', 'overview'],
    queryFn: () => statsService.getOverview(),
  });

  const { data: trendData, isLoading: trendLoading } = useQuery<RequestTrend[]>({
    queryKey: ['stats', 'trend'],
    queryFn: () => statsService.getRequestTrend(7),
  });

  const { data: modelUsage, isLoading: modelLoading } = useQuery<ModelUsage[]>({
    queryKey: ['stats', 'models'],
    queryFn: () => statsService.getModelUsage(),
  });

  const { data: recentLogs, isLoading: logsLoading } = useQuery({
    queryKey: ['logs', 'recent'],
    queryFn: () => statsService.getRecentLogs(5),
  });

  // 请求趋势图
  const trendChartOption: EChartsOption = {
    backgroundColor: 'transparent',
    tooltip: { trigger: 'axis', backgroundColor: '#1f1f1f', borderColor: '#333', textStyle: { color: '#fff' } },
    xAxis: {
      type: 'category',
      data: trendData?.map((item) => item.date) || [],
      axisLine: { lineStyle: { color: '#333' } },
      axisLabel: { color: '#a1a1aa' },
    },
    yAxis: {
      type: 'value',
      name: '请求数',
      nameTextStyle: { color: '#a1a1aa' },
      splitLine: { lineStyle: { type: 'dashed', color: 'rgba(255,255,255,0.05)' } },
      axisLabel: { color: '#a1a1aa' },
    },
    series: [
      {
        name: '请求数',
        type: 'line',
        smooth: true,
        symbol: 'none',
        data: trendData?.map((item) => item.count) || [],
        areaStyle: {
          color: {
            type: 'linear',
            x: 0, y: 0, x2: 0, y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(14, 165, 233, 0.3)' },
              { offset: 1, color: 'rgba(14, 165, 233, 0.01)' },
            ],
          },
        },
        lineStyle: { width: 3, color: '#0ea5e9' },
        itemStyle: { color: '#0ea5e9' },
      },
    ],
    grid: { left: 10, right: 10, bottom: 0, top: 30, containLabel: true },
  };

  // 模型使用排行
  const modelChartOption: EChartsOption = {
    backgroundColor: 'transparent',
    tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' }, backgroundColor: '#1f1f1f', borderColor: '#333', textStyle: { color: '#fff' } },
    xAxis: {
      type: 'value',
      name: '调用次数',
      nameTextStyle: { color: '#a1a1aa' },
      splitLine: { lineStyle: { type: 'dashed', color: 'rgba(255,255,255,0.05)' } },
      axisLabel: { color: '#a1a1aa' },
    },
    yAxis: {
      type: 'category',
      data: modelUsage?.map((item) => item.model).reverse() || [],
      axisLine: { lineStyle: { color: '#333' } },
      axisLabel: { color: '#a1a1aa', fontSize: 12 },
    },
    series: [
      {
        name: '调用次数',
        type: 'bar',
        barMaxWidth: 20,
        data: modelUsage?.map((item) => item.count).reverse() || [],
        itemStyle: {
          borderRadius: [0, 4, 4, 0],
          color: {
            type: 'linear',
            x: 0, y: 0, x2: 1, y2: 0,
            colorStops: [
              { offset: 0, color: '#0ea5e9' },
              { offset: 1, color: '#6366f1' },
            ],
          },
        },
      },
    ],
    grid: { left: 10, right: 30, bottom: 0, top: 30, containLabel: true },
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
      width: 100,
      render: (status: string) => (
        <span className={`px-2 py-1 rounded-md text-xs font-medium ${status === '成功'
          ? 'bg-green-500/10 text-green-400 border border-green-500/20'
          : 'bg-red-500/10 text-red-400 border border-red-500/20'
          }`}>
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
    <div className="space-y-6">
      {/* 页面头部 */}
      <div>
        <h1 className="text-2xl font-bold text-text-primary">统计概览</h1>
        <p className="text-text-secondary mt-1">查看系统运行状态和关键指标</p>
      </div>

      {/* 统计卡片 */}
      <Row gutter={[16, 16]}>
        {statCards.map((item) => (
          <Col xs={24} sm={12} lg={6} key={item.key}>
            <div className="glass-card p-6 h-full border-l-4" style={{ borderLeftColor: item.color }}>
              {overviewLoading ? (
                <Skeleton active paragraph={{ rows: 1 }} title={false} />
              ) : (
                <div className="flex items-start justify-between">
                  <div>
                    <div className="text-text-secondary text-sm mb-1">{item.title}</div>
                    <div className="text-2xl font-bold text-text-primary">
                      {item.key === 'total_requests' || item.key === 'today_requests'
                        ? formatNumber((overview as any)?.[item.key] || 0)
                        : (overview as any)?.[item.key] || 0
                      }
                    </div>
                  </div>
                  <div className="flex items-center justify-center w-10 h-10 rounded-lg" style={{ backgroundColor: item.bg, color: item.color }}>
                    <div className="text-lg">{item.icon}</div>
                  </div>
                </div>
              )}
            </div>
          </Col>
        ))}
      </Row>

      {/* 图表 */}
      <Row gutter={[16, 16]}>
        <Col xs={24} lg={12}>
          <div className="glass-card p-6 h-full">
            <h3 className="text-lg font-semibold text-text-primary mb-4">请求趋势</h3>
            {trendLoading ? (
              <Skeleton active paragraph={{ rows: 6 }} title={false} />
            ) : trendData?.length ? (
              <ReactECharts option={trendChartOption} style={{ height: 320 }} />
            ) : (
              <Empty description={<span className="text-text-secondary">暂无数据</span>} style={{ padding: '60px 0' }} />
            )}
          </div>
        </Col>
        <Col xs={24} lg={12}>
          <div className="glass-card p-6 h-full">
            <h3 className="text-lg font-semibold text-text-primary mb-4">模型使用排行</h3>
            {modelLoading ? (
              <Skeleton active paragraph={{ rows: 6 }} title={false} />
            ) : modelUsage?.length ? (
              <ReactECharts option={modelChartOption} style={{ height: 320 }} />
            ) : (
              <Empty description={<span className="text-text-secondary">暂无数据</span>} style={{ padding: '60px 0' }} />
            )}
          </div>
        </Col>
      </Row>

      {/* 最近活动 */}
      <div className="glass-card overflow-hidden">
        <div className="p-6 border-b border-white/5">
          <h3 className="text-lg font-semibold text-text-primary">最近活动</h3>
        </div>
        <Table
          columns={activityColumns}
          dataSource={activityData}
          pagination={false}
          loading={logsLoading}
          rowClassName="hover:bg-white/5 transition-colors"
          size="middle"
          locale={{ emptyText: <Empty description={<span className="text-text-secondary">暂无活动记录</span>} /> }}
        />
      </div>
    </div>
  );
};

export default DashboardPage;
