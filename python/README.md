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
| `map(url)` | Simple URL discovery on a domain (always sync) | `POST /v1/map` |
| `scan(url, criteria=...)` | **AI-assisted** URL discovery with plain-English criteria + map/deep routing | `POST /v1/scan` |
| `crawl_site(url, criteria=..., extract=...)` | **AI-assisted** full site crawl -- LLM generates scan config + extraction schema | `POST /v1/crawl/site` |
| `enrich(urls, schema)` | Build a data table from URLs -- per-URL enrichment with depth + search | `POST /v1/enrich` |

Each method returns a typed response object (`MarkdownResponse`, `ScreenshotResponse`, `ExtractResponse`, `MapResponse`, `ScanResult`, `SiteCrawlResponse`, `EnrichJobStatus`) with relevant status and data fields.

### AI-assisted flows (v0.4.0)

Pass a plain-English `criteria` and let the backend LLM pick scan mode, URL patterns, filters, and scorers. Pair with `extract` on `crawl_site()` to also auto-generate a CSS extraction schema from a sample URL. The generated config is echoed back so you can see and reuse it.

```python
async with AsyncWebCrawler(api_key="sk_live_...") as crawler:

    # AI-assisted scan — LLM picks map vs deep + generates patterns/query/threshold
    result = await crawler.scan(
        "https://docs.crawl4ai.com",
        criteria="API reference and core docs pages",
        max_urls=50,
    )
    print(f"Mode: {result.mode_used}")                  # "map" or "deep"
    print(f"Found: {result.total_urls} URLs")
    if result.generated_config:
        print(f"AI: {result.generated_config.reasoning}")

    # Explicit deep scan with async polling
    job = await crawler.scan(
        "https://directory.example.com",
        criteria="company profile pages",
        scan={"mode": "deep", "max_depth": 3},
        wait=True,                                       # block until done
        poll_interval=3.0,
    )

    # Flagship: crawl whole site + auto-extract structured data
    job = await crawler.crawl_site(
        "https://books.toscrape.com",
        criteria="all book listing pages",
        max_pages=50,
        strategy="http",
        extract={
            "query": "book title, price, rating",
            "json_example": {"title": "...", "price": "£0.00", "rating": 0},
            "method": "auto",                            # picks CSS schema vs LLM
        },
        include=["links"],                               # drop markdown — extract-only
    )
    print(f"Generated schema: {bool(job.schema_used)}")
    print(f"Method: {job.extraction_method_used}")      # "css_schema" or "llm"

    # Unified polling — one endpoint for scan + crawl phases
    while True:
        status = await crawler.get_site_crawl_job(job.job_id)
        print(f"{status.phase}: {status.progress.urls_crawled}/{status.progress.total}")
        if status.is_complete:
            print(f"Download: {status.download_url}")
            break
        await asyncio.sleep(3)
```

**Config objects** (optional — both `scan` and `extract` accept plain dicts or typed dataclasses):

```python
from crawl4ai_cloud import SiteScanConfig, SiteExtractConfig

scan_cfg = SiteScanConfig(
    mode="auto",                                         # "auto" | "map" | "deep"
    patterns=["*/docs/*", "*/guide/*"],
    scorers={"keywords": ["auth", "oauth"], "optimal_depth": 2},
    max_depth=3,
)

extract_cfg = SiteExtractConfig(
    query="book title, price, rating",
    json_example={"title": "...", "price": "£0.00", "rating": 0},
    method="auto",
)

job = await crawler.crawl_site(
    "https://books.toscrape.com",
    criteria="book listings",
    scan=scan_cfg,
    extract=extract_cfg,
)
```

**Drop markdown with `include`**: if you pass `include=["links", "media"]` without `"markdown"`, the worker force-strips markdown from every result -- saves bandwidth for extract-only crawls.

### Enrich (v0.5.0)

Build a data table from URLs. Define columns, provide URLs, and the pipeline crawls each URL, follows links to find missing fields, and optionally searches Google as a fallback.

```python
async with AsyncWebCrawler(api_key="sk_live_...") as crawler:

    # Basic enrichment -- depth 0, no search
    result = await crawler.enrich(
        urls=["https://kidocode.com", "https://brightchamps.com"],
        schema=[
            {"name": "Company Name"},
            {"name": "Email", "description": "primary contact email"},
            {"name": "Phone", "description": "phone number"},
        ],
        max_depth=0,
        enable_search=False,
    )

    for row in result.rows:
        print(f"{row.url}: {row.fields}")
        # Sources show where each field was found
        for field, src in row.sources.items():
            print(f"  {field}: {src.method} from {src.url}")

    # With depth + search fallback
    result = await crawler.enrich(
        urls=["https://brightchamps.com"],
        schema=[
            {"name": "Company Name"},
            {"name": "Email"},
            {"name": "Phone"},
            {"name": "Address", "description": "HQ or office address"},
        ],
        max_depth=1,           # follow internal links
        max_links=3,           # check up to 3 sub-pages
        enable_search=True,    # Google Search fallback
    )

    row = result.rows[0]
    print(f"Status: {row.status}")   # "complete", "partial", or "failed"
    print(f"Missing: {row.missing}") # fields that couldn't be found

    # Fire-and-forget + manual polling
    job = await crawler.enrich(
        urls=["https://example1.com", "https://example2.com"],
        schema=[{"name": "Title"}],
        wait=False,  # returns immediately with job_id
    )
    print(f"Job: {job.job_id}")

    # Poll
    status = await crawler.get_enrich_job(job.job_id)
    print(f"Progress: {status.progress.completed}/{status.progress.total}")

    # List recent jobs
    jobs = await crawler.list_enrich_jobs(limit=5)

    # Cancel
    await crawler.cancel_enrich_job(job.job_id)
```

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

    # Site crawl is always async — prefer criteria + extract over legacy discovery flag
    site = await crawler.crawl_site(
        "https://docs.example.com",
        criteria="all API reference pages",
        max_pages=100,
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

# Scan jobs (AI-assisted deep scans)
job   = await crawler.get_scan_job(job_id)              # unified status + URLs-so-far
await crawler.cancel_scan_job(job_id)                   # preserves partial results

# Site crawl jobs (unified scan + crawl polling)
job   = await crawler.get_site_crawl_job(job_id)        # phase: scan|crawl|done
# Cancel delegates to the underlying deep crawl job:
await crawler.cancel_deep_crawl(job_id)

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
