import { useState } from 'react';
import { Layout, Menu, Typography, Card, Button, Space, message, Divider, Tag, Alert } from 'antd';
import {
  CopyOutlined,
  DownloadOutlined,
  CheckOutlined,
  ApiOutlined,
  CodeOutlined,
  RocketOutlined,
  SafetyOutlined,
} from '@ant-design/icons';

const { Sider, Content } = Layout;
const { Title, Paragraph, Text } = Typography;

interface CodeBlockProps {
  code: string;
  language: string;
}

const CodeBlock = ({ code, language }: CodeBlockProps) => {
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(code);
      setCopied(true);
      message.success('代码已复制到剪贴板');
      setTimeout(() => setCopied(false), 2000);
    } catch (error) {
      message.error('复制失败，请手动复制');
    }
  };

  return (
    <div className="relative">
      <div className="flex justify-between items-center mb-2">
        <Tag color="blue">{language}</Tag>
        <Button
          type="text"
          size="small"
          icon={copied ? <CheckOutlined /> : <CopyOutlined />}
          onClick={handleCopy}
        >
          {copied ? '已复制' : '复制代码'}
        </Button>
      </div>
      <pre
        style={{
          background: '#f6f8fa',
          padding: '16px',
          borderRadius: '6px',
          overflow: 'auto',
          fontSize: '13px',
          lineHeight: '1.6',
        }}
      >
        <code>{code}</code>
      </pre>
    </div>
  );
};

