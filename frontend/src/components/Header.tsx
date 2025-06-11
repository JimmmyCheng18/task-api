import React from 'react';
import { TaskStats } from '../types';
import apiService from '../services/api';

interface HeaderProps {
  stats: TaskStats | null;
  onNotification: (message: string, type: 'success' | 'error') => void;
}

const Header: React.FC<HeaderProps> = ({ stats, onNotification }) => {
  const handleTestConnection = async () => {
    const result = await apiService.testConnection();
    onNotification(result.message, result.success ? 'success' : 'error');
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
          <button id="test-connection" onClick={handleTestConnection}>
            Test Connection
          </button>
        </div>
      </div>
    </header>
  );
};

export default Header;