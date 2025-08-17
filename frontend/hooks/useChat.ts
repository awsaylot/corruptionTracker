import { useState, useCallback, useEffect } from 'react';
import { ChatMessage, ChatServiceState } from '../types/chat';
import { ChatUIMessage, ChatUIState } from '../types/chat/ui';
import { chatService } from '../services/chat/chatService';

export function useChat() {
  const [state, setState] = useState<ChatUIState>({
    messages: [],
    isLoading: false,
    error: null,
    connectionStatus: 'disconnected'
  });

  const handleServiceStateChange = useCallback((serviceState: ChatServiceState) => {
    setState(prev => ({
      ...prev,
      connectionStatus: serviceState.status,
      error: serviceState.error?.message || null
    }));
  }, []);

  const handleMessage = useCallback((message: ChatMessage) => {
    setState(prev => {
      const messages = [...prev.messages];
      const lastMessage = messages[messages.length - 1];

      if (message.type === 'error') {
        if (lastMessage && lastMessage.status === 'sending') {
          messages[messages.length - 1] = {
            ...lastMessage,
            status: 'error',
            error: message.content
          };
        }
        return { ...prev, messages, error: message.content };
      }

      if (message.type === 'system') {
        if (lastMessage?.status === 'sending') {
          messages[messages.length - 1] = {
            ...lastMessage,
            status: 'complete'
          };
        }
        return { ...prev, messages };
      }

      // Handle assistant messages
      if (lastMessage?.role === 'assistant' && lastMessage.status === 'sending') {
        // Append to existing message
        messages[messages.length - 1] = {
          ...lastMessage,
          content: lastMessage.content + message.content,
          status: message.metadata?.done ? 'complete' : 'sending'
        };
      } else {
        // Create new message
        messages.push({
          id: Date.now().toString(),
          type: 'message',
          content: message.content,
          role: 'assistant',
          timestamp: new Date(),
          status: message.metadata?.done ? 'complete' : 'sending'
        } as ChatUIMessage);
      }

      return { ...prev, messages };
    });
  }, []);

  useEffect(() => {
    // Set up chat service handlers
    chatService.onMessage(handleMessage);
    chatService.onStateChange(handleServiceStateChange);

    // Connect to chat service
    chatService.connect().catch(error => {
      setState(prev => ({
        ...prev,
        error: error.message,
        isConnected: false
      }));
    });

    // Cleanup on unmount
    return () => {
      chatService.disconnect();
    };
  }, [handleMessage, handleServiceStateChange]);

  const sendMessage = useCallback(async (content: string) => {
    if (!content.trim() || state.isLoading) return;

    setState(prev => ({
      ...prev,
      isLoading: true,
      error: null,
      messages: [
        ...prev.messages,
        {
          id: Date.now().toString(),
          type: 'message',
          content: content.trim(),
          role: 'user',
          timestamp: new Date(),
          status: 'sending'
        } as ChatUIMessage
      ]
    }));

    try {
      await chatService.sendMessage(content);
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to send message',
        messages: prev.messages.map(msg => 
          msg.status === 'sending' 
            ? { ...msg, status: 'error' }
            : msg
        )
      }));
    } finally {
      setState(prev => ({ ...prev, isLoading: false }));
    }
  }, [state.isLoading]);

  const clearMessages = useCallback(() => {
    setState(prev => ({ ...prev, messages: [], error: null }));
  }, []);

  const retryLastMessage = useCallback(async () => {
    const lastUserMessage = [...state.messages]
      .reverse()
      .find(msg => msg.role === 'user');

    if (lastUserMessage) {
      // Remove failed assistant messages
      setState(prev => ({
        ...prev,
        messages: prev.messages.filter(msg => !(msg.role === 'assistant' && msg.status === 'error'))
      }));
      await sendMessage(lastUserMessage.content);
    }
  }, [state.messages, sendMessage]);

  return {
    ...state,
    sendMessage,
    clearMessages,
    retryLastMessage,
  };
}