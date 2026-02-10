# Prism API 实现待办清单

## 当前状态总结

### ✅ 已完成
1. 基础的三家厂商 API 适配器（OpenAI、Anthropic、Gemini）
2. 基本的请求代理和响应转换
3. 用户认证和 API 密钥管理
4. 负载均衡策略实现（轮询、加权、最少连接、随机）
5. 基础的流式响应支持（OpenAI、Anthropic）
6. 管理员用户管理和 API 配置管理
7. **工具调用（Function Calling）完整支持** ✅
   - OpenAI 格式：`tools`, `tool_choice`, `tool_calls`
   - Anthropic 格式：`tools`, `tool_use`, `tool_result`
   - Gemini 格式：`function_declarations`, `functionCall`
8. **负载均衡器配置管理 API** ✅
9. **响应格式完全对齐官方标准** ✅
   - OpenAI: `object`, `created`, `finish_reason`
   - Anthropic: 保持原始响应格式
   - Gemini: 保持原始响应格式

### ⚠️ 部分完成
1. 响应格式对齐（缺少部分字段）
2. 流式响应（Gemini 未完全实现）
3. 错误响应格式（未完全对齐官方标准）

### ❌ 未完成
1. 多模态输入（Vision）支持
2. 高级采样参数完整支持（部分已实现）
3. JSON 模式支持
4. 健康检查和监控
5. 使用历史记录 API

---

## 实现优先级和计划

### 🔴 P0 - 核心功能缺失（必须立即修复）

#### 1. 响应格式完全对齐官方标准 ✅ 已完成
**影响：** 使用官方 SDK 时会出错

**任务：**
- [x] 更新 `adapter.go` 中的 `ChatResponse` 和 `ChatChoice` 结构
- [x] 更新 `openai_adapter.go` 的 `convertResponse` 方法
- [x] 更新 `anthropic_adapter.go` 的 `convertResponse` 方法
- [x] 更新 `gemini_adapter.go` 的 `convertResponse` 方法
- [x] 确保 `proxy_handler.go` 不做额外的格式转换

**文件：**
- `backend/internal/adapter/adapter.go` ✅
- `backend/internal/adapter/openai_adapter.go` ✅
- `backend/internal/adapter/anthropic_adapter.go` ✅
- `backend/internal/adapter/gemini_adapter.go` ✅
- `backend/internal/api/proxy_handler.go` ✅

#### 2. 工具调用（Function Calling）支持 ✅ 已完成
**影响：** 无法使用 Agent 和工具调用功能

**任务：**
- [x] 创建 `types.go` 定义扩展类型
- [x] 更新 `proxy_handler.go` 的请求/响应结构
- [x] 更新三个适配器支持工具调用
  - [x] `openai_adapter.go`: 支持 `tools`, `tool_choice`, `tool_calls`
  - [x] `anthropic_adapter.go`: 支持 `tools`, `content[].type="tool_use"`
  - [x] `gemini_adapter.go`: 支持 `tools.function_declarations`, `functionCall`
- [x] 更新 `proxy_handler.go` 的转换逻辑
- [x] 清理过时的类型定义和转换函数

**文件：**
- `backend/internal/adapter/types.go` ✅
- `backend/internal/adapter/adapter.go` ✅
- `backend/internal/adapter/openai_adapter.go` ✅
- `backend/internal/adapter/anthropic_adapter.go` ✅
- `backend/internal/adapter/gemini_adapter.go` ✅
- `backend/internal/api/proxy_handler.go` ✅

#### 3. 负载均衡器配置管理 API ✅ 已完成
**影响：** 无法处理图片输入

**任务：**
- [ ] 更新 `Message` 结构支持 `content` 数组
- [ ] OpenAI: 支持 `type: "image_url"`
- [ ] Anthropic: 支持 `type: "image"` 和 `source`
- [ ] Gemini: 支持 `inline_data`
- [ ] 更新适配器的消息转换逻辑
- [ ] 测试图片输入

