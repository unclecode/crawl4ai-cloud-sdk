# Sessions Examples

This directory contains Go examples for browser session management with Crawl4AI Cloud.

## What are Sessions?

Browser sessions provide a persistent browser instance in the cloud. You get a WebSocket URL
that you can connect to using:

- **Crawl4AI** (local library with cloud browser)
- **Puppeteer** (puppeteer-core)
- **Playwright** (playwright-core)
- **chromedp** (Go Chrome DevTools Protocol client)

## Files

- **01_basic_session_http.go** - Create, use, and release sessions via HTTP API

## Usage

```bash
cd sessions
go run 01_basic_session_http.go
```

## Session Lifecycle

1. **Create**: Request a new browser session (returns WebSocket URL)
2. **Connect**: Use the WebSocket URL with your preferred tool
3. **Use**: Navigate, interact, extract data
4. **Disconnect**: Close your connection (don't close the browser)
5. **Release**: Call the release endpoint to free resources

## Go Chrome DevTools Example

```go
package main

import (
    "context"
    "github.com/chromedp/chromedp"
)

func main() {
    // Connect to cloud browser session
    wsURL := "wss://your-session-ws-url"

    allocCtx, cancel := chromedp.NewRemoteAllocator(context.Background(), wsURL)
    defer cancel()

    ctx, cancel := chromedp.NewContext(allocCtx)
    defer cancel()

    var title string
    err := chromedp.Run(ctx,
        chromedp.Navigate("https://example.com"),
        chromedp.Title(&title),
    )
    if err != nil {
        panic(err)
    }

    println("Title:", title)
}
```

## Important Notes

- Sessions have a timeout (default: 600 seconds)
- Always release sessions when done to avoid charges
- The cloud manages the browser lifecycle
