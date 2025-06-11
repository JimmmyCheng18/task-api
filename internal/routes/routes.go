package routes

import (
	"task-api/internal/handlers"
	"task-api/internal/interfaces"
	"task-api/internal/middleware"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// RouterConfig defines configuration for the router
type RouterConfig struct {
	EnableCORS      bool                       `json:"enable_cors"`       // Enable CORS middleware
	EnableLogging   bool                       `json:"enable_logging"`    // Enable request logging
	EnableSecurity  bool                       `json:"enable_security"`   // Enable security headers
	EnableRequestID bool                       `json:"enable_request_id"` // Enable request ID generation
	EnableRateLimit bool                       `json:"enable_rate_limit"` // Enable rate limiting
	TrustedProxies  []string                   `json:"trusted_proxies"`   // Trusted proxy IPs
	AllowedOrigins  []string                   `json:"allowed_origins"`   // CORS allowed origins
	DevelopmentMode bool                       `json:"development_mode"`  // Development mode flag
	RateLimitConfig middleware.RateLimitConfig `json:"rate_limit_config"` // Rate limiting configuration
}

// SetupRouterWithConfig configures and returns a Gin router with custom configuration
func SetupRouterWithConfig(storage interfaces.TaskStorage, config RouterConfig) *gin.Engine {
	// Set Gin mode based on configuration
	if config.DevelopmentMode {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	router := gin.New()

	// Set trusted proxies
	if len(config.TrustedProxies) > 0 {
		// Set trusted proxies and ignore error for router configuration
		_ = router.SetTrustedProxies(config.TrustedProxies)
	}

	// Recovery middleware (always enabled)
	router.Use(gin.Recovery())

	// Request ID middleware
	if config.EnableRequestID {
		router.Use(middleware.RequestID())
	}

	// Security headers middleware
	if config.EnableSecurity {
		router.Use(middleware.SecurityHeaders())
	}

	// CORS middleware
	if config.EnableCORS {
		if config.DevelopmentMode {
			router.Use(middleware.DevelopmentCORS())
		} else if len(config.AllowedOrigins) > 0 && config.AllowedOrigins[0] != "*" {
			router.Use(middleware.RestrictiveCORS(config.AllowedOrigins))
		} else {
			router.Use(middleware.CORS())
		}
	}

	// Rate limiting middleware (before logging to avoid logging blocked requests)
	if config.EnableRateLimit {
		if config.DevelopmentMode {
			// More lenient rate limiting for development
			devConfig := config.RateLimitConfig
			devConfig.PerIP = devConfig.PerIP * 2
			router.Use(middleware.SmartRateLimit(devConfig))
		} else {
			router.Use(middleware.SmartRateLimit(config.RateLimitConfig))
		}
	}

	// Logging middleware
	if config.EnableLogging {
		if config.DevelopmentMode {
			router.Use(middleware.DevelopmentLogger())
		} else {
			router.Use(middleware.ProductionLogger())
		}
	}

	// Error logging middleware
	router.Use(middleware.ErrorLogger())

	// Setup routes
	setupAPIRoutes(router, storage)

	return router
}

// setupAPIRoutes configures all API routes
func setupAPIRoutes(router *gin.Engine, storage interfaces.TaskStorage) {
	// Create task handler
	taskHandler := handlers.NewTaskHandler(storage)

	// API v1 group
	v1 := router.Group("/api/v1")
	{
		// Health check endpoint (outside of tasks group)
		v1.GET("/health", taskHandler.HealthCheck)

		// Statistics endpoint
		v1.GET("/stats", taskHandler.GetStorageStats)

		// Tasks group
		tasks := v1.Group("/tasks")
		{
			// Basic CRUD operations
			tasks.GET("", taskHandler.GetAllTasks)       // GET /api/v1/tasks
			tasks.POST("", taskHandler.CreateTask)       // POST /api/v1/tasks
			tasks.GET("/:id", taskHandler.GetTaskByID)   // GET /api/v1/tasks/:id
			tasks.PUT("/:id", taskHandler.UpdateTask)    // PUT /api/v1/tasks/:id
			tasks.DELETE("/:id", taskHandler.DeleteTask) // DELETE /api/v1/tasks/:id

			// Additional endpoints
			tasks.GET("/status/:status", taskHandler.GetTasksByStatus) // GET /api/v1/tasks/status/:status
			tasks.GET("/paginated", taskHandler.GetTasksPaginated)     // GET /api/v1/tasks/paginated
		}
	}

	// Add root health check for convenience
	router.GET("/health", taskHandler.HealthCheck)

	// Setup Swagger documentation
	setupSwaggerRoutes(router)

	// Add API documentation endpoint
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Task API",
			"version": "1.0.0",
			"endpoints": map[string]interface{}{
				"health":  "/health or /api/v1/health",
				"stats":   "/api/v1/stats",
				"swagger": "/swagger/index.html",
				"docs":    "/docs/swagger.json",
				"tasks": map[string]string{
					"list":      "GET /api/v1/tasks",
					"create":    "POST /api/v1/tasks",
					"get":       "GET /api/v1/tasks/:id",
					"update":    "PUT /api/v1/tasks/:id",
					"delete":    "DELETE /api/v1/tasks/:id",
					"by_status": "GET /api/v1/tasks/status/:status",
					"paginated": "GET /api/v1/tasks/paginated",
				},
			},
		})
	})
}

