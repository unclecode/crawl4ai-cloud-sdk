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

### Discover All URLs on a Domain (simple, sync)

```go
result, err := c.Map("https://crawl4ai.com", &crawl4ai.MapOptions{
    MaxURLs: 50,
})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Found %d URLs on %s\n", result.TotalUrls, result.Domain)
for _, u := range result.URLs {
    fmt.Println(u.URL)
}
```

### Scan URLs (AI-assisted)

Pass a plain-English `Criteria` and let the backend LLM pick scan mode, URL patterns, filters.

```go
result, err := c.Scan("https://docs.crawl4ai.com", &crawl4ai.ScanOptions{
    Criteria: "API reference and core documentation pages",
    MaxUrls:  50,
})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Mode used: %s\n", result.ModeUsed)           // "map" or "deep"
fmt.Printf("Found: %d URLs\n", result.TotalUrls)
if result.GeneratedConfig != nil {
    fmt.Printf("AI reasoning: %s\n", result.GeneratedConfig.Reasoning)
}
```

### Crawl an Entire Site with Auto-Extraction (flagship)

```go
site, err := c.CrawlSite("https://books.toscrape.com", &crawl4ai.SiteCrawlOptions{
    Criteria: "all book listing pages",
    MaxPages: 50,
    Strategy: "http",
    Extract: &crawl4ai.SiteExtractConfig{
        Query:       "book title, price, rating",
        JSONExample: map[string]interface{}{"title": "...", "price": "£0.00", "rating": 0},
        Method:      "auto", // picks CSS schema vs LLM
    },
})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Job %s started\n", site.JobID)
fmt.Printf("Extraction: %s\n", site.ExtractionMethodUsed) // "css_schema" or "llm"
fmt.Printf("Schema generated: %v\n", site.SchemaUsed != nil)

// Unified polling: one endpoint for scan + crawl phases
for {
    status, _ := c.GetSiteCrawlJob(site.JobID)
    fmt.Printf("%s: %d/%d\n", status.Phase, status.Progress.UrlsCrawled, status.Progress.Total)
    if status.IsComplete() {
        if status.DownloadURL != "" {
            fmt.Printf("Download: %s\n", status.DownloadURL)
        }
        break
    }
    time.Sleep(3 * time.Second)
}
```

### Enrich URLs (build a data table)

```go
result, err := crawler.Enrich(
    []string{"https://kidocode.com", "https://brightchamps.com"},
    []crawl4ai.EnrichFieldSpec{
        {Name: "Company Name"},
        {Name: "Email", Description: "primary contact email"},
        {Name: "Phone", Description: "phone number"},
    },
    &crawl4ai.EnrichOptions{
        MaxDepth:     1,
        EnableSearch: true,
        Wait:         true,
        Timeout:      120 * time.Second,
    },
)
if err != nil {
    log.Fatal(err)
}

for _, row := range result.Rows {
    fmt.Printf("%s: %v\n", row.URL, row.Fields)
    for field, src := range row.Sources {
        fmt.Printf("  %s: %s from %s\n", field, src.Method, src.URL)
    }
}

// Fire-and-forget + manual poll
job, _ := crawler.Enrich(urls, schema, &crawl4ai.EnrichOptions{Wait: false})
status, _ := crawler.GetEnrichJob(job.JobID)
jobs, _ := crawler.ListEnrichJobs(5, 0)
_ = crawler.CancelEnrichJob(job.JobID)
```

## Wrapper API Reference

