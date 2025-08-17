import { 
    ChatMessage, 
    ChatServiceConfig, 
    ChatServiceState, 
    MessageHandler, 
    StateChangeHandler, 
    WebSocketMessage 
} from '../../types/chat';
import { v4 as uuid } from 'uuid';

const DEFAULT_CONFIG: Required<ChatServiceConfig> = {
    baseUrl: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080',
    reconnectAttempts: 3,
    reconnectDelay: 1000,
    messageTimeout: 30000
};

// Backend WebSocket message types (matching Go backend)
interface BackendWebSocketMessage {
    type: 'chat' | 'ping' | 'pong';
    messages?: Array<{
        role: string;
        content: string;
    }>;
    content?: string;
}

interface BackendWebSocketResponse {
    type: 'chat_chunk' | 'error' | 'pong';
    content?: string;
    done?: boolean;
    error?: string;
}

export class ChatService {
    private ws: WebSocket | null = null;
    private config: Required<ChatServiceConfig>;
    private messageHandler: MessageHandler | null = null;
    private stateChangeHandler: StateChangeHandler | null = null;
    private reconnectCount = 0;
    private conversationId: string | null = null;
    private messageQueue: Array<() => void> = [];
    private state: ChatServiceState = { status: 'disconnected' };
    private currentStreamingMessage: ChatMessage | null = null;
    private pingInterval: NodeJS.Timeout | null = null;

    constructor(config?: ChatServiceConfig) {
        this.config = { ...DEFAULT_CONFIG, ...config };
    }

    private setState(newState: ChatServiceState) {
        this.state = newState;
        this.stateChangeHandler?.(newState);
    }

    private async waitForConnection(): Promise<void> {
        if (this.ws?.readyState === WebSocket.OPEN) {
            return;
        }

        return new Promise((resolve, reject) => {
            const timeout = setTimeout(() => {
                reject(new Error('Connection timeout'));
            }, this.config.messageTimeout);

            const checkConnection = setInterval(() => {
                if (this.ws?.readyState === WebSocket.OPEN) {
                    clearTimeout(timeout);
                    clearInterval(checkConnection);
                    resolve();
                }
            }, 100);
        });
    }

    private handleWebSocketError(error: Event) {
        console.error('WebSocket error:', error);
        this.stopPingInterval();
        
        // Check if the backend is available before attempting reconnection
        fetch(`${this.config.baseUrl}/health`)
            .then(response => {
                if (!response.ok) {
                    throw new Error('Backend health check failed');
                }
                
                // Backend is available, attempt reconnection
                if (this.reconnectCount < this.config.reconnectAttempts) {
                    this.reconnectCount++;
                    const delay = this.config.reconnectDelay * this.reconnectCount;
                    console.log(`Attempting reconnection in ${delay}ms (attempt ${this.reconnectCount})`);
                    
                    setTimeout(() => {
                        this.connect();
                    }, delay);
                } else {
                    this.setState({
                        status: 'error',
                        error: new Error('Maximum reconnection attempts reached')
                    });
                }
            })
            .catch(error => {
                console.error('Backend health check failed:', error);
                this.setState({
                    status: 'error',
                    error: new Error('Cannot connect to backend server')
                });
            });
    }

    private startPingInterval() {
        // Send ping every 30 seconds to keep connection alive
        this.pingInterval = setInterval(() => {
            if (this.ws?.readyState === WebSocket.OPEN) {
                const pingMessage: BackendWebSocketMessage = { type: 'ping' };
                this.ws.send(JSON.stringify(pingMessage));
            }
        }, 30000);
    }

    private stopPingInterval() {
        if (this.pingInterval) {
            clearInterval(this.pingInterval);
            this.pingInterval = null;
        }
    }

    private handleStreamingMessage(response: BackendWebSocketResponse) {
        if (response.type === 'chat_chunk') {
            if (!response.done && response.content) {
                // Start or continue streaming message
                if (!this.currentStreamingMessage) {
                    this.currentStreamingMessage = {
                        id: uuid(),
                        content: response.content,
                        timestamp: new Date(),
                        role: 'assistant',
                        status: 'streaming'
                    };
                    this.messageHandler?.(this.currentStreamingMessage);
                } else {
                    // Update existing streaming message
                    this.currentStreamingMessage.content += response.content;
                    this.currentStreamingMessage.status = 'streaming';
                    this.messageHandler?.(this.currentStreamingMessage);
                }
            } else if (response.done) {
                // Finish streaming message
                if (this.currentStreamingMessage) {
                    this.currentStreamingMessage.status = 'complete';
                    this.messageHandler?.(this.currentStreamingMessage);
                    this.currentStreamingMessage = null;
                }
                this.setState({ status: 'connected' });
            }
        } else if (response.type === 'error') {
            // Handle error
            const errorMessage: ChatMessage = {
                id: uuid(),
                content: `Error: ${response.error || 'Unknown error occurred'}`,
                timestamp: new Date(),
                role: 'assistant',
                status: 'error'
            };
            this.messageHandler?.(errorMessage);
            this.currentStreamingMessage = null;
            this.setState({ status: 'connected' });
        }
    }

