package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/bluesky-social/indigo/api/atproto"
	appbsky "github.com/bluesky-social/indigo/api/bsky"
	lexutil "github.com/bluesky-social/indigo/lex/util"
	"github.com/bluesky-social/indigo/xrpc"
	"golang.org/x/time/rate"

	"github.com/watzon/lining/models"
	"github.com/watzon/lining/post"
)

// Client interface defines the methods that a Bluesky client must implement
type Client interface {
	Connect(ctx context.Context) error
	PostToFeed(ctx context.Context, post appbsky.FeedPost) (string, string, error)
	UploadImage(ctx context.Context, image models.Image) (*lexutil.LexBlob, error)
	UploadImageFromURL(ctx context.Context, title string, imageURL string) (*lexutil.LexBlob, error)
	UploadImageFromFile(ctx context.Context, title string, filePath string) (*lexutil.LexBlob, error)
	UploadImages(ctx context.Context, images ...models.Image) ([]lexutil.LexBlob, error)
	GetProfile(ctx context.Context, handle string) (*appbsky.ActorDefs_ProfileViewDetailed, error)
	Follow(ctx context.Context, did string) error
	Unfollow(ctx context.Context, did string) error
}

// BskyClient implements the Client interface
type BskyClient struct {
	cfg     *Config
	client  *xrpc.Client
	limiter *rate.Limiter
	mu      sync.RWMutex
}

// NewClient creates a new Bluesky client with the given configuration
func NewClient(cfg *Config) (*BskyClient, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	// Validate config
	if cfg.Handle == "" || cfg.APIKey == "" {
		return nil, fmt.Errorf("handle and API key are required")
	}

	// Create rate limiter
	limiter := rate.NewLimiter(rate.Limit(cfg.RequestsPerMinute)/60, cfg.BurstSize)

	// Create HTTP client with proper configuration
	httpClient := &http.Client{
		Timeout: cfg.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:    cfg.MaxIdleConns,
			IdleConnTimeout: cfg.IdleConnTimeout,
		},
	}

	return &BskyClient{
		cfg: cfg,
		client: &xrpc.Client{
			Client: httpClient,
			Host:   cfg.ServerURL,
		},
		limiter: limiter,
	}, nil
}

// Connect authenticates with the Bluesky server
func (c *BskyClient) Connect(ctx context.Context) error {
	if err := c.limiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limit exceeded: %w", err)
	}

	input := &atproto.ServerCreateSession_Input{
		Identifier: c.cfg.Handle,
		Password:   c.cfg.APIKey,
	}

	session, err := atproto.ServerCreateSession(ctx, c.client, input)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	c.mu.Lock()
	c.client.Auth = &xrpc.AuthInfo{
		AccessJwt:  session.AccessJwt,
		RefreshJwt: session.RefreshJwt,
		Handle:     session.Handle,
		Did:        session.Did,
	}
	c.mu.Unlock()

	return nil
}

// GetProfile fetches a user's profile
func (c *BskyClient) GetProfile(ctx context.Context, handle string) (*appbsky.ActorDefs_ProfileViewDetailed, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	profile, err := appbsky.ActorGetProfile(ctx, c.client, handle)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	return profile, nil
}

// Follow follows a user by their DID
func (c *BskyClient) Follow(ctx context.Context, did string) error {
	if err := c.limiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limit exceeded: %w", err)
	}

	follow := &appbsky.GraphFollow{
		LexiconTypeID: "app.bsky.graph.follow",
		Subject:       did,
		CreatedAt:     time.Now().Format(time.RFC3339),
	}

	input := &atproto.RepoCreateRecord_Input{
		Collection: "app.bsky.graph.follow",
		Repo:       c.client.Auth.Did,
		Record:     &lexutil.LexiconTypeDecoder{Val: follow},
	}

	_, err := atproto.RepoCreateRecord(ctx, c.client, input)
	if err != nil {
		return fmt.Errorf("failed to follow user: %w", err)
	}

	return nil
}

