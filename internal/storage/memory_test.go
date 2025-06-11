package storage

import (
	"strconv"
	"sync"
	"task-api/internal/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMemoryStorage(t *testing.T) {
	storage := NewMemoryStorage(1000)

	assert.NotNil(t, storage)
	assert.NotNil(t, storage.shards)
	assert.Equal(t, 1000, storage.maxTasks)
	assert.True(t, storage.shardCount > 0)

	// Test with zero maxTasks - should use default
	storage2 := NewMemoryStorage(0)
	assert.Equal(t, 10000, storage2.maxTasks)

	// Test with negative maxTasks - should use default
	storage3 := NewMemoryStorage(-1)
	assert.Equal(t, 10000, storage3.maxTasks)
}

func TestMemoryStorage_Create(t *testing.T) {
	storage := NewMemoryStorage(1000)

	tests := []struct {
		name    string
		request *models.CreateTaskRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid task creation",
			request: &models.CreateTaskRequest{
				Name:   "Test Task",
				Status: models.TaskIncomplete,
			},
			wantErr: false,
		},
		{
			name: "valid task with completed status",
			request: &models.CreateTaskRequest{
				Name:   "Completed Task",
				Status: models.TaskCompleted,
			},
			wantErr: false,
		},
		{
			name: "empty name should fail",
			request: &models.CreateTaskRequest{
				Name:   "",
				Status: models.TaskIncomplete,
			},
			wantErr: true,
			errMsg:  "validation failed",
		},
		{
			name: "name too long should fail",
			request: &models.CreateTaskRequest{
				Name:   string(make([]byte, 256)), // 256 characters
				Status: models.TaskIncomplete,
			},
			wantErr: true,
			errMsg:  "validation failed",
		},
		{
			name: "invalid status should fail",
			request: &models.CreateTaskRequest{
				Name:   "Test Task",
				Status: models.TaskStatus(99),
			},
			wantErr: true,
			errMsg:  "validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, err := storage.Create(tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, task)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, task)
				assert.NotEmpty(t, task.ID)
				assert.Equal(t, tt.request.Name, task.Name)
				assert.Equal(t, tt.request.Status, task.Status)
				assert.False(t, task.CreatedAt.IsZero())
				assert.False(t, task.UpdatedAt.IsZero())
			}
		})
	}
}

func TestMemoryStorage_GetAll(t *testing.T) {
	storage := NewMemoryStorage(1000)

	// Initially empty
	tasks, err := storage.GetAll()
	assert.NoError(t, err)
	assert.Empty(t, tasks)

	// Create some tasks
	req1 := &models.CreateTaskRequest{Name: "Task 1", Status: models.TaskIncomplete}
	req2 := &models.CreateTaskRequest{Name: "Task 2", Status: models.TaskCompleted}

	task1, err := storage.Create(req1)
	require.NoError(t, err)
	task2, err := storage.Create(req2)
	require.NoError(t, err)

	// Get all tasks
	tasks, err = storage.GetAll()
	assert.NoError(t, err)
	assert.Len(t, tasks, 2)

	// Verify both tasks are returned
	taskIDs := make(map[string]bool)
	for _, task := range tasks {
		taskIDs[task.ID] = true
	}
	assert.True(t, taskIDs[task1.ID])
	assert.True(t, taskIDs[task2.ID])
}

func TestMemoryStorage_GetByID(t *testing.T) {
	storage := NewMemoryStorage(1000)

	// Create a task
	req := &models.CreateTaskRequest{Name: "Test Task", Status: models.TaskIncomplete}
	createdTask, err := storage.Create(req)
	require.NoError(t, err)

	// Get existing task
	task, err := storage.GetByID(createdTask.ID)
	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, createdTask.ID, task.ID)
	assert.Equal(t, createdTask.Name, task.Name)
	assert.Equal(t, createdTask.Status, task.Status)

	// Get non-existing task
	task, err = storage.GetByID("non-existing")
	assert.Error(t, err)
	assert.Nil(t, task)
	assert.Contains(t, err.Error(), "not found")
}

