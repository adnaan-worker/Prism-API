// 测试通过 Prism API 转发到 Kiro 的工具调用
const http = require('http');

// Prism API 配置
const PRISM_API_URL = 'http://localhost:8080';
const API_KEY = 'sk-afb6ec96bd1a90daba4b09b965ec586486a6a57c082ab8b4204a88244d9af117';

// OpenAI 格式的请求（会被 Prism 转换为 Kiro 格式）
const testPayload = {
  model: 'claude-sonnet-4.5',
  messages: [
    {
      role: 'user',
      content: '请列出当前目录的文件，然后读取 README.md 文件的内容'
    }
  ],
  tools: [
    {
      type: 'function',
      function: {
        name: 'list_directory',
        description: 'List directory contents. Use recursive=true for tree view.',
        parameters: {
          type: 'object',
          properties: {
            path: {
              type: 'string',
              description: 'Directory path relative to workspace root'
            },
            recursive: {
              type: 'boolean',
              description: 'Show subdirectories recursively'
            }
          },
          required: ['path']
        }
      }
    },
    {
      type: 'function',
      function: {
        name: 'read_file',
        description: 'Read one or more files with line numbers',
        parameters: {
          type: 'object',
          properties: {
            path: {
              type: 'string',
              description: 'File path or array of paths'
            },
            start_line: {
              type: 'number',
              description: 'Starting line (1-indexed)'
            },
            end_line: {
              type: 'number',
              description: 'Ending line (inclusive)'
            }
          },
          required: ['path']
        }
      }
    }
  ],
  stream: true,
  max_tokens: 2000,
  temperature: 0.7
};

const payload = JSON.stringify(testPayload);

console.log('=== Testing Kiro Tool Call via Prism API ===');
console.log('Prism API URL:', PRISM_API_URL);
console.log('Model:', testPayload.model);
console.log('Tools count:', testPayload.tools.length);
console.log('Payload size:', payload.length, 'bytes');
console.log('\nSending request...\n');

const options = {
  hostname: 'localhost',
  port: 8080,
  path: '/v1/chat/completions',
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${API_KEY}`,
    'Content-Length': Buffer.byteLength(payload)
  }
};

const req = http.request(options, (res) => {
  console.log('Status:', res.statusCode);
  console.log('Headers:', JSON.stringify(res.headers, null, 2));
  console.log('\n--- Response Stream ---\n');
  
  if (res.statusCode !== 200) {
    let errorBody = '';
    res.on('data', chunk => errorBody += chunk);
    res.on('end', () => {
      console.error('\n✗ Error response:', errorBody);
    });
    return;
  }

  let buffer = '';
  let toolCallsFound = [];
  let currentContent = '';

  res.on('data', (chunk) => {
    buffer += chunk.toString();
    
    // 解析 SSE 流
    const lines = buffer.split('\n');
    buffer = lines.pop() || ''; // 保留最后一个不完整的行

    for (const line of lines) {
      if (line.startsWith('data: ')) {
        const data = line.slice(6).trim();
        
        if (data === '[DONE]') {
          console.log('\n\n=== Stream completed ===');
          continue;
        }

        try {
          const chunk = JSON.parse(data);
          
          // 检查是否有内容
          if (chunk.choices && chunk.choices[0].delta) {
            const delta = chunk.choices[0].delta;
            
            if (delta.content) {
              currentContent += delta.content;
              process.stdout.write(delta.content);
            }
            
            // 检查工具调用
            if (delta.tool_calls && delta.tool_calls.length > 0) {
              for (const toolCall of delta.tool_calls) {
                const existingCall = toolCallsFound.find(tc => tc.id === toolCall.id);
                
                if (!existingCall) {
                  // 新的工具调用
                  toolCallsFound.push({
                    id: toolCall.id,
                    type: toolCall.type,
                    function: {
                      name: toolCall.function?.name || '',
                      arguments: toolCall.function?.arguments || ''
                    }
                  });
                  console.log('\n\n✓ Tool call detected!');
                  console.log('  ID:', toolCall.id);
                  console.log('  Name:', toolCall.function?.name);
                } else {
                  // 累积参数
                  if (toolCall.function?.arguments) {
                    existingCall.function.arguments += toolCall.function.arguments;
                  }
                }
              }
            }
          }
          
          // 检查完成原因
          if (chunk.choices && chunk.choices[0].finish_reason) {
            console.log('\n\nFinish reason:', chunk.choices[0].finish_reason);
          }
        } catch (e) {
          // 忽略解析错误
        }
      }
    }
  });

  res.on('end', () => {
    console.log('\n\n=== Analysis ===');
    console.log('Content length:', currentContent.length);
    console.log('Tool calls found:', toolCallsFound.length);
    
    if (toolCallsFound.length > 0) {
      console.log('\n--- Tool Calls ---');
      for (const toolCall of toolCallsFound) {
        console.log('\nTool:', toolCall.function.name);
        console.log('ID:', toolCall.id);
        console.log('Arguments length:', toolCall.function.arguments.length);
        console.log('Arguments:', toolCall.function.arguments);
        
        // 验证 JSON
        if (toolCall.function.arguments) {
          try {
            const parsed = JSON.parse(toolCall.function.arguments);
            console.log('✓ Arguments are valid JSON');
            console.log('Parsed:', JSON.stringify(parsed, null, 2));
          } catch (e) {
            console.log('✗ Arguments are NOT valid JSON:', e.message);
            console.log('Raw arguments:', toolCall.function.arguments);
          }
        } else {
          console.log('✗ No arguments provided!');
        }
      }
    } else {
      console.log('✗ No tool calls detected');
    }
  });
});

req.on('error', (e) => {
  console.error('✗ Request error:', e);
});

req.write(payload);
req.end();
