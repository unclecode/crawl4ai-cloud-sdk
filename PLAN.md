# Crawl4AI Cloud SDK

Multi-language SDK for Crawl4AI Cloud API that mirrors the OSS library API exactly.

## Vision

Users can copy-paste their existing OSS code, change the import, add an API key, and it works. Zero learning curve.

```python
# OSS (local) - what users have today
from crawl4ai import AsyncWebCrawler, CrawlerRunConfig
async with AsyncWebCrawler() as crawler:
    result = await crawler.arun(url, config=CrawlerRunConfig(...))

# Cloud SDK - nearly identical!
from crawl4ai_cloud import AsyncWebCrawler, CrawlerRunConfig
async with AsyncWebCrawler(api_key="sk_...") as crawler:
    result = await crawler.run(url, config=CrawlerRunConfig(...))
```

## Why This Approach

1. **New users** → Try cloud first → Become paying members
2. **OSS users** → Seamless migration → Just add API key
3. **Multi-language** → Python, Node.js, Go, Rust, Java, C# (future)
4. **Lightweight** → No playwright, no heavy deps (~5-10MB vs ~800MB)
5. **Separate package** → Clean PyPI/NPM/Go modules distribution

## Project Structure

```
crawl4ai-cloud/
├── python/
│   ├── crawl4ai_cloud/
│   │   ├── __init__.py          # Public exports
│   │   ├── crawler.py           # AsyncWebCrawler class
│   │   ├── configs.py           # CrawlerRunConfig, BrowserConfig
│   │   ├── models.py            # CrawlResult, MarkdownResult, etc.
│   │   ├── errors.py            # CloudError hierarchy
│   │   └── _client.py           # Internal HTTP client
│   ├── tests/
│   │   ├── conftest.py
│   │   ├── test_crawl.py
│   │   ├── test_batch.py
│   │   ├── test_deep_crawl.py
│   │   └── ...
│   ├── pyproject.toml
│   └── README.md
│
├── nodejs/
│   ├── src/
│   │   ├── index.ts             # Public exports
│   │   ├── crawler.ts           # AsyncWebCrawler class
│   │   ├── configs.ts           # CrawlerRunConfig, BrowserConfig
│   │   ├── models.ts            # Response types
│   │   └── errors.ts            # Error classes
│   ├── tests/
│   │   └── ...
│   ├── package.json
│   ├── tsconfig.json
│   └── README.md
│
├── go/
│   ├── pkg/
│   │   └── crawl4ai/
│   │       ├── crawler.go
│   │       ├── configs.go
│   │       ├── models.go
│   │       └── errors.go
│   ├── tests/
│   │   └── ...
│   ├── go.mod
│   └── README.md
│
├── .context/                    # Local planning (gitignored)
├── .gitignore
├── PLAN.md
└── README.md
```

## API Design Decisions

### Class Names (Mirror OSS Exactly)
- `AsyncWebCrawler` - main crawler class
- `CrawlerRunConfig` - crawl configuration
- `BrowserConfig` - browser settings
- `CrawlResult` - result object

### Method Names
| Method | Description |
|--------|-------------|
| `run(url, ...)` | Crawl single URL |
| `run_many(urls, ...)` | Crawl multiple URLs |
| `arun(url, ...)` | Alias for `run()` (OSS compatibility) |
| `arun_many(urls, ...)` | Alias for `run_many()` (OSS compatibility) |

**No "a" prefix** - but aliases provided for OSS code compatibility.

### Config Handling
- Copy `CrawlerRunConfig` and `BrowserConfig` from OSS
- Add **sanitizer/normalizer** that strips cloud-unsupported fields silently
- User's OSS code works without modification

### Fields to Sanitize (Remove Silently)

**CrawlerRunConfig:**
- `cache_mode`, `session_id`, `bypass_cache`
- `no_cache_read`, `no_cache_write`

**BrowserConfig:**
- `cdp_url`, `browser_mode`, `use_managed_browser`
- `user_data_dir`, `chrome_channel`

### Proxy Shorthand
```python
# String shorthand
result = await crawler.run(url, proxy="datacenter")

# Full config
result = await crawler.run(url, proxy={"mode": "residential", "country": "US"})
```

