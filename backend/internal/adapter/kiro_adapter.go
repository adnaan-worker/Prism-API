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
	ProfileArn        string                `json:"profileArn,omitempty"` // Required for Social Auth
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
	ToolUseID string                   `json:"toolUseId"`
	Status    string                   `json:"status"`
	Content   []map[string]interface{} `json:"content"`
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
func (a *KiroAdapter) convertRequest(req *ChatRequest) (*kiroRequest, error) {
	conversationID := uuid.New().String()

	// Get Kiro model ID from database mapping
	kiroModelID, err := a.modelMapper.GetModelMapping(context.Background(), req.Model)
	if err != nil {
		// Fallback: use model name as-is if mapping not found
		kiroModelID = req.Model
	}

	// Extract system prompt and process messages
	var systemPrompt string
	var history []kiroMessage
	var currentUserContent string

	// Process messages
	for i, msg := range req.Messages {
		if msg.Role == "system" {
			systemPrompt = msg.Content
			continue
		}

		// Last user message becomes currentMessage
		if msg.Role == "user" && i == len(req.Messages)-1 {
			currentUserContent = msg.Content
		} else {
			// Add to history
			if msg.Role == "user" {
				content := msg.Content
				// Prepend system prompt to first user message in history
				if systemPrompt != "" && len(history) == 0 {
					content = systemPrompt + "\n\n" + content
					systemPrompt = "" // Only add once
				}
				history = append(history, kiroMessage{
					UserInputMessage: &kiroUserMessage{
						Content: content,
						ModelID: kiroModelID,
						Origin:  "AI_EDITOR",
					},
				})
			} else if msg.Role == "assistant" {
				history = append(history, kiroMessage{
					AssistantResponseMessage: &kiroAssistantMessage{
						Content: msg.Content,
					},
				})
			}
		}
	}

	// Kiro API requires history to end with assistantResponseMessage
	if len(history) > 0 {
		lastMsg := history[len(history)-1]
		if lastMsg.AssistantResponseMessage == nil {
			// Add empty assistant message
			history = append(history, kiroMessage{
				AssistantResponseMessage: &kiroAssistantMessage{
					Content: "Continue",
				},
			})
		}
	}

	// Prepend system prompt to current message if not added to history
	if systemPrompt != "" {
		currentUserContent = systemPrompt + "\n\n" + currentUserContent
	}

	// Ensure current content is not empty
	if currentUserContent == "" {
		currentUserContent = "Continue"
	}

	// Build current message
	currentMsg := kiroMessage{
		UserInputMessage: &kiroUserMessage{
			Content: currentUserContent,
			ModelID: kiroModelID,
			Origin:  "AI_EDITOR",
		},
	}

	// Build tools context
	var kiroTools []kiroTool
	if len(req.Tools) > 0 {
		kiroTools = make([]kiroTool, 0, len(req.Tools))
		for _, tool := range req.Tools {
			kiroTools = append(kiroTools, kiroTool{
				ToolSpecification: kiroToolSpec{
					Name:        tool.Function.Name,
					Description: tool.Function.Description,
					InputSchema: kiroInputSchema{
						JSON: tool.Function.Parameters,
					},
				},
			})
		}
	}

	// Always add userInputMessageContext if tools are present
	if len(kiroTools) > 0 {
		currentMsg.UserInputMessage.UserInputMessageContext = &kiroMessageContext{
			Tools: kiroTools,
		}
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

	return kiroReq, nil
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
								// Complete previous tool call
								var args map[string]interface{}
								if err := json.Unmarshal([]byte(currentToolCall.Function.Arguments), &args); err == nil {
									argsBytes, _ := json.Marshal(args)
									currentToolCall.Function.Arguments = string(argsBytes)
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
							currentToolCall.Function.Arguments += input
						} else if inputObj, hasInputObj := actualEvent["input"].(map[string]interface{}); hasInputObj {
							// Complete input object provided
							inputBytes, _ := json.Marshal(inputObj)
							currentToolCall.Function.Arguments = string(inputBytes)
						}

						// Check if tool use is complete
						if stop, hasStop := actualEvent["stop"].(bool); hasStop && stop {
							// Validate and format arguments as JSON
							var args map[string]interface{}
							if err := json.Unmarshal([]byte(currentToolCall.Function.Arguments), &args); err == nil {
								argsBytes, _ := json.Marshal(args)
								currentToolCall.Function.Arguments = string(argsBytes)
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
								if stop, hasStop := actualEvent["stop"].(bool); hasStop && stop {
									// Tool use complete - send tool call chunk
									var inputArgs string
									if input, hasInput := actualEvent["input"].(string); hasInput {
										inputArgs = input
									} else if inputObj, hasInputObj := actualEvent["input"].(map[string]interface{}); hasInputObj {
										inputBytes, _ := json.Marshal(inputObj)
										inputArgs = string(inputBytes)
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
																Arguments: inputArgs,
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
