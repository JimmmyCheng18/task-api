import { Task, ApiResponse, TaskStats, TaskFormData } from '../types';

class ApiConfig {
  private static defaultPort = 3333;  // 後端 API 端口
  private static baseUrl: string = '';

  private static getApiBaseUrl(): string {
    // Get current hostname from window.location
    const currentHostname = window.location.hostname;
    
    // Check URL parameters for custom port
    const urlParams = new URLSearchParams(window.location.search);
    const customPort = urlParams.get('port');
    
    // Check localStorage for saved port
    const savedPort = localStorage.getItem('api-port');
    
    // Use custom port, saved port, or default port
    const port = customPort || savedPort || this.defaultPort;
    
    // Save port to localStorage if it's from URL params
    if (customPort) {
      localStorage.setItem('api-port', customPort);
    }
    
    // Use the same hostname as the current page to avoid CORS issues
    return `http://${currentHostname}:${port}/api/v1`;
  }

  static getBaseUrl(): string {
    if (!this.baseUrl) {
      this.baseUrl = this.getApiBaseUrl();
    }
    return this.baseUrl;
  }

  static updatePort(port: number): void {
    this.baseUrl = `http://localhost:${port}/api/v1`;
    localStorage.setItem('api-port', port.toString());
  }

  static getCurrentPort(): number {
    const url = new URL(this.getBaseUrl());
    return parseInt(url.port) || this.defaultPort;
  }
}

class ApiService {
  private async apiRequest<T>(endpoint: string, options: RequestInit = {}): Promise<ApiResponse<T>> {
    try {
      const response = await fetch(`${ApiConfig.getBaseUrl()}${endpoint}`, {
        headers: {
          'Content-Type': 'application/json',
          ...options.headers,
        },
        ...options,
      });

      const data: ApiResponse<T> = await response.json();
      return data;
    } catch (error) {
      console.error('API request failed:', error);
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Unknown error occurred'
      };
    }
  }

  async getTasks(offset = 0, limit = 10): Promise<ApiResponse<Task[]>> {
    return this.apiRequest<Task[]>(`/tasks/paginated?offset=${offset}&limit=${limit}`);
  }

  async getTasksByStatus(status: number): Promise<ApiResponse<Task[]>> {
    return this.apiRequest<Task[]>(`/tasks/status/${status}`);
  }

  async createTask(taskData: TaskFormData): Promise<ApiResponse<Task>> {
    return this.apiRequest<Task>('/tasks', {
      method: 'POST',
      body: JSON.stringify(taskData),
    });
  }

  async updateTask(id: string, taskData: TaskFormData): Promise<ApiResponse<Task>> {
    return this.apiRequest<Task>(`/tasks/${id}`, {
      method: 'PUT',
      body: JSON.stringify(taskData),
    });
  }

  async deleteTask(id: string): Promise<ApiResponse<void>> {
    return this.apiRequest(`/tasks/${id}`, {
      method: 'DELETE',
    });
  }

  async getStats(): Promise<ApiResponse<TaskStats>> {
    return this.apiRequest<TaskStats>('/stats');
  }

  async testConnection(): Promise<{ success: boolean; message: string }> {
    try {
      const currentHostname = window.location.hostname;
      const response = await fetch(`http://${currentHostname}:${ApiConfig.getCurrentPort()}/health`);
      if (response.ok) {
        const data = await response.json();
        return {
          success: true,
          message: `Connection successful! API version: ${data.version || 'unknown'}`
        };
      } else {
        return {
          success: false,
          message: `Connection failed: ${response.status} ${response.statusText}`
        };
      }
    } catch (error) {
      return {
        success: false,
        message: `Connection failed: ${error instanceof Error ? error.message : 'Unknown error'}`
      };
    }
  }
}

export { ApiConfig, ApiService };
export default new ApiService();