// Unfollow unfollows a user by their DID
func (c *BskyClient) Unfollow(ctx context.Context, did string) error {
	if err := c.limiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limit exceeded: %w", err)
	}

	// First, find the follow record
	records, err := atproto.RepoListRecords(ctx, c.client, "app.bsky.graph.follow", "", 100, c.client.Auth.Did, false, "", "")
	if err != nil {
		return fmt.Errorf("failed to list follow records: %w", err)
	}

	var rkey string
	for _, record := range records.Records {
		follow, ok := record.Value.Val.(*appbsky.GraphFollow)
		if ok && follow.Subject == did {
			rkey = record.Uri[strings.LastIndex(record.Uri, "/")+1:]
			break
		}
	}

	if rkey == "" {
		return fmt.Errorf("follow record not found")
	}

	// Delete the follow record
	input := &atproto.RepoDeleteRecord_Input{
		Collection: "app.bsky.graph.follow",
		Repo:       c.client.Auth.Did,
		Rkey:       rkey,
	}

	err = atproto.RepoDeleteRecord(ctx, c.client, input)
	if err != nil {
		return fmt.Errorf("failed to unfollow user: %w", err)
	}

	return nil
}

// UploadImage uploads an image to Bluesky
func (c *BskyClient) UploadImage(ctx context.Context, image models.Image) (*lexutil.LexBlob, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	resp, err := atproto.RepoUploadBlob(ctx, c.client, bytes.NewReader(image.Data))
	if err != nil {
		return nil, fmt.Errorf("failed to upload blob: %w", err)
	}

	blob := &lexutil.LexBlob{
		Ref:      resp.Blob.Ref,
		MimeType: resp.Blob.MimeType,
		Size:     resp.Blob.Size,
	}

	return blob, nil
}

// UploadImageFromURL uploads an image from a URL to Bluesky
func (c *BskyClient) UploadImageFromURL(ctx context.Context, title string, imageURL string) (*lexutil.LexBlob, error) {
	// Create a client with reasonable timeouts
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Fetch image
	resp, err := client.Get(imageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch image: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch image: HTTP %d", resp.StatusCode)
	}

	// Read response body
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read image data: %w", err)
	}

	return c.UploadImage(ctx, models.Image{
		Title: title,
		Data:  data,
	})
}

// UploadImageFromFile uploads an image from a local file to Bluesky
func (c *BskyClient) UploadImageFromFile(ctx context.Context, title string, filePath string) (*lexutil.LexBlob, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return c.UploadImage(ctx, models.Image{
		Title: title,
		Data:  data,
	})
}

// UploadImages uploads multiple images to Bluesky
func (c *BskyClient) UploadImages(ctx context.Context, images ...models.Image) ([]lexutil.LexBlob, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	var blobs []lexutil.LexBlob
	for _, img := range images {
		blob, err := c.UploadImage(ctx, img)
		if err != nil {
			return nil, fmt.Errorf("failed to upload image %s: %w", img.Title, err)
		}
		blobs = append(blobs, *blob)
	}

	return blobs, nil
}

// PostToFeed creates a new post in the user's feed
func (c *BskyClient) PostToFeed(ctx context.Context, post appbsky.FeedPost) (string, string, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return "", "", fmt.Errorf("rate limit exceeded: %w", err)
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.client == nil || c.client.Auth == nil {
		return "", "", fmt.Errorf("client not connected")
	}

	// Create a new post object
	newPost := &appbsky.FeedPost{
		LexiconTypeID: "app.bsky.feed.post",
		Text:          post.Text,
		CreatedAt:     time.Now().Format(time.RFC3339),
		Embed:         post.Embed,
	}

	resp, err := atproto.RepoCreateRecord(ctx, c.client, &atproto.RepoCreateRecord_Input{
		Collection: "app.bsky.feed.post",
		Repo:       c.client.Auth.Did,
		Record:     &lexutil.LexiconTypeDecoder{Val: newPost},
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to create post: %w", err)
	}

	return resp.Cid, resp.Uri, nil
}

// NewPostBuilder creates a new post builder with the specified options
func NewPostBuilder(opts ...post.BuilderOption) *post.Builder {
	return post.NewBuilder(opts...)
}
