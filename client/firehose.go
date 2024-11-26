package client

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/events"
	"github.com/bluesky-social/indigo/events/schedulers/sequential"
	"github.com/gorilla/websocket"
)

// RepoOperation represents an operation on a repository
type RepoOperation struct {
	Action string // create, update, delete
	Path   string // record path
}

// CommitEvent represents a commit to a repository
type CommitEvent struct {
	Repo string          // repository DID
	Time string          // timestamp
	Ops  []RepoOperation // operations performed
}

// HandleEvent represents a handle change event
type HandleEvent struct {
	Did    string // DID of the account
	Handle string // new handle
}

// InfoEvent represents repository information
type InfoEvent struct {
	Name    string // name of the event
	Message string // info message, may be empty
}

// MigrateEvent represents a repository migration
type MigrateEvent struct {
	Did       string // DID being migrated
	MigrateTo string // destination, may be empty
}

// TombstoneEvent represents a repository being tombstoned
type TombstoneEvent struct {
	Did  string // DID being tombstoned
	Time string // when it was tombstoned
}

// FirehoseCallbacks defines callbacks for different firehose events
type FirehoseCallbacks struct {
	// OnCommit is called when a repository commit occurs
	OnCommit func(evt *CommitEvent) error

	// OnHandle is called when a handle change occurs
	OnHandle func(evt *HandleEvent) error

	// OnInfo is called when repository information is received
	OnInfo func(evt *InfoEvent) error

	// OnMigrate is called when a repository migration occurs
	OnMigrate func(evt *MigrateEvent) error

	// OnTombstone is called when a repository is tombstoned
	OnTombstone func(evt *TombstoneEvent) error
}

// SubscribeToFirehose subscribes to the Bluesky firehose
func (c *BskyClient) SubscribeToFirehose(ctx context.Context, callbacks *FirehoseCallbacks) error {
	if callbacks == nil {
		return fmt.Errorf("callbacks cannot be nil")
	}

	// Create WebSocket connection
	dialer := websocket.Dialer{
		HandshakeTimeout: c.cfg.Timeout,
	}

	headers := http.Header{}
	if c.client.Auth != nil {
		headers.Set("Authorization", "Bearer "+c.client.Auth.AccessJwt)
	}

	conn, _, err := dialer.DialContext(ctx, c.cfg.FirehoseURL, headers)
	if err != nil {
		return fmt.Errorf("failed to connect to firehose: %w", err)
	}

	c.mu.Lock()
	c.wsConn = conn
	c.mu.Unlock()

	// Create repo stream callbacks that convert Indigo types to our types
	rsc := &events.RepoStreamCallbacks{
		RepoCommit: func(evt *atproto.SyncSubscribeRepos_Commit) error {
			if callbacks.OnCommit == nil {
				return nil
			}

			ops := make([]RepoOperation, len(evt.Ops))
			for i, op := range evt.Ops {
				ops[i] = RepoOperation{
					Action: op.Action,
					Path:   op.Path,
				}
			}

			return callbacks.OnCommit(&CommitEvent{
				Repo: evt.Repo,
				Time: evt.Time,
				Ops:  ops,
			})
		},
		RepoHandle: func(evt *atproto.SyncSubscribeRepos_Handle) error {
			if callbacks.OnHandle == nil {
				return nil
			}
			return callbacks.OnHandle(&HandleEvent{
				Did:    evt.Did,
				Handle: evt.Handle,
			})
		},
		RepoInfo: func(evt *atproto.SyncSubscribeRepos_Info) error {
			if callbacks.OnInfo == nil {
				return nil
			}
			message := ""
			if evt.Message != nil {
				message = *evt.Message
			}
			return callbacks.OnInfo(&InfoEvent{
				Name:    evt.Name,
				Message: message,
			})
		},
		RepoMigrate: func(evt *atproto.SyncSubscribeRepos_Migrate) error {
			if callbacks.OnMigrate == nil {
				return nil
			}
			migrateTo := ""
			if evt.MigrateTo != nil {
				migrateTo = *evt.MigrateTo
			}
			return callbacks.OnMigrate(&MigrateEvent{
				Did:       evt.Did,
				MigrateTo: migrateTo,
			})
		},
		RepoTombstone: func(evt *atproto.SyncSubscribeRepos_Tombstone) error {
			if callbacks.OnTombstone == nil {
				return nil
			}
			return callbacks.OnTombstone(&TombstoneEvent{
				Did:  evt.Did,
				Time: evt.Time,
			})
		},
	}

	// Create sequential scheduler
	sched := sequential.NewScheduler("bskyfirehose", rsc.EventHandler)

	// Start handling the repo stream
	go func() {
		if err := events.HandleRepoStream(ctx, conn, sched); err != nil {
			if c.cfg.Debug {
				fmt.Printf("firehose error: %v\n", err)
			}
			// Attempt to reconnect after delay
			time.Sleep(c.cfg.FirehoseReconnectDelay)
			if err := c.SubscribeToFirehose(ctx, callbacks); err != nil && c.cfg.Debug {
				fmt.Printf("reconnection failed: %v\n", err)
			}
		}
	}()

	return nil
}

// CloseFirehose closes the firehose connection
func (c *BskyClient) CloseFirehose() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.wsConn != nil {
		err := c.wsConn.Close()
		c.wsConn = nil
		return err
	}
	return nil
}
