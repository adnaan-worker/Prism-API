import { useState, useRef, useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { apiClient } from '../../lib/api';
import type { APIKey } from '../../types';
import { Button, Input, Select, Typography, Divider, Switch, Tooltip, Upload } from 'antd';
import {
    SendOutlined,
    SettingOutlined,
    PictureOutlined,
    ClearOutlined,
    SaveOutlined,
    PlayCircleOutlined,
    CodeOutlined
} from '@ant-design/icons';
import type { UploadProps } from 'antd';

const { TextArea } = Input;
const { Title, Text } = Typography;

// Basic message types for the UI
interface Message {
    id: string;
    role: 'user' | 'assistant' | 'system';
    content: string;
    reasoning_content?: string;
    isStreaming?: boolean;
    images?: string[]; // Array of base64 data URIs
}

export default function PlaygroundPage() {
    const [messages, setMessages] = useState<Message[]>([]);
    const [inputValue, setInputValue] = useState('');
    const [isGenerating, setIsGenerating] = useState(false);
    const [pendingImages, setPendingImages] = useState<string[]>([]);

    // Settings State
    const [selectedModel, setSelectedModel] = useState('claude-sonnet-4.5');
    const [selectedKey, setSelectedKey] = useState<string>('');
    const [selectedProtocol, setSelectedProtocol] = useState<'openai' | 'anthropic'>('openai');
    const [systemPrompt, setSystemPrompt] = useState('You are a helpful AI assistant. Please respond clearly and concisely.');
    const [temperature, setTemperature] = useState(0.7);
    const [thinkingMode, setThinkingMode] = useState(true);

    // Fetch API keys
    const { data: apiKeysData } = useQuery({
        queryKey: ['apiKeys'],
        queryFn: async () => {
            const response = await apiClient.get<{ keys: APIKey[] }>('/apikeys');
            return response.data;
        },
    });
    const apiKeys = apiKeysData?.keys || [];

    const bottomRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        // Scroll to bottom whenever messages change
        if (bottomRef.current) {
            bottomRef.current.scrollIntoView({ behavior: 'smooth' });
        }
    }, [messages]);

    const handleSend = async () => {
        if ((!inputValue.trim() && pendingImages.length === 0) || isGenerating) return;

        const newUserMsg: Message = {
            id: Date.now().toString(),
            role: 'user',
            content: inputValue,
            images: pendingImages.length > 0 ? [...pendingImages] : undefined
        };

        const currentMessages = [...messages, newUserMsg];
        setMessages(currentMessages);
        setInputValue('');
        setPendingImages([]);
        setIsGenerating(true);

        const respId = (Date.now() + 1).toString();

        // Add initial empty assistant message
        setMessages(prev => [...prev, {
            id: respId,
            role: 'assistant',
            content: '',
            reasoning_content: '',
            isStreaming: true
        }]);

        try {
            // Build the payload mapping our Message type to ChatMessage format required by the service
            const apiMessages = currentMessages.map(msg => {
                // To support images, if a message has images, we must wrap them in the standard format expected by Prisma-API adapter.
                // Our kiro_adapter.go looks for Message.Content, or Message.ContentParts if an array is provided.
                if (msg.images && msg.images.length > 0) {
                    return {
                        role: msg.role,
                        content: msg.content,
                        contentParts: [
                            { type: 'text', text: msg.content },
                            ...msg.images.map(img => ({ type: 'image_url', image_url: { url: img } }))
                        ]
                    };
                }
                return { role: msg.role, content: msg.content };
            });

            // Insert system prompt
            if (systemPrompt.trim() !== '') {
                apiMessages.unshift({ role: 'system', content: systemPrompt });
            }

            import('../../services/chatService').then(({ chatService }) => {
                chatService.streamChat(
                    {
                        model: selectedModel,
                        messages: apiMessages as any, // bypassing strict types for ContentParts injection
                        temperature: temperature,
                        apiKey: selectedKey || undefined,
                        protocol: selectedProtocol
                    },
                    {
                        onContent: (text) => {
                            setMessages(prev => prev.map(m => m.id === respId ? {
                                ...m,
                                content: m.content + text
                            } : m));
                        },
                        onReasoning: (text) => {
                            if (thinkingMode) {
                                setMessages(prev => prev.map(m => m.id === respId ? {
                                    ...m,
                                    reasoning_content: (m.reasoning_content || '') + text
                                } : m));
                            }
                        },
                        onError: (err) => {
                            console.error('Chat error:', err);
                            setMessages(prev => prev.map(m => m.id === respId ? {
                                ...m,
                                content: m.content + `\n\n[Error: ${err.message || 'Connection failed'}]`,
                                isStreaming: false
                            } : m));
                            setIsGenerating(false);
                        },
                        onFinish: () => {
                            setMessages(prev => prev.map(m => m.id === respId ? {
                                ...m,
                                isStreaming: false
                            } : m));
                            setIsGenerating(false);
                        }
                    }
                );
            });
        } catch (error) {
            console.error('Failed to initiate chat stream', error);
            setIsGenerating(false);
        }
    };

    const clearChat = () => {
        setMessages([]);
    };

    const uploadProps: UploadProps = {
        beforeUpload: (file) => {
            const reader = new FileReader();
            reader.onload = (e) => {
                if (e.target?.result && typeof e.target.result === 'string') {
                    setPendingImages(prev => [...prev, e.target!.result as string]);
                }
            };
            reader.readAsDataURL(file);
            return false; // Prevent default upload behavior
        },
        showUploadList: false,
        accept: "image/*"
    };

    const removePendingImage = (index: number) => {
        setPendingImages(prev => prev.filter((_, i) => i !== index));
    };

    return (
        <div className="flex flex-col h-full bg-transparent overflow-hidden animate-fade-in p-4 lg:p-6">
            <div className="flex items-center justify-between mb-4">
                <div>
                    <Title level={4} className="!m-0 text-text-primary">Chat Playground</Title>
                    <Text className="text-text-tertiary">Test models, tools, and multi-modal streams in real-time.</Text>
                </div>
                <div className="flex gap-2">
                    <Button icon={<ClearOutlined />} onClick={clearChat}>Clear</Button>
                    <Button type="primary" icon={<SaveOutlined />}>Save Preset</Button>
                </div>
            </div>

            <div className="flex-1 bg-white dark:bg-white/5 border border-border rounded-xl shadow-sm overflow-hidden flex h-[calc(100vh-140px)]">
                {/* Left Side: Settings Panel */}
                <div className="w-80 border-r border-border bg-slate-50 dark:bg-black/20 p-4 flex flex-col gap-4 overflow-y-auto custom-scrollbar">
                    <div className="mb-2">
                        <div className="flex items-center gap-2 mb-2 text-text-secondary font-medium">
                            <SettingOutlined /> Configuration
                        </div>

                        <div className="space-y-4">
                            <div>
                                <Text className="text-xs text-text-tertiary uppercase tracking-wider mb-1 block">Model Context</Text>
                                <Select
                                    value={selectedModel}
                                    onChange={setSelectedModel}
                                    className="w-full"
                                    options={[
                                        { value: 'claude-sonnet-4.5', label: 'Claude 4.5 Sonnet' },
                                        { value: 'claude-haiku-4.5', label: 'Claude Haiku 4.5' },
                                        { value: 'claude-sonnet-4', label: 'Claude Sonnet 4' },
                                        { value: 'deepseek-3.2', label: 'DeepSeek 3.2' },
                                        { value: 'minimax-m2.1', label: 'Minimax M2.1' },
                                        { value: 'qwen3-coder-next', label: 'Qwen3 Coder Next' },
                                    ]}
                                />
                            </div>

                            <div>
                                <Text className="text-xs text-text-tertiary uppercase tracking-wider mb-1 block">API Key</Text>
                                <Select
                                    value={selectedKey}
                                    onChange={setSelectedKey}
                                    className="w-full"
                                    options={[
                                        { value: '', label: 'Session Token (Default)' },
                                        ...apiKeys.map(k => ({ value: k.key, label: k.name }))
                                    ]}
                                />
                            </div>

                            <div>
                                <Text className="text-xs text-text-tertiary uppercase tracking-wider mb-1 block">Protocol Format</Text>
                                <Select
                                    value={selectedProtocol}
                                    onChange={setSelectedProtocol as any}
                                    className="w-full"
                                    options={[
                                        { value: 'openai', label: 'OpenAI (/v1/chat/completions)' },
                                        { value: 'anthropic', label: 'Anthropic (/v1/messages)' }
                                    ]}
                                />
                            </div>

                            <div>
                                <Text className="text-xs text-text-tertiary uppercase tracking-wider mb-1 block">System Prompt</Text>
                                <TextArea
                                    rows={4}
                                    value={systemPrompt}
                                    onChange={(e) => setSystemPrompt(e.target.value)}
                                    className="w-full text-sm font-mono custom-scrollbar"
                                    placeholder="You are a helpful assistant..."
                                />
                            </div>

                            <Divider className="my-2 border-border" />

                            <div>
                                <div className="flex justify-between items-center mb-1">
                                    <Text className="text-xs text-text-tertiary uppercase tracking-wider block">Temperature</Text>
                                    <Text className="text-xs font-mono">{temperature.toFixed(2)}</Text>
                                </div>
                                <input
                                    type="range"
                                    min="0" max="2" step="0.1"
                                    value={temperature}
                                    onChange={(e) => setTemperature(parseFloat(e.target.value))}
                                    className="w-full accent-primary"
                                />
                            </div>

                            <div className="flex items-center justify-between p-3 bg-white dark:bg-white/10 rounded-lg border border-border">
                                <div>
                                    <div className="text-sm font-medium">Enable Thinking</div>
                                    <div className="text-xs text-text-tertiary">Parse &lt;thinking&gt; tags</div>
                                </div>
                                <Switch checked={thinkingMode} onChange={setThinkingMode} />
                            </div>

                            <div className="flex items-center justify-between p-3 bg-white dark:bg-white/10 rounded-lg border border-border opacity-70 cursor-not-allowed">
                                <div>
                                    <div className="text-sm font-medium flex items-center gap-2"><CodeOutlined /> Tools</div>
                                    <div className="text-xs text-text-tertiary">Function calling (Soon)</div>
                                </div>
                                <Switch disabled />
                            </div>
                        </div>
                    </div>
                </div>

                {/* Right Side: Chat Area */}
                <div className="flex-1 flex flex-col relative bg-transparent">
                    {/* Messages Container */}
                    <div className="flex-1 overflow-y-auto p-6 space-y-6 custom-scrollbar">
                        {messages.length === 0 ? (
                            <div className="h-full flex flex-col items-center justify-center text-text-tertiary">
                                <div className="w-16 h-16 bg-primary/10 rounded-2xl flex items-center justify-center mb-4 text-primary text-2xl">
                                    <PlayCircleOutlined />
                                </div>
                                <h3 className="text-lg font-medium text-text-secondary mb-2">Playground Ready</h3>
                                <p className="max-w-md text-center text-sm">
                                    Send a message to start streaming responses. Use the configuration panel to inject system prompts or change models.
                                </p>
                            </div>
                        ) : (
                            messages.map(msg => (
                                <div key={msg.id} className={`flex ${msg.role === 'user' ? 'justify-end' : 'justify-start'}`}>
                                    <div className={`max-w-[85%] rounded-2xl p-4 ${msg.role === 'user'
                                        ? 'bg-primary text-white shadow-md'
                                        : 'bg-white dark:bg-[#1a1a1a] border border-border shadow-sm text-text-primary'
                                        }`}>
                                        {msg.role === 'assistant' && (
                                            <div className="flex items-center gap-2 mb-2 text-xs font-medium text-text-tertiary uppercase tracking-wider">
                                                <span className="w-4 h-4 rounded-full bg-primary/20 flex items-center justify-center">
                                                    <span className="w-2 h-2 rounded-full bg-primary animate-pulse"></span>
                                                </span>
                                                {selectedModel}
                                            </div>
                                        )}

                                        {/* Render images if any */}
                                        {msg.images && msg.images.length > 0 && (
                                            <div className="flex flex-wrap gap-2 mb-3">
                                                {msg.images.map((img, i) => (
                                                    <img key={i} src={img} alt="attachment" className="h-32 object-contain rounded border border-border bg-black/5 dark:bg-white/5" />
                                                ))}
                                            </div>
                                        )}

                                        {/* render thinking block if exists */}
                                        {msg.reasoning_content && (
                                            <div className="mb-3 rounded-lg bg-slate-50 dark:bg-black/40 border-l-2 border-primary/50 text-text-secondary text-sm overflow-hidden">
                                                <details className="group" open={msg.isStreaming}>
                                                    <summary className="cursor-pointer p-2 hover:bg-black/5 dark:hover:bg-white/5 transition-colors font-medium flex items-center outline-none">
                                                        <span className="mr-2 opacity-60">🤔</span>
                                                        Thinking Process
                                                        <span className="ml-auto text-xs text-text-tertiary opacity-0 group-hover:opacity-100 transition-opacity">Click to toggle</span>
                                                    </summary>
                                                    <div className="p-3 pt-1 border-t border-border/50 font-mono text-xs whitespace-pre-wrap opacity-80">
                                                        {msg.reasoning_content}
                                                        {msg.isStreaming && <span className="ml-1 inline-block w-1.5 h-3 bg-text-tertiary animate-pulse"></span>}
                                                    </div>
                                                </details>
                                            </div>
                                        )}

                                        <div className="whitespace-pre-wrap leading-relaxed">
                                            {msg.content}
                                            {msg.isStreaming && !msg.reasoning_content && <span className="ml-2 inline-block w-2 h-4 bg-primary animate-pulse align-middle"></span>}
                                        </div>
                                    </div>
                                </div>
                            ))
                        )}
                        <div ref={bottomRef} className="h-4" />
                    </div>

                    {/* Input Area */}
                    <div className="p-4 bg-white dark:bg-transparent border-t border-border flex flex-col gap-2">

                        {/* Pending Images Preview */}
                        {pendingImages.length > 0 && (
                            <div className="flex flex-wrap gap-2 px-2">
                                {pendingImages.map((img, i) => (
                                    <div key={i} className="relative group">
                                        <img src={img} alt="pending" className="h-16 w-16 object-cover rounded-lg border border-border" />
                                        <div
                                            className="absolute -top-2 -right-2 bg-red-500 text-white rounded-full w-5 h-5 flex items-center justify-center cursor-pointer opacity-0 group-hover:opacity-100 transition-opacity text-xs"
                                            onClick={() => removePendingImage(i)}
                                        >
                                            ×
                                        </div>
                                    </div>
                                ))}
                            </div>
                        )}

                        <div className="relative flex items-end gap-2 bg-slate-50 dark:bg-black/20 border border-border rounded-xl p-2 focus-within:border-primary/50 focus-within:ring-1 focus-within:ring-primary/20 transition-all shadow-sm">
                            <Upload {...uploadProps}>
                                <Tooltip title="Attach Image">
                                    <Button
                                        type="text"
                                        icon={<PictureOutlined />}
                                        className="h-10 w-10 text-text-secondary hover:text-primary hover:bg-primary/5"
                                    />
                                </Tooltip>
                            </Upload>

                            <TextArea
                                value={inputValue}
                                onChange={(e) => setInputValue(e.target.value)}
                                onPressEnter={(e) => {
                                    if (!e.shiftKey) {
                                        e.preventDefault();
                                        handleSend();
                                    }
                                }}
                                placeholder="Type a message... (Shift + Enter for new line)"
                                autoSize={{ minRows: 1, maxRows: 6 }}
                                className="!bg-transparent !border-none !shadow-none !ring-0 text-text-primary px-2 py-2 text-base resize-none custom-scrollbar"
                            />

                            <Button
                                type="primary"
                                shape="circle"
                                icon={<SendOutlined />}
                                onClick={handleSend}
                                disabled={!inputValue.trim() || isGenerating}
                                loading={isGenerating}
                                className="h-10 w-10 flex-shrink-0 shadow-md transition-transform active:scale-95"
                            />
                        </div>
                        <div className="text-center mt-2 text-xs text-text-tertiary">
                            Playground connects directly to Prism-API <code className="bg-black/5 dark:bg-white/10 px-1 rounded px-1">/v1/chat/completions</code> via standard SSE streams.
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}
