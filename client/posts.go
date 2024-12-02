package client

import (
	"context"
	"fmt"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/watzon/lining/post"
)

// GetPost retrieves a single post by its URI. The URI should be in the format
// "at://did:plc:xyz/app.bsky.feed.post/timestamp".
//
// Example:
//
//	post, err := client.GetPost(ctx, "at://did:plc:xyz/app.bsky.feed.post/123")
func (c *BskyClient) GetPost(ctx context.Context, uri string) (*post.Post, error) {
	if err := c.ensureValidSession(ctx); err != nil {
		return nil, err
	}

	// Extract repo and rkey from URI
	repo, collection, rkey, err := post.ParsePostURI(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to parse post URI: %w", err)
	}

	// Use bsky.FeedGetPostThread to get the post
	resp, err := bsky.FeedGetPostThread(ctx, c.client, 0, 0, uri)
	if err != nil {
		fmt.Printf("Debug info - Collection: %s, Repo: %s, Rkey: %s\n", collection, repo, rkey)
		if xerr, ok := err.(interface{ Unwrap() error }); ok {
			fmt.Printf("Underlying error: %v\n", xerr.Unwrap())
		}
		return nil, fmt.Errorf("failed to get post (repo=%s rkey=%s): %w", repo, rkey, err)
	}

	if resp == nil || resp.Thread == nil || resp.Thread.FeedDefs_ThreadViewPost == nil {
		return nil, fmt.Errorf("got nil response or post data")
	}

	// Convert the post view to our Post type
	return post.PostFromFeedDefs_PostView(resp.Thread.FeedDefs_ThreadViewPost.Post)
}

// GetPosts retrieves multiple posts by their URIs. This is more efficient than
// making multiple GetPost calls when you need to fetch several posts at once.
//
// Example:
//
//	posts, err := client.GetPosts(ctx,
//	    "at://did:plc:xyz/app.bsky.feed.post/123",
//	    "at://did:plc:xyz/app.bsky.feed.post/456",
//	)
func (c *BskyClient) GetPosts(ctx context.Context, uris ...string) ([]*post.Post, error) {
	if err := c.ensureValidSession(ctx); err != nil {
		return nil, err
	}

	var posts []*post.Post
	for _, uri := range uris {
		p, err := c.GetPost(ctx, uri)
		if err != nil {
			return nil, fmt.Errorf("failed to get post %s: %w", uri, err)
		}
		posts = append(posts, p)
	}

	return posts, nil
}
