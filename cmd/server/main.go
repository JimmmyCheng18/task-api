// Package main provides the entry point for the Task API application.
//
// @title Task API
// @version 1.0.0
// @description A modern, well-structured REST API for task management built with Go and Gin framework.
// @description This API provides comprehensive task management functionality with support for CRUD operations,
// @description status filtering, pagination, and more.
//
// @contact.name Task API Support
// @contact.url http://www.example.com/support
// @contact.email support@example.com
//
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
//
// @BasePath /api/v1
//
// @schemes http https
//
// @tag.name tasks
// @tag.description Task management operations
//
// @tag.name health
// @tag.description Health check and monitoring endpoints
//
// @tag.name stats
// @tag.description Statistics and metrics endpoints
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"task-api/internal/config"
	"task-api/internal/routes"
	"task-api/internal/storage"
	"time"

	"github.com/gin-gonic/gin"

	_ "task-api/docs" // Import swagger docs
)

// Application represents the main application structure
type Application struct {
	server  *http.Server
	storage *storage.MemoryStorage
	config  *config.Config
}

// NewApplication creates a new application instance with dependency injection
func NewApplication(cfg *config.Config) (*Application, error) {
	// Create storage instance (Factory Pattern)
	memStorage := storage.NewMemoryStorage(cfg.MaxTasks)

	// Create router based on environment
	var router *gin.Engine
	switch cfg.Environment {
	case "debug", "development":
		router = routes.SetupDevelopmentRouterWithConfig(memStorage, cfg)
		// Add debug routes in development
		routes.SetupDebugRoutes(router)
	case "test":
		router = routes.SetupTestRouter(memStorage)
	default:
		// Parse allowed origins for production
		var allowedOrigins []string
		if cfg.AllowedOrigins != "*" && cfg.AllowedOrigins != "" {
			// Parse comma-separated origins
			for _, origin := range strings.Split(cfg.AllowedOrigins, ",") {
				trimmed := strings.TrimSpace(origin)
				if trimmed != "" {
					allowedOrigins = append(allowedOrigins, trimmed)
				}
			}
		} else {
			allowedOrigins = []string{"*"}
		}
		router = routes.SetupProductionRouterWithConfig(memStorage, allowedOrigins, cfg)
	}

	// Add metrics endpoint
	routes.SetupMetricsEndpoint(router, memStorage)

	// Create HTTP server
	server := &http.Server{
		Addr:         cfg.GetServerAddress(),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.IdleTimeout) * time.Second,
	}

	return &Application{
		server:  server,
		storage: memStorage,
		config:  cfg,
	}, nil
}

// Start starts the application server
func (app *Application) Start() error {
	// Print startup information
	printStartupInfo(app.config)

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on %s", app.server.Addr)
		if err := app.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	return nil
}

// Stop gracefully stops the application
func (app *Application) Stop() error {
	log.Println("Shutting down server...")

	// Create context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(app.config.ShutdownTimeout)*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := app.server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
		return err
	}

	log.Println("Server stopped gracefully")
	return nil
}

// WaitForShutdown waits for shutdown signals and handles graceful shutdown
func (app *Application) WaitForShutdown() {
	// Create channel to receive OS signals
	quit := make(chan os.Signal, 1)

	// Register channel to receive specific signals
	signal.Notify(quit,
		syscall.SIGINT,  // Ctrl+C
		syscall.SIGTERM, // Termination signal
		syscall.SIGQUIT, // Quit signal
	)

	// Block until signal is received
	sig := <-quit
	log.Printf("Received signal: %v", sig)

	// Perform graceful shutdown
	if err := app.Stop(); err != nil {
		log.Printf("Error during shutdown: %v", err)
		os.Exit(1)
	}
}

// HealthCheck performs application health check
func (app *Application) HealthCheck() error {
	// Check storage health
	if err := app.storage.HealthCheck(); err != nil {
		return fmt.Errorf("storage health check failed: %w", err)
	}

	// Add more health checks here as needed
	// - Database connectivity
	// - External service availability
	// - Memory usage
	// - Disk space

	return nil
}

// GetStats returns application statistics
func (app *Application) GetStats() map[string]interface{} {
	stats := map[string]interface{}{
		"server_addr": app.server.Addr,
		"environment": app.config.Environment,
		"storage":     app.storage.GetStats(),
	}

	return stats
}

// main is the application entry point
func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Set Gin mode
	gin.SetMode(cfg.Environment)

	// Create application instance
	app, err := NewApplication(cfg)
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}

	// Perform initial health check
	if err := app.HealthCheck(); err != nil {
		log.Fatalf("Initial health check failed: %v", err)
	}

	// Start the application
	if err := app.Start(); err != nil {
		log.Fatalf("Failed to start application: %v", err)
	}

	// Wait for shutdown signal
	app.WaitForShutdown()
}

// printStartupInfo prints application startup information
func printStartupInfo(cfg *config.Config) {
	log.Println("=================================")
	log.Println("      Task API Starting Up       ")
	log.Println("=================================")
	log.Printf("Environment: %s", cfg.Environment)
	log.Printf("Address: %s", cfg.GetServerAddress())
	log.Printf("Read Timeout: %ds", cfg.ReadTimeout)
	log.Printf("Write Timeout: %ds", cfg.WriteTimeout)
	log.Printf("Idle Timeout: %ds", cfg.IdleTimeout)
	log.Printf("Shutdown Timeout: %ds", cfg.ShutdownTimeout)
	log.Printf("Allowed Origins: %s", cfg.AllowedOrigins)
	log.Println("=================================")

	// Print available endpoints
	log.Println("Available Endpoints:")
	log.Printf("  Health Check: http://%s/health", cfg.GetServerAddress())
	log.Printf("  API Documentation: http://%s/", cfg.GetServerAddress())
	log.Printf("  Tasks API: http://%s/api/v1/tasks", cfg.GetServerAddress())
	log.Printf("  Metrics: http://%s/metrics", cfg.GetServerAddress())
	log.Printf("  Stats: http://%s/api/v1/stats", cfg.GetServerAddress())

	if cfg.IsDevelopment() {
		log.Printf("  Debug Routes: http://%s/debug/routes", cfg.GetServerAddress())
		log.Printf("  Debug Echo: http://%s/debug/echo", cfg.GetServerAddress())
	}

	log.Println("=================================")
}

// init function for any initialization logic
func init() {
	// Set log format
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Set log output
	log.SetOutput(os.Stdout)

	// You could add more initialization here:
	// - Load additional configuration files
	// - Initialize external dependencies
	// - Set up monitoring/telemetry
}
