package interfaces

import "task-api/internal/models"

// TaskStorage defines the interface for task storage operations
// This interface implements the Repository Pattern, allowing for different storage implementations
type TaskStorage interface {
	// GetAll retrieves all tasks from storage
	// Returns a slice of all tasks and any error that occurred
	GetAll() ([]*models.Task, error)

	// GetByID retrieves a specific task by its ID
	// Returns the task if found, nil if not found, and any error that occurred
	GetByID(id string) (*models.Task, error)

	// Create creates a new task in storage
	// Takes a CreateTaskRequest and returns the created task with generated ID and timestamps
	Create(req *models.CreateTaskRequest) (*models.Task, error)

	// Update updates an existing task in storage
	// Takes the task ID and UpdateTaskRequest, returns the updated task
	// Returns error if task not found or update fails
	Update(id string, req *models.UpdateTaskRequest) (*models.Task, error)

	// Delete removes a task from storage by its ID
	// Returns error if task not found or deletion fails
	Delete(id string) error

	// Count returns the total number of tasks in storage
	// Useful for pagination and statistics
	Count() (int, error)

	// Clear removes all tasks from storage
	// Primarily used for testing purposes
	Clear() error
}

// HealthChecker defines the interface for health checking storage connections
// This is useful for monitoring and ensuring storage availability
type HealthChecker interface {
	// HealthCheck verifies if the storage is accessible and functioning
	// Returns error if storage is not healthy
	HealthCheck() error
}