// SetupTestRouter creates a router suitable for testing
func SetupTestRouter(storage interfaces.TaskStorage) *gin.Engine {
	gin.SetMode(gin.TestMode)

	config := RouterConfig{
		EnableCORS:      false, // Disable CORS for testing
		EnableLogging:   false, // Disable logging for cleaner test output
		EnableSecurity:  false, // Disable security headers for testing
		EnableRequestID: false, // Disable request ID for predictable tests
		EnableRateLimit: false, // Disable rate limiting for testing
		TrustedProxies:  []string{},
		AllowedOrigins:  []string{},
		DevelopmentMode: false,
		RateLimitConfig: middleware.DefaultRateLimitConfig(),
	}

	return SetupRouterWithConfig(storage, config)
}

// ConfigInterface defines the interface for app configuration
type ConfigInterface interface {
	GetRateLimitEnabled() bool
	GetRateLimitPerIP() int
	GetRateLimitPerAPIKey() int
	GetRateLimitCleanupTime() int
}

// SetupDevelopmentRouterWithConfig creates a router with development-friendly settings using app config
func SetupDevelopmentRouterWithConfig(storage interfaces.TaskStorage, appConfig ConfigInterface) *gin.Engine {
	rateLimitConfig := middleware.RateLimitConfig{
		Enabled:         appConfig.GetRateLimitEnabled(),
		PerIP:           appConfig.GetRateLimitPerIP() * 2, // More lenient for development
		PerAPIKey:       appConfig.GetRateLimitPerAPIKey() * 2,
		CleanupInterval: time.Duration(appConfig.GetRateLimitCleanupTime()) * time.Minute,
		WindowSize:      1 * time.Minute,
	}

	config := RouterConfig{
		EnableCORS:      true,
		EnableLogging:   true,
		EnableSecurity:  false, // Disable for easier debugging
		EnableRequestID: true,
		EnableRateLimit: true,
		TrustedProxies:  []string{"127.0.0.1", "::1"},
		AllowedOrigins:  []string{"*"},
		DevelopmentMode: true,
		RateLimitConfig: rateLimitConfig,
	}

	return SetupRouterWithConfig(storage, config)
}

// SetupProductionRouterWithConfig creates a router with production-ready settings using app config
func SetupProductionRouterWithConfig(storage interfaces.TaskStorage, allowedOrigins []string, appConfig ConfigInterface) *gin.Engine {
	rateLimitConfig := middleware.RateLimitConfig{
		Enabled:         appConfig.GetRateLimitEnabled(),
		PerIP:           appConfig.GetRateLimitPerIP(),
		PerAPIKey:       appConfig.GetRateLimitPerAPIKey(),
		CleanupInterval: time.Duration(appConfig.GetRateLimitCleanupTime()) * time.Minute,
		WindowSize:      1 * time.Minute,
	}

	config := RouterConfig{
		EnableCORS:      true,
		EnableLogging:   true,
		EnableSecurity:  true,
		EnableRequestID: true,
		EnableRateLimit: true,
		TrustedProxies:  []string{"127.0.0.1"},
		AllowedOrigins:  allowedOrigins,
		DevelopmentMode: false,
		RateLimitConfig: rateLimitConfig,
	}

	return SetupRouterWithConfig(storage, config)
}

// SetupMetricsEndpoint adds a metrics endpoint for monitoring
func SetupMetricsEndpoint(router *gin.Engine, storage interfaces.TaskStorage) {
	router.GET("/metrics", func(c *gin.Context) {
		// Basic metrics - could be extended to Prometheus format
		count, _ := storage.Count()

		metrics := map[string]interface{}{
			"total_tasks": count,
			"uptime":      "TODO: implement uptime tracking",
			"version":     "1.0.0",
		}

		// If storage supports more detailed stats
		if statsProvider, ok := storage.(interface{ GetStats() interface{} }); ok {
			metrics["storage_stats"] = statsProvider.GetStats()
		}

		c.JSON(200, gin.H{
			"metrics": metrics,
		})
	})

	// Add rate limit stats endpoint
	router.GET("/metrics/rate-limit", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Rate limit statistics endpoint",
			"note":    "Rate limit statistics are handled by middleware and would need middleware reference to display",
		})
	})
}

// RouteInfo represents information about a registered route
type RouteInfo struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	HandlerName string `json:"handler_name"`
}

// GetRegisteredRoutes returns information about all registered routes
func GetRegisteredRoutes(router *gin.Engine) []RouteInfo {
	var routes []RouteInfo

	for _, route := range router.Routes() {
		routes = append(routes, RouteInfo{
			Method:      route.Method,
			Path:        route.Path,
			HandlerName: route.Handler,
		})
	}

	return routes
}

// SetupDebugRoutes adds debug endpoints for development
func SetupDebugRoutes(router *gin.Engine) {
	debug := router.Group("/debug")
	{
		// List all registered routes
		debug.GET("/routes", func(c *gin.Context) {
			routes := GetRegisteredRoutes(router)
			c.JSON(200, gin.H{
				"routes": routes,
				"count":  len(routes),
			})
		})

		// Echo endpoint for testing
		debug.POST("/echo", func(c *gin.Context) {
			var body interface{}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(400, gin.H{"error": err.Error()})
				return
			}

			c.JSON(200, gin.H{
				"method":  c.Request.Method,
				"path":    c.Request.URL.Path,
				"headers": c.Request.Header,
				"body":    body,
			})
		})
	}
}

// setupSwaggerRoutes adds Swagger documentation routes
func setupSwaggerRoutes(router *gin.Engine) {
	// Swagger UI endpoint with custom configuration
	url := ginSwagger.URL("doc.json") // The url pointing to API definition
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

	// Direct access to swagger JSON
	router.Static("/docs", "./docs")

	// Redirect /swagger to /swagger/index.html for convenience
	router.GET("/swagger", func(c *gin.Context) {
		c.Redirect(301, "/swagger/index.html")
	})
}
