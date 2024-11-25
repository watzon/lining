package post

import (
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/bluesky-social/indigo/api/bsky"
	lexutil "github.com/bluesky-social/indigo/lex/util"
	"github.com/watzon/lining/models"
)

// Maximum length for a Bluesky post
const maxPostLength = 300

// ErrEmptyText is returned when attempting to add empty text
var ErrEmptyText = errors.New("text cannot be empty")

// ErrInvalidURL is returned when a URL is not valid
var ErrInvalidURL = errors.New("invalid URL")

// ErrInvalidMention is returned when a mention is not valid
var ErrInvalidMention = errors.New("invalid mention format")

// ErrInvalidTag is returned when a tag is not valid
var ErrInvalidTag = errors.New("invalid tag format")

// ErrMismatchedImages is returned when images and blobs arrays have different lengths
var ErrMismatchedImages = errors.New("images and blobs arrays must have the same length")

// ErrPostTooLong is returned when the post exceeds the maximum length
var ErrPostTooLong = errors.New("post exceeds maximum length")

// Builder constructs a Bluesky post with rich text features (facets).
// It provides a fluent interface for building posts with mentions, hashtags,
// and links, while automatically handling the byte indexing required by
// Bluesky's richtext format.
//
// Example usage:
//
//	post, err := NewBuilder().
//	    AddText("Check out this post by ").
//	    AddMention("alice", "did:plc:alice").
//	    AddText("! More info at ").
//	    AddLink("this link", "https://example.com").
//	    AddText(" #").
//	    AddTag("bluesky").
//	    Build()
type Builder struct {
	segments []segment
	embed    models.Embed
	err      error
}

// segment represents a piece of text with an optional facet.
// When the facet is nil, the segment is treated as plain text.
type segment struct {
	text  string
	facet *models.Facet
}

// NewBuilder creates a new post builder, optionally with initial text.
// If text is provided, it will be added as the first segment of the post.
//
// Example:
//
//	// Empty builder
//	builder := NewBuilder()
//
//	// Builder with initial text
//	builder := NewBuilder("Hello world")
func NewBuilder(text ...string) *Builder {
	b := &Builder{
		segments: []segment{},
	}
	if len(text) > 0 {
		b.AddText(text[0])
	}
	return b
}

// AddText adds plain text to the post. This text will not have any
// rich text features associated with it.
func (b *Builder) AddText(text string) *Builder {
	if text == "" {
		return b
	}
	if err := b.validatePostLength(text); err != nil {
		b.err = err
		return b
	}
	b.segments = append(b.segments, segment{text: text})
	return b
}

func (b *Builder) validatePostLength(additionalText string) error {
	totalLength := len(additionalText)
	for _, seg := range b.segments {
		totalLength += len(seg.text)
	}
	if totalLength > maxPostLength {
		return ErrPostTooLong
	}
	return nil
}

func validateURL(uri string) error {
	if uri == "" {
		return ErrInvalidURL
	}
	u, err := url.Parse(uri)
	if err != nil {
		return ErrInvalidURL
	}
	if u.Scheme == "" || u.Host == "" {
		return ErrInvalidURL
	}
	return nil
}

func validateMention(username string) error {
	if username == "" {
		return ErrInvalidMention
	}
	// Basic username validation - can be expanded based on Bluesky's rules
	if strings.ContainsAny(username, " \t\n@") {
		return ErrInvalidMention
	}
	return nil
}

func validateTag(tag string) error {
	if tag == "" {
		return ErrInvalidTag
	}
	// Basic tag validation - can be expanded based on Bluesky's rules
	if strings.ContainsAny(tag, " \t\n") {
		return ErrInvalidTag
	}
	return nil
}

// AddFacet adds text with an associated rich text feature to the post.
// This is a low-level method - prefer using AddMention, AddTag, or AddLink
// for common facet types.
func (b *Builder) AddFacet(text string, ftype models.FacetType, value string) *Builder {
	b.segments = append(b.segments, segment{
		text: text,
		facet: &models.Facet{
			Type:  ftype,
			Value: value,
			Text:  text,
		},
	})
	return b
}

// AddMention adds a mention facet (@username) to the post.
// The username should be provided without the @ prefix, as it will be added automatically.
// The did parameter should be the Bluesky DID for the mentioned user.
//
// Example:
//
//	builder.AddMention("alice", "did:plc:alice")  // Adds "@alice" to the post
func (b *Builder) AddMention(username string, did string) *Builder {
	if b.err != nil {
		return b
	}
	if err := validateMention(username); err != nil {
		b.err = err
		return b
	}
	return b.AddFacet("@"+username, models.FacetMention, did)
}

// AddTag adds a hashtag facet to the post. The tag can be provided with or
// without the # prefix. For double hashtags (##), provide the full tag including
// both # characters.
//
// Example:
//
//	builder.AddTag("bluesky")    // Adds "#bluesky"
//	builder.AddTag("#golang")    // Also adds "#golang"
//	builder.AddTag("##meta")     // Adds "##meta"
func (b *Builder) AddTag(tag string) *Builder {
	if b.err != nil {
		return b
	}

	// Remove # prefix if present for validation
	tagToValidate := tag
	if len(tag) > 0 && tag[0] == '#' {
		tagToValidate = tag[1:]
	}

	if err := validateTag(tagToValidate); err != nil {
		b.err = err
		return b
	}

	// Handle double hash tags (##tag)
	if len(tag) > 1 && tag[0] == '#' && tag[1] == '#' {
		return b.AddFacet(tag, models.FacetTag, tag[2:])
	}

	// Remove single # if present
	if len(tag) > 0 && tag[0] == '#' {
		tag = tag[1:]
	}

	return b.AddFacet("#"+tag, models.FacetTag, tag)
}

