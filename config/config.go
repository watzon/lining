package config

import (
	"time"
)

// Config holds all configuration for the Bluesky bot client
type Config struct {
	// API configuration
	Handle     string
	APIKey     string
	ServerURL  string
	UserAgent  string

	// HTTP client configuration
	Timeout           time.Duration
	RetryAttempts    int
	RetryWaitTime    time.Duration
	MaxIdleConns     int
	IdleConnTimeout  time.Duration

	// Rate limiting
	RequestsPerMinute int
	BurstSize        int

	// Logging
	Debug bool
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		ServerURL:         "https://bsky.social",
		Timeout:          30 * time.Second,
		RetryAttempts:    3,
		RetryWaitTime:    time.Second,
		MaxIdleConns:     10,
		IdleConnTimeout:  120 * time.Second,
		RequestsPerMinute: 60,
		BurstSize:        5,
		Debug:            false,
	}
}

// WithHandle sets the handle and returns the config
func (c *Config) WithHandle(handle string) *Config {
	c.Handle = handle
	return c
}

// WithAPIKey sets the API key and returns the config
func (c *Config) WithAPIKey(apiKey string) *Config {
	c.APIKey = apiKey
	return c
}
