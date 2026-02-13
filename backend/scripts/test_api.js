/**
 * API 完整功能测试脚本
 * 测试 OpenAI、Anthropic、Gemini 三种协议的各种功能
 * 
 * 使用方法:
 * node backend/scripts/test_api.js
 * 
 * 环境变量:
 * API_BASE_URL - API 基础地址 (默认: http://localhost:8080)
 * API_KEY - API 密钥 (必需)
 */

const https = require('https');
const http = require('http');

// 配置
const config = {
  baseUrl: process.env.API_BASE_URL || 'http://localhost:8080',
  apiKey: process.env.API_KEY || 'sk-afb6ec96bd1a90daba4b09b965ec586486a6a57c082ab8b4204a88244d9af117',
};

if (!config.apiKey) {
  console.error('错误: 请设置 API_KEY 环境变量');
  console.error('示例: API_KEY=your-api-key node backend/scripts/test_api.js');
  process.exit(1);
}

// 颜色输出
const colors = {
  reset: '\x1b[0m',
  bright: '\x1b[1m',
  green: '\x1b[32m',
  red: '\x1b[31m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
  cyan: '\x1b[36m',
};

function log(message, color = 'reset') {
  console.log(`${colors[color]}${message}${colors.reset}`);
}

function logSection(title) {
  console.log('\n' + '='.repeat(80));
  log(title, 'bright');
  console.log('='.repeat(80) + '\n');
}

function logTest(name) {
  log(`\n▶ 测试: ${name}`, 'cyan');
}

function logSuccess(message) {
  log(`✓ ${message}`, 'green');
}

function logError(message) {
  log(`✗ ${message}`, 'red');
}

function logWarning(message) {
  log(`⚠ ${message}`, 'yellow');
}

