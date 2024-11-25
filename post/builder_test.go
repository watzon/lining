package post

import (
	"testing"

	lexutil "github.com/bluesky-social/indigo/lex/util"
	"github.com/stretchr/testify/assert"
	"github.com/watzon/lining/models"
)

func TestBuilder(t *testing.T) {
	t.Run("creates empty post", func(t *testing.T) {
		post, err := NewBuilder().Build()
		assert.NoError(t, err)
		assert.Empty(t, post.Text)
		assert.Empty(t, post.Facets)
	})

	t.Run("creates post with text", func(t *testing.T) {
		post, err := NewBuilder().AddText("Hello world").Build()
		assert.NoError(t, err)
		assert.Equal(t, "Hello world", post.Text)
		assert.Empty(t, post.Facets)
	})

	t.Run("handles mentions", func(t *testing.T) {
		post, err := NewBuilder().
			AddText("Hello ").
			AddMention("alice", "did:plc:alice").
			AddText("!").
			Build()

		assert.NoError(t, err)
		assert.Equal(t, "Hello @alice!", post.Text)
		assert.Len(t, post.Facets, 1)

		facet := post.Facets[0]
		assert.Equal(t, int64(6), facet.Index.ByteStart) // After "Hello "
		assert.Equal(t, int64(12), facet.Index.ByteEnd)  // Length of "@alice"
		assert.NotNil(t, facet.Features[0].RichtextFacet_Mention)
		assert.Equal(t, "did:plc:alice", facet.Features[0].RichtextFacet_Mention.Did)
	})

	t.Run("handles single hashtags", func(t *testing.T) {
		tests := []struct {
			name     string
			input    string
			wantText string
			wantTag  string
		}{
			{
				name:     "with hash prefix",
				input:    "#golang",
				wantText: "#golang",
				wantTag:  "golang",
			},
			{
				name:     "without hash prefix",
				input:    "golang",
				wantText: "#golang",
				wantTag:  "golang",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				post, err := NewBuilder().AddTag(tt.input).Build()
				assert.NoError(t, err)
				assert.Equal(t, tt.wantText, post.Text)
				assert.Len(t, post.Facets, 1)

				facet := post.Facets[0]
				assert.Equal(t, int64(0), facet.Index.ByteStart)
				assert.Equal(t, int64(len(tt.wantText)), facet.Index.ByteEnd)
				assert.NotNil(t, facet.Features[0].RichtextFacet_Tag)
				assert.Equal(t, tt.wantTag, facet.Features[0].RichtextFacet_Tag.Tag)
			})
		}
	})

	t.Run("handles double hashtags", func(t *testing.T) {
		tests := []struct {
			name     string
			input    string
			wantText string
			wantTag  string
		}{
			{
				name:     "with double hash prefix",
				input:    "##meta",
				wantText: "##meta",
				wantTag:  "meta",
			},
			{
				name:     "with single hash prefix",
				input:    "#meta",
				wantText: "#meta",
				wantTag:  "meta",
			},
			{
				name:     "without hash prefix",
				input:    "meta",
				wantText: "#meta",
				wantTag:  "meta",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				post, err := NewBuilder().AddTag(tt.input).Build()
				assert.NoError(t, err)
				assert.Equal(t, tt.wantText, post.Text)
				assert.Len(t, post.Facets, 1)

				facet := post.Facets[0]
				assert.Equal(t, int64(0), facet.Index.ByteStart)
				assert.Equal(t, int64(len(tt.wantText)), facet.Index.ByteEnd)
				assert.NotNil(t, facet.Features[0].RichtextFacet_Tag)
				assert.Equal(t, tt.wantTag, facet.Features[0].RichtextFacet_Tag.Tag)
			})
		}
	})

	t.Run("handles links", func(t *testing.T) {
		t.Run("with custom text", func(t *testing.T) {
			post, err := NewBuilder().
				AddLink("click here", "https://example.com").
				Build()

			assert.NoError(t, err)
			assert.Equal(t, "click here", post.Text)
			assert.Len(t, post.Facets, 1)

			facet := post.Facets[0]
			assert.Equal(t, int64(0), facet.Index.ByteStart)
			assert.Equal(t, int64(10), facet.Index.ByteEnd) // Length of "click here"
			assert.NotNil(t, facet.Features[0].RichtextFacet_Link)
			assert.Equal(t, "https://example.com", facet.Features[0].RichtextFacet_Link.Uri)
		})

		t.Run("with URL as text", func(t *testing.T) {
			url := "https://example.com"
			post, err := NewBuilder().
				AddURLLink(url).
				Build()

			assert.NoError(t, err)
			assert.Equal(t, url, post.Text)
			assert.Len(t, post.Facets, 1)

			facet := post.Facets[0]
			assert.Equal(t, int64(0), facet.Index.ByteStart)
			assert.Equal(t, int64(len(url)), facet.Index.ByteEnd)
			assert.NotNil(t, facet.Features[0].RichtextFacet_Link)
			assert.Equal(t, url, facet.Features[0].RichtextFacet_Link.Uri)
		})
	})

	t.Run("handles multiple facets", func(t *testing.T) {
		post, err := NewBuilder().
			AddText("Hello ").
			AddMention("alice", "did:plc:alice").
			AddText("! Check out ").
			AddLink("this link", "https://example.com").
			AddText(" about ").
			AddTag("#golang").
			Build()

		assert.NoError(t, err)
		assert.Equal(t, "Hello @alice! Check out this link about #golang", post.Text)
		assert.Len(t, post.Facets, 3)

		// Check mention
		assert.NotNil(t, post.Facets[0].Features[0].RichtextFacet_Mention)
		assert.Equal(t, "did:plc:alice", post.Facets[0].Features[0].RichtextFacet_Mention.Did)

		// Check link
		assert.NotNil(t, post.Facets[1].Features[0].RichtextFacet_Link)
		assert.Equal(t, "https://example.com", post.Facets[1].Features[0].RichtextFacet_Link.Uri)

		// Check hashtag
		assert.NotNil(t, post.Facets[2].Features[0].RichtextFacet_Tag)
		assert.Equal(t, "golang", post.Facets[2].Features[0].RichtextFacet_Tag.Tag)
	})

	t.Run("handles spaces and newlines", func(t *testing.T) {
		post, err := NewBuilder().
			AddText("Line 1").
			AddNewLine().
			AddText("Line 2").
			AddSpace().
			AddText("continued").
			Build()

		assert.NoError(t, err)
		assert.Equal(t, "Line 1\nLine 2 continued", post.Text)
		assert.Empty(t, post.Facets)
	})

	t.Run("validation", func(t *testing.T) {
		t.Run("post length", func(t *testing.T) {
			// Create a string that exceeds maxPostLength
			longText := make([]byte, maxPostLength+1)
			for i := range longText {
				longText[i] = 'a'
			}

			post, err := NewBuilder().
				AddText(string(longText)).
				Build()

			assert.Error(t, err)
			assert.Equal(t, ErrPostTooLong, err)
			assert.Empty(t, post.Text)
		})

		t.Run("invalid mention", func(t *testing.T) {
			tests := []struct {
				name     string
				username string
			}{
				{
					name:     "empty username",
					username: "",
				},
				{
					name:     "username with spaces",
					username: "user name",
				},
				{
					name:     "username with @",
					username: "@username",
				},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					post, err := NewBuilder().
						AddMention(tt.username, "did:plc:test").
						Build()

					assert.Error(t, err)
					assert.Equal(t, ErrInvalidMention, err)
					assert.Empty(t, post.Text)
				})
			}
		})

		t.Run("invalid tag", func(t *testing.T) {
			tests := []struct {
				name string
				tag  string
			}{
				{
					name: "empty tag",
					tag:  "",
				},
				{
					name: "tag with spaces",
					tag:  "tag with spaces",
				},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					post, err := NewBuilder().
						AddTag(tt.tag).
						Build()

					assert.Error(t, err)
					assert.Equal(t, ErrInvalidTag, err)
					assert.Empty(t, post.Text)
				})
			}
		})

		t.Run("invalid URL", func(t *testing.T) {
			tests := []struct {
				name string
				url  string
			}{
				{
					name: "empty URL",
					url:  "",
				},
				{
					name: "invalid scheme",
					url:  "not-a-url",
				},
				{
					name: "missing host",
					url:  "http://",
				},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					t.Run("AddLink", func(t *testing.T) {
						post, err := NewBuilder().
							AddLink("click here", tt.url).
							Build()

						assert.Error(t, err)
						assert.Equal(t, ErrInvalidURL, err)
						assert.Empty(t, post.Text)
					})

					t.Run("AddURLLink", func(t *testing.T) {
						post, err := NewBuilder().
							AddURLLink(tt.url).
							Build()

						assert.Error(t, err)
						assert.Equal(t, ErrInvalidURL, err)
						assert.Empty(t, post.Text)
					})
				})
			}
		})

		t.Run("mismatched images", func(t *testing.T) {
			post, err := NewBuilder().
				WithImages([]lexutil.LexBlob{{}}, []models.Image{}).
				Build()

			assert.Error(t, err)
			assert.Equal(t, ErrMismatchedImages, err)
			assert.Empty(t, post.Text)
		})

		t.Run("error propagation", func(t *testing.T) {
			// Test that once an error occurs, subsequent operations are skipped
			post, err := NewBuilder().
				AddText("Hello ").
				AddMention("invalid user", "did:plc:test"). // This will fail
				AddText(" and more text").                  // This should be skipped
				Build()

			assert.Error(t, err)
			assert.Equal(t, ErrInvalidMention, err)
			assert.Empty(t, post.Text)
		})
	})
}
