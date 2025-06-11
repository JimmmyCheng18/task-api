# Task API Documentation

A comprehensive RESTful API for task management built with Go and Gin framework.

## Table of Contents

- [Overview](#overview)
- [Base URL](#base-url)
- [Authentication](#authentication)
- [Response Format](#response-format)
- [Error Handling](#error-handling)
- [Endpoints](#endpoints)
- [Data Models](#data-models)
- [Examples](#examples)
- [Rate Limiting](#rate-limiting)
- [Health Check](#health-check)

## Overview

The Task API provides a simple and efficient way to manage tasks with full CRUD operations. It supports task creation, retrieval, updating, and deletion with additional features like status filtering and pagination.

### Features

- ✅ Full CRUD operations for tasks
- ✅ Task status management (incomplete/completed)
- ✅ Pagination support
- ✅ Status-based filtering
- ✅ Thread-safe in-memory storage
- ✅ Health check endpoints
- ✅ Comprehensive error handling
- ✅ Request/Response validation

## Base URL

```
http://localhost:8080/api/v1
```

## Authentication

Currently, the API does not require authentication. This may be added in future versions.

## Response Format

All API responses follow a consistent JSON format:

### Success Response Format

```json
{
  "success": true,
  "message": "Operation completed successfully",
  "data": {
    // Response data here
  }
}
```

### List Response Format

```json
{
  "success": true,
  "data": [
    // Array of items
  ],
  "count": 10
}
```

### Error Response Format

```json
{
  "success": false,
  "message": "Error description",
  "error": "Detailed error message"
}
```

## Error Handling

The API uses standard HTTP status codes:

| Status Code | Description |
|-------------|-------------|
| 200 | OK - Request successful |
| 201 | Created - Resource created successfully |
| 400 | Bad Request - Invalid request data |
| 404 | Not Found - Resource not found |
| 500 | Internal Server Error - Server error |

## Endpoints

### Tasks

#### Get All Tasks

Retrieve all tasks from the system.

```http
GET /api/v1/tasks
```

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "1",
      "name": "Complete project documentation",
      "status": 0,
      "created_at": "2025-06-09T22:00:00Z",
      "updated_at": "2025-06-09T22:00:00Z"
    }
  ],
  "count": 1
}
```

#### Get Task by ID

Retrieve a specific task by its ID.

```http
GET /api/v1/tasks/{id}
```

**Parameters:**
- `id` (path parameter): Task ID

**Response:**
```json
{
  "success": true,
  "message": "Task retrieved successfully",
  "data": {
    "id": "1",
    "name": "Complete project documentation",
    "status": 0,
    "created_at": "2025-06-09T22:00:00Z",
    "updated_at": "2025-06-09T22:00:00Z"
  }
}
```

#### Create Task

Create a new task.

```http
POST /api/v1/tasks
```

**Request Body:**
```json
{
  "name": "New task name",
  "status": 0
}
```

**Response:**
```json
{
  "success": true,
  "message": "Task created successfully",
  "data": {
    "id": "2",
    "name": "New task name",
    "status": 0,
    "created_at": "2025-06-09T22:05:00Z",
    "updated_at": "2025-06-09T22:05:00Z"
  }
}
```

#### Update Task

Update an existing task.

```http
PUT /api/v1/tasks/{id}
```

**Parameters:**
- `id` (path parameter): Task ID

**Request Body:**
```json
{
  "name": "Updated task name",
  "status": 1
}
```

**Note:** Both `name` and `status` fields are optional. You can update just one field.

**Response:**
```json
{
  "success": true,
  "message": "Task updated successfully",
  "data": {
    "id": "1",
    "name": "Updated task name",
    "status": 1,
    "created_at": "2025-06-09T22:00:00Z",
    "updated_at": "2025-06-09T22:10:00Z"
  }
}
```

#### Delete Task

Delete a task by its ID.

```http
DELETE /api/v1/tasks/{id}
```

**Parameters:**
- `id` (path parameter): Task ID

**Response:**
```json
{
  "success": true,
  "message": "Task deleted successfully",
  "data": null
}
```

#### Get Tasks by Status

Retrieve tasks filtered by status.

```http
GET /api/v1/tasks/status/{status}
```

**Parameters:**
- `status` (path parameter): Task status (0 for incomplete, 1 for completed)

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "1",
      "name": "Incomplete task",
      "status": 0,
      "created_at": "2025-06-09T22:00:00Z",
      "updated_at": "2025-06-09T22:00:00Z"
    }
  ],
  "count": 1
}
```

#### Get Tasks with Pagination

Retrieve tasks with pagination support.

```http
GET /api/v1/tasks/paginated?offset=0&limit=10
```

**Query Parameters:**
- `offset` (optional): Number of items to skip (default: 0)
- `limit` (optional): Maximum number of items to return (default: 10, max: 100)

**Response Headers:**
- `X-Total-Count`: Total number of tasks
- `X-Offset`: Current offset
- `X-Limit`: Current limit

**Response:**
```json
{
  "success": true,
  "data": [
    // Array of tasks
  ],
  "count": 10
}
```

### Health Check

#### Health Status

Check the health status of the API.

```http
GET /health
```

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2025-06-09T22:00:00Z",
  "version": "1.0.0"
}
```

### Statistics

#### Get Storage Statistics

Get statistics about the task storage.

```http
GET /api/v1/stats
```

**Response:**
```json
{
  "success": true,
  "data": {
    "total_tasks": 10,
    "completed_tasks": 4,
    "incomplete_tasks": 6,
    "last_id": 10,
    "storage_type": "memory"
  }
}
```

## Data Models

### Task

| Field | Type | Description | Required |
|-------|------|-------------|----------|
| id | string | Unique task identifier | Auto-generated |
| name | string | Task name (max 255 characters) | Yes |
| status | integer | Task status (0=incomplete, 1=completed) | Yes |
| created_at | string | Creation timestamp (ISO 8601) | Auto-generated |
| updated_at | string | Last update timestamp (ISO 8601) | Auto-generated |

### Task Status

| Value | Description |
|-------|-------------|
| 0 | Incomplete |
| 1 | Completed |

## Examples

### Creating a Task

```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Review pull request",
    "status": 0
  }'
```

### Getting All Tasks

```bash
curl http://localhost:8080/api/v1/tasks
```

### Updating a Task

```bash
curl -X PUT http://localhost:8080/api/v1/tasks/1 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Review and merge pull request",
    "status": 1
  }'
```

### Deleting a Task

```bash
curl -X DELETE http://localhost:8080/api/v1/tasks/1
```

### Getting Tasks by Status

```bash
# Get incomplete tasks
curl http://localhost:8080/api/v1/tasks/status/0

# Get completed tasks
curl http://localhost:8080/api/v1/tasks/status/1
```

### Pagination

```bash
# Get first page (10 items)
curl "http://localhost:8080/api/v1/tasks/paginated?offset=0&limit=10"

# Get second page
curl "http://localhost:8080/api/v1/tasks/paginated?offset=10&limit=10"
```

## Rate Limiting

Currently, no rate limiting is implemented. This may be added in future versions.

## Validation Rules

### Task Name
- Required field
- Cannot be empty
- Maximum length: 255 characters
- Must be a valid string

### Task Status
- Required field
- Must be 0 (incomplete) or 1 (completed)
- Invalid values will return a 400 error

## Error Examples

### Validation Error

```json
{
  "success": false,
  "message": "Validation failed",
  "error": "task name cannot be empty"
}
```

### Not Found Error

```json
{
  "success": false,
  "message": "Task not found",
  "error": "task with ID 999 not found"
}
```

### Server Error

```json
{
  "success": false,
  "message": "Failed to create task",
  "error": "internal server error"
}
```

## HTTP Headers

### Request Headers

| Header | Required | Description |
|--------|----------|-------------|
| Content-Type | For POST/PUT | Must be `application/json` |
| Accept | Optional | Preferred response format |

### Response Headers

| Header | Description |
|--------|-------------|
| Content-Type | Always `application/json` |
| X-Request-ID | Unique request identifier |
| X-Total-Count | Total items (pagination endpoints) |
| X-Offset | Current offset (pagination endpoints) |
| X-Limit | Current limit (pagination endpoints) |

## API Versioning

The current API version is v1, indicated by the `/api/v1` prefix in all endpoints. Future versions will use `/api/v2`, etc.

## Development and Testing

### Running Locally

```bash
# Start the server
go run main.go

# Or using make
make run
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage
```

### Using Docker

```bash
# Build Docker image
make docker-build

# Run Docker container
make docker-run
```

## Support

For issues, questions, or contributions, please refer to the project repository.