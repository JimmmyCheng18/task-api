# Task API

A modern, well-structured REST API for task management built with Go and Gin framework, following clean architecture principles and industry best practices.

## 🏗️ Architecture

This project follows a clean, layered architecture with clear separation of concerns:

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
│   │   └── logger.go                 # Logging middleware
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
├── Dockerfile                        # Docker container definition
├── go.mod                            # Go modules file
├── go.sum                            # Go modules checksum
├── Makefile                          # Build automation
├── README.md                         # Project README
└── VERSION                           # Version file
```

## 🚀 Quick Start

### Prerequisites

- Go 1.24
- Git

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd task-api
   ```

2. **Install dependencies**
   ```bash
   make deps
   ```

3. **Build the application**
   ```bash
   make build
   ```

4. **Run the application**
   ```bash
   make run
   ```
   
   Or specify a custom port:
   ```bash
   PORT=3000 make run
   ```

The API will be available at `http://localhost:8080` (or your specified port)

### Development Mode

For development with hot reload:

```bash
make dev
```

Or with a custom port:
```bash
PORT=3000 make dev
```

## 🧪 Testing

### Run all tests
```bash
make test
```

### Run tests with coverage
```bash
make test-coverage
```

### Run specific test types
```bash
make test-unit        # Unit tests only
make benchmark        # Benchmark tests with coverage
```

### Benchmark Testing with Coverage

The `make benchmark` command now includes comprehensive coverage analysis:

```bash
make benchmark
```

This will:
- Run all benchmark tests with memory allocation tracking
- Generate coverage profile (`coverage/benchmark-coverage.out`)
- Display coverage summary in terminal
- Create HTML coverage report (`coverage/benchmark-coverage.html`)

**Coverage Reports Location:**
- **Profile**: `./coverage/benchmark-coverage.out`
- **HTML Report**: `./coverage/benchmark-coverage.html`

The benchmark coverage is separate from regular test coverage, allowing you to see which code paths are exercised during performance testing.

## 🔧 Development

### Code Quality

```bash
make format      # Format code
make lint        # Run linting
make security    # Security checks
make check       # Run all checks
```

### Development Tools Setup

```bash
make install-tools  # Install development tools
make setup          # Complete setup
```


## 📖 API Documentation

### 📚 Swagger UI - Interactive Documentation

**Access Swagger UI**:
- **Development**: `http://localhost:8080/swagger/index.html`
- **Custom Port**: `http://localhost:{PORT}/swagger/index.html`
- **Direct Access**: `http://localhost:8080/swagger`

**API Specifications**:
- **JSON Format**: `http://localhost:8080/docs/swagger.json`
- **YAML Format**: `http://localhost:8080/docs/swagger.yaml`

**Automatic Generation**: Swagger documentation is automatically generated during:
- `make dev` - Development mode
- `make build` - Build application
- `make build-all` - Multi-platform build
- `make docker-build` - Docker image build

**Manual Generation**:
```bash
make swagger-generate  # Generate Swagger docs
# or
make docs             # Alias for swagger-generate
```

**Quick Access**:
```bash
# Start the application
make dev

# Visit Swagger UI in your browser
open http://localhost:8080/swagger/index.html

# Or with custom port
PORT=3000 make dev
open http://localhost:3000/swagger/index.html
```

**Generated Files**:
- `docs/docs.go` - Go package for embedding
- `docs/swagger.json` - JSON API specification
- `docs/swagger.yaml` - YAML API specification

### 🔗 API Endpoints

#### Health Check
- `GET /health` - Basic health check
- `GET /api/v1/health` - Detailed health check

#### Tasks
- `GET /api/v1/tasks` - Get all tasks
- `GET /api/v1/tasks/paginated` - Get tasks with pagination
- `POST /api/v1/tasks` - Create a new task
- `GET /api/v1/tasks/{id}` - Get task by ID
- `PUT /api/v1/tasks/{id}` - Update task
- `DELETE /api/v1/tasks/{id}` - Delete task
- `GET /api/v1/tasks/status/{status}` - Get tasks by status

