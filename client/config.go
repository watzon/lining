package client

import (
	"strconv"
	"time"
)

// Config holds all configuration for the Bluesky bot client
type Config struct {
	// API configuration
	Handle    string
	APIKey    string
	ServerURL string
	UserAgent string

	// HTTP client configuration
	Timeout         time.Duration
	RetryAttempts   int
	RetryWaitTime   time.Duration
	MaxIdleConns    int
	IdleConnTimeout time.Duration

	// Rate limiting
	RequestsPerMinute int
	BurstSize         int

	// Firehose configuration
	FirehoseURL      string
	FirehoseReconnectDelay time.Duration
	FirehoseBufferSize     int

	// Logging
	Debug bool
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		ServerURL:         "https://bsky.social",
		Timeout:           30 * time.Second,
		RetryAttempts:     3,
		RetryWaitTime:     time.Second,
		MaxIdleConns:      10,
		IdleConnTimeout:   120 * time.Second,
		RequestsPerMinute: 60,
		BurstSize:         5,
		FirehoseURL:       "wss://bsky.network/xrpc/com.atproto.sync.subscribeRepos",
		FirehoseReconnectDelay: 5 * time.Second,
		FirehoseBufferSize:     1000,
		Debug:             false,
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

// WithServerURL sets the server URL and returns the config
func (c *Config) WithServerURL(serverURL string) *Config {
	c.ServerURL = serverURL
	return c
}

// WithUserAgent sets the user agent and returns the config
func (c *Config) WithUserAgent(userAgent string) *Config {
	c.UserAgent = userAgent
	return c
}

// WithTimeout sets the timeout and returns the config
func (c *Config) WithTimeout(timeout time.Duration) *Config {
	c.Timeout = timeout
	return c
}

// WithRetryAttempts sets the retry attempts and returns the config
func (c *Config) WithRetryAttempts(retryAttempts int) *Config {
	c.RetryAttempts = retryAttempts
	return c
}

// WithRetryWaitTime sets the retry wait time and returns the config
func (c *Config) WithRetryWaitTime(retryWaitTime time.Duration) *Config {
	c.RetryWaitTime = retryWaitTime
	return c
}

// WithMaxIdleConns sets the max idle connections and returns the config
func (c *Config) WithMaxIdleConns(maxIdleConns int) *Config {
	c.MaxIdleConns = maxIdleConns
	return c
}

// WithIdleConnTimeout sets the idle connection timeout and returns the config
func (c *Config) WithIdleConnTimeout(idleConnTimeout time.Duration) *Config {
	c.IdleConnTimeout = idleConnTimeout
	return c
}

// WithRequestsPerMinute sets the requests per minute and returns the config
func (c *Config) WithRequestsPerMinute(requestsPerMinute int) *Config {
	c.RequestsPerMinute = requestsPerMinute
	return c
}

// WithBurstSize sets the burst size and returns the config
func (c *Config) WithBurstSize(burstSize int) *Config {
	c.BurstSize = burstSize
	return c
}

// WithFirehoseURL sets the firehose URL and returns the config
func (c *Config) WithFirehoseURL(url string) *Config {
	c.FirehoseURL = url
	return c
}

// WithFirehoseReconnectDelay sets the firehose reconnect delay and returns the config
func (c *Config) WithFirehoseReconnectDelay(delay time.Duration) *Config {
	c.FirehoseReconnectDelay = delay
	return c
}

// WithFirehoseBufferSize sets the firehose buffer size and returns the config
func (c *Config) WithFirehoseBufferSize(size int) *Config {
	c.FirehoseBufferSize = size
	return c
}

// WithDebug sets the debug mode and returns the config
func (c *Config) WithDebug(debug bool) *Config {
	c.Debug = debug
	return c
}

func (c *Config) String() string {
	debug := "false"
	if c.Debug {
		debug = "true"
	}

	return "Config{" +
		"Handle: " + c.Handle + ", " +
		"APIKey: " + c.APIKey + ", " +
		"ServerURL: " + c.ServerURL + ", " +
		"UserAgent: " + c.UserAgent + ", " +
		"Timeout: " + c.Timeout.String() + ", " +
		"RetryAttempts: " + strconv.Itoa(c.RetryAttempts) + ", " +
		"RetryWaitTime: " + c.RetryWaitTime.String() + ", " +
		"MaxIdleConns: " + strconv.Itoa(c.MaxIdleConns) + ", " +
		"IdleConnTimeout: " + c.IdleConnTimeout.String() + ", " +
		"RequestsPerMinute: " + strconv.Itoa(c.RequestsPerMinute) + ", " +
		"BurstSize: " + strconv.Itoa(c.BurstSize) + ", " +
		"FirehoseURL: " + c.FirehoseURL + ", " +
		"FirehoseReconnectDelay: " + c.FirehoseReconnectDelay.String() + ", " +
		"FirehoseBufferSize: " + strconv.Itoa(c.FirehoseBufferSize) + ", " +
		"Debug: " + debug +
		"}"
}
