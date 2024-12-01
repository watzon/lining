package interaction

import "time"

// Interaction represents a base interaction type
type Interaction struct {
	Actor     string    // The DID of the user performing the action
	Subject   string    // The DID or URI being acted upon
	CreatedAt time.Time // When the interaction occurred
}

// Follow represents a follow interaction
type Follow struct {
	Interaction
}

// Like represents a like interaction
type Like struct {
	Interaction
	Uri string // The URI of the post being liked
}

// Repost represents a repost interaction
type Repost struct {
	Interaction
	Uri string // The URI of the post being reposted
}

// Comment represents a comment interaction
type Comment struct {
	Interaction
	Uri     string // The URI of the comment
	ReplyTo string // The URI of the post being replied to
	Text    string // The content of the comment
}

// Filter functions for each interaction type
type FollowFilter func(*Follow) bool
type LikeFilter func(*Like) bool
type RepostFilter func(*Repost) bool
type CommentFilter func(*Comment) bool

// Handler functions for each interaction type
type FollowHandler func(*Follow) error
type LikeHandler func(*Like) error
type RepostHandler func(*Repost) error
type CommentHandler func(*Comment) error

// HandlerWithFilter combines a handler with its filters
type FollowHandlerWithFilter struct {
	Handler FollowHandler
	Filters []FollowFilter
}

type LikeHandlerWithFilter struct {
	Handler LikeHandler
	Filters []LikeFilter
}

type RepostHandlerWithFilter struct {
	Handler RepostHandler
	Filters []RepostFilter
}

type CommentHandlerWithFilter struct {
	Handler CommentHandler
	Filters []CommentFilter
}
