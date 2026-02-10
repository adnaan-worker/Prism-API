# API 兼容性分析与修复建议

## 概述

本文档分析 Prism API 对外接口与 OpenAI、Anthropic、Gemini 官方 API 的兼容性，并提供修复建议。

## 当前实现问题

### 1. OpenAI 格式 (`/v1/chat/completions`)

#### ✅ 请求格式 - 已对齐
```json
{
  "model": "gpt-4",
  "messages": [{"role": "user", "content": "Hello"}],
  "temperature": 0.7,
  "max_tokens": 100,
  "stream": false
}
```

#### ⚠️ 响应格式 - 部分缺失
**官方标准响应：**
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
      "content": "Hello! How can I help you?"
    },
    "finish_reason": "stop"
  }],
  "usage": {
    "prompt_tokens": 9,
    "completion_tokens": 12,
    "total_tokens": 21
  }
}
```

**当前实现缺失字段：**
- ❌ `object` 字段（应为 "chat.completion"）
- ❌ `created` 字段（Unix 时间戳）
- ❌ `finish_reason` 字段（stop/length/content_filter/tool_calls）

#### ⚠️ 流式响应格式 - 需要验证
**官方 SSE 格式：**
```
data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1677652288,"model":"gpt-4","choices":[{"index":0,"delta":{"content":"Hello"},"finish_reason":null}]}

data: [DONE]
```

**需要确认：**
- 是否包含 `object: "chat.completion.chunk"`
- 是否使用 `delta` 而不是 `message`
- 是否以 `data: [DONE]` 结束

---

### 2. Anthropic 格式 (`/v1/messages`)

#### ✅ 请求格式 - 已对齐
```json
{
  "model": "claude-3-opus-20240229",
  "max_tokens": 1024,
  "messages": [{"role": "user", "content": "Hello"}],
  "system": "You are a helpful assistant",
  "temperature": 0.7,
  "stream": false
}
```

**请求头要求：**
- ✅ `x-api-key` 或 `Authorization: Bearer`
- ✅ `anthropic-version: 2023-06-01`
- ✅ `content-type: application/json`

#### ⚠️ 响应格式 - 部分缺失
**官方标准响应：**
```json
{
  "id": "msg_01XFDUDYJgAACzvnptvVoYEL",
  "type": "message",
  "role": "assistant",
  "content": [
    {
      "type": "text",
      "text": "Hello! How can I help you today?"
    }
  ],
  "model": "claude-3-opus-20240229",
  "stop_reason": "end_turn",
  "stop_sequence": null,
  "usage": {
    "input_tokens": 10,
    "output_tokens": 20
  }
}
```

**当前实现问题：**
- ✅ `type` 字段正确
- ✅ `content` 数组格式正确
- ❌ 缺少 `stop_sequence` 字段
- ⚠️ `stop_reason` 可能值不完整（应包含：end_turn, max_tokens, stop_sequence, tool_use）

#### ⚠️ 流式响应格式 - 需要验证
**官方 SSE 格式：**
```
event: message_start
data: {"type":"message_start","message":{"id":"msg_123","type":"message","role":"assistant","content":[],"model":"claude-3-opus-20240229","stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":10,"output_tokens":0}}}

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello"}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"end_turn","stop_sequence":null},"usage":{"output_tokens":20}}

event: message_stop
data: {"type":"message_stop"}
```

**需要确认：**
- Anthropic 使用 `event:` 和 `data:` 的 SSE 格式
- 需要多个事件类型（message_start, content_block_delta, message_stop 等）

---

### 3. Gemini 格式 (`/v1/models/{model}:generateContent`)

#### ✅ 请求格式 - 已对齐
```json
{
  "contents": [
    {
      "role": "user",
      "parts": [{"text": "Hello"}]
    }
  ],
  "generationConfig": {
    "temperature": 0.7,
    "maxOutputTokens": 100
  }
}
```

**URL 格式：**
- ✅ `/v1/models/{model}:generateContent?key={api_key}`

#### ⚠️ 响应格式 - 部分缺失
**官方标准响应：**
```json
{
  "candidates": [
    {
      "content": {
        "parts": [{"text": "Hello! How can I help you?"}],
        "role": "model"
      },
      "finishReason": "STOP",
      "index": 0,
      "safetyRatings": [
        {
          "category": "HARM_CATEGORY_SEXUALLY_EXPLICIT",
          "probability": "NEGLIGIBLE"
        }
      ]
    }
  ],
  "usageMetadata": {
    "promptTokenCount": 5,
    "candidatesTokenCount": 10,
    "totalTokenCount": 15
  },
  "modelVersion": "gemini-1.5-pro-001"
}
```

**当前实现缺失字段：**
- ❌ `safetyRatings` 数组（安全评级）
- ❌ `modelVersion` 字段
- ⚠️ `finishReason` 可能值不完整（应包含：STOP, MAX_TOKENS, SAFETY, RECITATION, OTHER）

#### ⚠️ 流式响应格式 - 需要验证
**官方 SSE 格式：**
```
data: {"candidates":[{"content":{"parts":[{"text":"Hello"}],"role":"model"},"finishReason":"STOP","index":0}],"usageMetadata":{"promptTokenCount":5,"candidatesTokenCount":1,"totalTokenCount":6}}

