# Crawl Examples

This directory contains TypeScript/JavaScript examples for the Crawl4AI Cloud API crawl endpoints.

## Files

### Basic Crawl (Single URL)
- **01_basic_crawl_sdk.ts** - Crawl a single URL using the SDK `AsyncWebCrawler` class
- **01_basic_crawl_http.ts** - Crawl a single URL using direct HTTP requests with `fetch`

### Batch Crawl (Multiple URLs, max 10)
- **02_batch_crawl_sdk.ts** - Crawl multiple URLs using SDK `runMany()` method
- **02_batch_crawl_http.ts** - Batch crawl using HTTP POST to `/v1/crawl/batch`

### Async Crawl (Background jobs)
- **03_async_crawl_sdk.ts** - Create async job with `wait: true` (SDK polls automatically)
- **03_async_crawl_http.ts** - Create async job with manual polling loop
- **04_async_webhook_http.ts** - Create async job with webhook notification (no polling)

## Usage

1. Install dependencies:
   ```bash
   npm install crawl4ai-cloud
   ```

2. Replace `YOUR_API_KEY` with your actual API key

3. Run any example:
   ```bash
   npx ts-node 01_basic_crawl_sdk.ts
   ```

   Or compile and run:
   ```bash
   npx tsc 01_basic_crawl_sdk.ts
   node 01_basic_crawl_sdk.js
   ```

## Strategy Options

All examples support two crawl strategies:

- **"browser"** - Full JavaScript support, slower but handles dynamic content
- **"http"** - Faster, lightweight, no JavaScript execution

## Quick Reference

| Method | Max URLs | Blocking | Best For |
|--------|----------|----------|----------|
| `run()` | 1 | Yes | Single pages |
| `runMany()` | 10 (batch) | Yes | Small batches |
| `runMany()` | Unlimited (async) | Optional | Large batches, background jobs |

## More Examples

See also:
- `/examples/deep_crawl/` - Deep crawl (sitemap discovery and crawling)
