package firehose

import (
	"context"
	"fmt"
	"strings"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/watzon/lining/post"
)

// EnhancedFirehose extends the base Firehose with additional functionality
type EnhancedFirehose struct {
	*Firehose
}

// NewEnhancedFirehose creates a new EnhancedFirehose instance
func NewEnhancedFirehose(auth AuthProvider) *EnhancedFirehose {
	return &EnhancedFirehose{
		Firehose: NewFirehose(auth),
	}
}

// Subscribe subscribes to the Bluesky firehose with enhanced functionality
func (f *EnhancedFirehose) Subscribe(ctx context.Context, callbacks *EnhancedFirehoseCallbacks) error {
	if callbacks == nil {
		callbacks = &EnhancedFirehoseCallbacks{}
	}

	baseCallbacks := &FirehoseCallbacks{
		OnCommit: func(evt *CommitEvent) error {
			for _, op := range evt.Ops {
				// Process through raw handlers
				for _, handler := range callbacks.Handlers {
					if err := handler.HandleRawOperation(&op); err != nil {
						return err
					}
				}

				// Handle posts if we have any post handlers
				if len(callbacks.PostHandlers) > 0 && op.Action == "create" && strings.HasPrefix(op.Path, "app.bsky.feed.post") {
					post, err := PostFromCommitEvent(*evt)
					if err != nil {
						return fmt.Errorf("failed to convert post: %w", err)
					}

					// Process through all post handlers
					for _, handler := range callbacks.PostHandlers {
						// Apply post filters
						shouldProcess := true
						for _, filter := range handler.Filters {
							if !filter(post) {
								shouldProcess = false
								break
							}
						}

						if shouldProcess {
							if err := handler.Handler(post); err != nil {
								return err
							}
						}
					}
				}
			}
			return nil
		},
		OnHandle: func(evt *HandleEvent) error {
			for _, handler := range callbacks.HandleHandlers {
				shouldProcess := true
				for _, filter := range handler.Filters {
					if !filter(evt) {
						shouldProcess = false
						break
					}
				}
				if shouldProcess {
					if err := handler.Handler(evt); err != nil {
						return err
					}
				}
			}
			return nil
		},
		OnInfo: func(evt *InfoEvent) error {
			for _, handler := range callbacks.InfoHandlers {
				shouldProcess := true
				for _, filter := range handler.Filters {
					if !filter(evt) {
						shouldProcess = false
						break
					}
				}
				if shouldProcess {
					if err := handler.Handler(evt); err != nil {
						return err
					}
				}
			}
			return nil
		},
		OnMigrate: func(evt *MigrateEvent) error {
			for _, handler := range callbacks.MigrateHandlers {
				shouldProcess := true
				for _, filter := range handler.Filters {
					if !filter(evt) {
						shouldProcess = false
						break
					}
				}
				if shouldProcess {
					if err := handler.Handler(evt); err != nil {
						return err
					}
				}
			}
			return nil
		},
		OnTombstone: func(evt *TombstoneEvent) error {
			for _, handler := range callbacks.TombstoneHandlers {
				shouldProcess := true
				for _, filter := range handler.Filters {
					if !filter(evt) {
						shouldProcess = false
						break
					}
				}
				if shouldProcess {
					if err := handler.Handler(evt); err != nil {
						return err
					}
				}
			}
			return nil
		},
	}

	return f.Firehose.Subscribe(ctx, baseCallbacks)
}

// PostFromCommitEvent converts a CommitEvent to a Post
func PostFromCommitEvent(evt CommitEvent) (*post.Post, error) {
	var p bsky.FeedPost
	if err := evt.Ops[0].DecodeRecord(&p); err != nil {
		return nil, fmt.Errorf("failed to decode post: %w", err)
	}

	// Extract the Rkey from the op path
	rkey := evt.Ops[0].Path[strings.LastIndex(evt.Ops[0].Path, "/")+1:]

	newPost, err := post.PostFromFeedPost(&p, evt.Repo, rkey)
	if err != nil {
		return nil, fmt.Errorf("failed to convert post: %w", err)
	}

	return newPost, nil
}