func TestMemoryStorage_Update(t *testing.T) {
	storage := NewMemoryStorage(1000)

	// Create a task first
	createReq := &models.CreateTaskRequest{Name: "Original Task", Status: models.TaskIncomplete}
	createdTask, err := storage.Create(createReq)
	require.NoError(t, err)

	tests := []struct {
		name    string
		taskID  string
		request *models.UpdateTaskRequest
		wantErr bool
		errMsg  string
	}{
		{
			name:   "update name only",
			taskID: createdTask.ID,
			request: &models.UpdateTaskRequest{
				Name: stringPtr("Updated Task"),
			},
			wantErr: false,
		},
		{
			name:   "update status only",
			taskID: createdTask.ID,
			request: &models.UpdateTaskRequest{
				Status: taskStatusPtr(models.TaskCompleted),
			},
			wantErr: false,
		},
		{
			name:   "update both name and status",
			taskID: createdTask.ID,
			request: &models.UpdateTaskRequest{
				Name:   stringPtr("Fully Updated Task"),
				Status: taskStatusPtr(models.TaskIncomplete),
			},
			wantErr: false,
		},
		{
			name:   "non-existing task should fail",
			taskID: "non-existing",
			request: &models.UpdateTaskRequest{
				Name: stringPtr("Updated Task"),
			},
			wantErr: true,
			errMsg:  "not found",
		},
		{
			name:    "empty update should fail",
			taskID:  createdTask.ID,
			request: &models.UpdateTaskRequest{},
			wantErr: true,
			errMsg:  "no updates provided",
		},
		{
			name:   "empty name should fail",
			taskID: createdTask.ID,
			request: &models.UpdateTaskRequest{
				Name: stringPtr(""),
			},
			wantErr: true,
			errMsg:  "validation failed",
		},
		{
			name:   "invalid status should fail",
			taskID: createdTask.ID,
			request: &models.UpdateTaskRequest{
				Status: taskStatusPtr(models.TaskStatus(99)),
			},
			wantErr: true,
			errMsg:  "validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updatedTask, err := storage.Update(tt.taskID, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, updatedTask)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, updatedTask)
				assert.Equal(t, tt.taskID, updatedTask.ID)

				if tt.request.Name != nil {
					assert.Equal(t, *tt.request.Name, updatedTask.Name)
				}
				if tt.request.Status != nil {
					assert.Equal(t, *tt.request.Status, updatedTask.Status)
				}

				// UpdatedAt should be changed
				assert.True(t, updatedTask.UpdatedAt.After(createdTask.CreatedAt) ||
					updatedTask.UpdatedAt.Equal(createdTask.CreatedAt))
			}
		})
	}
}