| Method | Endpoint | Returns | What it does |
|--------|----------|---------|--------------|
| `Markdown(url, opts)` | `POST /v1/markdown` | `*MarkdownResponse` | Clean markdown, optionally with links/media/metadata |
| `Screenshot(url, opts)` | `POST /v1/screenshot` | `*ScreenshotResponse` | Full-page screenshot (base64) or PDF |
| `Extract(url, opts)` | `POST /v1/extract` | `*ExtractResponse` | Structured data via auto, CSS, or LLM extraction |
| `Map(url, opts)` | `POST /v1/map` | `*MapResponse` | Simple URL discovery (always sync) |
| `Scan(url, opts)` | `POST /v1/scan` | `*ScanResult` | **AI-assisted** URL discovery with plain-English criteria |
| `CrawlSite(url, opts)` | `POST /v1/crawl/site` | `*SiteCrawlResponse` | **AI-assisted** whole-site crawl (always async) |
| `Enrich(urls, schema, opts)` | `POST /v1/enrich` | `*EnrichJobStatus` | Per-URL data enrichment with depth + search |

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

### Map Options (legacy, sync-only)

```go
threshold := 0.5
result, err := c.Map("https://example.com", &crawl4ai.MapOptions{
    MaxURLs:           200,
    IncludeSubdomains: true,
    Query:             "blog posts",
    ScoreThreshold:    &threshold,
})
```

### Scan Options (AI-assisted)

```go
// AI picks everything from criteria
result, err := c.Scan("https://docs.crawl4ai.com", &crawl4ai.ScanOptions{
    Criteria: "API reference pages",
    MaxUrls:  50,
})

// Explicit overrides on top of LLM output
maxDepth := 3
result, err := c.Scan("https://example.com", &crawl4ai.ScanOptions{
    Criteria: "product pages",
    Scan: &crawl4ai.SiteScanConfig{
        Mode:     "auto",  // "auto" | "map" | "deep"
        Patterns: []string{"*/p/*"},
        MaxDepth: &maxDepth,
    },
})

// Async deep scan (waits for completion)
result, err := c.Scan("https://directory.example.com", &crawl4ai.ScanOptions{
    Criteria:     "company profile pages",
    Scan:         &crawl4ai.SiteScanConfig{Mode: "deep", MaxDepth: &maxDepth},
    Wait:         true,
    PollInterval: 3 * time.Second,
})
```

### CrawlSite Options (AI-assisted flagship)

```go
site, err := c.CrawlSite("https://books.toscrape.com", &crawl4ai.SiteCrawlOptions{
    // AI-assisted fields (new in v0.4.0)
    Criteria: "all book listing pages",
    Scan: &crawl4ai.SiteScanConfig{                   // optional explicit override
        Mode: "auto",
    },
    Extract: &crawl4ai.SiteExtractConfig{
        Query:       "book title, price, rating",
        JSONExample: map[string]interface{}{"title": "...", "price": "£0.00", "rating": 0},
        Method:      "auto",                           // "auto" | "llm" | "schema"
    },
    Include: []string{"markdown", "links"},            // drop "markdown" to strip it
    // Standard fields
    MaxPages: 50,
    Strategy: "http",
    // Legacy fields (still supported)
    Discovery: "map",
    Pattern:   "/docs/*",
    // Waiting
    Wait:         true,
    PollInterval: 5 * time.Second,
    Timeout:      300 * time.Second,
})

// site.GeneratedConfig         -- LLM's scan + extract decisions
// site.ExtractionMethodUsed    -- "css_schema" | "llm"
// site.SchemaUsed              -- generated CSS schema (reusable!)
```

**Drop markdown with `Include`**: pass `Include: []string{"links", "media"}` (without `"markdown"`) and the worker strips markdown from every result.

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

// Scan jobs (AI-assisted deep scans)
scanJob, err := c.GetScanJob(jobID)                    // URLs discovered so far + progress
cancelled, err := c.CancelScanJob(jobID)               // preserves partial results

// Site crawl jobs (unified scan + crawl polling)
status, err := c.GetSiteCrawlJob(jobID)                // phase: "scan" | "crawl" | "done"
fmt.Printf("%s: %d/%d\n", status.Phase, status.Progress.UrlsCrawled, status.Progress.Total)
// status.DownloadURL is set when phase == "done" && status.Status == "completed"
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

`ScanJobStatus` and `SiteCrawlJobStatus` both provide an `IsComplete()` method with the same semantics.

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