    private setupWebSocket() {
        if (!this.ws) return;

        this.ws.onopen = () => {
            console.log('✅ WebSocket connection opened');
            this.setState({ status: 'connected' });
            this.reconnectCount = 0;
            this.startPingInterval();
            
            // Process any queued messages
            while (this.messageQueue.length > 0) {
                const sendMessage = this.messageQueue.shift();
                sendMessage?.();
            }
        };

        this.ws.onclose = (event) => {
            console.log(`🛑 WebSocket closed: code=${event.code}, reason=${event.reason}`);
            this.setState({ status: 'disconnected' });
            this.stopPingInterval();
            this.ws = null;
            this.conversationId = null;
            this.currentStreamingMessage = null;
        };

        this.ws.onerror = this.handleWebSocketError.bind(this);

        this.ws.onmessage = (event) => {
            try {
                // Try to parse as JSON first (structured messages)
                const response: BackendWebSocketResponse = JSON.parse(event.data);
                this.handleStreamingMessage(response);
            } catch (error) {
                // If not JSON, treat as plain text (fallback for simple messages)
                const content = event.data.toString();
                
                // Ignore ping/pong and empty messages
                if (content === 'ping' || content === 'pong' || !content.trim()) {
                    return;
                }

                // Create a simple message for plain text
                const message: ChatMessage = {
                    id: uuid(),
                    content: content,
                    timestamp: new Date(),
                    role: 'assistant',
                    status: 'complete'
                };

                this.messageHandler?.(message);
            }
        };
    }

    async connect(): Promise<void> {
        if (this.ws?.readyState === WebSocket.OPEN) {
            return;
        }

        this.setState({ status: 'connecting' });
        
        const wsUrl = `${this.config.baseUrl.replace(/^http/, 'ws')}/ws`;
        console.log('Connecting to WebSocket URL:', wsUrl);
        
        try {
            this.ws = new WebSocket(wsUrl);
            this.setupWebSocket();
        } catch (error) {
            console.error('Failed to create WebSocket:', error);
            this.setState({
                status: 'error',
                error: new Error(`Failed to create WebSocket connection: ${(error as Error).message}`)
            });
            throw error;
        }

        try {
            await this.waitForConnection();
        } catch (error) {
            this.setState({
                status: 'error',
                error: new Error('Failed to establish WebSocket connection')
            });
            throw error;
        }
    }

    async sendMessage(content: string, messages: ChatMessage[] = []): Promise<void> {
        // Set status to generating
        this.setState({ status: 'connected' }); // Keep connected but could add 'generating' status

        // Convert ChatMessage[] to backend format
        const backendMessages = messages.map(msg => ({
            role: msg.role,
            content: msg.content
        }));

        // Add current message
        backendMessages.push({
            role: 'user',
            content: content
        });

        const message: BackendWebSocketMessage = {
            type: 'chat',
            messages: backendMessages
        };

        const sendOperation = () => {
            if (this.ws?.readyState === WebSocket.OPEN) {
                console.log('Sending message:', message);
                this.ws.send(JSON.stringify(message));
            } else {
                throw new Error('WebSocket is not connected');
            }
        };

        if (this.ws?.readyState !== WebSocket.OPEN) {
            // Queue message if not connected
            this.messageQueue.push(sendOperation);
            await this.connect();
        } else {
            sendOperation();
        }
    }

    // Send a simple message (for backward compatibility)
    async sendSimpleMessage(content: string): Promise<void> {
        const message: BackendWebSocketMessage = {
            type: 'chat',
            content: content
        };

        const sendOperation = () => {
            if (this.ws?.readyState === WebSocket.OPEN) {
                console.log('Sending simple message:', message);
                this.ws.send(JSON.stringify(message));
            } else {
                throw new Error('WebSocket is not connected');
            }
        };

        if (this.ws?.readyState !== WebSocket.OPEN) {
            this.messageQueue.push(sendOperation);
            await this.connect();
        } else {
            sendOperation();
        }
    }

    onMessage(handler: MessageHandler) {
        this.messageHandler = handler;
    }

    onStateChange(handler: StateChangeHandler) {
        this.stateChangeHandler = handler;
    }

    getState(): ChatServiceState {
        return this.state;
    }

    getConversationId(): string | null {
        return this.conversationId;
    }

    isStreaming(): boolean {
        return this.currentStreamingMessage !== null;
    }

    disconnect() {
        this.stopPingInterval();
        if (this.ws) {
            this.ws.close();
            this.ws = null;
            this.conversationId = null;
            this.messageHandler = null;
            this.stateChangeHandler = null;
            this.messageQueue = [];
            this.currentStreamingMessage = null;
            this.setState({ status: 'disconnected' });
        }
    }
}

export const chatService = new ChatService();