package firehose

import (
	"bytes"
	"fmt"
	"io"

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-car"
	"github.com/watzon/lining/post"
)

// RepoOperation represents an operation on a repository
type RepoOperation struct {
	Action string // create, update, delete
	Path   string // record path
	Cid    string // content identifier
	Record []byte // raw record data
	Blocks []byte // CAR format blocks
}

// DecodeRecord attempts to decode the raw record into a specific type
func (op *RepoOperation) DecodeRecord(target any) error {
	if op.Blocks == nil {
		return fmt.Errorf("no blocks data available to decode")
	}

	if op.Cid == "" {
		return fmt.Errorf("no CID available for record")
	}

	// Parse the CID
	recordCid, err := cid.Parse(op.Cid)
	if err != nil {
		return fmt.Errorf("invalid CID: %w", err)
	}

	// Create a CAR reader
	cr, err := car.NewCarReader(bytes.NewReader(op.Blocks))
	if err != nil {
		return fmt.Errorf("failed to create CAR reader: %w", err)
	}

	// Read blocks until we find the one we want
	for {
		block, err := cr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading block: %w", err)
		}

		if block.Cid().Equals(recordCid) {
			// Found our block, decode it
			if v, ok := target.(cborer); ok {
				return v.UnmarshalCBOR(bytes.NewReader(block.RawData()))
			}
			return fmt.Errorf("target must implement UnmarshalCBOR")
		}
	}

	return fmt.Errorf("block not found in CAR data")
}

// cborer is an interface for types that can be unmarshaled from CBOR
type cborer interface {
	UnmarshalCBOR(io.Reader) error
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

// PostFilter is a function that filters Post events
type PostFilter func(post *post.Post) bool

// HandleFilter is a function that filters Handle events
type HandleFilter func(evt *HandleEvent) bool

// InfoFilter is a function that filters Info events
type InfoFilter func(evt *InfoEvent) bool

// MigrateFilter is a function that filters Migrate events
type MigrateFilter func(evt *MigrateEvent) bool

// TombstoneFilter is a function that filters Tombstone events
type TombstoneFilter func(evt *TombstoneEvent) bool

// OnPostHandler handles post events with optional filters
type OnPostHandler struct {
	Filters []PostFilter
	Handler func(post *post.Post) error
}

// OnHandleHandler handles handle events with optional filters
type OnHandleHandler struct {
	Filters []HandleFilter
	Handler func(evt *HandleEvent) error
}

// OnInfoHandler handles info events with optional filters
type OnInfoHandler struct {
	Filters []InfoFilter
	Handler func(evt *InfoEvent) error
}

// OnMigrateHandler handles migrate events with optional filters
type OnMigrateHandler struct {
	Filters []MigrateFilter
	Handler func(evt *MigrateEvent) error
}

// OnTombstoneHandler handles tombstone events with optional filters
type OnTombstoneHandler struct {
	Filters []TombstoneFilter
	Handler func(evt *TombstoneEvent) error
}

// FirehoseHandler represents a generic handler for firehose events
type FirehoseHandler interface {
	HandleRawOperation(op *RepoOperation) error
}

// EnhancedFirehoseCallbacks extends FirehoseCallbacks with additional functionality
type EnhancedFirehoseCallbacks struct {
	*FirehoseCallbacks
	Handlers          []FirehoseHandler
	PostHandlers      []*OnPostHandler
	HandleHandlers    []*OnHandleHandler
	InfoHandlers      []*OnInfoHandler
	MigrateHandlers   []*OnMigrateHandler
	TombstoneHandlers []*OnTombstoneHandler
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
