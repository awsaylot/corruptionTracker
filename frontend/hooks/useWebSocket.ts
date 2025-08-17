import { useEffect, useRef, useState, useCallback } from 'react';
import { ChatMessage, ConnectionStatus } from '../types/chat';

interface WebSocketHookReturn {
  sendMessage: (content: string) => void;
  messages: ChatMessage[];
  connectionStatus: ConnectionStatus;
  lastError: string | null;
}

export const useWebSocket = (url: string): WebSocketHookReturn => {
  const [connectionStatus, setConnectionStatus] = useState<ConnectionStatus>('connecting');
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [lastError, setLastError] = useState<string | null>(null);
  const ws = useRef<WebSocket | null>(null);

  const generateMessageId = () => `msg_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;

  useEffect(() => {
    const connect = () => {
      try {
        ws.current = new WebSocket(url);

        ws.current.onopen = () => {
          setConnectionStatus('connected');
          setLastError(null);
        };

        ws.current.onclose = () => {
          setConnectionStatus('disconnected');
          // Attempt to reconnect after 3 seconds
          setTimeout(connect, 3000);
        };

        ws.current.onerror = (error) => {
          setConnectionStatus('error');
          setLastError('WebSocket connection error');
        };

        ws.current.onmessage = (event) => {
          try {
            const data = JSON.parse(event.data);
            const message: ChatMessage = {
              id: generateMessageId(),
              content: data.content,
              timestamp: new Date(),
              role: 'assistant',
              status: 'sent'
            };
            setMessages(prev => [...prev, message]);
          } catch (error) {
            console.error('Failed to parse message:', error);
          }
        };
      } catch (error) {
        setConnectionStatus('error');
        setLastError('Failed to establish WebSocket connection');
      }
    };

    connect();

    return () => {
      if (ws.current) {
        ws.current.close();
      }
    };
  }, [url]);

  const sendMessage = useCallback((content: string) => {
    if (!content.trim() || !ws.current || ws.current.readyState !== WebSocket.OPEN) {
      return;
    }

    const message: ChatMessage = {
      id: generateMessageId(),
      content: content.trim(),
      timestamp: new Date(),
      role: 'user',
      status: 'sent'
    };

    try {
      ws.current.send(JSON.stringify({
        type: 'message',
        content: message.content
      }));
      
      setMessages(prev => [...prev, message]);
    } catch (error) {
      setLastError('Failed to send message');
    }
  }, []);

  return {
    sendMessage,
    messages,
    connectionStatus,
    lastError
  };
};
