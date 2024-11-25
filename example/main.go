package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/joho/godotenv"

	"github.com/watzon/lining/client"
	"github.com/watzon/lining/config"
	"github.com/watzon/lining/models"
)

func main() {
	godotenv.Load()
	handle := os.Getenv("HANDLE")
	apikey := os.Getenv("APIKEY")
	server := "https://bsky.social"

	cfg := &config.Config{
		Handle:            handle,
		APIKey:            apikey,
		ServerURL:         server,
		Timeout:           30 * time.Second,
		RequestsPerMinute: 60,
		BurstSize:         5,
	}

	// Create a new client
	c, err := client.NewClient(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Connect to the server
	err = c.Connect(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// Create a post with text and link
	uri, err := url.Parse("https://example.com")
	if err != nil {
		log.Fatal(err)
	}

	link := models.Link{
		Uri:         *uri,
		Title:       "Example Link",
		Description: "This is an example link",
	}

	// Create a post with text and link
	post := bsky.FeedPost{
		Text:          "Check out this link!",
		LexiconTypeID: "app.bsky.feed.post",
		CreatedAt:     time.Now().Format(time.RFC3339),
		Embed: &bsky.FeedPost_Embed{
			EmbedExternal: &bsky.EmbedExternal{
				LexiconTypeID: "app.bsky.embed.external",
				External: &bsky.EmbedExternal_External{
					Uri:         link.Uri.String(),
					Title:       link.Title,
					Description: link.Description,
				},
			},
		},
	}

	// Create the post
	cid, postUri, err := c.PostToFeed(context.Background(), post)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Post created with CID: %s and URI: %s\n", cid, postUri)

	// Create a post with an image
	img := models.Image{
		Title: "Example Image",
		Data:  []byte("example image data"), // In real usage, this would be actual image data
	}

	// Upload the image
	uploadedBlob, err := c.UploadImage(context.Background(), img)
	if err != nil {
		log.Fatal(err)
	}

	// Create a post with the image
	imagePost := bsky.FeedPost{
		Text:          "Check out this image!",
		LexiconTypeID: "app.bsky.feed.post",
		CreatedAt:     time.Now().Format(time.RFC3339),
		Embed: &bsky.FeedPost_Embed{
			EmbedImages: &bsky.EmbedImages{
				LexiconTypeID: "app.bsky.embed.images",
				Images: []*bsky.EmbedImages_Image{
					{
						Alt:   img.Title,
						Image: uploadedBlob,
					},
				},
			},
		},
	}

	// Create the post with image
	cid, postUri, err = c.PostToFeed(context.Background(), imagePost)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Image post created with CID: %s and URI: %s\n", cid, postUri)
}
