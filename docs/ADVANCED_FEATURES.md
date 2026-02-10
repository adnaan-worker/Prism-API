# 高级功能支持清单

## 1. 工具调用 (Function Calling / Tool Use)

### OpenAI Function Calling

#### 请求格式
```json
{
  "model": "gpt-4",
  "messages": [
    {"role": "user", "content": "What's the weather in Boston?"}
  ],
  "tools": [
    {
      "type": "function",
      "function": {
        "name": "get_weather",
        "description": "Get the current weather in a location",
        "parameters": {
          "type": "object",
          "properties": {
            "location": {
              "type": "string",
              "description": "The city and state, e.g. San Francisco, CA"
            },
            "unit": {
              "type": "string",
              "enum": ["celsius", "fahrenheit"]
            }
          },
          "required": ["location"]
        }
      }
    }
  ],
  "tool_choice": "auto"
}
```

#### 响应格式
```json
{
  "id": "chatcmpl-123",
  "object": "chat.completion",
  "created": 1677652288,
  "model": "gpt-4",
  "choices": [{
    "index": 0,
    "message": {
      "role": "assistant",
      "content": null,
      "tool_calls": [
        {
          "id": "call_abc123",
          "type": "function",
          "function": {
            "name": "get_weather",
            "arguments": "{\"location\": \"Boston, MA\"}"
          }
        }
      ]
    },
    "finish_reason": "tool_calls"
  }],
  "usage": {
    "prompt_tokens": 82,
    "completion_tokens": 17,
    "total_tokens": 99
  }
}
```

#### 需要支持的字段
- ✅ `tools` 数组（工具定义）
- ✅ `tool_choice`（auto/none/required/specific function）
- ✅ `message.tool_calls` 数组（工具调用结果）
- ✅ `finish_reason: "tool_calls"`

---

### Anthropic Tool Use

#### 请求格式
```json
{
  "model": "claude-3-opus-20240229",
  "max_tokens": 1024,
  "tools": [
    {
      "name": "get_weather",
      "description": "Get the current weather in a given location",
      "input_schema": {
        "type": "object",
        "properties": {
          "location": {
            "type": "string",
            "description": "The city and state, e.g. San Francisco, CA"
          }
        },
        "required": ["location"]
      }
    }
  ],
  "messages": [
    {"role": "user", "content": "What's the weather in Boston?"}
  ]
}
```

#### 响应格式
```json
{
  "id": "msg_01Aq9w938a90dw8q",
  "type": "message",
  "role": "assistant",
  "content": [
    {
      "type": "tool_use",
      "id": "toolu_01A09q90qw90lq917835lq9",
      "name": "get_weather",
      "input": {"location": "Boston, MA"}
    }
  ],
  "model": "claude-3-opus-20240229",
  "stop_reason": "tool_use",
  "usage": {
    "input_tokens": 385,
    "output_tokens": 48
  }
}
```

#### 需要支持的字段
- ✅ `tools` 数组（使用 `input_schema` 而不是 `parameters`）
- ✅ `content` 中的 `type: "tool_use"` 对象
- ✅ `stop_reason: "tool_use"`
- ✅ 工具结果回传：`role: "user"`, `content: [{"type": "tool_result", "tool_use_id": "...", "content": "..."}]`

---

### Gemini Function Calling

#### 请求格式
```json
{
  "contents": [
    {
      "role": "user",
      "parts": [{"text": "What's the weather in Boston?"}]
    }
  ],
  "tools": [
    {
      "function_declarations": [
        {
          "name": "get_weather",
          "description": "Get the current weather in a location",
          "parameters": {
            "type": "object",
            "properties": {
              "location": {
                "type": "string",
                "description": "The city and state"
              }
            },
            "required": ["location"]
          }
        }
      ]
    }
  ]
}
```

