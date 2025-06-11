package storage

import (
	"fmt"
	"sync"
	"sync/atomic"
	"task-api/internal/interfaces"
	"task-api/internal/models"
	"time"

	"github.com/google/uuid"
)

// shard represents a single shard with its own lock and task storage
type shard struct {
	tasks map[string]*models.Task // Task storage for this shard
	mutex sync.RWMutex            // Per-shard read-write lock
}

// MemoryStorage implements TaskStorage interface using sharded in-memory storage
// This implementation is thread-safe using sharding to reduce lock contention
type MemoryStorage struct {
	shards     []*shard  // Array of shards
	shardCount uint32    // Number of shards (using uint32 to match hash algorithm)
	maxTasks   int       // Maximum number of tasks allowed
	taskCount  int64     // Atomic task counter for fast count operations
	taskPool   sync.Pool // Object pool to reduce GC pressure
}

// Ensure MemoryStorage implements required interfaces at compile time
var (
	_ interfaces.TaskStorage   = (*MemoryStorage)(nil)
	_ interfaces.HealthChecker = (*MemoryStorage)(nil)
)

// NewMemoryStorage creates a new instance of MemoryStorage with sharding optimization
func NewMemoryStorage(maxTasks int) *MemoryStorage {
	if maxTasks <= 0 {
		maxTasks = 10000 // Default value
	}

	// Calculate optimal shard count based on maxTasks
	// More shards = less lock contention, but more memory overhead
	shardCount := 32 // Default for most use cases
	if maxTasks < 1000 {
		shardCount = 8
	} else if maxTasks > 100000 {
		shardCount = 64
	}

	// Initialize shards
	shards := make([]*shard, shardCount)
	for i := range shards {
		shards[i] = &shard{
			tasks: make(map[string]*models.Task),
			mutex: sync.RWMutex{},
		}
	}

	// Safe conversion with bounds checking to prevent integer overflow
	var safeShardCount uint32
	if shardCount < 0 || shardCount > int(^uint32(0)>>1) {
		// Use default safe value if out of bounds
		safeShardCount = 32
	} else {
		// #nosec G115 - Safe conversion with bounds checking above
		safeShardCount = uint32(shardCount)
	}

	return &MemoryStorage{
		shards:     shards,
		shardCount: safeShardCount,
		maxTasks:   maxTasks,
		taskCount:  0,
		taskPool: sync.Pool{
			New: func() interface{} {
				return &models.Task{}
			},
		},
	}
}

// getShard returns the shard for a given key using FNV-1a hash algorithm
func (ms *MemoryStorage) getShard(key string) *shard {
	hash := ms.fnv32Hash(key)
	shardIndex := hash % ms.shardCount
	return ms.shards[shardIndex]
}

// fnv32Hash implements FNV-1a 32-bit hash algorithm for fast key distribution
func (ms *MemoryStorage) fnv32Hash(key string) uint32 {
	hash := uint32(2166136261)     // FNV offset basis
	const prime = uint32(16777619) // FNV prime

	for i := 0; i < len(key); i++ {
		hash ^= uint32(key[i])
		hash *= prime
	}
	return hash
}

// GetAll retrieves all tasks from all shards
// Returns a copy of all tasks to prevent external modifications
func (ms *MemoryStorage) GetAll() ([]*models.Task, error) {
	// Pre-allocate slice with current task count for better performance
	currentCount := atomic.LoadInt64(&ms.taskCount)
	allTasks := make([]*models.Task, 0, currentCount)

	// Iterate through all shards and collect tasks
	for _, shard := range ms.shards {
		shard.mutex.RLock()
		for _, task := range shard.tasks {
			// Create a copy to prevent external modifications
			taskCopy := *task
			allTasks = append(allTasks, &taskCopy)
		}
		shard.mutex.RUnlock()
	}

	return allTasks, nil
}

// GetByID retrieves a specific task by its ID from the appropriate shard
func (ms *MemoryStorage) GetByID(id string) (*models.Task, error) {
	shard := ms.getShard(id)
	shard.mutex.RLock()
	defer shard.mutex.RUnlock()

	task, exists := shard.tasks[id]
	if !exists {
		return nil, fmt.Errorf("task with ID %s not found", id)
	}

	// Return a copy to prevent external modifications
	taskCopy := *task
	return &taskCopy, nil
}

// Create creates a new task in the appropriate shard
func (ms *MemoryStorage) Create(req *models.CreateTaskRequest) (*models.Task, error) {
	// Validate the request first
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if maximum tasks limit is reached using atomic operation
	currentCount := atomic.LoadInt64(&ms.taskCount)
	if int(currentCount) >= ms.maxTasks {
		return nil, fmt.Errorf("maximum tasks limit reached (%d)", ms.maxTasks)
	}

	// Generate UUID as task ID
	// UUID v4 collision probability is extremely low (~10^-15), so no need to check uniqueness
	taskID := uuid.New().String()

	// Create new task using factory method
	task := models.NewTask(req.Name, req.Status)
	task.ID = taskID

	// Get the appropriate shard and store the task
	shard := ms.getShard(taskID)
	shard.mutex.Lock()
	shard.tasks[taskID] = task
	shard.mutex.Unlock()

	// Increment task count atomically
	atomic.AddInt64(&ms.taskCount, 1)

	// Return a copy
	taskCopy := *task
	return &taskCopy, nil
}

