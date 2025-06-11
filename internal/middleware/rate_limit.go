package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimitConfig defines rate limiting configuration
type RateLimitConfig struct {
	Enabled         bool          // Enable rate limiting
	PerIP           int           // Requests per minute per IP
	PerAPIKey       int           // Requests per minute per API key
	CleanupInterval time.Duration // Interval for cleaning up expired records
	WindowSize      time.Duration // Time window size
}

// DefaultRateLimitConfig returns default rate limiting configuration
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Enabled:         true,
		PerIP:           100,             // 100 requests per minute per IP
		PerAPIKey:       1000,            // 1000 requests per minute per API key
		CleanupInterval: 5 * time.Minute, // Cleanup every 5 minutes
		WindowSize:      1 * time.Minute, // 1 minute time window
	}
}

// RequestRecord tracks request information
type RequestRecord struct {
	Count     int       // Request count
	FirstSeen time.Time // First request time
	LastSeen  time.Time // Last request time
}

// RateLimiter implements rate limiting functionality
type RateLimiter struct {
	config     RateLimitConfig
	ipRecords  map[string]*RequestRecord // IP request records
	keyRecords map[string]*RequestRecord // API key request records
	mu         sync.RWMutex              // Read-write mutex
	stopChan   chan struct{}             // Channel to stop cleanup routine
}

// NewRateLimiter creates a new rate limiter instance
func NewRateLimiter(config RateLimitConfig) *RateLimiter {
	limiter := &RateLimiter{
		config:     config,
		ipRecords:  make(map[string]*RequestRecord),
		keyRecords: make(map[string]*RequestRecord),
		stopChan:   make(chan struct{}),
	}

	// Start cleanup routine
	go limiter.startCleanupRoutine()

	return limiter
}

// RateLimit returns rate limiting middleware
func RateLimit(config RateLimitConfig) gin.HandlerFunc {
	if !config.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	limiter := NewRateLimiter(config)

	return func(c *gin.Context) {
		// Check rate limit
		if !limiter.Allow(c) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"message": "Too many requests. Please try again later.",
				"code":    "RATE_LIMIT_EXCEEDED",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Allow checks if the request is allowed
func (rl *RateLimiter) Allow(c *gin.Context) bool {
	clientIP := getClientIP(c)
	apiKey := c.GetHeader("X-API-Key")

	now := time.Now()

	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Check IP limit
	if !rl.checkLimit(clientIP, rl.config.PerIP, now, rl.ipRecords) {
		return false
	}

	// Check API key limit if present
	if apiKey != "" {
		if !rl.checkLimit(apiKey, rl.config.PerAPIKey, now, rl.keyRecords) {
			return false
		}
	}

	return true
}

// checkLimit checks the limit for a specific identifier
func (rl *RateLimiter) checkLimit(identifier string, limit int, now time.Time, records map[string]*RequestRecord) bool {
	record, exists := records[identifier]

	if !exists {
		// First request
		records[identifier] = &RequestRecord{
			Count:     1,
			FirstSeen: now,
			LastSeen:  now,
		}
		return true
	}

	// Check time window
	if now.Sub(record.FirstSeen) > rl.config.WindowSize {
		// Reset counter
		record.Count = 1
		record.FirstSeen = now
		record.LastSeen = now
		return true
	}

	// Check if limit exceeded
	if record.Count >= limit {
		return false
	}

	// Increment counter
	record.Count++
	record.LastSeen = now

	return true
}

// startCleanupRoutine starts the routine to clean up expired records
func (rl *RateLimiter) startCleanupRoutine() {
	ticker := time.NewTicker(rl.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.cleanup()
		case <-rl.stopChan:
			return
		}
	}
}

// cleanup removes expired records
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	expiry := rl.config.WindowSize * 2 // Keep records for two time windows

	// Clean up IP records
	for ip, record := range rl.ipRecords {
		if now.Sub(record.LastSeen) > expiry {
			delete(rl.ipRecords, ip)
		}
	}

	// Clean up API key records
	for key, record := range rl.keyRecords {
		if now.Sub(record.LastSeen) > expiry {
			delete(rl.keyRecords, key)
		}
	}
}

// Stop stops the rate limiter
func (rl *RateLimiter) Stop() {
	close(rl.stopChan)
}

// GetStats returns rate limiter statistics
func (rl *RateLimiter) GetStats() map[string]interface{} {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	return map[string]interface{}{
		"config": map[string]interface{}{
			"enabled":          rl.config.Enabled,
			"per_ip":           rl.config.PerIP,
			"per_api_key":      rl.config.PerAPIKey,
			"cleanup_interval": rl.config.CleanupInterval.String(),
			"window_size":      rl.config.WindowSize.String(),
		},
		"statistics": map[string]interface{}{
			"tracked_ips":      len(rl.ipRecords),
			"tracked_api_keys": len(rl.keyRecords),
		},
	}
}

// RateLimitWithConfig creates rate limiting middleware with custom configuration
func RateLimitWithConfig(config RateLimitConfig) gin.HandlerFunc {
	return RateLimit(config)
}

// SmartRateLimit creates intelligent rate limiting middleware with different limits for different endpoints
func SmartRateLimit(config RateLimitConfig) gin.HandlerFunc {
	if !config.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	limiter := NewRateLimiter(config)

	return func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method

		// Set different limit strategies for different endpoints
		customLimit := getCustomLimit(path, method, config)

		// Check rate limit
		if !limiter.allowWithCustomLimit(c, customLimit) {
			// Set appropriate response headers
			c.Header("X-RateLimit-Limit", "100") // Can be set dynamically based on actual limit
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", "60") // Reset time in seconds

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"message": "Too many requests. Please try again later.",
				"code":    "RATE_LIMIT_EXCEEDED",
				"details": map[string]interface{}{
					"path":   path,
					"method": method,
					"limit":  customLimit,
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// getCustomLimit returns custom limit based on path and method
func getCustomLimit(path, method string, baseConfig RateLimitConfig) int {
	// Health check endpoints allow more requests
	if path == "/health" || path == "/api/v1/health" {
		return baseConfig.PerIP * 5
	}

	// Read operations allow more requests
	if method == "GET" {
		return baseConfig.PerIP
	}

	// Write operations have stricter limits
	if method == "POST" || method == "PUT" || method == "DELETE" {
		return baseConfig.PerIP / 2
	}

	return baseConfig.PerIP
}

// allowWithCustomLimit checks requests with custom limits
func (rl *RateLimiter) allowWithCustomLimit(c *gin.Context, customLimit int) bool {
	clientIP := getClientIP(c)
	apiKey := c.GetHeader("X-API-Key")

	now := time.Now()

	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Check IP limit (with custom limit)
	if !rl.checkLimit(clientIP, customLimit, now, rl.ipRecords) {
		return false
	}

	// Check API key limit if present
	if apiKey != "" {
		if !rl.checkLimit(apiKey, rl.config.PerAPIKey, now, rl.keyRecords) {
			return false
		}
	}

	return true
}

// getClientIP extracts the client IP from various sources
func getClientIP(c *gin.Context) string {
	// Check X-Forwarded-For header first (for proxy/load balancer scenarios)
	if forwarded := c.GetHeader("X-Forwarded-For"); forwarded != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if idx := strings.Index(forwarded, ","); idx != -1 {
			return strings.TrimSpace(forwarded[:idx])
		}
		return strings.TrimSpace(forwarded)
	}

	// Check X-Real-IP header
	if realIP := c.GetHeader("X-Real-IP"); realIP != "" {
		return strings.TrimSpace(realIP)
	}

	// Fall back to Gin's ClientIP method
	return c.ClientIP()
}
