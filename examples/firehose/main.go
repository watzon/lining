package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/watzon/lining/client"
	"github.com/watzon/lining/config"
	"github.com/watzon/lining/firehose"
	"github.com/watzon/lining/interaction"
	"github.com/watzon/lining/post"
)

func main() {
	// Create a new client with default config
	cfg := config.Default()
	cfg.Debug = true // Enable debug logging

	// Set your Bluesky credentials
	cfg.Handle = os.Getenv("HANDLE")
	cfg.APIKey = os.Getenv("APIKEY")

	c, err := client.NewClient(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Create a context that we can cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Connect to Bluesky
	if err := c.Connect(ctx); err != nil {
		log.Fatal(err)
	}

	// Create a channel to handle signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create firehose callbacks
	callbacks := &firehose.EnhancedFirehoseCallbacks{
		// Simple post handler that prints out post text
		PostHandlers: []firehose.PostHandlerWithFilter{
			{
				Filters: []firehose.PostFilter{
					func(post *post.Post) bool {
						// Only process posts with text
						return post.Text != ""
					},
				},
				Handler: func(p *post.Post) error {
					fmt.Printf("New post from %s: %s\n", p.Repo, p.Text)
					return nil
				},
			},
		},

		// Handle follows
		FollowHandlers: []interaction.FollowHandlerWithFilter{
			{
				Handler: func(f *interaction.Follow) error {
					fmt.Printf("New follow from %s\n", f.Actor)
					return nil
				},
			},
		},

		// Handle likes
		LikeHandlers: []interaction.LikeHandlerWithFilter{
			{
				Handler: func(l *interaction.Like) error {
					fmt.Printf("New like from %s: %s\n", l.Actor, l.Uri)
					return nil
				},
			},
		},

		// Handle reposts
		RepostHandlers: []interaction.RepostHandlerWithFilter{
			{
				Handler: func(r *interaction.Repost) error {
					fmt.Printf("New repost from %s: %s\n", r.Actor, r.Uri)
					return nil
				},
			},
		},

		// Handle comments
		CommentHandlers: []interaction.CommentHandlerWithFilter{
			{
				Handler: func(c *interaction.Comment) error {
					fmt.Printf("New comment from %s on %s: %s\n", c.Actor, c.ReplyTo, c.Text)
					return nil
				},
			},
		},
	}

	// Subscribe to the firehose
	fmt.Println("Connecting to Bluesky firehose...")
	err = c.SubscribeToFirehose(ctx, callbacks)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected! Listening for events... (Press Ctrl+C to exit)")

	// Wait for interrupt signal
	<-sigChan
	fmt.Println("\nShutting down gracefully...")

	// Close the firehose connection
	if err := c.CloseFirehose(); err != nil {
		fmt.Printf("Error closing firehose: %v\n", err)
	}
}
