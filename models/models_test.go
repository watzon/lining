package models

import "testing"

func TestFacetType_String(t *testing.T) {
	tests := []struct {
		name string
		f    FacetType
		want string
	}{
		{
			name: "Link facet",
			f:    FacetLink,
			want: "app.bsky.richtext.facet#link",
		},
		{
			name: "Mention facet",
			f:    FacetMention,
			want: "app.bsky.richtext.facet#mention",
		},
		{
			name: "Tag facet",
			f:    FacetTag,
			want: "app.bsky.richtext.facet#tag",
		},
		{
			name: "Unknown facet",
			f:    FacetType(999), // Some invalid value
			want: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.f.String(); got != tt.want {
				t.Errorf("FacetType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