#### 响应格式
```json
{
  "candidates": [
    {
      "content": {
        "parts": [
          {
            "functionCall": {
              "name": "get_weather",
              "args": {
                "location": "Boston, MA"
              }
            }
          }
        ],
        "role": "model"
      },
      "finishReason": "STOP"
    }
  ],
  "usageMetadata": {
    "promptTokenCount": 41,
    "candidatesTokenCount": 7,
    "totalTokenCount": 48
  }
}
```

#### 需要支持的字段
- ✅ `tools` 数组（使用 `function_declarations`）
- ✅ `parts` 中的 `functionCall` 对象
- ✅ 工具结果回传：`parts: [{"functionResponse": {"name": "...", "response": {...}}}]`

---

## 2. 多模态输入 (Vision)

### OpenAI Vision

#### 请求格式
```json
{
  "model": "gpt-4-vision-preview",
  "messages": [
    {
      "role": "user",
      "content": [
        {"type": "text", "text": "What's in this image?"},
        {
          "type": "image_url",
          "image_url": {
            "url": "https://example.com/image.jpg",
            "detail": "high"
          }
        }
      ]
    }
  ],
  "max_tokens": 300
}
```

#### 需要支持的字段
- ✅ `content` 可以是字符串或数组
- ✅ `content` 数组中的 `type: "text"` 和 `type: "image_url"`
- ✅ `image_url.detail`（low/high/auto）

---

### Anthropic Vision

#### 请求格式
```json
{
  "model": "claude-3-opus-20240229",
  "max_tokens": 1024,
  "messages": [
    {
      "role": "user",
      "content": [
        {
          "type": "image",
          "source": {
            "type": "base64",
            "media_type": "image/jpeg",
            "data": "/9j/4AAQSkZJRg..."
          }
        },
        {
          "type": "text",
          "text": "What's in this image?"
        }
      ]
    }
  ]
}
```

#### 需要支持的字段
- ✅ `content` 数组中的 `type: "image"`
- ✅ `source.type`（base64/url）
- ✅ `source.media_type`（image/jpeg, image/png, image/gif, image/webp）

---

### Gemini Vision

#### 请求格式
```json
{
  "contents": [
    {
      "role": "user",
      "parts": [
        {"text": "What's in this image?"},
        {
          "inline_data": {
            "mime_type": "image/jpeg",
            "data": "/9j/4AAQSkZJRg..."
          }
        }
      ]
    }
  ]
}
```

#### 需要支持的字段
- ✅ `parts` 中的 `inline_data` 对象
- ✅ `mime_type` 和 `data`（base64）

---

## 3. 其他高级参数

### OpenAI 高级参数

```json
{
  "model": "gpt-4",
  "messages": [...],
  "temperature": 0.7,
  "top_p": 1.0,
  "n": 1,
  "stream": false,
  "stop": ["\n", "END"],
  "max_tokens": 100,
  "presence_penalty": 0.0,
  "frequency_penalty": 0.0,
  "logit_bias": {"50256": -100},
  "user": "user-123",
  "response_format": {"type": "json_object"},
  "seed": 42
}
```

#### 需要支持的字段
- ⚠️ `top_p`（核采样）
- ⚠️ `n`（生成多个响应）
- ⚠️ `stop`（停止序列）
- ⚠️ `presence_penalty`（存在惩罚）
- ⚠️ `frequency_penalty`（频率惩罚）
- ⚠️ `logit_bias`（logit 偏置）
- ⚠️ `user`（用户标识）
- ⚠️ `response_format`（JSON 模式）
- ⚠️ `seed`（确定性输出）

---

### Anthropic 高级参数

```json
{
  "model": "claude-3-opus-20240229",
  "max_tokens": 1024,
  "messages": [...],
  "system": "You are a helpful assistant",
  "temperature": 0.7,
  "top_p": 1.0,
  "top_k": 40,
  "stop_sequences": ["\n\nHuman:"],
  "metadata": {
    "user_id": "user-123"
  }
}
```

#### 需要支持的字段
- ✅ `system`（已支持）
- ⚠️ `top_p`
- ⚠️ `top_k`（Top-K 采样）
- ⚠️ `stop_sequences`
- ⚠️ `metadata`（元数据）

