export interface Task {
  id: string;
  name: string;
  status: number; // 0 = incomplete, 1 = completed
  created_at: string;
  updated_at: string;
}

export interface ApiResponse<T> {
  success: boolean;
  message?: string;
  data?: T;
  error?: string;
  count?: number;
}

export interface TaskStats {
  total_tasks: number;
  completed_tasks: number;
  incomplete_tasks: number;
  last_id: number;
  storage_type: string;
}

export type FilterType = 'all' | 'incomplete' | 'completed';

export interface TaskFormData {
  name: string;
  status: number;
}