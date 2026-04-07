# Endpoint Reference

Base URL: `https://api.crawl4ai.com`
Auth header: `X-API-Key: <key>`

---

## POST /v1/markdown

Get clean markdown from a URL.

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| url | string | **required** | URL to convert |
| strategy | `"browser"\|"http"` | `"browser"` | `http` is 5x faster for static pages |
| fit | bool | `true` | Prune nav/footer/boilerplate |
| include | string[] | `null` | Extra data: `"links"`, `"media"`, `"metadata"`, `"tables"` |
| crawler_config | object | `null` | CrawlerRunConfig overrides |
| browser_config | object | `null` | BrowserConfig overrides |
| proxy | ProxyConfig | `null` | `{"mode":"datacenter\|residential\|auto","country":"US"}` |
| bypass_cache | bool | `false` | Skip server-side cache |

**Response:**
```json
{"success":true,"url":"...","markdown":"...","fit_markdown":"...","fit_html":"...",
 "links":{},"media":{},"metadata":{},"tables":[],
 "duration_ms":1200,"usage":{"credits_used":1,"credits_remaining":99},
 "error_message":null}
```

---

## POST /v1/markdown/async

Async markdown job. Single URL or batch (max 100).

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| url | string | `null` | Single URL (use `url` OR `urls`, not both) |
| urls | string[] | `null` | Batch URLs, max 100 |
| webhook_url | string | `null` | POST callback on completion |
| priority | int (1-10) | `5` | Job priority (1=highest) |
| *(plus all /v1/markdown params except url)* | | | |

**Response:** `{"job_id":"...","status":"pending","urls_count":3,"created_at":"ISO"}`

---

## GET /v1/markdown/jobs/{job_id}

**Response:**
```json
{"job_id":"...","status":"pending|running|completed|partial|failed|cancelled",
 "progress":{"total":3,"completed":2,"failed":0},"progress_percent":66,
 "results":[...],"download_url":"...","error":null,
 "created_at":"ISO","started_at":"ISO","completed_at":"ISO"}
```

## GET /v1/markdown/jobs

Query: `?limit=20&offset=0&status=completed`

**Response:** `{"jobs":[{"job_id":"...","status":"...","urls_count":3,"progress_percent":100,"created_at":"ISO","completed_at":"ISO"}],"total":5,"limit":20,"offset":0}`

## DELETE /v1/markdown/jobs/{job_id}

**Response:** `{"job_id":"...","status":"cancelled"}`

---

## POST /v1/screenshot

Capture screenshot or PDF. Always uses browser strategy.

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| url | string | **required** | URL to capture |
| full_page | bool | `true` | Full scrollable page vs viewport only |
| pdf | bool | `false` | Generate PDF |
| wait_for | string | `null` | CSS selector or seconds (e.g. `".content"` or `"2"`) |
| crawler_config | object | `null` | CrawlerRunConfig overrides |
| browser_config | object | `null` | BrowserConfig overrides |
| proxy | ProxyConfig | `null` | Proxy config |
| bypass_cache | bool | `false` | Skip cache |

**Response:**
```json
{"success":true,"url":"...","screenshot":"base64...","pdf":"base64...",
 "duration_ms":3200,"usage":{"credits_used":1,"credits_remaining":99},
 "error_message":null}
```

## POST /v1/screenshot/async

Same as markdown async pattern. `url` OR `urls`, `webhook_url`, `priority`.

## GET /v1/screenshot/jobs/{job_id} | GET /v1/screenshot/jobs | DELETE /v1/screenshot/jobs/{job_id}

Same response shapes as markdown job endpoints.

---

## POST /v1/extract

Extract structured data. AUTO mode picks best method.

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| url | string | **required** | URL to extract from |
| query | string | `null` | What to extract: `"get all products with title and price"` |
| json_example | object | `null` | Desired output shape: `{"title":"...","price":"$0"}` |
| schema | object | `null` | Pre-built CSS schema (reuse from previous run) |
| method | `"auto"\|"llm"\|"schema"` | `"auto"` | AUTO analyzes page, picks CSS or LLM |
| strategy | `"browser"\|"http"` | `"http"` | Page fetch strategy |
| crawler_config | object | `null` | CrawlerRunConfig overrides |
| browser_config | object | `null` | BrowserConfig overrides |
| llm_config | object | `null` | BYOK: `{"provider":"openai","model":"gpt-4o-mini"}` |
| proxy | ProxyConfig | `null` | Proxy config |

