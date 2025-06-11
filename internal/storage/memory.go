package storage

import (
	"fmt"
	"sync"
	"task-api/internal/interfaces"
	"task-api/internal/models"
	"time"

	"github.com/google/uuid"
)

// MemoryStorage implements TaskStorage interface using in-memory storage
// This implementation is thread-safe using sync.RWMutex
type MemoryStorage struct {
	tasks    map[string]*models.Task // In-memory task storage
	mutex    sync.RWMutex            // Read-write mutex for thread safety
	maxTasks int                     // Maximum number of tasks allowed
}

// Ensure MemoryStorage implements required interfaces at compile time
var (
	_ interfaces.TaskStorage   = (*MemoryStorage)(nil)
	_ interfaces.HealthChecker = (*MemoryStorage)(nil)
)

// NewMemoryStorage creates a new instance of MemoryStorage (Factory Pattern)
func NewMemoryStorage(maxTasks int) *MemoryStorage {
	if maxTasks <= 0 {
		maxTasks = 10000 // 默認值
	}
	return &MemoryStorage{
		tasks:    make(map[string]*models.Task),
		mutex:    sync.RWMutex{},
		maxTasks: maxTasks,
	}
}

// GetAll retrieves all tasks from memory storage
// Returns a copy of all tasks to prevent external modifications
func (ms *MemoryStorage) GetAll() ([]*models.Task, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	// Create a slice to hold all tasks
	tasks := make([]*models.Task, 0, len(ms.tasks))

	// Copy all tasks to prevent external modifications
	for _, task := range ms.tasks {
		taskCopy := *task // Create a copy of the task
		tasks = append(tasks, &taskCopy)
	}

	return tasks, nil
}

// GetByID retrieves a specific task by its ID
func (ms *MemoryStorage) GetByID(id string) (*models.Task, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	task, exists := ms.tasks[id]
	if !exists {
		return nil, fmt.Errorf("task with ID %s not found", id)
	}

	// Return a copy to prevent external modifications
	taskCopy := *task
	return &taskCopy, nil
}

// Create creates a new task in memory storage
func (ms *MemoryStorage) Create(req *models.CreateTaskRequest) (*models.Task, error) {
	// Validate the request first
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	// Check if maximum tasks limit is reached
	if len(ms.tasks) >= ms.maxTasks {
		return nil, fmt.Errorf("maximum tasks limit reached (%d)", ms.maxTasks)
	}

	// Generate UUID as task ID
	taskID := uuid.New().String()
	// UUID 碰撞極少，但仍檢查唯一性
	for ms.tasks[taskID] != nil {
		taskID = uuid.New().String()
	}

	// Create new task using factory method
	task := models.NewTask(req.Name, req.Status)
	task.ID = taskID

	// Store the task
	ms.tasks[taskID] = task

	// Return a copy
	taskCopy := *task
	return &taskCopy, nil
}

// Update updates an existing task in memory storage
func (ms *MemoryStorage) Update(id string, req *models.UpdateTaskRequest) (*models.Task, error) {
	// Validate the request first
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if there are any updates to apply
	if !req.HasUpdates() {
		return nil, fmt.Errorf("no updates provided")
	}

	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	// Check if task exists
	task, exists := ms.tasks[id]
	if !exists {
		return nil, fmt.Errorf("task with ID %s not found", id)
	}

	// Create a copy of the existing task to modify
	updatedTask := *task

	// Apply updates to the copy
	req.ApplyTo(&updatedTask)

	// Store the updated task
	ms.tasks[id] = &updatedTask

	// Return a copy
	taskCopy := updatedTask
	return &taskCopy, nil
}

// Delete removes a task from memory storage
func (ms *MemoryStorage) Delete(id string) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	// Check if task exists
	if _, exists := ms.tasks[id]; !exists {
		return fmt.Errorf("task with ID %s not found", id)
	}

	// Delete the task
	delete(ms.tasks, id)

	return nil
}

// Count returns the total number of tasks in storage
func (ms *MemoryStorage) Count() (int, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	return len(ms.tasks), nil
}

// Clear removes all tasks from storage (primarily for testing)
func (ms *MemoryStorage) Clear() error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	// Clear the map
	ms.tasks = make(map[string]*models.Task)

	return nil
}

// HealthCheck verifies if the storage is accessible and functioning
func (ms *MemoryStorage) HealthCheck() error {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	// For memory storage, we can check if the map is initialized
	if ms.tasks == nil {
		return fmt.Errorf("memory storage is not properly initialized")
	}

	return nil
}

// GetStats returns statistics about the memory storage
func (ms *MemoryStorage) GetStats() StorageStats {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	completedCount := 0
	incompleteCount := 0

	for _, task := range ms.tasks {
		if task.Status == models.TaskCompleted {
			completedCount++
		} else {
			incompleteCount++
		}
	}

	return StorageStats{
		TotalTasks:      len(ms.tasks),
		CompletedTasks:  completedCount,
		IncompleteTasks: incompleteCount,
		LastID:          0, // UUID 不使用數字 ID，設為 0
		StorageType:     "memory",
	}
}

// GetMaxTasks returns the maximum number of tasks allowed
func (ms *MemoryStorage) GetMaxTasks() int {
	return ms.maxTasks
}

// GetUsage returns current storage usage information
func (ms *MemoryStorage) GetUsage() map[string]interface{} {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	return map[string]interface{}{
		"current_tasks": len(ms.tasks),
		"max_tasks":     ms.maxTasks,
		"usage_percent": float64(len(ms.tasks)) / float64(ms.maxTasks) * 100,
		"available":     ms.maxTasks - len(ms.tasks),
	}
}

// GetTasksByStatus returns all tasks with the specified status
func (ms *MemoryStorage) GetTasksByStatus(status models.TaskStatus) ([]*models.Task, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	var tasks []*models.Task

	for _, task := range ms.tasks {
		if task.Status == status {
			taskCopy := *task
			tasks = append(tasks, &taskCopy)
		}
	}

	return tasks, nil
}

// GetTasksCreatedAfter returns tasks created after the specified time
func (ms *MemoryStorage) GetTasksCreatedAfter(after time.Time) ([]*models.Task, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	var tasks []*models.Task

	for _, task := range ms.tasks {
		if task.CreatedAt.After(after) {
			taskCopy := *task
			tasks = append(tasks, &taskCopy)
		}
	}

	return tasks, nil
}

// GetTasksPaginated returns a paginated list of tasks
func (ms *MemoryStorage) GetTasksPaginated(offset, limit int) ([]*models.Task, int, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	// Convert map to slice for pagination
	allTasks := make([]*models.Task, 0, len(ms.tasks))
	for _, task := range ms.tasks {
		taskCopy := *task
		allTasks = append(allTasks, &taskCopy)
	}

	total := len(allTasks)

	// Handle edge cases
	if offset >= total {
		return []*models.Task{}, total, nil
	}

	end := offset + limit
	if end > total {
		end = total
	}

	paginatedTasks := allTasks[offset:end]

	return paginatedTasks, total, nil
}

// StorageStats represents statistics about the storage
type StorageStats struct {
	TotalTasks      int    `json:"total_tasks"`      // Total number of tasks
	CompletedTasks  int    `json:"completed_tasks"`  // Number of completed tasks
	IncompleteTasks int    `json:"incomplete_tasks"` // Number of incomplete tasks
	LastID          int    `json:"last_id"`          // Last generated ID
	StorageType     string `json:"storage_type"`     // Type of storage (memory, database, etc.)
}
