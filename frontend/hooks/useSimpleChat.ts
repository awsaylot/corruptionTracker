import { useState } from 'react';
import { Message } from '../types/chat';

interface SimpleChatResponse {
  choices: Array<{
    message: Message;
  }>;
}

export function useSimpleChat() {
  const [messages, setMessages] = useState<Message[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const sendMessage = async (content: string) => {
    try {
      setIsLoading(true);
      setError(null);

      // Add user message to the list
      const userMessage: Message = {
        role: 'user',
        content,
        timestamp: new Date(),
        id: `msg_${Date.now()}`,
        status: 'sent'
      };
      setMessages(prev => [...prev, userMessage]);

      // Send request to the non-streaming endpoint
      const response = await fetch('http://localhost:8080/llm/chat', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          messages: [userMessage]
        }),
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const data: SimpleChatResponse = await response.json();
      
      // Add assistant's response to messages
      const assistantMessage: Message = {
        ...data.choices[0].message,
        timestamp: new Date(),
        id: `msg_${Date.now()}`,
        status: 'sent'
      };
      setMessages(prev => [...prev, assistantMessage]);

    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
    } finally {
      setIsLoading(false);
    }
  };

  return {
    messages,
    sendMessage,
    isLoading,
    error
  };
}
