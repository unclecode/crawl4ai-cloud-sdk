---
name: crawl4ai
description: Web crawling, screenshots, data extraction, URL discovery. Use when the user asks to crawl, scrape, extract data from, screenshot, or discover URLs on a website.
argument-hint: [url or instruction]
---

# Crawl4AI

Crawl, extract, screenshot, or discover URLs from any website via the Crawl4AI Cloud API.

## Scripts

Run these directly. They handle auth, errors, and output JSON.

```
SCRIPTS="${CLAUDE_PLUGIN_ROOT}/scripts"

python3 $SCRIPTS/crawl.py markdown URL [options]     # Get clean markdown
python3 $SCRIPTS/crawl.py screenshot URL [options]   # Capture screenshot/PDF
python3 $SCRIPTS/crawl.py extract URL [options]      # Extract structured data
python3 $SCRIPTS/crawl.py map URL [options]           # Discover all URLs on domain
python3 $SCRIPTS/crawl.py site URL [options]          # Crawl entire website
python3 $SCRIPTS/crawl.py crawl URL [options]         # Full power /v1/crawl
python3 $SCRIPTS/crawl.py async TYPE URLS [options]   # Batch async job

python3 $SCRIPTS/poll-job.py JOB_ID --type TYPE       # Poll async job
bash $SCRIPTS/health.sh                               # API health + credits
bash $SCRIPTS/test-key.sh                              # Validate API key
```

## Decision Logic

```
User wants...                    Use this
--------------------------------------------------------------
Clean text/markdown              crawl.py markdown URL
Screenshot or PDF                crawl.py screenshot URL
Structured data (JSON)           crawl.py extract URL --query "..."
Find all URLs on a domain        crawl.py map URL
Crawl all pages of a site        crawl.py site URL --max-pages N
Full control (40+ params)        crawl.py crawl URL --crawler-config '{...}'
Batch multiple URLs              crawl.py async TYPE URL1,URL2 --wait
```

## Strategy Selection

| Strategy | Speed | JavaScript | When to use |
|----------|-------|:----------:|-------------|
| `--strategy http` | Fast (100-500ms) | No | Blogs, news, static sites |
| `--strategy browser` | Slower (2-5s) | Yes | SPAs, React/Vue, lazy-loaded content |

Default is `browser`. Use `http` when the page doesn't need JavaScript.

## The 5 Wrapper Endpoints

### 1. Markdown
```bash
python3 $SCRIPTS/crawl.py markdown "https://example.com"
python3 $SCRIPTS/crawl.py markdown "https://example.com" --strategy http --fit true
python3 $SCRIPTS/crawl.py markdown "https://example.com" --include links,media,metadata
```
Key params: `--strategy`, `--fit` (default true, prunes nav/footer), `--include`, `--bypass-cache`
Response: `success`, `markdown`, `fit_markdown`, `links`, `media`, `metadata`, `usage`

### 2. Screenshot
```bash
python3 $SCRIPTS/crawl.py screenshot "https://example.com"
python3 $SCRIPTS/crawl.py screenshot "https://example.com" --pdf
python3 $SCRIPTS/crawl.py screenshot "https://example.com" --full-page false --wait-for ".content"
```
Key params: `--full-page` (default true), `--pdf`, `--wait-for` (CSS selector or seconds)
Response: `success`, `screenshot` (base64), `pdf` (base64), `usage`

### 3. Extract
```bash
python3 $SCRIPTS/crawl.py extract "https://books.toscrape.com" --query "get all books with title and price"
python3 $SCRIPTS/crawl.py extract "https://example.com" --method llm --query "summarize in 3 bullets"
python3 $SCRIPTS/crawl.py extract "https://example.com" --method schema --json-example '{"title":"...","price":"$0"}'
```
Key params: `--query`, `--method` (auto/llm/schema), `--json-example`, `--strategy`
Response: `success`, `data` (array of objects), `method_used`, `schema_used`

AUTO mode (default): system analyzes the page and picks CSS Schema (for structured pages) or LLM (for prose/unstructured).

### 4. Map
```bash
python3 $SCRIPTS/crawl.py map "https://crawl4ai.com" --max-urls 50
python3 $SCRIPTS/crawl.py map "https://docs.crawl4ai.com" --query "extraction" --score-threshold 0.3
```
Key params: `--mode` (default/deep), `--max-urls`, `--query` (BM25 scoring), `--score-threshold`, `--force`
Response: `success`, `domain`, `total_urls`, `urls` (array with relevance scores)
Results cached 7 days. Use `--force` to bypass.

