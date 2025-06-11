# Frontend Docker Management Makefile

# Variables
IMAGE_NAME = task-frontend
CONTAINER_NAME = task-frontend
FRONTEND_PORT = 3666
NGINX_PORT = 80

# Build the Docker image
build:
	@echo "Building frontend Docker image..."
	docker build -t $(IMAGE_NAME):latest .

# Run the Docker container
run:
	@echo "Running frontend Docker container..."
	docker run -d -p $(FRONTEND_PORT):$(NGINX_PORT) --name $(CONTAINER_NAME) $(IMAGE_NAME):latest

# Stop and remove the container
stop:
	@echo "Stopping and removing frontend container..."
	docker stop $(CONTAINER_NAME) || true
	docker rm $(CONTAINER_NAME) || true

# Restart the container
restart: stop run

# View container logs
logs:
	docker logs $(CONTAINER_NAME)

# Access container shell
shell:
	docker exec -it $(CONTAINER_NAME) /bin/sh

# Check container status
status:
	docker ps --filter name=$(CONTAINER_NAME)

# Clean up - remove container and image
clean: stop
	@echo "Cleaning up frontend Docker resources..."
	docker rmi $(IMAGE_NAME):latest || true

# Build and run
deploy: build run

# Show help
help:
	@echo "Available commands:"
	@echo "  build    - Build the Docker image"
	@echo "  run      - Run the Docker container"
	@echo "  stop     - Stop and remove the container"
	@echo "  restart  - Restart the container"
	@echo "  logs     - View container logs"
	@echo "  shell    - Access container shell"
	@echo "  status   - Check container status"
	@echo "  clean    - Remove container and image"
	@echo "  deploy   - Build and run"
	@echo "  help     - Show this help"

.PHONY: build run stop restart logs shell status clean deploy help