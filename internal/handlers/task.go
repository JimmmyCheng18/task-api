package handlers

import (
	"net/http"
	"strconv"
	"task-api/internal/interfaces"
	"task-api/internal/models"
	"task-api/internal/storage"

	"github.com/gin-gonic/gin"
)

// TaskHandler handles HTTP requests for task operations
// This implements the MVC pattern's Controller layer
type TaskHandler struct {
	storage interfaces.TaskStorage // Dependency injection via interface
}

// NewTaskHandler creates a new TaskHandler instance (Factory Pattern)
func NewTaskHandler(storage interfaces.TaskStorage) *TaskHandler {
	return &TaskHandler{
		storage: storage,
	}
}

// GetAllTasks handles GET /tasks - retrieve all tasks
// @Summary Get all tasks
// @Description Get all tasks from the storage
// @Tags tasks
// @Accept json
// @Produce json
// @Success 200 {object} models.TaskListResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /tasks [get]
func (h *TaskHandler) GetAllTasks(c *gin.Context) {
	tasks, err := h.storage.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Failed to retrieve tasks",
			err,
		))
		return
	}

	response := models.NewTaskListResponse(tasks)
	c.JSON(http.StatusOK, response)
}

// GetTaskByID handles GET /tasks/:id - retrieve a specific task
// @Summary Get a task by ID
// @Description Get a specific task by its ID
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Success 200 {object} models.TaskResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /tasks/{id} [get]
func (h *TaskHandler) GetTaskByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Task ID is required",
			nil,
		))
		return
	}

	task, err := h.storage.GetByID(id)
	if err != nil {
		// Check if it's a "not found" error
		if contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				"Task not found",
				err,
			))
			return
		}

		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Failed to retrieve task",
			err,
		))
		return
	}

	response := models.NewTaskResponse(task, "Task retrieved successfully")
	c.JSON(http.StatusOK, response)
}

// CreateTask handles POST /tasks - create a new task
// @Summary Create a new task
// @Description Create a new task with the provided data
// @Tags tasks
// @Accept json
// @Produce json
// @Param task body models.CreateTaskRequest true "Task data"
// @Success 201 {object} models.TaskResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /tasks [post]
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req models.CreateTaskRequest

	// Bind JSON request to struct with validation
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Invalid request data",
			err,
		))
		return
	}

	// Additional validation (business logic)
	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Validation failed",
			err,
		))
		return
	}

	// Create the task
	task, err := h.storage.Create(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Failed to create task",
			err,
		))
		return
	}

	response := models.NewTaskResponse(task, "Task created successfully")
	c.JSON(http.StatusCreated, response)
}

// UpdateTask handles PUT /tasks/:id - update an existing task
// @Summary Update a task
// @Description Update an existing task with the provided data
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Param task body models.UpdateTaskRequest true "Task update data"
// @Success 200 {object} models.TaskResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /tasks/{id} [put]
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Task ID is required",
			nil,
		))
		return
	}

	var req models.UpdateTaskRequest

	// Bind JSON request to struct
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Invalid request data",
			err,
		))
		return
	}

	// Additional validation (business logic)
	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Validation failed",
			err,
		))
		return
	}

	// Check if there are any updates
	if !req.HasUpdates() {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"No updates provided",
			nil,
		))
		return
	}

	// Update the task
	task, err := h.storage.Update(id, &req)
	if err != nil {
		// Check if it's a "not found" error
		if contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				"Task not found",
				err,
			))
			return
		}

		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Failed to update task",
			err,
		))
		return
	}

	response := models.NewTaskResponse(task, "Task updated successfully")
	c.JSON(http.StatusOK, response)
}

// DeleteTask handles DELETE /tasks/:id - delete a task
// @Summary Delete a task
// @Description Delete a task by its ID
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Success 200 {object} models.TaskResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /tasks/{id} [delete]
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Task ID is required",
			nil,
		))
		return
	}

	// Check if task exists before deletion
	_, err := h.storage.GetByID(id)
	if err != nil {
		// Check if it's a "not found" error
		if contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				"Task not found",
				err,
			))
			return
		}

		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Failed to retrieve task",
			err,
		))
		return
	}

	// Delete the task
	err = h.storage.Delete(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Failed to delete task",
			err,
		))
		return
	}

	response := &models.TaskResponse{
		Success: true,
		Message: "Task deleted successfully",
		Data:    nil,
	}
	c.JSON(http.StatusOK, response)
}

