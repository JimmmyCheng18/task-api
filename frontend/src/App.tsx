import React, { useState } from 'react';
import { Task } from './types';
import { useTasks } from './hooks/useTasks';
import Header from './components/Header';
import TaskForm from './components/TaskForm';
import TaskFilter from './components/TaskFilter';
import TaskList from './components/TaskList';
import Pagination from './components/Pagination';
import EditTaskModal from './components/EditTaskModal';
import Notification from './components/Notification';

interface NotificationState {
  message: string;
  type: 'success' | 'error';
}

const App: React.FC = () => {
  const {
    tasks,
    stats,
    loading,
    error,
    currentFilter,
    currentPage,
    totalTasks,
    limit,
    createTask,
    updateTask,
    deleteTask,
    toggleTaskStatus,
    changeFilter,
    changePage
  } = useTasks();

  const [editingTask, setEditingTask] = useState<Task | null>(null);
  const [notification, setNotification] = useState<NotificationState | null>(null);

  const showNotification = (message: string, type: 'success' | 'error') => {
    setNotification({ message, type });
  };

  const hideNotification = () => {
    setNotification(null);
  };

  const handleToggleStatus = async (task: Task) => {
    const result = await toggleTaskStatus(task);
    if (result) {
      showNotification(result.message, result.success ? 'success' : 'error');
    }
  };

  const handleEditTask = (task: Task) => {
    setEditingTask(task);
  };

  const handleCloseEditModal = () => {
    setEditingTask(null);
  };

  const handleSaveTask = async (id: string, name: string, status: number) => {
    return await updateTask(id, name, status);
  };

  const handleDeleteTask = async (id: string) => {
    const result = await deleteTask(id);
    if (result) {
      showNotification(result.message, result.success ? 'success' : 'error');
    }
  };

  const showPagination = currentFilter === 'all' && totalTasks > limit;

  return (
    <div className="container">
      <Header 
        stats={stats} 
        onNotification={showNotification}
      />
      
      <main>
        <TaskForm 
          onSubmit={createTask}
          onNotification={showNotification}
        />
        
        <TaskFilter 
          currentFilter={currentFilter}
          onFilterChange={changeFilter}
        />
        
        <TaskList 
          tasks={tasks}
          loading={loading}
          error={error}
          onToggleStatus={handleToggleStatus}
          onEdit={handleEditTask}
          onDelete={handleDeleteTask}
        />
        
        <Pagination 
          currentPage={currentPage}
          totalTasks={totalTasks}
          limit={limit}
          onPageChange={changePage}
          showPagination={showPagination}
        />
      </main>

      <EditTaskModal
        task={editingTask}
        isOpen={!!editingTask}
        onClose={handleCloseEditModal}
        onSave={handleSaveTask}
        onNotification={showNotification}
      />

      {notification && (
        <Notification
          message={notification.message}
          type={notification.type}
          onClose={hideNotification}
        />
      )}
    </div>
  );
};

export default App;