---

### Gemini 高级参数

```json
{
  "contents": [...],
  "generationConfig": {
    "temperature": 0.7,
    "topP": 0.95,
    "topK": 40,
    "maxOutputTokens": 1024,
    "stopSequences": ["END"],
    "candidateCount": 1,
    "responseMimeType": "application/json"
  },
  "safetySettings": [
    {
      "category": "HARM_CATEGORY_HARASSMENT",
      "threshold": "BLOCK_MEDIUM_AND_ABOVE"
    }
  ]
}
```

#### 需要支持的字段
- ⚠️ `topP`
- ⚠️ `topK`
- ⚠️ `stopSequences`
- ⚠️ `candidateCount`
- ⚠️ `responseMimeType`（JSON 模式）
- ⚠️ `safetySettings`（安全设置）

---

## 4. 流式响应中的工具调用

### OpenAI 流式工具调用

```
data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1677652288,"model":"gpt-4","choices":[{"index":0,"delta":{"role":"assistant","content":null,"tool_calls":[{"index":0,"id":"call_abc123","type":"function","function":{"name":"get_weather","arguments":""}}]},"finish_reason":null}]}

data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1677652288,"model":"gpt-4","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"function":{"arguments":"{\"lo"}}]},"finish_reason":null}]}

data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1677652288,"model":"gpt-4","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"function":{"arguments":"cation"}}]},"finish_reason":null}]}

data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1677652288,"model":"gpt-4","choices":[{"index":0,"delta":{},"finish_reason":"tool_calls"}]}

data: [DONE]
```

### Anthropic 流式工具调用

```
event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"tool_use","id":"toolu_01A09q90qw90lq917835lq9","name":"get_weather","input":{}}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"{\"location\": "}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"\"Boston, MA\"}"}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"tool_use"},"usage":{"output_tokens":48}}

event: message_stop
data: {"type":"message_stop"}
```

---

## 5. 实现优先级

### 🔴 P0 - 必须支持（影响核心功能）

1. **工具调用基础支持**
   - OpenAI: `tools`, `tool_choice`, `message.tool_calls`
   - Anthropic: `tools`, `content[].type="tool_use"`
   - Gemini: `tools.function_declarations`, `parts[].functionCall`

2. **多模态输入（Vision）**
   - OpenAI: `content` 数组，`type: "image_url"`
   - Anthropic: `content[].type="image"`
   - Gemini: `parts[].inline_data`

3. **流式响应中的工具调用**
   - 确保工具调用在流式模式下正确传输

### 🟡 P1 - 应该支持（增强功能）

4. **高级采样参数**
   - `top_p`, `top_k`, `stop_sequences`
   - `presence_penalty`, `frequency_penalty`（OpenAI）

5. **JSON 模式**
   - OpenAI: `response_format: {"type": "json_object"}`
   - Gemini: `responseMimeType: "application/json"`

6. **多个响应生成**
   - OpenAI: `n` 参数

### 🟢 P2 - 可选支持（完善功能）

7. **安全设置**
   - Gemini: `safetySettings`

8. **元数据和用户标识**
   - OpenAI: `user`
   - Anthropic: `metadata`

9. **确定性输出**
   - OpenAI: `seed`

---

## 6. 数据结构修改建议

### Message 结构扩展