// GetTasksByStatus handles GET /tasks/status/:status - get tasks by status
// @Summary Get tasks by status
// @Description Get all tasks with a specific status
// @Tags tasks
// @Accept json
// @Produce json
// @Param status path int true "Task Status (0=incomplete, 1=completed)"
// @Success 200 {object} models.TaskListResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /tasks/status/{status} [get]
func (h *TaskHandler) GetTasksByStatus(c *gin.Context) {
	statusStr := c.Param("status")
	if statusStr == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Status is required",
			nil,
		))
		return
	}

	// Parse status
	statusInt, err := strconv.Atoi(statusStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Invalid status format",
			err,
		))
		return
	}

	status := models.TaskStatus(statusInt)
	if !status.IsValid() {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Invalid status value. Must be 0 (incomplete) or 1 (completed)",
			nil,
		))
		return
	}

	// Get tasks by status (if storage supports it)
	if memStorage, ok := h.storage.(*storage.MemoryStorage); ok {
		tasks, err := memStorage.GetTasksByStatus(status)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
				"Failed to retrieve tasks by status",
				err,
			))
			return
		}

		response := models.NewTaskListResponse(tasks)
		c.JSON(http.StatusOK, response)
		return
	}

	// Fallback: get all tasks and filter
	allTasks, err := h.storage.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Failed to retrieve tasks",
			err,
		))
		return
	}

	// Filter tasks by status
	var filteredTasks []*models.Task
	for _, task := range allTasks {
		if task.Status == status {
			filteredTasks = append(filteredTasks, task)
		}
	}

	response := models.NewTaskListResponse(filteredTasks)
	c.JSON(http.StatusOK, response)
}

// GetTasksPaginated handles GET /tasks/paginated - get tasks with pagination
// @Summary Get tasks with pagination
// @Description Get tasks with pagination support
// @Tags tasks
// @Accept json
// @Produce json
// @Param offset query int false "Offset for pagination (default: 0)"
// @Param limit query int false "Limit for pagination (default: 10)"
// @Success 200 {object} models.TaskListResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /tasks/paginated [get]
func (h *TaskHandler) GetTasksPaginated(c *gin.Context) {
	// Parse query parameters
	offsetStr := c.DefaultQuery("offset", "0")
	limitStr := c.DefaultQuery("limit", "10")

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Invalid offset parameter",
			err,
		))
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Invalid limit parameter (must be between 1 and 100)",
			err,
		))
		return
	}

	// Get paginated tasks (if storage supports it)
	if memStorage, ok := h.storage.(*storage.MemoryStorage); ok {
		tasks, total, err := memStorage.GetTasksPaginated(offset, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
				"Failed to retrieve paginated tasks",
				err,
			))
			return
		}

		response := models.NewTaskListResponse(tasks)
		// Add pagination metadata to response headers
		c.Header("X-Total-Count", strconv.Itoa(total))
		c.Header("X-Offset", strconv.Itoa(offset))
		c.Header("X-Limit", strconv.Itoa(limit))

		c.JSON(http.StatusOK, response)
		return
	}

	// Fallback: get all tasks and slice
	allTasks, err := h.storage.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Failed to retrieve tasks",
			err,
		))
		return
	}

	total := len(allTasks)

	// Handle pagination manually
	if offset >= total {
		response := models.NewTaskListResponse([]*models.Task{})
		c.Header("X-Total-Count", strconv.Itoa(total))
		c.Header("X-Offset", strconv.Itoa(offset))
		c.Header("X-Limit", strconv.Itoa(limit))
		c.JSON(http.StatusOK, response)
		return
	}

	end := offset + limit
	if end > total {
		end = total
	}

	paginatedTasks := allTasks[offset:end]
	response := models.NewTaskListResponse(paginatedTasks)

	// Add pagination metadata to response headers
	c.Header("X-Total-Count", strconv.Itoa(total))
	c.Header("X-Offset", strconv.Itoa(offset))
	c.Header("X-Limit", strconv.Itoa(limit))

	c.JSON(http.StatusOK, response)
}

// HealthCheck handles GET /health - health check endpoint
// @Summary Health check
// @Description Check if the service is healthy
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} models.HealthResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /health [get]
func (h *TaskHandler) HealthCheck(c *gin.Context) {
	// Check storage health if it implements HealthChecker
	if healthChecker, ok := h.storage.(interfaces.HealthChecker); ok {
		if err := healthChecker.HealthCheck(); err != nil {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
				"Storage health check failed",
				err,
			))
			return
		}
	}

	response := models.NewHealthResponse("1.0.0")
	c.JSON(http.StatusOK, response)
}

// GetStorageStats handles GET /stats - get storage statistics
// @Summary Get storage statistics
// @Description Get statistics about the storage
// @Tags stats
// @Accept json
// @Produce json
// @Success 200 {object} storage.StorageStats
// @Failure 500 {object} models.ErrorResponse
// @Router /stats [get]
func (h *TaskHandler) GetStorageStats(c *gin.Context) {
	// Check if storage supports stats
	if memStorage, ok := h.storage.(*storage.MemoryStorage); ok {
		stats := memStorage.GetStats()
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    stats,
		})
		return
	}

	// Fallback: basic stats
	count, err := h.storage.Count()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Failed to get task count",
			err,
		))
		return
	}

	stats := map[string]interface{}{
		"total_tasks":  count,
		"storage_type": "unknown",
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// Helper function to check if a string contains a substring
func contains(str, substr string) bool {
	return len(str) >= len(substr) && (str == substr || (len(str) > len(substr) &&
		(str[:len(substr)] == substr || str[len(str)-len(substr):] == substr ||
			containsMiddle(str, substr))))
}

func containsMiddle(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
