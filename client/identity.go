package client

import (
	"context"
	"time"

	"github.com/bluesky-social/indigo/atproto/identity"
	"github.com/bluesky-social/indigo/atproto/syntax"
	// api "github.com/bluesky-social/indigo/api/atproto"
)

// newIdentityCache creates a new identity cache
// TODO: make this configurable
func newIdentityCache() *identity.CacheDirectory {
	dir := identity.DefaultDirectory()
	cache := identity.NewCacheDirectory(dir, 100_000, 30*time.Minute, 5*time.Minute, 5*time.Minute)
	return &cache
}

// ResolveDID resolves a DID to an identity, using a cache to avoid repeated lookups
func (c *BskyClient) ResolveDID(ctx context.Context, did string) (*identity.Identity, error) {
	if err := c.ensureValidSession(ctx); err != nil {
		return nil, err
	}

	// Not in cache or expired, need to resolve
	identity, err := c.cache.LookupDID(ctx, syntax.DID(did))
	if err != nil {
		return nil, err
	}

	return identity, nil
}

// ResolveHandle resolves a handle to an identity, using a cache to avoid repeated lookups
func (c *BskyClient) ResolveHandle(ctx context.Context, handle string) (*identity.Identity, error) {
	if err := c.ensureValidSession(ctx); err != nil {
		return nil, err
	}

	// Not in cache or expired, need to resolve
	identity, err := c.cache.LookupHandle(ctx, syntax.Handle(handle))
	if err != nil {
		return nil, err
	}

	return identity, nil
}

// GetDIDForHandle resolves a handle (e.g., "user.bsky.social") to its DID
// (Decentralized Identifier). The result is cached to improve performance
// of future lookups.
//
// Example:
//
//	did, err := client.GetDIDForHandle(ctx, "user.bsky.social")
//	// did will be something like "did:plc:xyz123..."
func (c *BskyClient) GetDIDForHandle(ctx context.Context, handle string) (string, error) {
	identity, err := c.ResolveHandle(ctx, handle)
	if err != nil {
		return "", err
	}

	return string(identity.DID), nil
}

// GetHandleForDID resolves a DID to its current handle. The result is cached
// to improve performance of future lookups.
//
// Example:
//
//	handle, err := client.GetHandleForDID(ctx, "did:plc:xyz123...")
//	// handle will be something like "user.bsky.social"
func (c *BskyClient) GetHandleForDID(ctx context.Context, did string) (string, error) {
	identity, err := c.ResolveDID(ctx, did)
	if err != nil {
		return "", err
	}

	return string(identity.Handle), nil
}
