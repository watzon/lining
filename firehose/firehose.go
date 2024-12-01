package firehose

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/events"
	"github.com/bluesky-social/indigo/events/schedulers/sequential"
	"github.com/gorilla/websocket"
)

// AuthProvider defines the minimal interface needed for firehose authentication
type AuthProvider interface {
	GetAccessToken() string
	GetFirehoseURL() string
	GetTimeout() time.Duration
}

// Firehose manages the connection to the Bluesky firehose
type Firehose struct {
	auth   AuthProvider
	wsConn *websocket.Conn
	mu     sync.RWMutex
}

// NewFirehose creates a new Firehose instance
func NewFirehose(auth AuthProvider) *Firehose {
	return &Firehose{
		auth: auth,
	}
}

// Subscribe subscribes to the Bluesky firehose
func (f *Firehose) Subscribe(ctx context.Context, callbacks *FirehoseCallbacks) error {
	if callbacks == nil {
		return fmt.Errorf("callbacks cannot be nil")
	}

	// Create WebSocket connection
	dialer := websocket.Dialer{
		HandshakeTimeout: f.auth.GetTimeout(),
	}

	headers := http.Header{}
	if token := f.auth.GetAccessToken(); token != "" {
		headers.Set("Authorization", "Bearer "+token)
	}

	conn, _, err := dialer.DialContext(ctx, f.auth.GetFirehoseURL(), headers)
	if err != nil {
		return fmt.Errorf("failed to connect to firehose: %w", err)
	}

	f.mu.Lock()
	f.wsConn = conn
	f.mu.Unlock()

	// Create repo stream callbacks that convert Indigo types to our types
	rsc := &events.RepoStreamCallbacks{
		RepoCommit: func(evt *atproto.SyncSubscribeRepos_Commit) error {
			if callbacks.OnCommit == nil {
				return nil
			}

			ops := make([]RepoOperation, len(evt.Ops))
			for i, op := range evt.Ops {
				cid := ""
				if op.Cid != nil {
					cid = op.Cid.String()
				}
				ops[i] = RepoOperation{
					Action: op.Action,
					Path:   op.Path,
					Cid:    cid,
					Blocks: evt.Blocks,
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
			if true {
				fmt.Printf("firehose error: %v\n", err)
			}
			// Attempt to reconnect after delay
			time.Sleep(5 * time.Second)
			if err := f.Subscribe(ctx, callbacks); err != nil && true {
				fmt.Printf("reconnection failed: %v\n", err)
			}
		}
	}()

	return nil
}

// Close closes the firehose connection
func (f *Firehose) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.wsConn != nil {
		err := f.wsConn.Close()
		f.wsConn = nil
		return err
	}
	return nil
}

// Deprecated: Use Firehose.Subscribe instead
func SubscribeToFirehose(ctx context.Context, auth AuthProvider, callbacks *FirehoseCallbacks) error {
	f := NewFirehose(auth)
	return f.Subscribe(ctx, callbacks)
}
