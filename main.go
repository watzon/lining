package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/watzon/lining/client"
)

func main() {
	// Create a new client with default config
	cfg := client.DefaultConfig()
	cfg.Debug = true // Enable debug logging

	// Set your Bluesky credentials
	cfg.Handle = "watzon.tech"
	cfg.APIKey = "aplc-i3ph-epbj-xpn7"

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
	callbacks := &client.FirehoseCallbacks{
		OnCommit: func(evt *client.CommitEvent) error {
			fmt.Printf("Commit from repo: %s\n", evt.Repo)
			for _, op := range evt.Ops {
				fmt.Printf(" - %s record %s\n", op.Action, op.Path)
			}
			return nil
		},
		OnHandle: func(evt *client.HandleEvent) error {
			fmt.Printf("Handle change: %s -> %s\n", evt.Did, evt.Handle)
			return nil
		},
		OnInfo: func(evt *client.InfoEvent) error {
			fmt.Printf("Repo info: name=%s, message=%s\n", evt.Name, evt.Message)
			return nil
		},
		OnMigrate: func(evt *client.MigrateEvent) error {
			fmt.Printf("Repo migrate: %s -> %s\n", evt.Did, evt.MigrateTo)
			return nil
		},
		OnTombstone: func(evt *client.TombstoneEvent) error {
			fmt.Printf("Repo tombstone: %s (time: %s)\n", evt.Did, evt.Time)
			return nil
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
