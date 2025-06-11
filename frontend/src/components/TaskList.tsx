import React from 'react';
import { Task } from '../types';
import TaskItem from './TaskItem';

interface TaskListProps {
  tasks: Task[];
  loading: boolean;
  error: string | null;
  onToggleStatus: (task: Task) => void;
  onEdit: (task: Task) => void;
  onDelete: (id: string) => void;
}

const TaskList: React.FC<TaskListProps> = ({
  tasks,
  loading,
  error,
  onToggleStatus,
  onEdit,
  onDelete
}) => {
  if (loading) {
    return (
      <section className="tasks-section">
        <h2>Task List</h2>
        <div className="loading">Loading...</div>
      </section>
    );
  }

  if (error) {
    return (
      <section className="tasks-section">
        <h2>Task List</h2>
        <div className="error">{error}</div>
      </section>
    );
  }

  if (tasks.length === 0) {
    return (
      <section className="tasks-section">
        <h2>Task List</h2>
        <div className="empty-state">
          <p>📋 No tasks available</p>
          <p>Add your first task to get started!</p>
        </div>
      </section>
    );
  }

  return (
    <section className="tasks-section">
      <h2>Task List</h2>
      <div className="tasks-container">
        {tasks.map(task => (
          <TaskItem
            key={task.id}
            task={task}
            onToggleStatus={onToggleStatus}
            onEdit={onEdit}
            onDelete={onDelete}
          />
        ))}
      </div>
    </section>
  );
};

export default TaskList;