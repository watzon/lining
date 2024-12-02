package post

import (
	"fmt"

	"github.com/bluesky-social/indigo/api/bsky"
)

type FacetTypeMention struct {
	Did string
}

type FacetTypeTag struct {
	Tag string
}

type FacetTypeLink struct {
	Uri string
}

type FacetType struct {
	Type             string
	FacetTypeMention *FacetTypeMention
	FacetTypeTag     *FacetTypeTag
	FacetTypeLink    *FacetTypeLink
}

type FacetByteSlice struct {
	ByteStart int64
	ByteEnd   int64
}

type Facet struct {
	Type  FacetType
	Index FacetByteSlice
	Text  string
}

func ExtractFacetsFromFeedPost(feedPost *bsky.FeedPost) []Facet {
	var facets []Facet
	for _, facet := range feedPost.Facets {
		for _, feature := range facet.Features {
			index := FacetByteSlice{
				ByteStart: facet.Index.ByteStart,
				ByteEnd:   min(facet.Index.ByteEnd, int64(len(feedPost.Text))),
			}
			text := feedPost.Text[index.ByteStart:index.ByteEnd]
			switch {
			case feature.RichtextFacet_Link != nil:
				facets = append(facets, Facet{
					Type: FacetType{
						Type: "app.bsky.richtext.facet#link",
						FacetTypeLink: &FacetTypeLink{
							Uri: feature.RichtextFacet_Link.Uri,
						},
					},
					Index: index,
					Text:  text,
				})
			case feature.RichtextFacet_Mention != nil:
				facets = append(facets, Facet{
					Type: FacetType{
						Type: "app.bsky.richtext.facet#mention",
						FacetTypeMention: &FacetTypeMention{
							Did: feature.RichtextFacet_Mention.Did,
						},
					},
					Index: index,
					Text:  text,
				})
			case feature.RichtextFacet_Tag != nil:
				facets = append(facets, Facet{
					Type: FacetType{
						Type: "app.bsky.richtext.facet#tag",
						FacetTypeTag: &FacetTypeTag{
							Tag: feature.RichtextFacet_Tag.Tag,
						},
					},
					Index: index,
					Text:  text,
				})
			default:
				fmt.Printf("Unknown facet type: %v\n", feature)
			}
		}
	}
	return facets
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
