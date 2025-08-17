import React, { useState } from 'react';
import MainLayout from '../../components/layout/MainLayout';
import { ChatContainer } from '../../components/chat/ChatContainer';

const ChatPage: React.FC = () => {
  const [connectionStatus, setConnectionStatus] = useState<'connected' | 'disconnected' | 'error'>('disconnected');
  const [lastError, setLastError] = useState<string | null>(null);

  const systemStatus = {
    state: {
      connection: connectionStatus,
      error: lastError
    },
    resources: {
      cpuUsage: 0,
      memoryUsage: 0,
      neuralCore: true,
      encryptionActive: true
    }
  };

  return (
    <MainLayout>
      <div className="flex flex-col h-[calc(100vh-4rem)]">
        <div className="flex-1 min-h-0 p-4">
          <div className="bg-white shadow-lg rounded-lg h-full flex flex-col">
            <div className="p-4 border-b">
              <h1 className="text-xl font-semibold text-gray-800">
                AI Assistant Chat
              </h1>
              <p className="text-sm text-gray-500">
                Chat with the AI to analyze corruption networks and patterns
              </p>
            </div>

            <ChatContainer
              systemStatus={systemStatus}
              setConnectionStatus={setConnectionStatus}
              setLastError={setLastError}
            />
          </div>
        </div>
      </div>
    </MainLayout>
  );
};

export default ChatPage;
