package adapter

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// KiroModelMapper 接口用于获取模型映射
type KiroModelMapper interface {
	GetModelMapping(ctx context.Context, modelName string) (string, error)
}

// KiroAdapter implements the Adapter interface for Kiro API (AWS CodeWhisperer)
// Kiro provides Claude models through AWS CodeWhisperer
// This adapter converts between OpenAI/Anthropic/Gemini formats and Kiro's native format
type KiroAdapter struct {
	config       *Config
	accessToken  string
	profileArn   string // Required for Social Auth
	region       string
	machineID    string
	modelMapper  KiroModelMapper
}

// NewKiroAdapter creates a new Kiro adapter
func NewKiroAdapter(config *Config, accessToken, profileArn, region string, modelMapper KiroModelMapper) *KiroAdapter {
	if config.Client == nil {
		timeout := 120 * time.Second // Kiro needs longer timeout
		if config.Timeout > 0 {
			timeout = time.Duration(config.Timeout) * time.Second
		}
		config.Client = &http.Client{
			Timeout: timeout,
		}
	}

	// Generate machine ID for this session
	machineID := generateMachineID()

	return &KiroAdapter{
		config:       config,
		accessToken:  accessToken,
		profileArn:   profileArn,
		region:       region,
		machineID:    machineID,
		modelMapper:  modelMapper,
	}
}

// GetType returns the adapter type
func (a *KiroAdapter) GetType() string {
	return "kiro"
}

// Kiro request/response structures
type kiroRequest struct {
	ConversationState kiroConversationState `json:"conversationState"`
	ProfileArn        string                `json:"profileArn,omitempty"`        // Required for Social Auth
	InferenceConfig   *kiroInferenceConfig  `json:"inferenceConfig,omitempty"`   // Optional inference parameters
}