**文件：**
- `backend/internal/adapter/types.go` ✅（已定义）
- `backend/internal/adapter/openai_adapter.go`
- `backend/internal/adapter/anthropic_adapter.go`
- `backend/internal/adapter/gemini_adapter.go`

#### 4. 多模态输入（Vision）支持
**影响：** 前端管理页面无法使用

**任务：**
- [x] 创建 `load_balancer_handler.go`
- [x] 实现以下端点：
  - `GET /api/admin/load-balancer/configs`
  - `GET /api/admin/load-balancer/models/:model/endpoints`
  - `POST /api/admin/load-balancer/configs`
  - `PUT /api/admin/load-balancer/configs/:id`
  - `DELETE /api/admin/load-balancer/configs/:id`
  - `GET /api/admin/models`
- [x] 创建 `load_balancer_service.go`
- [x] 创建 `load_balancer_repository.go`
- [x] 在 `main.go` 中注册路由
- [x] 创建单元测试

**文件：**
- `backend/internal/api/load_balancer_handler.go` ✅
- `backend/internal/service/load_balancer_service.go` ✅
- `backend/internal/service/load_balancer_service_test.go` ✅
- `backend/internal/repository/load_balancer_repository.go` ✅
- `backend/cmd/server/main.go` ✅

---

### 🟡 P1 - 重要功能（应尽快实现）

#### 5. 流式响应完善
**任务：**
- [ ] 验证 OpenAI 流式格式是否完全符合官方标准
- [ ] 验证 Anthropic 流式格式（event-based SSE）
- [ ] 完善 Gemini 流式支持（使用 `:streamGenerateContent`）
- [ ] 支持流式响应中的工具调用

**文件：**
- `backend/internal/adapter/openai_adapter.go`
- `backend/internal/adapter/anthropic_adapter.go`
- `backend/internal/adapter/gemini_adapter.go`
- `backend/internal/api/proxy_handler.go`

#### 6. 高级采样参数支持
**任务：**
- [ ] 支持 `top_p`, `top_k`
- [ ] 支持 `stop` / `stop_sequences`
- [ ] 支持 `presence_penalty`, `frequency_penalty`（OpenAI）
- [ ] 支持 `n` 参数（多个响应）
- [ ] 更新适配器传递这些参数

**文件：**
- `backend/internal/adapter/types.go` ✅（已定义）
- `backend/internal/adapter/openai_adapter.go`
- `backend/internal/adapter/anthropic_adapter.go`
- `backend/internal/adapter/gemini_adapter.go`

#### 7. JSON 模式支持
**任务：**
- [ ] OpenAI: 支持 `response_format: {"type": "json_object"}`
- [ ] Gemini: 支持 `responseMimeType: "application/json"`
- [ ] 更新适配器传递这些参数

**文件：**
- `backend/internal/adapter/openai_adapter.go`
- `backend/internal/adapter/gemini_adapter.go`

#### 8. 错误响应格式对齐
**任务：**
- [ ] OpenAI 错误格式：`{"error": {"message": "...", "type": "...", "code": "..."}}`
- [ ] Anthropic 错误格式：`{"type": "error", "error": {"type": "...", "message": "..."}}`
- [ ] Gemini 错误格式：`{"error": {"code": 400, "message": "...", "status": "..."}}`
- [ ] 更新 `proxy_handler.go` 的错误处理

**文件：**
- `backend/internal/api/proxy_handler.go`

#### 9. 使用历史记录 API
**任务：**
- [ ] 实现 `GET /api/user/usage-history` 端点
- [ ] 返回用户的使用趋势数据
- [ ] 支持按时间范围、模型筛选

**文件：**
- `backend/internal/api/quota_handler.go`
- `backend/internal/service/quota_service.go`

---

### 🟢 P2 - 增强功能（可以后续实现）

