package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"task-api/internal/models"
	"task-api/internal/storage"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestHandler creates a handler with memory storage for testing
func setupTestHandler() (*TaskHandler, *gin.Engine) {
	// Create memory storage
	memStorage := storage.NewMemoryStorage(1000)

	// Create handler
	handler := NewTaskHandler(memStorage)

	// Setup Gin in test mode
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Register routes
	api := router.Group("/api/v1")
	{
		api.GET("/tasks", handler.GetAllTasks)
		api.GET("/tasks/:id", handler.GetTaskByID)
		api.POST("/tasks", handler.CreateTask)
		api.PUT("/tasks/:id", handler.UpdateTask)
		api.DELETE("/tasks/:id", handler.DeleteTask)
		api.GET("/tasks/status/:status", handler.GetTasksByStatus)
		api.GET("/tasks/paginated", handler.GetTasksPaginated)
		api.GET("/health", handler.HealthCheck)
		api.GET("/stats", handler.GetStorageStats)
	}

	return handler, router
}

// setupBenchmarkHandler creates a handler with larger storage limit for benchmarks
func setupBenchmarkHandler() (*TaskHandler, *gin.Engine) {
	// Create memory storage with much larger limit for benchmarks
	memStorage := storage.NewMemoryStorage(100000)

	// Create handler
	handler := NewTaskHandler(memStorage)

	// Setup Gin in test mode
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Register routes
	api := router.Group("/api/v1")
	{
		api.GET("/tasks", handler.GetAllTasks)
		api.GET("/tasks/:id", handler.GetTaskByID)
		api.POST("/tasks", handler.CreateTask)
		api.PUT("/tasks/:id", handler.UpdateTask)
		api.DELETE("/tasks/:id", handler.DeleteTask)
		api.GET("/tasks/status/:status", handler.GetTasksByStatus)
		api.GET("/tasks/paginated", handler.GetTasksPaginated)
		api.GET("/health", handler.HealthCheck)
		api.GET("/stats", handler.GetStorageStats)
	}

	return handler, router
}

// createTestTask is a helper to create a task for testing
func createTestTask(tb testing.TB, handler *TaskHandler, name string, status models.TaskStatus) *models.Task {
	req := &models.CreateTaskRequest{
		Name:   name,
		Status: status,
	}

	task, err := handler.storage.Create(req)
	require.NoError(tb, err)

	return task
}

func TestNewTaskHandler(t *testing.T) {
	storage := storage.NewMemoryStorage(1000)
	handler := NewTaskHandler(storage)

	assert.NotNil(t, handler)
	assert.Equal(t, storage, handler.storage)
}

func TestTaskHandler_GetAllTasks(t *testing.T) {
	handler, router := setupTestHandler()

	tests := []struct {
		name           string
		setupTasks     func()
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "empty storage",
			setupTasks:     func() {},
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			name: "with tasks",
			setupTasks: func() {
				createTestTask(t, handler, "Task 1", models.TaskIncomplete)
				createTestTask(t, handler, "Task 2", models.TaskCompleted)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear storage and setup test data
			// Clear storage for test setup (ignore error in test context)
			_ = handler.storage.Clear()
			tt.setupTasks()

			// Make request
			req, _ := http.NewRequest("GET", "/api/v1/tasks", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify response
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response models.TaskListResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.True(t, response.Success)
			assert.Equal(t, tt.expectedCount, response.Count)
			assert.Len(t, response.Data, tt.expectedCount)
		})
	}
}

func TestTaskHandler_GetTaskByID(t *testing.T) {
	handler, router := setupTestHandler()

	// Create a test task
	task := createTestTask(t, handler, "Test Task", models.TaskIncomplete)

	tests := []struct {
		name           string
		taskID         string
		expectedStatus int
		expectedFound  bool
	}{
		{
			name:           "existing task",
			taskID:         task.ID,
			expectedStatus: http.StatusOK,
			expectedFound:  true,
		},
		{
			name:           "non-existing task",
			taskID:         "999",
			expectedStatus: http.StatusNotFound,
			expectedFound:  false,
		},
		{
			name:           "empty task ID",
			taskID:         "",
			expectedStatus: http.StatusMovedPermanently, // Gin redirects for empty params
			expectedFound:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/api/v1/tasks/%s", tt.taskID)
			req, _ := http.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedFound {
				var response models.TaskResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.True(t, response.Success)
				assert.NotNil(t, response.Data)
				assert.Equal(t, task.ID, response.Data.ID)
				assert.Equal(t, task.Name, response.Data.Name)
			} else if tt.expectedStatus == http.StatusMovedPermanently {
				// For redirect responses, we don't expect JSON
				assert.Contains(t, w.Body.String(), "Moved Permanently")
			} else {
				var response models.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.False(t, response.Success)
				assert.NotEmpty(t, response.Message)
			}
		})
	}
}