type kiroInferenceConfig struct {
	MaxTokens   int     `json:"maxTokens,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
	TopP        float64 `json:"topP,omitempty"`
}

type kiroConversationState struct {
	ConversationID  string        `json:"conversationId"`
	History         []kiroMessage `json:"history,omitempty"`
	CurrentMessage  kiroMessage   `json:"currentMessage"`
	ChatTriggerType string        `json:"chatTriggerType"`
}

type kiroMessage struct {
	UserInputMessage         *kiroUserMessage      `json:"userInputMessage,omitempty"`
	AssistantResponseMessage *kiroAssistantMessage `json:"assistantResponseMessage,omitempty"`
}

type kiroUserMessage struct {
	Content                 string              `json:"content"`
	ModelID                 string              `json:"modelId"`                           // Required: Kiro model ID
	Origin                  string              `json:"origin"`                            // Required: "AI_EDITOR"
	UserInputMessageContext *kiroMessageContext `json:"userInputMessageContext,omitempty"` // Optional: tools and tool results
}

type kiroAssistantMessage struct {
	Content  string         `json:"content"`
	ToolUses []kiroToolUse  `json:"toolUses,omitempty"`
}

type kiroMessageContext struct {
	Tools       []kiroTool       `json:"tools,omitempty"`
	ToolResults []kiroToolResult `json:"toolResults,omitempty"`
}

type kiroTool struct {
	ToolSpecification kiroToolSpec `json:"toolSpecification"`
}

type kiroToolSpec struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema kiroInputSchema `json:"inputSchema"`
}

type kiroInputSchema struct {
	JSON map[string]interface{} `json:"json"`
}

type kiroToolUse struct {
	Name      string                 `json:"name"`
	ToolUseID string                 `json:"toolUseId"`
	Input     map[string]interface{} `json:"input"`
}

type kiroToolResult struct {
	ToolUseID string                  `json:"toolUseId"`
	Status    string                  `json:"status"`
	Content   []kiroToolResultContent `json:"content"`
}

type kiroToolResultContent struct {
	Text string `json:"text"`
}

type kiroResponse struct {
	ConversationID           string `json:"conversationId"`
	AssistantResponseMessage string `json:"$amazonq.streaming#assistantResponseMessage,omitempty"`
	Message                  string `json:"message,omitempty"`
}

// Call makes a request to Kiro API
func (a *KiroAdapter) Call(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	// Convert unified request to Kiro format
	kiroReq, err := a.convertRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to convert request: %w", err)
	}

	// Marshal request
	reqBody, err := json.Marshal(kiroReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("https://q.%s.amazonaws.com/generateAssistantResponse", a.region)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	a.setKiroHeaders(httpReq)

	// Make request
	resp, err := a.config.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Handle gzip encoding
	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	// Read response body
	respBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("[Kiro] API error %d: %s\n", resp.StatusCode, string(respBody))
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse AWS EventStream response
	parsedContent, toolCalls, err := a.parseEventStreamChunk(respBody)
	if err != nil {
		return nil, fmt.Errorf("failed to parse EventStream response: %w", err)
	}

	// Convert to unified response
	return a.convertEventStreamResponse(parsedContent, toolCalls, req.Model), nil
}

// CallStream makes a streaming request to Kiro API
// Returns a wrapped response that converts EventStream to SSE format
func (a *KiroAdapter) CallStream(ctx context.Context, req *ChatRequest) (*http.Response, error) {
	// Convert unified request to Kiro format
	kiroReq, err := a.convertRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to convert request: %w", err)
	}

	// Marshal request
	reqBody, err := json.Marshal(kiroReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("https://q.%s.amazonaws.com/generateAssistantResponse", a.region)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers for streaming
	a.setKiroHeaders(httpReq)
	httpReq.Header.Set("Accept", "application/vnd.amazon.eventstream")

	// Make request
	resp, err := a.config.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Create a pipe to convert EventStream to SSE
	pr, pw := io.Pipe()

	// Start goroutine to parse EventStream and write SSE
	go func() {
		defer pw.Close()
		defer resp.Body.Close()

		if err := a.streamEventStreamToSSE(resp.Body, pw, req.Model); err != nil {
			pw.CloseWithError(err)
		}
	}()

	// Create a new response with the pipe as body
	streamResp := &http.Response{
		Status:        "200 OK",
		StatusCode:    200,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Body:          pr,
		ContentLength: -1,
		Header:        make(http.Header),
	}
	streamResp.Header.Set("Content-Type", "text/event-stream")
	streamResp.Header.Set("Cache-Control", "no-cache")
	streamResp.Header.Set("Connection", "keep-alive")

	return streamResp, nil
}

// convertRequest converts unified ChatRequest to Kiro format
// Following Kiro-account-manager implementation exactly
func (a *KiroAdapter) convertRequest(req *ChatRequest) (*kiroRequest, error) {
	conversationID := uuid.New().String()

	// Get Kiro model ID from database mapping
	kiroModelID, err := a.modelMapper.GetModelMapping(context.Background(), req.Model)
	if err != nil {
		// Fallback: use model name as-is if mapping not found
		kiroModelID = req.Model
	}

	origin := "AI_EDITOR"

	// Extract system prompt
	var systemPrompt string
	var nonSystemMessages []Message

	for _, msg := range req.Messages {
		if msg.Role == "system" {
			if systemPrompt != "" {
				systemPrompt += "\n"
			}
			systemPrompt += msg.Content
		} else {
			nonSystemMessages = append(nonSystemMessages, msg)
		}
	}

	// Inject timestamp
	timestamp := time.Now().Format(time.RFC3339)
	systemPrompt = fmt.Sprintf("[Context: Current time is %s]\n\n%s", timestamp, systemPrompt)

	// Inject execution directive (prevent AI from losing focus)
	executionDirective := `
<execution_discipline>
当用户要求执行特定任务时，你必须遵循以下纪律：
1. **目标锁定**：在整个会话中始终牢记用户的原始目标，不要在代码探索过程中迷失方向
2. **行动优先**：优先执行任务而非仅分析或总结，除非用户明确只要求分析
3. **计划执行**：为任务创建明确的步骤计划，逐步执行并标记完成状态
4. **禁止确认性收尾**：在任务未完成前，禁止输出"需要我继续吗？"、"需要深入分析吗？"等确认性问题
5. **持续推进**：如果发现部分任务已完成，立即继续执行剩余未完成的任务
6. **完整交付**：直到所有任务步骤都执行完毕才算完成
</execution_discipline>
`
	systemPrompt = systemPrompt + "\n\n" + executionDirective

	// Build history messages - collect all messages first
	var allMessages []kiroMessage
	var pendingToolResults []kiroToolResult
	systemPromptMerged := false

	for i, msg := range nonSystemMessages {
		if msg.Role == "user" {
			userContent := msg.Content

			// Merge system prompt into first user message
			if !systemPromptMerged && systemPrompt != "" {
				userContent = fmt.Sprintf("%s\n\n%s", systemPrompt, userContent)
				systemPromptMerged = true
			}

			if userContent == "" {
				userContent = "Continue"
			}

			allMessages = append(allMessages, kiroMessage{
				UserInputMessage: &kiroUserMessage{
					Content: userContent,
					ModelID: kiroModelID,
					Origin:  origin,
				},
			})
		} else if msg.Role == "assistant" {
			// Kiro API requires content to be non-empty
			assistantContent := msg.Content
			if assistantContent == "" && len(msg.ToolCalls) > 0 {
				assistantContent = "Using tools."
			} else if assistantContent == "" {
				assistantContent = "I understand."
			}

			var toolUses []kiroToolUse
			for _, tc := range msg.ToolCalls {
				if tc.Type == "function" {
					var input map[string]interface{}
					if err := json.Unmarshal([]byte(tc.Function.Arguments), &input); err != nil {
						input = make(map[string]interface{})
					}
					toolUses = append(toolUses, kiroToolUse{
						ToolUseID: tc.ID,
						Name:      tc.Function.Name,
						Input:     input,
					})
				}
			}

			allMessages = append(allMessages, kiroMessage{
				AssistantResponseMessage: &kiroAssistantMessage{
					Content:  assistantContent,
					ToolUses: toolUses,
				},
			})
		} else if msg.Role == "tool" {
			// Tool result - collect for processing
			if msg.ToolCallID != "" {
				pendingToolResults = append(pendingToolResults, kiroToolResult{
					ToolUseID: msg.ToolCallID,
					Content: []kiroToolResultContent{
						{Text: msg.Content},
					},
					Status: "success",
				})
			}

			// Check if next message is also a tool message
			var nextMsg *Message
			if i+1 < len(nonSystemMessages) {
				nextMsg = &nonSystemMessages[i+1]
			}
			shouldFlush := nextMsg == nil || nextMsg.Role != "tool"

			// Flush tool results as a user message if needed
			if shouldFlush && len(pendingToolResults) > 0 {
				allMessages = append(allMessages, kiroMessage{
					UserInputMessage: &kiroUserMessage{
						Content: "Tool results provided.",
						ModelID: kiroModelID,
						Origin:  origin,
						UserInputMessageContext: &kiroMessageContext{
							ToolResults: pendingToolResults,
						},
					},
				})
				pendingToolResults = nil
			}
		}
	}

	// Sanitize conversation (ensure proper alternation and structure)
	sanitized := sanitizeConversation(allMessages, kiroModelID, origin)

	// Split into history and currentMessage
	// currentMessage is the last message, history is everything before
	var history []kiroMessage
	var currentMsg kiroMessage

	if len(sanitized) > 0 {
		history = sanitized[:len(sanitized)-1]
		currentMsg = sanitized[len(sanitized)-1]
	} else {
		// No messages - create a default current message
		currentMsg = kiroMessage{
			UserInputMessage: &kiroUserMessage{
				Content: "Continue.",
				ModelID: kiroModelID,
				Origin:  origin,
			},
		}
	}

	// Ensure currentMessage is a user message (required by Kiro API)
	if currentMsg.UserInputMessage == nil {
		// Last message is assistant - add a Continue message
		history = append(history, currentMsg)
		currentMsg = kiroMessage{
			UserInputMessage: &kiroUserMessage{
				Content: "Continue.",
				ModelID: kiroModelID,
				Origin:  origin,
			},
		}
	}

	// If system prompt not merged yet, add to current message
	if !systemPromptMerged && systemPrompt != "" {
		currentMsg.UserInputMessage.Content = fmt.Sprintf("%s\n\n%s", systemPrompt, currentMsg.UserInputMessage.Content)
	}

	// Convert tools
	var kiroTools []kiroTool
	if len(req.Tools) > 0 {
		kiroTools = make([]kiroTool, 0, len(req.Tools))
		for _, tool := range req.Tools {
			description := tool.Function.Description
			if description == "" {
				description = fmt.Sprintf("Tool: %s", tool.Function.Name)
			}
			// Truncate long descriptions
			const maxDescLen = 10237
			if len(description) > maxDescLen {
				description = description[:maxDescLen] + "..."
			}

			kiroTools = append(kiroTools, kiroTool{
				ToolSpecification: kiroToolSpec{
					Name:        shortenToolName(tool.Function.Name),
					Description: description,
					InputSchema: kiroInputSchema{
						JSON: tool.Function.Parameters,
					},
				},
			})
		}
	}

	// Add tools to current message context (tools only go in currentMessage)
	if len(kiroTools) > 0 {
		if currentMsg.UserInputMessage.UserInputMessageContext == nil {
			currentMsg.UserInputMessage.UserInputMessageContext = &kiroMessageContext{}
		}
		currentMsg.UserInputMessage.UserInputMessageContext.Tools = kiroTools
	}

	// Build request
	kiroReq := &kiroRequest{
		ConversationState: kiroConversationState{
			ConversationID:  conversationID,
			History:         history,
			CurrentMessage:  currentMsg,
			ChatTriggerType: "MANUAL",
		},
	}

	// Add profileArn for Social Auth (required)
	if a.profileArn != "" {
		kiroReq.ProfileArn = a.profileArn
	}

	// Add inference config
	if req.MaxTokens > 0 || req.Temperature > 0 || req.TopP > 0 {
		kiroReq.InferenceConfig = &kiroInferenceConfig{}
		if req.MaxTokens > 0 {
			kiroReq.InferenceConfig.MaxTokens = req.MaxTokens
		}
		if req.Temperature > 0 {
			kiroReq.InferenceConfig.Temperature = req.Temperature
		}
		if req.TopP > 0 {
			kiroReq.InferenceConfig.TopP = req.TopP
		}
	}

	return kiroReq, nil
}

// sanitizeConversation ensures proper message alternation and structure
// Following Kiro-account-manager implementation
func sanitizeConversation(messages []kiroMessage, modelID, origin string) []kiroMessage {
	if len(messages) == 0 {
		return []kiroMessage{
			{
				UserInputMessage: &kiroUserMessage{
					Content: "Hello",
					ModelID: modelID,
					Origin:  origin,
				},
			},
		}
	}

	// Step 1: Ensure starts with user message
	sanitized := ensureStartsWithUserMessage(messages, modelID, origin)

	// Step 2: Remove empty user messages (except first)
	sanitized = removeEmptyUserMessages(sanitized)

	// Step 3: Ensure valid tool uses and results
	sanitized = ensureValidToolUsesAndResults(sanitized, modelID, origin)

	// Step 4: Ensure alternating messages
	sanitized = ensureAlternatingMessages(sanitized, modelID, origin)

	// Step 5: Ensure ends with user message
	sanitized = ensureEndsWithUserMessage(sanitized, modelID, origin)

	return sanitized
}

// ensureStartsWithUserMessage ensures conversation starts with user message
func ensureStartsWithUserMessage(messages []kiroMessage, modelID, origin string) []kiroMessage {
	if len(messages) == 0 || messages[0].UserInputMessage != nil {
		return messages
	}

	// Prepend a hello message
	hello := kiroMessage{
		UserInputMessage: &kiroUserMessage{
			Content: "Hello",
			ModelID: modelID,
			Origin:  origin,
		},
	}
	return append([]kiroMessage{hello}, messages...)
}

// ensureEndsWithUserMessage ensures conversation ends with user message
func ensureEndsWithUserMessage(messages []kiroMessage, modelID, origin string) []kiroMessage {
	if len(messages) == 0 {
		return []kiroMessage{
			{
				UserInputMessage: &kiroUserMessage{
					Content: "Hello",
					ModelID: modelID,
					Origin:  origin,
				},
			},
		}
	}

	if messages[len(messages)-1].UserInputMessage != nil {
		return messages
	}

	// Append a continue message
	cont := kiroMessage{
		UserInputMessage: &kiroUserMessage{
			Content: "Continue",
			ModelID: modelID,
			Origin:  origin,
		},
	}
	return append(messages, cont)
}

// ensureAlternatingMessages ensures messages alternate between user and assistant
func ensureAlternatingMessages(messages []kiroMessage, modelID, origin string) []kiroMessage {
	if len(messages) <= 1 {
		return messages
	}

	result := []kiroMessage{messages[0]}
	for i := 1; i < len(messages); i++ {
		prevMsg := result[len(result)-1]
		currentMsg := messages[i]

		// Check if both are user messages
		if prevMsg.UserInputMessage != nil && currentMsg.UserInputMessage != nil {
			// Insert understood message
			result = append(result, kiroMessage{
				AssistantResponseMessage: &kiroAssistantMessage{
					Content: "understood",
				},
			})
		} else if prevMsg.AssistantResponseMessage != nil && currentMsg.AssistantResponseMessage != nil {
			// Insert continue message
			result = append(result, kiroMessage{
				UserInputMessage: &kiroUserMessage{
					Content: "Continue",
					ModelID: modelID,
					Origin:  origin,
				},
			})
		}

		result = append(result, currentMsg)
	}

	return result
}

// ensureValidToolUsesAndResults ensures tool uses have corresponding results
func ensureValidToolUsesAndResults(messages []kiroMessage, modelID, origin string) []kiroMessage {
	result := make([]kiroMessage, 0, len(messages))

	for i := 0; i < len(messages); i++ {
		msg := messages[i]
		result = append(result, msg)

		// Check if this is an assistant message with tool uses
		if msg.AssistantResponseMessage != nil && len(msg.AssistantResponseMessage.ToolUses) > 0 {
			// Check if next message has tool results
			var nextMsg *kiroMessage
			if i+1 < len(messages) {
				nextMsg = &messages[i+1]
			}

			hasToolResults := nextMsg != nil &&
				nextMsg.UserInputMessage != nil &&
				nextMsg.UserInputMessage.UserInputMessageContext != nil &&
				len(nextMsg.UserInputMessage.UserInputMessageContext.ToolResults) > 0

			if !hasToolResults {
				// No tool results - add failed tool results
				toolUseIDs := make([]string, len(msg.AssistantResponseMessage.ToolUses))
				for j, tu := range msg.AssistantResponseMessage.ToolUses {
					toolUseIDs[j] = tu.ToolUseID
				}

				failedResults := make([]kiroToolResult, len(toolUseIDs))
				for j, id := range toolUseIDs {
					failedResults[j] = kiroToolResult{
						ToolUseID: id,
						Content: []kiroToolResultContent{
							{Text: "Tool execution failed"},
						},
						Status: "error",
					}
				}

				result = append(result, kiroMessage{
					UserInputMessage: &kiroUserMessage{
						Content: "",
						ModelID: modelID,
						Origin:  origin,
						UserInputMessageContext: &kiroMessageContext{
							ToolResults: failedResults,
						},
					},
				})
			}
		}
	}

	return result
}

// removeEmptyUserMessages removes empty user messages (except first)
func removeEmptyUserMessages(messages []kiroMessage) []kiroMessage {
	if len(messages) <= 1 {
		return messages
	}

	firstUserIdx := -1
	for i, msg := range messages {
		if msg.UserInputMessage != nil {
			firstUserIdx = i
			break
		}
	}

	result := make([]kiroMessage, 0, len(messages))
	for i, msg := range messages {
		// Keep assistant messages
		if msg.AssistantResponseMessage != nil {
			result = append(result, msg)
			continue
		}

		// Keep first user message
		if i == firstUserIdx {
			result = append(result, msg)
			continue
		}

		// Keep user messages with content or tool results
		if msg.UserInputMessage != nil {
			hasContent := strings.TrimSpace(msg.UserInputMessage.Content) != ""
			hasToolResults := msg.UserInputMessage.UserInputMessageContext != nil &&
				len(msg.UserInputMessage.UserInputMessageContext.ToolResults) > 0

			if hasContent || hasToolResults {
				result = append(result, msg)
			}
		}
	}

	return result
}

// shortenToolName shortens tool names that are too long
func shortenToolName(name string) string {
	const limit = 64
	if len(name) <= limit {
		return name
	}

	// MCP tools: mcp__server__tool -> mcp__tool
	if strings.HasPrefix(name, "mcp__") {
		lastIdx := strings.LastIndex(name, "__")
		if lastIdx > 5 {
			shortened := "mcp__" + name[lastIdx+2:]
			if len(shortened) <= limit {
				return shortened
			}
		}
	}

	return name[:limit]
}

// convertResponse converts Kiro response to unified format
func (a *KiroAdapter) convertResponse(resp *kiroResponse, model string) *ChatResponse {
	// Extract content
	content := resp.Message
	if content == "" {
		content = resp.AssistantResponseMessage
	}

	// Parse tool calls if present
	toolCalls := parseKiroToolCalls(content)

	msg := Message{
		Role:    "assistant",
		Content: content,
	}

	if len(toolCalls) > 0 {
		msg.ToolCalls = toolCalls
	}

	// Estimate token usage (Kiro doesn't provide token counts)
	promptTokens := estimateTokens(content) / 2
	completionTokens := estimateTokens(content)

	return &ChatResponse{
		ID:      fmt.Sprintf("kiro-%s", resp.ConversationID),
		Model:   model,
		Created: time.Now().Unix(),
		Choices: []ChatChoice{
			{
				Index:        0,
				Message:      msg,
				FinishReason: "stop",
			},
		},
		Usage: UsageInfo{
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      promptTokens + completionTokens,
		},
	}
}

// setKiroHeaders sets Kiro-specific headers
func (a *KiroAdapter) setKiroHeaders(req *http.Request) {
	// Generate unique invocation ID for each request
	invocationID := uuid.New().String()
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.accessToken)
	req.Header.Set("amz-sdk-invocation-id", invocationID) // Required!
	req.Header.Set("x-amzn-kiro-agent-mode", "vibe")
	req.Header.Set("x-amz-user-agent", fmt.Sprintf("aws-sdk-js/1.0.0 KiroIDE-0.8.140-%s", a.machineID))
	req.Header.Set("user-agent", "aws-sdk-js/1.0.0 ua/2.1 os/linux lang/js md/nodejs#18.0.0 api/codewhispererruntime#1.0.0 m/E")
	req.Header.Set("amz-sdk-request", "attempt=1; max=1")
	req.Header.Set("Connection", "close")
	req.Header.Set("Accept-Encoding", "gzip, deflate") // Accept gzip encoding
}

// parseEventStreamChunk parses AWS EventStream binary format response
// AWS EventStream format:
// - 4 bytes: total message length (big-endian uint32)
// - 4 bytes: headers length (big-endian uint32)
// - 4 bytes: prelude CRC (skip)
// - N bytes: headers (key-value pairs)
// - M bytes: payload (JSON)
// - 4 bytes: message CRC (skip)
func (a *KiroAdapter) parseEventStreamChunk(rawData []byte) (string, []ToolCall, error) {
	var fullContent strings.Builder
	var toolCalls []ToolCall
	var currentToolCall *ToolCall

	offset := 0
	for offset < len(rawData) {
		// Need at least 12 bytes for length headers + prelude CRC
		if offset+12 > len(rawData) {
			break
		}

		// Read total message length (big-endian)
		totalLength := int(rawData[offset])<<24 | int(rawData[offset+1])<<16 | int(rawData[offset+2])<<8 | int(rawData[offset+3])
		headersLength := int(rawData[offset+4])<<24 | int(rawData[offset+5])<<16 | int(rawData[offset+6])<<8 | int(rawData[offset+7])
		// Skip prelude CRC (4 bytes at offset+8 to offset+11)

		// Check if we have the complete message
		if offset+totalLength > len(rawData) {
			break
		}

		// Skip to payload (12 bytes for lengths+CRC + headers length)
		payloadStart := offset + 12 + headersLength
		// Payload ends 4 bytes before message end (message CRC)
		payloadEnd := offset + totalLength - 4

		if payloadStart < payloadEnd && payloadEnd <= len(rawData) {
			payload := rawData[payloadStart:payloadEnd]

			// Try to parse payload as JSON
			var eventData map[string]interface{}
			if err := json.Unmarshal(payload, &eventData); err == nil {
				// Extract the actual event data (may be nested)
				var actualEvent map[string]interface{}
				
				// Check for assistantResponseEvent
				if assistantResp, ok := eventData["assistantResponseEvent"].(map[string]interface{}); ok {
					actualEvent = assistantResp
				} else if toolUseEvt, ok := eventData["toolUseEvent"].(map[string]interface{}); ok {
					actualEvent = toolUseEvt
				} else {
					// Use the event data directly
					actualEvent = eventData
				}

				// Handle assistantResponseEvent - text content
				if content, hasContent := actualEvent["content"].(string); hasContent {
					// Skip followupPrompt events
					if followup, hasFollowup := actualEvent["followupPrompt"]; !hasFollowup || followup == nil {
						// Process content - handle escape sequences
						decodedContent := strings.ReplaceAll(content, `\n`, "\n")
						fullContent.WriteString(decodedContent)
					}
				}

				// Handle toolUseEvent - tool calls
				if name, hasName := actualEvent["name"].(string); hasName {
					if toolUseID, hasID := actualEvent["toolUseId"].(string); hasID {
						// New tool use starting
						if currentToolCall == nil || currentToolCall.ID != toolUseID {
							if currentToolCall != nil {
								// Complete previous tool call - ensure valid JSON
								if currentToolCall.Function.Arguments != "" {
									// Try to parse and reformat to ensure valid JSON
									var args map[string]interface{}
									if err := json.Unmarshal([]byte(currentToolCall.Function.Arguments), &args); err == nil {
										argsBytes, _ := json.Marshal(args)
										currentToolCall.Function.Arguments = string(argsBytes)
									} else {
										// If parsing fails, create error object
										fmt.Printf("[Kiro] Failed to parse tool input: %v, Buffer: %s\n", err, currentToolCall.Function.Arguments[:min(100, len(currentToolCall.Function.Arguments))])
										errorObj := map[string]interface{}{
											"_error":        "Tool input truncated by Kiro API (output token limit exceeded)",
											"_partialInput": currentToolCall.Function.Arguments[:min(500, len(currentToolCall.Function.Arguments))],
										}
										argsBytes, _ := json.Marshal(errorObj)
										currentToolCall.Function.Arguments = string(argsBytes)
									}
								} else {
									currentToolCall.Function.Arguments = "{}"
								}
								toolCalls = append(toolCalls, *currentToolCall)
							}
							
							currentToolCall = &ToolCall{
								ID:   toolUseID,
								Type: "function",
								Function: FunctionCall{
									Name:      name,
									Arguments: "",
								},
							}
						}

						// Accumulate input fragments
						if input, hasInput := actualEvent["input"].(string); hasInput {
							// Accumulate string fragments
							currentToolCall.Function.Arguments += input
						} else if inputObj, hasInputObj := actualEvent["input"].(map[string]interface{}); hasInputObj {
							// Complete input object provided - this is the final form
							inputBytes, _ := json.Marshal(inputObj)
							currentToolCall.Function.Arguments = string(inputBytes)
						}

						// Check if tool use is complete
						if stop, hasStop := actualEvent["stop"].(bool); hasStop && stop {
							// Validate and format arguments as JSON
							if currentToolCall.Function.Arguments != "" {
								var args map[string]interface{}
								if err := json.Unmarshal([]byte(currentToolCall.Function.Arguments), &args); err == nil {
									argsBytes, _ := json.Marshal(args)
									currentToolCall.Function.Arguments = string(argsBytes)
								} else {
									// If parsing fails, create error object
									errorObj := map[string]interface{}{
										"_error":        "Tool input truncated by Kiro API (output token limit exceeded)",
										"_partialInput": currentToolCall.Function.Arguments[:min(500, len(currentToolCall.Function.Arguments))],
									}
									argsBytes, _ := json.Marshal(errorObj)
									currentToolCall.Function.Arguments = string(argsBytes)
								}
							} else {
								currentToolCall.Function.Arguments = "{}"
							}
							toolCalls = append(toolCalls, *currentToolCall)
							currentToolCall = nil
						}
					}
				}
			}
		}

		// Move to next message
		offset += totalLength
	}

	// Add any incomplete tool call
	if currentToolCall != nil {
		// Ensure valid JSON for incomplete tool call
		if currentToolCall.Function.Arguments != "" {
			var args map[string]interface{}
			if err := json.Unmarshal([]byte(currentToolCall.Function.Arguments), &args); err == nil {
				argsBytes, _ := json.Marshal(args)
				currentToolCall.Function.Arguments = string(argsBytes)
			} else {
				// If parsing fails, wrap in empty object
				currentToolCall.Function.Arguments = "{}"
			}
		} else {
			currentToolCall.Function.Arguments = "{}"
		}
		toolCalls = append(toolCalls, *currentToolCall)
	}

	content := fullContent.String()

	// Check for bracket-format tool calls in the text
	bracketToolCalls := parseKiroToolCalls(content)
	if len(bracketToolCalls) > 0 {
		toolCalls = append(toolCalls, bracketToolCalls...)

		// Remove tool call text from response
		for _, tc := range bracketToolCalls {
			funcName := regexp.QuoteMeta(tc.Function.Name)
			pattern := regexp.MustCompile(`(?s)\[Called\s+` + funcName + `\s+with\s+args:\s*\{[^}]*(?:\{[^}]*\}[^}]*)*\}\]`)
			content = pattern.ReplaceAllString(content, "")
		}
		// Clean up extra whitespace
		content = regexp.MustCompile(`\s+`).ReplaceAllString(content, " ")
		content = strings.TrimSpace(content)
	}

	// Deduplicate tool calls
	uniqueToolCalls := deduplicateToolCalls(toolCalls)

	return content, uniqueToolCalls, nil
}

// convertEventStreamResponse converts parsed EventStream data to unified format
func (a *KiroAdapter) convertEventStreamResponse(content string, toolCalls []ToolCall, model string) *ChatResponse {
	msg := Message{
		Role:    "assistant",
		Content: content,
	}

	if len(toolCalls) > 0 {
		msg.ToolCalls = toolCalls
	}

	// Estimate token usage (Kiro doesn't provide token counts)
	promptTokens := estimateTokens(content) / 2
	completionTokens := estimateTokens(content)

	return &ChatResponse{
		ID:      fmt.Sprintf("kiro-%s", uuid.New().String()),
		Model:   model,
		Created: time.Now().Unix(),
		Choices: []ChatChoice{
			{
				Index:        0,
				Message:      msg,
				FinishReason: "stop",
			},
		},
		Usage: UsageInfo{
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      promptTokens + completionTokens,
		},
	}
}

// deduplicateToolCalls removes duplicate tool calls
func deduplicateToolCalls(toolCalls []ToolCall) []ToolCall {
	seen := make(map[string]bool)
	var unique []ToolCall

	for _, tc := range toolCalls {
		key := tc.Function.Name + "-" + tc.Function.Arguments
		if !seen[key] {
			seen[key] = true
			unique = append(unique, tc)
		}
	}

	return unique
}

// Helper functions

// generateMachineID generates a unique machine ID
func generateMachineID() string {
	return uuid.New().String()[:32]
}

// estimateTokens estimates token count from text
func estimateTokens(text string) int {
	// Rough estimate: 1 token ≈ 4 characters
	return len(text) / 4
}

// parseKiroToolCalls parses Kiro's tool call format from response
// Format: [Called function_name with args: {"arg1": "value1"}]
func parseKiroToolCalls(content string) []ToolCall {
	var toolCalls []ToolCall

	// Find all [Called ...] patterns
	start := 0
	for {
		idx := strings.Index(content[start:], "[Called")
		if idx == -1 {
			break
		}
		idx += start

		// Find matching closing bracket
		endIdx := strings.Index(content[idx:], "]")
		if endIdx == -1 {
			break
		}
		endIdx += idx

		// Extract tool call text
		toolCallText := content[idx : endIdx+1]

		// Parse function name and arguments
		// Pattern: [Called function_name with args: {...}]
		parts := strings.SplitN(toolCallText, " with args:", 2)
		if len(parts) == 2 {
			funcName := strings.TrimPrefix(parts[0], "[Called ")
			funcName = strings.TrimSpace(funcName)

			argsJSON := strings.TrimSuffix(strings.TrimSpace(parts[1]), "]")

			// Try to parse JSON arguments
			var args map[string]interface{}
			if err := json.Unmarshal([]byte(argsJSON), &args); err == nil {
				argsBytes, _ := json.Marshal(args)
				toolCalls = append(toolCalls, ToolCall{
					ID:   fmt.Sprintf("call_%d", len(toolCalls)),
					Type: "function",
					Function: FunctionCall{
						Name:      funcName,
						Arguments: string(argsBytes),
					},
				})
			}
		}

		start = endIdx + 1
	}

	return toolCalls
}

// streamEventStreamToSSE converts AWS EventStream to SSE format
func (a *KiroAdapter) streamEventStreamToSSE(eventStreamBody io.Reader, sseWriter io.Writer, model string) error {
	buffer := make([]byte, 0)
	readBuf := make([]byte, 4096)
	chunkID := 0
	
	// Track tool use state for accumulating input fragments
	type toolUseState struct {
		id        string
		name      string
		argsBuffer string
	}
	currentToolUse := make(map[string]*toolUseState) // key: toolUseId

	for {
		n, err := eventStreamBody.Read(readBuf)
		if n > 0 {
			// Append to buffer
			buffer = append(buffer, readBuf[:n]...)

			// Try to parse complete messages from buffer
			offset := 0
			for offset < len(buffer) {
				// Need at least 12 bytes for headers
				if offset+12 > len(buffer) {
					break
				}

				// Read total message length
				totalLength := int(buffer[offset])<<24 | int(buffer[offset+1])<<16 | int(buffer[offset+2])<<8 | int(buffer[offset+3])
				headersLength := int(buffer[offset+4])<<24 | int(buffer[offset+5])<<16 | int(buffer[offset+6])<<8 | int(buffer[offset+7])

				// Check if we have the complete message
				if offset+totalLength > len(buffer) {
					break
				}

				// Extract payload
				payloadStart := offset + 12 + headersLength
				payloadEnd := offset + totalLength - 4

				if payloadStart < payloadEnd && payloadEnd <= len(buffer) {
					payload := buffer[payloadStart:payloadEnd]

					// Parse JSON payload
					var eventData map[string]interface{}
					if err := json.Unmarshal(payload, &eventData); err == nil {
						// Extract actual event
						var actualEvent map[string]interface{}
						if assistantResp, ok := eventData["assistantResponseEvent"].(map[string]interface{}); ok {
							actualEvent = assistantResp
						} else if toolUseEvt, ok := eventData["toolUseEvent"].(map[string]interface{}); ok {
							actualEvent = toolUseEvt
						} else {
							actualEvent = eventData
						}

						// Handle text content
						if content, hasContent := actualEvent["content"].(string); hasContent {
							// Skip followupPrompt events
							if followup, hasFollowup := actualEvent["followupPrompt"]; !hasFollowup || followup == nil {
								// Write SSE chunk
								chunkID++
								chunk := ChatStreamChunk{
									ID:      fmt.Sprintf("chatcmpl-%d", chunkID),
									Object:  "chat.completion.chunk",
									Created: time.Now().Unix(),
									Model:   model,
									Choices: []StreamChoice{
										{
											Index: 0,
											Delta: StreamDelta{
												Content: content,
											},
											FinishReason: "",
										},
									},
								}

								chunkJSON, _ := json.Marshal(chunk)
								fmt.Fprintf(sseWriter, "data: %s\n\n", string(chunkJSON))
							}
						}

						// Handle tool use events
						if name, hasName := actualEvent["name"].(string); hasName {
							if toolUseID, hasID := actualEvent["toolUseId"].(string); hasID {
								// Initialize or get existing tool use state
								if _, exists := currentToolUse[toolUseID]; !exists {
									currentToolUse[toolUseID] = &toolUseState{
										id:         toolUseID,
										name:       name,
										argsBuffer: "",
									}
								}
								
								state := currentToolUse[toolUseID]
								
								// Accumulate input fragments
								if input, hasInput := actualEvent["input"].(string); hasInput {
									state.argsBuffer += input
								} else if inputObj, hasInputObj := actualEvent["input"].(map[string]interface{}); hasInputObj {
									// Complete input object provided
									inputBytes, _ := json.Marshal(inputObj)
									state.argsBuffer = string(inputBytes)
								}
								
								// Check if tool use is complete
								if stop, hasStop := actualEvent["stop"].(bool); hasStop && stop {
									// Tool use complete - validate and send
									var finalArgs string
									if state.argsBuffer != "" {
										// Try to parse and validate JSON
										var args map[string]interface{}
										if err := json.Unmarshal([]byte(state.argsBuffer), &args); err == nil {
											argsBytes, _ := json.Marshal(args)
											finalArgs = string(argsBytes)
										} else {
											// Parsing failed - create error object
											errorObj := map[string]interface{}{
												"_error":        "Tool input parsing failed",
												"_partialInput": state.argsBuffer[:min(500, len(state.argsBuffer))],
											}
											argsBytes, _ := json.Marshal(errorObj)
											finalArgs = string(argsBytes)
										}
									} else {
										finalArgs = "{}"
									}

									chunkID++
									chunk := ChatStreamChunk{
										ID:      fmt.Sprintf("chatcmpl-%d", chunkID),
										Object:  "chat.completion.chunk",
										Created: time.Now().Unix(),
										Model:   model,
										Choices: []StreamChoice{
											{
												Index: 0,
												Delta: StreamDelta{
													ToolCalls: []ToolCall{
														{
															ID:   toolUseID,
															Type: "function",
															Function: FunctionCall{
																Name:      name,
																Arguments: finalArgs,
															},
														},
													},
												},
												FinishReason: "",
											},
										},
									}

									chunkJSON, _ := json.Marshal(chunk)
									fmt.Fprintf(sseWriter, "data: %s\n\n", string(chunkJSON))
									
									// Clean up completed tool use
									delete(currentToolUse, toolUseID)
								}
							}
						}
					}
				}

				// Move to next message
				offset += totalLength
			}

			// Keep remaining incomplete data in buffer
			if offset > 0 {
				buffer = buffer[offset:]
			}
		}

		if err != nil {
			if err == io.EOF {
				// Send final chunk
				finalChunk := ChatStreamChunk{
					ID:      fmt.Sprintf("chatcmpl-%d", chunkID+1),
					Object:  "chat.completion.chunk",
					Created: time.Now().Unix(),
					Model:   model,
					Choices: []StreamChoice{
						{
							Index:        0,
							Delta:        StreamDelta{},
							FinishReason: "stop",
						},
					},
				}
				chunkJSON, _ := json.Marshal(finalChunk)
				fmt.Fprintf(sseWriter, "data: %s\n\n", string(chunkJSON))
				fmt.Fprintf(sseWriter, "data: [DONE]\n\n")
				return nil
			}
			return fmt.Errorf("read error: %w", err)
		}
	}
}

// ChatStreamChunk represents a streaming response chunk
type ChatStreamChunk struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []StreamChoice `json:"choices"`
}

// StreamChoice represents a choice in streaming response
type StreamChoice struct {
	Index        int          `json:"index"`
	Delta        StreamDelta  `json:"delta"`
	FinishReason string       `json:"finish_reason,omitempty"`
}

// StreamDelta represents the delta content in streaming
type StreamDelta struct {
	Role      string     `json:"role,omitempty"`
	Content   string     `json:"content,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// parseSSEStream parses Server-Sent Events stream from Kiro
func parseSSEStream(reader *bufio.Reader) (string, error) {
	var fullContent strings.Builder

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}

		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}

		// Parse SSE field
		if strings.HasPrefix(line, "data:") {
			data := strings.TrimPrefix(line, "data:")
			data = strings.TrimSpace(data)

			// Try to parse as JSON
			var event map[string]interface{}
			if err := json.Unmarshal([]byte(data), &event); err == nil {
				// Extract content from various possible fields
				if content, ok := event["$amazonq.streaming#assistantResponseMessage"].(string); ok {
					fullContent.WriteString(content)
				} else if content, ok := event["message"].(string); ok {
					fullContent.WriteString(content)
				}
			}
		}
	}

	return fullContent.String(), nil
}
