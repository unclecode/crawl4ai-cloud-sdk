# Crawl4AI Cloud SDK

The fastest way to turn any URL into markdown, screenshots, structured data, or a multi-entity table.

[![PyPI version](https://badge.fury.io/py/crawl4ai-cloud-sdk.svg)](https://badge.fury.io/py/crawl4ai-cloud-sdk)
[![Python Version](https://img.shields.io/pypi/pyversions/crawl4ai-cloud-sdk)](https://pypi.org/project/crawl4ai-cloud-sdk/)

## What's new in 0.7.0

- **`scrape()` + `scrape_many()`** are the canonical methods (replacing `markdown` / `markdown_many`). Same shape, same response — they hit the new `/v1/scrape(/async)` endpoints.
- `markdown()` / `markdown_many()` are kept as **deprecated aliases** that route to `/v1/scrape` and emit a `DeprecationWarning`. Removed in 0.8.0.
- **`extract_many()`** signature fixed: now takes `url + extra_urls=` (was `urls=[]`, which the server rejected). `method="auto"` is now allowed for batch.
- **`sources=`** kwarg added to `scan()` and `map()` (`"primary"` | `"extended"`). Replaces `mode=`, which is kept as a deprecated alias.
- **`crawl_site()` and `deep_crawl()`** target deprecated endpoints (`/v1/crawl/site`, `/v1/crawl/deep`). They still work and emit a `DeprecationWarning`. **Migrate to the composable scan + scrape/extract chain** — see below.

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

        # Scrape a page → clean markdown (+ optional links/media/metadata/tables)
        page = await crawler.scrape("https://example.com")
        print(page.markdown)

        # Take a full-page screenshot
        shot = await crawler.screenshot("https://example.com")
        # shot.screenshot is base64-encoded PNG

        # Extract structured data — AUTO picks css_schema vs llm
        job = await crawler.extract_many(
            url="https://news.ycombinator.com",
            method="auto",
            query="get each story title, URL, and points",
            wait=True,
        )
        for r in job.results:
            print(r.data)

        # Discover all URLs on a domain
        sitemap = await crawler.map("https://docs.python.org", sources="primary")
        for u in sitemap.urls[:10]:
            print(u.url)

asyncio.run(main())
```

## Crawl a whole site — composable two-step

There's no single bundled "crawl this site" method anymore. Compose it:

```python
async with AsyncWebCrawler(api_key="sk_live_...") as crawler:

    # Step 1 — discover URLs (criteria + scan.patterns narrow the result)
    scan = await crawler.scan(
        "https://books.toscrape.com",
        criteria="all book detail pages",
        max_urls=20,
    )
    urls = [u.url for u in scan.urls]

    # Step 2A — pipe to scrape_many for markdown
    md_job = await crawler.scrape_many(urls, strategy="http", wait=True)

    # Step 2B — OR pipe to extract_many for structured fields. The base url is
    # the schema TEMPLATE — schema is generated once, then re-applied across
    # extra_urls for free in css_schema mode (10-100× cheaper than per-page LLM).
    base, *rest = urls
    ex_job = await crawler.extract_many(
        url=base,
        extra_urls=rest,
        method="auto",
        query="book title, price, rating",
        wait=True,
    )
    for r in ex_job.results:
        print(r.data)
```

## Wrapper API Reference

| Method | What it does | Endpoint |
|--------|-------------|----------|
| `scrape(url)` | Fetch a page → clean markdown + optional `include=[links\|media\|metadata\|tables]` | `POST /v1/scrape` |
| `scrape_many(urls=[...])` | Async batch scrape (≤100 URLs) | `POST /v1/scrape/async` |
| `screenshot(url)` | Base64 PNG (and optional PDF) | `POST /v1/screenshot` |
| `screenshot_many(urls=[...])` | Async batch screenshot | `POST /v1/screenshot/async` |
| `extract(url, query=...)` | Sync extract — single URL | `POST /v1/extract` |
| `extract_many(url, extra_urls=[...])` | Async extract — base URL + followers, schema reused for free in css_schema mode | `POST /v1/extract/async` |
| `scan(url, sources=, criteria=, scan=)` | URL discovery — sources=primary/extended, optional AI criteria | `POST /v1/scan` |
| `map(url, sources=)` | Simpler URL discovery wrapper (no criteria) | `POST /v1/map` |
| `enrich(query=…)` | Multi-entity table from a brief, list of entities, or list of URLs (multi-phase) | `POST /v1/enrich/async` |
| `configure(...)` | _(internal preview, surfaced later)_ — natural-language → ready-to-POST body | `POST /v1/configure` |

Each method returns a typed response (`MarkdownResponse`, `ScreenshotResponse`, `ExtractResponse`, `MapResponse`, `ScanResult`, `WrapperJob`, `EnrichJobStatus`).

### Deprecated (still work, emit warnings)

| Method | Migration |
|---|---|
| `markdown(url)` | Use `scrape(url)`. |
| `markdown_many(urls)` | Use `scrape_many(urls)`. |
| `crawl_site(url, criteria=…, extract=…)` | Use `scan(url, criteria=…)` + `extract_many(url, extra_urls=…)`. See "Crawl a whole site" above. |
| `deep_crawl(url, …)` | Use `scan(url, scan={"mode": "deep"})` + the chain above. |
| `scan(url, mode=…)` | Use `sources="primary"` or `sources="extended"`. |
| `map(url, mode=…)` | Same — `sources=`. |

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

### Enrich v2 (v0.6.0)

Multi-phase enrichment. Give a brief, a list of entities, or a list of URLs and get back a structured table with per-field provenance, certainty, and disputed-value markers.

The job walks through phases: `queued → planning → plan_ready → resolving_urls → urls_ready → extracting → merging → completed`. Defaults `auto_confirm_plan=True, auto_confirm_urls=True` make it run straight through (one-shot). Set either to `False` for human-in-loop review and resume via `resume_enrich_job(...)`.

```python
async with AsyncWebCrawler(api_key="sk_live_...") as crawler:

    # 1. Agent one-shot — give a brief, get a table back
    result = await crawler.enrich(
        query="licensed nurseries in North York Toronto with extended hours",
        country="ca",
        top_k_per_entity=3,
    )
    for row in result.rows:
        print(row.input_key, row.fields)
        # certainty + sources are per-field
        for f, c in row.certainty.items():
            print(f"  {f}: {c:.2f}  (from {row.sources[f]['url']})")
    print(f"Crawls: {result.usage.crawls}, Searches: {result.usage.searches}")
    print(f"LLM totals: {result.usage.llm_totals}")

    # 2. Pre-resolved URLs — skip planning + URL resolution
    result = await crawler.enrich(
        urls=["https://example.com/a", "https://example.com/b"],
        features=["price", "hours"],   # string shortcut: same as [{"name": "price"}, ...]
    )

    # 3. Human review flow — pause for editing the plan
    job = await crawler.enrich(
        query="best Italian restaurants in Brooklyn",
        country="us",
        auto_confirm_plan=False,
        auto_confirm_urls=False,
        wait=False,
    )

    # Wait for the planning phase to land
    job = await crawler.wait_enrich_job(job.job_id, until="plan_ready")
    print(job.plan.entities, job.plan.features)

    # Edit and resume — the server applies your edits then advances
    await crawler.resume_enrich_job(
        job.job_id,
        entities=[{"name": "Lucali"}, {"name": "Roberta's"}],
        features=[{"name": "address"}, {"name": "hours"}],
    )

    # Wait again for the URL-resolution pause, then resume to completion
    job = await crawler.wait_enrich_job(job.job_id, until="urls_ready")
    await crawler.resume_enrich_job(job.job_id)   # accept server's URL picks
    final = await crawler.wait_enrich_job(job.job_id)

    # 4. Live progress via SSE
    async for event in crawler.stream_enrich_job(job.job_id):
        if event.type == "phase":  print("→", event.status)
        elif event.type == "row":  print("✓", event.row.input_key)
        elif event.type == "complete": break

    # Job management
    jobs = await crawler.list_enrich_jobs(limit=5)
    await crawler.cancel_enrich_job(job.job_id)
```

**Vocabulary:**
- **Entity** — one row identifier (e.g. `"Franklin Barbecue"`).
- **Criterion** — a search-side filter when finding URLs per entity (`{"text": "Austin TX", "kind": "location"}`).
- **Feature** — one extraction column read off each crawled page.

**Per-purpose usage** is reported in `result.usage.llm_tokens_by_purpose` with five buckets: `plan_intent`, `url_plan`, `paywall_classify`, `extract`, `merge_tiebreak` (only buckets that ran appear).

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