// Update updates an existing task in the appropriate shard
func (ms *MemoryStorage) Update(id string, req *models.UpdateTaskRequest) (*models.Task, error) {
	// Validate the request first
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if there are any updates to apply
	if !req.HasUpdates() {
		return nil, fmt.Errorf("no updates provided")
	}

	shard := ms.getShard(id)
	shard.mutex.Lock()
	defer shard.mutex.Unlock()

	// Check if task exists
	task, exists := shard.tasks[id]
	if !exists {
		return nil, fmt.Errorf("task with ID %s not found", id)
	}

	// Create a copy of the existing task to modify
	updatedTask := *task

	// Apply updates to the copy
	req.ApplyTo(&updatedTask)

	// Store the updated task
	shard.tasks[id] = &updatedTask

	// Return a copy
	taskCopy := updatedTask
	return &taskCopy, nil
}

// Delete removes a task from the appropriate shard
func (ms *MemoryStorage) Delete(id string) error {
	shard := ms.getShard(id)
	shard.mutex.Lock()
	defer shard.mutex.Unlock()

	// Check if task exists
	if _, exists := shard.tasks[id]; !exists {
		return fmt.Errorf("task with ID %s not found", id)
	}

	// Delete the task
	delete(shard.tasks, id)

	// Decrement task count atomically
	atomic.AddInt64(&ms.taskCount, -1)

	return nil
}

// Count returns the total number of tasks using atomic operation for O(1) performance
func (ms *MemoryStorage) Count() (int, error) {
	count := atomic.LoadInt64(&ms.taskCount)
	return int(count), nil
}

// Clear removes all tasks from all shards (primarily for testing)
func (ms *MemoryStorage) Clear() error {
	// Clear all shards
	for _, shard := range ms.shards {
		shard.mutex.Lock()
		shard.tasks = make(map[string]*models.Task)
		shard.mutex.Unlock()
	}

	// Reset task count
	atomic.StoreInt64(&ms.taskCount, 0)

	return nil
}

// HealthCheck verifies if the storage is accessible and functioning
func (ms *MemoryStorage) HealthCheck() error {
	// Check if shards are properly initialized
	if len(ms.shards) == 0 {
		return fmt.Errorf("memory storage shards are not properly initialized")
	}

	// Check each shard
	for i, shard := range ms.shards {
		if shard == nil || shard.tasks == nil {
			return fmt.Errorf("memory storage shard %d is not properly initialized", i)
		}
	}

	return nil
}

// GetStats returns statistics about the memory storage
func (ms *MemoryStorage) GetStats() StorageStats {
	completedCount := 0
	incompleteCount := 0

	// Collect stats from all shards
	for _, shard := range ms.shards {
		shard.mutex.RLock()
		for _, task := range shard.tasks {
			if task.Status == models.TaskCompleted {
				completedCount++
			} else {
				incompleteCount++
			}
		}
		shard.mutex.RUnlock()
	}

	currentCount := atomic.LoadInt64(&ms.taskCount)

	return StorageStats{
		TotalTasks:      int(currentCount),
		CompletedTasks:  completedCount,
		IncompleteTasks: incompleteCount,
		LastID:          0, // UUID doesn't use numeric IDs, set to 0
		StorageType:     "sharded_memory",
	}
}

// GetMaxTasks returns the maximum number of tasks allowed
func (ms *MemoryStorage) GetMaxTasks() int {
	return ms.maxTasks
}

// GetUsage returns current storage usage information including shard statistics
func (ms *MemoryStorage) GetUsage() map[string]interface{} {
	currentCount := atomic.LoadInt64(&ms.taskCount)

	// Calculate per-shard distribution
	shardDistribution := make([]int, int(ms.shardCount))
	for i, shard := range ms.shards {
		shard.mutex.RLock()
		shardDistribution[i] = len(shard.tasks)
		shard.mutex.RUnlock()
	}

	return map[string]interface{}{
		"current_tasks":      int(currentCount),
		"max_tasks":          ms.maxTasks,
		"usage_percent":      float64(currentCount) / float64(ms.maxTasks) * 100,
		"available":          ms.maxTasks - int(currentCount),
		"shard_count":        int(ms.shardCount),
		"shard_distribution": shardDistribution,
		"storage_type":       "sharded_memory",
	}
}

// GetTasksByStatus returns all tasks with the specified status from all shards
func (ms *MemoryStorage) GetTasksByStatus(status models.TaskStatus) ([]*models.Task, error) {
	var tasks []*models.Task

	// Collect tasks from all shards
	for _, shard := range ms.shards {
		shard.mutex.RLock()
		for _, task := range shard.tasks {
			if task.Status == status {
				taskCopy := *task
				tasks = append(tasks, &taskCopy)
			}
		}
		shard.mutex.RUnlock()
	}

	return tasks, nil
}

// GetTasksCreatedAfter returns tasks created after the specified time from all shards
func (ms *MemoryStorage) GetTasksCreatedAfter(after time.Time) ([]*models.Task, error) {
	var tasks []*models.Task

	// Collect tasks from all shards
	for _, shard := range ms.shards {
		shard.mutex.RLock()
		for _, task := range shard.tasks {
			if task.CreatedAt.After(after) {
				taskCopy := *task
				tasks = append(tasks, &taskCopy)
			}
		}
		shard.mutex.RUnlock()
	}

	return tasks, nil
}

// GetTasksPaginated returns a paginated list of tasks from all shards
func (ms *MemoryStorage) GetTasksPaginated(offset, limit int) ([]*models.Task, int, error) {
	// Get all tasks first (could be optimized further with shard-level pagination)
	allTasks := make([]*models.Task, 0, atomic.LoadInt64(&ms.taskCount))

	for _, shard := range ms.shards {
		shard.mutex.RLock()
		for _, task := range shard.tasks {
			taskCopy := *task
			allTasks = append(allTasks, &taskCopy)
		}
		shard.mutex.RUnlock()
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
	StorageType     string `json:"storage_type"`     // Type of storage (sharded_memory, database, etc.)
}
