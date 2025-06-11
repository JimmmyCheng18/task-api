#!/bin/bash

# Task API Examples using cURL
# Make sure the server is running on http://localhost:${PORT:-8080}
# Usage: PORT=3111 ./examples/curl_examples.sh

PORT=${PORT:-8080}
BASE_URL="http://localhost:${PORT}/api/v1"
HEALTH_URL="http://localhost:${PORT}/health"

echo "=================================="
echo "    Task API cURL Examples"
echo "=================================="

# Function to print headers
print_header() {
    echo ""
    echo "=== $1 ==="
}

# Function to check if server is running
check_server() {
    print_header "Checking Server Health"
    
    if curl -f -s $HEALTH_URL > /dev/null; then
        echo "✅ Server is running!"
        curl -s $HEALTH_URL | jq '.' 2>/dev/null || curl -s $HEALTH_URL
    else
        echo "❌ Server is not running. Please start the server first:"
        echo "   go run main.go"
        echo "   or"
        echo "   make dev"
        exit 1
    fi
}

# 1. Health Check
check_server

# 2. Get all tasks (initially empty)
print_header "Get All Tasks (Empty)"
curl -s "$BASE_URL/tasks" | jq '.' 2>/dev/null || curl -s "$BASE_URL/tasks"

# 3. Create first task
print_header "Create First Task"
TASK1_RESPONSE=$(curl -s -X POST "$BASE_URL/tasks" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Complete project documentation",
    "status": 0
  }')

echo "$TASK1_RESPONSE" | jq '.' 2>/dev/null || echo "$TASK1_RESPONSE"

# Extract task ID for later use
TASK1_ID=$(echo "$TASK1_RESPONSE" | jq -r '.data.id' 2>/dev/null || echo "1")

# 4. Create second task
print_header "Create Second Task"
TASK2_RESPONSE=$(curl -s -X POST "$BASE_URL/tasks" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Write unit tests",
    "status": 0
  }')

echo "$TASK2_RESPONSE" | jq '.' 2>/dev/null || echo "$TASK2_RESPONSE"

TASK2_ID=$(echo "$TASK2_RESPONSE" | jq -r '.data.id' 2>/dev/null || echo "2")

# 5. Create third task (completed)
print_header "Create Third Task (Completed)"
TASK3_RESPONSE=$(curl -s -X POST "$BASE_URL/tasks" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Set up CI/CD pipeline",
    "status": 1
  }')

echo "$TASK3_RESPONSE" | jq '.' 2>/dev/null || echo "$TASK3_RESPONSE"

TASK3_ID=$(echo "$TASK3_RESPONSE" | jq -r '.data.id' 2>/dev/null || echo "3")

# 6. Get all tasks (now with data)
print_header "Get All Tasks (With Data)"
curl -s "$BASE_URL/tasks" | jq '.' 2>/dev/null || curl -s "$BASE_URL/tasks"

# 7. Get specific task by ID
print_header "Get Task by ID ($TASK1_ID)"
curl -s "$BASE_URL/tasks/$TASK1_ID" | jq '.' 2>/dev/null || curl -s "$BASE_URL/tasks/$TASK1_ID"

# 8. Update task (mark as completed)
print_header "Update Task ($TASK1_ID) - Mark as Completed"
curl -s -X PUT "$BASE_URL/tasks/$TASK1_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "status": 1
  }' | jq '.' 2>/dev/null || curl -s -X PUT "$BASE_URL/tasks/$TASK1_ID" \
  -H "Content-Type: application/json" \
  -d '{"status": 1}'

# 9. Update task name
print_header "Update Task ($TASK2_ID) - Change Name"
curl -s -X PUT "$BASE_URL/tasks/$TASK2_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Write comprehensive unit tests with >90% coverage"
  }' | jq '.' 2>/dev/null || curl -s -X PUT "$BASE_URL/tasks/$TASK2_ID" \
  -H "Content-Type: application/json" \
  -d '{"name": "Write comprehensive unit tests with >90% coverage"}'

# 10. Get incomplete tasks
print_header "Get Incomplete Tasks (Status = 0)"
curl -s "$BASE_URL/tasks/status/0" | jq '.' 2>/dev/null || curl -s "$BASE_URL/tasks/status/0"

# 11. Get completed tasks
print_header "Get Completed Tasks (Status = 1)"
curl -s "$BASE_URL/tasks/status/1" | jq '.' 2>/dev/null || curl -s "$BASE_URL/tasks/status/1"

# 12. Test pagination
print_header "Test Pagination (Limit 2, Offset 0)"
curl -s "$BASE_URL/tasks/paginated?limit=2&offset=0" \
  -H "Accept: application/json" | jq '.' 2>/dev/null || curl -s "$BASE_URL/tasks/paginated?limit=2&offset=0"

# 13. Get storage statistics
print_header "Get Storage Statistics"
curl -s "$BASE_URL/stats" | jq '.' 2>/dev/null || curl -s "$BASE_URL/stats"

# 14. Test error cases
print_header "Test Error Cases"

echo "14a. Create task with empty name (should fail):"
curl -s -X POST "$BASE_URL/tasks" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "",
    "status": 0
  }' | jq '.' 2>/dev/null || curl -s -X POST "$BASE_URL/tasks" \
  -H "Content-Type: application/json" \
  -d '{"name": "", "status": 0}'

echo ""
echo "14b. Get non-existent task (should fail):"
curl -s "$BASE_URL/tasks/999" | jq '.' 2>/dev/null || curl -s "$BASE_URL/tasks/999"

echo ""
echo "14c. Update non-existent task (should fail):"
curl -s -X PUT "$BASE_URL/tasks/999" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "This should fail"
  }' | jq '.' 2>/dev/null || curl -s -X PUT "$BASE_URL/tasks/999" \
  -H "Content-Type: application/json" \
  -d '{"name": "This should fail"}'

echo ""
echo "14d. Invalid status value (should fail):"
curl -s -X POST "$BASE_URL/tasks" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test task",
    "status": 99
  }' | jq '.' 2>/dev/null || curl -s -X POST "$BASE_URL/tasks" \
  -H "Content-Type: application/json" \
  -d '{"name": "Test task", "status": 99}'

# 15. Delete a task
print_header "Delete Task ($TASK3_ID)"
curl -s -X DELETE "$BASE_URL/tasks/$TASK3_ID" | jq '.' 2>/dev/null || curl -s -X DELETE "$BASE_URL/tasks/$TASK3_ID"

# 16. Verify deletion
print_header "Verify Task Deletion (Get All Tasks)"
curl -s "$BASE_URL/tasks" | jq '.' 2>/dev/null || curl -s "$BASE_URL/tasks"

# 17. Final statistics
print_header "Final Statistics"
curl -s "$BASE_URL/stats" | jq '.' 2>/dev/null || curl -s "$BASE_URL/stats"

print_header "API Endpoints Summary"
echo "Health Check:     GET  $HEALTH_URL"
echo "List Tasks:       GET  $BASE_URL/tasks"
echo "Get Task:         GET  $BASE_URL/tasks/{id}"
echo "Create Task:      POST $BASE_URL/tasks"
echo "Update Task:      PUT  $BASE_URL/tasks/{id}"
echo "Delete Task:      DELETE $BASE_URL/tasks/{id}"
echo "Tasks by Status:  GET  $BASE_URL/tasks/status/{status}"
echo "Paginated Tasks:  GET  $BASE_URL/tasks/paginated?offset=0&limit=10"
echo "Statistics:       GET  $BASE_URL/stats"

echo ""
echo "=================================="
echo "   Examples completed successfully!"
echo "=================================="