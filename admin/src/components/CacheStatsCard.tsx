import React from 'react';
import { Card, Statistic, Row, Col, Spin } from 'antd';
import {
  ThunderboltOutlined,
  SaveOutlined,
  DatabaseOutlined,
} from '@ant-design/icons';
import { useQuery } from '@tanstack/react-query';
import { cacheService } from '../services/cacheService';
import { formatNumber } from '../utils/format';

/**
 * 缓存统计卡片组件
 * 显示缓存命中次数、节省的 tokens 和缓存条目数
 */
const CacheStatsCard: React.FC = () => {
  const { data: stats, isLoading } = useQuery({
    queryKey: ['cache-stats'],
    queryFn: cacheService.getCacheStats,
    refetchInterval: 30000, // 每 30 秒刷新一次
  });

  if (isLoading) {
    return (
      <Card>
        <Spin />
      </Card>
    );
  }

  // 计算节省的成本（假设 1000 tokens = 1 credit）
  const creditsSaved = stats ? Math.floor(stats.tokens_saved / 1000) : 0;

  return (
    <Card title="缓存统计" bordered={false}>
      <Row gutter={16}>
        <Col xs={24} sm={8}>
          <Statistic
            title="缓存命中次数"
            value={stats?.total_hits || 0}
            prefix={<ThunderboltOutlined />}
            valueStyle={{ color: '#3f8600' }}
          />
        </Col>
        <Col xs={24} sm={8}>
          <Statistic
            title="节省 Tokens"
            value={formatNumber(stats?.tokens_saved || 0)}
            prefix={<SaveOutlined />}
            valueStyle={{ color: '#cf1322' }}
            suffix={`(${formatNumber(creditsSaved)} credits)`}
          />
        </Col>
        <Col xs={24} sm={8}>
          <Statistic
            title="缓存条目"
            value={stats?.cache_entries || 0}
            prefix={<DatabaseOutlined />}
            valueStyle={{ color: '#1890ff' }}
          />
        </Col>
      </Row>
    </Card>
  );
};

export default CacheStatsCard;
