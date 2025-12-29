# Crawl Examples

This directory contains Go examples for the Crawl4AI Cloud API crawl endpoints.

## Files

### Basic Crawl (Single URL)
- **01_basic_crawl_sdk.go** - Crawl a single URL using the SDK
- **01_basic_crawl_http.go** - Crawl a single URL using direct HTTP requests

### Batch Crawl (Multiple URLs, max 10)
- **02_batch_crawl_sdk.go** - Crawl multiple URLs using SDK `RunMany()`
- **02_batch_crawl_http.go** - Batch crawl using HTTP POST to `/v1/crawl/batch`

### Async Crawl (Background jobs)
- **03_async_crawl_sdk.go** - Create async job with `Wait: true` (SDK polls automatically)
- **03_async_crawl_http.go** - Create async job with manual polling loop
- **04_async_webhook_http.go** - Create async job with webhook notification (no polling)

## Usage

1. Replace `YOUR_API_KEY` with your actual API key

2. Run any example:
   ```bash
   cd crawl
   go run 01_basic_crawl_sdk.go
   ```

## Strategy Options

All examples support two crawl strategies:

- **"browser"** - Full JavaScript support, slower but handles dynamic content
- **"http"** - Faster, lightweight, no JavaScript execution

## Quick Reference

| Method | Max URLs | Blocking | Best For |
|--------|----------|----------|----------|
| `Run()` | 1 | Yes | Single pages |
| `RunMany()` | 10 (batch) | Yes | Small batches |
| `RunMany()` | Unlimited (async) | Optional | Large batches, background jobs |

## More Examples

See also:
- `/examples/deep_crawl/` - Deep crawl (sitemap discovery and crawling)
