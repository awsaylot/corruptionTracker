import { ChatMessage, ConnectionStatus } from './index';

// UI-specific message type
export interface ChatUIMessage extends ChatMessage {
    // Add any UI-specific message properties here
}

// UI-specific state
export interface ChatUIState {
    messages: ChatUIMessage[];
    isLoading: boolean;
    error: string | null;
    connectionStatus: ConnectionStatus;
}

// System status for UI display
export interface SystemStatus {
    cpuUsage: number;
    memoryUsage: number;
    neuralCore: boolean;
    encryptionActive: boolean;
    sessionActive: boolean;
}