```go
type Message struct {
    Role      string        `json:"role"`
    Content   interface{}   `json:"content"` // string or []ContentPart
    Name      string        `json:"name,omitempty"`
    ToolCalls []ToolCall    `json:"tool_calls,omitempty"`
    ToolCallID string       `json:"tool_call_id,omitempty"`
}

type ContentPart struct {
    Type     string        `json:"type"` // text, image_url, image, inline_data
    Text     string        `json:"text,omitempty"`
    ImageURL *ImageURL     `json:"image_url,omitempty"`
    Source   *ImageSource  `json:"source,omitempty"`
    InlineData *InlineData `json:"inline_data,omitempty"`
}

type ImageURL struct {
    URL    string `json:"url"`
    Detail string `json:"detail,omitempty"` // low, high, auto
}

type ImageSource struct {
    Type      string `json:"type"` // base64, url
    MediaType string `json:"media_type"`
    Data      string `json:"data,omitempty"`
    URL       string `json:"url,omitempty"`
}

type InlineData struct {
    MimeType string `json:"mime_type"`
    Data     string `json:"data"`
}

type ToolCall struct {
    ID       string       `json:"id"`
    Type     string       `json:"type"` // function
    Function FunctionCall `json:"function"`
}

type FunctionCall struct {
    Name      string `json:"name"`
    Arguments string `json:"arguments"` // JSON string
}
```

### ChatRequest 结构扩展

```go
type ChatRequest struct {
    Model            string        `json:"model"`
    Messages         []Message     `json:"messages"`
    Temperature      float64       `json:"temperature,omitempty"`
    TopP             float64       `json:"top_p,omitempty"`
    TopK             int           `json:"top_k,omitempty"`
    MaxTokens        int           `json:"max_tokens,omitempty"`
    Stream           bool          `json:"stream,omitempty"`
    Stop             interface{}   `json:"stop,omitempty"` // string or []string
    N                int           `json:"n,omitempty"`
    PresencePenalty  float64       `json:"presence_penalty,omitempty"`
    FrequencyPenalty float64       `json:"frequency_penalty,omitempty"`
    LogitBias        map[string]int `json:"logit_bias,omitempty"`
    User             string        `json:"user,omitempty"`
    ResponseFormat   *ResponseFormat `json:"response_format,omitempty"`
    Seed             *int          `json:"seed,omitempty"`
    Tools            []Tool        `json:"tools,omitempty"`
    ToolChoice       interface{}   `json:"tool_choice,omitempty"` // string or object
}

type Tool struct {
    Type     string        `json:"type"` // function
    Function ToolFunction  `json:"function"`
}

type ToolFunction struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description,omitempty"`
    Parameters  map[string]interface{} `json:"parameters"`
}

type ResponseFormat struct {
    Type string `json:"type"` // text, json_object
}
```

---

## 7. 测试用例

### 工具调用测试

```bash
# OpenAI 格式
curl http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "What is the weather in Boston?"}],
    "tools": [{
      "type": "function",
      "function": {
        "name": "get_weather",
        "description": "Get weather",
        "parameters": {
          "type": "object",
          "properties": {
            "location": {"type": "string"}
          },
          "required": ["location"]
        }
      }
    }]
  }'

# Anthropic 格式
curl http://localhost:8080/v1/messages \
  -H "x-api-key: YOUR_API_KEY" \
  -H "anthropic-version: 2023-06-01" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-opus-20240229",
    "max_tokens": 1024,
    "tools": [{
      "name": "get_weather",
      "description": "Get weather",
      "input_schema": {
        "type": "object",
        "properties": {
          "location": {"type": "string"}
        },
        "required": ["location"]
      }
    }],
    "messages": [{"role": "user", "content": "What is the weather in Boston?"}]
  }'
```

### Vision 测试

```bash
# OpenAI Vision
curl http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4-vision-preview",
    "messages": [{
      "role": "user",
      "content": [
        {"type": "text", "text": "What is in this image?"},
        {"type": "image_url", "image_url": {"url": "https://example.com/image.jpg"}}
      ]
    }],
    "max_tokens": 300
  }'
```

---

## 参考文档

- [OpenAI Function Calling](https://platform.openai.com/docs/guides/function-calling)
- [OpenAI Vision](https://platform.openai.com/docs/guides/vision)
- [Anthropic Tool Use](https://docs.anthropic.com/en/docs/tool-use)
- [Anthropic Vision](https://docs.anthropic.com/en/docs/vision)
- [Gemini Function Calling](https://ai.google.dev/gemini-api/docs/function-calling)
- [Gemini Multimodal](https://ai.google.dev/gemini-api/docs/vision)
