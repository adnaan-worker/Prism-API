import {
  Tag,
  Button,
  message,
  Avatar
} from 'antd';
import {
  UserOutlined,
  MailOutlined,
  CalendarOutlined,
  ClockCircleOutlined,
  CheckCircleOutlined,
  CopyOutlined,
  EditOutlined,
  SafetyCertificateOutlined
} from '@ant-design/icons';
import { useQuery } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { useOutletContext } from 'react-router-dom';
import { authService } from '../../services/authService';
import { quotaService } from '../../services/quotaService';
import { Column } from '@ant-design/charts';

const ProfilePage = () => {
  const { t } = useTranslation();
  const { isDarkMode } = useOutletContext<{ isDarkMode: boolean }>();

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

  // Fetch usage history (7 days)
  const { data: usageHistory, isLoading: historyLoading } = useQuery({
    queryKey: ['usageHistory', 7],
    queryFn: () => quotaService.getUsageHistory(7),
  });

  // Calculate usage percentage
  const usagePercentage = quotaInfo
    ? Math.round((quotaInfo.used_quota / quotaInfo.total_quota) * 100)
    : 0;

  // Format date
  const formatDate = (dateString?: string) => {
    if (!dateString) return 'Never';
    return new Date(dateString).toLocaleString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  // Transform usage history data for chart
  const usageHistoryData = usageHistory?.map((item) => {
    const date = new Date(item.date);
    const dayNames = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
    return {
      date: dayNames[date.getDay()],
      tokens: item.tokens,
    };
  }) || [];

  const chartConfig = {
    data: usageHistoryData,
    xField: 'date',
    yField: 'tokens',
    color: '#0ea5e9',
    columnStyle: {
      radius: [4, 4, 0, 0],
      fillOpacity: 0.8,
    },
    label: {
      position: 'top' as const,
      style: {
        fill: isDarkMode ? '#a1a1aa' : '#4b5563', // gray-600 for light mode
        opacity: 0.8,
      },
      formatter: (v: any) => (v.tokens / 1000).toFixed(1) + 'k',
    },
    xAxis: {
      label: { style: { fill: isDarkMode ? '#a1a1aa' : '#4b5563' } },
      line: { style: { stroke: 'transparent' } },
    },
    yAxis: {
      label: { style: { fill: isDarkMode ? '#a1a1aa' : '#4b5563' }, formatter: (v: string) => (Number(v) / 1000) + 'k' },
      grid: { line: { style: { stroke: isDarkMode ? '#333' : '#e5e7eb', lineDash: [4, 4] } } },
    },
    tooltip: {
      formatter: (datum: any) => {
        return { name: 'Usage', value: datum.tokens + ' tokens' };
      },
    },
    theme: isDarkMode ? 'dark' : 'light',
  };

  const InfoItem = ({ icon, label, value, copyable }: any) => (
    <div className="flex items-center justify-between p-4 rounded-xl bg-slate-50 dark:bg-white/5 border border-slate-100 dark:border-white/5 hover:bg-slate-100 dark:hover:bg-white/10 transition-colors">
      <div className="flex items-center gap-3">
        <div className="text-slate-400 dark:text-gray-500">{icon}</div>
        <div>
          <p className="text-xs text-slate-500 dark:text-gray-500">{label}</p>
          <p className="text-sm text-slate-900 dark:text-white font-medium">{value || '-'}</p>
        </div>
      </div>
      {/* ... */}
    </div>
  );

  return (
    <div className="max-w-7xl mx-auto space-y-8 animate-fade-in pb-10">

      {/* Profile Banner */}
      <div className="relative rounded-3xl overflow-hidden glass-card">
        {/* Banner Background */}
        <div className="h-48 bg-gradient-to-r from-primary-900 via-primary-800 to-black relative">
          <div className="absolute inset-0 bg-[url('/grid-pattern.svg')] opacity-20"></div>
          <div className="absolute inset-0 bg-gradient-to-t from-black/80 to-transparent"></div>
        </div>

        {/* Profile Content */}
        <div className="relative px-8 pb-8 -mt-16 flex flex-col md:flex-row items-end md:items-center gap-6">
          <Avatar
            size={128}
            icon={<UserOutlined />}
            className="ring-4 ring-white dark:ring-black bg-primary text-white text-4xl shadow-2xl"
          />
          <div className="flex-1 mb-2">
            <div className="flex items-center gap-3 mb-1">
              <h1 className="text-3xl font-bold text-slate-900 dark:text-white md:text-white">{user?.username}</h1>
              {user?.is_admin && (
                <Tag color="gold" className="border-none px-2 py-0.5 rounded bg-yellow-500/20 text-yellow-500">
                  <SafetyCertificateOutlined className="mr-1" /> {t('profile.role.admin')}
                </Tag>
              )}
              <Tag color={user?.status === 'active' ? 'success' : 'default'} className="border-none px-2 py-0.5 rounded bg-green-500/20 text-green-500">
                {user?.status === 'active' ? t('profile.status.active') : user?.status}
              </Tag>
            </div>
            <p className="text-slate-500 dark:text-gray-400 flex items-center gap-2">
              <MailOutlined /> {user?.email}
            </p>
          </div>
          <div className="flex gap-3 mb-2">
            <Button icon={<EditOutlined />} className="bg-white/10 border-white/20 text-white hover:bg-white/20">
              {t('profile.edit')}
            </Button>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">

        {/* Left Column: Stats & Info */}
        <div className="lg:col-span-1 space-y-6">
          {/* Account Details */}
          <div className="glass-card p-6 rounded-2xl">
            <h3 className="text-lg font-bold text-slate-900 dark:text-white mb-4">{t('profile.accountDetails')}</h3>
            <div className="space-y-3">
              <InfoItem icon={<UserOutlined />} label={t('profile.userId')} value={user?.id} copyable />
              <InfoItem icon={<CalendarOutlined />} label={t('profile.joined')} value={formatDate(user?.created_at)} />
              <InfoItem icon={<ClockCircleOutlined />} label={t('profile.lastLogin')} value={formatDate(user?.last_sign_in)} />
            </div>
          </div>

          {/* Quota Ring */}
          <div className="glass-card p-6 rounded-2xl text-center">
            <h3 className="text-lg font-bold text-slate-900 dark:text-white mb-6">{t('profile.quotaUsage')}</h3>
            <div className="relative w-48 h-48 mx-auto mb-6 flex items-center justify-center">
              <svg className="w-full h-full transform -rotate-90">
                <circle cx="96" cy="96" r="88" stroke={isDarkMode ? "#1f1f1f" : "#e2e8f0"} strokeWidth="12" fill="transparent" />
                <circle
                  cx="96"
                  cy="96"
                  r="88"
                  stroke="#0ea5e9"
                  strokeWidth="12"
                  fill="transparent"
                  strokeDasharray={2 * Math.PI * 88}
                  strokeDashoffset={2 * Math.PI * 88 * (1 - usagePercentage / 100)}
                  strokeLinecap="round"
                />
              </svg>
              <div className="absolute inset-0 flex flex-col items-center justify-center">
                <span className="text-4xl font-bold text-slate-900 dark:text-white">{usagePercentage}%</span>
                <span className="text-xs text-slate-500 dark:text-gray-500">{t('profile.quota.used')}</span>
              </div>
            </div>
            <div className="grid grid-cols-2 gap-4 text-left">
              <div className="p-3 rounded-lg bg-slate-50 dark:bg-white/5 border border-slate-100 dark:border-transparent">
                <p className="text-xs text-slate-500 dark:text-gray-500">{t('profile.quota.remaining')}</p>
                <p className="text-lg font-bold text-emerald-500">{quotaInfo?.remaining_quota.toLocaleString()}</p>
              </div>
              <div className="p-3 rounded-lg bg-slate-50 dark:bg-white/5 border border-slate-100 dark:border-transparent">
                <p className="text-xs text-slate-500 dark:text-gray-500">{t('profile.quota.total')}</p>
                <p className="text-lg font-bold text-blue-500">{quotaInfo?.total_quota.toLocaleString()}</p>
              </div>
            </div>
          </div >
        </div >

        {/* Right Column: Usage History */}
        < div className="lg:col-span-2 space-y-6" >
          <div className="glass-card p-6 rounded-2xl h-[400px]">
            <div className="flex items-center justify-between mb-6">
              <h3 className="text-lg font-bold text-slate-900 dark:text-white">{t('profile.activity')}</h3>
              <div className="flex gap-2">
                <Tag className="bg-primary/10 text-primary border-none px-3 py-1 scale-105">Tokens Consumed</Tag>
              </div>
            </div>
            {historyLoading ? (
              <div className="h-full flex items-center justify-center text-text-tertiary">{t('common.loading')}</div>
            ) : (
              <div className="h-[300px] w-full">
                <Column {...chartConfig} />
              </div>
            )}
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <div className="glass-card p-5 rounded-xl">
              <p className="text-sm text-slate-500 dark:text-gray-500 mb-1">{t('profile.stats.total')}</p>
              <p className="text-2xl font-bold text-slate-900 dark:text-white">
                {usageHistoryData.reduce((sum, item) => sum + item.tokens, 0).toLocaleString()}
              </p>
            </div>
            <div className="glass-card p-5 rounded-xl">
              <p className="text-sm text-slate-500 dark:text-gray-500 mb-1">{t('profile.stats.dailyAvg')}</p>
              <p className="text-2xl font-bold text-slate-900 dark:text-white">
                {Math.round(usageHistoryData.reduce((sum, item) => sum + item.tokens, 0) / 7).toLocaleString()}
              </p>
            </div>
            <div className="glass-card p-5 rounded-xl">
              <p className="text-sm text-slate-500 dark:text-gray-500 mb-1">{t('profile.stats.peak')}</p>
              <p className="text-2xl font-bold text-slate-900 dark:text-white">
                {Math.max(0, ...usageHistoryData.map((item) => item.tokens)).toLocaleString()}
              </p>
            </div>
          </div>
        </div >
      </div >
    </div >
  );
};

export default ProfilePage;
