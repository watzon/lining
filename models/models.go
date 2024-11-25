package models

import (
	"net/url"

	lexutil "github.com/bluesky-social/indigo/lex/util"
)

// Image represents an image to be uploaded to Bluesky
type Image struct {
	Title string
	Data  []byte
}

// Link represents an external link in a post
type Link struct {
	Title       string
	Uri         url.URL
	Description string
	Thumb       lexutil.LexBlob
}

// Embed represents embedded content in a post
type Embed struct {
	Link           Link
	Images         []Image
	UploadedImages []lexutil.LexBlob
}

// Facet represents rich text features in a post
type Facet struct {
	Type  FacetType
	Value string
	Text  string
}

// FacetType represents the type of a facet
type FacetType int

const (
	FacetLink FacetType = iota + 1
	FacetMention
	FacetTag
)

// String returns the string representation of a FacetType
func (f FacetType) String() string {
	switch f {
	case FacetLink:
		return "app.bsky.richtext.facet#link"
	case FacetMention:
		return "app.bsky.richtext.facet#mention"
	case FacetTag:
		return "app.bsky.richtext.facet#tag"
	default:
		return "unknown"
	}
}
