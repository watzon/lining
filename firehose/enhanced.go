package firehose

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/watzon/lining/interaction"
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
				if len(callbacks.PostHandlers) > 0 && strings.HasPrefix(op.Path, "app.bsky.feed.post") {
					// Only try to convert to post if it's a create operation and has a CID
					if op.Action == "create" && op.Cid != "" {
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

						// Handle comments (replies) if this post is a reply
						if len(callbacks.CommentHandlers) > 0 && post.ReplyUri != "" {
							comment := &interaction.Comment{
								Interaction: interaction.Interaction{
									Actor:     evt.Repo,
									Subject:   op.Path,
									CreatedAt: time.Now(),
								},
								Uri:     post.Uri(),
								ReplyTo: post.ReplyUri,
								Text:    post.Text,
							}

							for _, handler := range callbacks.CommentHandlers {
								shouldProcess := true
								for _, filter := range handler.Filters {
									if !filter(comment) {
										shouldProcess = false
										break
									}
								}

								if shouldProcess {
									if err := handler.Handler(comment); err != nil {
										return err
									}
								}
							}
						}
					}
				}

				// Handle follows
				if len(callbacks.FollowHandlers) > 0 && strings.HasPrefix(op.Path, "app.bsky.graph.follow") {
					follow := &interaction.Follow{
						Interaction: interaction.Interaction{
							Actor:     evt.Repo,
							Subject:   op.Path,
							CreatedAt: time.Now(),
						},
					}

					for _, handler := range callbacks.FollowHandlers {
						shouldProcess := true
						for _, filter := range handler.Filters {
							if !filter(follow) {
								shouldProcess = false
								break
							}
						}

						if shouldProcess {
							if err := handler.Handler(follow); err != nil {
								return err
							}
						}
					}
				}

				// Handle likes
				if len(callbacks.LikeHandlers) > 0 && strings.HasPrefix(op.Path, "app.bsky.feed.like") && op.Action == "create" {
					like := &interaction.Like{
						Interaction: interaction.Interaction{
							Actor:     evt.Repo,
							Subject:   op.Path,
							CreatedAt: time.Now(),
						},
						Uri: op.Path,
					}

					for _, handler := range callbacks.LikeHandlers {
						shouldProcess := true
						for _, filter := range handler.Filters {
							if !filter(like) {
								shouldProcess = false
								break
							}
						}

						if shouldProcess {
							if err := handler.Handler(like); err != nil {
								return err
							}
						}
					}
				}

				// Handle reposts
				if len(callbacks.RepostHandlers) > 0 && strings.HasPrefix(op.Path, "app.bsky.feed.repost") && op.Action == "create" {
					repost := &interaction.Repost{
						Interaction: interaction.Interaction{
							Actor:     evt.Repo,
							Subject:   op.Path,
							CreatedAt: time.Now(),
						},
						Uri: op.Path,
					}

					for _, handler := range callbacks.RepostHandlers {
						shouldProcess := true
						for _, filter := range handler.Filters {
							if !filter(repost) {
								shouldProcess = false
								break
							}
						}

						if shouldProcess {
							if err := handler.Handler(repost); err != nil {
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
	if len(evt.Ops) == 0 {
		return nil, fmt.Errorf("no operations in commit event")
	}

	op := evt.Ops[0]
	if op.Cid == "" {
		return nil, fmt.Errorf("no CID available for record")
	}

	if op.Blocks == nil {
		return nil, fmt.Errorf("no blocks data available")
	}

	var p bsky.FeedPost
	if err := op.DecodeRecord(&p); err != nil {
		return nil, fmt.Errorf("failed to decode post: %w", err)
	}

	// Extract the Rkey from the op path
	rkey := op.Path[strings.LastIndex(op.Path, "/")+1:]

	newPost, err := post.PostFromFeedPost(&p, evt.Repo, rkey)
	if err != nil {
		return nil, fmt.Errorf("failed to convert post: %w", err)
	}

	return newPost, nil
}
