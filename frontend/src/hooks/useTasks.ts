import { useState, useEffect, useCallback } from 'react';
import { Task, FilterType, TaskStats } from '../types';
import apiService from '../services/api';

export const useTasks = () => {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [stats, setStats] = useState<TaskStats | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [currentFilter, setCurrentFilter] = useState<FilterType>('all');
  const [currentPage, setCurrentPage] = useState(0);
  const [totalTasks, setTotalTasks] = useState(0);
  const limit = 10;

  const showError = useCallback((message: string) => {
    setError(message);
    setTimeout(() => setError(null), 5000);
  }, []);

  const loadTasks = useCallback(async () => {
    setLoading(true);
    setError(null);

    try {
      let response;
      if (currentFilter === 'all') {
        response = await apiService.getTasks(currentPage * limit, limit);
      } else {
        const status = currentFilter === 'completed' ? 1 : 0;
        response = await apiService.getTasksByStatus(status);
      }

      if (response.success && response.data) {
        setTasks(response.data);
        setTotalTasks(response.count || response.data.length);
      } else {
        showError(response.error || 'Failed to load tasks');
      }
    } catch (error) {
      showError('Failed to load tasks');
    } finally {
      setLoading(false);
    }
  }, [currentFilter, currentPage, limit, showError]);

  const loadStats = useCallback(async () => {
    try {
      const response = await apiService.getStats();
      if (response.success && response.data) {
        setStats(response.data);
      }
    } catch (error) {
      console.error('Failed to load stats:', error);
    }
  }, []);

  const createTask = useCallback(async (name: string) => {
    const response = await apiService.createTask({ name, status: 0 });
    if (response.success) {
      await loadTasks();
      await loadStats();
      return { success: true, message: 'Task created successfully' };
    } else {
      return { success: false, message: response.error || 'Failed to create task' };
    }
  }, [loadTasks, loadStats]);

  const updateTask = useCallback(async (id: string, name: string, status: number) => {
    const response = await apiService.updateTask(id, { name, status });
    if (response.success) {
      await loadTasks();
      await loadStats();
      return { success: true, message: 'Task updated successfully' };
    } else {
      return { success: false, message: response.error || 'Failed to update task' };
    }
  }, [loadTasks, loadStats]);

  const deleteTask = useCallback(async (id: string) => {
    const response = await apiService.deleteTask(id);
    if (response.success) {
      await loadTasks();
      await loadStats();
      return { success: true, message: 'Task deleted successfully' };
    } else {
      return { success: false, message: response.error || 'Failed to delete task' };
    }
  }, [loadTasks, loadStats]);

  const toggleTaskStatus = useCallback(async (task: Task) => {
    const newStatus = task.status === 0 ? 1 : 0;
    return updateTask(task.id, task.name, newStatus);
  }, [updateTask]);

  const changeFilter = useCallback((filter: FilterType) => {
    setCurrentFilter(filter);
    setCurrentPage(0);
  }, []);

  const changePage = useCallback((direction: number) => {
    const newPage = currentPage + direction;
    const maxPage = Math.ceil(totalTasks / limit) - 1;
    
    if (newPage >= 0 && newPage <= maxPage) {
      setCurrentPage(newPage);
    }
  }, [currentPage, totalTasks, limit]);

  useEffect(() => {
    loadTasks();
  }, [loadTasks]);

  useEffect(() => {
    loadStats();
  }, [loadStats]);

  return {
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
    changePage,
    loadTasks,
    loadStats
  };
};