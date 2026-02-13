// 测试流式请求的日志记录和扣费功能
const http = require('http');

const API_BASE = 'http://localhost:8080';
const API_KEY = 'sk-afb6ec96bd1a90daba4b09b965ec586486a6a57c082ab8b4204a88244d9af117'; // 使用测试用户的 API Key

async function testStreamRequest() {
  console.log('=== Testing Stream Request ===\n');

  const requestBody = JSON.stringify({
    model: 'claude-sonnet-4.5',
    messages: [
      {
        role: 'user',
        content: 'Say "Hello, World!" in 5 different languages.'
      }
    ],
    stream: true,
    max_tokens: 200
  });

  const options = {
    hostname: 'localhost',
    port: 8080,
    path: '/v1/chat/completions',
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${API_KEY}`,
      'Content-Length': Buffer.byteLength(requestBody)
    }
  };

  return new Promise((resolve, reject) => {
    const req = http.request(options, (res) => {
      console.log(`Status: ${res.statusCode}`);
      console.log(`Headers:`, res.headers);
      console.log('\n=== Stream Response ===\n');

      let chunks = [];
      
      res.on('data', (chunk) => {
        const data = chunk.toString();
        chunks.push(data);
        process.stdout.write(data);
      });

      res.on('end', () => {
        console.log('\n\n=== Stream Completed ===');
        console.log(`Total chunks received: ${chunks.length}`);
        resolve();
      });
    });

    req.on('error', (error) => {
      console.error('Request error:', error);
      reject(error);
    });

    req.write(requestBody);
    req.end();
  });
}

// 运行测试
testStreamRequest()
  .then(() => {
    console.log('\n✓ Test completed successfully');
    console.log('\nPlease check:');
    console.log('1. Backend logs for stream processing details');
    console.log('2. Database logs table for the request record');
    console.log('3. User quota deduction');
    process.exit(0);
  })
  .catch((error) => {
    console.error('\n✗ Test failed:', error);
    process.exit(1);
  });