data: {"candidates":[{"content":{"parts":[{"text":"!"}],"role":"model"},"finishReason":"STOP","index":0}],"usageMetadata":{"promptTokenCount":5,"candidatesTokenCount":2,"totalTokenCount":7}}
```

**需要确认：**
- 流式 URL 应为 `/v1/models/{model}:streamGenerateContent?key={api_key}&alt=sse`
- 每个 chunk 都是完整的 JSON 对象

---

## 修复优先级

### 🔴 高优先级（影响兼容性）

1. **OpenAI 响应格式补全**
   - 添加 `object` 字段
   - 添加 `created` 字段
   - 添加 `finish_reason` 字段

2. **Anthropic 流式响应格式**
   - 实现完整的事件流格式（message_start, content_block_delta 等）
   - 当前可能只是直接透传，需要验证

3. **Gemini 流式响应 URL**
   - 确认使用 `:streamGenerateContent` 而不是 `:generateContent`

### 🟡 中优先级（增强兼容性）

4. **错误响应格式对齐**
   - OpenAI: `{"error": {"message": "...", "type": "...", "code": "..."}}`
   - Anthropic: `{"type": "error", "error": {"type": "...", "message": "..."}}`
   - Gemini: `{"error": {"code": 400, "message": "...", "status": "INVALID_ARGUMENT"}}`

5. **添加缺失的可选字段**
   - Anthropic: `stop_sequence`
   - Gemini: `safetyRatings`, `modelVersion`

### 🟢 低优先级（完善功能）

6. **支持更多请求参数**
   - OpenAI: `top_p`, `n`, `presence_penalty`, `frequency_penalty`
   - Anthropic: `top_p`, `top_k`, `metadata`
   - Gemini: `topP`, `topK`, `stopSequences`, `safetySettings`

---

## 建议的修复步骤

### Step 1: 修复 OpenAI 响应格式

修改 `backend/internal/adapter/openai_adapter.go`:

```go
func (a *OpenAIAdapter) convertResponse(resp *openAIResponse) *ChatResponse {
    choices := make([]ChatChoice, len(resp.Choices))
    for i, choice := range resp.Choices {
        choices[i] = ChatChoice{
            Index:        choice.Index,
            Message:      choice.Message,
            FinishReason: choice.FinishReason, // 添加此字段
        }
    }

    return &ChatResponse{
        ID:      resp.ID,
        Object:  "chat.completion", // 添加此字段
        Created: resp.Created,      // 添加此字段
        Model:   resp.Model,
        Choices: choices,
        Usage: UsageInfo{
            PromptTokens:     resp.Usage.PromptTokens,
            CompletionTokens: resp.Usage.CompletionTokens,
            TotalTokens:      resp.Usage.TotalTokens,
        },
    }
}
```

### Step 2: 修复 proxy_handler.go 的响应转换

确保 `ChatCompletions` 处理器直接返回 OpenAI 格式，不做额外转换。

### Step 3: 验证流式响应格式

测试三家厂商的流式响应是否符合官方格式。

### Step 4: 统一错误响应格式

为每个厂商实现符合其标准的错误响应格式。

---

## 测试建议

### 1. 兼容性测试

使用官方 SDK 测试：

```python
# OpenAI SDK
import openai
client = openai.OpenAI(
    api_key="your-prism-api-key",
    base_url="http://localhost:8080/v1"
)
response = client.chat.completions.create(
    model="gpt-4",
    messages=[{"role": "user", "content": "Hello"}]
)

# Anthropic SDK
import anthropic
client = anthropic.Anthropic(
    api_key="your-prism-api-key",
    base_url="http://localhost:8080"
)
message = client.messages.create(
    model="claude-3-opus-20240229",
    max_tokens=1024,
    messages=[{"role": "user", "content": "Hello"}]
)

# Google Gemini SDK
import google.generativeai as genai
genai.configure(
    api_key="your-prism-api-key",
    transport="rest",
    client_options={"api_endpoint": "http://localhost:8080"}
)
model = genai.GenerativeModel('gemini-pro')
response = model.generate_content("Hello")
```

### 2. 字段验证测试

编写测试用例验证所有必需字段是否存在。

### 3. 流式响应测试

测试 SSE 流式响应的格式和事件顺序。

---

## 参考文档

- [OpenAI Chat Completions API](https://platform.openai.com/docs/api-reference/chat/create)
- [Anthropic Messages API](https://docs.anthropic.com/en/api/messages)
- [Google Gemini API](https://ai.google.dev/gemini-api/docs/text-generation)