func TestMemoryStorage_Delete(t *testing.T) {
	storage := NewMemoryStorage(1000)

	// Create a task
	req := &models.CreateTaskRequest{Name: "Test Task", Status: models.TaskIncomplete}
	createdTask, err := storage.Create(req)
	require.NoError(t, err)

	// Verify task exists
	task, err := storage.GetByID(createdTask.ID)
	assert.NoError(t, err)
	assert.NotNil(t, task)

	// Delete the task
	err = storage.Delete(createdTask.ID)
	assert.NoError(t, err)

	// Verify task no longer exists
	task, err = storage.GetByID(createdTask.ID)
	assert.Error(t, err)
	assert.Nil(t, task)

	// Delete non-existing task should fail
	err = storage.Delete("non-existing")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestMemoryStorage_Count(t *testing.T) {
	storage := NewMemoryStorage(1000)

	// Initially zero
	count, err := storage.Count()
	assert.NoError(t, err)
	assert.Equal(t, 0, count)

	// Create some tasks
	req := &models.CreateTaskRequest{Name: "Test Task", Status: models.TaskIncomplete}

	for i := 0; i < 5; i++ {
		_, err := storage.Create(req)
		require.NoError(t, err)
	}

	// Count should be 5
	count, err = storage.Count()
	assert.NoError(t, err)
	assert.Equal(t, 5, count)
}

func TestMemoryStorage_Clear(t *testing.T) {
	storage := NewMemoryStorage(1000)

	// Create some tasks
	req := &models.CreateTaskRequest{Name: "Test Task", Status: models.TaskIncomplete}

	for i := 0; i < 3; i++ {
		_, err := storage.Create(req)
		require.NoError(t, err)
	}

	// Verify tasks exist
	count, err := storage.Count()
	assert.NoError(t, err)
	assert.Equal(t, 3, count)

	// Clear storage
	err = storage.Clear()
	assert.NoError(t, err)

	// Verify storage is empty
	count, err = storage.Count()
	assert.NoError(t, err)
	assert.Equal(t, 0, count)

	// Verify maxTasks is preserved
	assert.Equal(t, 1000, storage.GetMaxTasks())
}

func TestMemoryStorage_HealthCheck(t *testing.T) {
	storage := NewMemoryStorage(1000)

	// Normal storage should be healthy
	err := storage.HealthCheck()
	assert.NoError(t, err)

	// Nil shards should fail health check
	storage.shards = nil
	err = storage.HealthCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not properly initialized")
}

func TestMemoryStorage_GetStats(t *testing.T) {
	storage := NewMemoryStorage(1000)

	// Create tasks with different statuses
	incompleteReq := &models.CreateTaskRequest{Name: "Incomplete Task", Status: models.TaskIncomplete}
	completedReq := &models.CreateTaskRequest{Name: "Completed Task", Status: models.TaskCompleted}

	// Create 3 incomplete and 2 completed tasks
	for i := 0; i < 3; i++ {
		_, err := storage.Create(incompleteReq)
		require.NoError(t, err)
	}

	for i := 0; i < 2; i++ {
		_, err := storage.Create(completedReq)
		require.NoError(t, err)
	}

	stats := storage.GetStats()
	assert.Equal(t, 5, stats.TotalTasks)
	assert.Equal(t, 3, stats.IncompleteTasks)
	assert.Equal(t, 2, stats.CompletedTasks)
	assert.Equal(t, 0, stats.LastID)
	assert.Equal(t, "sharded_memory", stats.StorageType)
}

func TestMemoryStorage_GetTasksByStatus(t *testing.T) {
	storage := NewMemoryStorage(1000)

	// Create tasks with different statuses
	incompleteReq := &models.CreateTaskRequest{Name: "Incomplete Task", Status: models.TaskIncomplete}
	completedReq := &models.CreateTaskRequest{Name: "Completed Task", Status: models.TaskCompleted}

	// Create 2 incomplete and 1 completed task
	for i := 0; i < 2; i++ {
		_, err := storage.Create(incompleteReq)
		require.NoError(t, err)
	}

	_, err := storage.Create(completedReq)
	require.NoError(t, err)

	// Get incomplete tasks
	incompleteTasks, err := storage.GetTasksByStatus(models.TaskIncomplete)
	assert.NoError(t, err)
	assert.Len(t, incompleteTasks, 2)
	for _, task := range incompleteTasks {
		assert.Equal(t, models.TaskIncomplete, task.Status)
	}

	// Get completed tasks
	completedTasks, err := storage.GetTasksByStatus(models.TaskCompleted)
	assert.NoError(t, err)
	assert.Len(t, completedTasks, 1)
	for _, task := range completedTasks {
		assert.Equal(t, models.TaskCompleted, task.Status)
	}
}

func TestMemoryStorage_GetTasksPaginated(t *testing.T) {
	storage := NewMemoryStorage(1000)

	// Create 10 tasks
	req := &models.CreateTaskRequest{Name: "Test Task", Status: models.TaskIncomplete}
	for i := 0; i < 10; i++ {
		_, err := storage.Create(req)
		require.NoError(t, err)
	}

	// Test pagination
	tasks, total, err := storage.GetTasksPaginated(0, 5)
	assert.NoError(t, err)
	assert.Len(t, tasks, 5)
	assert.Equal(t, 10, total)

	// Test second page
	tasks, total, err = storage.GetTasksPaginated(5, 5)
	assert.NoError(t, err)
	assert.Len(t, tasks, 5)
	assert.Equal(t, 10, total)

	// Test beyond available data
	tasks, total, err = storage.GetTasksPaginated(10, 5)
	assert.NoError(t, err)
	assert.Len(t, tasks, 0)
	assert.Equal(t, 10, total)

	// Test partial page
	tasks, total, err = storage.GetTasksPaginated(8, 5)
	assert.NoError(t, err)
	assert.Len(t, tasks, 2)
	assert.Equal(t, 10, total)
}

// Test thread safety with concurrent operations
func TestMemoryStorage_ConcurrentOperations(t *testing.T) {
	storage := NewMemoryStorage(1000)
	const numGoroutines = 10
	const operationsPerGoroutine = 100

	var wg sync.WaitGroup

	// Concurrent creates
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				req := &models.CreateTaskRequest{
					Name:   "Task " + strconv.Itoa(id) + "-" + strconv.Itoa(j),
					Status: models.TaskIncomplete,
				}
				_, err := storage.Create(req)
				assert.NoError(t, err)
			}
		}(i)
	}
	wg.Wait()

	// Verify all tasks were created
	count, err := storage.Count()
	assert.NoError(t, err)
	assert.Equal(t, numGoroutines*operationsPerGoroutine, count)

	// Concurrent reads
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				tasks, err := storage.GetAll()
				assert.NoError(t, err)
				assert.NotNil(t, tasks)
			}
		}()
	}
	wg.Wait()
}

