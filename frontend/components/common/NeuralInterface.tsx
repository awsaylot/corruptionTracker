'use client';

import { useState } from 'react';
import { useTime } from '../../hooks/useTime';
import { useConnection } from '../../hooks/useConnection';
import { Header } from '../ui/Header';
import { StatusBar } from '../ui/StatusBar';
import { Terminal } from '../ui/Terminal';
import { ChatContainer } from '../chat/ChatContainer';

export function NeuralInterface() {
  const { currentTime, currentDate } = useTime();
  const { connectionStatus, systemStatus, isConnected } = useConnection();
  const [lastError, setLastError] = useState<string | null>(null);
  const [chatConnectionStatus, setChatConnectionStatus] = useState<'connected' | 'disconnected' | 'error'>('disconnected');

  return (
    <div className="neural-interface">
      {/* Header */}
      <Header 
        currentTime={currentTime}
        currentDate={currentDate}
      />

      {/* Status Bar */}
      <StatusBar 
        connectionStatus={connectionStatus}
        isConnected={isConnected}
      />

      {/* Main Terminal Content */}
      <Terminal 
        currentTime={currentTime}
        isConnected={isConnected}
      >
        <div className="terminal-messages">
          {/* This space will show chat messages */}
        </div>
      </Terminal>

      {/* Bottom Terminal with Chat */}
      <div className="bottom-terminal">
        <ChatContainer 
          systemStatus={systemStatus}
          setConnectionStatus={setChatConnectionStatus}
          setLastError={setLastError}
        />
      </div>
    </div>
  );
}