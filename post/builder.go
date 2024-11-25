package post

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bluesky-social/indigo/api/bsky"
	lexutil "github.com/bluesky-social/indigo/lex/util"
	"github.com/watzon/lining/models"
)

// Builder constructs a post
type Builder struct {
	Text   string
	Facets []models.Facet
	Embed  models.Embed
}

// NewBuilder creates a simple post with text
func NewBuilder(text string) *Builder {
	return &Builder{
		Text:   text,
		Facets: []models.Facet{},
	}
}

// WithFacet adds a rich text feature to the post
func (b *Builder) WithFacet(ftype models.FacetType, value string, text string) *Builder {
	b.Facets = append(b.Facets, models.Facet{
		Type:  ftype,
		Value: value,
		Text:  text,
	})
	return b
}

// WithExternalLink adds an external link to the post
func (b *Builder) WithExternalLink(link models.Link) *Builder {
	b.Embed.Link = link
	return b
}

// WithImages adds images to the post
func (b *Builder) WithImages(blobs []lexutil.LexBlob, images []models.Image) *Builder {
	b.Embed.Images = images
	b.Embed.UploadedImages = blobs
	return b
}

// Build creates the final post
func (b *Builder) Build() (bsky.FeedPost, error) {
	post := bsky.FeedPost{
		Text:          b.Text,
		LexiconTypeID: "app.bsky.feed.post",
		CreatedAt:     time.Now().Format(time.RFC3339),
	}

	// Convert facets
	facets := []*bsky.RichtextFacet{}
	for _, f := range b.Facets {
		facet := &bsky.RichtextFacet{}
		features := []*bsky.RichtextFacet_Features_Elem{}
		feature := &bsky.RichtextFacet_Features_Elem{}

		switch f.Type {
		case models.FacetLink:
			feature = &bsky.RichtextFacet_Features_Elem{
				RichtextFacet_Link: &bsky.RichtextFacet_Link{
					LexiconTypeID: f.Type.String(),
					Uri:           f.Value,
				},
			}
		case models.FacetMention:
			feature = &bsky.RichtextFacet_Features_Elem{
				RichtextFacet_Mention: &bsky.RichtextFacet_Mention{
					LexiconTypeID: f.Type.String(),
					Did:           f.Value,
				},
			}
		case models.FacetTag:
			feature = &bsky.RichtextFacet_Features_Elem{
				RichtextFacet_Tag: &bsky.RichtextFacet_Tag{
					LexiconTypeID: f.Type.String(),
					Tag:           f.Value,
				},
			}
		}

		features = append(features, feature)
		facet.Features = features

		byteStart, byteEnd, err := findSubstring(post.Text, f.Text)
		if err != nil {
			return post, fmt.Errorf("unable to find the substring: %v , %w", f.Text, err)
		}

		facet.Index = &bsky.RichtextFacet_ByteSlice{
			ByteStart: int64(byteStart),
			ByteEnd:   int64(byteEnd),
		}

		facets = append(facets, facet)
	}
	post.Facets = facets

	// Handle embeds
	if b.Embed.Link != (models.Link{}) {
		post.Embed = &bsky.FeedPost_Embed{
			EmbedExternal: &bsky.EmbedExternal{
				LexiconTypeID: "app.bsky.embed.external",
				External: &bsky.EmbedExternal_External{
					Title:       b.Embed.Link.Title,
					Uri:         b.Embed.Link.Uri.String(),
					Description: b.Embed.Link.Description,
					Thumb:       &b.Embed.Link.Thumb,
				},
			},
		}
	} else if len(b.Embed.Images) > 0 && len(b.Embed.Images) == len(b.Embed.UploadedImages) {
		images := make([]*bsky.EmbedImages_Image, len(b.Embed.Images))
		for i, img := range b.Embed.Images {
			images[i] = &bsky.EmbedImages_Image{
				Alt:   img.Title,
				Image: &b.Embed.UploadedImages[i],
			}
		}

		post.Embed = &bsky.FeedPost_Embed{
			EmbedImages: &bsky.EmbedImages{
				LexiconTypeID: "app.bsky.embed.images",
				Images:        images,
			},
		}
	}

	return post, nil
}

// findSubstring finds the byte indices of a substring within a string
func findSubstring(s, substr string) (int, int, error) {
	start := strings.Index(s, substr)
	if start == -1 {
		return 0, 0, errors.New("substring not found")
	}
	return start, start + len(substr), nil
}
