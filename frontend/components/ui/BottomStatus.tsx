import { SystemStatus } from '../../types/chat/index';

interface BottomStatusProps {
  systemStatus: SystemStatus;
  isConnected: boolean;
  error?: string | null;
}

export function BottomStatus({ systemStatus, isConnected, error }: BottomStatusProps) {
  return (
    <div className="bottom-status">
      <div className="bottom-status-left">
        <span>
          NEURAL_LINK: {isConnected ? 'ACTIVE' : 'STANDBY'}
        </span>
        <span>
          SESSION: {systemStatus.sessionActive ? 'ACTIVE' : 'INACTIVE'}
        </span>
        <span>
          ENCRYPTION: {systemStatus.encryptionActive ? 'AES-256' : 'DISABLED'}
        </span>
        {error && (
          <span className="error-status">
            ERROR: {error.toUpperCase()}
          </span>
        )}
      </div>
      <div className="bottom-status-right">
        <span className="cpu-status">
          🔥 CPU: {systemStatus.cpuUsage}%
        </span>
        <span className="mem-status">
          📊 MEM: {systemStatus.memoryUsage}%
        </span>
      </div>
    </div>
  );
}