### Authentication
```python
# Via constructor
crawler = AsyncWebCrawler(api_key="sk_live_...")

# Via environment variable
# export CRAWL4AI_API_KEY=sk_live_...
crawler = AsyncWebCrawler()  # Auto-reads from env
```

## Cloud API Endpoints

| SDK Method | HTTP Endpoint |
|------------|---------------|
| `run(url)` | `POST /v1/crawl` |
| `run_many(urls)` ≤10 | `POST /v1/crawl/batch` |
| `run_many(urls)` >10 | `POST /v1/crawl/async` |
| `get_job(id)` | `GET /v1/crawl/jobs/{id}` |
| `list_jobs()` | `GET /v1/crawl/jobs` |
| `cancel_job(id)` | `DELETE /v1/crawl/jobs/{id}` |
| `deep_crawl(url)` | `POST /v1/crawl/deep` |
| `context(query)` | `POST /v1/context` |
| `generate_schema(html)` | `POST /v1/schema/generate` |
| `storage()` | `GET /v1/crawl/storage` |

**Base URL:** `https://api.crawl4ai.com`
**Auth Header:** `X-API-Key: sk_live_...`

## Reusable Code from Previous Work

We built a working cloud client in `~/crawl4ai/crawl4ai/cloud/` that can be adapted:

### `_client.py` (from client.py)
- HTTP request handling with retries
- Error mapping (401→AuthError, 429→RateLimit, etc.)
- Async context manager pattern

### `errors.py`
```python
class CloudError(Exception): ...
class AuthenticationError(CloudError): ...
class RateLimitError(CloudError): ...
class QuotaExceededError(CloudError): ...
class NotFoundError(CloudError): ...
class ValidationError(CloudError): ...
class TimeoutError(CloudError): ...
class ServerError(CloudError): ...
```

### `models.py`
- `CrawlResult` / `CloudCrawlResult`
- `MarkdownResult`
- `CrawlJob`, `JobProgress`
- `DeepCrawlResult`
- `ContextResult`
- `GeneratedSchema`
- `StorageUsage`

### `config.py`
- `sanitize_crawler_config()` - strips cloud-controlled fields
- `sanitize_browser_config()` - strips cloud-controlled fields
- `normalize_proxy()` - handles string/dict/object proxy config

## Testing Strategy

### Python (pytest)
```bash
cd python && pytest tests/ -v
```

### Node.js (jest/vitest)
```bash
cd nodejs && npm test
```

### Go (go test)
```bash
cd go && go test ./...
```

### Test Categories
1. **Basic crawl** - `run()` with various configs
2. **Batch crawl** - `run_many()` with wait=True/False
3. **Async jobs** - job management (get, wait, list, cancel)
4. **Deep crawl** - BFS, DFS, best-first strategies
5. **Context API** - PAA expansion
6. **Schema generation** - CSS/XPath schema
7. **Proxy modes** - datacenter, residential, auto
8. **Config sanitization** - OSS config compatibility
9. **Error handling** - all error types

## Implementation Order

### Phase 1: Python SDK
1. Project setup (pyproject.toml, structure)
2. Copy and adapt code from previous work
3. Rename `CloudCrawler` → `AsyncWebCrawler`
4. Add `arun()`/`arun_many()` aliases
5. Copy configs from OSS, add sanitizers
6. Write tests
7. Test against live API

### Phase 2: Node.js SDK
1. Project setup (package.json, TypeScript)
2. Port Python implementation to TypeScript
3. Match API exactly
4. Write tests

### Phase 3: Go SDK
1. Project setup (go.mod)
2. Implement in idiomatic Go
3. Match API as closely as Go allows
4. Write tests

## Package Distribution

| Language | Package Name | Registry |
|----------|--------------|----------|
| Python | `crawl4ai-cloud` | PyPI |
| Node.js | `crawl4ai-cloud` | NPM |
| Go | `github.com/unclecode/crawl4ai-cloud/go` | Go Modules |

## Success Criteria

- [ ] `pip install crawl4ai-cloud` works (~10MB)
- [ ] `from crawl4ai_cloud import AsyncWebCrawler` works
- [ ] OSS code runs with just import change + API key
- [ ] All tests pass for Python
- [ ] Node.js SDK mirrors Python API
- [ ] Go SDK works idiomatically
