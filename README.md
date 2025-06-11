# Task API

High-performance REST API for task management built with Go, featuring clean architecture and advanced concurrency patterns.

## 🏗️ Architecture

Clean, layered architecture with clear separation of concerns:

```
task-api-service/
├── cmd/                              # 🚀 Application entry points
│   └── server/
│       └── main.go                   # Main application entry point
│
├── internal/                         # 🔒 Internal packages (Go 1.4+ feature)
│   ├── models/                       # Data models and DTOs
│   │   └── task.go                   # Task model, requests, responses
│   │
│   ├── interfaces/                   # Abstract interfaces
│   │   └── storage.go                # Storage interface
│   │
│   ├── storage/                      # Storage layer implementations
│   │   ├── memory.go                 # In-memory storage
│   │   └── memory_test.go            # Memory storage tests
│   │
│   ├── handlers/                     # HTTP handlers
│   │   ├── task.go                   # Task CRUD handlers
│   │   └── task_test.go              # Handler tests
│   │
│   ├── middleware/                   # Middleware components
│   │   ├── cors.go                   # CORS middleware
│   │   ├── logger.go                 # Logging middleware
│   │   ├── rate_limit.go             # Rate limiting
│   │   └── rate_limit_test.go        # Rate limit tests
│   │
│   ├── routes/                       # Route configuration
│   │   └── routes.go                 # Route definitions
│   │
│   └── config/                       # Configuration management
│       └── config.go                 # Configuration struct
│
├── scripts/                          # 🔧 Build and development scripts
│   ├── build.sh                      # Build script
│   └── test.sh                       # Test script
│
├── docs/                             # 📚 Documentation & Swagger
│   ├── API_DOCUMENTATION.md          # API documentation
│   ├── docs.go                       # Swagger Go package
│   ├── swagger.json                  # Swagger JSON specification
│   └── swagger.yaml                  # Swagger YAML specification
│
├── examples/                         # 💡 Usage examples
│   └── curl_examples.sh              # cURL examples
│
├── .dockerignore                     # Docker ignore patterns
├── .env.example                      # Environment variables example
├── .gitignore                        # Git ignore patterns
├── docker-compose.yml                # Multi-service orchestration
├── Dockerfile                        # Backend container definition
├── go.mod                            # Go modules file
├── go.sum                            # Go modules checksum
├── Makefile                          # Build automation
├── README.md                         # Project README
└── VERSION                           # Version file
```

## ⚡ High Performance Features

**Sharded Memory Storage**
- 32 independent shards with FNV-1a hash distribution
- Per-shard RWMutex for minimal lock contention
- Atomic operations and object pooling
- O(1) performance for most operations

## 🚀 Quick Start

```bash
# Clone and setup
git clone <repository-url> && cd task-api
make deps

# Development with hot reload
make dev

# Custom port
PORT=3000 make dev
```

API available at `http://localhost:8080` | Swagger UI: `/swagger/index.html`

## 🧪 Testing & Quality

```bash
make test              # Run all tests
make test-coverage     # Tests with coverage
make benchmark         # Performance tests
make check             # All quality checks
```

## 📖 API Endpoints

**Core Endpoints:**
- `GET /api/v1/tasks` - List tasks (with pagination support)
- `POST /api/v1/tasks` - Create task
- `GET /api/v1/tasks/{id}` - Get task by ID
- `PUT /api/v1/tasks/{id}` - Update task
- `DELETE /api/v1/tasks/{id}` - Delete task
- `GET /api/v1/tasks/status/{status}` - Filter by status
- `GET /api/v1/stats` - Storage statistics

**Usage Example:**
```bash
# Create task
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{"name": "Complete project", "status": 0}'
```

**Interactive Documentation:** Access Swagger UI at `/swagger/index.html`

## ⚙️ Configuration

Key environment variables:
- `PORT` - Server port (default: 8080)
- `GIN_MODE` - debug/release/test (default: release)
- `ALLOWED_ORIGINS` - CORS origins (default: *)

```bash
# Quick configuration
cp .env.example .env
PORT=3000 GIN_MODE=debug make dev
```

## 🐳 Docker & Deployment

**Single Service:**
```bash
make docker-build
make docker-run
```

**Full Stack (Frontend + Backend):**
```bash
make compose-deploy
```
- Frontend: http://localhost:3666
- Backend: http://localhost:3333
- API Docs: http://localhost:3333/swagger/index.html

## 🛠️ Development

```bash
make format           # Format code
make lint            # Run linting
make install-tools   # Setup dev tools
make build-all       # Multi-platform build
```

Built with clean architecture principles, comprehensive testing, and production-ready Docker deployment.
