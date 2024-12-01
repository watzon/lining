package post

import (
	"fmt"
	"strings"

	"github.com/bluesky-social/indigo/api/bsky"
)

// Post represents a Bluesky post, but in a more user-friendly format
type Post struct {
	Repo string // Author DID
	Rkey string // Post ID

	// Counts
	Likes   int64
	Quotes  int64
	Replies int64
	Reposts int64

	// FeedPost stuff
	Text      string
	CreatedAt string
	Embed     *Embed
	Facets    []Facet
	Labels    []string
	Langs     []string
	Tags      []string

	// Reply
	ReplyUri string
	ReplyRef *bsky.FeedPost_ReplyRef
}

// Uri returns the AT URI for the post
func (p *Post) Uri() string {
	return fmt.Sprintf("at://%s/app.bsky.feed.post/%s", p.Repo, p.Rkey)
}

// Url returns the URL for the post
func (p *Post) Url() string {
	return fmt.Sprintf("https://bsky.app/profile/%s/post/%s", p.Repo, p.Rkey)
}

// String returns a string representation of the Post
func (p *Post) String() string {
	return fmt.Sprintf("Post{Text: %s, CreatedAt: %s, Embed: %v, Facets: %v, Labels: %s, Langs: %s, Reply: %v, Tags: %s}",
		p.Text, p.CreatedAt, p.Embed, p.Facets, p.Labels, p.Langs, p.ReplyRef, p.Tags)
}

// PostsFromGetPostsResponse converts a bsky.FeedGetPosts_Output to a slice of Post
func PostsFromGetPostsResponse(resp *bsky.FeedGetPosts_Output) (posts []*Post, err error) {
	for _, post := range resp.Posts {
		feedPost, ok := post.Record.Val.(*bsky.FeedPost)
		if !ok {
			return nil, fmt.Errorf("unexpected record type: %T", post.Record.Val)
		}

		// Extract labels
		var labels []string
		if feedPost.Labels != nil && feedPost.Labels.LabelDefs_SelfLabels != nil {
			for _, label := range feedPost.Labels.LabelDefs_SelfLabels.Values {
				labels = append(labels, label.Val)
			}
		}

		embed, err := ExtractEmbedFromFeedPost(feedPost)
		if err != nil {
			return nil, err
		}

		replyUri := ""
		if feedPost.Reply != nil && feedPost.Reply.Root != nil {
			replyUri = feedPost.Reply.Root.Uri
		}

		posts = append(posts, &Post{
			Text:      feedPost.Text,
			CreatedAt: feedPost.CreatedAt,
			Embed:     embed,
			Facets:    ExtractFacetsFromFeedPost(feedPost),
			Labels:    labels,
			Langs:     feedPost.Langs,
			Tags:      feedPost.Tags,
			ReplyUri:  replyUri,
			ReplyRef:  feedPost.Reply,
		})
	}
	return
}

// PostFromFeedDefs_PostView converts a bsky.FeedDefs_PostView to a Post
func PostFromFeedDefs_PostView(post *bsky.FeedDefs_PostView) (*Post, error) {
	feedPost, ok := post.Record.Val.(*bsky.FeedPost)
	if !ok {
		return nil, fmt.Errorf("unexpected record type: %T", post.Record.Val)
	}

	repo, _, rkey, err := ParsePostURI(post.Uri)
	if err != nil {
		return nil, err
	}

	extracted, err := PostFromFeedPost(feedPost, repo, rkey)
	if err != nil {
		return nil, err
	}

	extracted.Repo = repo
	extracted.Rkey = rkey

	return extracted, nil
}

// PostFromFeedPost converts a bsky.FeedPost to a Post
func PostFromFeedPost(post *bsky.FeedPost, repo, rkey string) (*Post, error) {
	var labels []string
	if post.Labels != nil && post.Labels.LabelDefs_SelfLabels != nil {
		for _, label := range post.Labels.LabelDefs_SelfLabels.Values {
			labels = append(labels, label.Val)
		}
	}

	embed, err := ExtractEmbedFromFeedPost(post)
	if err != nil {
		return nil, err
	}

	replyUri := ""
	if post.Reply != nil && post.Reply.Root != nil {
		replyUri = post.Reply.Root.Uri
	}

	return &Post{
		Repo:      repo,
		Rkey:      rkey,
		Text:      post.Text,
		CreatedAt: post.CreatedAt,
		Embed:     embed,
		Facets:    ExtractFacetsFromFeedPost(post),
		Labels:    labels,
		Langs:     post.Langs,
		Tags:      post.Tags,
		ReplyUri:  replyUri,
		ReplyRef:  post.Reply,
	}, nil
}

// ParsePostURI parses an AT URI into repo, collection and rkey
// Example URI: at://did:plc:xyz/app.bsky.feed.post/123
func ParsePostURI(uri string) (repo string, collection string, rkey string, err error) {
	// Remove the at:// prefix
	if !strings.HasPrefix(uri, "at://") {
		return "", "", "", fmt.Errorf("invalid URI prefix: %s", uri)
	}
	uri = strings.TrimPrefix(uri, "at://")

	// Split into parts
	parts := strings.Split(uri, "/")
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("invalid URI format: %s", uri)
	}

	repo = parts[0]
	collection = parts[1]
	rkey = parts[2]

	return repo, collection, rkey, nil
}
