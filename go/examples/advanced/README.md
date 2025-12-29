# Advanced Examples

This directory contains Go examples for advanced Crawl4AI Cloud features.

## Files

### Screenshots and PDFs
- **01_screenshots_sdk.go** - Capture screenshots and generate PDFs with SDK
- **01_screenshots_http.go** - Screenshots and PDFs via HTTP API

### Custom Browser Configuration
- **02_custom_browser_config.go** - Viewport, headers, proxy with SDK

### Error Handling
- **03_error_handling.go** - Handle rate limits, quota errors, retries

### Storage Monitoring
- **04_storage_usage.go** - Monitor and manage storage quota

## Usage

```bash
cd advanced
go run 01_screenshots_sdk.go
```

## Screenshot Options

```go
result, err := crawler.Run(url, &crawl4ai.RunOptions{
    Config: &crawl4ai.CrawlerRunConfig{
        Screenshot: true,           // Capture screenshot
        WaitFor:    ".content",     // Wait for selector
        PDF:        true,           // Generate PDF
    },
})
```

## Custom Browser Options

```go
result, err := crawler.Run(url, &crawl4ai.RunOptions{
    Strategy: "browser",
    BrowserConfig: &crawl4ai.BrowserConfig{
        Viewport: map[string]int{"width": 1920, "height": 1080},
        Headers:  map[string]string{"User-Agent": "CustomBot/1.0"},
    },
    Proxy: map[string]interface{}{"mode": "datacenter"},
})
```

## Error Types

| Error | Retryable | Description |
|-------|-----------|-------------|
| `AuthenticationError` | No | Invalid API key |
| `RateLimitError` | Yes | Too many requests |
| `QuotaExceededError` | No | Daily/storage limit |
| `ValidationError` | No | Invalid parameters |
| `TimeoutError` | Yes | Request timed out |
| `ServerError` | Yes | API server issue |