// HTTP 请求函数
function makeRequest(options, data = null) {
  return new Promise((resolve, reject) => {
    const url = new URL(options.path, config.baseUrl);
    const isHttps = url.protocol === 'https:';
    const client = isHttps ? https : http;

    const requestOptions = {
      hostname: url.hostname,
      port: url.port || (isHttps ? 443 : 80),
      path: url.pathname + url.search,
      method: options.method || 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${config.apiKey}`,
        ...options.headers,
      },
    };

    const req = client.request(requestOptions, (res) => {
      let body = '';
      let chunks = [];

      res.on('data', (chunk) => {
        if (options.stream) {
          chunks.push(chunk.toString());
        } else {
          body += chunk;
        }
      });

      res.on('end', () => {
        if (options.stream) {
          resolve({ statusCode: res.statusCode, chunks });
        } else {
          try {
            const parsed = body ? JSON.parse(body) : {};
            resolve({ statusCode: res.statusCode, data: parsed, body });
          } catch (e) {
            resolve({ statusCode: res.statusCode, body });
          }
        }
      });
    });

    req.on('error', reject);

    if (data) {
      req.write(JSON.stringify(data));
    }

    req.end();
  });
}

// 测试用例
const tests = {
  // ==================== OpenAI 协议测试 ====================
  async testOpenAISimple() {
    logTest('OpenAI - 简单对话');
    const response = await makeRequest({
      path: '/v1/chat/completions',
      method: 'POST',
    }, {
      model: 'claude-sonnet-4.5',
      messages: [
        { role: 'user', content: '你好，请用一句话介绍你自己' }
      ],
      max_tokens: 100,
    });

    if (response.statusCode === 200 && response.data.choices) {
      logSuccess(`状态码: ${response.statusCode}`);
      logSuccess(`模型: ${response.data.model}`);
      logSuccess(`回复: ${response.data.choices[0].message.content}`);
      logSuccess(`Token 使用: ${JSON.stringify(response.data.usage)}`);
      return true;
    } else {
      logError(`失败: ${response.statusCode} - ${JSON.stringify(response.data || response.body)}`);
      return false;
    }
  },

  async testOpenAIStreaming() {
    logTest('OpenAI - 流式响应');
    const response = await makeRequest({
      path: '/v1/chat/completions',
      method: 'POST',
      stream: true,
    }, {
      model: 'claude-sonnet-4.5',
      messages: [
        { role: 'user', content: '数到5，每个数字一行' }
      ],
      stream: true,
      max_tokens: 50,
    });

    if (response.statusCode === 200 && response.chunks.length > 0) {
      logSuccess(`状态码: ${response.statusCode}`);
      logSuccess(`收到 ${response.chunks.length} 个数据块`);
      
      let fullContent = '';
      for (const chunk of response.chunks) {
        const lines = chunk.split('\n').filter(line => line.trim() && line.includes('data: '));
        for (const line of lines) {
          const data = line.replace('data: ', '').trim();
          if (data === '[DONE]') continue;
          try {
            const parsed = JSON.parse(data);
            if (parsed.choices?.[0]?.delta?.content) {
              fullContent += parsed.choices[0].delta.content;
            }
          } catch (e) {
            // 忽略解析错误
          }
        }
      }
      
      logSuccess(`完整内容: ${fullContent}`);
      return true;
    } else {
      logError(`失败: ${response.statusCode}`);
      return false;
    }
  },

  async testOpenAIMultiTurn() {
    logTest('OpenAI - 多轮对话');
    const response = await makeRequest({
      path: '/v1/chat/completions',
      method: 'POST',
    }, {
      model: 'claude-sonnet-4.5',
      messages: [
        { role: 'user', content: '我叫小明' },
        { role: 'assistant', content: '你好小明，很高兴认识你！' },
        { role: 'user', content: '我刚才说我叫什么？' }
      ],
      max_tokens: 50,
    });

    if (response.statusCode === 200 && response.data.choices) {
      const reply = response.data.choices[0].message.content;
      logSuccess(`状态码: ${response.statusCode}`);
      logSuccess(`回复: ${reply}`);
      if (reply.includes('小明')) {
        logSuccess('✓ 正确记住了上下文');
        return true;
      } else {
        logWarning('⚠ 可能没有正确记住上下文');
        return true;
      }
    } else {
      logError(`失败: ${response.statusCode} - ${JSON.stringify(response.data || response.body)}`);
      return false;
    }
  },

  async testOpenAISystemPrompt() {
    logTest('OpenAI - 系统提示词');
    const response = await makeRequest({
      path: '/v1/chat/completions',
      method: 'POST',
    }, {
      model: 'claude-sonnet-4.5',
      messages: [
        { role: 'system', content: '你是一个海盗，说话要像海盗一样' },
        { role: 'user', content: '你好' }
      ],
      max_tokens: 100,
    });

    if (response.statusCode === 200 && response.data.choices) {
      logSuccess(`状态码: ${response.statusCode}`);
      logSuccess(`回复: ${response.data.choices[0].message.content}`);
      return true;
    } else {
      logError(`失败: ${response.statusCode} - ${JSON.stringify(response.data || response.body)}`);
      return false;
    }
  },

  async testOpenAITools() {
    logTest('OpenAI - 工具调用');
    const response = await makeRequest({
      path: '/v1/chat/completions',
      method: 'POST',
    }, {
      model: 'claude-sonnet-4.5',
      messages: [
        { role: 'user', content: '北京今天天气怎么样？' }
      ],
      tools: [
        {
          type: 'function',
          function: {
            name: 'get_weather',
            description: '获取指定城市的天气信息',
            parameters: {
              type: 'object',
              properties: {
                city: {
                  type: 'string',
                  description: '城市名称，例如：北京、上海'
                }
              },
              required: ['city']
            }
          }
        }
      ],
      tool_choice: 'auto',
      max_tokens: 200,
    });

    if (response.statusCode === 200) {
      logSuccess(`状态码: ${response.statusCode}`);
      const choice = response.data.choices?.[0];
      if (choice?.message?.tool_calls) {
        logSuccess(`工具调用: ${JSON.stringify(choice.message.tool_calls, null, 2)}`);
        return true;
      } else {
        logSuccess(`普通回复: ${choice?.message?.content}`);
        logWarning('模型选择了不使用工具');
        return true;
      }
    } else {
      logError(`失败: ${response.statusCode} - ${JSON.stringify(response.data || response.body)}`);
      return false;
    }
  },

  // ==================== Anthropic 协议测试 ====================
  async testAnthropicSimple() {
    logTest('Anthropic - 简单对话');
    const response = await makeRequest({
      path: '/v1/messages',
      method: 'POST',
      headers: {
        'anthropic-version': '2023-06-01',
      },
    }, {
      model: 'claude-sonnet-4.5',
      messages: [
        { role: 'user', content: '你好，请用一句话介绍你自己' }
      ],
      max_tokens: 100,
    });

    if (response.statusCode === 200 && response.data.content) {
      logSuccess(`状态码: ${response.statusCode}`);
      logSuccess(`模型: ${response.data.model}`);
      logSuccess(`回复: ${response.data.content[0].text}`);
      logSuccess(`Token 使用: ${JSON.stringify(response.data.usage)}`);
      return true;
    } else {
      logError(`失败: ${response.statusCode} - ${JSON.stringify(response.data || response.body)}`);
      return false;
    }
  },

  async testAnthropicStreaming() {
    logTest('Anthropic - 流式响应');
    const response = await makeRequest({
      path: '/v1/messages',
      method: 'POST',
      stream: true,
      headers: {
        'anthropic-version': '2023-06-01',
      },
    }, {
      model: 'claude-sonnet-4.5',
      messages: [
        { role: 'user', content: '数到5，每个数字一行' }
      ],
      stream: true,
      max_tokens: 50,
    });

    if (response.statusCode === 200 && response.chunks.length > 0) {
      logSuccess(`状态码: ${response.statusCode}`);
      logSuccess(`收到 ${response.chunks.length} 个数据块`);
      
      let fullContent = '';
      for (const chunk of response.chunks) {
        const lines = chunk.split('\n').filter(line => line.trim() && line.includes('data: '));
        for (const line of lines) {
          const data = line.replace('data: ', '').trim();
          try {
            const parsed = JSON.parse(data);
            if (parsed.type === 'content_block_delta' && parsed.delta?.text) {
              fullContent += parsed.delta.text;
            }
          } catch (e) {
            // 忽略解析错误
          }
        }
      }
      
      logSuccess(`完整内容: ${fullContent}`);
      return true;
    } else {
      logError(`失败: ${response.statusCode}`);
      return false;
    }
  },

  async testAnthropicMultiTurn() {
    logTest('Anthropic - 多轮对话');
    const response = await makeRequest({
      path: '/v1/messages',
      method: 'POST',
      headers: {
        'anthropic-version': '2023-06-01',
      },
    }, {
      model: 'claude-sonnet-4.5',
      messages: [
        { role: 'user', content: '我叫小明' },
        { role: 'assistant', content: '你好小明，很高兴认识你！' },
        { role: 'user', content: '我刚才说我叫什么？' }
      ],
      max_tokens: 50,
    });

    if (response.statusCode === 200 && response.data.content) {
      const reply = response.data.content[0].text;
      logSuccess(`状态码: ${response.statusCode}`);
      logSuccess(`回复: ${reply}`);
      if (reply.includes('小明')) {
        logSuccess('✓ 正确记住了上下文');
        return true;
      } else {
        logWarning('⚠ 可能没有正确记住上下文');
        return true;
      }
    } else {
      logError(`失败: ${response.statusCode} - ${JSON.stringify(response.data || response.body)}`);
      return false;
    }
  },

  async testAnthropicSystemPrompt() {
    logTest('Anthropic - 系统提示词');
    const response = await makeRequest({
      path: '/v1/messages',
      method: 'POST',
      headers: {
        'anthropic-version': '2023-06-01',
      },
    }, {
      model: 'claude-sonnet-4.5',
      system: '你是一个海盗，说话要像海盗一样',
      messages: [
        { role: 'user', content: '你好' }
      ],
      max_tokens: 100,
    });

    if (response.statusCode === 200 && response.data.content) {
      logSuccess(`状态码: ${response.statusCode}`);
      logSuccess(`回复: ${response.data.content[0].text}`);
      return true;
    } else {
      logError(`失败: ${response.statusCode} - ${JSON.stringify(response.data || response.body)}`);
      return false;
    }
  },

  async testAnthropicTools() {
    logTest('Anthropic - 工具调用');
    const response = await makeRequest({
      path: '/v1/messages',
      method: 'POST',
      headers: {
        'anthropic-version': '2023-06-01',
      },
    }, {
      model: 'claude-sonnet-4.5',
      messages: [
        { role: 'user', content: '北京今天天气怎么样？' }
      ],
      tools: [
        {
          name: 'get_weather',
          description: '获取指定城市的天气信息',
          input_schema: {
            type: 'object',
            properties: {
              city: {
                type: 'string',
                description: '城市名称，例如：北京、上海'
              }
            },
            required: ['city']
          }
        }
      ],
      max_tokens: 200,
    });

    if (response.statusCode === 200) {
      logSuccess(`状态码: ${response.statusCode}`);
      const content = response.data.content;
      const toolUse = content?.find(c => c.type === 'tool_use');
      if (toolUse) {
        logSuccess(`工具调用: ${JSON.stringify(toolUse, null, 2)}`);
        return true;
      } else {
        const textContent = content?.find(c => c.type === 'text');
        logSuccess(`普通回复: ${textContent?.text}`);
        logWarning('模型选择了不使用工具');
        return true;
      }
    } else {
      logError(`失败: ${response.statusCode} - ${JSON.stringify(response.data || response.body)}`);
      return false;
    }
  },

  // ==================== Gemini 协议测试 ====================
  async testGeminiSimple() {
    logTest('Gemini - 简单对话');
    const response = await makeRequest({
      path: '/v1/models/claude-sonnet-4.5:generateContent',
      method: 'POST',
    }, {
      contents: [
        {
          role: 'user',
          parts: [{ text: '你好，请用一句话介绍你自己' }]
        }
      ],
      generationConfig: {
        maxOutputTokens: 100,
      },
    });

    if (response.statusCode === 200 && response.data.candidates) {
      logSuccess(`状态码: ${response.statusCode}`);
      logSuccess(`回复: ${response.data.candidates[0].content.parts[0].text}`);
      if (response.data.usageMetadata) {
        logSuccess(`Token 使用: ${JSON.stringify(response.data.usageMetadata)}`);
      }
      return true;
    } else {
      logError(`失败: ${response.statusCode} - ${JSON.stringify(response.data || response.body)}`);
      return false;
    }
  },

  async testGeminiStreaming() {
    logTest('Gemini - 流式响应');
    const response = await makeRequest({
      path: '/v1/models/claude-sonnet-4.5:streamGenerateContent',
      method: 'POST',
      stream: true,
    }, {
      contents: [
        {
          role: 'user',
          parts: [{ text: '数到5，每个数字一行' }]
        }
      ],
      generationConfig: {
        maxOutputTokens: 50,
      },
    });

    if (response.statusCode === 200 && response.chunks.length > 0) {
      logSuccess(`状态码: ${response.statusCode}`);
      logSuccess(`收到 ${response.chunks.length} 个数据块`);
      
      let fullContent = '';
      for (const chunk of response.chunks) {
        const lines = chunk.split('\n').filter(line => line.trim());
        for (const line of lines) {
          try {
            const parsed = JSON.parse(line);
            if (parsed.candidates?.[0]?.content?.parts?.[0]?.text) {
              fullContent += parsed.candidates[0].content.parts[0].text;
            }
          } catch (e) {
            // 忽略解析错误
          }
        }
      }
      
      logSuccess(`完整内容: ${fullContent}`);
      return true;
    } else {
      logError(`失败: ${response.statusCode}`);
      return false;
    }
  },

  async testGeminiMultiTurn() {
    logTest('Gemini - 多轮对话');
    const response = await makeRequest({
      path: '/v1/models/claude-sonnet-4.5:generateContent',
      method: 'POST',
    }, {
      contents: [
        {
          role: 'user',
          parts: [{ text: '我叫小明' }]
        },
        {
          role: 'model',
          parts: [{ text: '你好小明，很高兴认识你！' }]
        },
        {
          role: 'user',
          parts: [{ text: '我刚才说我叫什么？' }]
        }
      ],
      generationConfig: {
        maxOutputTokens: 50,
      },
    });

    if (response.statusCode === 200 && response.data.candidates) {
      const reply = response.data.candidates[0].content.parts[0].text;
      logSuccess(`状态码: ${response.statusCode}`);
      logSuccess(`回复: ${reply}`);
      if (reply.includes('小明')) {
        logSuccess('✓ 正确记住了上下文');
        return true;
      } else {
        logWarning('⚠ 可能没有正确记住上下文');
        return true;
      }
    } else {
      logError(`失败: ${response.statusCode} - ${JSON.stringify(response.data || response.body)}`);
      return false;
    }
  },

  async testGeminiSystemPrompt() {
    logTest('Gemini - 系统提示词');
    const response = await makeRequest({
      path: '/v1/models/claude-sonnet-4.5:generateContent',
      method: 'POST',
    }, {
      systemInstruction: {
        parts: [{ text: '你是一个海盗，说话要像海盗一样' }]
      },
      contents: [
        {
          role: 'user',
          parts: [{ text: '你好' }]
        }
      ],
      generationConfig: {
        maxOutputTokens: 100,
      },
    });

    if (response.statusCode === 200 && response.data.candidates) {
      logSuccess(`状态码: ${response.statusCode}`);
      logSuccess(`回复: ${response.data.candidates[0].content.parts[0].text}`);
      return true;
    } else {
      logError(`失败: ${response.statusCode} - ${JSON.stringify(response.data || response.body)}`);
      return false;
    }
  },

  async testGeminiTools() {
    logTest('Gemini - 工具调用');
    const response = await makeRequest({
      path: '/v1/models/claude-sonnet-4.5:generateContent',
      method: 'POST',
    }, {
      contents: [
        {
          role: 'user',
          parts: [{ text: '北京今天天气怎么样？' }]
        }
      ],
      tools: [
        {
          functionDeclarations: [
            {
              name: 'get_weather',
              description: '获取指定城市的天气信息',
              parameters: {
                type: 'object',
                properties: {
                  city: {
                    type: 'string',
                    description: '城市名称，例如：北京、上海'
                  }
                },
                required: ['city']
              }
            }
          ]
        }
      ],
      generationConfig: {
        maxOutputTokens: 200,
      },
    });

    if (response.statusCode === 200) {
      logSuccess(`状态码: ${response.statusCode}`);
      const parts = response.data.candidates?.[0]?.content?.parts;
      const functionCall = parts?.find(p => p.functionCall);
      if (functionCall) {
        logSuccess(`工具调用: ${JSON.stringify(functionCall, null, 2)}`);
        return true;
      } else {
        const textPart = parts?.find(p => p.text);
        logSuccess(`普通回复: ${textPart?.text}`);
        logWarning('模型选择了不使用工具');
        return true;
      }
    } else {
      logError(`失败: ${response.statusCode} - ${JSON.stringify(response.data || response.body)}`);
      return false;
    }
  },
};

// 运行所有测试
async function runAllTests() {
  log('API 完整功能测试', 'bright');
  log(`基础地址: ${config.baseUrl}`, 'blue');
  log(`API 密钥: ${config.apiKey.substring(0, 10)}...`, 'blue');

  const results = {
    total: 0,
    passed: 0,
    failed: 0,
  };

  // OpenAI 测试
  logSection('OpenAI 协议测试');
  for (const test of [
    'testOpenAISimple',
    'testOpenAIStreaming',
    'testOpenAIMultiTurn',
    'testOpenAISystemPrompt',
    'testOpenAITools',
  ]) {
    results.total++;
    try {
      const passed = await tests[test]();
      if (passed) {
        results.passed++;
      } else {
        results.failed++;
      }
    } catch (error) {
      logError(`异常: ${error.message}`);
      results.failed++;
    }
    await new Promise(resolve => setTimeout(resolve, 1000)); // 延迟1秒
  }

  // Anthropic 测试
  logSection('Anthropic 协议测试');
  for (const test of [
    'testAnthropicSimple',
    'testAnthropicStreaming',
    'testAnthropicMultiTurn',
    'testAnthropicSystemPrompt',
    'testAnthropicTools',
  ]) {
    results.total++;
    try {
      const passed = await tests[test]();
      if (passed) {
        results.passed++;
      } else {
        results.failed++;
      }
    } catch (error) {
      logError(`异常: ${error.message}`);
      results.failed++;
    }
    await new Promise(resolve => setTimeout(resolve, 1000));
  }

  // Gemini 测试
  logSection('Gemini 协议测试');
  for (const test of [
    'testGeminiSimple',
    'testGeminiStreaming',
    'testGeminiMultiTurn',
    'testGeminiSystemPrompt',
    'testGeminiTools',
  ]) {
    results.total++;
    try {
      const passed = await tests[test]();
      if (passed) {
        results.passed++;
      } else {
        results.failed++;
      }
    } catch (error) {
      logError(`异常: ${error.message}`);
      results.failed++;
    }
    await new Promise(resolve => setTimeout(resolve, 1000));
  }

  // 输出总结
  logSection('测试总结');
  log(`总测试数: ${results.total}`, 'blue');
  log(`通过: ${results.passed}`, 'green');
  log(`失败: ${results.failed}`, 'red');
  log(`成功率: ${((results.passed / results.total) * 100).toFixed(1)}%`, 'bright');

  process.exit(results.failed > 0 ? 1 : 0);
}

// 运行测试
runAllTests().catch(error => {
  logError(`致命错误: ${error.message}`);
  console.error(error);
  process.exit(1);
});
