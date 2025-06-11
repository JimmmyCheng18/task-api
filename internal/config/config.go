package config

import (
	"log"
	"os"
	"strconv"
)

// Config holds application configuration
type Config struct {
	Port            string `json:"port"`
	Host            string `json:"host"`
	Environment     string `json:"environment"`
	ShutdownTimeout int    `json:"shutdown_timeout"`
	ReadTimeout     int    `json:"read_timeout"`
	WriteTimeout    int    `json:"write_timeout"`
	IdleTimeout     int    `json:"idle_timeout"`
	AllowedOrigins  string `json:"allowed_origins"`
	MaxTasks        int    `json:"max_tasks"`

	// Rate limiting configuration
	RateLimitEnabled     bool `json:"rate_limit_enabled"`
	RateLimitPerIP       int  `json:"rate_limit_per_ip"`       // Requests per minute per IP
	RateLimitPerAPIKey   int  `json:"rate_limit_per_api_key"`  // Requests per minute per API key
	RateLimitCleanupTime int  `json:"rate_limit_cleanup_time"` // Cleanup interval in minutes
}

// LoadConfig loads configuration from environment variables with defaults
func LoadConfig() *Config {
	config := &Config{
		Port:            getEnv("PORT", "8080"),
		Host:            getEnv("HOST", "0.0.0.0"),
		Environment:     getEnv("GIN_MODE", "release"),
		ShutdownTimeout: getEnvAsInt("SHUTDOWN_TIMEOUT", 30),
		ReadTimeout:     getEnvAsInt("READ_TIMEOUT", 60),
		WriteTimeout:    getEnvAsInt("WRITE_TIMEOUT", 60),
		IdleTimeout:     getEnvAsInt("IDLE_TIMEOUT", 120),
		AllowedOrigins:  getEnv("ALLOWED_ORIGINS", "*"),
		MaxTasks:        getEnvAsInt("MAX_TASKS", 10000),

		// Rate limiting defaults
		RateLimitEnabled:     getEnvAsBool("RATE_LIMIT_ENABLED", true),
		RateLimitPerIP:       getEnvAsInt("RATE_LIMIT_PER_IP", 100),       // 100 requests per minute per IP
		RateLimitPerAPIKey:   getEnvAsInt("RATE_LIMIT_PER_API_KEY", 1000), // 1000 requests per minute per API key
		RateLimitCleanupTime: getEnvAsInt("RATE_LIMIT_CLEANUP_TIME", 5),   // Cleanup every 5 minutes
	}

	return config
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets environment variable as integer with default value
func getEnvAsInt(key string, defaultValue int) int {
	if valueStr := os.Getenv(key); valueStr != "" {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		}
		log.Printf("Invalid integer value for %s: %s, using default: %d", key, valueStr, defaultValue)
	}
	return defaultValue
}

// getEnvAsBool gets environment variable as boolean with default value
func getEnvAsBool(key string, defaultValue bool) bool {
	if valueStr := os.Getenv(key); valueStr != "" {
		if value, err := strconv.ParseBool(valueStr); err == nil {
			return value
		}
		log.Printf("Invalid boolean value for %s: %s, using default: %t", key, valueStr, defaultValue)
	}
	return defaultValue
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Environment == "debug" || c.Environment == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Environment == "release" || c.Environment == "production"
}

// GetServerAddress returns the full server address
func (c *Config) GetServerAddress() string {
	return c.Host + ":" + c.Port
}

// GetRateLimitEnabled returns whether rate limiting is enabled
func (c *Config) GetRateLimitEnabled() bool {
	return c.RateLimitEnabled
}

// GetRateLimitPerIP returns the rate limit per IP
func (c *Config) GetRateLimitPerIP() int {
	return c.RateLimitPerIP
}

// GetRateLimitPerAPIKey returns the rate limit per API key
func (c *Config) GetRateLimitPerAPIKey() int {
	return c.RateLimitPerAPIKey
}

// GetRateLimitCleanupTime returns the rate limit cleanup time
func (c *Config) GetRateLimitCleanupTime() int {
	return c.RateLimitCleanupTime
}
