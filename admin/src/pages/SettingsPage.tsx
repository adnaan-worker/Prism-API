import React from 'react';
import { Card, Typography } from 'antd';

const { Title, Paragraph } = Typography;

const SettingsPage: React.FC = () => {
  return (
    <Card>
      <Title level={3}>系统设置</Title>
      <Paragraph>
        系统配置和管理功能（待实现）
      </Paragraph>
    </Card>
  );
};

export default SettingsPage;
