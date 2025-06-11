import React, { useState, useEffect } from 'react';
import { Task } from '../types';

interface EditTaskModalProps {
  task: Task | null;
  isOpen: boolean;
  onClose: () => void;
  onSave: (id: string, name: string, status: number) => Promise<{ success: boolean; message: string }>;
  onNotification: (message: string, type: 'success' | 'error') => void;
}

const EditTaskModal: React.FC<EditTaskModalProps> = ({
  task,
  isOpen,
  onClose,
  onSave,
  onNotification
}) => {
  const [taskName, setTaskName] = useState('');
  const [taskStatus, setTaskStatus] = useState(0);
  const [isSubmitting, setIsSubmitting] = useState(false);

  useEffect(() => {
    if (task) {
      setTaskName(task.name);
      setTaskStatus(task.status);
    }
  }, [task]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!task) return;
    
    const trimmedName = taskName.trim();
    if (!trimmedName) {
      onNotification('Task name cannot be empty', 'error');
      return;
    }

    setIsSubmitting(true);
    try {
      const result = await onSave(task.id, trimmedName, taskStatus);
      if (result.success) {
        onClose();
        onNotification(result.message, 'success');
      } else {
        onNotification(result.message, 'error');
      }
    } catch (error) {
      onNotification('Failed to update task', 'error');
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleBackgroundClick = (e: React.MouseEvent) => {
    if (e.target === e.currentTarget) {
      onClose();
    }
  };

  if (!isOpen || !task) {
    return null;
  }

  return (
    <div className="modal" onClick={handleBackgroundClick}>
      <div className="modal-content">
        <h3>Edit Task</h3>
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label htmlFor="edit-task-name">Task Name:</label>
            <input
              type="text"
              id="edit-task-name"
              value={taskName}
              onChange={(e) => setTaskName(e.target.value)}
              maxLength={255}
              disabled={isSubmitting}
              required
            />
          </div>
          <div className="form-group">
            <label htmlFor="edit-task-status">Status:</label>
            <select
              id="edit-task-status"
              value={taskStatus}
              onChange={(e) => setTaskStatus(parseInt(e.target.value))}
              disabled={isSubmitting}
            >
              <option value={0}>Incomplete</option>
              <option value={1}>Completed</option>
            </select>
          </div>
          <div className="modal-actions">
            <button type="button" onClick={onClose} disabled={isSubmitting}>
              Cancel
            </button>
            <button type="submit" disabled={isSubmitting}>
              {isSubmitting ? 'Saving...' : 'Save'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default EditTaskModal;