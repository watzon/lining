package post

import (
	"testing"

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
				WithImages([]models.UploadedImage{}).
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

func TestBuilderJoinStrategies(t *testing.T) {
	t.Run("JoinAsIs strategy", func(t *testing.T) {
		post, err := NewBuilder().
			AddText("Hello").
			AddText("world").
			AddText("!").
			Build()

		assert.NoError(t, err)
		assert.Equal(t, "Helloworld!", post.Text)
	})

	t.Run("JoinWithSpaces strategy", func(t *testing.T) {
		post, err := NewBuilder(WithJoinStrategy(JoinWithSpaces)).
			AddText("Hello").
			AddText("world").
			AddText("!").
			Build()

		assert.NoError(t, err)
		assert.Equal(t, "Hello world !", post.Text)
	})

	t.Run("JoinWithSpaces with facets", func(t *testing.T) {
		post, err := NewBuilder(WithJoinStrategy(JoinWithSpaces)).
			AddText("Hello").
			AddMention("alice", "did:plc:alice").
			AddText("!").
			AddText("Check out").
			AddLink("this", "https://example.com").
			Build()

		assert.NoError(t, err)
		assert.Equal(t, "Hello @alice ! Check out this", post.Text)
		assert.Len(t, post.Facets, 2)
	})

	t.Run("JoinWithSpaces with empty segments", func(t *testing.T) {
		post, err := NewBuilder(WithJoinStrategy(JoinWithSpaces)).
			AddText("Hello").
			AddText(""). // Empty segment
			AddText("world").
			Build()

		assert.NoError(t, err)
		assert.Equal(t, "Hello world", post.Text)
	})
}

func TestBuilderOptions(t *testing.T) {
	t.Run("default options", func(t *testing.T) {
		builder := NewBuilder()
		assert.Equal(t, JoinAsIs, builder.options.JoinStrategy)
	})

	t.Run("with join strategy option", func(t *testing.T) {
		builder := NewBuilder(WithJoinStrategy(JoinWithSpaces))
		assert.Equal(t, JoinWithSpaces, builder.options.JoinStrategy)
	})

	t.Run("multiple options (future-proofing)", func(t *testing.T) {
		// This test ensures our options system can handle multiple options
		// when we add more in the future
		builder := NewBuilder(
			WithJoinStrategy(JoinWithSpaces),
			// Add more options here as they're added
		)
		assert.Equal(t, JoinWithSpaces, builder.options.JoinStrategy)
	})
}

func TestBuilderAutoDetection(t *testing.T) {
	t.Run("auto hashtags", func(t *testing.T) {
		post, err := NewBuilder(WithAutoHashtag(true)).
			AddText("Check out #golang and #programming!").
			Build()

		assert.NoError(t, err)
		assert.Equal(t, "Check out #golang and #programming!", post.Text)
		assert.Len(t, post.Facets, 2)

		// Verify hashtags
		for _, facet := range post.Facets {
			assert.NotNil(t, facet.Features[0].RichtextFacet_Tag)
			tag := facet.Features[0].RichtextFacet_Tag.Tag
			assert.Contains(t, []string{"golang", "programming"}, tag)
		}
	})

	t.Run("auto mentions", func(t *testing.T) {
		post, err := NewBuilder(WithAutoMention(true)).
			AddText("Hello @alice and @bob!").
			Build()

		assert.NoError(t, err)
		assert.Equal(t, "Hello @alice and @bob!", post.Text)
		assert.Len(t, post.Facets, 2)

		// Verify mentions
		for _, facet := range post.Facets {
			assert.NotNil(t, facet.Features[0].RichtextFacet_Mention)
			did := facet.Features[0].RichtextFacet_Mention.Did
			assert.Contains(t, []string{"did:plc:alice", "did:plc:bob"}, did)
		}
	})

	t.Run("auto links", func(t *testing.T) {
		post, err := NewBuilder(WithAutoLink(true)).
			AddText("Check https://example.com and https://test.com").
			Build()

		assert.NoError(t, err)
		assert.Equal(t, "Check https://example.com and https://test.com", post.Text)
		assert.Len(t, post.Facets, 2)

		// Verify links
		for _, facet := range post.Facets {
			assert.NotNil(t, facet.Features[0].RichtextFacet_Link)
			uri := facet.Features[0].RichtextFacet_Link.Uri
			assert.Contains(t, []string{"https://example.com", "https://test.com"}, uri)
		}
	})

	t.Run("all auto features", func(t *testing.T) {
		post, err := NewBuilder(
			WithAutoHashtag(true),
			WithAutoMention(true),
			WithAutoLink(true),
		).AddText("Hi @alice! Check #golang at https://golang.org #programming").
			Build()

		assert.NoError(t, err)
		assert.Equal(t, "Hi @alice! Check #golang at https://golang.org #programming", post.Text)
		assert.Len(t, post.Facets, 4) // 1 mention + 2 hashtags + 1 link

		var hashtags, mentions, links int
		for _, facet := range post.Facets {
			if facet.Features[0].RichtextFacet_Tag != nil {
				hashtags++
			}
			if facet.Features[0].RichtextFacet_Mention != nil {
				mentions++
			}
			if facet.Features[0].RichtextFacet_Link != nil {
				links++
			}
		}

		assert.Equal(t, 2, hashtags)
		assert.Equal(t, 1, mentions)
		assert.Equal(t, 1, links)
	})

	t.Run("invalid auto-detected items", func(t *testing.T) {
		post, err := NewBuilder(
			WithAutoHashtag(true),
			WithAutoMention(true),
			WithAutoLink(true),
		).AddText("@invalid user #invalid tag https://").
			Build()

		assert.NoError(t, err)
		assert.Equal(t, "@invalid user #invalid tag https://", post.Text)
		assert.Empty(t, post.Facets) // All items should be invalid and ignored
	})
}

func TestBuilderMaxLength(t *testing.T) {
	t.Run("custom max length", func(t *testing.T) {
		builder := NewBuilder(WithMaxLength(10))
		post, err := builder.AddText("12345").Build()
		assert.NoError(t, err)
		assert.Equal(t, "12345", post.Text)

		_, err = builder.AddText("123456").Build()
		assert.ErrorIs(t, err, ErrPostTooLong)
	})

	t.Run("invalid max length", func(t *testing.T) {
		assert.Panics(t, func() {
			NewBuilder(WithMaxLength(0))
		})
		assert.Panics(t, func() {
			NewBuilder(WithMaxLength(-1))
		})
		assert.Panics(t, func() {
			NewBuilder(WithMaxLength(maxPostLength + 1))
		})
	})
}
