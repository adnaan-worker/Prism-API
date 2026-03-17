import React from 'react';
import { Row, Col, Table, Skeleton, Empty, Card } from 'antd';
import {
  UserOutlined,
  TeamOutlined,
  ApiOutlined,
  RiseOutlined,
  ArrowUpOutlined,
  ArrowDownOutlined,
} from '@ant-design/icons';
import { useQuery } from '@tanstack/react-query';
import ReactECharts from 'echarts-for-react';
import { statsService, type StatsOverview, type RequestTrend, type ModelUsage } from '../services/statsService';
import type { EChartsOption } from 'echarts';
import { formatNumber } from '../utils/format';
import PageContainer from '../components/PageContainer';

// 统计卡片配置
const statCards = [
  { 
    key: 'total_users', 
    title: '总用户数', 
    icon: <UserOutlined />, 
    gradient: 'from-blue-500 to-cyan-500',
    iconBg: 'bg-blue-500/10',
    iconColor: 'text-blue-400',
  },
  { 
    key: 'active_users', 
    title: '活跃用户', 
    icon: <TeamOutlined />, 
    gradient: 'from-green-500 to-emerald-500',
    iconBg: 'bg-green-500/10',
    iconColor: 'text-green-400',
  },
  { 
    key: 'total_requests', 
    title: '总请求数', 
    icon: <ApiOutlined />, 
    gradient: 'from-purple-500 to-pink-500',
    iconBg: 'bg-purple-500/10',
    iconColor: 'text-purple-400',
  },
  { 
    key: 'today_requests', 
    title: '今日请求', 
    icon: <RiseOutlined />, 
    gradient: 'from-amber-500 to-orange-500',
    iconBg: 'bg-amber-500/10',
    iconColor: 'text-amber-400',
  },
] as const;

