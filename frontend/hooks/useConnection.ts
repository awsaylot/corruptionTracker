import { useState, useEffect } from 'react';
import { ConnectionStatus, SystemStatus } from '../types/chat/index';
import { UI_CONSTANTS } from '../utils/constants';

export function useConnection() {
  const [connectionStatus, setConnectionStatus] = useState<ConnectionStatus>('connecting');
  const [systemStatus, setSystemStatus] = useState<SystemStatus>({
    cpuUsage: 23,
    memoryUsage: 67,
    neuralCore: false,
    encryptionActive: true,
    sessionActive: true,
  });

  useEffect(() => {
    // Simulate connection process
    const connectTimer = setTimeout(() => {
      setConnectionStatus('connected');
      setSystemStatus(prev => ({ ...prev, neuralCore: true }));
    }, UI_CONSTANTS.CONNECTION_TIMEOUT);

    // Simulate system metrics updates
    const metricsTimer = setInterval(() => {
      setSystemStatus(prev => ({
        ...prev,
        cpuUsage: Math.floor(Math.random() * 30) + 20, // 20-50%
        memoryUsage: Math.floor(Math.random() * 20) + 60, // 60-80%
      }));
    }, 5000);

    return () => {
      clearTimeout(connectTimer);
      clearInterval(metricsTimer);
    };
  }, []);

  const disconnect = () => {
    setConnectionStatus('disconnected');
    setSystemStatus(prev => ({ ...prev, neuralCore: false }));
  };

  const reconnect = () => {
    setConnectionStatus('connecting');
    setTimeout(() => {
      setConnectionStatus('connected');
      setSystemStatus(prev => ({ ...prev, neuralCore: true }));
    }, UI_CONSTANTS.CONNECTION_TIMEOUT);
  };

  return {
    connectionStatus,
    systemStatus,
    disconnect,
    reconnect,
    isConnected: connectionStatus === 'connected',
  };
}