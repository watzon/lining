package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	cfg := Default()
	assert.Equal(t, "https://bsky.social", cfg.ServerURL)
	assert.Equal(t, 30*time.Second, cfg.Timeout)
	assert.Equal(t, 3, cfg.RetryAttempts)
	assert.Equal(t, time.Second, cfg.RetryWaitTime)
	assert.Equal(t, 10, cfg.MaxIdleConns)
	assert.Equal(t, 120*time.Second, cfg.IdleConnTimeout)
	assert.Equal(t, 60, cfg.RequestsPerMinute)
	assert.Equal(t, 5, cfg.BurstSize)
	assert.False(t, cfg.Debug)
}

func TestConfigChaining(t *testing.T) {
	cfg := Default().
		WithHandle("test.bsky.social").
		WithServerURL("https://example.com").
		WithUserAgent("TestBot/1.0").
		WithTimeout(60 * time.Second).
		WithRetryAttempts(5).
		WithRetryWaitTime(2 * time.Second).
		WithMaxIdleConns(20).
		WithIdleConnTimeout(240 * time.Second).
		WithRequestsPerMinute(120).
		WithBurstSize(10).
		WithDebug(true)

	assert.Equal(t, "test.bsky.social", cfg.Handle)
	assert.Equal(t, "https://example.com", cfg.ServerURL)
	assert.Equal(t, "TestBot/1.0", cfg.UserAgent)
	assert.Equal(t, 60*time.Second, cfg.Timeout)
	assert.Equal(t, 5, cfg.RetryAttempts)
	assert.Equal(t, 2*time.Second, cfg.RetryWaitTime)
	assert.Equal(t, 20, cfg.MaxIdleConns)
	assert.Equal(t, 240*time.Second, cfg.IdleConnTimeout)
	assert.Equal(t, 120, cfg.RequestsPerMinute)
	assert.Equal(t, 10, cfg.BurstSize)
	assert.True(t, cfg.Debug)
}

func TestConfigString(t *testing.T) {
	cfg := Default().
		WithHandle("test.bsky.social").
		WithServerURL("https://example.com").
		WithUserAgent("TestBot/1.0")

	str := cfg.String()
	assert.Contains(t, str, "test.bsky.social")
	assert.Contains(t, str, "https://example.com")
	assert.Contains(t, str, "TestBot/1.0")
	assert.Contains(t, str, "RetryAttempts: 3")
}