#### Statistics
- `GET /api/v1/stats` - Get storage statistics

### 💡 API Usage Examples

**Create a task:**
```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{"name": "Complete project", "status": 0}'
```

**Get all tasks:**
```bash
curl http://localhost:8080/api/v1/tasks
```

**Update a task:**
```bash
curl -X PUT http://localhost:8080/api/v1/tasks/{id} \
  -H "Content-Type: application/json" \
  -d '{"name": "Updated task", "status": 1}'
```

For more examples, see `examples/curl_examples.sh`

## ⚙️ Configuration

The application can be configured using environment variables:

- `PORT` - Server port (default: 8080)
- `HOST` - Server host (default: 0.0.0.0)
- `GIN_MODE` - Gin mode: debug, release, test (default: release)
- `ALLOWED_ORIGINS` - CORS allowed origins (default: *)
- `SHUTDOWN_TIMEOUT` - Graceful shutdown timeout in seconds (default: 30)
- `READ_TIMEOUT` - HTTP read timeout in seconds (default: 60)
- `WRITE_TIMEOUT` - HTTP write timeout in seconds (default: 60)
- `IDLE_TIMEOUT` - HTTP idle timeout in seconds (default: 120)

### Environment Configuration

1. **Copy the example environment file:**
   ```bash
   cp .env.example .env
   ```

2. **Edit the `.env` file with your preferences:**
   ```bash
   PORT=3000
   GIN_MODE=debug
   ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080
   ```

3. **Run with environment file:**
   ```bash
   source .env && make dev
   ```

### Quick Port Configuration Examples

**Development mode with custom port:**
```bash
PORT=3000 make dev
```

**Production build with custom port:**
```bash
PORT=9000 make run
```

**Direct binary execution:**
```bash
PORT=3000 GIN_MODE=debug ./bin/task-api
```

## 🐳 Docker

### Build Docker image
```bash
make docker-build
```

### Run with Docker

**Default port (8080):**
```bash
make docker-run
```

**Custom host port:**
```bash
PORT=3000 make docker-run
```

This maps host port 3000 to container port 8080.

### Docker Environment Variables

You can use environment variables when running Docker containers:

1. **Create a `.env` file:**
   ```env
   PORT=8080
   GIN_MODE=release
   ALLOWED_ORIGINS=*
   ```

### Docker Run Examples

**Manual Docker run with custom settings:**
```bash
# Run on custom port with environment variables
docker run --rm -p 3000:8080 \
  -e PORT=8080 \
  -e GIN_MODE=debug \
  -e ALLOWED_ORIGINS="*" \
  task-api:latest

# Run with custom internal port
docker run --rm -p 3000:9000 \
  -e PORT=9000 \
  -e HOST=0.0.0.0 \
  task-api:latest
```

## 📦 Project Structure Principles

### Clean Architecture
- **Separation of Concerns**: Each layer has a single responsibility
- **Dependency Inversion**: Higher-level modules don't depend on lower-level modules
- **Interface Segregation**: Clients don't depend on interfaces they don't use

### Package Organization
- `cmd/` - Application entry points
- `internal/` - Private application code
- `docs/` - Documentation and API specifications
- `examples/` - Usage examples and scripts
- `scripts/` - Build and deployment scripts

### Code Quality Standards
- Comprehensive test coverage
- Clear error handling
- Consistent naming conventions
- Proper documentation
- Security best practices

## 🛠️ Development Workflow

1. **Feature Development**
   ```bash
   git checkout -b feature/new-feature
   make dev  # Start development server
   # Write code and tests
   make check  # Run quality checks
   ```

2. **Testing**
   ```bash
   make test-coverage  # Ensure good coverage
   make check         # All quality checks
   ```

3. **Building**
   ```bash
   make build      # Single platform
   make build-all  # All platforms
   ```

4. **Release**
   ```bash
   make release  # Create release build
   ```