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

Has to provide:

* a handle -  example bluesky handle: "example.bsky.social"
* an apikey - is used for authentication and the retrieval of the access token and refresh token. To create a new one: Settings -> App Passwords 
* the server (PDS) - the Bluesky's "PDS Service" is bsky.social. 

```go
import "github.com/watzon/lining"

func main() {

	godotenv.Load()
	handle := os.Getenv("HANDLE")
	apikey := os.Getenv("APIKEY")
	server := "https://bsky.social"

	ctx := context.Background()

	agent := lining.NewAgent(ctx, server, handle, apikey)
	agent.Connect(ctx)

	// Facets Section
	// =======================================
	// Facet_type coulf be Facet_Link, Facet_Mention or Facet_Tag
	// based on the selected type it expect the second argument to be URI, DID, or TAG
	// the last function argument is the text, part of the original text that is modifiend in Richtext

	post1, err := lining.NewPostBuilder("Hello to Bluesky, the coolest open social network").
		WithFacet(lining.Facet_Link, "https://docs.bsky.app/", "Bluesky").
		WithFacet(lining.Facet_Tag, "bsky", "open social").
		Build()
	if err != nil {
		fmt.Printf("Got error: %v", err)
	}

	cid1, uri1, err := agent.PostToFeed(ctx, post1)
	if err != nil {
		fmt.Printf("Got error: %v", err)
	} else {
		fmt.Printf("Succes: Cid = %v , Uri = %v", cid1, uri1)
	}

	// Embed Links section
	// =======================================

	u, err := url.Parse("https://go.dev/")
	if err != nil {
		log.Fatalf("Parse error, %v", err)
	}

	previewUrl, err := url.Parse("https://www.freecodecamp.org/news/content/images/2021/10/golang.png")
	if err != nil {
		log.Fatalf("Parse error, %v", err)
	}
	previewImage := lining.Image{
		Title: "Golang",
		Uri:   *previewUrl,
	}
	previewImageBlob, err := agent.UploadImage(ctx, previewImage)
	if err != nil {
		log.Fatalf("Parse error, %v", err)
	}

	post2, err := lining.NewPostBuilder("Hello to Go on Bluesky").
		WithExternalLink("Go Programming Language", *u, "Build simple, secure, scalable systems with Go", *previewImageBlob).
		Build()
	if err != nil {
		fmt.Printf("Got error: %v", err)
	}

	cid2, uri2, err := agent.PostToFeed(ctx, post2)
	if err != nil {
		fmt.Printf("Got error: %v", err)
	} else {
		fmt.Printf("Succes: Cid = %v , Uri = %v", cid2, uri2)
	}

	// Embed Images section
	// =======================================
	images := []lining.Image{}

	url1, err := url.Parse("https://www.freecodecamp.org/news/content/images/2021/10/golang.png")
	if err != nil {
		log.Fatalf("Parse error, %v", err)
	}
	images = append(images, lining.Image{
		Title: "Golang",
		Uri:   *url1,
	})

	blobs, err := agent.UploadImages(ctx, images...)
	if err != nil {
		log.Fatalf("Parse error, %v", err)
	}

	post3, err := lining.NewPostBuilder("Lining - a simple golang lib to write Bluesky bots").
		WithImages(blobs, images).
		Build()
	if err != nil {
		fmt.Printf("Got error: %v", err)
	}

	cid3, uri3, err := agent.PostToFeed(ctx, post3)
	if err != nil {
		fmt.Printf("Got error: %v", err)
	} else {
		fmt.Printf("Succes: Cid = %v , Uri = %v", cid3, uri3)
	}

}

## Examples

### Simple Text Post

```go
ctx := context.Background()
if err := client.Connect(ctx); err != nil {
    log.Fatal(err)
}

post, err := lining.NewPostBuilder("Hello Bluesky!").Build()
if err != nil {
    log.Fatal(err)
}

cid, uri, err := client.PostToFeed(ctx, post)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Posted successfully: %s\n", uri)
```

### Rich Text Post with Mentions and Tags

```go
post, err := lining.NewPostBuilder("Check out @someone's #awesome post about #bluesky").
    WithFacet(lining.Facet_Mention, "did:plc:someone", "@someone").
    WithFacet(lining.Facet_Tag, "awesome", "#awesome").
    WithFacet(lining.Facet_Tag, "bluesky", "#bluesky").
    Build()
```

### Post with Image

```go
imageUrl, _ := url.Parse("https://example.com/image.png")
image := lining.Image{
    Title: "My Image",
    Uri:   *imageUrl,
}

blob, err := client.UploadImage(ctx, image)
if err != nil {
    log.Fatal(err)
}

post, err := lining.NewPostBuilder("Check out this image!").
    WithImages([]lexutil.LexBlob{*blob}, []lining.Image{image}).
    Build()
```

### Follow a User

```go
err := client.Follow(ctx, "did:plc:someuser")
if err != nil {
    log.Fatal(err)
}
```

### Get User Profile

```go
profile, err := client.GetProfile(ctx, "user.bsky.social")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Display Name: %s\n", profile.DisplayName)
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Acknowledgements

- Thanks to [Dan Rusei](https://github.com/danrusei) for his work on [gobot-bsky](https://github.com/danrusei/gobot-bsky) for providing the inspiration and initial base for this project.
- Thanks to [bluesky-social](https://github.com/bluesky-social) for providing the Bluesky API documentation and examples.

## License

This project, like the original, is licensed under the Apache License, Version 2.0. For more information, please see the [LICENSE](LICENSE) file.