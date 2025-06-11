package models

import (
	"fmt"
	"time"
)

// TaskStatus defines the enumeration values for task status
type TaskStatus int

const (
	// TaskIncomplete represents an incomplete task
	TaskIncomplete TaskStatus = 0
	// TaskCompleted represents a completed task
	TaskCompleted TaskStatus = 1
)

// String implements the Stringer interface, providing string representation of status
func (ts TaskStatus) String() string {
	switch ts {
	case TaskIncomplete:
		return "incomplete"
	case TaskCompleted:
		return "completed"
	default:
		return "unknown"
	}
}

// IsValid checks if the status value is valid
func (ts TaskStatus) IsValid() bool {
	return ts == TaskIncomplete || ts == TaskCompleted
}

// Task represents a task entity
type Task struct {
	ID        string     `json:"id"`                      // Unique identifier
	Name      string     `json:"name" binding:"required"` // Task name (required)
	Status    TaskStatus `json:"status"`                  // Task status
	CreatedAt time.Time  `json:"created_at"`              // Creation time
	UpdatedAt time.Time  `json:"updated_at"`              // Last update time
}

// CreateTaskRequest represents the DTO for creating a task
type CreateTaskRequest struct {
	Name   string     `json:"name" binding:"required"` // Task name (required)
	Status TaskStatus `json:"status"`                  // Task status (optional, defaults to incomplete)
}

// Validate validates the create request
func (req *CreateTaskRequest) Validate() error {
	if req.Name == "" {
		return fmt.Errorf("task name cannot be empty")
	}
	if len(req.Name) > 255 {
		return fmt.Errorf("task name cannot exceed 255 characters")
	}
	if !req.Status.IsValid() {
		return fmt.Errorf("invalid task status: %d", req.Status)
	}
	return nil
}

// UpdateTaskRequest represents the DTO for updating a task
type UpdateTaskRequest struct {
	Name   *string     `json:"name,omitempty"`   // Task name (optional)
	Status *TaskStatus `json:"status,omitempty"` // Task status (optional)
}

// Validate validates the update request
func (req *UpdateTaskRequest) Validate() error {
	if req.Name != nil {
		if *req.Name == "" {
			return fmt.Errorf("task name cannot be empty")
		}
		if len(*req.Name) > 255 {
			return fmt.Errorf("task name cannot exceed 255 characters")
		}
	}
	if req.Status != nil && !req.Status.IsValid() {
		return fmt.Errorf("invalid task status: %d", *req.Status)
	}
	return nil
}

// HasUpdates checks if there are any fields to update
func (req *UpdateTaskRequest) HasUpdates() bool {
	return req.Name != nil || req.Status != nil
}

// ApplyTo applies the update request to an existing task
func (req *UpdateTaskRequest) ApplyTo(task *Task) {
	now := time.Now()

	if req.Name != nil {
		task.Name = *req.Name
		task.UpdatedAt = now
	}

	if req.Status != nil {
		task.Status = *req.Status
		task.UpdatedAt = now
	}
}

// TaskResponse represents the DTO for single task response
type TaskResponse struct {
	Success bool   `json:"success"`           // Whether the operation was successful
	Message string `json:"message,omitempty"` // Response message
	Data    *Task  `json:"data,omitempty"`    // Task data
}

// TaskListResponse represents the DTO for task list response
type TaskListResponse struct {
	Success bool    `json:"success"`        // Whether the operation was successful
	Data    []*Task `json:"data,omitempty"` // Task list
	Count   int     `json:"count"`          // Total number of tasks
}

// ErrorResponse represents the DTO for error response
type ErrorResponse struct {
	Success bool   `json:"success"`         // Always false
	Message string `json:"message"`         // Error message
	Error   string `json:"error,omitempty"` // Detailed error information
}

// HealthResponse represents the DTO for health check response
type HealthResponse struct {
	Status    string    `json:"status"`    // Service status
	Timestamp time.Time `json:"timestamp"` // Check timestamp
	Version   string    `json:"version"`   // Service version
}

// NewTask creates a new task entity (Factory Pattern)
func NewTask(name string, status TaskStatus) *Task {
	now := time.Now()
	return &Task{
		Name:      name,
		Status:    status,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// NewTaskResponse creates a successful task response (Factory Pattern)
func NewTaskResponse(task *Task, message string) *TaskResponse {
	return &TaskResponse{
		Success: true,
		Message: message,
		Data:    task,
	}
}

// NewTaskListResponse creates a task list response (Factory Pattern)
func NewTaskListResponse(tasks []*Task) *TaskListResponse {
	return &TaskListResponse{
		Success: true,
		Data:    tasks,
		Count:   len(tasks),
	}
}

// NewErrorResponse creates an error response (Factory Pattern)
func NewErrorResponse(message string, err error) *ErrorResponse {
	response := &ErrorResponse{
		Success: false,
		Message: message,
	}

	if err != nil {
		response.Error = err.Error()
	}

	return response
}

// NewHealthResponse creates a health check response (Factory Pattern)
func NewHealthResponse(version string) *HealthResponse {
	return &HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   version,
	}
}
