// types/chat.ts

export interface ChatMessage {
    id: string;
    content: string;
    role: 'user' | 'assistant' | 'system';
    timestamp: Date;
    status: 'streaming' | 'complete' | 'error';
    metadata?: {
        model?: string;
        tokens?: number;
        processingTime?: number;
    };
}

export interface ChatServiceConfig {
    baseUrl?: string;
    reconnectAttempts?: number;
    reconnectDelay?: number;
    messageTimeout?: number;
    enableLogging?: boolean;
}

export interface ChatServiceState {
    status: 'disconnected' | 'connecting' | 'connected' | 'error' | 'generating';
    error?: Error;
    connectionId?: string;
}

export interface WebSocketMessage {
    type: 'message' | 'ping' | 'start' | 'stop';
    content?: string;
    conversationId?: string | null;
    messages?: ChatMessage[];
    metadata?: {
        model?: string;
        temperature?: number;
        maxTokens?: number;
    };
}

export interface Conversation {
    id: string;
    title: string;
    messages: ChatMessage[];
    createdAt: Date;
    updatedAt: Date;
    metadata?: {
        model?: string;
        totalTokens?: number;
        averageResponseTime?: number;
    };
}

// Event handler types
export type MessageHandler = (message: ChatMessage) => void;
export type StateChangeHandler = (state: ChatServiceState) => void;
export type ErrorHandler = (error: Error) => void;

// API Response types
export interface ApiResponse<T = any> {
    success: boolean;
    data?: T;
    error?: string;
    message?: string;
}

export interface StreamingResponse {
    type: 'chunk' | 'complete' | 'error';
    content?: string;
    done?: boolean;
    error?: string;
    metadata?: {
        tokens?: number;
        model?: string;
    };
}

// Utility types for React components
export interface ChatComponentProps {
    conversation?: Conversation;
    onMessageSend?: (content: string) => void;
    onConversationChange?: (conversation: Conversation) => void;
    className?: string;
    theme?: 'light' | 'dark';
    autoFocus?: boolean;
    placeholder?: string;
    maxLength?: number;
}

export interface MessageComponentProps {
    message: ChatMessage;
    onEdit?: (messageId: string, newContent: string) => void;
    onDelete?: (messageId: string) => void;
    onRegenerate?: (messageId: string) => void;
    showTimestamp?: boolean;
    showActions?: boolean;
    className?: string;
}

// Configuration types
export interface LLMConfig {
    model: string;
    temperature?: number;
    maxTokens?: number;
    topP?: number;
    frequencyPenalty?: number;
    presencePenalty?: number;
    systemMessage?: string;
}

export interface ChatSettings {
    llm: LLMConfig;
    ui: {
        theme: 'light' | 'dark' | 'auto';
        showTimestamps: boolean;
        showTokenCounts: boolean;
        enableMarkdown: boolean;
        enableCodeHighlighting: boolean;
    };
    behavior: {
        autoSave: boolean;
        clearOnRefresh: boolean;
        maxConversationLength: number;
        enableNotifications: boolean;
    };
}

// Error types
export class ChatServiceError extends Error {
    constructor(
        message: string,
        public code?: string,
        public originalError?: Error
    ) {
        super(message);
        this.name = 'ChatServiceError';
    }
}

export class WebSocketError extends ChatServiceError {
    constructor(message: string, originalError?: Error) {
        super(message, 'WEBSOCKET_ERROR', originalError);
        this.name = 'WebSocketError';
    }
}

export class APIError extends ChatServiceError {
    constructor(
        message: string,
        public statusCode?: number,
        originalError?: Error
    ) {
        super(message, 'API_ERROR', originalError);
        this.name = 'APIError';
    }
}