#### 10. 健康检查和监控
**任务：**
- [ ] API 配置的健康状态检查
- [ ] 响应时间统计
- [ ] 成功率统计
- [ ] 定期健康检查任务

**文件：**
- `backend/internal/service/health_check_service.go`（新建）
- `backend/internal/api/health_check_handler.go`（新建）

#### 11. API 密钥增强功能
**任务：**
- [ ] 支持更新 API 密钥（名称、限流）
- [ ] 支持启用/禁用 API 密钥
- [ ] API 密钥使用统计

**文件：**
- `backend/internal/api/api_key_handler.go`
- `backend/internal/service/api_key_service.go`

#### 12. 配额管理增强
**任务：**
- [ ] 配额使用详细记录
- [ ] 配额预警机制
- [ ] 配额使用趋势分析

**文件：**
- `backend/internal/service/quota_service.go`

#### 13. 安全设置支持
**任务：**
- [ ] Gemini: 支持 `safetySettings`
- [ ] 返回 `safetyRatings`

**文件：**
- `backend/internal/adapter/gemini_adapter.go`

#### 14. 元数据和用户标识
**任务：**
- [ ] OpenAI: 支持 `user` 字段
- [ ] Anthropic: 支持 `metadata` 字段
- [ ] 记录到日志中

**文件：**
- `backend/internal/adapter/openai_adapter.go`
- `backend/internal/adapter/anthropic_adapter.go`

#### 15. 确定性输出
**任务：**
- [ ] OpenAI: 支持 `seed` 参数

**文件：**
- `backend/internal/adapter/openai_adapter.go`

---

## 测试计划

### 单元测试
- [ ] 适配器测试（请求/响应转换）
- [ ] 服务层测试
- [ ] 负载均衡器测试

### 集成测试
- [ ] 使用官方 SDK 测试（OpenAI、Anthropic、Gemini）
- [ ] 工具调用端到端测试
- [ ] Vision 端到端测试
- [ ] 流式响应测试

### 兼容性测试
- [ ] 使用 OpenAI Python SDK
- [ ] 使用 Anthropic Python SDK
- [ ] 使用 Google Gemini SDK
- [ ] 验证所有响应字段

---

## 文档更新

- [x] `docs/API_COMPATIBILITY.md` - API 兼容性分析
- [x] `docs/ADVANCED_FEATURES.md` - 高级功能说明
- [x] `docs/IMPLEMENTATION_TODO.md` - 实现待办清单
- [ ] `docs/API.md` - 更新 API 文档，添加工具调用和 Vision 示例
- [ ] `README.md` - 更新功能列表

---

## 开发建议

### 开发顺序
1. **先修复响应格式**（P0-1）- 确保基础兼容性
2. **实现工具调用**（P0-2）- 核心功能
3. **实现多模态输入**（P0-3）- 核心功能
4. **补充负载均衡器 API**（P0-4）- 前端依赖
5. **完善流式响应**（P1-5）- 重要功能
6. **其他 P1 功能**
7. **P2 增强功能**

### 测试策略
- 每完成一个功能，立即编写测试
- 使用官方 SDK 进行集成测试
- 对比官方 API 的响应格式

### 代码审查重点
- 确保所有字段名和类型与官方文档一致
- 确保错误处理符合各厂商的标准
- 确保流式响应格式正确

---

## 参考资源

### 官方文档
- [OpenAI API Reference](https://platform.openai.com/docs/api-reference)
- [Anthropic API Reference](https://docs.anthropic.com/en/api/messages)
- [Google Gemini API Reference](https://ai.google.dev/gemini-api/docs)

### 官方 SDK
- [OpenAI Python SDK](https://github.com/openai/openai-python)
- [Anthropic Python SDK](https://github.com/anthropics/anthropic-sdk-python)
- [Google Generative AI Python SDK](https://github.com/google/generative-ai-python)

### 测试工具
- Postman/Insomnia（API 测试）
- curl（命令行测试）
- 官方 SDK（兼容性测试）
