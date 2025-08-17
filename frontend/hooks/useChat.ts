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
      
      // Handle error messages
      if (message.status === 'error') {
        const lastMessage = messages[messages.length - 1];
        if (lastMessage && lastMessage.status === 'streaming') {
          // Update the last message to error state
          messages[messages.length - 1] = {
            ...lastMessage,
            status: 'error',
            error: message.content
          };
        } else {
          // Add new error message
          messages.push({
            id: message.id,
            content: message.content,
            role: message.role,
            timestamp: message.timestamp,
            status: 'error'
          } as ChatUIMessage);
        }
        return { ...prev, messages, error: message.content };
      }

      // Handle assistant messages (streaming or complete)
      if (message.role === 'assistant') {
        // Find if there's already a streaming message from assistant
        const lastMessage = messages[messages.length - 1];
        
        if (lastMessage?.role === 'assistant' && 
            (lastMessage.status === 'streaming' || lastMessage.status === 'sending')) {
          // Update existing streaming message
          // IMPORTANT: Replace content, don't concatenate (chatService already handles accumulation)
          messages[messages.length - 1] = {
            ...lastMessage,
            content: message.content, // Replace, not append
            status: message.status === 'complete' ? 'complete' : 'streaming'
          };
        } else {
          // Create new assistant message
          messages.push({
            id: message.id,
            content: message.content,
            role: message.role,
            timestamp: message.timestamp,
            status: message.status === 'complete' ? 'complete' : 'streaming'
          } as ChatUIMessage);
        }
      } else {
        // Handle other message types (user, system)
        messages.push({
          id: message.id,
          content: message.content,
          role: message.role,
          timestamp: message.timestamp,
          status: message.status
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
        connectionStatus: 'error'
      }));
    });

    // Cleanup on unmount
    return () => {
      chatService.disconnect();
    };
  }, [handleMessage, handleServiceStateChange]);

  const sendMessage = useCallback(async (content: string) => {
    if (!content.trim() || state.isLoading) return;

    // Add user message immediately
    const userMessage: ChatUIMessage = {
      id: `user_${Date.now()}`,
      content: content.trim(),
      role: 'user',
      timestamp: new Date(),
      status: 'sending'
    };

    setState(prev => ({
      ...prev,
      isLoading: true,
      error: null,
      messages: [...prev.messages, userMessage]
    }));

    try {
      // Send message with current conversation context
      const conversationMessages = state.messages
        .filter(msg => msg.status === 'complete')
        .map(msg => ({
          id: msg.id,
          content: msg.content,
          role: msg.role,
          timestamp: msg.timestamp,
          status: msg.status
        }));

      await chatService.sendMessage(content.trim(), conversationMessages);
      
      // Update user message status to sent
      setState(prev => ({
        ...prev,
        messages: prev.messages.map(msg => 
          msg.id === userMessage.id 
            ? { ...msg, status: 'complete' }
            : msg
        )
      }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to send message',
        messages: prev.messages.map(msg => 
          msg.id === userMessage.id
            ? { ...msg, status: 'error', error: error instanceof Error ? error.message : 'Failed to send' }
            : msg
        )
      }));
    } finally {
      setState(prev => ({ ...prev, isLoading: false }));
    }
  }, [state.isLoading, state.messages]);

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