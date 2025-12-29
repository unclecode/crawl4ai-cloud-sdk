# Crawl4AI Cloud SDK for Go

Lightweight Go SDK for [Crawl4AI Cloud API](https://api.crawl4ai.com). Idiomatic Go implementation.

[![Go Reference](https://pkg.go.dev/badge/github.com/unclecode/crawl4ai-cloud/go.svg)](https://pkg.go.dev/github.com/unclecode/crawl4ai-cloud/go)

## Installation

```bash
go get github.com/unclecode/crawl4ai-cloud/go
```

## Get Your API Key

1. Go to [api.crawl4ai.com](https://api.crawl4ai.com)
2. Sign up and get your API key

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/unclecode/crawl4ai-cloud/go/pkg/crawl4ai"
)

func main() {
    crawler, err := crawl4ai.NewAsyncWebCrawler(crawl4ai.CrawlerOptions{
        APIKey: "sk_live_...",
    })
    if err != nil {
        panic(err)
    }

    result, err := crawler.Run("https://example.com", nil)
    if err != nil {
        panic(err)
    }

    fmt.Println(result.Markdown.RawMarkdown)
}
```

## Features

### Single URL Crawl

```go
result, err := crawler.Run("https://example.com", nil)
if err != nil {
    log.Fatal(err)
}

fmt.Println(result.Success)
fmt.Println(result.Markdown.RawMarkdown)
fmt.Println(result.HTML)
```

### Batch Crawl

```go
urls := []string{"https://example.com", "https://httpbin.org/html"}

// Wait for results
result, err := crawler.RunMany(urls, &crawl4ai.RunManyOptions{Wait: true})
if err != nil {
    log.Fatal(err)
}

for _, r := range result.Results {
    fmt.Printf("%s: %v\n", r.URL, r.Success)
}

// Fire and forget (returns job)
result, err = crawler.RunMany(urls, &crawl4ai.RunManyOptions{Wait: false})
fmt.Printf("Job ID: %s\n", result.Job.ID)
```

### Configuration

```go
config := &crawl4ai.CrawlerRunConfig{
    WordCountThreshold:   10,
    ExcludeExternalLinks: true,
    Screenshot:           true,
}

browserConfig := &crawl4ai.BrowserConfig{
    ViewportWidth:  1920,
    ViewportHeight: 1080,
}

result, err := crawler.Run("https://example.com", &crawl4ai.RunOptions{
    Config:        config,
    BrowserConfig: browserConfig,
})
```

### Proxy Support

```go
// Shorthand
result, err := crawler.Run(url, &crawl4ai.RunOptions{
    Proxy: "datacenter",
})

// Full config
result, err := crawler.Run(url, &crawl4ai.RunOptions{
    Proxy: &crawl4ai.ProxyConfig{
        Mode:    "residential",
        Country: "US",
    },
})
```

### Deep Crawl

```go
result, err := crawler.DeepCrawl("https://docs.example.com", &crawl4ai.DeepCrawlOptions{
    Strategy: "bfs",
    MaxDepth: 2,
    MaxURLs:  50,
    Wait:     true,
})
```

### Job Management

```go
// List jobs
jobs, err := crawler.ListJobs(&crawl4ai.ListJobsOptions{
    Status: "completed",
    Limit:  10,
})

// Get job status
job, err := crawler.GetJob(jobID, false)

// Wait for job
job, err := crawler.WaitJob(jobID, 2*time.Second, 5*time.Minute, true)

// Cancel job
err := crawler.CancelJob(jobID)
```

## OSS Compatibility

The SDK provides `Arun()` and `ArunMany()` aliases for seamless migration:

```go
// These are equivalent
result, err := crawler.Run(url, nil)
result, err := crawler.Arun(url, nil)

result, err := crawler.RunMany(urls, opts)
result, err := crawler.ArunMany(urls, opts)
```

## Environment Variables

```bash
export CRAWL4AI_API_KEY=sk_live_...
```

```go
// API key auto-loaded from environment
crawler, err := crawl4ai.NewAsyncWebCrawler(crawl4ai.CrawlerOptions{})
```

## Error Handling

```go
import "github.com/unclecode/crawl4ai-cloud/go/pkg/crawl4ai"

result, err := crawler.Run(url, nil)
if err != nil {
    switch e := err.(type) {
    case *crawl4ai.AuthenticationError:
        fmt.Println("Invalid API key")
    case *crawl4ai.RateLimitError:
        fmt.Printf("Rate limited. Retry after %ds\n", e.RetryAfter())
    case *crawl4ai.QuotaExceededError:
        fmt.Println("Quota exceeded")
    case *crawl4ai.NotFoundError:
        fmt.Println("Resource not found")
    default:
        fmt.Printf("Error: %v\n", err)
    }
}
```

## Types

```go
// Main types
type AsyncWebCrawler struct { ... }
type CrawlResult struct { ... }
type CrawlJob struct { ... }
type MarkdownResult struct { ... }

// Config types
type CrawlerRunConfig struct { ... }
type BrowserConfig struct { ... }
type ProxyConfig struct { ... }

// Options types
type RunOptions struct { ... }
type RunManyOptions struct { ... }
type DeepCrawlOptions struct { ... }
```

## Links

- [Cloud API](https://api.crawl4ai.com) - Get your API key
- [Documentation](https://docs.crawl4ai.com)
- [OSS Repository](https://github.com/unclecode/crawl4ai) - Self-hosted option
- [Discord](https://discord.gg/jP8KfhDhyN) - Community & support

## License

Apache 2.0
