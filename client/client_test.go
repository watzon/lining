package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/watzon/lining/config"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &config.Config{
				Handle:            "test.bsky.social",
				APIKey:            "test-key",
				ServerURL:         "https://bsky.social",
				Timeout:           30 * time.Second,
				RequestsPerMinute: 60,
				BurstSize:         5,
			},
			wantErr: false,
		},
		{
			name: "missing handle",
			cfg: &config.Config{
				APIKey:            "test-key",
				ServerURL:         "https://bsky.social",
				Timeout:           30 * time.Second,
				RequestsPerMinute: 60,
				BurstSize:         5,
			},
			wantErr: true,
		},
		{
			name: "missing api key",
			cfg: &config.Config{
				Handle:            "test.bsky.social",
				ServerURL:         "https://bsky.social",
				Timeout:           30 * time.Second,
				RequestsPerMinute: 60,
				BurstSize:         5,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.cfg)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

func TestConnect(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/xrpc/com.atproto.server.createSession" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"accessJwt": "test-access-token",
				"refreshJwt": "test-refresh-token",
				"handle": "test.bsky.social",
				"did": "did:plc:test"
			}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	cfg := &config.Config{
		Handle:            "test.bsky.social",
		APIKey:            "test-key",
		ServerURL:         server.URL,
		Timeout:           30 * time.Second,
		RequestsPerMinute: 60,
		BurstSize:         5,
	}

	client, err := NewClient(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, client)

	err = client.Connect(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, client.client)
	assert.NotNil(t, client.client.Auth)
	assert.Equal(t, "test-access-token", client.client.Auth.AccessJwt)
	assert.Equal(t, "test-refresh-token", client.client.Auth.RefreshJwt)
}

func TestGetProfile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/xrpc/app.bsky.actor.getProfile" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"did": "did:plc:test",
				"handle": "test.bsky.social",
				"displayName": "Test User",
				"description": "Test description"
			}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	cfg := &config.Config{
		Handle:            "test.bsky.social",
		APIKey:            "test-key",
		ServerURL:         server.URL,
		Timeout:           30 * time.Second,
		RequestsPerMinute: 60,
		BurstSize:         5,
	}

	client, err := NewClient(cfg)
	assert.NoError(t, err)

	profile, err := client.GetProfile(context.Background(), "test.bsky.social")
	assert.NoError(t, err)
	assert.NotNil(t, profile)
	assert.Equal(t, "did:plc:test", profile.Did)
	assert.Equal(t, "test.bsky.social", profile.Handle)
	assert.Equal(t, "Test User", *profile.DisplayName)
	assert.Equal(t, "Test description", *profile.Description)
}
