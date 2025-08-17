import { ConnectionStatus } from '../../types/chat';
import { UI_CONSTANTS } from '../../utils/constants';

interface StatusBarProps {
  connectionStatus: ConnectionStatus;
  isConnected: boolean;
}

export function StatusBar({ connectionStatus, isConnected }: StatusBarProps) {
  const getStatusText = (status: ConnectionStatus) => {
    switch (status) {
      case 'connecting':
        return 'CONNECTING';
      case 'connected':
        return 'ONLINE';
      case 'disconnected':
        return 'OFFLINE';
      case 'error':
        return 'ERROR';
      default:
        return 'UNKNOWN';
    }
  };

  return (
    <div className="status-bar">
      <div className="status-item">
        <span className="status-icon">⚡</span>
        <span>{UI_CONSTANTS.MODEL_NAME}</span>
      </div>
      <div className="status-item">
        <span className="status-icon">⚡</span>
        <span>STATUS: {getStatusText(connectionStatus)}</span>
      </div>
      <div className="status-item">
        <span className="status-icon">◯</span>
        <span>SECURE_CHANNEL</span>
      </div>
      <div className="status-item ready">
        <span>{isConnected ? 'READY' : 'STANDBY'}</span>
        <div className="ready-indicator"></div>
      </div>
      <div className="status-right">
        <span className="quantum-core">⚛ QUANTUM_CORE</span>
      </div>
    </div>
  );
}