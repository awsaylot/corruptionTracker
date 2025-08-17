import { useEffect, useRef } from 'react';
import { ChatUIMessage } from '../../types/chat/ui';
import { ChatMessage } from './ChatMessage';

interface MessageListProps {
  messages: ChatUIMessage[];
  onRetry: () => void;
}

export function MessageList({ messages, onRetry }: MessageListProps) {
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  return (
    <div className="chat-messages">
      {messages.map((message) => (
        <ChatMessage
          key={message.id}
          message={message}
          onRetry={message.status === 'error' ? onRetry : undefined}
        />
      ))}
      <div ref={messagesEndRef} />
    </div>
  );
}