**Response:**
```json
{"success":true,"data":[{"title":"...","price":"$9.99"}],
 "method_used":"css_schema|llm","schema_used":{...},"query_used":"...",
 "url":"...","llm_usage":{"prompt_tokens":500,"completion_tokens":200,"total_tokens":700},
 "duration_ms":4500,"error_message":null}
```

## POST /v1/extract/async

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| url/urls | string/string[] | **required** | Single or batch (max 100) |
| method | `"auto"\|"llm"\|"schema"` | `"auto"` | **AUTO not allowed for batch** |
| webhook_url | string | `null` | Callback URL |
| priority | int (1-10) | `5` | Job priority |
| *(plus query, json_example, schema, strategy, llm_config, proxy, bypass_cache)* | | | |

## GET /v1/extract/jobs/{job_id} | GET /v1/extract/jobs | DELETE /v1/extract/jobs/{job_id}

Same job pattern as markdown/screenshot.

---

## POST /v1/map

Discover all URLs on a domain. Synchronous.

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| url | string | **required** | Domain to map (e.g. `https://example.com`) |
| mode | `"default"\|"deep"` | `"default"` | `default` ~2-15s, `deep` ~30-60s |
| max_urls | int | `null` | Limit results |
| include_subdomains | bool | `false` | Include subdomains |
| extract_head | bool | `true` | Fetch title/description per URL |
| query | string | `null` | BM25 relevance scoring |
| score_threshold | float (0-1) | `null` | Min relevance score (requires query) |
| force | bool | `false` | Bypass 7-day cache |
| proxy | ProxyConfig | `null` | Proxy config |

**Response:**
```json
{"success":true,"domain":"example.com","total_urls":42,"hosts_found":3,"mode":"default",
 "urls":[{"url":"...","host":"...","status":"valid","relevance_score":0.85,"head_data":{"title":"..."}}],
 "duration_ms":5200,"error_message":null}
```

No async mode. Results cached 7 days.

---

## POST /v1/crawl/site

Crawl entire website. Always async.

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| url | string | **required** | Starting URL |
| max_pages | int (1-1000) | `20` | Max pages to crawl |
| discovery | `"map"\|"bfs"\|"dfs"\|"best_first"` | `"map"` | URL discovery strategy |
| strategy | `"browser"\|"http"` | `"browser"` | Per-page crawl strategy |
| fit | bool | `true` | Prune content |
| include | string[] | `null` | Extra data per page |
| pattern | string | `null` | URL glob filter: `"*/blog/*"` |
| max_depth | int (1-10) | `null` | Max link depth (bfs/dfs/best_first) |
| crawler_config | object | `null` | CrawlerRunConfig overrides |
| browser_config | object | `null` | BrowserConfig overrides |
| proxy | ProxyConfig | `null` | Proxy config |
| webhook_url | string | `null` | Callback URL |
| priority | int (1-10) | `5` | Job priority |

**Response:** `{"job_id":"...","status":"pending","strategy":"map","discovered_urls":42,"queued_urls":20,"created_at":"ISO"}`

**Poll:** `GET /v1/crawl/deep/jobs/{job_id}` (uses deep crawl job system)

---

## POST /v1/crawl

Full power endpoint. All 40+ CrawlerRunConfig fields available.

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| url | string | **required** | URL to crawl |
| strategy | `"browser"\|"http"` | `"browser"` | Crawl strategy |
| crawler_config | object | `null` | Full CrawlerRunConfig (see crawler-config.md) |
| browser_config | object | `null` | Full BrowserConfig (see browser-config.md) |
| proxy | ProxyConfig | `null` | Proxy config |
| bypass_cache | bool | `false` | Skip cache |
| include_fields | string[] | `null` | Filter response fields |

**Response:** Full CrawlResponse with `html`, `cleaned_html`, `markdown`, `media`, `links`, `metadata`, `screenshot`, `pdf`, `extracted_content`, `usage`, etc.

## POST /v1/crawl/async | GET /v1/crawl/jobs/{job_id} | GET /v1/crawl/jobs | DELETE /v1/crawl/jobs/{job_id}

Async version accepts `urls` (max 100), `webhook_url`, `priority`. Job endpoints follow same pattern.

---

## ProxyConfig Object

```json
{"mode": "none|datacenter|residential|auto", "country": "US", "sticky_session": false}
```

| Mode | Surcharge | Use when |
|------|-----------|----------|
| `none` | +0 credits | Default, fast |
| `datacenter` | +1 credit | Geo-targeting, basic anti-bot |
| `residential` | +4 credits | Aggressive anti-bot sites |
| `auto` | varies | Let system decide escalation |