### 5. Site Crawl
```bash
python3 $SCRIPTS/crawl.py site "https://docs.crawl4ai.com" --max-pages 10 --strategy http
python3 $SCRIPTS/crawl.py site "https://example.com" --discovery bfs --max-depth 2 --pattern "*/blog/*"
```
Key params: `--max-pages`, `--discovery` (map/bfs/dfs/best_first), `--strategy`, `--pattern`, `--max-depth`
Always async. Returns `job_id`. Use `--wait` to poll until complete.
Poll: `python3 $SCRIPTS/poll-job.py SCAN_ID --type deep`

## Full Power Mode (/v1/crawl)

When wrappers aren't enough (extraction strategies, magic mode, session management):

```bash
python3 $SCRIPTS/crawl.py crawl "https://example.com" \
  --crawler-config '{"screenshot":true,"extraction_strategy":{"type":"llm","provider":"openai/gpt-4o-mini","instruction":"extract products"}}'
```

All crawler_config fields: see `reference/crawler-config.md`
All browser_config fields: see `reference/browser-config.md`

## Config Passthrough (Power User)

Every wrapper endpoint accepts `--crawler-config` and `--browser-config` for full control:

```bash
# CSS selector to target specific content
python3 $SCRIPTS/crawl.py markdown "https://blog.example.com" \
  --crawler-config '{"css_selector":"article.post-body","excluded_tags":["nav","footer"]}'

# Custom headers
python3 $SCRIPTS/crawl.py markdown "https://example.com" \
  --browser-config '{"headers":{"Accept-Language":"fr-FR"}}'

# Wait for dynamic content
python3 $SCRIPTS/crawl.py markdown "https://spa-app.com" --strategy browser \
  --crawler-config '{"wait_for":".content-loaded","delay_before_return_html":2}'

# Execute JavaScript before extraction
python3 $SCRIPTS/crawl.py markdown "https://example.com" --strategy browser \
  --crawler-config '{"js_code":"document.querySelector(\".load-more\").click()"}'
```

### Critical crawler_config Fields

| Field | Type | What it does |
|-------|------|-------------|
| `css_selector` | string | Extract only matching elements (e.g., `"main"`, `"article"`) |
| `excluded_tags` | string[] | Strip these HTML tags (e.g., `["nav","footer","aside"]`) |
| `wait_for` | string | CSS selector to wait for before extracting |
| `wait_until` | string | Page event: `"domcontentloaded"`, `"networkidle"`, `"load"` |
| `delay_before_return_html` | float | Seconds to wait after load (for animations) |
| `js_code` | string | JavaScript to execute before extraction |
| `page_timeout` | int | Max ms for page load (default 30000) |
| `screenshot` | bool | Capture screenshot |
| `pdf` | bool | Generate PDF |

Full reference: `reference/crawler-config.md`

## Async / Batch Pattern

For multiple URLs:
```bash
# Create batch job
python3 $SCRIPTS/crawl.py async markdown "https://a.com,https://b.com,https://c.com"
# Returns: {"job_id": "job_xxx", "status": "pending", "urls_count": 3}

# Poll until done
python3 $SCRIPTS/poll-job.py job_xxx --type markdown

# Or use --wait to auto-poll
python3 $SCRIPTS/crawl.py async markdown "https://a.com,https://b.com" --wait
```

Batch types: `markdown`, `screenshot`, `extract`
For extract batch: `--method` must be `llm` or `schema` (not `auto`).

## Common Mistakes

1. **Don't build complex crawler_config when a wrapper exists.** `crawl.py markdown URL` is simpler than `crawl.py crawl URL --crawler-config '{"content_filter":{"type":"PruningContentFilter"}}'`

2. **Use `--strategy http` for static pages.** 5x faster than browser.

3. **Don't use `--method auto` for batch extract.** Specify `llm` or `schema`.

4. **Check the `success` field** before using response data.

5. **Credits are always deducted** (even for cached results). Check `usage.credits_used` in response.

6. **Proxy is ON by default.** No need to configure unless you need geo-targeting.

## Diagnostics

If something isn't working:
```bash
bash $SCRIPTS/health.sh       # Is API reachable? What's storage usage?
bash $SCRIPTS/test-key.sh     # Is the API key valid? What plan?
```

## Reference

- `reference/endpoints.md` -- all endpoints with full params and response schemas
- `reference/crawler-config.md` -- every crawler_config field explained
- `reference/browser-config.md` -- every browser_config field
- `reference/errors.md` -- error codes and recovery
- `examples/common.md` -- 10 common patterns
- `examples/advanced.md` -- power user patterns
