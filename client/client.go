// Package client provides a high-level interface to interact with the Bluesky social network.
// It handles authentication, rate limiting, and provides methods for common operations
// like posting, retrieving posts, and managing the firehose subscription.
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
	"github.com/bluesky-social/indigo/atproto/identity"
	lexutil "github.com/bluesky-social/indigo/lex/util"
	"github.com/bluesky-social/indigo/xrpc"
	"golang.org/x/time/rate"

	"github.com/watzon/lining/config"
	"github.com/watzon/lining/firehose"
	"github.com/watzon/lining/models"
	"github.com/watzon/lining/post"
)

// BskyClient implements the main interface for interacting with Bluesky.
// It handles authentication, rate limiting, and provides methods for all
// supported Bluesky operations.
type BskyClient struct {
	cfg      *config.Config
	client   *xrpc.Client
	limiter  *rate.Limiter
	mu       sync.RWMutex
	cache    *identity.CacheDirectory
	firehose *firehose.EnhancedFirehose
}

// NewClient creates a new Bluesky client with the given configuration.
// The configuration must include at minimum a Handle and APIKey.
// If no configuration is provided, default values will be used.
//
// Example:
//
//	cfg := &config.Config{
//	    Handle: "user.bsky.social",
//	    APIKey: "your-api-key",
//	    ServerURL: "https://bsky.social",
//	}
//	client, err := NewClient(cfg)
func NewClient(cfg *config.Config) (*BskyClient, error) {
	if cfg == nil {
		cfg = config.Default()
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

	client := &BskyClient{
		cfg:     cfg,
		client:  &xrpc.Client{Client: httpClient, Host: cfg.ServerURL},
		limiter: limiter,
		cache:   newIdentityCache(),
	}

	return client, nil
}

// Connect establishes a connection to the Bluesky network and authenticates the user.
// This must be called before using any other methods that require authentication.
// The context can be used to cancel the connection attempt.
//
// Example:
//
//	ctx := context.Background()
//	if err := client.Connect(ctx); err != nil {
//	    log.Fatal("Failed to connect:", err)
//	}
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

// GetConfig returns the client's configuration
func (c *BskyClient) GetConfig() *config.Config {
	return c.cfg
}

// ensureValidSession checks if the current session is valid and refreshes it if necessary.
// This is called automatically by methods that require authentication.
func (c *BskyClient) ensureValidSession(ctx context.Context) error {
	c.mu.RLock()
	hasAuth := c.client.Auth != nil && c.client.Auth.AccessJwt != ""
	c.mu.RUnlock()

	if !hasAuth {
		return c.Connect(ctx)
	}

	// Try the refresh. If it fails, we'll attempt a full reconnect
	if err := c.RefreshSession(ctx); err != nil {
		return c.Connect(ctx)
	}

	return nil
}

// RefreshSession refreshes the access token using the refresh token.
// This is called automatically by ensureValidSession when needed, but
// you can call it manually if you want to force a refresh.
//
// Returns an error if the refresh fails or if there is no valid refresh token.
func (c *BskyClient) RefreshSession(ctx context.Context) error {
	c.mu.RLock()
	refreshJwt := c.client.Auth.RefreshJwt
	c.mu.RUnlock()

	if refreshJwt == "" {
		return fmt.Errorf("no refresh token available")
	}

	if err := c.limiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limit exceeded: %w", err)
	}

	session, err := atproto.ServerRefreshSession(ctx, c.client)
	if err != nil {
		return fmt.Errorf("failed to refresh session: %w", err)
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
	if err := c.ensureValidSession(ctx); err != nil {
		return nil, err
	}

	profile, err := appbsky.ActorGetProfile(ctx, c.client, handle)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	return profile, nil
}

// Follow follows a user by their DID
func (c *BskyClient) Follow(ctx context.Context, did string) error {
	if err := c.ensureValidSession(ctx); err != nil {
		return err
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
	if err := c.ensureValidSession(ctx); err != nil {
		return err
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

	_, err = atproto.RepoDeleteRecord(ctx, c.client, input)
	if err != nil {
		return fmt.Errorf("failed to unfollow user: %w", err)
	}

	return nil
}

// UploadImage uploads an image to Bluesky. The image data should be provided in the
// Image struct, which includes the raw bytes and metadata like title.
//
// The returned UploadedImage contains the blob reference needed for including
// the image in posts.
//
// Example:
//
//	img := models.Image{
//	    Title: "My Photo",
//	    Data:  imageBytes,
//	}
//	uploaded, err := client.UploadImage(ctx, img)
func (c *BskyClient) UploadImage(ctx context.Context, image models.Image) (*models.UploadedImage, error) {
	if err := c.ensureValidSession(ctx); err != nil {
		return nil, err
	}

	resp, err := atproto.RepoUploadBlob(ctx, c.client, bytes.NewReader(image.Data))
	if err != nil {
		return nil, fmt.Errorf("failed to upload blob: %w", err)
	}

	uploaded := &models.UploadedImage{
		LexBlob: &lexutil.LexBlob{
			Ref:      resp.Blob.Ref,
			MimeType: resp.Blob.MimeType,
			Size:     resp.Blob.Size,
		},
		Image: image,
	}

	return uploaded, nil
}

// UploadImageFromURL downloads an image from the given URL and uploads it to Bluesky.
// This is a convenience method that handles both downloading and uploading.
//
// The title parameter will be used as the alt text for the image.
//
// Example:
//
//	uploaded, err := client.UploadImageFromURL(ctx, "My Photo", "https://example.com/photo.jpg")
func (c *BskyClient) UploadImageFromURL(ctx context.Context, title string, imageURL string) (*models.UploadedImage, error) {
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

// UploadImageFromFile reads an image from the local filesystem and uploads it to Bluesky.
// This is a convenience method that handles both reading and uploading.
//
// The title parameter will be used as the alt text for the image.
//
// Example:
//
//	uploaded, err := client.UploadImageFromFile(ctx, "My Photo", "/path/to/photo.jpg")
func (c *BskyClient) UploadImageFromFile(ctx context.Context, title string, filePath string) (*models.UploadedImage, error) {
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
func (c *BskyClient) UploadImages(ctx context.Context, images ...models.Image) ([]*models.UploadedImage, error) {
	if err := c.ensureValidSession(ctx); err != nil {
		return nil, err
	}

	var uploads []*models.UploadedImage
	for _, img := range images {
		blob, err := c.UploadImage(ctx, img)
		if err != nil {
			return nil, fmt.Errorf("failed to upload image %s: %w", img.Title, err)
		}
		uploads = append(uploads, blob)
	}

	return uploads, nil
}

// PostToFeed creates a new post in the user's feed. The post parameter should be a
// fully constructed FeedPost object, which you can create using the post.Builder.
//
// Returns the CID (Content Identifier) and URI of the created post.
//
// Example:
//
//	post, _ := post.NewBuilder().
//	    AddText("Hello, Bluesky!").
//	    WithImages([]models.UploadedImage{*uploadedImage}).
//	    Build()
//
//	cid, uri, err := client.PostToFeed(ctx, post)
func (c *BskyClient) PostToFeed(ctx context.Context, post appbsky.FeedPost) (string, string, error) {
	if err := c.ensureValidSession(ctx); err != nil {
		return "", "", err
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
		Facets:        post.Facets,
		Entities:      post.Entities,
		Labels:        post.Labels,
		Langs:         post.Langs,
		Reply:         post.Reply,
		Tags:          post.Tags,
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
func (c *BskyClient) NewPostBuilder(opts ...post.BuilderOption) *post.Builder {
	// Add the client option first, then any user-provided options
	allOpts := append([]post.BuilderOption{
		post.WithClient(c.client),
	}, opts...)
	return post.NewBuilder(allOpts...)
}

// GetAccessToken returns the current access token
func (c *BskyClient) GetAccessToken() string {
	if c.client != nil && c.client.Auth != nil {
		return c.client.Auth.AccessJwt
	}
	return ""
}

// GetFirehoseURL returns the configured firehose URL
func (c *BskyClient) GetFirehoseURL() string {
	return c.cfg.FirehoseURL
}

// GetTimeout returns the configured timeout duration for API requests.
// This is used internally by the firehose and other long-running operations.
func (c *BskyClient) GetTimeout() time.Duration {
	return c.cfg.Timeout
}

// SubscribeToFirehose connects to the Bluesky firehose and starts processing events
// using the provided callbacks. The firehose provides a real-time stream of all
// public activities on the network.
//
// The callbacks parameter should define handlers for the types of events you're
// interested in (posts, likes, follows, etc.).
//
// Example:
//
//	callbacks := &firehose.EnhancedFirehoseCallbacks{
//	    PostHandlers: []*firehose.OnPostHandler{
//	        {
//	            Filters: []firehose.PostFilter{
//	                func(post *post.Post) bool {
//	                    return len(post.Embed.Images) > 0
//	                },
//	            },
//	            Handler: func(post *post.Post) error {
//	                fmt.Printf("New post with images: %s\n", post.Url())
//	                return nil
//	            },
//	        },
//	    },
//	}
//	err := client.SubscribeToFirehose(ctx, callbacks)
func (c *BskyClient) SubscribeToFirehose(ctx context.Context, callbacks *firehose.EnhancedFirehoseCallbacks) error {
	if c.firehose == nil {
		c.firehose = firehose.NewEnhancedFirehose(c)
	}
	return c.firehose.Subscribe(ctx, callbacks)
}

// CloseFirehose stops the firehose subscription and cleans up resources.
// It's important to call this when you're done with the firehose to prevent
// resource leaks.
func (c *BskyClient) CloseFirehose() error {
	if c.firehose != nil {
		return c.firehose.Close()
	}
	return nil
}

// DownloadBlob downloads a blob (like an image) from the Bluesky network using its CID and owner's DID.
// The CID (Content Identifier) can be found in several places:
//   - post.Embed.Images[].Ref for direct image embeds
//   - post.Embed.RecordWithMedia.Media.Images[].Ref for images in quoted posts
//   - post.Embed.External.ThumbRef for external link thumbnails
//
// The DID (Decentralized Identifier) can be found in:
//   - post.Repo for the post author's DID
//   - post.Record.Uri for quoted/embedded post DIDs
//
// It returns the raw blob data, the detected content type (e.g., "image/jpeg"),
// and any error that occurred during download.
//
// Example:
//
//	data, contentType, err := client.DownloadBlob(ctx, img.Ref, post.Repo)
//	if err != nil {
//	    return fmt.Errorf("failed to download: %w", err)
//	}
func (c *BskyClient) DownloadBlob(ctx context.Context, cid string, did string) ([]byte, string, error) {
	if err := c.ensureValidSession(ctx); err != nil {
		return nil, "", err
	}

	data, err := atproto.SyncGetBlob(ctx, c.client, cid, did)
	if err != nil {
		return nil, "", fmt.Errorf("failed to download blob: %w", err)
	}

	// Try to detect content type from the first few bytes
	contentType := http.DetectContentType(data)

	return data, contentType, nil
}
