# Lining - A Silver Lining for Your Bluesky Bot

A simple Go library for interacting with the Bluesky API

## About

Lining (inspired by "silver lining", as in "every cloud has a silver lining") is a Go library that
provides a simple interface for creating bots on Bluesky. It handles authentication,
posting, and other common bot operations through a clean and intuitive API.

## Installation

```bash
go get github.com/watzon/lining
```

## Features

- Simple and intuitive API for interacting with Bluesky
- Rate limiting to prevent API abuse
- Automatic token refresh
- Support for rich text posts with mentions, links, and tags
- Image upload support
- Follow/unfollow functionality
- Profile fetching
- Comprehensive error handling
- Fully tested with unit tests

## Configuration

The library uses a configuration struct for initialization:

```go
cfg := config.DefaultConfig().
    WithHandle("your-handle.bsky.social").
    WithAPIKey("your-api-key")

client, err := client.NewClient(cfg)
if err != nil {
    log.Fatal(err)
}
```

Available configuration options:
- Handle: Your Bluesky handle
- APIKey: Your API key (create one in Settings -> App Passwords)
- ServerURL: Bluesky PDS URL (defaults to https://bsky.social)
- Timeout: HTTP client timeout
- RetryAttempts: Number of retry attempts for failed requests
- RequestsPerMinute: Rate limiting configuration
- Debug: Enable debug logging

## Usage example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/joho/godotenv"
    "github.com/watzon/lining/client"
    "github.com/watzon/lining/config"
    "github.com/watzon/lining/models"
)

func main() {
    // Load environment variables from .env file
    if err := godotenv.Load(); err != nil {
        log.Fatal(err)
    }

    // Create a new client
    cfg := config.DefaultConfig().
        WithHandle(os.Getenv("HANDLE")).
        WithAPIKey(os.Getenv("APIKEY"))

    client, err := client.NewClient(cfg)
    if err != nil {
        log.Fatal(err)
    }

    // Connect to Bluesky
    ctx := context.Background()
    if err := client.Connect(ctx); err != nil {
        log.Fatal(err)
    }

    // Create a post with a mention and a tag
    post, err := client.NewPostBuilder("Hello @watzon! Check out this #bluesky bot!").
        WithFacet(models.FacetMention, "did:plc:watzon", "@watzon").
        WithFacet(models.FacetTag, "bluesky", "#bluesky").
        Build()
    if err != nil {
        log.Fatal(err)
    }

    // Post to your feed
    cid, uri, err := client.PostToFeed(ctx, post)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Posted successfully! URI: %s\n", uri)
}
```

## Examples

### Simple Text Post

```go
post, err := client.NewPostBuilder("Hello Bluesky!").Build()
if err != nil {
    log.Fatal(err)
}

cid, uri, err := client.PostToFeed(ctx, post)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Posted successfully: %s\n", uri)
```

### Rich Text with Mentions, Links, and Tags

```go
post, err := client.NewPostBuilder("Check out @someone's link to docs.bsky.app #bluesky").
    WithFacet(models.FacetMention, "did:plc:someone", "@someone").
    WithFacet(models.FacetLink, "https://docs.bsky.app", "docs.bsky.app").
    WithFacet(models.FacetTag, "bluesky", "#bluesky").
    Build()
```

### Image Upload Methods

#### 1. Upload from Image struct

```go
// Upload using an Image struct with raw bytes
imageData, err := os.ReadFile("path/to/image.jpg")
if err != nil {
    log.Fatal(err)
}

image := models.Image{
    Title: "My Cool Image",
    Data:  imageData,
}

blob, err := client.UploadImage(ctx, image)
if err != nil {
    log.Fatal(err)
}

post, err := client.NewPostBuilder("Check out this image!").
    WithImages([]models.Image{image}, []blob.Blob{*blob}).
    Build()
```

#### 2. Upload directly from URL

```go
// Upload directly from a URL - no need to create Image struct
blob, err := client.UploadImageFromURL(ctx, "https://example.com/image.jpg", "Cool Image From URL")
if err != nil {
    log.Fatal(err)
}

image := models.Image{
    Title: "Cool Image From URL",
}

post, err := client.NewPostBuilder("Check out this image I found!").
    WithImages([]models.Image{image}, []blob.Blob{*blob}).
    Build()
```

#### 3. Upload from local file

```go
// Upload from a local file - no need to handle the bytes manually
blob, err := client.UploadImageFromFile(ctx, "/path/to/local/image.jpg", "My Local Image")
if err != nil {
    log.Fatal(err)
}

image := models.Image{
    Title: "My Local Image",
}

post, err := client.NewPostBuilder("Just took this photo!").
    WithImages([]models.Image{image}, []blob.Blob{*blob}).
    Build()
```

### Rich Text with Mentions, Links, and Tags

```go
post, err := client.NewPostBuilder("Check out @someone's link to docs.bsky.app #bluesky").
    WithFacet(models.FacetMention, "did:plc:someone", "@someone").
    WithFacet(models.FacetLink, "https://docs.bsky.app", "docs.bsky.app").
    WithFacet(models.FacetTag, "bluesky", "#bluesky").
    Build()
```

### Profile Operations

```go
// Get a user's profile
profile, err := client.GetProfile(ctx, "watzon.bsky.social")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Display Name: %s\n", profile.DisplayName)

// Follow a user
err = client.Follow(ctx, "did:plc:someuser")
if err != nil {
    log.Fatal(err)
}
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Acknowledgements

- Thanks to [Dan Rusei](https://github.com/danrusei) for his work on [gobot-bsky](https://github.com/danrusei/gobot-bsky) for providing the inspiration and initial base for this project.
- Thanks to [bluesky-social](https://github.com/bluesky-social) for providing the Bluesky API documentation and examples.

## License

This project, like the original, is licensed under the Apache License, Version 2.0. For more information, please see the [LICENSE](LICENSE) file.