package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/joho/godotenv"
	"github.com/watzon/lining/client"
	"github.com/watzon/lining/config"
	"github.com/watzon/lining/models"
)

func main() {
	godotenv.Load()
	handle := os.Getenv("HANDLE")
	apikey := os.Getenv("APIKEY")

	cfg := config.Default()
	cfg.Handle = handle
	cfg.APIKey = apikey

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
	post, err := client.NewPostBuilder().
		AddText("Check out this link!").
		WithExternalLink(link).
		Build()
	if err != nil {
		log.Fatal(err)
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
	uploadedImage, err := c.UploadImage(context.Background(), img)
	if err != nil {
		log.Fatal(err)
	}

	// Create a post with the image
	imagePost, err := client.NewPostBuilder().
		AddText("Check out this image!").
		WithImages([]models.UploadedImage{*uploadedImage}).
		Build()
	if err != nil {
		log.Fatal(err)
	}

	// Create the post with image
	cid, postUri, err = c.PostToFeed(context.Background(), imagePost)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Image post created with CID: %s and URI: %s\n", cid, postUri)
}