const DashboardPage: React.FC = () => {
  // 数据查询
  const { data: overview, isLoading: overviewLoading } = useQuery<StatsOverview>({
    queryKey: ['stats', 'overview'],
    queryFn: () => statsService.getOverview(),
    refetchInterval: 30000, // 每30秒刷新一次
  });

  const { data: trendData, isLoading: trendLoading } = useQuery<RequestTrend[]>({
    queryKey: ['stats', 'trend'],
    queryFn: () => statsService.getRequestTrend(7),
    refetchInterval: 60000, // 每分钟刷新一次
  });

  const { data: modelUsage, isLoading: modelLoading } = useQuery<ModelUsage[]>({
    queryKey: ['stats', 'models'],
    queryFn: () => statsService.getModelUsage(),
    refetchInterval: 60000,
  });

  const { data: recentLogs, isLoading: logsLoading } = useQuery({
    queryKey: ['logs', 'recent'],
    queryFn: () => statsService.getRecentLogs(5),
    refetchInterval: 30000,
  });

  // 请求趋势图
  const trendChartOption: EChartsOption = {
    backgroundColor: 'transparent',
    tooltip: { 
      trigger: 'axis',
      backgroundColor: 'rgba(17, 17, 17, 0.95)',
      borderColor: 'rgba(255, 255, 255, 0.1)',
      borderWidth: 1,
      textStyle: { color: '#fff' },
      padding: [12, 16],
      formatter: (params: any) => {
        const data = params[0];
        return `<div style="font-weight: 600; margin-bottom: 4px;">${data.axisValue}</div>
                <div style="color: #0ea5e9;">请求数: ${formatNumber(data.value)}</div>`;
      }
    },
    xAxis: {
      type: 'category',
      data: trendData?.map((item) => item.date) || [],
      axisLine: { lineStyle: { color: '#333' } },
      axisLabel: { color: '#a1a1aa', fontSize: 12 },
      axisTick: { show: false },
    },
    yAxis: {
      type: 'value',
      name: '请求数',
      nameTextStyle: { color: '#a1a1aa', fontSize: 12 },
      splitLine: { lineStyle: { type: 'dashed', color: 'rgba(255,255,255,0.05)' } },
      axisLabel: { 
        color: '#a1a1aa',
        fontSize: 12,
        formatter: (value: number) => {
          if (value >= 1000000) return (value / 1000000).toFixed(1) + 'M';
          if (value >= 1000) return (value / 1000).toFixed(1) + 'K';
          return value.toString();
        }
      },
      axisLine: { show: false },
      axisTick: { show: false },
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
              { offset: 0, color: 'rgba(14, 165, 233, 0.25)' },
              { offset: 1, color: 'rgba(14, 165, 233, 0)' },
            ],
          },
        },
        lineStyle: { width: 2.5, color: '#0ea5e9' },
        itemStyle: { 
          color: '#0ea5e9',
          borderColor: '#0a0a0a',
          borderWidth: 2,
        },
        emphasis: {
          itemStyle: {
            color: '#38bdf8',
            borderColor: '#0ea5e9',
            borderWidth: 2,
            shadowBlur: 10,
            shadowColor: 'rgba(14, 165, 233, 0.5)',
          }
        }
      },
    ],
    grid: { left: 16, right: 16, bottom: 8, top: 40, containLabel: true },
  };

  // 模型使用排行
  const modelChartOption: EChartsOption = {
    backgroundColor: 'transparent',
    tooltip: { 
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
      backgroundColor: 'rgba(17, 17, 17, 0.95)',
      borderColor: 'rgba(255, 255, 255, 0.1)',
      borderWidth: 1,
      textStyle: { color: '#fff' },
      padding: [12, 16],
      formatter: (params: any) => {
        const data = params[0];
        return `<div style="font-weight: 600; margin-bottom: 4px;">${data.name}</div>
                <div style="color: #0ea5e9;">调用次数: ${formatNumber(data.value)}</div>`;
      }
    },
    xAxis: {
      type: 'value',
      name: '调用次数',
      nameTextStyle: { color: '#a1a1aa', fontSize: 12 },
      splitLine: { lineStyle: { type: 'dashed', color: 'rgba(255,255,255,0.05)' } },
      axisLabel: { 
        color: '#a1a1aa',
        fontSize: 12,
        formatter: (value: number) => {
          if (value >= 1000000) return (value / 1000000).toFixed(1) + 'M';
          if (value >= 1000) return (value / 1000).toFixed(1) + 'K';
          return value.toString();
        }
      },
      axisLine: { show: false },
      axisTick: { show: false },
    },
    yAxis: {
      type: 'category',
      data: modelUsage?.map((item) => item.model).reverse() || [],
      axisLine: { show: false },
      axisTick: { show: false },
      axisLabel: { 
        color: '#a1a1aa',
        fontSize: 12,
        formatter: (value: string) => {
          return value.length > 20 ? value.substring(0, 20) + '...' : value;
        }
      },
    },
    series: [
      {
        name: '调用次数',
        type: 'bar',
        barMaxWidth: 24,
        data: modelUsage?.map((item) => item.count).reverse() || [],
        itemStyle: {
          borderRadius: [0, 6, 6, 0],
          color: {
            type: 'linear',
            x: 0, y: 0, x2: 1, y2: 0,
            colorStops: [
              { offset: 0, color: '#0ea5e9' },
              { offset: 1, color: '#6366f1' },
            ],
          },
        },
        emphasis: {
          itemStyle: {
            color: {
              type: 'linear',
              x: 0, y: 0, x2: 1, y2: 0,
              colorStops: [
                { offset: 0, color: '#38bdf8' },
                { offset: 1, color: '#818cf8' },
              ],
            },
          }
        }
      },
    ],
    grid: { left: 16, right: 24, bottom: 8, top: 40, containLabel: true },
  };

  // 最近活动表格
  const activityColumns = [
    { 
      title: '时间',
      dataIndex: 'time',
      key: 'time',
      width: 180,
      render: (time: string) => (
        <span className="text-text-secondary text-sm">{time}</span>
      )
    },
    { 
      title: '用户',
      dataIndex: 'user',
      key: 'user',
      render: (user: string) => (
        <span className="text-text-primary font-medium">{user}</span>
      )
    },
    { 
      title: '操作',
      dataIndex: 'action',
      key: 'action',
      render: (action: string) => (
        <span className="text-text-secondary">{action}</span>
      )
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => (
        <span className={`inline-flex items-center px-2.5 py-1 rounded-md text-xs font-medium ${status === '成功'
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
    <PageContainer
      title="统计概览"
      description="实时监控系统运行状态和关键指标"
    >
      {/* 统计卡片 */}
      <Row gutter={[16, 16]} className="mb-6">
        {statCards.map((card) => (
          <Col key={card.key} xs={24} sm={12} lg={6}>
            {overviewLoading ? (
              <div className="glass-card p-6 h-[140px] animate-pulse">
                <div className="flex items-start justify-between mb-4">
                  <div className="w-12 h-12 rounded-xl bg-white/5"></div>
                </div>
                <div className="h-8 bg-white/5 rounded mb-2 w-24"></div>
                <div className="h-4 bg-white/5 rounded w-20"></div>
              </div>
            ) : (
              <div className="glass-card p-6 cursor-pointer hover:scale-[1.02] transition-all group">
                <div className="flex items-start justify-between mb-4">
                  <div className={`w-12 h-12 rounded-xl flex items-center justify-center text-xl ${card.iconBg} ${card.iconColor} group-hover:scale-110 transition-transform`}>
                    {card.icon}
                  </div>
                </div>
                <div className="text-3xl font-bold text-white mb-1 tracking-tight">
                  {formatNumber(overview?.[card.key as keyof StatsOverview] || 0)}
                </div>
                <div className="text-sm text-text-secondary font-medium">{card.title}</div>
              </div>
            )}
          </Col>
        ))}
      </Row>

      {/* 图表区域 */}
      <Row gutter={[16, 16]} className="mb-6">
        {/* 请求趋势 */}
        <Col xs={24} lg={16}>
          <div className="glass-card p-6">
            <div className="flex items-center justify-between mb-4">
              <div>
                <h3 className="text-lg font-semibold text-white mb-1">请求趋势</h3>
                <p className="text-sm text-text-secondary">最近 7 天的请求量变化</p>
              </div>
            </div>
            {trendLoading ? (
              <div className="h-[300px] flex items-center justify-center">
                <div className="animate-spin rounded-full h-8 w-8 border-2 border-primary border-t-transparent"></div>
              </div>
            ) : trendData && trendData.length > 0 ? (
              <ReactECharts option={trendChartOption} style={{ height: '300px' }} />
            ) : (
              <div className="h-[300px] flex items-center justify-center">
                <Empty description="暂无数据" />
              </div>
            )}
          </div>
        </Col>

        {/* 模型使用排行 */}
        <Col xs={24} lg={8}>
          <div className="glass-card p-6">
            <div className="flex items-center justify-between mb-4">
              <div>
                <h3 className="text-lg font-semibold text-white mb-1">模型排行</h3>
                <p className="text-sm text-text-secondary">调用次数 Top 10</p>
              </div>
            </div>
            {modelLoading ? (
              <div className="h-[300px] flex items-center justify-center">
                <div className="animate-spin rounded-full h-8 w-8 border-2 border-primary border-t-transparent"></div>
              </div>
            ) : modelUsage && modelUsage.length > 0 ? (
              <ReactECharts option={modelChartOption} style={{ height: '300px' }} />
            ) : (
              <div className="h-[300px] flex items-center justify-center">
                <Empty description="暂无数据" />
              </div>
            )}
          </div>
        </Col>
      </Row>

      {/* 最近活动 */}
      <div className="glass-card overflow-hidden">
        <div className="p-6 border-b border-border/40">
          <h3 className="text-lg font-semibold text-white">最近活动</h3>
          <p className="text-sm text-text-secondary mt-1">最近 5 条 API 调用记录</p>
        </div>
        {logsLoading ? (
          <div className="p-8 flex items-center justify-center">
            <div className="animate-spin rounded-full h-8 w-8 border-2 border-primary border-t-transparent"></div>
          </div>
        ) : (
          <Table
            columns={activityColumns}
            dataSource={activityData}
            pagination={false}
            rowClassName="hover:bg-white/5 transition-colors cursor-pointer"
            size="middle"
            locale={{ emptyText: <Empty description={<span className="text-text-secondary">暂无活动记录</span>} /> }}
          />
        )}
      </div>
    </PageContainer>
  );
};

export default DashboardPage;
