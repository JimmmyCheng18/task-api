package middleware

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// LogLevel defines the log level enum
type LogLevel int

const (
	// LogLevelDebug for debug messages
	LogLevelDebug LogLevel = iota
	// LogLevelInfo for info messages
	LogLevelInfo
	// LogLevelWarn for warning messages
	LogLevelWarn
	// LogLevelError for error messages
	LogLevelError
)

// String returns the string representation of log level
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// LoggerConfig defines the configuration for logger middleware
type LoggerConfig struct {
	Output      io.Writer `json:"-"`            // Output destination
	TimeFormat  string    `json:"time_format"`  // Time format for logs
	LogLevel    LogLevel  `json:"log_level"`    // Minimum log level
	SkipPaths   []string  `json:"skip_paths"`   // Paths to skip logging
	EnableColor bool      `json:"enable_color"` // Enable colored output
}

// LoggerWithConfig returns a logger middleware with custom configuration
func LoggerWithConfig(config LoggerConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip logging for specified paths
		path := c.Request.URL.Path
		for _, skipPath := range config.SkipPaths {
			if path == skipPath {
				c.Next()
				return
			}
		}

		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get request information
		statusCode := c.Writer.Status()
		method := c.Request.Method
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		// Determine log level based on status code
		logLevel := getLogLevelFromStatus(statusCode)

		// Skip if log level is below configured level
		if logLevel < config.LogLevel {
			return
		}

		// Format and write log
		logMessage := formatLogMessage(LogEntry{
			Timestamp:  start,
			Method:     method,
			Path:       path,
			StatusCode: statusCode,
			Latency:    latency,
			ClientIP:   clientIP,
			UserAgent:  userAgent,
			LogLevel:   logLevel,
		}, config)

		config.Output.Write([]byte(logMessage + "\n"))
	}
}

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp  time.Time     `json:"timestamp"`
	Method     string        `json:"method"`
	Path       string        `json:"path"`
	StatusCode int           `json:"status_code"`
	Latency    time.Duration `json:"latency"`
	ClientIP   string        `json:"client_ip"`
	UserAgent  string        `json:"user_agent"`
	LogLevel   LogLevel      `json:"log_level"`
}

// formatLogMessage formats a log entry into a readable string
func formatLogMessage(entry LogEntry, config LoggerConfig) string {
	timestamp := entry.Timestamp.Format(config.TimeFormat)

	// Color codes for different status codes (if enabled)
	var statusColor, resetColor string
	if config.EnableColor {
		statusColor = getStatusColor(entry.StatusCode)
		resetColor = "\033[0m"
	}

	// Format latency
	latencyStr := formatLatency(entry.Latency)

	// Build log message
	logMsg := fmt.Sprintf("[%s] %s%3d%s %13v | %15s | %-7s %s",
		timestamp,
		statusColor,
		entry.StatusCode,
		resetColor,
		latencyStr,
		entry.ClientIP,
		entry.Method,
		entry.Path,
	)

	// Add user agent for debug level
	if entry.LogLevel == LogLevelDebug && entry.UserAgent != "" {
		logMsg += fmt.Sprintf(" | %s", entry.UserAgent)
	}

	return logMsg
}

// getLogLevelFromStatus determines log level based on HTTP status code
func getLogLevelFromStatus(statusCode int) LogLevel {
	switch {
	case statusCode >= 500:
		return LogLevelError
	case statusCode >= 400:
		return LogLevelWarn
	case statusCode >= 300:
		return LogLevelInfo
	default:
		return LogLevelDebug
	}
}

// getStatusColor returns ANSI color code for status code
func getStatusColor(statusCode int) string {
	switch {
	case statusCode >= 200 && statusCode < 300:
		return "\033[32m" // Green
	case statusCode >= 300 && statusCode < 400:
		return "\033[33m" // Yellow
	case statusCode >= 400 && statusCode < 500:
		return "\033[31m" // Red
	case statusCode >= 500:
		return "\033[35m" // Magenta
	default:
		return "\033[37m" // White
	}
}

// formatLatency formats duration to a readable string
func formatLatency(latency time.Duration) string {
	switch {
	case latency > time.Minute:
		return fmt.Sprintf("%.2fm", latency.Minutes())
	case latency > time.Second:
		return fmt.Sprintf("%.2fs", latency.Seconds())
	case latency > time.Millisecond:
		return fmt.Sprintf("%.2fms", float64(latency.Nanoseconds())/1e6)
	case latency > time.Microsecond:
		return fmt.Sprintf("%.2fμs", float64(latency.Nanoseconds())/1e3)
	default:
		return fmt.Sprintf("%dns", latency.Nanoseconds())
	}
}

// ProductionLogger returns a logger configuration suitable for production
func ProductionLogger() gin.HandlerFunc {
	config := LoggerConfig{
		Output:      os.Stdout,
		TimeFormat:  time.RFC3339,
		LogLevel:    LogLevelInfo,
		SkipPaths:   []string{"/health", "/metrics"},
		EnableColor: false,
	}

	return LoggerWithConfig(config)
}

// DevelopmentLogger returns a logger configuration suitable for development
func DevelopmentLogger() gin.HandlerFunc {
	config := LoggerConfig{
		Output:      os.Stdout,
		TimeFormat:  "2006/01/02 - 15:04:05",
		LogLevel:    LogLevelDebug,
		SkipPaths:   []string{},
		EnableColor: true,
	}

	return LoggerWithConfig(config)
}

// RequestID middleware adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to get request ID from header first
		requestID := c.GetHeader("X-Request-ID")

		// Generate one if not present
		if requestID == "" {
			requestID = generateRequestID()
		}

		// Set request ID in context and response header
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// generateRequestID generates a simple request ID
func generateRequestID() string {
	// Simple implementation using timestamp and random component
	now := time.Now()
	return fmt.Sprintf("%d-%d", now.Unix(), now.Nanosecond()%1000000)
}

// ErrorLogger logs errors that occur during request processing
func ErrorLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there were any errors
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				logError(c, err)
			}
		}
	}
}

// logError logs an error with context information
func logError(c *gin.Context, err *gin.Error) {
	timestamp := time.Now().Format(time.RFC3339)
	method := c.Request.Method
	path := c.Request.URL.Path
	clientIP := c.ClientIP()
	requestID := c.GetString("request_id")

	errorMsg := fmt.Sprintf("[ERROR] %s | %s | %s %s | Request-ID: %s | Error: %s",
		timestamp,
		clientIP,
		method,
		path,
		requestID,
		err.Error(),
	)

	fmt.Fprintln(os.Stderr, errorMsg)
}
