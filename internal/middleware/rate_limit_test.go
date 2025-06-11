package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRateLimiter_Allow(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := RateLimitConfig{
		Enabled:         true,
		PerIP:           2, // Very low limit for testing
		PerAPIKey:       5,
		CleanupInterval: 1 * time.Minute,
		WindowSize:      1 * time.Minute,
	}

	// Create test router with rate limiting
	router := gin.New()
	router.Use(RateLimit(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	// Test IP-based rate limiting
	t.Run("IP Rate Limiting", func(t *testing.T) {
		// First request should succeed
		req1, _ := http.NewRequest("GET", "/test", nil)
		req1.Header.Set("X-Forwarded-For", "192.168.1.1")
		w1 := httptest.NewRecorder()
		router.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusOK, w1.Code)

		// Second request should succeed
		req2, _ := http.NewRequest("GET", "/test", nil)
		req2.Header.Set("X-Forwarded-For", "192.168.1.1")
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusOK, w2.Code)

		// Third request should be rate limited
		req3, _ := http.NewRequest("GET", "/test", nil)
		req3.Header.Set("X-Forwarded-For", "192.168.1.1")
		w3 := httptest.NewRecorder()
		router.ServeHTTP(w3, req3)
		assert.Equal(t, http.StatusTooManyRequests, w3.Code)
	})

	t.Run("Different IPs Should Have Separate Limits", func(t *testing.T) {
		// Request from different IP should succeed
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Forwarded-For", "192.168.1.2")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestSmartRateLimit_CustomLimits(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := RateLimitConfig{
		Enabled:         true,
		PerIP:           4, // Base limit
		PerAPIKey:       10,
		CleanupInterval: 1 * time.Minute,
		WindowSize:      1 * time.Minute,
	}

	// Create test router with smart rate limiting
	router := gin.New()
	router.Use(SmartRateLimit(config))

	// Health endpoint should have higher limit (4 * 5 = 20)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Regular GET endpoint should have base limit (4)
	router.GET("/api/data", func(c *gin.Context) {
		c.JSON(200, gin.H{"data": "test"})
	})

	// POST endpoint should have lower limit (4 / 2 = 2)
	router.POST("/api/data", func(c *gin.Context) {
		c.JSON(200, gin.H{"created": "test"})
	})

	t.Run("Health Endpoint Has Higher Limit", func(t *testing.T) {
		ip := "192.168.1.10"
		successCount := 0

		// Try many requests to health endpoint
		for i := 0; i < 10; i++ {
			req, _ := http.NewRequest("GET", "/health", nil)
			req.Header.Set("X-Forwarded-For", ip)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code == http.StatusOK {
				successCount++
			}
		}

		// Should allow more than base limit
		assert.Greater(t, successCount, 4)
	})

	t.Run("POST Has Lower Limit Than GET", func(t *testing.T) {
		ip := "192.168.1.11"

		// Test POST endpoint (lower limit)
		postSuccessCount := 0
		for i := 0; i < 5; i++ {
			req, _ := http.NewRequest("POST", "/api/data", nil)
			req.Header.Set("X-Forwarded-For", ip)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code == http.StatusOK {
				postSuccessCount++
			}
		}

		// Should be limited to 2 requests (4 / 2)
		assert.LessOrEqual(t, postSuccessCount, 2)
	})
}

func TestRateLimiter_APIKeyLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := RateLimitConfig{
		Enabled:         true,
		PerIP:           10, // High IP limit
		PerAPIKey:       2,  // Low API key limit
		CleanupInterval: 1 * time.Minute,
		WindowSize:      1 * time.Minute,
	}

	router := gin.New()
	router.Use(RateLimit(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	t.Run("API Key Rate Limiting", func(t *testing.T) {
		apiKey := "test-api-key"
		successCount := 0

		for i := 0; i < 5; i++ {
			req, _ := http.NewRequest("GET", "/test", nil)
			req.Header.Set("X-API-Key", apiKey)
			req.Header.Set("X-Forwarded-For", "192.168.1.20") // Different IP each time
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code == http.StatusOK {
				successCount++
			}
		}

		// Should be limited by API key limit (2), not IP limit (10)
		assert.Equal(t, 2, successCount)
	})
}

func TestRateLimiter_Cleanup(t *testing.T) {
	config := RateLimitConfig{
		Enabled:         true,
		PerIP:           1,
		PerAPIKey:       1,
		CleanupInterval: 100 * time.Millisecond, // Very short for testing
		WindowSize:      1 * time.Minute,
	}

	limiter := NewRateLimiter(config)
	defer limiter.Stop()

	// Create a mock gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("X-Forwarded-For", "192.168.1.30")

	// Make a request to populate records
	limiter.Allow(c)

	// Check that records exist
	limiter.mu.RLock()
	initialCount := len(limiter.ipRecords)
	limiter.mu.RUnlock()
	assert.Equal(t, 1, initialCount)

	// Wait for cleanup
	time.Sleep(150 * time.Millisecond)

	// Records should still exist (not expired yet)
	limiter.mu.RLock()
	currentCount := len(limiter.ipRecords)
	limiter.mu.RUnlock()
	assert.Equal(t, 1, currentCount)
}

func TestRateLimiter_GetStats(t *testing.T) {
	config := DefaultRateLimitConfig()
	limiter := NewRateLimiter(config)
	defer limiter.Stop()

	stats := limiter.GetStats()

	assert.Contains(t, stats, "config")
	assert.Contains(t, stats, "statistics")

	configStats := stats["config"].(map[string]interface{})
	assert.Equal(t, true, configStats["enabled"])
	assert.Equal(t, 100, configStats["per_ip"])
	assert.Equal(t, 1000, configStats["per_api_key"])

	statisticsStats := stats["statistics"].(map[string]interface{})
	assert.Equal(t, 0, statisticsStats["tracked_ips"])
	assert.Equal(t, 0, statisticsStats["tracked_api_keys"])
}

func TestRateLimiter_Disabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := RateLimitConfig{
		Enabled: false, // Disabled
	}

	router := gin.New()
	router.Use(RateLimit(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	// All requests should succeed when rate limiting is disabled
	for i := 0; i < 10; i++ {
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Forwarded-For", "192.168.1.40")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}
}