// AddLink adds a link facet with custom display text to the post.
//
// Example:
//
//	builder.AddLink("click here", "https://example.com")
func (b *Builder) AddLink(text string, uri string) *Builder {
	if b.err != nil {
		return b
	}
	if err := validateURL(uri); err != nil {
		b.err = err
		return b
	}
	return b.AddFacet(text, models.FacetLink, uri)
}

// AddURLLink adds a link facet where the URL itself is used as the display text.
//
// Example:
//
//	builder.AddURLLink("https://example.com")
func (b *Builder) AddURLLink(uri string) *Builder {
	if b.err != nil {
		return b
	}
	if err := validateURL(uri); err != nil {
		b.err = err
		return b
	}
	return b.AddFacet(uri, models.FacetLink, uri)
}

// AddSpace adds a single space character to the post.
// This is a convenience method equivalent to AddText(" ").
func (b *Builder) AddSpace() *Builder {
	return b.AddText(" ")
}

// AddNewLine adds a newline character to the post.
// This is a convenience method equivalent to AddText("\n").
func (b *Builder) AddNewLine() *Builder {
	return b.AddText("\n")
}

// WithExternalLink adds an external link to the post. This link will be
// displayed as a card in the Bluesky interface, separate from any link
// facets in the text.
func (b *Builder) WithExternalLink(link models.Link) *Builder {
	b.embed.Link = link
	return b
}

// WithImages adds images to the post. The images will be displayed
// in a gallery format in the Bluesky interface.
//
// The blobs parameter should contain the already-uploaded image blobs,
// and the images parameter should contain the corresponding image metadata.
// Both slices must be the same length.
func (b *Builder) WithImages(blobs []lexutil.LexBlob, images []models.Image) *Builder {
	if b.err != nil {
		return b
	}
	if len(blobs) != len(images) {
		b.err = ErrMismatchedImages
		return b
	}
	b.embed.Images = images
	b.embed.UploadedImages = blobs
	return b
}

// Build creates the final Bluesky post, combining all the added text,
// facets, and embeds into a complete post structure.
func (b *Builder) Build() (bsky.FeedPost, error) {
	if b.err != nil {
		return bsky.FeedPost{}, b.err
	}
	var post bsky.FeedPost
	var text strings.Builder
	facets := []*bsky.RichtextFacet{}
	byteIndex := 0

	// Build text and facets together
	for _, seg := range b.segments {
		text.WriteString(seg.text)

		if seg.facet != nil {
			facet := &bsky.RichtextFacet{
				Index: &bsky.RichtextFacet_ByteSlice{
					ByteStart: int64(byteIndex),
					ByteEnd:   int64(byteIndex + len(seg.text)),
				},
				Features: []*bsky.RichtextFacet_Features_Elem{},
			}

			feature := &bsky.RichtextFacet_Features_Elem{}
			switch seg.facet.Type {
			case models.FacetLink:
				feature.RichtextFacet_Link = &bsky.RichtextFacet_Link{
					LexiconTypeID: seg.facet.Type.String(),
					Uri:           seg.facet.Value,
				}
			case models.FacetMention:
				feature.RichtextFacet_Mention = &bsky.RichtextFacet_Mention{
					LexiconTypeID: seg.facet.Type.String(),
					Did:           seg.facet.Value,
				}
			case models.FacetTag:
				feature.RichtextFacet_Tag = &bsky.RichtextFacet_Tag{
					LexiconTypeID: seg.facet.Type.String(),
					Tag:           seg.facet.Value,
				}
			}

			facet.Features = append(facet.Features, feature)
			facets = append(facets, facet)
		}

		byteIndex += len(seg.text)
	}

	post.Text = text.String()
	post.Facets = facets
	post.LexiconTypeID = "app.bsky.feed.post"
	post.CreatedAt = time.Now().Format(time.RFC3339)

	// Handle embeds
	if len(b.embed.Images) > 0 && len(b.embed.Images) == len(b.embed.UploadedImages) {
		images := make([]*bsky.EmbedImages_Image, len(b.embed.Images))
		for i, img := range b.embed.Images {
			images[i] = &bsky.EmbedImages_Image{
				Alt:   img.Title,
				Image: &b.embed.UploadedImages[i],
			}
		}

		post.Embed = &bsky.FeedPost_Embed{
			EmbedImages: &bsky.EmbedImages{
				LexiconTypeID: "app.bsky.embed.images",
				Images:        images,
			},
		}
	} else if b.embed.Link.Uri.String() != "" {
		post.Embed = &bsky.FeedPost_Embed{
			EmbedExternal: &bsky.EmbedExternal{
				LexiconTypeID: "app.bsky.embed.external",
				External: &bsky.EmbedExternal_External{
					Uri:         b.embed.Link.Uri.String(),
					Title:       b.embed.Link.Title,
					Description: b.embed.Link.Description,
					Thumb:       &b.embed.Link.Thumb,
				},
			},
		}
	}

	return post, nil
}