const DocsPage = () => {
  const [selectedKey, setSelectedKey] = useState('quick-start');

  const menuItems = [
    {
      key: 'quick-start',
      icon: <RocketOutlined />,
      label: '快速开始',
    },
    {
      key: 'authentication',
      icon: <SafetyOutlined />,
      label: '认证方式',
    },
    {
      key: 'openai-api',
      icon: <ApiOutlined />,
      label: 'OpenAI格式',
    },
    {
      key: 'anthropic-api',
      icon: <ApiOutlined />,
      label: 'Anthropic格式',
    },
    {
      key: 'gemini-api',
      icon: <ApiOutlined />,
      label: 'Gemini格式',
    },
    {
      key: 'sdk',
      icon: <CodeOutlined />,
      label: 'SDK下载',
    },
  ];

  const renderContent = () => {
    switch (selectedKey) {
      case 'quick-start':
        return (
          <>
            <Title level={2}>快速开始</Title>
            <Paragraph>
              欢迎使用 Prism API！本文档将帮助您快速上手，通过统一的接口调用多个AI模型。
            </Paragraph>

            <Title level={3}>1. 获取API密钥</Title>
            <Paragraph>
              首先，您需要在 <a href="/dashboard/api-keys">API密钥管理</a> 页面创建一个API密钥。
              密钥格式为 <Text code>sk-xxxxxxxxxxxxxxxx</Text>。
            </Paragraph>

            <Title level={3}>2. 发起第一个请求</Title>
            <Paragraph>使用curl命令快速测试：</Paragraph>
            <CodeBlock
              language="bash"
              code={`curl http://localhost:8080/v1/chat/completions \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer YOUR_API_KEY" \\
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [
      {
        "role": "user",
        "content": "Hello, how are you?"
      }
    ]
  }'`}
            />

            <Alert
              message="提示"
              description="请将 YOUR_API_KEY 替换为您的实际API密钥。"
              type="info"
              showIcon
              style={{ marginTop: 16 }}
            />

            <Title level={3} style={{ marginTop: 24 }}>
              3. 查看可用模型
            </Title>
            <Paragraph>
              访问 <a href="/dashboard/models">模型列表</a> 页面查看所有可用的AI模型。
            </Paragraph>
          </>
        );

      case 'authentication':
        return (
          <>
            <Title level={2}>认证方式</Title>
            <Paragraph>
              Prism API 使用 Bearer Token 认证方式。您需要在HTTP请求头中包含您的API密钥。
            </Paragraph>

            <Title level={3}>请求头格式</Title>
            <CodeBlock
              language="http"
              code={`Authorization: Bearer YOUR_API_KEY
Content-Type: application/json`}
            />

            <Title level={3} style={{ marginTop: 24 }}>
              安全建议
            </Title>
            <ul>
              <li>
                <Paragraph>不要在客户端代码中硬编码API密钥</Paragraph>
              </li>
              <li>
                <Paragraph>使用环境变量存储API密钥</Paragraph>
              </li>
              <li>
                <Paragraph>定期轮换API密钥</Paragraph>
              </li>
              <li>
                <Paragraph>为不同的应用创建不同的API密钥</Paragraph>
              </li>
            </ul>
          </>
        );

      case 'openai-api':
        return (
          <>
            <Title level={2}>OpenAI格式API</Title>
            <Paragraph>
              我们提供完全兼容OpenAI的API接口，您可以直接使用OpenAI的SDK或切换Base URL。
            </Paragraph>

            <Title level={3}>端点</Title>
            <CodeBlock language="text" code="POST http://localhost:8080/v1/chat/completions" />

            <Title level={3} style={{ marginTop: 24 }}>
              Python示例
            </Title>
            <CodeBlock
              language="python"
              code={`from openai import OpenAI

client = OpenAI(
    api_key="YOUR_API_KEY",
    base_url="http://localhost:8080/v1"
)

response = client.chat.completions.create(
    model="gpt-4",
    messages=[
        {"role": "system", "content": "You are a helpful assistant."},
        {"role": "user", "content": "Hello!"}
    ],
    temperature=0.7,
    max_tokens=1000
)

print(response.choices[0].message.content)`}
            />

            <Title level={3} style={{ marginTop: 24 }}>
              Node.js示例
            </Title>
            <CodeBlock
              language="javascript"
              code={`import OpenAI from 'openai';

const client = new OpenAI({
  apiKey: 'YOUR_API_KEY',
  baseURL: 'http://localhost:8080/v1'
});

async function main() {
  const response = await client.chat.completions.create({
    model: 'gpt-4',
    messages: [
      { role: 'system', content: 'You are a helpful assistant.' },
      { role: 'user', content: 'Hello!' }
    ],
    temperature: 0.7,
    max_tokens: 1000
  });

  console.log(response.choices[0].message.content);
}

main();`}
            />

            <Title level={3} style={{ marginTop: 24 }}>
              请求参数
            </Title>
            <ul>
              <li>
                <Text strong>model</Text> (必需): 模型名称，如 "gpt-4", "gpt-3.5-turbo"
              </li>
              <li>
                <Text strong>messages</Text> (必需): 消息数组
              </li>
              <li>
                <Text strong>temperature</Text> (可选): 0-2之间，控制随机性
              </li>
              <li>
                <Text strong>max_tokens</Text> (可选): 最大生成token数
              </li>
              <li>
                <Text strong>stream</Text> (可选): 是否流式返回
              </li>
            </ul>
          </>
        );

      case 'anthropic-api':
        return (
          <>
            <Title level={2}>Anthropic格式API</Title>
            <Paragraph>支持Anthropic Claude模型的原生API格式。</Paragraph>

            <Title level={3}>端点</Title>
            <CodeBlock language="text" code="POST http://localhost:8080/v1/messages" />

            <Title level={3} style={{ marginTop: 24 }}>
              Python示例
            </Title>
            <CodeBlock
              language="python"
              code={`import anthropic

client = anthropic.Anthropic(
    api_key="YOUR_API_KEY",
    base_url="http://localhost:8080/v1"
)

message = client.messages.create(
    model="claude-3-opus-20240229",
    max_tokens=1024,
    messages=[
        {"role": "user", "content": "Hello, Claude!"}
    ]
)

print(message.content[0].text)`}
            />

            <Title level={3} style={{ marginTop: 24 }}>
              cURL示例
            </Title>
            <CodeBlock
              language="bash"
              code={`curl http://localhost:8080/v1/messages \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer YOUR_API_KEY" \\
  -H "anthropic-version: 2023-06-01" \\
  -d '{
    "model": "claude-3-opus-20240229",
    "max_tokens": 1024,
    "messages": [
      {
        "role": "user",
        "content": "Hello, Claude!"
      }
    ]
  }'`}
            />

            <Title level={3} style={{ marginTop: 24 }}>
              请求参数
            </Title>
            <ul>
              <li>
                <Text strong>model</Text> (必需): 模型名称，如 "claude-3-opus-20240229"
              </li>
              <li>
                <Text strong>messages</Text> (必需): 消息数组
              </li>
              <li>
                <Text strong>max_tokens</Text> (必需): 最大生成token数
              </li>
              <li>
                <Text strong>temperature</Text> (可选): 0-1之间
              </li>
              <li>
                <Text strong>system</Text> (可选): 系统提示词
              </li>
            </ul>
          </>
        );

      case 'gemini-api':
        return (
          <>
            <Title level={2}>Gemini格式API</Title>
            <Paragraph>支持Google Gemini模型的API格式。</Paragraph>

            <Title level={3}>端点</Title>
            <CodeBlock
              language="text"
              code="POST http://localhost:8080/v1/models/{model}/generateContent"
            />

            <Title level={3} style={{ marginTop: 24 }}>
              Python示例
            </Title>
            <CodeBlock
              language="python"
              code={`import requests

url = "http://localhost:8080/v1/models/gemini-pro/generateContent"
headers = {
    "Content-Type": "application/json",
    "Authorization": "Bearer YOUR_API_KEY"
}

data = {
    "contents": [
        {
            "parts": [
                {"text": "Hello, Gemini!"}
            ]
        }
    ],
    "generationConfig": {
        "temperature": 0.7,
        "maxOutputTokens": 1000
    }
}

response = requests.post(url, json=data, headers=headers)
result = response.json()
print(result['candidates'][0]['content']['parts'][0]['text'])`}
            />

            <Title level={3} style={{ marginTop: 24 }}>
              cURL示例
            </Title>
            <CodeBlock
              language="bash"
              code={`curl http://localhost:8080/v1/models/gemini-pro/generateContent \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer YOUR_API_KEY" \\
  -d '{
    "contents": [
      {
        "parts": [
          {"text": "Hello, Gemini!"}
        ]
      }
    ],
    "generationConfig": {
      "temperature": 0.7,
      "maxOutputTokens": 1000
    }
  }'`}
            />

            <Title level={3} style={{ marginTop: 24 }}>
              请求参数
            </Title>
            <ul>
              <li>
                <Text strong>contents</Text> (必需): 内容数组
              </li>
              <li>
                <Text strong>generationConfig</Text> (可选): 生成配置
              </li>
              <li>
                <Text strong>temperature</Text> (可选): 0-1之间
              </li>
              <li>
                <Text strong>maxOutputTokens</Text> (可选): 最大输出token数
              </li>
            </ul>
          </>
        );

      case 'sdk':
        return (
          <>
            <Title level={2}>SDK下载</Title>
            <Paragraph>
              我们推荐使用官方SDK，只需修改Base URL即可无缝接入我们的平台。
            </Paragraph>

            <Divider />

            <Card
              title={
                <Space>
                  <CodeOutlined />
                  <Text strong>Python SDK</Text>
                </Space>
              }
              extra={
                <Button
                  type="primary"
                  icon={<DownloadOutlined />}
                  href="https://pypi.org/project/openai/"
                  target="_blank"
                >
                  安装
                </Button>
              }
              style={{ marginBottom: 16 }}
            >
              <Paragraph>
                <Text strong>安装命令：</Text>
              </Paragraph>
              <CodeBlock language="bash" code="pip install openai" />
              <Paragraph style={{ marginTop: 16 }}>
                <Text strong>使用示例：</Text>
              </Paragraph>
              <CodeBlock
                language="python"
                code={`from openai import OpenAI

client = OpenAI(
    api_key="YOUR_API_KEY",
    base_url="http://localhost:8080/v1"
)

# 使用任何支持的模型
response = client.chat.completions.create(
    model="gpt-4",
    messages=[{"role": "user", "content": "Hello!"}]
)`}
              />
            </Card>

            <Card
              title={
                <Space>
                  <CodeOutlined />
                  <Text strong>Node.js SDK</Text>
                </Space>
              }
              extra={
                <Button
                  type="primary"
                  icon={<DownloadOutlined />}
                  href="https://www.npmjs.com/package/openai"
                  target="_blank"
                >
                  安装
                </Button>
              }
              style={{ marginBottom: 16 }}
            >
              <Paragraph>
                <Text strong>安装命令：</Text>
              </Paragraph>
              <CodeBlock language="bash" code="npm install openai" />
              <Paragraph style={{ marginTop: 16 }}>
                <Text strong>使用示例：</Text>
              </Paragraph>
              <CodeBlock
                language="javascript"
                code={`import OpenAI from 'openai';

const client = new OpenAI({
  apiKey: 'YOUR_API_KEY',
  baseURL: 'http://localhost:8080/v1'
});

// 使用任何支持的模型
const response = await client.chat.completions.create({
  model: 'gpt-4',
  messages: [{ role: 'user', content: 'Hello!' }]
});`}
              />
            </Card>

            <Card
              title={
                <Space>
                  <CodeOutlined />
                  <Text strong>Anthropic SDK</Text>
                </Space>
              }
              extra={
                <Button
                  type="primary"
                  icon={<DownloadOutlined />}
                  href="https://pypi.org/project/anthropic/"
                  target="_blank"
                >
                  安装
                </Button>
              }
              style={{ marginBottom: 16 }}
            >
              <Paragraph>
                <Text strong>安装命令：</Text>
              </Paragraph>
              <CodeBlock language="bash" code="pip install anthropic" />
              <Paragraph style={{ marginTop: 16 }}>
                <Text strong>使用示例：</Text>
              </Paragraph>
              <CodeBlock
                language="python"
                code={`import anthropic

client = anthropic.Anthropic(
    api_key="YOUR_API_KEY",
    base_url="http://localhost:8080/v1"
)

message = client.messages.create(
    model="claude-3-opus-20240229",
    max_tokens=1024,
    messages=[{"role": "user", "content": "Hello!"}]
)`}
              />
            </Card>

            <Alert
              message="提示"
              description="所有SDK都支持通过修改base_url/baseURL参数来使用我们的平台。您无需学习新的API，只需更改一个配置即可。"
              type="success"
              showIcon
              style={{ marginTop: 16 }}
            />
          </>
        );

      default:
        return null;
    }
  };

  return (
    <Layout style={{ background: 'transparent' }}>
      <Sider
        width={240}
        style={{
          background: '#fff',
          borderRadius: '8px',
          marginRight: '24px',
          overflow: 'auto',
          height: 'calc(100vh - 160px)',
          position: 'sticky',
          top: 88,
        }}
      >
        <div style={{ padding: '16px' }}>
          <Title level={4}>文档目录</Title>
        </div>
        <Menu
          mode="inline"
          selectedKeys={[selectedKey]}
          items={menuItems}
          onClick={({ key }) => setSelectedKey(key)}
          style={{ borderRight: 0 }}
        />
      </Sider>
      <Content
        style={{
          background: '#fff',
          padding: '32px',
          borderRadius: '8px',
          minHeight: 'calc(100vh - 160px)',
        }}
      >
        {renderContent()}
      </Content>
    </Layout>
  );
};

export default DocsPage;
