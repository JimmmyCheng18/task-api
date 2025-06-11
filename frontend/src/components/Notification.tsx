import React, { useEffect } from 'react';

interface NotificationProps {
  message: string;
  type: 'success' | 'error';
  onClose: () => void;
}

const Notification: React.FC<NotificationProps> = ({ message, type, onClose }) => {
  useEffect(() => {
    const timer = setTimeout(() => {
      onClose();
    }, type === 'success' ? 3000 : 5000);

    return () => clearTimeout(timer);
  }, [onClose, type]);

  const baseStyles: React.CSSProperties = {
    position: 'fixed',
    top: '20px',
    right: '20px',
    zIndex: 1001,
    maxWidth: '300px',
    padding: '16px',
    borderRadius: '8px',
    boxShadow: '0 4px 12px rgba(0, 0, 0, 0.1)',
    cursor: 'pointer'
  };

  const typeStyles: React.CSSProperties = type === 'success'
    ? { background: '#c6f6d5', color: '#276749', border: '1px solid #9ae6b4' }
    : { background: '#fed7d7', color: '#c53030', border: '1px solid #feb2b2' };

  return (
    <div
      style={{ ...baseStyles, ...typeStyles }}
      onClick={onClose}
    >
      {message}
    </div>
  );
};

export default Notification;