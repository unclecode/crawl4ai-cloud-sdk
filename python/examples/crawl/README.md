# Crawl Examples

This directory contains Python examples for the Crawl4AI Cloud API crawl endpoints.

## Files

### Basic Crawl (Single URL)
- **01_basic_crawl_sdk.py** - Crawl a single URL using the SDK `Crawl4AI` class
- **01_basic_crawl_http.py** - Crawl a single URL using direct HTTP requests with `httpx`

### Batch Crawl (Multiple URLs, max 10)
- **02_batch_crawl_sdk.py** - Crawl multiple URLs using SDK `crawl_batch()` method
- **02_batch_crawl_http.py** - Batch crawl using HTTP POST to `/v1/crawl/batch`

### Async Crawl (Background jobs)
- **03_async_crawl_sdk.py** - Create async job with `wait=True` (SDK polls automatically)
- **03_async_crawl_http.py** - Create async job with manual polling loop
- **04_async_webhook_http.py** - Create async job with webhook notification (no polling)

## Usage

1. Install dependencies:
   ```bash
   pip install crawl4ai-cloud httpx
   ```

2. Replace `YOUR_API_KEY` with your actual API key

3. Run any example:
   ```bash
   python 01_basic_crawl_sdk.py
   ```

## Strategy Options

All examples support two crawl strategies:

- **"browser"** - Full JavaScript support, slower but handles dynamic content
- **"http"** - Faster, lightweight, no JavaScript execution

## Quick Reference

| Method | Max URLs | Blocking | Best For |
|--------|----------|----------|----------|
| `crawl()` | 1 | Yes | Single pages |
| `crawl_batch()` | 10 | Yes | Small batches |
| `crawl_async()` | Unlimited | No | Large batches, background jobs |

## More Examples

See also:
- `/examples/deep_crawl/` - Deep crawl (sitemap discovery and crawling)
