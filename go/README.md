# Crawl4AI Cloud SDK for Go

The official Go SDK for [Crawl4AI Cloud](https://api.crawl4ai.com). One method per task, idiomatic error handling, zero dependencies beyond the standard library.

> **v0.3.0** -- This SDK is for **Crawl4AI Cloud** (api.crawl4ai.com), the managed cloud service.
> For the self-hosted open-source version, see [github.com/unclecode/crawl4ai](https://github.com/unclecode/crawl4ai).

[![Go Reference](https://pkg.go.dev/badge/github.com/unclecode/crawl4ai-cloud-sdk/go.svg)](https://pkg.go.dev/github.com/unclecode/crawl4ai-cloud-sdk/go)

## Install

```bash
go get github.com/unclecode/crawl4ai-cloud-sdk/go@v0.3.0
```

Set your API key (sign up at [api.crawl4ai.com](https://api.crawl4ai.com)):

```bash
export CRAWL4AI_API_KEY=sk_live_...
```

## Quick Examples

### Get Markdown

```go
c, err := crawl4ai.NewAsyncWebCrawler(crawl4ai.CrawlerOptions{})
if err != nil {
    log.Fatal(err)
}

md, err := c.Markdown("https://example.com", nil)
if err != nil {
    log.Fatal(err)
}

fmt.Println(md.Markdown)
```

### Take a Screenshot

```go
ss, err := c.Screenshot("https://example.com", nil)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Base64 length: %d\n", len(ss.Screenshot))
```

### Extract Structured Data

```go
data, err := c.Extract("https://books.toscrape.com", &crawl4ai.ExtractOptions{
    Query: "extract all books with title and price",
})
if err != nil {
    log.Fatal(err)
}

for _, item := range data.Data {
    fmt.Printf("%s - %s\n", item["title"], item["price"])
}
```

### Discover All URLs on a Domain

```go
maxURLs := 50
result, err := c.Map("https://crawl4ai.com", &crawl4ai.MapOptions{
    MaxURLs: &maxURLs,
})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Found %d URLs on %s\n", result.TotalUrls, result.Domain)
for _, u := range result.URLs {
    fmt.Println(u.URL)
}
```

### Crawl an Entire Site

```go
site, err := c.CrawlSite("https://docs.example.com", &crawl4ai.SiteCrawlOptions{
    MaxPages:  100,
    Discovery: "map",
    Strategy:  "browser",
})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Job %s queued, %d URLs discovered\n", site.JobID, site.DiscoveredURLs)
```

## Wrapper API Reference

| Method | Endpoint | Returns | What it does |
|--------|----------|---------|--------------|
| `Markdown(url, opts)` | `POST /v1/markdown` | `*MarkdownResponse` | Clean markdown, optionally with links/media/metadata |
| `Screenshot(url, opts)` | `POST /v1/screenshot` | `*ScreenshotResponse` | Full-page screenshot (base64) or PDF |
| `Extract(url, opts)` | `POST /v1/extract` | `*ExtractResponse` | Structured data via auto, CSS, or LLM extraction |
| `Map(url, opts)` | `POST /v1/map` | `*MapResponse` | Discover all URLs on a domain (sitemap + crawl) |
| `CrawlSite(url, opts)` | `POST /v1/crawl/site` | `*SiteCrawlResponse` | Crawl an entire site (always async, returns job ID) |

Every wrapper method accepts `nil` for options to use sensible defaults.

### Markdown Options

```go
md, err := c.Markdown("https://example.com", &crawl4ai.MarkdownOptions{
    Strategy:  "browser",            // "browser" (default) or "http"
    Include:   []string{"links", "media", "metadata"},
    BypassCache: true,
    CrawlerConfig: map[string]interface{}{
        "css_selector": "article.main",
    },
})
```

### Screenshot Options

```go
ss, err := c.Screenshot("https://example.com", &crawl4ai.ScreenshotOptions{
    PDF:     true,                    // Generate PDF instead of PNG
    WaitFor: "css:.loaded",           // Wait for element before capture
})
```

### Extract Options

```go
data, err := c.Extract("https://example.com", &crawl4ai.ExtractOptions{
    Method:   "auto",                 // "auto" (default), "css", or "llm"
    Strategy: "http",                 // "http" (default) or "browser"
    Query:    "extract product names and prices",
    Schema: map[string]interface{}{   // Optional JSON schema
        "type": "array",
        "items": map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "name":  map[string]interface{}{"type": "string"},
                "price": map[string]interface{}{"type": "number"},
            },
        },
    },
})
```

### Map Options

```go
maxURLs := 200
threshold := 0.5
result, err := c.Map("https://example.com", &crawl4ai.MapOptions{
    MaxURLs:           &maxURLs,
    IncludeSubdomains: true,
    Query:             "blog posts",
    ScoreThreshold:    &threshold,
})
```

### CrawlSite Options

```go
maxDepth := 3
site, err := c.CrawlSite("https://docs.example.com", &crawl4ai.SiteCrawlOptions{
    MaxPages:  200,
    Discovery: "map",               // "map" (default) or "crawl"
    Strategy:  "browser",           // "browser" (default) or "http"
    MaxDepth:  &maxDepth,
    Pattern:   "/docs/*",
    Include:   []string{"links"},
    Priority:  10,                  // 1 (low), 5 (default), 10 (high)
    WebhookURL: "https://example.com/webhook",
})
```

## Job Management

`CrawlSite` always runs async. The other wrapper methods can also produce async jobs for batch operations. Use these methods to track and manage them:

```go
// Check status of a markdown job
job, err := c.GetMarkdownJob(jobID)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Status: %s, Progress: %d%%\n", job.Status, job.ProgressPercent)

// Check status of other job types
job, err = c.GetScreenshotJob(jobID)
job, err = c.GetExtractJob(jobID)

// Cancel jobs
err = c.CancelMarkdownJob(jobID)
err = c.CancelScreenshotJob(jobID)
err = c.CancelExtractJob(jobID)
```

The `WrapperJob` struct provides:

```go
type WrapperJob struct {
    JobID           string
    Status          string   // "pending", "running", "completed", "partial", "failed", "cancelled"
    Progress        *WrapperJobProgress
    ProgressPercent int
    URLsCount       int
    Error           string
    CreatedAt       string
    CompletedAt     string
}

job.IsComplete() // true for completed, partial, failed, cancelled
```

## Power User: CrawlerConfig and BrowserConfig

All wrapper methods accept `CrawlerConfig` and `BrowserConfig` maps for fine-grained control over the crawl engine. These are pass-through maps; the SDK strips cloud-controlled fields automatically.

```go
md, err := c.Markdown("https://spa-site.com", &crawl4ai.MarkdownOptions{
    Strategy: "browser",
    CrawlerConfig: map[string]interface{}{
        "wait_for":        "css:.content-loaded",
        "js_code":         "window.scrollTo(0, document.body.scrollHeight)",
        "scan_full_page":  true,
        "page_timeout":    30000,
    },
    BrowserConfig: map[string]interface{}{
        "viewport_width":  1920,
        "viewport_height": 1080,
        "user_agent":      "MyBot/1.0",
    },
})
```

All wrapper methods also accept a `Proxy` map:

```go
ss, err := c.Screenshot("https://geo-restricted.com", &crawl4ai.ScreenshotOptions{
    Proxy: map[string]interface{}{
        "mode":    "residential",
        "country": "US",
    },
})
```

## Full Power Mode

For advanced use cases (custom extraction strategies, batch crawls with polling, deep recursive crawls), the SDK provides the original low-level methods.

### Single URL Crawl

```go
result, err := c.Run("https://example.com", &crawl4ai.RunOptions{
    Config: &crawl4ai.CrawlerRunConfig{
        Screenshot:         true,
        WordCountThreshold: 10,
    },
    BrowserConfig: &crawl4ai.BrowserConfig{
        ViewportWidth: 1920,
    },
    Proxy: "datacenter",
})
if err != nil {
    log.Fatal(err)
}
fmt.Println(result.Markdown.RawMarkdown)
```

### Batch Crawl

```go
urls := []string{"https://example.com", "https://httpbin.org/html"}

// Fire and forget (returns job)
result, err := c.RunMany(urls, &crawl4ai.RunManyOptions{Priority: 10})
fmt.Printf("Job: %s\n", result.Job.JobID)

// Or wait for completion
result, err = c.RunMany(urls, &crawl4ai.RunManyOptions{Wait: true, Timeout: 5 * time.Minute})
```

### Deep Crawl

```go
deep, err := c.DeepCrawl("https://docs.example.com", &crawl4ai.DeepCrawlOptions{
    Strategy: "bfs",
    MaxDepth: 2,
    MaxURLs:  50,
    Wait:     true,
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Discovered %d URLs\n", deep.DeepResult.DiscoveredCount)
```

### Job Polling (for RunMany / DeepCrawl)

```go
// Poll until complete
job, err := c.WaitJob(jobID, 2*time.Second, 5*time.Minute)

// Or check manually
job, err := c.GetJob(jobID)
if job.IsComplete() {
    fmt.Printf("Done: %s\n", job.Status)
}

// List recent jobs
jobs, err := c.ListJobs(&crawl4ai.ListJobsOptions{Status: "completed", Limit: 10})

// Cancel
err = c.CancelJob(jobID)
```

## Error Handling

All errors are typed. Use type assertions to handle specific cases:

```go
result, err := c.Markdown(url, nil)
if err != nil {
    switch e := err.(type) {
    case *crawl4ai.AuthenticationError:
        log.Fatal("Bad API key")
    case *crawl4ai.RateLimitError:
        log.Printf("Rate limited, retry after %ds", e.RetryAfter())
    case *crawl4ai.QuotaExceededError:
        log.Fatal("Credits exhausted")
    case *crawl4ai.ValidationError:
        log.Printf("Bad request: %s", e.Message)
    case *crawl4ai.NotFoundError:
        log.Printf("Not found: %s", e.Message)
    case *crawl4ai.TimeoutError:
        log.Printf("Timed out: %s", e.Message)
    case *crawl4ai.ServerError:
        log.Printf("Server error (%d): %s", e.StatusCode, e.Message)
    case *crawl4ai.CloudError:
        log.Printf("API error (%d): %s", e.StatusCode, e.Message)
    default:
        log.Printf("Unexpected: %v", err)
    }
}
```

All error types embed `*CloudError`, which carries `StatusCode`, `Message`, `Response` (raw JSON body), and `Headers`.

## Links

- [Cloud Dashboard](https://api.crawl4ai.com) -- Sign up and get your API key
- [Cloud API Docs](https://api.crawl4ai.com/docs) -- Full API reference
- [Go Reference](https://pkg.go.dev/github.com/unclecode/crawl4ai-cloud-sdk/go) -- Godoc
- [OSS Repository](https://github.com/unclecode/crawl4ai) -- Self-hosted option
- [Discord](https://discord.gg/jP8KfhDhyN) -- Community and support

## License

Apache 2.0
