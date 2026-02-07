import { Button, Card, Space, Typography } from 'antd';
import { CheckCircleOutlined } from '@ant-design/icons';

const { Title, Paragraph, Text } = Typography;

function App() {
  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 p-8">
      <div className="max-w-4xl mx-auto">
        <Card className="shadow-lg">
          <Space direction="vertical" size="large" className="w-full">
            <div className="text-center">
              <Title level={1} className="!mb-2">
                Prism API - User Portal
              </Title>
              <Paragraph className="text-lg text-gray-600">
                Welcome to Prism API
              </Paragraph>
            </div>

            <div className="bg-green-50 border border-green-200 rounded-lg p-6">
              <Title level={3} className="!mb-4 text-green-700">
                <CheckCircleOutlined className="mr-2" />
                项目初始化完成
              </Title>
              <Space direction="vertical" size="middle" className="w-full">
                <div className="flex items-center">
                  <CheckCircleOutlined className="text-green-600 mr-2" />
                  <Text>React 18 + Vite - 已配置</Text>
                </div>
                <div className="flex items-center">
                  <CheckCircleOutlined className="text-green-600 mr-2" />
                  <Text>Ant Design 5 - 已集成</Text>
                </div>
                <div className="flex items-center">
                  <CheckCircleOutlined className="text-green-600 mr-2" />
                  <Text>Tailwind CSS 3 - 已配置</Text>
                </div>
                <div className="flex items-center">
                  <CheckCircleOutlined className="text-green-600 mr-2" />
                  <Text>React Router 6 - 已配置</Text>
                </div>
                <div className="flex items-center">
                  <CheckCircleOutlined className="text-green-600 mr-2" />
                  <Text>TanStack Query 5 - 已配置</Text>
                </div>
              </Space>
            </div>

            <div className="text-center">
              <Button type="primary" size="large">
                开始使用
              </Button>
            </div>
          </Space>
        </Card>
      </div>
    </div>
  );
}

export default App;
