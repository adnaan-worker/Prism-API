// A service dedicated to handling the streaming connection for the Chat Playground.
import { authService } from './authService';

export interface ChatMessage {
    role: 'system' | 'user' | 'assistant';
    content: string;
    reasoning_content?: string;
    // We can expand this with ToolCalls later
}

interface ChatRequestOptions {
    model: string;
    messages: ChatMessage[];
    temperature?: number;
    stream?: boolean;
    apiKey?: string; // Optional custom API key
    protocol?: 'openai' | 'anthropic'; // Request protocol
}

export const chatService = {
    /**
     * Send a streaming request to the chat completion API
     */
    async streamChat(
        options: ChatRequestOptions,
        handlers: {
            onContent: (text: string) => void;
            onReasoning: (text: string) => void;
            onError: (err: any) => void;
            onFinish: () => void;
        }
    ) {
        try {
            // First check if the custom API key is provided, if not fallback to dashboard session token (for testing).
            const token = options.apiKey || authService.getToken();
            if (!token) throw new Error('No authentication token or API key provided.');

            const protocol = options.protocol || 'openai';
            let url = `${import.meta.env.VITE_API_URL || 'http://localhost:8080'}/v1/chat/completions`;
            let reqBody: any = {
                stream: true,
                model: options.model,
                temperature: options.temperature,
            };

            // Format Request
            if (protocol === 'anthropic') {
                url = `${import.meta.env.VITE_API_URL || 'http://localhost:8080'}/v1/messages`;
                // Extract system message for Anthropic format
                const sysMsg = options.messages.find(m => m.role === 'system');
                if (sysMsg) {
                    reqBody.system = sysMsg.content;
                }
                const anthropicMessages = options.messages
                    .filter(m => m.role !== 'system')
                    .map(m => ({
                        role: m.role,
                        content: m.content
                        // Note: If content is already an array of parts due to images, the backend Anthropic adapter needs to handle it or we re-format it precisely.
                    }));
                reqBody.messages = anthropicMessages;
                reqBody.max_tokens = 4096; // Anthropic requires max_tokens
            } else {
                reqBody.messages = options.messages;
            }

            const response = await fetch(url, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`
                },
                body: JSON.stringify(reqBody)
            });

            if (!response.ok) {
                const errText = await response.text();
                handlers.onError(new Error(`API Error ${response.status}: ${errText}`));
                return;
            }

            if (!response.body) {
                handlers.onError(new Error('No response body returned from server'));
                return;
            }

            // Read the SSE stream
            const reader = response.body.getReader();
            const decoder = new TextDecoder('utf-8');
            let buffer = '';

            while (true) {
                const { done, value } = await reader.read();
                if (done) break;

                buffer += decoder.decode(value, { stream: true });

                // Process complete chunks
                const lines = buffer.split('\n');
                // keep the last string in the buffer since it might be incomplete
                buffer = lines.pop() || '';

                let currentEvent = '';

                for (const line of lines) {
                    if (line.trim() === '') continue;

                    if (protocol === 'anthropic') {
                        // Anthropic SSE Format Handling
                        if (line.startsWith('event: ')) {
                            currentEvent = line.substring(7).trim();
                            continue;
                        }
                        if (line.startsWith('data: ')) {
                            const dataStr = line.substring(6).trim();
                            if (currentEvent === 'ping') continue;
                            if (currentEvent === 'message_stop' || currentEvent === 'error') continue;

                            try {
                                const data = JSON.parse(dataStr);
                                if (currentEvent === 'content_block_delta') {
                                    if (data.delta && data.delta.type === 'text_delta') {
                                        handlers.onContent(data.delta.text);
                                    } else if (data.delta && data.delta.type === 'reasoning_delta' && data.delta.reasoning) {
                                        // Specific thinking representation for some models like Claude 3.7
                                        handlers.onReasoning(data.delta.reasoning);
                                    }
                                }
                            } catch (e) {
                                console.warn('Failed to parse Anthropic SSE chunk', dataStr, e);
                            }
                        }
                    } else {
                        // OpenAI SSE Format Handling
                        if (line.startsWith('data: ')) {
                            const dataStr = line.substring(6).trim();
                            if (dataStr === '[DONE]') {
                                continue; // Stream finished
                            }

                            try {
                                const data = JSON.parse(dataStr);
                                if (data.choices && data.choices[0] && data.choices[0].delta) {
                                    const delta = data.choices[0].delta;

                                    if (delta.content) {
                                        handlers.onContent(delta.content);
                                    }
                                    if (delta.reasoning_content) {
                                        handlers.onReasoning(delta.reasoning_content);
                                    }
                                }
                            } catch (e) {
                                console.warn('Failed to parse OpenAI SSE JSON chunk', dataStr, e);
                            }
                        }
                    }
                }
            }

            handlers.onFinish();
        } catch (error) {
            handlers.onError(error);
        }
    }
};
