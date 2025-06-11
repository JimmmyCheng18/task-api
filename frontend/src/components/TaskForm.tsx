import React, { useState } from 'react';

interface TaskFormProps {
  onSubmit: (name: string) => Promise<{ success: boolean; message: string }>;
  onNotification: (message: string, type: 'success' | 'error') => void;
}

const TaskForm: React.FC<TaskFormProps> = ({ onSubmit, onNotification }) => {
  const [taskName, setTaskName] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    const trimmedName = taskName.trim();
    if (!trimmedName) {
      onNotification('Task name cannot be empty', 'error');
      return;
    }

    setIsSubmitting(true);
    try {
      const result = await onSubmit(trimmedName);
      if (result.success) {
        setTaskName('');
        onNotification(result.message, 'success');
      } else {
        onNotification(result.message, 'error');
      }
    } catch (error) {
      onNotification('Failed to create task', 'error');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <section className="add-task-section">
      <h2>Add New Task</h2>
      <form onSubmit={handleSubmit} className="add-task-form">
        <div className="form-group">
          <input
            type="text"
            value={taskName}
            onChange={(e) => setTaskName(e.target.value)}
            placeholder="Enter task name..."
            maxLength={255}
            disabled={isSubmitting}
            required
          />
          <button type="submit" disabled={isSubmitting}>
            {isSubmitting ? 'Adding...' : 'Add Task'}
          </button>
        </div>
      </form>
    </section>
  );
};

export default TaskForm;