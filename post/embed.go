package post

import (
	"fmt"

	"github.com/bluesky-social/indigo/api/bsky"
)

type AspectRatio struct {
	Width  int64
	Height int64
}

type EmbedImage struct {
	Alt         string
	AspectRatio *AspectRatio
	Ref         string
	MimeType    string
	Size        int64
}

type EmbedVideoCaption struct {
	Lang string
	Text string
}

type EmbedVideo struct {
	Alt         string
	AspectRatio *AspectRatio
	Captions    []*EmbedVideoCaption
	Ref         string
	MimeType    string
	Size        int64
}

type EmbedExternal struct {
	Description string
	Uri         string
	ThumbRef    string
	Title       string
}

type EmbedRecord struct {
	Cid string
	Uri string
}

type EmbedRecordWithMedia_Media struct {
	Images   []*EmbedImage
	Video    *EmbedVideo
	External *EmbedExternal
}

type EmbedRecordWithMedia struct {
	Media  *EmbedRecordWithMedia_Media
	Record *EmbedRecord
}

type Embed struct {
	Images          []*EmbedImage
	External        *EmbedExternal
	Record          *EmbedRecord
	Video           *EmbedVideo
	RecordWithMedia *EmbedRecordWithMedia
}

func ExtractEmbedFromFeedPost(feedPost *bsky.FeedPost) (*Embed, error) {
	embed := &Embed{
		Images: make([]*EmbedImage, 0),
	}

	if feedPost.Embed != nil {
		switch {
		case feedPost.Embed.EmbedImages != nil:
			images := make([]*EmbedImage, len(feedPost.Embed.EmbedImages.Images))
			for i, image := range feedPost.Embed.EmbedImages.Images {
				var aspectRatio *AspectRatio
				if image.AspectRatio != nil {
					aspectRatio = &AspectRatio{
						Width:  image.AspectRatio.Width,
						Height: image.AspectRatio.Height,
					}
				}

				if image.Image == nil {
					return nil, fmt.Errorf("image blob is nil")
				}

				images[i] = &EmbedImage{
					Alt:         image.Alt,
					AspectRatio: aspectRatio,
					Ref:         image.Image.Ref.String(),
					MimeType:    image.Image.MimeType,
					Size:        image.Image.Size,
				}
			}
			embed = &Embed{
				Images: images,
			}
		case feedPost.Embed.EmbedVideo != nil:
			var aspectRatio *AspectRatio
			if feedPost.Embed.EmbedVideo.AspectRatio != nil {
				aspectRatio = &AspectRatio{
					Width:  feedPost.Embed.EmbedVideo.AspectRatio.Width,
					Height: feedPost.Embed.EmbedVideo.AspectRatio.Height,
				}
			}

			if feedPost.Embed.EmbedVideo.Video == nil {
				return nil, fmt.Errorf("video blob is nil")
			}

			var alt string
			if feedPost.Embed.EmbedVideo.Alt != nil {
				alt = *feedPost.Embed.EmbedVideo.Alt
			}

			embed = &Embed{
				Video: &EmbedVideo{
					Alt:         alt,
					AspectRatio: aspectRatio,
					Ref:         feedPost.Embed.EmbedVideo.Video.Ref.String(),
					MimeType:    feedPost.Embed.EmbedVideo.Video.MimeType,
					Size:        feedPost.Embed.EmbedVideo.Video.Size,
				},
				Images: make([]*EmbedImage, 0),
			}
		case feedPost.Embed.EmbedExternal != nil:
			var thumbRef string
			if feedPost.Embed.EmbedExternal.External.Thumb != nil {
				thumbRef = feedPost.Embed.EmbedExternal.External.Thumb.Ref.String()
			}

			embed = &Embed{
				External: &EmbedExternal{
					Description: feedPost.Embed.EmbedExternal.External.Description,
					Uri:         feedPost.Embed.EmbedExternal.External.Uri,
					ThumbRef:    thumbRef,
					Title:       feedPost.Embed.EmbedExternal.External.Title,
				},
				Images: make([]*EmbedImage, 0),
			}
		case feedPost.Embed.EmbedRecord != nil:
			embed = &Embed{
				Record: &EmbedRecord{
					Cid: feedPost.Embed.EmbedRecord.Record.Cid,
					Uri: feedPost.Embed.EmbedRecord.Record.Uri,
				},
				Images: make([]*EmbedImage, 0),
			}
		case feedPost.Embed.EmbedRecordWithMedia != nil:
			media := feedPost.Embed.EmbedRecordWithMedia.Media
			switch {
			case media.EmbedImages != nil:
				images := make([]*EmbedImage, len(media.EmbedImages.Images))
				for i, image := range media.EmbedImages.Images {
					var aspectRatio *AspectRatio
					if image.AspectRatio != nil {
						aspectRatio = &AspectRatio{
							Width:  image.AspectRatio.Width,
							Height: image.AspectRatio.Height,
						}
					}

					if image.Image == nil {
						return nil, fmt.Errorf("image blob is nil")
					}

					images[i] = &EmbedImage{
						Alt:         image.Alt,
						AspectRatio: aspectRatio,
						Ref:         image.Image.Ref.String(),
						MimeType:    image.Image.MimeType,
						Size:        image.Image.Size,
					}
				}
				embed = &Embed{
					RecordWithMedia: &EmbedRecordWithMedia{
						Media: &EmbedRecordWithMedia_Media{
							Images: images,
						},
					},
				}
			case media.EmbedVideo != nil:
				var aspectRatio *AspectRatio
				if media.EmbedVideo.AspectRatio != nil {
					aspectRatio = &AspectRatio{
						Width:  media.EmbedVideo.AspectRatio.Width,
						Height: media.EmbedVideo.AspectRatio.Height,
					}
				}

				if media.EmbedVideo.Video == nil {
					return nil, fmt.Errorf("video blob is nil")
				}

				var alt string
				if media.EmbedVideo.Alt != nil {
					alt = *media.EmbedVideo.Alt
				}

				embed = &Embed{
					RecordWithMedia: &EmbedRecordWithMedia{
						Media: &EmbedRecordWithMedia_Media{
							Video: &EmbedVideo{
								Alt:         alt,
								AspectRatio: aspectRatio,
								Ref:         media.EmbedVideo.Video.Ref.String(),
								MimeType:    media.EmbedVideo.Video.MimeType,
								Size:        media.EmbedVideo.Video.Size,
							},
						},
					},
				}
			case media.EmbedExternal != nil:
				var thumbRef string
				if media.EmbedExternal.External.Thumb != nil {
					thumbRef = media.EmbedExternal.External.Thumb.Ref.String()
				}

				embed = &Embed{
					RecordWithMedia: &EmbedRecordWithMedia{
						Media: &EmbedRecordWithMedia_Media{
							External: &EmbedExternal{
								Description: media.EmbedExternal.External.Description,
								Uri:         media.EmbedExternal.External.Uri,
								ThumbRef:    thumbRef,
								Title:       media.EmbedExternal.External.Title,
							},
						},
					},
				}
			}
		}
	}
	return embed, nil
}