func TestTaskHandler_CreateTask(t *testing.T) {
	_, router := setupTestHandler()

	tests := []struct {
		name           string
		request        interface{}
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "valid task",
			request: models.CreateTaskRequest{
				Name:   "Test Task",
				Status: models.TaskIncomplete,
			},
			expectedStatus: http.StatusCreated,
			expectedError:  false,
		},
		{
			name: "valid completed task",
			request: models.CreateTaskRequest{
				Name:   "Completed Task",
				Status: models.TaskCompleted,
			},
			expectedStatus: http.StatusCreated,
			expectedError:  false,
		},
		{
			name: "empty name",
			request: models.CreateTaskRequest{
				Name:   "",
				Status: models.TaskIncomplete,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "invalid status",
			request: models.CreateTaskRequest{
				Name:   "Test Task",
				Status: models.TaskStatus(99),
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:           "invalid JSON",
			request:        `{"name": "Test", "status": "invalid"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:           "missing required field",
			request:        map[string]interface{}{"status": 0},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error

			if str, ok := tt.request.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.request)
				require.NoError(t, err)
			}

			req, _ := http.NewRequest("POST", "/api/v1/tasks", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError {
				var response models.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.False(t, response.Success)
				assert.NotEmpty(t, response.Message)
			} else {
				var response models.TaskResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.True(t, response.Success)
				assert.NotNil(t, response.Data)
				assert.NotEmpty(t, response.Data.ID)
				assert.False(t, response.Data.CreatedAt.IsZero())
			}
		})
	}
}

func TestTaskHandler_UpdateTask(t *testing.T) {
	handler, router := setupTestHandler()

	// Create a test task
	task := createTestTask(t, handler, "Original Task", models.TaskIncomplete)

	tests := []struct {
		name           string
		taskID         string
		request        interface{}
		expectedStatus int
		expectedError  bool
	}{
		{
			name:   "update name only",
			taskID: task.ID,
			request: models.UpdateTaskRequest{
				Name: stringPtr("Updated Task"),
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:   "update status only",
			taskID: task.ID,
			request: models.UpdateTaskRequest{
				Status: taskStatusPtr(models.TaskCompleted),
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:   "update both fields",
			taskID: task.ID,
			request: models.UpdateTaskRequest{
				Name:   stringPtr("Fully Updated Task"),
				Status: taskStatusPtr(models.TaskIncomplete),
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:   "non-existing task",
			taskID: "999",
			request: models.UpdateTaskRequest{
				Name: stringPtr("Updated Task"),
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  true,
		},
		{
			name:           "empty update",
			taskID:         task.ID,
			request:        models.UpdateTaskRequest{},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:   "invalid name",
			taskID: task.ID,
			request: models.UpdateTaskRequest{
				Name: stringPtr(""),
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:   "invalid status",
			taskID: task.ID,
			request: models.UpdateTaskRequest{
				Status: taskStatusPtr(models.TaskStatus(99)),
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.request)
			require.NoError(t, err)

			url := fmt.Sprintf("/api/v1/tasks/%s", tt.taskID)
			req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError {
				var response models.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.False(t, response.Success)
				assert.NotEmpty(t, response.Message)
			} else {
				var response models.TaskResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.True(t, response.Success)
				assert.NotNil(t, response.Data)
				assert.Equal(t, tt.taskID, response.Data.ID)
			}
		})
	}
}

func TestTaskHandler_DeleteTask(t *testing.T) {
	handler, router := setupTestHandler()

	// Create a test task
	task := createTestTask(t, handler, "Test Task", models.TaskIncomplete)

	tests := []struct {
		name           string
		taskID         string
		expectedStatus int
		expectedError  bool
	}{
		{
			name:           "existing task",
			taskID:         task.ID,
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:           "non-existing task",
			taskID:         "999",
			expectedStatus: http.StatusNotFound,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/api/v1/tasks/%s", tt.taskID)
			req, _ := http.NewRequest("DELETE", url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError {
				var response models.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.False(t, response.Success)
				assert.NotEmpty(t, response.Message)
			} else {
				var response models.TaskResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.True(t, response.Success)
				assert.Nil(t, response.Data)

				// Verify task is actually deleted
				_, err = handler.storage.GetByID(tt.taskID)
				assert.Error(t, err)
			}
		})
	}
}

func TestTaskHandler_GetTasksByStatus(t *testing.T) {
	handler, router := setupTestHandler()

	// Create test tasks with different statuses
	createTestTask(t, handler, "Incomplete Task 1", models.TaskIncomplete)
	createTestTask(t, handler, "Incomplete Task 2", models.TaskIncomplete)
	createTestTask(t, handler, "Completed Task", models.TaskCompleted)

	tests := []struct {
		name           string
		status         string
		expectedStatus int
		expectedCount  int
		expectedError  bool
	}{
		{
			name:           "get incomplete tasks",
			status:         "0",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
			expectedError:  false,
		},
		{
			name:           "get completed tasks",
			status:         "1",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			expectedError:  false,
		},
		{
			name:           "invalid status",
			status:         "2",
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:           "invalid status format",
			status:         "invalid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/api/v1/tasks/status/%s", tt.status)
			req, _ := http.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError {
				var response models.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.False(t, response.Success)
			} else {
				var response models.TaskListResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.True(t, response.Success)
				assert.Equal(t, tt.expectedCount, response.Count)
				assert.Len(t, response.Data, tt.expectedCount)

				// Verify all tasks have the expected status
				expectedTaskStatus := models.TaskStatus(mustAtoi(tt.status))
				for _, task := range response.Data {
					assert.Equal(t, expectedTaskStatus, task.Status)
				}
			}
		})
	}
}

func TestTaskHandler_GetTasksPaginated(t *testing.T) {
	handler, router := setupTestHandler()

	// Create 10 test tasks
	for i := 0; i < 10; i++ {
		createTestTask(t, handler, fmt.Sprintf("Task %d", i+1), models.TaskIncomplete)
	}

	tests := []struct {
		name           string
		offset         string
		limit          string
		expectedStatus int
		expectedCount  int
		expectedTotal  string
		expectedError  bool
	}{
		{
			name:           "first page",
			offset:         "0",
			limit:          "5",
			expectedStatus: http.StatusOK,
			expectedCount:  5,
			expectedTotal:  "10",
			expectedError:  false,
		},
		{
			name:           "second page",
			offset:         "5",
			limit:          "5",
			expectedStatus: http.StatusOK,
			expectedCount:  5,
			expectedTotal:  "10",
			expectedError:  false,
		},
		{
			name:           "partial page",
			offset:         "8",
			limit:          "5",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
			expectedTotal:  "10",
			expectedError:  false,
		},
		{
			name:           "beyond data",
			offset:         "15",
			limit:          "5",
			expectedStatus: http.StatusOK,
			expectedCount:  0,
			expectedTotal:  "10",
			expectedError:  false,
		},
		{
			name:           "default pagination",
			offset:         "",
			limit:          "",
			expectedStatus: http.StatusOK,
			expectedCount:  10,
			expectedTotal:  "10",
			expectedError:  false,
		},
		{
			name:           "invalid offset",
			offset:         "invalid",
			limit:          "5",
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:           "invalid limit",
			offset:         "0",
			limit:          "invalid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:           "limit too large",
			offset:         "0",
			limit:          "200",
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/v1/tasks/paginated"
			if tt.offset != "" || tt.limit != "" {
				url += "?"
				if tt.offset != "" {
					url += "offset=" + tt.offset
				}
				if tt.limit != "" {
					if tt.offset != "" {
						url += "&"
					}
					url += "limit=" + tt.limit
				}
			}

			req, _ := http.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError {
				var response models.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.False(t, response.Success)
			} else {
				var response models.TaskListResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.True(t, response.Success)
				assert.Equal(t, tt.expectedCount, response.Count)
				assert.Len(t, response.Data, tt.expectedCount)

				// Check headers
				assert.Equal(t, tt.expectedTotal, w.Header().Get("X-Total-Count"))
			}
		})
	}
}

func TestTaskHandler_HealthCheck(t *testing.T) {
	_, router := setupTestHandler()

	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "healthy", response.Status)
	assert.Equal(t, "1.0.0", response.Version)
	assert.False(t, response.Timestamp.IsZero())
}

func TestTaskHandler_GetStorageStats(t *testing.T) {
	handler, router := setupTestHandler()

	// Create some test tasks
	createTestTask(t, handler, "Incomplete Task", models.TaskIncomplete)
	createTestTask(t, handler, "Completed Task", models.TaskCompleted)

	req, _ := http.NewRequest("GET", "/api/v1/stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response["success"].(bool))
	assert.NotNil(t, response["data"])

	data := response["data"].(map[string]interface{})
	assert.Equal(t, float64(2), data["total_tasks"])
	assert.Equal(t, "sharded_memory", data["storage_type"])
}

// Helper functions for tests
func stringPtr(s string) *string {
	return &s
}

func taskStatusPtr(status models.TaskStatus) *models.TaskStatus {
	return &status
}

func mustAtoi(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return i
}

// API Benchmark Tests for High Concurrency

// BenchmarkAPI_CreateTask tests task creation performance
func BenchmarkAPI_CreateTask(b *testing.B) {
	_, router := setupBenchmarkHandler()

	requestBody := `{"name":"Benchmark Task","status":0}`

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, _ := http.NewRequest("POST", "/api/v1/tasks", bytes.NewBufferString(requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusCreated {
				b.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
			}
		}
	})
}

// BenchmarkAPI_GetAllTasks tests getting all tasks performance
func BenchmarkAPI_GetAllTasks(b *testing.B) {
	handler, router := setupBenchmarkHandler()

	// Pre-populate with some tasks
	for i := 0; i < 100; i++ {
		createTestTask(b, handler, fmt.Sprintf("Task %d", i), models.TaskIncomplete)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, _ := http.NewRequest("GET", "/api/v1/tasks", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				b.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			}
		}
	})
}

// BenchmarkAPI_GetTaskByID tests getting task by ID performance
func BenchmarkAPI_GetTaskByID(b *testing.B) {
	handler, router := setupBenchmarkHandler()

	// Create a test task
	task := createTestTask(b, handler, "Benchmark Task", models.TaskIncomplete)
	url := fmt.Sprintf("/api/v1/tasks/%s", task.ID)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, _ := http.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				b.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			}
		}
	})
}

// BenchmarkAPI_UpdateTask tests task update performance
func BenchmarkAPI_UpdateTask(b *testing.B) {
	handler, router := setupBenchmarkHandler()

	// Create a test task
	task := createTestTask(b, handler, "Original Task", models.TaskIncomplete)
	url := fmt.Sprintf("/api/v1/tasks/%s", task.ID)
	requestBody := `{"name":"Updated Task"}`

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, _ := http.NewRequest("PUT", url, bytes.NewBufferString(requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				b.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			}
		}
	})
}

// BenchmarkAPI_HealthCheck tests health check endpoint performance
func BenchmarkAPI_HealthCheck(b *testing.B) {
	_, router := setupBenchmarkHandler()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, _ := http.NewRequest("GET", "/api/v1/health", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				b.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			}
		}
	})
}

// BenchmarkAPI_MixedWorkload tests mixed read/write operations
func BenchmarkAPI_MixedWorkload(b *testing.B) {
	handler, router := setupBenchmarkHandler()

	// Pre-populate with some tasks
	tasks := make([]*models.Task, 50)
	for i := 0; i < 50; i++ {
		tasks[i] = createTestTask(b, handler, fmt.Sprintf("Task %d", i), models.TaskIncomplete)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			switch i % 4 {
			case 0: // GET all tasks (25%)
				req, _ := http.NewRequest("GET", "/api/v1/tasks", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

			case 1: // GET task by ID (25%)
				taskID := tasks[i%len(tasks)].ID
				url := fmt.Sprintf("/api/v1/tasks/%s", taskID)
				req, _ := http.NewRequest("GET", url, nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

			case 2: // CREATE task (25%)
				body := fmt.Sprintf(`{"name":"Benchmark Task %d","status":0}`, i)
				req, _ := http.NewRequest("POST", "/api/v1/tasks", bytes.NewBufferString(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

			case 3: // UPDATE task (25%)
				taskID := tasks[i%len(tasks)].ID
				url := fmt.Sprintf("/api/v1/tasks/%s", taskID)
				body := fmt.Sprintf(`{"name":"Updated Task %d"}`, i)
				req, _ := http.NewRequest("PUT", url, bytes.NewBufferString(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
			}
			i++
		}
	})
}

// BenchmarkAPI_HighConcurrency tests API under high concurrent load
func BenchmarkAPI_HighConcurrency(b *testing.B) {
	handler, router := setupBenchmarkHandler()

	// Pre-populate with tasks
	for i := 0; i < 1000; i++ {
		createTestTask(b, handler, fmt.Sprintf("Task %d", i), models.TaskIncomplete)
	}

	b.ResetTimer()

	// Set high parallelism for concurrency testing
	b.SetParallelism(100)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Simulate high concurrent read operations
			req, _ := http.NewRequest("GET", "/api/v1/tasks", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				b.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			}
		}
	})
}
