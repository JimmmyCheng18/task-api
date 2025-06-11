import React from 'react';
import { Task } from '../types';

interface TaskItemProps {
  task: Task;
  onToggleStatus: (task: Task) => void;
  onEdit: (task: Task) => void;
  onDelete: (id: string) => void;
}

const TaskItem: React.FC<TaskItemProps> = ({ task, onToggleStatus, onEdit, onDelete }) => {
  const isCompleted = task.status === 1;
  const createdDate = new Date(task.created_at).toLocaleDateString();
  const updatedDate = new Date(task.updated_at).toLocaleDateString();

  const handleDelete = () => {
    if (window.confirm('Are you sure you want to delete this task?')) {
      onDelete(task.id);
    }
  };

  return (
    <div className="task-item">
      <div 
        className={`task-checkbox ${isCompleted ? 'completed' : ''}`}
        onClick={() => onToggleStatus(task)}
      />
      <div className="task-content">
        <div className={`task-name ${isCompleted ? 'completed' : ''}`}>
          {task.name}
        </div>
        <div className="task-meta">
          Created: {createdDate} | Updated: {updatedDate}
        </div>
      </div>
      <div className="task-actions">
        <button className="edit-btn" onClick={() => onEdit(task)}>
          Edit
        </button>
        <button className="delete-btn" onClick={handleDelete}>
          Delete
        </button>
      </div>
    </div>
  );
};

export default TaskItem;