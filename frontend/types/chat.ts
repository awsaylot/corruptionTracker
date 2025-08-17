export type MessageStatus = 'sending' | 'sent' | 'error';

export interface Message {
  id: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  timestamp: Date;
  status: MessageStatus;
}

export type ConnectionStatus = 'connecting' | 'connected' | 'disconnected' | 'error';
