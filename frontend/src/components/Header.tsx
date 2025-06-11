import React, { useState } from 'react';
import { TaskStats } from '../types';
import { ApiConfig } from '../services/api';
import apiService from '../services/api';

interface HeaderProps {
  stats: TaskStats | null;
  onNotification: (message: string, type: 'success' | 'error') => void;
}

const Header: React.FC<HeaderProps> = ({ stats, onNotification }) => {
  const [port, setPort] = useState(ApiConfig.getCurrentPort());

  const handleUpdatePort = () => {
    if (port && port > 0 && port <= 65535) {
      ApiConfig.updatePort(port);
      onNotification(`API port updated to ${port}`, 'success');
    } else {
      onNotification('Please enter a valid port number (1-65535)', 'error');
    }
  };

  const handleTestConnection = async () => {
    const result = await apiService.testConnection();
    onNotification(result.message, result.success ? 'success' : 'error');
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleUpdatePort();
    }
  };

  return (
    <header className="header">
      <h1>📝 Task Management System</h1>
      
      {stats && (
        <div className="stats">
          <div className="stat-item">
            <div className="stat-number">{stats.total_tasks}</div>
            <div className="stat-label">Total</div>
          </div>
          <div className="stat-item">
            <div className="stat-number">{stats.incomplete_tasks}</div>
            <div className="stat-label">Incomplete</div>
          </div>
          <div className="stat-item">
            <div className="stat-number">{stats.completed_tasks}</div>
            <div className="stat-label">Completed</div>
          </div>
        </div>
      )}

      <div className="port-config">
        <div className="port-settings">
          <label htmlFor="api-port">API Port:</label>
          <input
            type="number"
            id="api-port"
            value={port}
            onChange={(e) => setPort(parseInt(e.target.value))}
            onKeyPress={handleKeyPress}
            min="1"
            max="65535"
          />
          <button onClick={handleUpdatePort}>Update</button>
          <button id="test-connection" onClick={handleTestConnection}>
            Test Connection
          </button>
        </div>
      </div>
    </header>
  );
};

export default Header;