# Multi-stage build for Task API
# Stage 1: Build stage
FROM golang:1.24-alpine AS builder

# Set necessary environment variables
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Create and set the working directory
WORKDIR /build

# Install git and ca-certificates for dependency resolution
RUN apk update && apk add --no-cache git ca-certificates tzdata

# Copy go mod and sum files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy the source code
COPY . .

# Build the application
RUN go build -a -installsuffix cgo -ldflags="-w -s" -o task-api ./cmd/server

# Stage 2: Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests and tzdata for timezone
RUN apk --no-cache add ca-certificates tzdata

# Create a non-root user for security
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /build/task-api .

# Copy the docs directory for Swagger UI
COPY --from=builder /build/docs ./docs

# Copy any additional files if needed (like config files)
# COPY --from=builder /build/config ./config

# Change ownership to non-root user
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Add health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the binary
ENTRYPOINT ["./task-api"]