package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORSConfig defines the configuration for CORS middleware
type CORSConfig struct {
	AllowOrigins     []string `json:"allow_origins"`     // Allowed origins
	AllowMethods     []string `json:"allow_methods"`     // Allowed HTTP methods
	AllowHeaders     []string `json:"allow_headers"`     // Allowed headers
	ExposeHeaders    []string `json:"expose_headers"`    // Headers to expose to client
	AllowCredentials bool     `json:"allow_credentials"` // Allow credentials
	MaxAge           int      `json:"max_age"`           // Preflight cache duration
}

// DefaultCORSConfig returns a default CORS configuration
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodHead,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Length",
			"Content-Type",
			"Authorization",
			"X-Requested-With",
			"Accept",
			"Accept-Encoding",
			"Accept-Language",
			"Connection",
			"Host",
			"Referer",
			"User-Agent",
		},
		ExposeHeaders: []string{
			"Content-Length",
			"X-Total-Count",
			"X-Offset",
			"X-Limit",
		},
		AllowCredentials: false,
		MaxAge:           86400, // 24 hours
	}
}

// CORS returns a CORS middleware with default configuration
func CORS() gin.HandlerFunc {
	return CORSWithConfig(DefaultCORSConfig())
}

// CORSWithConfig returns a CORS middleware with custom configuration
func CORSWithConfig(config CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Set Access-Control-Allow-Origin
		if len(config.AllowOrigins) == 1 && config.AllowOrigins[0] == "*" {
			c.Header("Access-Control-Allow-Origin", "*")
		} else if isOriginAllowed(origin, config.AllowOrigins) {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		// Set Access-Control-Allow-Methods
		if len(config.AllowMethods) > 0 {
			c.Header("Access-Control-Allow-Methods", joinStrings(config.AllowMethods, ", "))
		}

		// Set Access-Control-Allow-Headers
		if len(config.AllowHeaders) > 0 {
			c.Header("Access-Control-Allow-Headers", joinStrings(config.AllowHeaders, ", "))
		}

		// Set Access-Control-Expose-Headers
		if len(config.ExposeHeaders) > 0 {
			c.Header("Access-Control-Expose-Headers", joinStrings(config.ExposeHeaders, ", "))
		}

		// Set Access-Control-Allow-Credentials
		if config.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		// Set Access-Control-Max-Age for preflight requests
		if config.MaxAge > 0 {
			c.Header("Access-Control-Max-Age", formatInt(config.MaxAge))
		}

		// Handle preflight requests
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RestrictiveCORS returns a CORS middleware with restrictive configuration
// This is suitable for production environments
func RestrictiveCORS(allowedOrigins []string) gin.HandlerFunc {
	config := CORSConfig{
		AllowOrigins: allowedOrigins,
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Authorization",
			"Accept",
		},
		ExposeHeaders: []string{
			"Content-Length",
			"X-Total-Count",
		},
		AllowCredentials: true,
		MaxAge:           3600, // 1 hour
	}

	return CORSWithConfig(config)
}

// DevelopmentCORS returns a CORS middleware suitable for development
// This allows all origins and methods
func DevelopmentCORS() gin.HandlerFunc {
	config := CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: false,
		MaxAge:           86400,
	}

	return CORSWithConfig(config)
}

// Helper functions

// isOriginAllowed checks if the origin is in the allowed list
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
		// You could add wildcard matching here if needed
		// For example: *.example.com
	}
	return false
}

// joinStrings joins a slice of strings with a separator
func joinStrings(slice []string, sep string) string {
	if len(slice) == 0 {
		return ""
	}

	result := slice[0]
	for i := 1; i < len(slice); i++ {
		result += sep + slice[i]
	}
	return result
}

// formatInt converts an integer to string
func formatInt(i int) string {
	// Simple integer to string conversion
	if i == 0 {
		return "0"
	}

	var result string
	negative := i < 0
	if negative {
		i = -i
	}

	for i > 0 {
		result = string(rune('0'+(i%10))) + result
		i /= 10
	}

	if negative {
		result = "-" + result
	}

	return result
}

// SecurityHeaders adds common security headers
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent XSS attacks
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")

		// HSTS (only for HTTPS)
		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		// Content Security Policy (relaxed for Swagger UI)
		// Allow unsafe-inline for styles and scripts needed by Swagger UI
		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline' 'unsafe-eval'; " +
			"style-src 'self' 'unsafe-inline'; " +
			"img-src 'self' data: https:; " +
			"font-src 'self' data:"
		c.Header("Content-Security-Policy", csp)

		// Referrer Policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		c.Next()
	}
}
