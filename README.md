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
    "net/url"
    "os"

    "github.com/joho/godotenv"
    "github.com/watzon/lining/client"
    "github.com/watzon/lining/config"
    "github.com/watzon/lining/models"
)

func main() {
    if err := godotenv.Load(); err != nil {
        log.Fatal(err)
    }

    cfg := config.DefaultConfig().
        WithHandle(os.Getenv("HANDLE")).
        WithAPIKey(os.Getenv("APIKEY"))

    client, err := client.NewClient(cfg)
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()
    if err := client.Connect(ctx); err != nil {
        log.Fatal(err)
    }

    // Facets Section
    // =======================================
    // Facet type can be Link, Mention or Tag
    post1, err := client.NewPostBuilder("Hello to Bluesky, the coolest open social network").
        WithFacet(models.FacetLink, "https://docs.bsky.app/", "Bluesky").
        WithFacet(models.FacetTag, "bsky", "open social").
        Build()
    if err != nil {
        log.Fatal(err)
    }

    cid1, uri1, err := client.PostToFeed(ctx, post1)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Success: Cid = %v, Uri = %v\n", cid1, uri1)

    // Embed Links section
    // =======================================
    u, err := url.Parse("https://go.dev/")
    if err != nil {
        log.Fatal(err)
    }

    previewUrl, err := url.Parse("https://www.freecodecamp.org/news/content/images/2021/10/golang.png")
    if err != nil {
        log.Fatal(err)
    }
    previewImage := models.Image{
        Title: "Golang",
        Uri:   *previewUrl,
    }
    previewImageBlob, err := client.UploadImage(ctx, previewImage)
    if err != nil {
        log.Fatal(err)
    }

    post2, err := client.NewPostBuilder("Hello to Go on Bluesky").
        WithExternalLink("Go Programming Language", *u, "Build simple, secure, scalable systems with Go", *previewImageBlob).
        Build()
    if err != nil {
        log.Fatal(err)
    }

    cid2, uri2, err := client.PostToFeed(ctx, post2)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Success: Cid = %v, Uri = %v\n", cid2, uri2)

    // Embed Images section
    // =======================================
    images := []models.Image{}

    url1, err := url.Parse("https://www.freecodecamp.org/news/content/images/2021/10/golang.png")
    if err != nil {
        log.Fatal(err)
    }
    images = append(images, models.Image{
        Title: "Golang",
        Uri:   *url1,
    })

    blobs, err := client.UploadImages(ctx, images...)
    if err != nil {
        log.Fatal(err)
    }

    post3, err := client.NewPostBuilder("Lining - a simple golang lib to write Bluesky bots").
        WithImages(images, blobs).
        Build()
    if err != nil {
        log.Fatal(err)
    }

    cid3, uri3, err := client.PostToFeed(ctx, post3)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Success: Cid = %v, Uri = %v\n", cid3, uri3)
}

## Examples

### Simple Text Post

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/watzon/lining/client"
    "github.com/watzon/lining/config"
)

func main() {
    cfg := config.DefaultConfig().
        WithHandle("your-handle.bsky.social").
        WithAPIKey("your-api-key")

    client, err := client.NewClient(cfg)
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()
    if err := client.Connect(ctx); err != nil {
        log.Fatal(err)
    }

    post, err := client.NewPostBuilder("Hello Bluesky!").Build()
    if err != nil {
        log.Fatal(err)
    }

    cid, uri, err := client.PostToFeed(ctx, post)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Posted successfully: %s\n", uri)
}
```

### Rich Text Post with Mentions and Tags

```go
package main

import (
    "github.com/watzon/lining/client"
    "github.com/watzon/lining/config"
    "github.com/watzon/lining/models"
)

func main() {
    // ... client setup code ...

    post, err := client.NewPostBuilder("Check out @someone's #awesome post about #bluesky").
        WithFacet(models.FacetMention, "did:plc:someone", "@someone").
        WithFacet(models.FacetTag, "awesome", "#awesome").
        WithFacet(models.FacetTag, "bluesky", "#bluesky").
        Build()
}
```

### Post with Image

```go
package main

import (
    "net/url"

    "github.com/watzon/lining/client"
    "github.com/watzon/lining/config"
    "github.com/watzon/lining/models"
)

func main() {
    // ... client setup code ...

    imageUrl, _ := url.Parse("https://example.com/image.png")
    image := models.Image{
        Title: "My Image",
        Uri:   *imageUrl,
    }

    blob, err := client.UploadImage(ctx, image)
    if err != nil {
        log.Fatal(err)
    }

    post, err := client.NewPostBuilder("Check out this cool image!").
        WithImages([]models.Image{image}, []blob.Blob{*blob}).
        Build()
}
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Acknowledgements

- Thanks to [Dan Rusei](https://github.com/danrusei) for his work on [gobot-bsky](https://github.com/danrusei/gobot-bsky) for providing the inspiration and initial base for this project.
- Thanks to [bluesky-social](https://github.com/bluesky-social) for providing the Bluesky API documentation and examples.

## License

This project, like the original, is licensed under the Apache License, Version 2.0. For more information, please see the [LICENSE](LICENSE) file.