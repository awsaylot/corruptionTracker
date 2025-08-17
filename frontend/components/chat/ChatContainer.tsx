import React from 'react';
import { MessageList } from './MessageList';
import { InputForm } from '../ui/InputForm';
import { BottomStatus } from '../ui/BottomStatus';
import { useChat } from '../../hooks/useChat';

interface ChatContainerProps {
  systemStatus: any;
  setConnectionStatus: (status: 'connected' | 'disconnected' | 'error') => void;
  setLastError: (err: string | null) => void;
}

export function ChatContainer({ systemStatus }: ChatContainerProps) {
  const {
    messages,
    isLoading,
    error,
    connectionStatus,
    sendMessage,
    retryLastMessage,
    clearMessages
  } = useChat();

  return (
    <div className="flex flex-col h-full">
      <div className="flex-1 overflow-y-auto p-4">
        <MessageList messages={messages} onRetry={retryLastMessage} />
      </div>

      <div className="border-t bg-white p-4">
        <InputForm 
          onSubmit={sendMessage} 
          isLoading={isLoading} 
          isConnected={connectionStatus === 'connected'} 
        />
        <div className="mt-2">
          <BottomStatus 
            systemStatus={systemStatus} 
            isConnected={connectionStatus === 'connected'} 
            error={error} 
          />
        </div>

        {messages.length > 0 && (
          <div className="mt-2 text-xs">
            <button
              onClick={clearMessages}
              className="px-2 py-1 text-red-600 border border-red-600 bg-red-50 hover:bg-red-100 rounded"
            >
              Clear Chat
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
