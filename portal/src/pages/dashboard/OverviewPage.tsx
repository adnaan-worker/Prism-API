import { useState, useEffect } from 'react';
import {
  Button,
  message,
  Alert
} from 'antd';
import {
  GiftOutlined,
  ThunderboltOutlined,
  CheckCircleOutlined,
  ApiOutlined,
  CodeOutlined,
  RocketOutlined,
  ArrowRightOutlined,
  FireOutlined
} from '@ant-design/icons';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { useOutletContext, useNavigate } from 'react-router-dom';
import { quotaService } from '../../services/quotaService';
import { Line } from '@ant-design/charts';

const OverviewPage = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { isDarkMode } = useOutletContext<{ isDarkMode: boolean }>();
  const [hasSignedInToday, setHasSignedInToday] = useState(false);

  // Fetch quota info
  const { data: quotaInfo, isLoading } = useQuery({
    queryKey: ['quotaInfo'],
    queryFn: quotaService.getQuotaInfo,
    refetchInterval: 30000,
  });

  // Fetch usage history (7 days)
  const { data: usageHistory, isLoading: historyLoading } = useQuery({
    queryKey: ['usageHistory', 7],
    queryFn: () => quotaService.getUsageHistory(7),
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
      message.success(t('common.success') + '! ' + t('common.tokens') + ': ' + data.quota_awarded);
      setHasSignedInToday(true);
      queryClient.invalidateQueries({ queryKey: ['quotaInfo'] });
    },
    onError: (error: any) => {
      if (error.response?.data?.error?.code === 409002) {
        message.warning(t('dashboard.status.degraded')); // fallback/placeholder message or specific translation needed
        setHasSignedInToday(true);
      } else {
        message.error(t('common.error'));
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

  // Transform usage history data for chart
  const usageTrendData = usageHistory?.map((item) => {
    const date = new Date(item.date);
    const dayNames = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
    return {
      date: dayNames[date.getDay()],
      usage: item.tokens,
    };
  }) || [];

  const chartConfig = {
    data: usageTrendData,
    xField: 'date',
    yField: 'usage',
    smooth: true,
    color: '#0ea5e9', // Primary Sky 500
    point: {
      size: 0,
    },
    lineStyle: {
      lineWidth: 3,
      shadowColor: 'rgba(14, 165, 233, 0.5)',
      shadowBlur: 10,
    },
    area: {
      style: {
        fill: 'l(270) 0:rgba(14, 165, 233, 0.2) 1:rgba(14, 165, 233, 0)',
      },
    },
    xAxis: {
      grid: { line: { style: { stroke: isDarkMode ? '#333' : '#e5e7eb', lineDash: [4, 4] } } },
      line: { style: { stroke: 'transparent' } },
      label: { style: { fill: isDarkMode ? '#666' : '#9ca3af' } },
    },
    yAxis: {
      grid: { line: { style: { stroke: isDarkMode ? '#333' : '#e5e7eb', lineDash: [4, 4] } } },
      label: { style: { fill: isDarkMode ? '#666' : '#9ca3af' }, formatter: (v: string) => (Number(v) / 1000) + 'k' },
    },
    tooltip: { showMarkers: true },
    theme: isDarkMode ? 'dark' : 'light',
  };

  const StatCard = ({ title, value, icon, color, subValue }: any) => (
    <div className="glass-card p-6 rounded-2xl relative overflow-hidden group hover:bg-white hover:shadow-xl dark:hover:bg-white/5 transition-all duration-300">
      <div className={'absolute top-0 right-0 p-4 opacity-20 group-hover:opacity-40 transition-opacity ' + color}>
        {icon}
      </div>
      <div className="relative z-10">
        <p className="text-slate-600 dark:text-gray-400 text-sm font-medium mb-1">{title}</p>
        <h3 className="text-2xl font-bold text-slate-900 dark:text-white tracking-tight">{value}</h3>
        {subValue && <p className="text-slate-400 dark:text-gray-500 text-xs mt-1">{subValue}</p>}
      </div>
      <div className={'absolute bottom-0 left-0 h-1 w-full bg-gradient-to-r opacity-50 ' + color}></div>
    </div>
  );

  return (
    <div className="max-w-7xl mx-auto space-y-8 animate-fade-in">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold text-slate-900 dark:text-white mb-2">{t('dashboard.welcome')}</h1>
          <p className="text-slate-600 dark:text-gray-400">{t('dashboard.welcomeSub')}</p>
        </div>
        <div className="flex gap-3">
          <Button
            type="primary"
            size="large"
            icon={<GiftOutlined />}
            onClick={handleSignIn}
            loading={signInMutation.isPending}
            disabled={hasSignedInToday}
            className={`
              h-12 px-6 rounded-xl border-none shadow-lg
              ${hasSignedInToday
                ? 'bg-black/10 dark:bg-white/10 text-text-tertiary cursor-not-allowed'
                : 'bg-gradient-to-r from-primary-600 to-primary-500 hover:scale-105 transition-transform text-white'
              }
            `}
          >
            {hasSignedInToday ? t('common.success') : 'Daily Check-in'}
          </Button>
          <Button
            size="large"
            className="h-12 px-6 rounded-xl bg-white/5 border border-black/10 dark:border-white/10 text-text-primary hover:bg-black/5 dark:hover:bg-white/10"
            icon={<RocketOutlined />}
          >
            {t('dashboard.upgrade')}
          </Button>
        </div>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <StatCard
          title={t('dashboard.quota.total')}
          value={isLoading ? '-' : quotaInfo?.total_quota.toLocaleString()}
          icon={<ThunderboltOutlined style={{ fontSize: 40 }} />}
          color="from-blue-500 to-cyan-500"
          subValue={t('dashboard.quota.lifetime')}
        />
        <StatCard
          title={t('dashboard.quota.used')}
          value={isLoading ? '-' : quotaInfo?.used_quota.toLocaleString()}
          icon={<ApiOutlined style={{ fontSize: 40 }} />}
          color="from-orange-500 to-red-500"
          subValue={`${usagePercentage}${t('dashboard.quota.consumed')}`}
        />
        <StatCard
          title={t('dashboard.quota.remaining')}
          value={isLoading ? '-' : quotaInfo?.remaining_quota.toLocaleString()}
          icon={<CheckCircleOutlined style={{ fontSize: 40 }} />}
          color="from-green-500 to-emerald-500"
          subValue={t('dashboard.quota.available')}
        />
        <div className="glass-card p-6 rounded-2xl flex flex-col justify-center items-center relative overflow-hidden">
          <div className="absolute inset-0 bg-gradient-to-br from-primary-900/20 to-transparent"></div>
          <FireOutlined className="text-3xl text-primary mb-2 animate-pulse" />
          <p className="text-text-secondary text-sm">{t('dashboard.status.title')}</p>
          <p className="text-xl font-bold text-green-500 flex items-center gap-2">
            <span className="w-2 h-2 rounded-full bg-green-500 animate-pulse"></span>
            {t('dashboard.status.operational')}
          </p>
        </div>
      </div>

      {/* Main Content Split */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Left: Usage Chart */}
        <div className="lg:col-span-2 space-y-6">
          <div className="glass-card p-6 rounded-2xl h-[400px]">
            <div className="flex items-center justify-between mb-6">
              <h3 className="text-lg font-bold text-text-primary">{t('dashboard.trends.title')}</h3>
              <select className="bg-black/5 dark:bg-black/30 border border-black/10 dark:border-white/10 rounded-lg px-3 py-1 text-xs text-text-secondary outline-none focus:border-primary">
                <option>{t('dashboard.trends.last7Days')}</option>
                <option>{t('dashboard.trends.last30Days')}</option>
              </select>
            </div>
            {historyLoading ? (
              <div className="h-full flex items-center justify-center text-text-tertiary">{t('common.loading')}</div>
            ) : (
              <div className="h-[300px] w-full">
                <Line {...chartConfig} />
              </div>
            )}
          </div>

          <div className="glass-card p-6 rounded-2xl">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-bold text-text-primary">{t('dashboard.quota.used')}</h3>
              <span className="text-primary font-mono">{usagePercentage}%</span>
            </div>
            <div className="h-3 w-full bg-black/10 dark:bg-black/40 rounded-full overflow-hidden mb-2">
              <div
                className={`h-full rounded-full transition-all duration-1000 ${usagePercentage > 90 ? 'bg-red-500' : 'bg-primary'}`}
                style={{ width: `${usagePercentage}%` }}
              ></div>
            </div>
            {usagePercentage > 80 && (
              <div className="mt-4 p-3 bg-orange-500/10 border border-orange-500/20 rounded-lg flex items-start gap-3">
                <Alert message="Quota Warning" type="warning" showIcon className="bg-transparent border-none p-0" />
                <span className="text-orange-500 text-sm">You have used over 80% of your quota.</span>
              </div>
            )}
          </div>
        </div>

        {/* Right: Quick Start & Tips */}
        <div className="space-y-6">
          <div className="glass-card p-6 rounded-2xl">
            <h3 className="text-lg font-bold text-text-primary mb-4 flex items-center gap-2">
              <RocketOutlined className="text-primary" /> {t('dashboard.quickStart.title')}
            </h3>
            <div className="space-y-4">
              {[
                { title: t('dashboard.quickStart.generateKey.title'), desc: t('dashboard.quickStart.generateKey.desc'), icon: <ApiOutlined />, link: "/dashboard/api-keys" },
                { title: t('dashboard.quickStart.exploreModels.title'), desc: t('dashboard.quickStart.exploreModels.desc'), icon: <ThunderboltOutlined />, link: "/dashboard/models" },
                { title: t('dashboard.quickStart.docs.title'), desc: t('dashboard.quickStart.docs.desc'), icon: <CodeOutlined />, link: "/dashboard/docs" },
              ].map((item, idx) => (
                <div
                  key={idx}
                  className="group p-3 rounded-xl hover:bg-slate-100 dark:hover:bg-white/5 border border-transparent hover:border-slate-200 dark:hover:border-white/5 transition-all cursor-pointer flex items-center gap-4"
                  onClick={() => navigate(item.link)}
                >
                  <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center text-primary group-hover:scale-110 transition-transform">
                    {item.icon}
                  </div>
                  <div className="flex-1">
                    <h4 className="text-text-primary font-medium text-sm">{item.title}</h4>
                    <p className="text-text-tertiary text-xs">{item.desc}</p>
                  </div>
                  <ArrowRightOutlined className="text-text-tertiary group-hover:text-text-primary group-hover:translate-x-1 transition-all" />
                </div>
              ))}
            </div >
          </div >

          <div className="glass-card p-6 rounded-2xl bg-gradient-to-br from-primary-900/10 to-transparent border-primary/20">
            <h3 className="text-text-primary font-bold mb-2">Need more power?</h3>
            <p className="text-text-secondary text-sm mb-4">Upgrade to Pro for higher limits and priority support.</p>
            <Button type="primary" block className="bg-primary/20 border-primary text-primary hover:bg-primary hover:text-white">
              View Plans
            </Button>
          </div>
        </div >
      </div >
    </div >
  );
};

export default OverviewPage;