// Helper functions for tests
func stringPtr(s string) *string {
	return &s
}

func taskStatusPtr(status models.TaskStatus) *models.TaskStatus {
	return &status
}

// Benchmark tests
func BenchmarkMemoryStorage_Create(b *testing.B) {
	storage := NewMemoryStorage(1000)
	req := &models.CreateTaskRequest{Name: "Benchmark Task", Status: models.TaskIncomplete}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := storage.Create(req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMemoryStorage_GetAll(b *testing.B) {
	storage := NewMemoryStorage(1000)
	req := &models.CreateTaskRequest{Name: "Benchmark Task", Status: models.TaskIncomplete}

	// Pre-populate with 1000 tasks
	for i := 0; i < 1000; i++ {
		_, err := storage.Create(req)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := storage.GetAll()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// TestMemoryStorage_MaxTasksLimit tests the maximum tasks limit functionality
func TestMemoryStorage_MaxTasksLimit(t *testing.T) {
	// Create storage with limit of 3 tasks
	storage := NewMemoryStorage(3)
	req := &models.CreateTaskRequest{Name: "Test Task", Status: models.TaskIncomplete}

	// Create 3 tasks - should succeed
	for i := 0; i < 3; i++ {
		task, err := storage.Create(req)
		assert.NoError(t, err)
		assert.NotNil(t, task)
	}

	// Fourth task should fail
	task, err := storage.Create(req)
	assert.Error(t, err)
	assert.Nil(t, task)
	assert.Contains(t, err.Error(), "maximum tasks limit reached")
	assert.Contains(t, err.Error(), "(3)")

	// Verify count is still 3
	count, err := storage.Count()
	assert.NoError(t, err)
	assert.Equal(t, 3, count)

	// Delete one task and try again - should succeed
	tasks, err := storage.GetAll()
	assert.NoError(t, err)
	assert.Len(t, tasks, 3)

	err = storage.Delete(tasks[0].ID)
	assert.NoError(t, err)

	// Now creating should work again
	task, err = storage.Create(req)
	assert.NoError(t, err)
	assert.NotNil(t, task)
}

// TestMemoryStorage_UUIDGeneration tests UUID generation and uniqueness
func TestMemoryStorage_UUIDGeneration(t *testing.T) {
	storage := NewMemoryStorage(1000)
	req := &models.CreateTaskRequest{Name: "Test Task", Status: models.TaskIncomplete}

	// Create multiple tasks and verify all IDs are unique UUIDs
	createdIDs := make(map[string]bool)

	for i := 0; i < 100; i++ {
		task, err := storage.Create(req)
		assert.NoError(t, err)
		assert.NotNil(t, task)

		// Check UUID format (36 characters with hyphens)
		assert.Len(t, task.ID, 36)
		assert.Contains(t, task.ID, "-")

		// Check uniqueness
		assert.False(t, createdIDs[task.ID], "Duplicate UUID found: %s", task.ID)
		createdIDs[task.ID] = true
	}

	// Verify all IDs are different from the old numeric format
	for id := range createdIDs {
		assert.NotRegexp(t, `^task_\d+$`, id, "ID should not be in old numeric format")
	}
}

// TestMemoryStorage_LimitReached tests behavior when limit is reached
func TestMemoryStorage_LimitReached(t *testing.T) {
	storage := NewMemoryStorage(2)
	req := &models.CreateTaskRequest{Name: "Test Task", Status: models.TaskIncomplete}

	// Fill to limit
	task1, err := storage.Create(req)
	assert.NoError(t, err)
	assert.NotNil(t, task1)

	task2, err := storage.Create(req)
	assert.NoError(t, err)
	assert.NotNil(t, task2)

	// Should fail now
	task3, err := storage.Create(req)
	assert.Error(t, err)
	assert.Nil(t, task3)

	// Test GetMaxTasks and GetUsage methods
	assert.Equal(t, 2, storage.GetMaxTasks())

	usage := storage.GetUsage()
	assert.Equal(t, 2, usage["current_tasks"])
	assert.Equal(t, 2, usage["max_tasks"])
	assert.Equal(t, float64(100), usage["usage_percent"])
	assert.Equal(t, 0, usage["available"])
	assert.Contains(t, usage, "shard_count")
	assert.Contains(t, usage, "shard_distribution")
	assert.Equal(t, "sharded_memory", usage["storage_type"])

	// Delete one task and check usage again
	err = storage.Delete(task1.ID)
	assert.NoError(t, err)

	usage = storage.GetUsage()
	assert.Equal(t, 1, usage["current_tasks"])
	assert.Equal(t, 2, usage["max_tasks"])
	assert.Equal(t, float64(50), usage["usage_percent"])
	assert.Equal(t, 1, usage["available"])
	assert.Contains(t, usage, "shard_count")
	assert.Contains(t, usage, "shard_distribution")
}

// TestMemoryStorage_GetMaxTasks tests the GetMaxTasks method
func TestMemoryStorage_GetMaxTasks(t *testing.T) {
	storage1 := NewMemoryStorage(100)
	assert.Equal(t, 100, storage1.GetMaxTasks())

	storage2 := NewMemoryStorage(5000)
	assert.Equal(t, 5000, storage2.GetMaxTasks())

	storage3 := NewMemoryStorage(0) // Should use default
	assert.Equal(t, 10000, storage3.GetMaxTasks())
}

// TestMemoryStorage_GetUsage tests the GetUsage method
func TestMemoryStorage_GetUsage(t *testing.T) {
	storage := NewMemoryStorage(10)
	req := &models.CreateTaskRequest{Name: "Test Task", Status: models.TaskIncomplete}

	// Empty storage
	usage := storage.GetUsage()
	assert.Equal(t, 0, usage["current_tasks"])
	assert.Equal(t, 10, usage["max_tasks"])
	assert.Equal(t, float64(0), usage["usage_percent"])
	assert.Equal(t, 10, usage["available"])
	assert.Contains(t, usage, "shard_count")
	assert.Contains(t, usage, "shard_distribution")
	assert.Equal(t, "sharded_memory", usage["storage_type"])

	// Add 3 tasks
	for i := 0; i < 3; i++ {
		_, err := storage.Create(req)
		assert.NoError(t, err)
	}

	usage = storage.GetUsage()
	assert.Equal(t, 3, usage["current_tasks"])
	assert.Equal(t, 10, usage["max_tasks"])
	assert.Equal(t, float64(30), usage["usage_percent"])
	assert.Equal(t, 7, usage["available"])
	assert.Contains(t, usage, "shard_count")
	assert.Contains(t, usage, "shard_distribution")
	assert.Equal(t, "sharded_memory", usage["storage_type"])
}

func BenchmarkMemoryStorage_GetByID(b *testing.B) {
	storage := NewMemoryStorage(1000)
	req := &models.CreateTaskRequest{Name: "Benchmark Task", Status: models.TaskIncomplete}

	task, err := storage.Create(req)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := storage.GetByID(task.ID)
		if err != nil {
			b.Fatal(err)
		}
	}
}
