# Scryfall Go Client

[![Go Reference](https://pkg.go.dev/badge/github.com/repricah/scryfall.svg)](https://pkg.go.dev/github.com/repricah/scryfall)
[![Go Report Card](https://goreportcard.com/badge/github.com/repricah/scryfall)](https://goreportcard.com/report/github.com/repricah/scryfall)
[![CI](https://github.com/repricah/scryfall/actions/workflows/ci.yml/badge.svg)](https://github.com/repricah/scryfall/actions/workflows/ci.yml)

A lightweight Go client for the [Scryfall API](https://scryfall.com/docs/api).

- Focused on the core API + bulk data endpoints
- Configurable base URL, user agent, rate limiter, and HTTP client
- Streaming bulk data parsing for large payloads

## Installation

```bash
go get github.com/repricah/scryfall
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"

    "github.com/repricah/scryfall"
)

func main() {
    client := scryfall.NewClient(
        scryfall.WithUserAgent("my-app/1.0"),
    )

    card, err := client.GetCardByID(context.Background(), "0cc2cb3f-b1e7-4d6c-8ae6-6f4d095e0b6c")
    if err != nil {
        panic(err)
    }

    fmt.Println(card.Name)
}
```

## Bulk Data Streaming

```go
err := client.DownloadBulkDataStream(ctx, bulk.DownloadURI, func(card scryfall.Card) error {
    // process each card
    return nil
}, func(current, total int64) {
    // progress tracking
})
```

## Configuration

```go
client := scryfall.NewClient(
    scryfall.WithBaseURL("https://api.scryfall.com"),
    scryfall.WithUserAgent("my-app/1.0"),
    scryfall.WithLimiter(rate.NewLimiter(rate.Limit(10), 10)),
    scryfall.WithHTTPClient(&http.Client{Timeout: 30 * time.Second}),
)
```

## API Notes

Scryfall requests should include a clear user agent that identifies your app.
See the official documentation and API announcements for usage policies and
bulk data details:

- https://scryfall.com/docs/api
- https://scryfall.com/docs/api/bulk-data
- https://scryfall.com/blog/category/api

## License

MIT License - see LICENSE for details.
