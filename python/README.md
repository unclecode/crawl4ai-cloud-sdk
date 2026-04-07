# Crawl4AI Cloud SDK

The fastest way to turn any URL into markdown, screenshots, structured data, or a full site crawl.

[![PyPI version](https://badge.fury.io/py/crawl4ai-cloud-sdk.svg)](https://badge.fury.io/py/crawl4ai-cloud-sdk)
[![Python Version](https://img.shields.io/pypi/pyversions/crawl4ai-cloud-sdk)](https://pypi.org/project/crawl4ai-cloud-sdk/)

## Install

```bash
pip install crawl4ai-cloud-sdk
```

Get your API key at [api.crawl4ai.com](https://api.crawl4ai.com).

## Quick Start

```python
import asyncio
from crawl4ai_cloud import AsyncWebCrawler

async def main():
    async with AsyncWebCrawler(api_key="sk_live_...") as crawler:

        # Get clean markdown from any page
        md = await crawler.markdown("https://example.com")
        print(md.markdown)

        # Take a full-page screenshot
        ss = await crawler.screenshot("https://example.com")
        # ss.screenshot is base64-encoded PNG

        # Extract structured data with natural language
        data = await crawler.extract(
            "https://news.ycombinator.com",
            query="get each story title, URL, and points",
        )
        print(data.data)  # list of dicts

        # Discover all URLs on a domain
        sitemap = await crawler.map("https://docs.python.org")
        for u in sitemap.urls[:10]:
            print(u.url)

        # Crawl an entire site (async, returns job_id)
        job = await crawler.crawl_site(
            "https://docs.example.com",
            max_pages=50,
            wait=True,
        )
        print(f"Done: {job.discovered_urls} pages crawled")

asyncio.run(main())
```

## Wrapper API Reference

| Method | What it does | Endpoint |
|--------|-------------|----------|
| `markdown(url)` | Returns clean markdown (with optional fit/pruning) | `POST /v1/markdown` |
| `screenshot(url)` | Returns base64 screenshot (PNG) and optional PDF | `POST /v1/screenshot` |
| `extract(url, query=...)` | Extracts structured data (auto/LLM/CSS schema) | `POST /v1/extract` |
| `map(url)` | Discovers all URLs on a domain via sitemap + probing | `POST /v1/map` |
| `crawl_site(url)` | Full site crawl with discovery + extraction (async) | `POST /v1/crawl/site` |

Each method returns a typed response object (`MarkdownResponse`, `ScreenshotResponse`, `ExtractResponse`, `MapResponse`, `SiteCrawlResponse`) with `.success`, `.duration_ms`, and `.usage` fields.

## Async / Batch

Every wrapper method has a `_many` variant for processing multiple URLs as an async job.

```python
async with AsyncWebCrawler(api_key="sk_live_...") as crawler:

    # Batch markdown (fire-and-forget)
    job = await crawler.markdown_many(
        ["https://a.com", "https://b.com", "https://c.com"],
    )
    print(f"Job {job.job_id} started, {job.urls_count} URLs queued")

    # Batch markdown (wait for results)
    job = await crawler.markdown_many(urls, wait=True, timeout=120)

    # Batch screenshots
    job = await crawler.screenshot_many(urls, full_page=True, wait=True)

    # Batch extraction (note: method must be "llm" or "schema", not "auto")
    job = await crawler.extract_many(
        urls, method="llm", query="get product name and price", wait=True,
    )

    # Site crawl is always async
    site = await crawler.crawl_site(
        "https://docs.example.com",
        max_pages=100,
        discovery="bfs",
        wait=True,
    )
```

## Job Management

Each wrapper namespace has its own job management methods.

```python
# Markdown jobs
job   = await crawler.get_markdown_job(job_id)
jobs  = await crawler.list_markdown_jobs(status="completed", limit=10)
await crawler.cancel_markdown_job(job_id)

# Screenshot jobs
job   = await crawler.get_screenshot_job(job_id)
jobs  = await crawler.list_screenshot_jobs()
await crawler.cancel_screenshot_job(job_id)

# Extract jobs
job   = await crawler.get_extract_job(job_id)
jobs  = await crawler.list_extract_jobs()
await crawler.cancel_extract_job(job_id)

# Core crawl jobs (from run_many / deep_crawl)
job   = await crawler.get_job(job_id)
jobs  = await crawler.list_jobs(status="running")
await crawler.cancel_job(job_id)
url   = await crawler.download_url(job_id)  # presigned S3 ZIP
```

## Power User: Config Passthrough

All wrapper methods accept `crawler_config` and `browser_config` dicts for full control. These are the same fields you would pass to the core `/v1/crawl` endpoint.

```python
md = await crawler.markdown(
    "https://example.com",
    strategy="browser",
    fit=True,
    include=["links", "media", "tables"],
    crawler_config={
        "css_selector": "article",
        "exclude_external_links": True,
        "wait_for": ".content-loaded",
        "js_code": "window.scrollTo(0, document.body.scrollHeight)",
    },
    browser_config={
        "viewport_width": 1920,
        "viewport_height": 1080,
        "headers": {"Accept-Language": "en-US"},
    },
    proxy="residential",
)
```

Works the same way for `screenshot()`, `extract()`, `map()`, and `crawl_site()`.

## Full Power Mode

For advanced use cases where you need full control over the crawl pipeline, the core methods give you direct access to the `/v1/crawl` endpoint with every configuration option.

### Single URL

```python
from crawl4ai_cloud import CrawlerRunConfig, BrowserConfig

config = CrawlerRunConfig(
    screenshot=True,
    word_count_threshold=10,
    exclude_external_links=True,
    process_iframes=True,
    css_selector="article",
)
browser_config = BrowserConfig(
    viewport_width=1920,
    viewport_height=1080,
)

result = await crawler.run(
    "https://example.com",
    config=config,
    browser_config=browser_config,
    proxy="datacenter",
)
print(result.markdown.raw_markdown)
print(result.screenshot)  # base64
```

### Batch Crawl

```python
job = await crawler.run_many(
    ["https://a.com", "https://b.com"],
    config=config,
    wait=True,
    priority=1,
)
# Results available via download
url = await crawler.download_url(job.id)
```

### Deep Crawl

```python
result = await crawler.deep_crawl(
    "https://docs.example.com",
    strategy="bfs",       # bfs, dfs, best_first, map
    max_depth=3,
    max_urls=100,
    include_patterns=["docs", "api"],
    exclude_patterns=["download"],
    wait=True,
)
```

### Domain Scan

```python
scan = await crawler.scan("https://example.com", mode="deep", max_urls=200)
for url_info in scan.urls:
    print(f"{url_info.url} (score: {url_info.relevance_score})")
```

Full reference: [Cloud API Docs](https://api.crawl4ai.com/docs)

## Configuration

### CrawlerRunConfig

Controls what gets extracted and how pages are processed.

```python
from crawl4ai_cloud import CrawlerRunConfig

config = CrawlerRunConfig(
    css_selector="main",          # target specific elements
    excluded_tags=["nav", "footer"],
    word_count_threshold=10,
    screenshot=True,
    wait_for=".loaded",           # wait for CSS selector
    js_code="document.querySelector('.show-more').click()",
    magic=True,                   # anti-bot mode
)
```

### BrowserConfig

Controls the browser environment.

```python
from crawl4ai_cloud import BrowserConfig

browser = BrowserConfig(
    viewport_width=1920,
    viewport_height=1080,
    user_agent="MyBot/1.0",
    headers={"Authorization": "Bearer token"},
    cookies=[{"name": "session", "value": "abc", "domain": "example.com"}],
    profile_id="my-saved-profile",  # cloud browser profile
)
```

### ProxyConfig

```python
from crawl4ai_cloud import ProxyConfig

# Shorthand (works on all methods)
result = await crawler.markdown(url, proxy="datacenter")
result = await crawler.markdown(url, proxy="residential")

# Full config
proxy = ProxyConfig(mode="residential", country="US", sticky_session=True)
result = await crawler.markdown(url, proxy=proxy)
```

Proxy modes: `"none"` (direct, 1x credits), `"datacenter"` (fast, 2x), `"residential"` (premium, 5x), `"auto"` (smart selection).

## Environment Variables

```bash
export CRAWL4AI_API_KEY=sk_live_...
```

```python
# API key auto-loaded from environment
async with AsyncWebCrawler() as crawler:
    md = await crawler.markdown("https://example.com")
```

## Error Handling

```python
from crawl4ai_cloud import (
    CloudError,
    AuthenticationError,
    RateLimitError,
    QuotaExceededError,
    NotFoundError,
    ValidationError,
    TimeoutError,
    ServerError,
)

try:
    result = await crawler.markdown(url)
except AuthenticationError:
    print("Invalid API key")
except RateLimitError as e:
    print(f"Rate limited. Retry after {e.retry_after}s")
except QuotaExceededError as e:
    print(f"Quota exceeded ({e.quota_type})")
except TimeoutError:
    print("Request timed out")
except ValidationError:
    print("Invalid request parameters")
except ServerError:
    print("Server error, try again later")
except CloudError as e:
    print(f"[{e.status_code}] {e.message}")
```

## Claude Code Plugin

Use Crawl4AI directly inside [Claude Code](https://claude.com/claude-code) with 9 built-in tools.

```
/plugin marketplace add unclecode/crawl4ai-cloud-sdk
/plugin install crawl4ai@crawl4ai-claude-plugins
```

See [plugin README](./claude-plugin/README.md) for details.

## Links

- [Cloud Dashboard](https://api.crawl4ai.com) - Sign up and manage your API key
- [Cloud API Docs](https://api.crawl4ai.com/docs) - Full API reference
- [PyPI](https://pypi.org/project/crawl4ai-cloud-sdk/) - Package page
- [GitHub](https://github.com/unclecode/crawl4ai-cloud-sdk) - Source code
- [OSS Crawl4AI](https://github.com/unclecode/crawl4ai) - Self-hosted option
- [Discord](https://discord.gg/jP8KfhDhyN) - Community and support

## License

Apache 2.0
