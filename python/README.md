# Crawl4AI Cloud SDK for Python

Lightweight Python SDK for [Crawl4AI Cloud](https://api.crawl4ai.com). Mirrors the OSS API exactly.

> **Note:** This SDK is for **Crawl4AI Cloud** (api.crawl4ai.com), the managed cloud service. For the self-hosted open-source version, see [github.com/unclecode/crawl4ai](https://github.com/unclecode/crawl4ai).

[![PyPI version](https://badge.fury.io/py/crawl4ai-cloud-sdk.svg)](https://badge.fury.io/py/crawl4ai-cloud-sdk)
[![Python Version](https://img.shields.io/pypi/pyversions/crawl4ai-cloud-sdk)](https://pypi.org/project/crawl4ai-cloud-sdk/)

## Installation

```bash
pip install crawl4ai-cloud-sdk
```

## Get Your API Key

1. Go to [api.crawl4ai.com](https://api.crawl4ai.com)
2. Sign up and get your API key

## Quick Start

```python
import asyncio
from crawl4ai_cloud import AsyncWebCrawler

async def main():
    async with AsyncWebCrawler(api_key="sk_live_...") as crawler:
        result = await crawler.run("https://example.com")
        print(result.markdown.raw_markdown)

asyncio.run(main())
```

## Features

### Single URL Crawl

```python
result = await crawler.run("https://example.com")
print(result.success)
print(result.markdown.raw_markdown)
print(result.html)
```

### Batch Crawl

```python
urls = ["https://example.com", "https://httpbin.org/html"]

# Wait for results
results = await crawler.run_many(urls, wait=True)
for r in results:
    print(f"{r.url}: {r.success}")

# Fire and forget (returns job)
job = await crawler.run_many(urls, wait=False)
print(f"Job ID: {job.id}")
```

### Configuration

```python
from crawl4ai_cloud import CrawlerRunConfig, BrowserConfig

config = CrawlerRunConfig(
    word_count_threshold=10,
    exclude_external_links=True,
    screenshot=True,
)

browser_config = BrowserConfig(
    viewport_width=1920,
    viewport_height=1080,
)

result = await crawler.run(
    "https://example.com",
    config=config,
    browser_config=browser_config,
)
```

### Proxy Support

```python
# Shorthand
result = await crawler.run(url, proxy="datacenter")
result = await crawler.run(url, proxy="residential")

# Full config
result = await crawler.run(url, proxy={
    "mode": "residential",
    "country": "US"
})
```

### Deep Crawl

```python
result = await crawler.deep_crawl(
    "https://docs.example.com",
    strategy="bfs",
    max_depth=2,
    max_urls=50,
    wait=True,
)
```

### Job Management

```python
# List jobs
jobs = await crawler.list_jobs(status="completed", limit=10)

# Get job status
job = await crawler.get_job(job_id)

# Wait for job
job = await crawler.wait_job(job_id, poll_interval=2.0)

# Cancel job
await crawler.cancel_job(job_id)
```

## Migration from OSS

Zero learning curve — your existing code works:

```python
# Before (OSS)
from crawl4ai import AsyncWebCrawler
async with AsyncWebCrawler() as crawler:
    result = await crawler.arun(url)

# After (Cloud)
from crawl4ai_cloud import AsyncWebCrawler
async with AsyncWebCrawler(api_key="sk_...") as crawler:
    result = await crawler.run(url)  # arun() also works!
```

## Environment Variables

```bash
export CRAWL4AI_API_KEY=sk_live_...
```

```python
# API key auto-loaded from environment
crawler = AsyncWebCrawler()
```

## Error Handling

```python
from crawl4ai_cloud import (
    CloudError,
    AuthenticationError,
    RateLimitError,
    QuotaExceededError,
    NotFoundError,
)

try:
    result = await crawler.run(url)
except AuthenticationError:
    print("Invalid API key")
except RateLimitError as e:
    print(f"Rate limited. Retry after {e.retry_after}s")
except QuotaExceededError:
    print("Quota exceeded")
```

## Links

- [Cloud Dashboard](https://api.crawl4ai.com) - Sign up & get your API key
- [Cloud API Docs](https://api.crawl4ai.com/docs) - Full API reference
- [OSS Repository](https://github.com/unclecode/crawl4ai) - Self-hosted option
- [Discord](https://discord.gg/jP8KfhDhyN) - Community & support

## License

Apache 2.0
